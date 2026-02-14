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

	port := "9093"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}
	
	s := grpc.NewServer()
	
	// TODO: Register Chatbot Service

	logger.Log.Info("Chatbot service starting on port " + port)
	if err := s.Serve(lis); err != nil {
		logger.Log.Fatal("failed to serve: " + err.Error())
	}
}
