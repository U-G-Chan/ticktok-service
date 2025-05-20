package model

import (
	"time"
)

// Message 消息模型
type Message struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SessionID  string    `json:"sessionId" gorm:"column:session_id;size:50;not null"`
	SenderID   uint      `json:"senderId" gorm:"column:sender_id;not null"`
	ReceiverID uint      `json:"receiverId" gorm:"column:receiver_id;not null"`
	Type       string    `json:"type" gorm:"type:enum('text','voice','image');not null"`
	Content    string    `json:"content" gorm:"type:text"`
	Duration   string    `json:"duration,omitempty" gorm:"size:10"`
	Caption    string    `json:"caption,omitempty" gorm:"size:255"`
	Timestamp  int64     `json:"timestamp" gorm:"not null"`
	Status     string    `json:"status" gorm:"type:enum('sending','sent','read','failed');default:'sending'"`
	CreatedAt  time.Time `json:"createdAt" gorm:"not null"`
	Sender     User      `json:"-" gorm:"foreignKey:SenderID"`
	Receiver   User      `json:"-" gorm:"foreignKey:ReceiverID"`
}

// MessageRequest 发送消息请求
type MessageRequest struct {
	SenderID   uint   `json:"senderId" binding:"required"`
	ReceiverID uint   `json:"receiverId" binding:"required"`
	IsSelf     bool   `json:"isSelf" binding:"required"`
	Type       string `json:"type" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Timestamp  int64  `json:"timestamp" binding:"required"`
	Duration   string `json:"duration,omitempty"`
	Caption    string `json:"caption,omitempty"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID         uint   `json:"id"`
	SenderID   uint   `json:"senderId"`
	ReceiverID uint   `json:"receiverId"`
	IsSelf     bool   `json:"isSelf"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
	Status     string `json:"status"`
	SessionID  string `json:"sessionId"`
	Duration   string `json:"duration,omitempty"`
	Caption    string `json:"caption,omitempty"`
}

// ChatMessage 聊天消息响应
type ChatMessage struct {
	ID         uint   `json:"id"`
	SenderID   uint   `json:"senderId"`
	ReceiverID uint   `json:"receiverId"`
	IsSelf     bool   `json:"isSelf"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Timestamp  int64  `json:"timestamp"`
	Status     string `json:"status"`
	SessionID  string `json:"sessionId"`
	Duration   string `json:"duration,omitempty"`
	Caption    string `json:"caption,omitempty"`
}

// MessageListResponse 消息列表项
type MessageListResponse struct {
	ID     uint          `json:"id"`
	Sender FriendResponse `json:"sender"`
	Text   string        `json:"text"`
	Time   string        `json:"time"`
	Unread int           `json:"unread"`
} 