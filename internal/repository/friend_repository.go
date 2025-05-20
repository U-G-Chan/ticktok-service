package repository

import (
	"ticktok-service/internal/model"

	"gorm.io/gorm"
)

// FriendRepository 好友数据仓库接口
type FriendRepository interface {
	GetFriendsByUserID(userID uint) ([]*model.Friendship, error)
}

// friendRepository 好友数据仓库实现
type friendRepository struct {
	db *gorm.DB
}

// NewFriendRepository 创建好友数据仓库
func NewFriendRepository(db *gorm.DB) FriendRepository {
	return &friendRepository{
		db: db,
	}
}

// GetFriendsByUserID 获取用户的好友列表
func (r *friendRepository) GetFriendsByUserID(userID uint) ([]*model.Friendship, error) {
	var friendships []*model.Friendship
	if err := r.db.Preload("Friend").Where("user_id = ?", userID).Find(&friendships).Error; err != nil {
		return nil, err
	}
	return friendships, nil
} 