package main

import (
	"log"
	"net"
	"ticktok-service/api/user/v1"
	"ticktok-service/internal/user/model"
	"ticktok-service/internal/user/service"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/mysql"

	"google.golang.org/grpc"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)
	mysql.Init()

	if err := model.AutoMigrate(mysql.DB); err != nil {
		logger.Log.Fatal("failed to auto migrate: " + err.Error())
	}

	lis, err := net.Listen("tcp", ":"+config.Config.Server.GrpcPort)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}

	s := grpc.NewServer()

	user.RegisterUserServiceServer(s, service.NewUserService())

	logger.Log.Info("User service starting on port " + config.Config.Server.GrpcPort)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}
