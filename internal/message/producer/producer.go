package producer

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	messagev1 "ticktok-service/api/message/v1"
	"ticktok-service/pkg/kafka"
	"ticktok-service/pkg/logger"
	"ticktok-service/pkg/snowflake"
)

// SendMessage handles the business logic of creating a message and sending it to Kafka
func SendMessage(ctx context.Context, fromUserID, toUserID int64, content string) (int64, error) {
	msgID := snowflake.GenerateMsgID()

	msg := &messagev1.Message{
		Id:         msgID,
		ToUserId:   toUserID,
		FromUserId: fromUserID,
		Content:    content,
		CreateTime: time.Now().UnixMilli(),
	}

	// Serialize the message
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to marshal message: %v", err)
		return 0, err
	}

	// Produce to Kafka
	err = kafka.SendMessage(ctx, []byte(strconv.FormatInt(fromUserID, 10)), msgBytes)
	if err != nil {
		logger.Log.Sugar().Errorf("Failed to produce message to Kafka: %v", err)
		return 0, err
	}

	return msgID, nil
}
