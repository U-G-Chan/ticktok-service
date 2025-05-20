package main

import (
	"log"
	"ticktok-service/config"
	"ticktok-service/internal/handler"
	"ticktok-service/internal/middleware"
	"ticktok-service/internal/model"
)

func main() {

	// 加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := model.SetupDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}



	// 设置路由
	r := handler.SetupRouter(model.DB)

	// 添加中间件
	r.Use(middleware.Cors())

	// 启动服务器
	port := config.AppConfig.Server.Port
	log.Printf("服务监听端口: %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}