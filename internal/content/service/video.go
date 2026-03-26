package service

import (
	"context"
	"fmt"
	"time"

	contentv1 "ticktok-service/api/content/v1"
	userv1 "ticktok-service/api/user/v1"
	"ticktok-service/internal/content/model"
	"ticktok-service/internal/content/repository"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/snowflake"
)

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
	videos, err := s.repo.GetFeed(ctx, lastScore, lastID, limit)
	if err != nil {
		return nil, 0, 0, err
	}

	var pbVideos []*contentv1.Video
	var nextScore int32
	var nextID int64

	for i, v := range videos {
		// 调用 User 服务获取作者信息
		var author *userv1.User
		userInfoResp, err := s.userClient.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserId: v.AuthorID})
		if err == nil && userInfoResp.Code == int32(errno.Success.Code) {
			author = userInfoResp.User
		} else {
			// 降级处理
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

	return s.repo.Update(ctx, video)
}

func (s *contentService) GetPublishList(ctx context.Context, userID int64) ([]*contentv1.Video, error) {
	videos, err := s.repo.GetByAuthorID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var pbVideos []*contentv1.Video
	for _, v := range videos {
		// 调用 User 服务获取作者信息
		var author *userv1.User
		userInfoResp, err := s.userClient.GetUserInfo(ctx, &userv1.GetUserInfoRequest{UserId: v.AuthorID})
		if err == nil && userInfoResp.Code == int32(errno.Success.Code) {
			author = userInfoResp.User
		} else {
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
	}
	return pbVideos, nil
}
