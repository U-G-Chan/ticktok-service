package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"ticktok-service/internal/content/repository"
	"ticktok-service/internal/content/worker"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/minio"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/redis"
)

func main() {
	// Initialize configurations
	if err := config.Init("config/config.yaml"); err != nil {
		panic(err)
	}

	// Initialize logger
	logger.Init(config.Config.LogLevel)

	// Initialize resources
	mysql.Init()
	redis.Init()
	minio.Init()

	repo := repository.NewVideoRepository(mysql.DB)
	coverWorker := worker.NewCoverWorker(repo)

	ctx, cancel := context.WithCancel(context.Background())

	// Start the worker in a goroutine
	go coverWorker.Start(ctx)

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down worker...")
	cancel()
	logger.Log.Info("Worker stopped")
}
