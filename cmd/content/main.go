package main

import (
	"log"
	"net"

	contentv1 "ticktok-service/api/content/v1"
	"ticktok-service/internal/content/handler"
	"ticktok-service/internal/content/model"
	"ticktok-service/internal/content/repository"
	"ticktok-service/internal/content/service"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/mysql"

	"google.golang.org/grpc"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)

	// 初始化 MySQL
	mysql.Init()
	// 自动迁移建表
	if err := model.AutoMigrate(mysql.DB); err != nil {
		logger.Log.Fatal("AutoMigrate failed: " + err.Error())
	}
	logger.Log.Info("Database AutoMigrate successfully")

	// 初始化 MinIO
	minio.Init()

	port := ":" + config.Config.ContentService.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	// 依赖注入并注册服务
	repo := repository.NewVideoRepository(mysql.DB)
	svc := service.NewContentService(repo)
	h := handler.NewContentHandler(svc)
	contentv1.RegisterContentServiceServer(s, h)

	logger.Log.Info("Content service starting on port " + port)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}


