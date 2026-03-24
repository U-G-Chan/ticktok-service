package handler

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/rpc"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
)

func MessageConnection(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Token is required"})
		return
	}

	claims, err := util.ParseToken(tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "msg": "Invalid token"})
		return
	}

	// Token is valid. Proxy the request to the message service HTTP port
	targetStr := "http://" + config.Config.Microservices.Message
	target, err := url.Parse(targetStr)
	if err != nil {
		logger.Log.Error("Invalid message service url: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Internal error"})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify the request to include user_id in the header or query for the message service to consume
	// Or message service can re-parse the token. Here we pass the user_id via a custom header.
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-User-Id", strconv.FormatInt(claims.UserID, 10))
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func SyncMessages(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	syncKeyStr := c.Query("sync_key")
	syncKey, err := strconv.ParseInt(syncKeyStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "msg": "Invalid sync_key"})
		return
	}

	req := &messagev1.SyncMessageListRequest{
		UserId:  userID,
		SyncKey: syncKey,
	}

	resp, err := rpc.GetClientManager().MessageClient.SyncMessageList(context.Background(), req)
	if err != nil {
		logger.Log.Error("Failed to sync messages: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "msg": "Internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":          resp.Code,
		"msg":           resp.Msg,
		"message_list":  resp.MessageList,
		"next_sync_key": resp.NextSyncKey,
	})
}
