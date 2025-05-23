package handler

import (
	"encoding/json"
	"net/http"
	"ticktok-service/internal/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PublishHandler 发布相关处理器
type PublishHandler struct {
	db *gorm.DB
}

// NewPublishHandler 创建新的发布处理器
func NewPublishHandler(db *gorm.DB) *PublishHandler {
	return &PublishHandler{
		db: db,
	}
}

// SaveDraft 保存草稿
func (h *PublishHandler) SaveDraft(c *gin.Context) {
	// 开发模式：使用固定用户ID
	userID := uint(1)
	
	// 解析请求参数
	var req model.PublishMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.SaveDraftResponse{
			Success: false,
		})
		return
	}
	
	// 生成草稿ID
	draftID := uuid.New().String()
	
	// 序列化JSON字段
	mediaItemsJSON, _ := json.Marshal(req.MediaItems)
	topicsJSON, _ := json.Marshal(req.Topics)
	mentionsJSON, _ := json.Marshal(req.Mentions)
	tagsJSON, _ := json.Marshal(req.Tags)
	
	// 处理时间
	now := time.Now()
	createdAt := now
	if req.CreatedAt > 0 {
		createdAt = time.Unix(0, req.CreatedAt*int64(time.Millisecond))
	}
	
	// 创建草稿
	draft := model.Draft{
		ID:          draftID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		MediaItems:  string(mediaItemsJSON),
		Topics:      string(topicsJSON),
		Mentions:    string(mentionsJSON),
		Tags:        string(tagsJSON),
		Visibility:  req.Visibility,
		IsDaily:     req.IsDaily,
		CreatedAt:   createdAt,
		UpdatedAt:   now,
	}
	
	// 保存到数据库
	if err := h.db.Create(&draft).Error; err != nil {
		c.JSON(http.StatusOK, model.SaveDraftResponse{
			Success: false,
		})
		return
	}
	
	// 返回成功响应
	c.JSON(http.StatusOK, model.SaveDraftResponse{
		Success: true,
		DraftID: draftID,
	})
}

// PublishContent 发布内容
func (h *PublishHandler) PublishContent(c *gin.Context) {
	// 开发模式：使用固定用户ID
	userID := uint(1)
	
	// 解析请求参数
	var req model.PublishMessage
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, model.PublishContentResponse{
			Success: false,
		})
		return
	}
	
	// 生成发布ID
	publishID := uuid.New().String()
	
	// 序列化JSON字段
	mediaItemsJSON, _ := json.Marshal(req.MediaItems)
	topicsJSON, _ := json.Marshal(req.Topics)
	mentionsJSON, _ := json.Marshal(req.Mentions)
	tagsJSON, _ := json.Marshal(req.Tags)
	
	// 处理时间
	now := time.Now()
	createdAt := now
	if req.CreatedAt > 0 {
		createdAt = time.Unix(0, req.CreatedAt*int64(time.Millisecond))
	}
	
	// 创建发布内容
	content := model.Content{
		ID:          publishID,
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		MediaItems:  string(mediaItemsJSON),
		Topics:      string(topicsJSON),
		Mentions:    string(mentionsJSON),
		Tags:        string(tagsJSON),
		Visibility:  req.Visibility,
		IsDaily:     req.IsDaily,
		ViewCount:   0,
		LikeCount:   0,
		CommentCount: 0,
		ShareCount:  0,
		CreatedAt:   createdAt,
		UpdatedAt:   now,
	}
	
	// 保存到数据库
	if err := h.db.Create(&content).Error; err != nil {
		c.JSON(http.StatusOK, model.PublishContentResponse{
			Success: false,
		})
		return
	}
	
	// 返回成功响应
	c.JSON(http.StatusOK, model.PublishContentResponse{
		Success:   true,
		PublishID: publishID,
	})
}

// GetTopics 获取话题列表
func (h *PublishHandler) GetTopics(c *gin.Context) {
	// 预定义话题列表
	topics := []string{
		"记录生活",
		"旅行足迹", 
		"世界那么大我想去",
		"美食分享",
		"宠物日常",
		"学习笔记",
		"工作日常",
		"健身打卡",
		"摄影分享",
		"音乐推荐",
		"电影评论",
		"读书心得",
		"技术分享",
		"创意设计",
		"时尚穿搭",
	}
	
	// 返回话题列表
	c.JSON(http.StatusOK, model.TopicsResponse{
		Topics: topics,
	})
} 