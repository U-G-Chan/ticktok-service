package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	contentv1 "ticktok-service/api/content/v1"
	userv1 "ticktok-service/api/user/v1"
	"ticktok-service/internal/content/model"
	"ticktok-service/internal/content/repository"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"

	goredis "github.com/go-redis/redis/v8"
)

const feedKey = "feed:recommend_score"

type ContentService interface {
	GetFeed(ctx context.Context, lastScore int32, lastID int64) ([]*contentv1.Video, int32, int64, error)
	GetVideoUploadURL(ctx context.Context, authorID int64, title string) (string, int64, error)
	PublishVideo(ctx context.Context, videoID int64) error
	GetPublishList(ctx context.Context, userID int64) ([]*contentv1.Video, error)
}

type contentService struct {
	repo       repository.VideoRepository
	userClient userv1.UserServiceClient
}

func NewContentService(repo repository.VideoRepository, userClient userv1.UserServiceClient) ContentService {
	return &contentService{
		repo:       repo,
		userClient: userClient,
	}
}

func (s *contentService) GetFeed(ctx context.Context, lastScore int32, lastID int64) ([]*contentv1.Video, int32, int64, error) {
	limit := 30
	var videos []*model.Video
	var err error
	var cacheHit bool

	// 1. Try to get Feed from Redis ZSet
	startRank := int64(0)
	if lastID > 0 {
		rank, err := redis.RDB.ZRevRank(ctx, feedKey, strconv.FormatInt(lastID, 10)).Result()
		if err == nil {
			startRank = rank + 1
		} else {
			// Cache miss for cursor, fallback to DB
			startRank = -1
		}
	}

	if startRank >= 0 {
		// Fetch video IDs from Redis
		videoIDStrs, err := redis.RDB.ZRevRange(ctx, feedKey, startRank, startRank+int64(limit)-1).Result()
		if err == nil && len(videoIDStrs) > 0 {
			var ids []int64
			for _, idStr := range videoIDStrs {
				if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
					ids = append(ids, id)
				}
			}

			// Batch fetch from MySQL
			unsortedVideos, err := s.repo.GetByIDs(ctx, ids)
			if err == nil && len(unsortedVideos) > 0 {
				cacheHit = true
				// MySQL IN query doesn't preserve order, we need to map and sort by Redis ZSet order
				videoMap := make(map[int64]*model.Video)
				for _, v := range unsortedVideos {
					videoMap[v.ID] = v
				}
				for _, id := range ids {
					if v, ok := videoMap[id]; ok {
						videos = append(videos, v)
					}
				}
			}
		}
	}

	// 2. Fallback to DB if cache miss
	if !cacheHit {
		videos, err = s.repo.GetFeed(ctx, lastScore, lastID, limit)
		if err != nil {
			return nil, 0, 0, err
		}
	}

	if len(videos) == 0 {
		return nil, 0, 0, nil
	}

	// 3. Collect unique author IDs
	authorIDs := make([]int64, 0)
	authorMap := make(map[int64]*userv1.User)
	seenAuthor := make(map[int64]struct{})

	for _, v := range videos {
		if _, ok := seenAuthor[v.AuthorID]; !ok {
			seenAuthor[v.AuthorID] = struct{}{}
			authorIDs = append(authorIDs, v.AuthorID)
		}
	}

	// 2. Batch fetch user info
	userInfoResp, err := s.userClient.MGetUserInfo(ctx, &userv1.MGetUserInfoRequest{UserIds: authorIDs})
	if err == nil && userInfoResp.Code == int32(errno.Success.Code) {
		for _, u := range userInfoResp.Users {
			authorMap[u.Id] = u
		}
	}

	var pbVideos []*contentv1.Video
	var nextScore int32
	var nextID int64

	for i, v := range videos {
		// 3. Match author info
		author, ok := authorMap[v.AuthorID]
		if !ok {
			// Fallback
			author = &userv1.User{Id: v.AuthorID}
		}

		pbVideos = append(pbVideos, &contentv1.Video{
			Id:             v.ID,
			Author:         author,
			PlayUrl:        v.PlayURL,
			CoverUrl:       v.CoverURL,
			FavoriteCount:  v.FavoriteCount,
			CommentCount:   v.CommentCount,
			Title:          v.Title,
			RecommendScore: v.RecommendScore,
		})

		// 记录最后一条作为下一页游标
		if i == len(videos)-1 {
			nextScore = v.RecommendScore
			nextID = v.ID
		}
	}

	return pbVideos, nextScore, nextID, nil
}

func (s *contentService) GetVideoUploadURL(ctx context.Context, authorID int64, title string) (string, int64, error) {
	// 1. 生成全局唯一 ID
	videoID := snowflake.GenerateMsgID()

	// 2. 预先创建 Pending 状态的记录
	objectName := fmt.Sprintf("%d.mp4", videoID)
	video := &model.Video{
		ID:       videoID,
		AuthorID: authorID,
		Title:    title,
		PlayURL:  minio.GetObjectURL(objectName),
		CoverURL: minio.GetObjectURL(fmt.Sprintf("%d.jpg", videoID)), // 预留封面
		Status:   0,                                                 // Pending
	}

	if err := s.repo.Create(ctx, video); err != nil {
		return "", 0, err
	}

	// 3. 生成 15 分钟有效的预签名 URL
	uploadURL, err := minio.GeneratePresignedPutURL(ctx, objectName, 15*time.Minute)
	if err != nil {
		return "", 0, err
	}

	return uploadURL, videoID, nil
}

func (s *contentService) PublishVideo(ctx context.Context, videoID int64) error {
	video, err := s.repo.GetByID(ctx, videoID)
	if err != nil {
		return errno.ErrVideoNotFound
	}

	// 更新状态为 Published
	video.Status = 1
	// 这里模拟生成推荐分数，比如随机 1-1000
	video.RecommendScore = 500 // 实际可以接入推荐模型

	err = s.repo.Update(ctx, video)
	if err != nil {
		return err
	}

	// 异步更新 Redis Feed 流缓存和发送 Kafka 消息用于抽帧
	go func() {
		bgCtx := context.Background()
		redis.RDB.ZAdd(bgCtx, feedKey, &goredis.Z{
			Score:  float64(video.RecommendScore),
			Member: strconv.FormatInt(videoID, 10),
		})
		// 限制 ZSet 大小，例如只保留最新/最热的 5000 条，防止 BigKey
		redis.RDB.ZRemRangeByRank(bgCtx, feedKey, 0, -5001)

		// 发送 Kafka 消息，通知 worker 截取封面
		// 消息的 value 就是 videoID 的字符串形式
		err := kafka.SendMessageToTopic(bgCtx, "video_publish_events", []byte(strconv.FormatInt(videoID, 10)), []byte(strconv.FormatInt(videoID, 10)))
		if err != nil {
			fmt.Printf("failed to send kafka message for video %d: %v\n", videoID, err)
		}
	}()

	return nil
}

func (s *contentService) GetPublishList(ctx context.Context, userID int64) ([]*contentv1.Video, error) {
	videos, err := s.repo.GetByAuthorID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(videos) == 0 {
		return nil, nil
	}

	// 1. Fetch user info (only one author for all these videos)
	var author *userv1.User
	userInfoResp, err := s.userClient.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserId: userID})
	if err == nil && userInfoResp.Code == int32(errno.Success.Code) {
		author = userInfoResp.User
	} else {
		// Fallback
		author = &userv1.User{Id: userID}
	}

	var pbVideos []*contentv1.Video
	for _, v := range videos {
		pbVideos = append(pbVideos, &contentv1.Video{
			Id:             v.ID,
			Author:         author,
			PlayUrl:        v.PlayURL,
			CoverUrl:       v.CoverURL,
			FavoriteCount:  v.FavoriteCount,
			CommentCount:   v.CommentCount,
			Title:          v.Title,
			RecommendScore: v.RecommendScore,
		})
	}
	return pbVideos, nil
}
