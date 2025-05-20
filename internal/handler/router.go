package handler

import (
	"github.com/gin-gonic/gin"
)

// SetupRouter 配置路由
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 创建博客处理器
	blogHandler := NewBlogHandler()

	// 博客相关API
	apiGroup := r.Group("/api")
	{
		// 博客列表
		apiGroup.GET("/blogs", blogHandler.GetBlogs)
		
		// 搜索博客 - 注意：这个路由必须放在/:id前面，否则会被误认为是id参数
		apiGroup.GET("/blogs/search", blogHandler.SearchBlogs)
		
		// 博客详情
		apiGroup.GET("/blogs/:id", blogHandler.GetBlogDetail)
	}

	return r
} 