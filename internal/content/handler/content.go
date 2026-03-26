package handler

import (
	"context"

	contentv1 "ticktok-service/api/content/v1"
	"ticktok-service/internal/content/service"
	"ticktok-service/pkg/errno"
)

type ContentHandler struct {
	contentv1.UnimplementedContentServiceServer
	svc service.ContentService
}

func NewContentHandler(svc service.ContentService) *ContentHandler {
	return &ContentHandler{svc: svc}
}

func (h *ContentHandler) GetFeed(ctx context.Context, req *contentv1.GetFeedRequest) (*contentv1.GetFeedResponse, error) {
	videos, nextScore, nextID, err := h.svc.GetFeed(ctx, req.LastScore, req.LastId)
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
	videos, err := h.svc.GetPublishList(ctx, req.UserId)
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
