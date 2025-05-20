package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UID       uint      `json:"uid" gorm:"not null"`
	Nickname  string    `json:"nickname" gorm:"size:100;not null"`
	Avatar    string    `json:"avatar" gorm:"size:255;not null"`
	Status    string    `json:"status" gorm:"type:enum('online','offline','away');default:'offline'"`
	LastSeen  string    `json:"lastSeen" gorm:"column:last_seen;size:100"`
	Signature string    `json:"signature" gorm:"size:255"`
	CreatedAt time.Time `json:"createdAt" gorm:"not null"`
}

// UserResponse 用户响应模型
type UserResponse struct {
	ID        uint   `json:"id"`
	UID       uint   `json:"uid"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	Status    string `json:"status"`
	LastSeen  string `json:"lastSeen"`
	Signature string `json:"signature"`
}

// FriendResponse 好友响应模型
type FriendResponse struct {
	ID         uint   `json:"id"`
	Name       string `json:"name"`
	Avatar     string `json:"avatar"`
	Online     bool   `json:"online"`
	IsOfficial bool   `json:"isOfficial"`
	LastActive string `json:"lastActive"`
	FriendType string `json:"friendType"`
} 