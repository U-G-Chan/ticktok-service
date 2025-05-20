package handler

import (
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FriendHandler 好友相关处理器
type FriendHandler struct {
	friendService service.FriendService
}

// NewFriendHandler 创建新的好友处理器
func NewFriendHandler(db *gorm.DB) *FriendHandler {
	return &FriendHandler{
		friendService: service.NewFriendService(db),
	}
}

// GetFriends 获取好友列表
func (h *FriendHandler) GetFriends(c *gin.Context) {
	// 这里简化处理，实际应用中应该从JWT或会话中获取当前用户ID
	userID := uint(1) // 假设当前用户ID为1

	// 获取好友列表
	friends, err := h.friendService.GetFriendsByUserID(userID)
	if err != nil {
		util.Fail(c, 500, "获取好友列表失败: "+err.Error())
		return
	}

	util.Success(c, friends)
} 