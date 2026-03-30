package main

import (
	"context"
	"log"
	"net"
	"net/http"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/internal/message/model"
	"ticktok-service/internal/message/service"
	"ticktok-service/internal/message/worker"
	"ticktok-service/internal/message/ws"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/mysql"
	"ticktok-service/pkg/redis"
	"ticktok-service/pkg/snowflake"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
)

func main() {
	// 1. Initialize configs
	if err := config.Init("config/config.yaml"); err != nil {
		log.Fatalf("Init config failed: %v", err)
	}
	logger.Init(config.Config.LogLevel)

	// 2. Initialize DB, Redis, Kafka
	mysql.Init()
	redis.Init()
	kafka.InitProducer()
	defer kafka.CloseProducer()

	// 3. AutoMigrate Models
	if err := model.AutoMigrate(mysql.DB); err != nil {
		logger.Log.Fatal("failed to auto migrate: " + err.Error())
	}

	// 4. Initialize Snowflake
	snowflake.Init()

	// 5. Start Workers
	ctx := context.Background()
	go worker.StartPushWorker(ctx)
	go worker.StartStoreWorker(ctx)

	// 6. Setup cmux on the configured port
	port := ":" + config.Config.MessageService.Port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		logger.Log.Fatal("failed to listen: " + err.Error())
	}

	m := cmux.New(lis)

	// Match gRPC connections first
	grpcL := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	// Match HTTP connections
	httpL := m.Match(cmux.Any())

	// 7. Setup gRPC Server
	grpcServer := grpc.NewServer()
	messagev1.RegisterMessageServiceServer(grpcServer, service.NewMessageServiceServer())
	go func() {
		if err := grpcServer.Serve(grpcL); err != nil {
			logger.Log.Fatal("grpc server error: " + err.Error())
		}
	}()

	// 8. Setup HTTP Server (WebSocket)
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/api/v1/message/connection", ws.WsHandler)
	httpServer := &http.Server{
		Handler: httpMux,
	}
	go func() {
		if err := httpServer.Serve(httpL); err != nil {
			logger.Log.Fatal("http server error: " + err.Error())
		}
	}()

	logger.Log.Info("Message service starting on port " + port)
	if err := m.Serve(); err != nil {
		logger.Log.Fatal("cmux serve error: " + err.Error())
	}
}
