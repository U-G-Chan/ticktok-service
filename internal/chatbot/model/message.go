package model

import "time"

type ChatMessage struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	SessionID string    `gorm:"column:session_id;type:varchar(64);uniqueIndex:idx_session_msg;not null"`
	MessageID int64     `gorm:"column:message_id;uniqueIndex:idx_session_msg;not null"` // Snowflake ID
	Role      string    `gorm:"column:role;type:varchar(20);not null"`
	Content   string    `gorm:"column:content;type:text;not null"`
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}
