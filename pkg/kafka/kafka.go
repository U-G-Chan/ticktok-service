package kafka

import (
	"context"
	"fmt"
	"ticktok-service/pkg/config"

	"github.com/segmentio/kafka-go"
)

var (
	Producer *kafka.Writer
)

// InitProducer initializes the global Kafka producer
func InitProducer() {
	Producer = &kafka.Writer{
		Addr:     kafka.TCP(config.Config.Kafka.Brokers...),
		Topic:    config.Config.Kafka.ChatTopic,
		Balancer: &kafka.Hash{},
	}
}

// CloseProducer closes the global Kafka producer
func CloseProducer() error {
	if Producer != nil {
		return Producer.Close()
	}
	return nil
}

// SendMessage sends a message to the default chat topic
func SendMessage(ctx context.Context, key, value []byte) error {
	if Producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}
	err := Producer.WriteMessages(ctx, kafka.Message{
		Key:   key,
		Value: value,
	})
	return err
}

// NewConsumer creates a new Kafka reader (consumer)
func NewConsumer(groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Config.Kafka.Brokers,
		GroupID:  groupID,
		Topic:    config.Config.Kafka.ChatTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
