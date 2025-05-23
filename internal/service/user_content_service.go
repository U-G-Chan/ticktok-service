package service

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"
	"time"
)

// UserContentService 用户内容服务接口
type UserContentService interface {
	// 内容管理
	CreateContent(params *model.CreateContentParams) (*model.ContentItem, error)
	UpdateContent(params *model.UpdateContentParams) (*model.ContentItem, error)
	DeleteContent(itemID, userID string) error
	GetContentList(params *model.ContentListParams) (*model.ContentListResult, error)
	
	// 点赞管理
	ToggleLike(params *model.ToggleLikeParams) (*model.ToggleLikeResult, error)
}

// userContentService 用户内容服务实现
type userContentService struct {
	repo repository.UserContentRepository
}

// NewUserContentService 创建用户内容服务实例
func NewUserContentService(repo repository.UserContentRepository) UserContentService {
	return &userContentService{repo: repo}
}

// CreateContent 创建内容项
func (s *userContentService) CreateContent(params *model.CreateContentParams) (*model.ContentItem, error) {
	var itemID string
	var err error
	
	// 如果前端传入了itemID，则使用传入的；否则自动生成
	if params.ItemID != "" {
		itemID = params.ItemID
		// 检查itemID是否已存在
		if _, err := s.repo.GetByItemID(itemID); err == nil {
			return nil, fmt.Errorf("ItemID已存在: %s", itemID)
		}
	} else {
		// 生成唯一的ItemID
		itemID, err = s.generateItemID(params.ListType)
		if err != nil {
			return nil, fmt.Errorf("生成ItemID失败: %v", err)
		}
	}
	
	// 设置默认值
	if params.WorkType == "" {
		params.WorkType = model.WorkTypePublished
	}
	
	isPublic := true
	if params.IsPublic != nil {
		isPublic = *params.IsPublic
	}
	
	// 设置点赞数，如果前端传入了则使用传入的值，否则默认为0
	likes := params.Likes
	if likes < 0 {
		likes = 0 // 确保点赞数不为负数
	}
	
	// 序列化标签
	tagsJSON := "[]"
	if params.Tags != nil && len(params.Tags) > 0 {
		tagsBytes, err := json.Marshal(params.Tags)
		if err == nil {
			tagsJSON = string(tagsBytes)
		}
	}
	
	// 序列化Other字段
	otherJSON := "{}"
	if params.Other != nil {
		otherBytes, err := json.Marshal(params.Other)
		if err == nil {
			otherJSON = string(otherBytes)
		}
	}
	
	// 创建数据库模型
	now := time.Now()
	content := &model.UserContent{
		ItemID:      itemID,
		UserID:      params.UserID,
		ListType:    params.ListType,
		Thumbnail:   params.Thumbnail,
		Likes:       likes,
		WorkType:    params.WorkType,
		Title:       params.Title,
		Description: params.Description,
		PublishTime: &now,
		Duration:    params.Duration,
		Tags:        tagsJSON,
		IsPublic:    isPublic,
		Other:       otherJSON,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	
	// 保存到数据库
	if err := s.repo.Create(content); err != nil {
		return nil, fmt.Errorf("创建内容失败: %v", err)
	}
	
	// 转换为API模型返回
	return repository.ConvertToContentItem(content)
}

// UpdateContent 更新内容项
func (s *userContentService) UpdateContent(params *model.UpdateContentParams) (*model.ContentItem, error) {
	// 获取现有内容
	content, err := s.repo.GetByItemID(params.ItemID)
	if err != nil {
		return nil, fmt.Errorf("内容不存在: %v", err)
	}
	
	// 检查权限
	if content.UserID != params.UserID {
		return nil, fmt.Errorf("无权限更新此内容")
	}
	
	// 更新字段
	if params.Title != "" {
		content.Title = params.Title
	}
	if params.Description != "" {
		content.Description = params.Description
	}
	if params.Thumbnail != "" {
		content.Thumbnail = params.Thumbnail
	}
	if params.WorkType != "" {
		content.WorkType = params.WorkType
	}
	if params.IsPublic != nil {
		content.IsPublic = *params.IsPublic
	}
	
	// 更新标签
	if params.Tags != nil {
		tagsBytes, err := json.Marshal(params.Tags)
		if err == nil {
			content.Tags = string(tagsBytes)
		}
	}
	
	// 更新Other字段
	if params.Other != nil {
		// 先解析现有的Other字段
		var existingOther map[string]interface{}
		if content.Other != "" {
			json.Unmarshal([]byte(content.Other), &existingOther)
		}
		if existingOther == nil {
			existingOther = make(map[string]interface{})
		}
		
		// 合并新的Other字段
		for k, v := range params.Other {
			existingOther[k] = v
		}
		
		otherBytes, err := json.Marshal(existingOther)
		if err == nil {
			content.Other = string(otherBytes)
		}
	}
	
	content.UpdatedAt = time.Now()
	
	// 保存到数据库
	if err := s.repo.Update(content); err != nil {
		return nil, fmt.Errorf("更新内容失败: %v", err)
	}
	
	// 转换为API模型返回
	return repository.ConvertToContentItem(content)
}

// DeleteContent 删除内容项
func (s *userContentService) DeleteContent(itemID, userID string) error {
	return s.repo.Delete(itemID, userID)
}

// GetContentList 获取内容列表
func (s *userContentService) GetContentList(params *model.ContentListParams) (*model.ContentListResult, error) {
	// 设置默认分页参数
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 10
	}
	if params.PageSize > 100 {
		params.PageSize = 100
	}
	
	// 从仓库获取数据
	contents, total, err := s.repo.GetList(params)
	if err != nil {
		return nil, fmt.Errorf("获取内容列表失败: %v", err)
	}
	
	// 转换为API模型
	items := make([]model.ContentItem, 0, len(contents))
	for _, content := range contents {
		item, err := repository.ConvertToContentItem(&content)
		if err == nil && item != nil {
			items = append(items, *item)
		}
	}
	
	// 计算是否还有更多数据
	hasMore := int64(params.Page*params.PageSize) < total
	
	return &model.ContentListResult{
		Items:    items,
		Total:    total,
		Page:     params.Page,
		PageSize: params.PageSize,
		HasMore:  hasMore,
	}, nil
}

// ToggleLike 切换点赞状态
func (s *userContentService) ToggleLike(params *model.ToggleLikeParams) (*model.ToggleLikeResult, error) {
	// 检查内容是否存在
	_, err := s.repo.GetByItemID(params.ItemID)
	if err != nil {
		return nil, fmt.Errorf("内容不存在: %v", err)
	}
	
	// 切换点赞状态
	likes, isLiked, err := s.repo.ToggleLike(params.UserID, params.ItemID, params.IsLiked)
	if err != nil {
		return nil, fmt.Errorf("切换点赞状态失败: %v", err)
	}
	
	return &model.ToggleLikeResult{
		Likes:   likes,
		IsLiked: isLiked,
	}, nil
}

// generateItemID 生成唯一的ItemID
func (s *userContentService) generateItemID(listType model.ListType) (string, error) {
	// 生成随机字符串
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 9)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	
	return fmt.Sprintf("%s_%d_%s", listType, time.Now().Unix(), string(b)), nil
} 