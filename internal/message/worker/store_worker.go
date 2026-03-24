package worker

import (
	"context"
	"encoding/json"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/internal/message/model"
	"ticktok-service/internal/message/repository"
	"ticktok-service/pkg/config"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
)

func StartStoreWorker(ctx context.Context) {
	reader := kafka.NewConsumer(config.Config.Kafka.StoreGroupID)
	defer reader.Close()

	logger.Log.Info("Store Worker started...")

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			logger.Log.Sugar().Errorf("Store Worker read error: %v", err)
			break
		}

		var msg messagev1.Message
		if err := json.Unmarshal(m.Value, &msg); err != nil {
			logger.Log.Sugar().Errorf("Store Worker unmarshal error: %v", err)
			continue
		}

		// 1. Save to MySQL
		dbMsg := &model.Message{
			ID:         msg.Id,
			ToUserID:   msg.ToUserId,
			FromUserID: msg.FromUserId,
			Content:    msg.Content,
			CreateTime: msg.CreateTime,
		}
		if err := repository.CreateMessage(ctx, dbMsg); err != nil {
			logger.Log.Sugar().Errorf("Failed to save message to DB: %v", err)
			continue
		}

		// 2. Add to Redis Inbox
		if err := repository.AddToInbox(ctx, msg.ToUserId, msg.Id); err != nil {
			logger.Log.Sugar().Errorf("Failed to add message to inbox: %v", err)
		}
	}
}
