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

	videoRepo := repository.NewVideoRepository(mysql.DB)
	favoriteRepo := repository.NewFavoriteRepository(mysql.DB)
	commentRepo := repository.NewCommentRepository(mysql.DB)
	coverWorker := worker.NewCoverWorker(videoRepo)
	interactionWorker := worker.NewInteractionWorker(videoRepo, favoriteRepo, commentRepo)

	ctx, cancel := context.WithCancel(context.Background())

	go coverWorker.Start(ctx)
	go interactionWorker.StartFavoriteConsumer(ctx)
	go interactionWorker.StartStatsFlusher(ctx)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down worker...")
	cancel()
	logger.Log.Info("Worker stopped")
}
