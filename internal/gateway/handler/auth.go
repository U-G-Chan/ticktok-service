package handler

import (
	"context"
	"net/http"
	"ticktok-service/api/user/v1"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/rpc"
	"ticktok-service/pkg/util"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func Register(c *gin.Context) {
	var req user.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	resp, err := rpc.GetClientManager().UserClient.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	// Store refresh token in Redis
	err = redis.RDB.Set(context.Background(), resp.RefreshToken, resp.UserId, time.Duration(config.Config.JWT.RefreshExpire)*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func Login(c *gin.Context) {
	var req user.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	resp, err := rpc.GetClientManager().UserClient.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": err.Error()})
		return
	}

	if resp.Code != 0 {
		c.JSON(http.StatusOK, resp)
		return
	}

	// Store refresh token in Redis
	err = redis.RDB.Set(context.Background(), resp.RefreshToken, resp.UserId, time.Duration(config.Config.JWT.RefreshExpire)*time.Second).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to store refresh token"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func Logout(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// Remove from Redis
	if err := redis.RDB.Del(context.Background(), req.RefreshToken).Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to delete refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "Logged out successfully"})
}

func Refresh(c *gin.Context) {
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": err.Error()})
		return
	}

	// Validate refresh token from Redis
	_, err := redis.RDB.Get(context.Background(), req.RefreshToken).Result()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid or expired refresh token"})
		return
	}

	// Parse to verify signature and expiration
	claims, err := util.ParseToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid refresh token"})
		return
	}

	// Generate new Access Token
	accessToken, _, err := util.GenerateToken(claims.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":          0,
		"msg":           "Success",
		"access_token":  accessToken,
		"refresh_token": req.RefreshToken,
	})
}
