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

// SendMessage sends a message to the default chat topic (backward compatibility)
func SendMessage(ctx context.Context, key, value []byte) error {
	return SendMessageToTopic(ctx, config.Config.Kafka.ChatTopic, key, value)
}

// SendMessageToTopic sends a message to a specific topic
func SendMessageToTopic(ctx context.Context, topic string, key, value []byte) error {
	if Producer == nil {
		return fmt.Errorf("kafka producer not initialized")
	}
	err := Producer.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
	return err
}

// NewConsumer creates a new Kafka reader (consumer) for chat topic (backward compatibility)
func NewConsumer(groupID string) *kafka.Reader {
	return NewConsumerForTopic(config.Config.Kafka.ChatTopic, groupID)
}

// NewConsumerForTopic creates a new Kafka reader for a specific topic
func NewConsumerForTopic(topic, groupID string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:  config.Config.Kafka.Brokers,
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
}
