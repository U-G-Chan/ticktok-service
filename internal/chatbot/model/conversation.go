package model

import "gorm.io/gorm"

type Conversation struct {
	gorm.Model
	SessionID string `gorm:"column:session_id;type:varchar(64);uniqueIndex;not null"`
	UserID    int64  `gorm:"column:user_id;index;not null"`
	Title     string `gorm:"column:title;type:varchar(255)"`
}

func (Conversation) TableName() string {
	return "conversations"
}
