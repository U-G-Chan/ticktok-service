package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS 返回配置好的CORS中间件
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",  // React开发服务器
			"http://127.0.0.1:3000",
			"http://localhost:3001",  // 备用端口
			"http://127.0.0.1:3001",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS",
		},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Content-Length", "Accept-Encoding",
			"X-CSRF-Token", "Authorization", "accept", "Cache-Control", "X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
} 