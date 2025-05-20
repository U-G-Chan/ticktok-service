package service

import (
	"fmt"
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"
	"time"

	"gorm.io/gorm"
)

// MessageService 消息服务接口
type MessageService interface {
	GetMessageList(userID uint) ([]*model.MessageListResponse, error)
	GetChatHistory(userID, friendID uint) ([]*model.ChatMessage, error)
	SendMessage(req *model.MessageRequest) (*model.MessageResponse, error)
	MarkAsRead(userID, friendID uint) (bool, error)
}

// messageService 消息服务实现
type messageService struct {
	messageRepo repository.MessageRepository
	userRepo    repository.UserRepository
	friendRepo  repository.FriendRepository
}

// NewMessageService 创建消息服务
func NewMessageService(db *gorm.DB) MessageService {
	return &messageService{
		messageRepo: repository.NewMessageRepository(db),
		userRepo:    repository.NewUserRepository(db),
		friendRepo:  repository.NewFriendRepository(db),
	}
}

// GetMessageList 获取消息列表
func (s *messageService) GetMessageList(userID uint) ([]*model.MessageListResponse, error) {
	// 获取用户的最近消息
	messages, err := s.messageRepo.GetLastMessages(userID)
	if err != nil {
		return nil, err
	}
	
	var messageResponses []*model.MessageListResponse
	for _, message := range messages {
		// 确定发送者
		var senderID uint
		if message.SenderID == userID {
			senderID = message.ReceiverID
		} else {
			senderID = message.SenderID
		}
		
		// 获取发送者信息
		sender, err := s.userRepo.GetUserByID(senderID)
		if err != nil {
			continue
		}
		
		// 格式化时间
		timeStr := formatTime(message.Timestamp)
		
		// 获取未读消息数量
		unreadCount, err := s.messageRepo.GetUnreadCount(userID, senderID)
		if err != nil {
			unreadCount = 0
		}
		
		// 检查是否是官方账号
		isOfficial := false // 默认不是官方账号，实际应用中可以通过数据库判断
		
		// 创建响应
		messageResponses = append(messageResponses, &model.MessageListResponse{
			ID: message.ID,
			Sender: model.FriendResponse{
				ID:         sender.ID,
				Name:       sender.Nickname,
				Avatar:     sender.Avatar,
				Online:     sender.Status == "online",
				IsOfficial: isOfficial,
				LastActive: sender.LastSeen,
				FriendType: "normal", // 默认值，实际应用中应该从数据库获取
			},
			Text:   message.Content,
			Time:   timeStr,
			Unread: unreadCount,
		})
	}
	
	return messageResponses, nil
}

// GetChatHistory 获取聊天历史记录
func (s *messageService) GetChatHistory(userID, friendID uint) ([]*model.ChatMessage, error) {
	// 获取聊天记录
	messages, err := s.messageRepo.GetChatHistory(userID, friendID)
	if err != nil {
		return nil, err
	}
	
	var chatMessages []*model.ChatMessage
	for _, message := range messages {
		isSelf := message.SenderID == userID
		
		chatMessages = append(chatMessages, &model.ChatMessage{
			ID:         message.ID,
			SenderID:   message.SenderID,
			ReceiverID: message.ReceiverID,
			IsSelf:     isSelf,
			Type:       message.Type,
			Content:    message.Content,
			Timestamp:  message.Timestamp,
			Status:     message.Status,
			SessionID:  message.SessionID,
			Duration:   message.Duration,
			Caption:    message.Caption,
		})
	}
	
	return chatMessages, nil
}

// SendMessage 发送消息
func (s *messageService) SendMessage(req *model.MessageRequest) (*model.MessageResponse, error) {
	// 创建消息记录
	sessionID := fmt.Sprintf("chat_%d_%d", req.SenderID, req.ReceiverID)
	message := &model.Message{
		SessionID:  sessionID,
		SenderID:   req.SenderID,
		ReceiverID: req.ReceiverID,
		Type:       req.Type,
		Content:    req.Content,
		Timestamp:  req.Timestamp,
		Duration:   req.Duration,
		Caption:    req.Caption,
		Status:     "sent",
		CreatedAt:  time.Now(),
	}
	
	// 保存消息
	if err := s.messageRepo.CreateMessage(message); err != nil {
		return nil, err
	}
	
	// 返回响应
	return &model.MessageResponse{
		ID:         message.ID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		IsSelf:     req.IsSelf,
		Type:       message.Type,
		Content:    message.Content,
		Timestamp:  message.Timestamp,
		Status:     message.Status,
		SessionID:  message.SessionID,
		Duration:   message.Duration,
		Caption:    message.Caption,
	}, nil
}

// MarkAsRead 标记消息为已读
func (s *messageService) MarkAsRead(userID, friendID uint) (bool, error) {
	err := s.messageRepo.MarkMessagesAsRead(userID, friendID)
	if err != nil {
		return false, err
	}
	return true, nil
}

// formatTime 格式化时间
func formatTime(timestamp int64) string {
	t := time.Unix(timestamp/1000, 0)
	now := time.Now()
	
	// 如果是今天，返回小时:分钟
	if t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day() {
		return t.Format("15:04")
	}
	
	// 如果是今年，返回月-日
	if t.Year() == now.Year() {
		return t.Format("01-02")
	}
	
	// 否则返回年-月-日
	return t.Format("2006-01-02")
} 