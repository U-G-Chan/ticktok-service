package model

import (
	"time"
)

// Session 消息会话模型
type Session struct {
	ID            string    `json:"id" gorm:"primaryKey;size:50"`
	User1ID       uint      `json:"user1Id" gorm:"column:user1_id;not null"`
	User2ID       uint      `json:"user2Id" gorm:"column:user2_id;not null"`
	LastMessageID uint      `json:"lastMessageId" gorm:"column:last_message_id"`
	UpdatedAt     time.Time `json:"updatedAt" gorm:"not null"`
	User1         User      `json:"-" gorm:"foreignKey:User1ID"`
	User2         User      `json:"-" gorm:"foreignKey:User2ID"`
	LastMessage   Message   `json:"-" gorm:"foreignKey:LastMessageID"`
}

// UnreadMessage 未读消息模型
type UnreadMessage struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"userId" gorm:"column:user_id;not null"`
	SenderID  uint      `json:"senderId" gorm:"column:sender_id;not null"`
	Count     int       `json:"count" gorm:"default:0"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"not null"`
	User      User      `json:"-" gorm:"foreignKey:UserID"`
	Sender    User      `json:"-" gorm:"foreignKey:SenderID"`
} 