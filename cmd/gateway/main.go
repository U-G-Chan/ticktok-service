package main

import (
	"ticktok-service/internal/gateway/router"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/rpc"
)

func main() {
	config.Init("config/config.yaml")
	logger.Init(config.Config.LogLevel)
	redis.Init()
	// Initialize RPC clients
	serviceAddrs := map[string]string{
		"user":    config.Config.Microservices.User,
		"content": config.Config.Microservices.Content,
		"message": config.Config.Microservices.Message,
		"chatbot": config.Config.Microservices.Chatbot,
	}
	if _, err := rpc.InitClientManager(serviceAddrs); err != nil {
		logger.Log.Fatal("Failed to init RPC clients: " + err.Error())
	}

	r := router.NewRouter()

	logger.Log.Info("Gateway server starting on port " + config.Config.Server.HttpPort)
	r.Run(":" + config.Config.Server.HttpPort)
}
