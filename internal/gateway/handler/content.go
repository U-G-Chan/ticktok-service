package handler

import (
	"net/http"
	"strconv"
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
		// Token 在 gRPC 侧暂未使用，可为空
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
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
