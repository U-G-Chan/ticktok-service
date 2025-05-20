package model

import (
	"time"
)

// Friendship 好友关系模型
type Friendship struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	UserID     uint      `json:"userId" gorm:"column:user_id;not null"`
	FriendID   uint      `json:"friendId" gorm:"column:friend_id;not null"`
	FriendType string    `json:"friendType" gorm:"column:friend_type;type:enum('normal','aibot','system');default:'normal'"`
	CreatedAt  time.Time `json:"createdAt" gorm:"not null"`
	User       User      `json:"-" gorm:"foreignKey:UserID"`
	Friend     User      `json:"-" gorm:"foreignKey:FriendID"`
} 