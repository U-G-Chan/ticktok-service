package handler

import (
	"net/http"
	"strconv"
	"ticktok-service/api/user/v1"
	"ticktok-service/pkg/rpc"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	var userId int64
	userIdStr := c.Query("user_id")
	if userIdStr != "" {
		parsed, err := strconv.ParseInt(userIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid user_id"})
			return
		}
		userId = parsed
	} else {
		userIDRaw, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "user_id is required"})
			return
		}
		userId = userIDRaw.(int64)
	}

	userIDRaw, _ := c.Get("userID")
	var tUid int64
	if userIDRaw != nil {
		tUid = userIDRaw.(int64)
	}

	resp, err := rpc.GetClientManager().UserClient.GetUserInfo(c.Request.Context(), &user.GetUserInfoRequest{
		UserId:      userId,
		TokenUserId: tUid,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func RelationAction(c *gin.Context) {
	var req struct {
		ToUserID   int64 `json:"to_user_id" binding:"required"`
		ActionType int32 `json:"action_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "invalid request body"})
		return
	}

	userIDRaw, _ := c.Get("userID")
	userID := userIDRaw.(int64)

	resp, err := rpc.GetClientManager().UserClient.RelationAction(c.Request.Context(), &user.RelationActionRequest{
		UserId:     userID,
		ToUserId:   req.ToUserID,
		ActionType: req.ActionType,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetFollowList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid user_id"})
		return
	}

	tokenUserIdRaw, _ := c.Get("userID")
	var tokenUserId int64
	if tokenUserIdRaw != nil {
		tokenUserId = tokenUserIdRaw.(int64)
	}

	resp, err := rpc.GetClientManager().UserClient.GetFollowList(c.Request.Context(), &user.GetFollowListRequest{
		UserId:      userId,
		TokenUserId: tokenUserId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func GetFollowerList(c *gin.Context) {
	userIdStr := c.Query("user_id")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid user_id"})
		return
	}

	tokenUserIdRaw, _ := c.Get("userID")
	var tokenUserId int64
	if tokenUserIdRaw != nil {
		tokenUserId = tokenUserIdRaw.(int64)
	}

	resp, err := rpc.GetClientManager().UserClient.GetFollowerList(c.Request.Context(), &user.GetFollowerListRequest{
		UserId:      userId,
		TokenUserId: tokenUserId,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
