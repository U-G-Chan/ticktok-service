package repository

import (
	"encoding/json"
	"fmt"
	"ticktok-service/internal/model"
	"time"

	"gorm.io/gorm"
)

// UserContentRepository 用户内容仓库接口
type UserContentRepository interface {
	// 基础CRUD操作
	Create(content *model.UserContent) error
	GetByItemID(itemID string) (*model.UserContent, error)
	Update(content *model.UserContent) error
	Delete(itemID, userID string) error

	// 列表查询
	GetList(params *model.ContentListParams) ([]model.UserContent, int64, error)

	// 点赞相关
	ToggleLike(userID, itemID string, isLiked bool) (int, bool, error)
	IsLiked(userID, itemID string) (bool, error)
	GetLikeCount(itemID string) (int, error)
}

// userContentRepository 用户内容仓库实现
type userContentRepository struct {
	db *gorm.DB
}

// NewUserContentRepository 创建用户内容仓库实例
func NewUserContentRepository(db *gorm.DB) UserContentRepository {
	return &userContentRepository{db: db}
}

// Create 创建内容项
func (r *userContentRepository) Create(content *model.UserContent) error {
	return r.db.Create(content).Error
}

// GetByItemID 根据项目ID获取内容
func (r *userContentRepository) GetByItemID(itemID string) (*model.UserContent, error) {
	var content model.UserContent
	err := r.db.Where("item_id = ?", itemID).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// Update 更新内容项
func (r *userContentRepository) Update(content *model.UserContent) error {
	return r.db.Save(content).Error
}

// Delete 删除内容项
func (r *userContentRepository) Delete(itemID, userID string) error {
	result := r.db.Where("item_id = ? AND user_id = ?", itemID, userID).Delete(&model.UserContent{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("内容项不存在或无权限删除")
	}
	return nil
}

// GetList 获取内容列表
func (r *userContentRepository) GetList(params *model.ContentListParams) ([]model.UserContent, int64, error) {
	var contents []model.UserContent
	var total int64

	// 设置默认分页参数
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// 构建查询条件
	query := r.db.Model(&model.UserContent{})

	// 根据列表类型和用户ID筛选
	switch params.ListType {
	case model.ListTypeWorks:
		// 作品列表：当前用户的所有内容
		query = query.Where("user_id = ? AND list_type = ?", params.UserID, model.ListTypeWorks)
	case model.ListTypeRecommend:
		// 推荐列表：公开的已发布内容
		query = query.Where("is_public = ? AND work_type = ?", true, model.ListTypeRecommend)
	case model.ListTypeCollection:
		// 收藏列表：TODO - 需要额外的收藏表关联
		query = query.Where("user_id = ? AND list_type = ?", params.UserID, model.ListTypeCollection)
	case model.ListTypeLikes:
		// 点赞列表：用户点赞过的内容
		query = query.Where("user_id = ? AND list_type = ?", params.UserID, model.ListTypeLikes)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&contents).Error; err != nil {
		return nil, 0, err
	}

	return contents, total, nil
}

// ToggleLike 切换点赞状态
func (r *userContentRepository) ToggleLike(userID, itemID string, isLiked bool) (int, bool, error) {
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return 0, false, tx.Error
	}

	// 检查是否已点赞
	var existingLike model.ContentLike
	err := tx.Where("user_id = ? AND item_id = ?", userID, itemID).First(&existingLike).Error

	if isLiked {
		// 要点赞
		if err == gorm.ErrRecordNotFound {
			// 不存在点赞记录，创建新的
			like := model.ContentLike{
				UserID:    userID,
				ItemID:    itemID,
				CreatedAt: time.Now(),
			}
			if err := tx.Create(&like).Error; err != nil {
				tx.Rollback()
				return 0, false, err
			}
		} else if err != nil {
			tx.Rollback()
			return 0, false, err
		}
		// 如果已存在点赞记录，不做任何操作
	} else {
		// 要取消点赞
		if err == nil {
			// 存在点赞记录，删除它
			if err := tx.Delete(&existingLike).Error; err != nil {
				tx.Rollback()
				return 0, false, err
			}
		} else if err != gorm.ErrRecordNotFound {
			tx.Rollback()
			return 0, false, err
		}
		// 如果不存在点赞记录，不做任何操作
	}

	// 更新内容项的点赞数
	var likeCount int64
	if err := tx.Model(&model.ContentLike{}).Where("item_id = ?", itemID).Count(&likeCount).Error; err != nil {
		tx.Rollback()
		return 0, false, err
	}

	if err := tx.Model(&model.UserContent{}).Where("item_id = ?", itemID).Update("likes", likeCount).Error; err != nil {
		tx.Rollback()
		return 0, false, err
	}

	if err := tx.Commit().Error; err != nil {
		return 0, false, err
	}

	// 检查最终点赞状态
	finalIsLiked, err := r.IsLiked(userID, itemID)
	if err != nil {
		return int(likeCount), isLiked, err
	}

	return int(likeCount), finalIsLiked, nil
}

// IsLiked 检查用户是否已点赞
func (r *userContentRepository) IsLiked(userID, itemID string) (bool, error) {
	var count int64
	err := r.db.Model(&model.ContentLike{}).Where("user_id = ? AND item_id = ?", userID, itemID).Count(&count).Error
	return count > 0, err
}

// GetLikeCount 获取点赞数
func (r *userContentRepository) GetLikeCount(itemID string) (int, error) {
	var count int64
	err := r.db.Model(&model.ContentLike{}).Where("item_id = ?", itemID).Count(&count).Error
	return int(count), err
}

// ConvertToContentItem 将数据库模型转换为API模型
func ConvertToContentItem(content *model.UserContent) (*model.ContentItem, error) {
	item := &model.ContentItem{
		ItemID:    content.ItemID,
		ListType:  content.ListType,
		Thumbnail: content.Thumbnail,
		Likes:     content.Likes,
	}

	// 构建Other信息
	other := &model.OtherInfo{
		WorkType:    content.WorkType,
		Title:       content.Title,
		Description: content.Description,
		PublishTime: content.PublishTime,
		Duration:    content.Duration,
		IsPublic:    content.IsPublic,
		CreateTime:  content.CreatedAt,
		UpdateTime:  content.UpdatedAt,
	}

	// 解析Tags
	if content.Tags != "" {
		var tags []string
		if err := json.Unmarshal([]byte(content.Tags), &tags); err == nil {
			other.Tags = tags
		}
	}

	item.Other = other
	return item, nil
}
