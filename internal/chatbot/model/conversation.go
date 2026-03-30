package model

import "time"

type Conversation struct {
	ID        int64     `gorm:"primaryKey;autoIncrement:false" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	SessionID string    `gorm:"column:session_id;type:varchar(64);uniqueIndex;not null"`
	UserID    int64     `gorm:"column:user_id;index;not null"`
	Title     string    `gorm:"column:title;type:varchar(255)"`
}

func (Conversation) TableName() string {
	return "conversations"
}
