package service

import (
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"

	"gorm.io/gorm"
)

// FriendService 好友服务接口
type FriendService interface {
	GetFriendsByUserID(userID uint) ([]*model.FriendResponse, error)
}

// friendService 好友服务实现
type friendService struct {
	friendRepo repository.FriendRepository
}

// NewFriendService 创建好友服务
func NewFriendService(db *gorm.DB) FriendService {
	return &friendService{
		friendRepo: repository.NewFriendRepository(db),
	}
}

// GetFriendsByUserID 获取用户的好友列表
func (s *friendService) GetFriendsByUserID(userID uint) ([]*model.FriendResponse, error) {
	friendships, err := s.friendRepo.GetFriendsByUserID(userID)
	if err != nil {
		return nil, err
	}
	
	var friendResponses []*model.FriendResponse
	for _, friendship := range friendships {
		friend := friendship.Friend
		isOnline := friend.Status == "online"
		isOfficial := friendship.FriendType == "aibot" || friendship.FriendType == "system"
		
		friendResponses = append(friendResponses, &model.FriendResponse{
			ID:         friend.ID,
			Name:       friend.Nickname,
			Avatar:     friend.Avatar,
			Online:     isOnline,
			IsOfficial: isOfficial,
			LastActive: friend.LastSeen,
			FriendType: friendship.FriendType,
		})
	}
	
	return friendResponses, nil
} 