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
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"

	goredis "github.com/go-redis/redis/v8"
)

const feedKey = "feed:recommend_score"

type ContentService interface {
	GetFeed(ctx context.Context, lastScore int32, lastID int64, token string) ([]*contentv1.Video, int32, int64, error)
	GetFollowFeed(ctx context.Context, userID int64, lastTime int64, token string) ([]*contentv1.Video, int64, error)
	GetVideoUploadURL(ctx context.Context, authorID int64, title string) (string, int64, error)
	PublishVideo(ctx context.Context, videoID int64) error
	GetPublishList(ctx context.Context, userID int64, token string) ([]*contentv1.Video, error)
	FavoriteAction(ctx context.Context, userID, videoID int64, actionType int32) error
	CommentAction(ctx context.Context, userID, videoID int64, actionType int32, commentText string, commentID int64) (*contentv1.Comment, error)
	GetCommentList(ctx context.Context, videoID int64) ([]*contentv1.Comment, error)
}

type contentService struct {
	repo        repository.VideoRepository
	favoriteRepo repository.FavoriteRepository
	commentRepo repository.CommentRepository
	userClient  userv1.UserServiceClient
}

func NewContentService(
	repo repository.VideoRepository,
	favoriteRepo repository.FavoriteRepository,
	commentRepo repository.CommentRepository,
	userClient userv1.UserServiceClient,
) ContentService {
	return &contentService{
		repo:        repo,
		favoriteRepo: favoriteRepo,
		commentRepo: commentRepo,
		userClient:  userClient,
	}
}

func (s *contentService) GetFeed(ctx context.Context, lastScore int32, lastID int64, token string) ([]*contentv1.Video, int32, int64, error) {
	limit := 30
	var videos []*model.Video
	var err error
	var cacheHit bool
	viewerID := parseUserIDFromToken(token)

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
		author, ok := authorMap[v.AuthorID]
		if !ok {
			author = &userv1.User{Id: v.AuthorID}
		}

		pbVideo, buildErr := s.buildPBVideo(ctx, v, author, viewerID)
		if buildErr != nil {
			return nil, 0, 0, buildErr
		}
		pbVideos = append(pbVideos, pbVideo)

		if i == len(videos)-1 {
			nextScore = v.RecommendScore
			nextID = v.ID
		}
	}

	return pbVideos, nextScore, nextID, nil
}

func (s *contentService) GetFollowFeed(ctx context.Context, userID int64, lastTime int64, token string) ([]*contentv1.Video, int64, error) {
	limit := 30
	if lastTime == 0 {
		lastTime = time.Now().UnixMilli()
	}

	var allVideoIDs []string
	var allScores []float64

	// 1. 从收件箱 (Inbox) 获取 Push 过来的视频 (普通用户的更新)
	inboxKey := fmt.Sprintf("user:inbox:%d", userID)
	inboxVideos, err := redis.RDB.ZRevRangeByScoreWithScores(ctx, inboxKey, &goredis.ZRangeBy{
		Max:   strconv.FormatInt(lastTime, 10),
		Min:   "0",
		Count: int64(limit),
	}).Result()

	if err == nil {
		for _, z := range inboxVideos {
			allVideoIDs = append(allVideoIDs, z.Member.(string))
			allScores = append(allScores, z.Score)
		}
	}

	// 2. 获取关注列表，筛选出大V，从他们的发件箱 (Outbox) Pull 视频
	followResp, err := s.userClient.GetFollowList(ctx, &userv1.GetFollowListRequest{
		UserId:      userID,
		TokenUserId: userID,
	})

	if err == nil && followResp.Code == int32(errno.Success.Code) {
		for _, followedUser := range followResp.UserList {
			// 假设粉丝数大于等于 10000 的是大V，走 Pull 模式
			// 为了测试方便，这里假设 follower_count >= 0 都可以 pull (由于没有真实的 follower_count 统计，我们以一定条件判断，
			// 或者在项目中简化为直接 Pull 所有关注者，但在文档中强调 Push-Pull 结合。为了体现代码，我们这里写死一个阈值，测试时可以修改)
			if followedUser.FollowerCount >= 10000 {
				outboxKey := fmt.Sprintf("user:outbox:%d", followedUser.Id)
				outboxVideos, err := redis.RDB.ZRevRangeByScoreWithScores(ctx, outboxKey, &goredis.ZRangeBy{
					Max:   strconv.FormatInt(lastTime, 10),
					Min:   "0",
					Count: int64(limit),
				}).Result()
				if err == nil {
					for _, z := range outboxVideos {
						allVideoIDs = append(allVideoIDs, z.Member.(string))
						allScores = append(allScores, z.Score)
					}
				}
			}
		}
	}

	// 3. 内存排序与去重
	type videoItem struct {
		id    int64
		score float64
	}
	var items []videoItem
	seen := make(map[int64]bool)

	for i, idStr := range allVideoIDs {
		id, _ := strconv.ParseInt(idStr, 10, 64)
		if !seen[id] {
			seen[id] = true
			items = append(items, videoItem{id: id, score: allScores[i]})
		}
	}

	// 按时间倒序排序
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].score < items[j].score {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// 取前 limit 个
	if len(items) > limit {
		items = items[:limit]
	}

	if len(items) == 0 {
		return nil, 0, nil
	}

	var fetchIDs []int64
	var nextTime int64
	for i, item := range items {
		fetchIDs = append(fetchIDs, item.id)
		if i == len(items)-1 {
			// 下一页的 lastTime 是最后一条记录的时间戳 - 1
			nextTime = int64(item.score) - 1
		}
	}

	// 4. 从数据库批量获取视频信息并组装
	unsortedVideos, err := s.repo.GetByIDs(ctx, fetchIDs)
	if err != nil {
		return nil, 0, err
	}

	videoMap := make(map[int64]*model.Video)
	for _, v := range unsortedVideos {
		videoMap[v.ID] = v
	}

	// 收集作者信息
	authorIDs := make([]int64, 0)
	authorMap := make(map[int64]*userv1.User)
	seenAuthor := make(map[int64]struct{})

	for _, v := range unsortedVideos {
		if _, ok := seenAuthor[v.AuthorID]; !ok {
			seenAuthor[v.AuthorID] = struct{}{}
			authorIDs = append(authorIDs, v.AuthorID)
		}
	}

	userInfoResp, err := s.userClient.MGetUserInfo(ctx, &userv1.MGetUserInfoRequest{
		UserIds:     authorIDs,
		TokenUserId: userID,
	})
	if err == nil && userInfoResp.Code == int32(errno.Success.Code) {
		for _, u := range userInfoResp.Users {
			authorMap[u.Id] = u
		}
	}

	var pbVideos []*contentv1.Video
	for _, id := range fetchIDs {
		if v, ok := videoMap[id]; ok {
			author, aok := authorMap[v.AuthorID]
			if !aok {
				author = &userv1.User{Id: v.AuthorID}
			}
			pbVideo, buildErr := s.buildPBVideo(ctx, v, author, userID)
			if buildErr != nil {
				continue
			}
			pbVideos = append(pbVideos, pbVideo)
		}
	}

	return pbVideos, nextTime, nil
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

	// 异步更新 Redis Feed 流缓存、Timeline 和发送 Kafka 消息用于抽帧
	go func() {
		bgCtx := context.Background()
		now := time.Now().UnixMilli()
		videoIDStr := strconv.FormatInt(videoID, 10)

		// 1. 全局 Feed 流 (Recommend)
		redis.RDB.ZAdd(bgCtx, feedKey, &goredis.Z{
			Score:  float64(video.RecommendScore),
			Member: videoIDStr,
		})
		redis.RDB.ZRemRangeByRank(bgCtx, feedKey, 0, -5001)

		// 2. 写入发件箱 (Outbox) - 用于 Pull 模式
		outboxKey := fmt.Sprintf("user:outbox:%d", video.AuthorID)
		redis.RDB.ZAdd(bgCtx, outboxKey, &goredis.Z{
			Score:  float64(now),
			Member: videoIDStr,
		})
		redis.RDB.ZRemRangeByRank(bgCtx, outboxKey, 0, -1001) // 限制每个用户发件箱大小

		// 3. 写扩散 (Push 模式) - 将视频推送到粉丝的收件箱 (Inbox)
		// 获取粉丝列表
		followerResp, err := s.userClient.GetFollowerList(bgCtx, &userv1.GetFollowerListRequest{
			UserId: video.AuthorID,
		})
		if err == nil && followerResp.Code == int32(errno.Success.Code) {
			// 如果粉丝数量很大（大V），不进行全量推，让活跃粉丝拉取（Pull）
			// 这里假设粉丝数小于 10000 走写扩散 (Push)
			if len(followerResp.UserList) < 10000 {
				for _, follower := range followerResp.UserList {
					inboxKey := fmt.Sprintf("user:inbox:%d", follower.Id)
					redis.RDB.ZAdd(bgCtx, inboxKey, &goredis.Z{
						Score:  float64(now),
						Member: videoIDStr,
					})
					// 限制收件箱大小
					redis.RDB.ZRemRangeByRank(bgCtx, inboxKey, 0, -1001)
				}
			}
		}

		// 4. 发送 Kafka 消息用于异步抽帧
		topic := config.Config.Kafka.VideoPublishTopic
		if topic == "" {
			topic = "video_publish_events"
		}
		err = kafka.SendMessageToTopic(bgCtx, topic, []byte(videoIDStr), []byte(videoIDStr))
		if err != nil {
			fmt.Printf("failed to send kafka message for video %d: %v\n", videoID, err)
		}
	}()

	return nil
}

func (s *contentService) GetPublishList(ctx context.Context, userID int64, token string) ([]*contentv1.Video, error) {
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
		author = &userv1.User{Id: userID}
	}

	viewerID := parseUserIDFromToken(token)
	var pbVideos []*contentv1.Video
	for _, v := range videos {
		pbVideo, buildErr := s.buildPBVideo(ctx, v, author, viewerID)
		if buildErr != nil {
			return nil, buildErr
		}
		pbVideos = append(pbVideos, pbVideo)
	}
	return pbVideos, nil
}
