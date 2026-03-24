package worker

import (
	"context"
	"encoding/json"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/internal/message/router"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func StartPushWorker(ctx context.Context) {
	reader := kafka.NewConsumer(config.Config.Kafka.PushGroupID)
	defer reader.Close()

	logger.Log.Info("Push Worker started...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Sugar().Errorf("Push Worker read error: %v", err)
			break
		}

		var msg messagev1.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logger.Log.Sugar().Errorf("Push Worker unmarshal error: %v", err)
			continue
		}

		// Check if target user is online
		nodeIP, err := router.GetRoute(ctx, msg.ToUserId)
		if err == nil && nodeIP != "" {
			// Try local push first if nodeIP is this node's IP
			// (Assuming nodeIP is actually ip:port)
			
			// Setup RPC connection to the target node
			conn, err := grpc.NewClient(nodeIP, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logger.Log.Sugar().Errorf("Failed to connect to node %s: %v", nodeIP, err)
				continue
			}

			client := messagev1.NewMessageServiceClient(conn)
			_, err = client.PushMsgToClient(ctx, &messagev1.PushMsgRequest{Message: &msg})
			if err != nil {
				logger.Log.Sugar().Errorf("Failed to push message via RPC to node %s: %v", nodeIP, err)
			}
			conn.Close()
		} else {
			// User offline, do nothing in Push Worker
			logger.Log.Sugar().Infof("User %d is offline, push skipped", msg.ToUserId)
		}
	}
}
