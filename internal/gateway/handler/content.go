package handler

import (
	"net/http"
	"strconv"
	"strings"
	contentv1 "ticktok-service/api/content/v1"
	"ticktok-service/pkg/rpc"

	"github.com/gin-gonic/gin"
)

// GetFeed 获取视频流
func GetFeed(c *gin.Context) {
	lastScoreStr := c.DefaultQuery("last_score", "0")
	lastIDStr := c.DefaultQuery("last_id", "0")
	token := c.DefaultQuery("token", "")

	lastScore, _ := strconv.ParseInt(lastScoreStr, 10, 32)
	lastID, _ := strconv.ParseInt(lastIDStr, 10, 64)

	resp, err := rpc.GetClientManager().ContentClient.GetFeed(c.Request.Context(), &contentv1.GetFeedRequest{
		LastScore: int32(lastScore),
		LastId:    lastID,
		Token:     token,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetVideoUploadURL 获取预签名上传URL
func GetVideoUploadURL(c *gin.Context) {
	// 从 token 解析出的用户ID
	authorIDRaw, _ := c.Get("userID")
	authorID := authorIDRaw.(int64)

	title := c.Query("title")
	if title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "title is required"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.GetVideoUploadURL(c.Request.Context(), &contentv1.GetVideoUploadURLRequest{
		AuthorId: authorID,
		Title:    title,
		Token:    bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// PublishVideo 确认发布视频
func PublishVideo(c *gin.Context) {
	var req struct {
		VideoID int64 `json:"video_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid request body"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.PublishVideo(c.Request.Context(), &contentv1.PublishVideoRequest{
		VideoId: req.VideoID,
		Token:   bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPublishList 获取用户的发布列表
func GetPublishList(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid user_id"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.GetPublishList(c.Request.Context(), &contentv1.GetPublishListRequest{
		UserId: userID,
		Token:  bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func FavoriteAction(c *gin.Context) {
	var req struct {
		VideoID    int64 `json:"video_id" binding:"required"`
		ActionType int32 `json:"action_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid request body"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.FavoriteAction(c.Request.Context(), &contentv1.FavoriteActionRequest{
		VideoId:    req.VideoID,
		ActionType: req.ActionType,
		Token:      bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func CommentAction(c *gin.Context) {
	var req struct {
		VideoID     int64  `json:"video_id" binding:"required"`
		ActionType  int32  `json:"action_type" binding:"required"`
		CommentText string `json:"comment_text"`
		CommentID   int64  `json:"comment_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid request body"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.CommentAction(c.Request.Context(), &contentv1.CommentActionRequest{
		VideoId:     req.VideoID,
		ActionType:  req.ActionType,
		CommentText: req.CommentText,
		CommentId:   req.CommentID,
		Token:       bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func GetCommentList(c *gin.Context) {
	videoIDStr := c.Query("video_id")
	videoID, err := strconv.ParseInt(videoIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid video_id"})
		return
	}

	resp, err := rpc.GetClientManager().ContentClient.GetCommentList(c.Request.Context(), &contentv1.GetCommentListRequest{
		VideoId: videoID,
		Token:   bearerToken(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func bearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	}
	return ""
}
