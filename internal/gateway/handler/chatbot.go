package handler

import (
	"context"
	"io"
	"net/http"

	pb "ticktok-service/api/chatbot/v1"
	"ticktok-service/pkg/rpc"

	"github.com/gin-gonic/gin"
)

type ChatCompletionRequest struct {
	Model       string            `json:"model"`
	Messages    []*pb.ChatMessage `json:"messages"`
	Temperature float64           `json:"temperature"`
	Stream      bool              `json:"stream"`
}

// ChatCompletions handles POST /v1/chat/completions
func ChatCompletions(c *gin.Context) {
	var req ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Get UserID from context (set by JWTMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check X-Session-ID header
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-Session-ID header is required"})
		return
	}

	streamReq := &pb.ChatStreamRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		SessionId:   sessionID,
		UserId:      userID.(int64),
		Temperature: req.Temperature,
	}

	client := rpc.GetClientManager().ChatbotClient
	stream, err := client.ChatStream(context.Background(), streamReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call chatbot service"})
		return
	}

	// Set headers for SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	c.Stream(func(w io.Writer) bool {
		resp, err := stream.Recv()
		if err == io.EOF {
			// Send [DONE]
			c.SSEvent("", "[DONE]")
			return false
		}
		if err != nil {
			// Log error?
			return false
		}

		if resp.IsFinish {
			c.SSEvent("", "[DONE]")
			return false
		}

		// Format as OpenAI chunk
		// data: {"id":"...","object":"chat.completion.chunk","created":...,"model":"...",
		// "choices":[{"index":0,"delta":{"content":"..."},"finish_reason":null}]}

		chunk := gin.H{
			"id":      "chatcmpl-" + sessionID,
			"object":  "chat.completion.chunk",
			"created": 0, // TODO: Timestamp
			"model":   req.Model,
			"choices": []gin.H{
				{
					"index": 0,
					"delta": gin.H{
						"content": resp.ContentChunk,
					},
					"finish_reason": nil,
				},
			},
		}
		c.SSEvent("", chunk)
		return true
	})
}

// GetChatHistory handles GET /v1/chatbot/history/:session_id
func GetChatHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID required"})
		return
	}

	client := rpc.GetClientManager().ChatbotClient
	resp, err := client.GetChatHistory(context.Background(), &pb.GetChatHistoryRequest{SessionId: sessionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

type CreateConversationRequest struct {
	Title string `json:"title"`
}

// CreateConversation handles POST /v1/chatbot/newConversation
func CreateConversation(c *gin.Context) {
	var req CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Use default title if empty
	if req.Title == "" {
		req.Title = "New Chat"
	}

	client := rpc.GetClientManager().ChatbotClient
	resp, err := client.CreateConversation(context.Background(), &pb.CreateConversationRequest{
		UserId: userID.(int64),
		Title:  req.Title,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListConversations handles GET /v1/chatbot/history/list
func ListConversations(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	client := rpc.GetClientManager().ChatbotClient
	resp, err := client.ListConversations(context.Background(), &pb.ListConversationsRequest{UserId: userID.(int64)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
