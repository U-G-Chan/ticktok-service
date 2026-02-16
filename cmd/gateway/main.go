package main

import (
	"ticktok-service/internal/gateway/router"
	"ticktok-service/internal/gateway/rpc"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/redis"
)

func main() {
	config.Init("config/config.yaml")
	logger.Init(config.Config.LogLevel)
	redis.Init()
	rpc.Init()

	r := router.NewRouter()

	logger.Log.Info("Gateway server starting on port " + config.Config.Server.HttpPort)
	r.Run(":" + config.Config.Server.HttpPort)
}
