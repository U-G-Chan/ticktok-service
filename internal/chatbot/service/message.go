package service

import (
	"context"
	"ticktok-service/internal/chatbot/model"
	"ticktok-service/internal/chatbot/repository"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/mysql"

	"github.com/bwmarrin/snowflake"
)

type MessageService struct {
	messageRepo *repository.MessageRepo
	snowflake   *snowflake.Node
}

func NewMessageService() *MessageService {
	node, err := snowflake.NewNode(1)
	if err != nil {
		logger.Log.Fatal("failed to initialize snowflake node: " + err.Error())
	}
	return &MessageService{
		messageRepo: repository.NewMessageRepo(mysql.DB),
		snowflake:   node,
	}
}

// GetChatHistory implements retrieving chat history for a session
func (s *MessageService) GetChatHistory(ctx context.Context, sessionID string) ([]*model.ChatMessage, error) {
	return s.messageRepo.GetBySessionID(sessionID)
}

// AddMessage adds a new message to the database
func (s *MessageService) AddMessage(ctx context.Context, sessionID string, role string, content string) error {
	msgID := s.snowflake.Generate().Int64()
	return s.messageRepo.Create(&model.ChatMessage{
		SessionID: sessionID,
		MessageID: msgID,
		Role:      role,
		Content:   content,
	})
}

// SaveMessage is an alias for AddMessage, used for compatibility or semantic clarity
func (s *MessageService) SaveMessage(ctx context.Context, sessionID string, role string, content string) error {
	return s.AddMessage(ctx, sessionID, role, content)
}
