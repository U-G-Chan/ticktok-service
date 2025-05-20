package repository

import (
	"fmt"
	"ticktok-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// MessageRepository 消息数据仓库接口
type MessageRepository interface {
	GetLastMessages(userID uint) ([]*model.Message, error)
	GetChatHistory(userID, friendID uint) ([]*model.Message, error)
	CreateMessage(message *model.Message) error
	MarkMessagesAsRead(userID, friendID uint) error
	GetUnreadCount(userID, senderID uint) (int, error)
}

// messageRepository 消息数据仓库实现
type messageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 创建消息数据仓库
func NewMessageRepository(db *gorm.DB) MessageRepository {
	return &messageRepository{
		db: db,
	}
}

// GetLastMessages 获取最近消息列表
func (r *messageRepository) GetLastMessages(userID uint) ([]*model.Message, error) {
	var messages []*model.Message
	// 获取用户参与的所有会话的最后一条消息
	query := `
		SELECT m.* FROM messages m
		INNER JOIN (
			SELECT MAX(id) as id, session_id
			FROM messages
			WHERE sender_id = ? OR receiver_id = ?
			GROUP BY session_id
		) AS latest ON latest.id = m.id
		ORDER BY m.timestamp DESC
	`
	if err := r.db.Raw(query, userID, userID).Preload("Sender").Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// GetChatHistory 获取聊天历史记录
func (r *messageRepository) GetChatHistory(userID, friendID uint) ([]*model.Message, error) {
	var messages []*model.Message
	sessionID1 := fmt.Sprintf("chat_%d_%d", userID, friendID)
	sessionID2 := fmt.Sprintf("chat_%d_%d", friendID, userID)
	
	if err := r.db.Where("(session_id = ? OR session_id = ?)", sessionID1, sessionID2).
		Order("timestamp ASC").
		Find(&messages).Error; err != nil {
		return nil, err
	}
	return messages, nil
}

// CreateMessage 创建消息
func (r *messageRepository) CreateMessage(message *model.Message) error {
	message.CreatedAt = time.Now()
	// 如果未设置会话ID，则自动生成
	if message.SessionID == "" {
		message.SessionID = fmt.Sprintf("chat_%d_%d", message.SenderID, message.ReceiverID)
	}
	
	tx := r.db.Begin()
	
	// 插入消息
	if err := tx.Create(message).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// 更新或创建会话
	sessionID := message.SessionID
	var session model.Session
	if err := tx.Where("id = ?", sessionID).First(&session).Error; err != nil {
		// 会话不存在，创建新会话
		session = model.Session{
			ID:            sessionID,
			User1ID:       message.SenderID,
			User2ID:       message.ReceiverID,
			LastMessageID: message.ID,
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(&session).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 会话已存在，更新最后消息ID
		if err := tx.Model(&session).Updates(map[string]interface{}{
			"last_message_id": message.ID,
			"updated_at":      time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	
	// 更新未读消息计数
	var unreadMessage model.UnreadMessage
	result := tx.Where("user_id = ? AND sender_id = ?", message.ReceiverID, message.SenderID).First(&unreadMessage)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		tx.Rollback()
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		// 创建新的未读记录
		unreadMessage = model.UnreadMessage{
			UserID:    message.ReceiverID,
			SenderID:  message.SenderID,
			Count:     1,
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(&unreadMessage).Error; err != nil {
			tx.Rollback()
			return err
		}
	} else {
		// 更新未读计数
		if err := tx.Model(&unreadMessage).Updates(map[string]interface{}{
			"count":      unreadMessage.Count + 1,
			"updated_at": time.Now(),
		}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	
	return tx.Commit().Error
}

// MarkMessagesAsRead 标记消息为已读
func (r *messageRepository) MarkMessagesAsRead(userID, friendID uint) error {
	tx := r.db.Begin()
	
	// 更新消息状态
	sessionID1 := fmt.Sprintf("chat_%d_%d", friendID, userID)
	sessionID2 := fmt.Sprintf("chat_%d_%d", userID, friendID)
	
	if err := tx.Model(&model.Message{}).
		Where("(session_id = ? OR session_id = ?) AND sender_id = ? AND status <> ?", 
			sessionID1, sessionID2, friendID, "read").
		Update("status", "read").Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// 重置未读计数
	if err := tx.Where("user_id = ? AND sender_id = ?", userID, friendID).
		Delete(&model.UnreadMessage{}).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	return tx.Commit().Error
}

// GetUnreadCount 获取未读消息数量
func (r *messageRepository) GetUnreadCount(userID, senderID uint) (int, error) {
	var unreadMessage model.UnreadMessage
	result := r.db.Select("count").Where("user_id = ? AND sender_id = ?", userID, senderID).First(&unreadMessage)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, result.Error
	}
	return unreadMessage.Count, nil
} 