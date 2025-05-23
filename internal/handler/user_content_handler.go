package handler

import (
	"net/http"
	"strconv"
	"ticktok-service/internal/model"
	"ticktok-service/internal/repository"
	"ticktok-service/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserContentHandler 用户内容处理器
type UserContentHandler struct {
	service service.UserContentService
}

// NewUserContentHandler 创建用户内容处理器实例
func NewUserContentHandler(db *gorm.DB) *UserContentHandler {
	repo := repository.NewUserContentRepository(db)
	svc := service.NewUserContentService(repo)
	return &UserContentHandler{service: svc}
}

// GetContentList 获取用户内容列表
// GET /user/content/list
func (h *UserContentHandler) GetContentList(c *gin.Context) {
	var params model.ContentListParams
	
	// 绑定查询参数
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	
	// 调用服务层
	result, err := h.service.GetContentList(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    result,
	})
}

// CreateContent 创建内容项
// POST /user/content/create
func (h *UserContentHandler) CreateContent(c *gin.Context) {
	var params model.CreateContentParams
	
	// 绑定JSON参数
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	
	// 调用服务层
	result, err := h.service.CreateContent(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "创建成功",
		"data":    result,
	})
}

// UpdateContent 更新内容项
// PUT /user/content/update
func (h *UserContentHandler) UpdateContent(c *gin.Context) {
	var params model.UpdateContentParams
	
	// 绑定JSON参数
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	
	// 调用服务层
	result, err := h.service.UpdateContent(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "更新成功",
		"data":    result,
	})
}

// DeleteContent 删除内容项
// DELETE /user/content/delete
func (h *UserContentHandler) DeleteContent(c *gin.Context) {
	// 获取查询参数
	itemID := c.Query("itemId")
	userID := c.Query("userId")
	
	if itemID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少必要参数: itemId 和 userId",
			"data":    nil,
		})
		return
	}
	
	// 调用服务层
	if err := h.service.DeleteContent(itemID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "删除成功",
		"data": gin.H{
			"success": true,
			"message": "内容已删除",
		},
	})
}

// ToggleLike 切换点赞状态
// POST /user/content/toggle-like
func (h *UserContentHandler) ToggleLike(c *gin.Context) {
	var params model.ToggleLikeParams
	
	// 绑定JSON参数
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	
	// 调用服务层
	result, err := h.service.ToggleLike(&params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "操作成功",
		"data":    result,
	})
}

// GetContentDetail 获取内容详情 (可选的扩展API)
// GET /user/content/detail/:itemId
func (h *UserContentHandler) GetContentDetail(c *gin.Context) {
	itemID := c.Param("itemId")
	userID := c.Query("userId") // 用于检查点赞状态
	
	if itemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "缺少内容ID",
			"data":    nil,
		})
		return
	}
	
	// 这里可以扩展为详情查询，目前返回基础信息
	// 可以通过调用 repository 的方法获取详细信息
	
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "功能待实现",
		"data":    gin.H{"itemId": itemID, "userId": userID},
	})
}

// BatchCreateContent 批量创建内容 (可选的扩展API)
// POST /user/content/batch-create
func (h *UserContentHandler) BatchCreateContent(c *gin.Context) {
	var batchParams struct {
		Items []model.CreateContentParams `json:"items" binding:"required"`
	}
	
	// 绑定JSON参数
	if err := c.ShouldBindJSON(&batchParams); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data":    nil,
		})
		return
	}
	
	// 批量创建
	var results []model.ContentItem
	var errors []string
	
	for i, params := range batchParams.Items {
		result, err := h.service.CreateContent(&params)
		if err != nil {
			errors = append(errors, "第"+strconv.Itoa(i+1)+"项创建失败: "+err.Error())
		} else {
			results = append(results, *result)
		}
	}
	
	response := gin.H{
		"code":    200,
		"message": "批量操作完成",
		"data": gin.H{
			"success": results,
			"errors":  errors,
			"total":   len(batchParams.Items),
			"created": len(results),
			"failed":  len(errors),
		},
	}
	
	c.JSON(http.StatusOK, response)
} 