package repository

import (
	"ticktok-service/internal/chatbot/model"
	"ticktok-service/pkg/common"

	"gorm.io/gorm"
)

type MessageRepo struct {
	*common.BaseRepo[model.ChatMessage]
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{
		BaseRepo: common.NewBaseRepo[model.ChatMessage](db),
	}
}

func (r *MessageRepo) GetBySessionID(sessionID string) ([]*model.ChatMessage, error) {
	var messages []*model.ChatMessage
	err := r.DB.Where("session_id = ?", sessionID).Order("message_id asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepo) Create(message *model.ChatMessage) error {
	return r.DB.Create(message).Error
}

