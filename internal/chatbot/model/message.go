package model

import "gorm.io/gorm"

type ChatMessage struct {
	gorm.Model
	SessionID string `gorm:"column:session_id;type:varchar(64);uniqueIndex:idx_session_msg;not null"`
	MessageID int64  `gorm:"column:message_id;uniqueIndex:idx_session_msg;not null"` // Snowflake ID
	Role      string `gorm:"column:role;type:varchar(20);not null"`
	Content   string `gorm:"column:content;type:text;not null"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}