package repository

import (
	"ticktok-service/internal/chatbot/model"
	"ticktok-service/pkg/common"
	"ticktok-service/pkg/snowflake"

	"gorm.io/gorm"
)

type ConversationRepo struct {
	*common.BaseRepo[model.Conversation]
}

func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{
		BaseRepo: common.NewBaseRepo[model.Conversation](db),
	}
}

func (r *ConversationRepo) CreateConversation(sessionID string, userID int64, title string) error {
	return r.Create(&model.Conversation{
		ID:        snowflake.GenerateMsgID(),
		SessionID: sessionID,
		UserID:    userID,
		Title:     title,
	})
}

func (r *ConversationRepo) ListConversations(userID int64) ([]*model.Conversation, error) {
	var conversations []*model.Conversation
	err := r.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&conversations).Error
	return conversations, err
}

func (r *ConversationRepo) GetConversation(sessionID string) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.DB.Where("session_id = ?", sessionID).First(&conversation).Error
	return &conversation, err
}
