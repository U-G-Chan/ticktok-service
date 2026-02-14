package main

import (
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	config.Init("config/config.yaml")
	logger.Init(config.Config.LogLevel)

	r := gin.Default()
	
	// TODO: Register Routes
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	logger.Log.Info("Gateway server starting on port " + config.Config.Server.HttpPort)
	r.Run(":" + config.Config.Server.HttpPort)
}
