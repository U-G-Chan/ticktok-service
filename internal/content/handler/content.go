package handler

import (
	"context"

	contentv1 "ticktok-service/api/content/v1"
	"ticktok-service/internal/content/service"
	"ticktok-service/pkg/errno"
	"ticktok-service/pkg/util"
)

type ContentHandler struct {
	contentv1.UnimplementedContentServiceServer
	svc service.ContentService
}

func NewContentHandler(svc service.ContentService) *ContentHandler {
	return &ContentHandler{svc: svc}
}

func (h *ContentHandler) GetFeed(ctx context.Context, req *contentv1.GetFeedRequest) (*contentv1.GetFeedResponse, error) {
	videos, nextScore, nextID, err := h.svc.GetFeed(ctx, req.LastScore, req.LastId, req.Token)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.GetFeedResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	return &contentv1.GetFeedResponse{
		Code:      int32(errno.Success.Code),
		Msg:       errno.Success.Message,
		VideoList: videos,
		NextScore: nextScore,
		NextId:    nextID,
	}, nil
}

func (h *ContentHandler) GetFollowFeed(ctx context.Context, req *contentv1.GetFollowFeedRequest) (*contentv1.GetFollowFeedResponse, error) {
	videos, nextTime, err := h.svc.GetFollowFeed(ctx, req.UserId, req.LastTime, req.Token)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.GetFollowFeedResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	return &contentv1.GetFollowFeedResponse{
		Code:      int32(errno.Success.Code),
		Msg:       errno.Success.Message,
		VideoList: videos,
		NextTime:  nextTime,
	}, nil
}

func (h *ContentHandler) GetVideoUploadURL(ctx context.Context, req *contentv1.GetVideoUploadURLRequest) (*contentv1.GetVideoUploadURLResponse, error) {
	url, videoID, err := h.svc.GetVideoUploadURL(ctx, req.AuthorId, req.Title)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.GetVideoUploadURLResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	return &contentv1.GetVideoUploadURLResponse{
		Code:      int32(errno.Success.Code),
		Msg:       errno.Success.Message,
		UploadUrl: url,
		VideoId:   videoID,
	}, nil
}

func (h *ContentHandler) PublishVideo(ctx context.Context, req *contentv1.PublishVideoRequest) (*contentv1.PublishVideoResponse, error) {
	err := h.svc.PublishVideo(ctx, req.VideoId)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.PublishVideoResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	return &contentv1.PublishVideoResponse{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Message,
	}, nil
}

func (h *ContentHandler) GetPublishList(ctx context.Context, req *contentv1.GetPublishListRequest) (*contentv1.GetPublishListResponse, error) {
	videos, err := h.svc.GetPublishList(ctx, req.UserId, req.Token)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.GetPublishListResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}

	return &contentv1.GetPublishListResponse{
		Code:      int32(errno.Success.Code),
		Msg:       errno.Success.Message,
		VideoList: videos,
	}, nil
}

func (h *ContentHandler) FavoriteAction(ctx context.Context, req *contentv1.FavoriteActionRequest) (*contentv1.FavoriteActionResponse, error) {
	userID, err := parseTokenUserID(req.Token)
	if err != nil {
		return &contentv1.FavoriteActionResponse{
			Code: int32(errno.ErrUnauthorized.Code),
			Msg:  errno.ErrUnauthorized.Message,
		}, nil
	}
	if err := h.svc.FavoriteAction(ctx, userID, req.VideoId, req.ActionType); err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.FavoriteActionResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}
	return &contentv1.FavoriteActionResponse{
		Code: int32(errno.Success.Code),
		Msg:  errno.Success.Message,
	}, nil
}

func (h *ContentHandler) CommentAction(ctx context.Context, req *contentv1.CommentActionRequest) (*contentv1.CommentActionResponse, error) {
	userID, err := parseTokenUserID(req.Token)
	if err != nil {
		return &contentv1.CommentActionResponse{
			Code: int32(errno.ErrUnauthorized.Code),
			Msg:  errno.ErrUnauthorized.Message,
		}, nil
	}
	comment, err := h.svc.CommentAction(ctx, userID, req.VideoId, req.ActionType, req.CommentText, req.CommentId)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.CommentActionResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}
	return &contentv1.CommentActionResponse{
		Code:    int32(errno.Success.Code),
		Msg:     errno.Success.Message,
		Comment: comment,
	}, nil
}

func (h *ContentHandler) GetCommentList(ctx context.Context, req *contentv1.GetCommentListRequest) (*contentv1.GetCommentListResponse, error) {
	comments, err := h.svc.GetCommentList(ctx, req.VideoId)
	if err != nil {
		code, msg := errno.DecodeErr(err)
		return &contentv1.GetCommentListResponse{
			Code: int32(code),
			Msg:  msg,
		}, nil
	}
	return &contentv1.GetCommentListResponse{
		Code:        int32(errno.Success.Code),
		Msg:         errno.Success.Message,
		CommentList: comments,
	}, nil
}

func parseTokenUserID(token string) (int64, error) {
	claims, err := util.ParseToken(token)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
