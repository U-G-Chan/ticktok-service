package handler

import (
	"ticktok-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册路由
func RegisterRoutes(r *gin.Engine, db *gorm.DB) {
	// 创建各种处理器
	friendHandler := NewFriendHandler(db)
	messageHandler := NewMessageHandler(db)
	userHandler := NewUserHandler(db)
	blogHandler := NewBlogHandler()
	productHandler := NewProductHandler(db)
	slideHandler := NewSlideHandler(db)
	uploadHandler := NewUploadHandler(db)
	publishHandler := NewPublishHandler(db)

	// API路由组
	api := r.Group("/api")
	{
		// 好友相关路由
		api.GET("/friends", friendHandler.GetFriends)

		// 消息相关路由
		api.GET("/messages", messageHandler.GetMessages)
		api.GET("/chat/:userId", messageHandler.GetChatHistory)
		api.POST("/chat/send", messageHandler.SendMessage)
		api.PUT("/chat/read/:userId", messageHandler.MarkAsRead)

		// 用户相关路由
		api.GET("/user/:userId", userHandler.GetUser)
		api.POST("/users/batch", userHandler.GetUsersBatch)
		
		// 博客相关路由
		api.GET("/blogs", blogHandler.GetBlogs)
		// 搜索博客 - 注意：这个路由必须放在/:id前面，否则会被误认为是id参数
		api.GET("/blogs/search", blogHandler.SearchBlogs)
		// 博客详情
		api.GET("/blogs/:id", blogHandler.GetBlogDetail)
		
		// 商城相关路由
		mall := api.Group("/mall")
		{
			// 商品列表
			mall.GET("/products", productHandler.GetProducts)
			// 商品详情
			mall.GET("/products/:id", productHandler.GetProductDetail)
			// 标签
			mall.GET("/labels", productHandler.GetLabels)
		}
		
		// 轮播内容相关路由
		slide := api.Group("/slide")
		{
			// 获取轮播内容列表
			slide.GET("/items", slideHandler.GetSlideItems)
			
			// 注意：搜索和类型路由要放在/:itemId前面，否则会被误认为是itemId参数
			
			// 按内容类型获取轮播内容
			slide.GET("/items/type/:contentType", slideHandler.GetSlideItemsByType)
			
			// 按标签搜索轮播内容
			slide.GET("/items/search", slideHandler.SearchSlideItems)
			
			// 获取轮播内容详情
			slide.GET("/items/:itemId", slideHandler.GetSlideItemDetail)
		}
		
		// 发布相关路由
		publish := api.Group("/publish")
		{
			// 保存草稿
			publish.POST("/drafts", publishHandler.SaveDraft)
			// 发布内容
			publish.POST("/contents", publishHandler.PublishContent)
			// 获取话题列表
			publish.POST("/topics", publishHandler.GetTopics)
		}
		
		// 文件上传相关路由
		api.POST("/upload/media", uploadHandler.UploadMedia)
	}
	
	// 配置静态文件服务
	r.Static("/uploads", "./uploads")
}

// SetupRouter 配置路由
func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	
	// 添加CORS中间件
	r.Use(middleware.CORS())
	
	// 注册所有路由
	RegisterRoutes(r, db)
	
	return r
} 