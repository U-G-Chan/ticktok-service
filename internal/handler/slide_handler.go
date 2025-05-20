package handler

import (
	"strconv"
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SlideHandler 轮播内容相关处理器
type SlideHandler struct {
	slideService service.SlideService
}

// NewSlideHandler 创建新的轮播内容处理器
func NewSlideHandler(db *gorm.DB) *SlideHandler {
	return &SlideHandler{
		slideService: service.NewSlideService(db),
	}
}

// GetSlideItems 获取轮播内容列表
func (h *SlideHandler) GetSlideItems(c *gin.Context) {
	// 解析分页参数
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "0"))
	if startIndex < 0 {
		startIndex = 0
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取轮播内容列表
	result, err := h.slideService.GetSlideItems(startIndex, pageSize)
	if err != nil {
		util.Fail(c, 500, "获取轮播内容列表失败: "+err.Error())
		return
	}

	util.Success(c, result.Items)
}

// GetSlideItemDetail 获取轮播内容详情
func (h *SlideHandler) GetSlideItemDetail(c *gin.Context) {
	// 获取路径参数
	itemID := c.Param("itemId")
	if itemID == "" {
		util.Fail(c, 400, "内容ID不能为空")
		return
	}

	// 获取轮播内容详情
	item, err := h.slideService.GetSlideItemByItemID(itemID)
	if err != nil {
		util.Fail(c, 404, "轮播内容不存在: "+err.Error())
		return
	}

	util.Success(c, item)
}

// GetSlideItemsByType 根据内容类型获取轮播内容
func (h *SlideHandler) GetSlideItemsByType(c *gin.Context) {
	// 获取路径参数
	contentType := c.Param("contentType")
	if contentType != "video" && contentType != "picture" {
		util.Fail(c, 400, "内容类型必须是video或picture")
		return
	}

	// 解析分页参数
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "0"))
	if startIndex < 0 {
		startIndex = 0
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取指定类型的轮播内容
	result, err := h.slideService.GetSlideItemsByType(contentType, startIndex, pageSize)
	if err != nil {
		util.Fail(c, 500, "获取轮播内容失败: "+err.Error())
		return
	}

	util.Success(c, result.Items)
}

// SearchSlideItems 搜索轮播内容
func (h *SlideHandler) SearchSlideItems(c *gin.Context) {
	// 获取查询参数
	keyword := c.Query("keyword")
	if keyword == "" {
		util.Fail(c, 400, "搜索关键词不能为空")
		return
	}

	// 解析分页参数
	startIndex, _ := strconv.Atoi(c.DefaultQuery("startIndex", "0"))
	if startIndex < 0 {
		startIndex = 0
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 搜索轮播内容
	result, err := h.slideService.SearchSlideItems(keyword, startIndex, pageSize)
	if err != nil {
		util.Fail(c, 500, "搜索轮播内容失败: "+err.Error())
		return
	}

	util.Success(c, result.Items)
} 