package repository

import (
	"ticktok-service/internal/model"

	"gorm.io/gorm"
)

// SlideRepository 轮播内容数据仓库接口
type SlideRepository interface {
	GetSlideItems(startIndex, pageSize int) ([]*model.SlideItem, int64, error)
	GetSlideItemByItemID(itemID string) (*model.SlideItem, error)
	GetSlideItemsByType(contentType string, startIndex, pageSize int) ([]*model.SlideItem, int64, error)
	SearchSlideItems(keyword string, startIndex, pageSize int) ([]*model.SlideItem, int64, error)
}

// slideRepository 轮播内容数据仓库实现
type slideRepository struct {
	db *gorm.DB
}

// NewSlideRepository 创建轮播内容数据仓库
func NewSlideRepository(db *gorm.DB) SlideRepository {
	return &slideRepository{
		db: db,
	}
}

// GetSlideItems 获取轮播内容列表
func (r *slideRepository) GetSlideItems(startIndex, pageSize int) ([]*model.SlideItem, int64, error) {
	var items []*model.SlideItem
	var total int64
	
	// 查询总记录数
	if err := r.db.Model(&model.SlideItem{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询轮播内容列表并预加载关联数据
	if err := r.db.Preload("Labels").
		Preload("Album", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Offset(startIndex).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	
	return items, total, nil
}

// GetSlideItemByItemID 根据ItemID获取轮播内容详情
func (r *slideRepository) GetSlideItemByItemID(itemID string) (*model.SlideItem, error) {
	var item model.SlideItem
	
	// 查询轮播内容详情并预加载关联数据
	if err := r.db.Preload("Labels").
		Preload("Album", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("item_id = ?", itemID).
		First(&item).Error; err != nil {
		return nil, err
	}
	
	return &item, nil
}

// GetSlideItemsByType 根据内容类型获取轮播内容
func (r *slideRepository) GetSlideItemsByType(contentType string, startIndex, pageSize int) ([]*model.SlideItem, int64, error) {
	var items []*model.SlideItem
	var total int64
	
	// 查询指定类型的总记录数
	if err := r.db.Model(&model.SlideItem{}).
		Where("content_type = ?", contentType).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询指定类型的轮播内容并预加载关联数据
	if err := r.db.Preload("Labels").
		Preload("Album", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("content_type = ?", contentType).
		Offset(startIndex).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	
	return items, total, nil
}

// SearchSlideItems 搜索轮播内容
func (r *slideRepository) SearchSlideItems(keyword string, startIndex, pageSize int) ([]*model.SlideItem, int64, error) {
	var items []*model.SlideItem
	var total int64
	
	// 查询子查询获取匹配标签的item_id
	subQuery := r.db.Model(&model.SlideItemLabel{}).
		Select("item_id").
		Where("label_content LIKE ?", "%"+keyword+"%").
		Group("item_id")
	
	// 查询匹配条件的总记录数
	if err := r.db.Model(&model.SlideItem{}).
		Where("title LIKE ? OR author LIKE ? OR item_id IN (?)", 
			"%"+keyword+"%", "%"+keyword+"%", subQuery).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// 查询匹配条件的轮播内容并预加载关联数据
	if err := r.db.Preload("Labels").
		Preload("Album", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Where("title LIKE ? OR author LIKE ? OR item_id IN (?)", 
			"%"+keyword+"%", "%"+keyword+"%", subQuery).
		Offset(startIndex).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&items).Error; err != nil {
		return nil, 0, err
	}
	
	return items, total, nil
} 