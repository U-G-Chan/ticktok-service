package service

import (
	"context"
	"ticktok-service/internal/chatbot/model"
	"ticktok-service/internal/chatbot/repository"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/snowflake"
)

type MessageService struct {
	messageRepo *repository.MessageRepo
}

func NewMessageService() *MessageService {
	return &MessageService{
		messageRepo: repository.NewMessageRepo(mysql.DB),
	}
}

// GetChatHistory implements retrieving chat history for a session
func (s *MessageService) GetChatHistory(ctx context.Context, sessionID string) ([]*model.ChatMessage, error) {
	return s.messageRepo.GetBySessionID(sessionID)
}

// AddMessage adds a new message to the database
func (s *MessageService) AddMessage(ctx context.Context, sessionID string, role string, content string) error {
	msgID := snowflake.GenerateMsgID()
	return s.messageRepo.Create(&model.ChatMessage{
		ID:        snowflake.GenerateMsgID(),
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
