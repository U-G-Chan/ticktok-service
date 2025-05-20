package handler

import (
	"strconv"
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProductHandler 商品相关处理器
type ProductHandler struct {
	productService service.ProductService
}

// NewProductHandler 创建新的商品处理器
func NewProductHandler(db *gorm.DB) *ProductHandler {
	return &ProductHandler{
		productService: service.NewProductService(db),
	}
}

// GetProducts 获取商品列表
func (h *ProductHandler) GetProducts(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取商品列表
	result, err := h.productService.GetProductList(page, pageSize)
	if err != nil {
		util.Fail(c, 500, "获取商品列表失败: "+err.Error())
		return
	}

	util.Success(c, result)
}

// GetProductDetail 获取商品详情
func (h *ProductHandler) GetProductDetail(c *gin.Context) {
	// 解析路径参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的商品ID")
		return
	}

	// 获取商品详情
	product, err := h.productService.GetProductDetail(uint(id))
	if err != nil {
		util.Fail(c, 404, "商品不存在: "+err.Error())
		return
	}

	util.Success(c, product)
}

// GetLabels 获取商品标签
func (h *ProductHandler) GetLabels(c *gin.Context) {
	// 获取标签
	labels, err := h.productService.GetLabels()
	if err != nil {
		util.Fail(c, 500, "获取标签失败: "+err.Error())
		return
	}

	util.Success(c, labels)
} 