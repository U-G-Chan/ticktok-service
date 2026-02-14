package main

import (
	"log"
	"net"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	
	"google.golang.org/grpc"
)

func main() {
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)

	lis, err := net.Listen("tcp", ":"+config.Config.Server.GrpcPort)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}
	
	s := grpc.NewServer()
	
	// TODO: Register User Service
	// user.RegisterUserServiceServer(s, &service.UserService{})

	logger.Log.Info("User service starting on port " + config.Config.Server.GrpcPort)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}
