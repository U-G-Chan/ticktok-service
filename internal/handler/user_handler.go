package handler

import (
	"strconv"
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler 用户相关处理器
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建新的用户处理器
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		userService: service.NewUserService(db),
	}
}

// GetUser 获取用户信息
func (h *UserHandler) GetUser(c *gin.Context) {
	// 解析路径参数
	userIDStr := c.Param("userId")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的用户ID")
		return
	}

	// 获取用户信息
	user, err := h.userService.GetUserByID(uint(userID))
	if err != nil {
		util.Fail(c, 404, "用户不存在: "+err.Error())
		return
	}

	util.Success(c, user)
}

// GetUsersBatch 批量获取用户信息
func (h *UserHandler) GetUsersBatch(c *gin.Context) {
	// 解析请求参数
	var req struct {
		UserIDs []uint `json:"userIds" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, 400, "无效的请求参数: "+err.Error())
		return
	}

	// 获取用户信息
	users, err := h.userService.GetUsersByIDs(req.UserIDs)
	if err != nil {
		util.Fail(c, 500, "获取用户信息失败: "+err.Error())
		return
	}

	util.Success(c, users)
} 