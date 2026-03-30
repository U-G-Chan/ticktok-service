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
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/rpc"
	"ticktok-service/pkg/snowflake"

	"google.golang.org/grpc"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)

	// 初始化资源
	snowflake.Init()
	mysql.Init()
	redis.Init()
	minio.Init()
	kafka.InitProducer()

	// 自动迁移建表
	if err := model.AutoMigrate(mysql.DB); err != nil {
		logger.Log.Fatal("AutoMigrate failed: " + err.Error())
	}
	logger.Log.Info("Database AutoMigrate successfully")

	// 初始化 gRPC 客户端 (连接 User 服务)
	_, err := rpc.InitClientManager(map[string]string{
		"user": config.Config.Microservices.User,
	})
	if err != nil {
		logger.Log.Fatal("Failed to initialize RPC clients: " + err.Error())
	}

	port := ":" + config.Config.ContentService.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	// 依赖注入并注册服务
	videoRepo := repository.NewVideoRepository(mysql.DB)
	favoriteRepo := repository.NewFavoriteRepository(mysql.DB)
	commentRepo := repository.NewCommentRepository(mysql.DB)
	userClient := rpc.GetClientManager().UserClient
	svc := service.NewContentService(videoRepo, favoriteRepo, commentRepo, userClient)
	h := handler.NewContentHandler(svc)
	contentv1.RegisterContentServiceServer(s, h)

	logger.Log.Info("Content service starting on port " + port)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}

