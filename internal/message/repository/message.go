package repository

import (
	"context"
	"ticktok-service/internal/message/model"
	"ticktok-service/pkg/mysql"
)

// CreateMessage inserts a message into the database
func CreateMessage(ctx context.Context, msg *model.Message) error {
	return mysql.DB.WithContext(ctx).Create(msg).Error
}

// GetMessagesByIDs retrieves multiple messages by their IDs
func GetMessagesByIDs(ctx context.Context, ids []int64) ([]*model.Message, error) {
	var messages []*model.Message
	err := mysql.DB.WithContext(ctx).Where("id IN ?", ids).Find(&messages).Error
	return messages, err
}
