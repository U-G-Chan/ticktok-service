package repository

import (
	"ticktok-service/internal/model"

	"gorm.io/gorm"
)

// UserRepository 用户数据仓库接口
type UserRepository interface {
	GetUserByID(id uint) (*model.User, error)
	GetUsersByIDs(ids []uint) ([]*model.User, error)
}

// userRepository 用户数据仓库实现
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository 创建用户数据仓库
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// GetUserByID 根据ID获取用户
func (r *userRepository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersByIDs 批量获取用户
func (r *userRepository) GetUsersByIDs(ids []uint) ([]*model.User, error) {
	var users []*model.User
	if err := r.db.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
} 