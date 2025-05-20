package service

import (
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"

	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	GetUserByID(id uint) (*model.UserResponse, error)
	GetUsersByIDs(ids []uint) ([]*model.UserResponse, error)
}

// userService 用户服务实现
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(db *gorm.DB) UserService {
	return &userService{
		userRepo: repository.NewUserRepository(db),
	}
}

// GetUserByID 根据ID获取用户信息
func (s *userService) GetUserByID(id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetUserByID(id)
	if err != nil {
		return nil, err
	}
	
	return &model.UserResponse{
		ID:        user.ID,
		UID:       user.UID,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		Status:    user.Status,
		LastSeen:  user.LastSeen,
		Signature: user.Signature,
	}, nil
}

// GetUsersByIDs 批量获取用户信息
func (s *userService) GetUsersByIDs(ids []uint) ([]*model.UserResponse, error) {
	users, err := s.userRepo.GetUsersByIDs(ids)
	if err != nil {
		return nil, err
	}
	
	var userResponses []*model.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, &model.UserResponse{
			ID:        user.ID,
			UID:       user.UID,
			Nickname:  user.Nickname,
			Avatar:    user.Avatar,
			Status:    user.Status,
			LastSeen:  user.LastSeen,
			Signature: user.Signature,
		})
	}
	
	return userResponses, nil
} 