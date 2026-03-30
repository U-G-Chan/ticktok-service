package router

import (
	"ticktok-service/internal/gateway/handler"
	"ticktok-service/internal/gateway/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	r := gin.Default()

	// Public routes
	auth := r.Group("/api/v1/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.POST("/refresh", handler.Refresh)
	}

	// 开放的视频流接口 (不需要 token 也可以看，也可以带 token 看)
	r.GET("/api/v1/feed", handler.GetFeed)

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTMiddleware())
	{
		api.GET("/feed/follow", handler.GetFollowFeed)

		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
				"user_id": c.MustGet("userID"),
			})
		})

		// User routes
		userGroup := api.Group("/user")
		{
			userGroup.GET("/info", handler.GetUserInfo)
		}

		relationGroup := api.Group("/relation")
		{
			relationGroup.POST("/action", handler.RelationAction)
			relationGroup.GET("/follow/list", handler.GetFollowList)
			relationGroup.GET("/follower/list", handler.GetFollowerList)
		}

		// Content routes (Publish)
		publishGroup := api.Group("/publish")
		{
			publishGroup.GET("/action/url", handler.GetVideoUploadURL) // 获取上传URL
			publishGroup.POST("/action/confirm", handler.PublishVideo) // 确认发布
			publishGroup.GET("/list", handler.GetPublishList)          // 获取发布列表
		}

		favoriteGroup := api.Group("/favorite")
		{
			favoriteGroup.POST("/action", handler.FavoriteAction)
		}

		commentGroup := api.Group("/comment")
		{
			commentGroup.POST("/action", handler.CommentAction)
			commentGroup.GET("/list", handler.GetCommentList)
		}

		// Chatbot routes
		chat := api.Group("/chat")
		{
			chat.POST("/completions", handler.ChatCompletions)
		}

		chatbot := api.Group("/chatbot")
		{
			chatbot.POST("/newConversation", handler.CreateConversation)
			chatbot.GET("/history/list", handler.ListConversations)
			chatbot.GET("/history/:session_id", handler.GetChatHistory)
		}

		// Message routes
		msgGroup := api.Group("/message")
		{
			msgGroup.GET("/sync", handler.SyncMessages)
		}
	}

	// WebSocket route for messaging, outside JWTMiddleware because it uses query param token
	r.GET("/api/v1/message/connection", handler.MessageConnection)

	return r
}
