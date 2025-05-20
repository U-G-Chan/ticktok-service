package handler

import (
	"strconv"
	"ticktok-service/internal/service"
	"ticktok-service/pkg/util"

	"github.com/gin-gonic/gin"
)

// BlogHandler 博客相关处理器
type BlogHandler struct {
	blogService service.BlogService
}

// NewBlogHandler 创建新的博客处理器
func NewBlogHandler() *BlogHandler {
	return &BlogHandler{
		blogService: service.NewBlogService(),
	}
}

// GetBlogs 获取博客列表
func (h *BlogHandler) GetBlogs(c *gin.Context) {
	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 获取博客列表
	blogs, total, err := h.blogService.GetBlogs(page, pageSize)
	if err != nil {
		util.Fail(c, 500, "获取博客列表失败: "+err.Error())
		return
	}

	// 处理图片和标签数据，转换为前端需要的格式
	for i := range blogs {
		// 处理图片URL
		var images []string
		for _, img := range blogs[i].Images {
			images = append(images, img.ImageURL)
		}
		blogs[i].Images = nil

		// 处理标签
		var tags []string
		for _, tag := range blogs[i].Tags {
			tags = append(tags, tag.TagContent)
		}
		blogs[i].Tags = nil

		// 设置假数据
		blogs[i].IsFollowing = true
	}

	// 返回分页结果
	pageResult := util.NewPageResult(blogs, total, page, pageSize)
	util.Success(c, pageResult)
}

// GetBlogDetail 获取博客详情
func (h *BlogHandler) GetBlogDetail(c *gin.Context) {
	// 解析ID参数
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		util.Fail(c, 400, "无效的博客ID")
		return
	}

	// 获取博客详情
	blog, err := h.blogService.GetBlogByID(uint(id))
	if err != nil {
		util.Fail(c, 404, "博客不存在: "+err.Error())
		return
	}

	// // 处理图片和标签数据，转换为前端需要的格式
	// var images []string
	// for _, img := range blog.Images {
	// 	images = append(images, img.ImageURL)
	// }
	// blog.Images = nil

	// var tags []string
	// for _, tag := range blog.Tags {
	// 	tags = append(tags, tag.TagContent)
	// }
	// // blog.Tags = nil
	

	// // 设置假数据
	// blog.IsFollowing = true

	util.Success(c, blog)
}

// SearchBlogs 搜索博客
func (h *BlogHandler) SearchBlogs(c *gin.Context) {
	// 获取关键词
	keyword := c.Query("keyword")
	if keyword == "" {
		util.Fail(c, 400, "搜索关键词不能为空")
		return
	}

	// 解析分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 搜索博客
	blogs, total, err := h.blogService.SearchBlogs(keyword, page, pageSize)
	if err != nil {
		util.Fail(c, 500, "搜索博客失败: "+err.Error())
		return
	}

	// 处理图片和标签数据，转换为前端需要的格式
	for i := range blogs {
		// 处理图片URL
		var images []string
		for _, img := range blogs[i].Images {
			images = append(images, img.ImageURL)
		}
		blogs[i].Images = nil

		// 处理标签
		var tags []string
		for _, tag := range blogs[i].Tags {
			tags = append(tags, tag.TagContent)
		}
		blogs[i].Tags = nil

		// 设置假数据
		blogs[i].IsFollowing = true
	}

	// 返回分页结果
	pageResult := util.NewPageResult(blogs, total, page, pageSize)
	util.Success(c, pageResult)
} 