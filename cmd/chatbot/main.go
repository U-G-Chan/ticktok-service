package main

import (
	"log"
	"net"
	pb "ticktok-service/api/chatbot/v1"
	"ticktok-service/internal/chatbot/model"
	"ticktok-service/internal/chatbot/service"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"

	"google.golang.org/grpc"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)
	snowflake.Init()
	mysql.Init()
	redis.Init()

	if err := model.AutoMigrate(mysql.DB); err != nil {
		logger.Log.Fatal("failed to auto migrate: " + err.Error())
	}

	port := ":" + config.Config.ChatbotService.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	// Register Chatbot Service
	pb.RegisterChatbotServiceServer(s, service.NewChatbotService())

	logger.Log.Info("Chatbot service starting on port " + port)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}
