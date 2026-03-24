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

	// Protected routes
	api := r.Group("/api/v1")
	api.Use(middleware.JWTMiddleware())
	{
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"message": "pong",
				"user_id": c.MustGet("userID"),
			})
		})

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
