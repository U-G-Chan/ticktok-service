package handler

import (
	"strconv"
	"ticktok-service/internal/model"
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MessageHandler 消息相关处理器
type MessageHandler struct {
	messageService service.MessageService
}

// NewMessageHandler 创建新的消息处理器
func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{
		messageService: service.NewMessageService(db),
	}
}

// GetMessages 获取消息列表
func (h *MessageHandler) GetMessages(c *gin.Context) {
	// 这里简化处理，实际应用中应该从JWT或会话中获取当前用户ID
	userID := uint(1) // 假设当前用户ID为1

	// 获取消息列表
	messages, err := h.messageService.GetMessageList(userID)
	if err != nil {
		util.Fail(c, 500, "获取消息列表失败: "+err.Error())
		return
	}

	util.Success(c, messages)
}

// GetChatHistory 获取聊天历史记录
func (h *MessageHandler) GetChatHistory(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的用户ID")
		return
	}

	// 获取当前用户ID（查询参数）
	currentUserIDStr := c.DefaultQuery("currentUserId", "1") // 默认为1
	currentUserID, err := strconv.ParseUint(currentUserIDStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的当前用户ID")
		return
	}

	// 获取聊天历史记录
	messages, err := h.messageService.GetChatHistory(uint(currentUserID), uint(userID))
	if err != nil {
		util.Fail(c, 500, "获取聊天历史记录失败: "+err.Error())
		return
	}

	util.Success(c, messages)
}

// SendMessage 发送消息
func (h *MessageHandler) SendMessage(c *gin.Context) {
	// 解析请求参数
	var req model.MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, 400, "无效的请求参数: "+err.Error())
		return
	}

	// 发送消息
	message, err := h.messageService.SendMessage(&req)
	if err != nil {
		util.Fail(c, 500, "发送消息失败: "+err.Error())
		return
	}

	util.Success(c, message)
}

// MarkAsRead 标记消息已读
func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的用户ID")
		return
	}

	// 获取当前用户ID（查询参数）
	currentUserIDStr := c.DefaultQuery("currentUserId", "1") // 默认为1
	currentUserID, err := strconv.ParseUint(currentUserIDStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的当前用户ID")
		return
	}

	// 标记消息为已读
	success, err := h.messageService.MarkAsRead(uint(currentUserID), uint(userID))
	if err != nil {
		util.Fail(c, 500, "标记消息已读失败: "+err.Error())
		return
	}

	util.Success(c, success)
} 