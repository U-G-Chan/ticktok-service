package service

import (
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"

	"gorm.io/gorm"
)

// SlideService 轮播内容服务接口
type SlideService interface {
	GetSlideItems(startIndex, pageSize int) (*model.SlideResponse, error)
	GetSlideItemByItemID(itemID string) (*model.SlideItemResponse, error)
	GetSlideItemsByType(contentType string, startIndex, pageSize int) (*model.SlideResponse, error)
	SearchSlideItems(keyword string, startIndex, pageSize int) (*model.SlideResponse, error)
}

// slideService 轮播内容服务实现
type slideService struct {
	slideRepo repository.SlideRepository
}

// NewSlideService 创建轮播内容服务
func NewSlideService(db *gorm.DB) SlideService {
	return &slideService{
		slideRepo: repository.NewSlideRepository(db),
	}
}

// 将SlideItem转换为SlideItemResponse
func convertToSlideItemResponse(item *model.SlideItem) *model.SlideItemResponse {
	// 收集标签
	var labels []string
	for _, label := range item.Labels {
		labels = append(labels, label.LabelContent)
	}
	
	// 收集相册图片
	var album []string
	if len(item.Album) > 0 {
		for _, img := range item.Album {
			album = append(album, img.ImageURL)
		}
	}
	
	// 创建响应
	response := &model.SlideItemResponse{
		ID:          item.ID,
		ItemID:      item.ItemID,
		ContentType: item.ContentType,
		Title:       item.Title,
		Author:      item.Author,
		Likes:       item.Likes,
		Comments:    item.Comments,
		Stars:       item.Stars,
		Forwards:    item.Forwards,
		Labels:      labels,
		Avatar:      item.Avatar,
	}
	
	// 根据内容类型设置相应字段
	if item.ContentType == "video" {
		response.VideoURL = item.VideoURL
	} else if item.ContentType == "picture" {
		response.Album = album
	}
	
	return response
}

// GetSlideItems 获取轮播内容列表
func (s *slideService) GetSlideItems(startIndex, pageSize int) (*model.SlideResponse, error) {
	// 获取轮播内容列表
	items, total, err := s.slideRepo.GetSlideItems(startIndex, pageSize)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	var itemResponses []*model.SlideItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, convertToSlideItemResponse(item))
	}
	
	// 判断是否有更多数据
	hasMore := int64(startIndex+len(itemResponses)) < total
	
	// 创建响应
	response := &model.SlideResponse{
		Items:   itemResponses,
		Total:   total,
		HasMore: hasMore,
	}
	
	return response, nil
}

// GetSlideItemByItemID 根据ItemID获取轮播内容详情
func (s *slideService) GetSlideItemByItemID(itemID string) (*model.SlideItemResponse, error) {
	// 获取轮播内容详情
	item, err := s.slideRepo.GetSlideItemByItemID(itemID)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	response := convertToSlideItemResponse(item)
	
	return response, nil
}

// GetSlideItemsByType 根据内容类型获取轮播内容
func (s *slideService) GetSlideItemsByType(contentType string, startIndex, pageSize int) (*model.SlideResponse, error) {
	// 获取指定类型的轮播内容
	items, total, err := s.slideRepo.GetSlideItemsByType(contentType, startIndex, pageSize)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	var itemResponses []*model.SlideItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, convertToSlideItemResponse(item))
	}
	
	// 判断是否有更多数据
	hasMore := int64(startIndex+len(itemResponses)) < total
	
	// 创建响应
	response := &model.SlideResponse{
		Items:   itemResponses,
		Total:   total,
		HasMore: hasMore,
	}
	
	return response, nil
}

// SearchSlideItems 搜索轮播内容
func (s *slideService) SearchSlideItems(keyword string, startIndex, pageSize int) (*model.SlideResponse, error) {
	// 搜索轮播内容
	items, total, err := s.slideRepo.SearchSlideItems(keyword, startIndex, pageSize)
	if err != nil {
		return nil, err
	}
	
	// 转换为响应格式
	var itemResponses []*model.SlideItemResponse
	for _, item := range items {
		itemResponses = append(itemResponses, convertToSlideItemResponse(item))
	}
	
	// 判断是否有更多数据
	hasMore := int64(startIndex+len(itemResponses)) < total
	
	// 创建响应
	response := &model.SlideResponse{
		Items:   itemResponses,
		Total:   total,
		HasMore: hasMore,
	}
	
	return response, nil
} 