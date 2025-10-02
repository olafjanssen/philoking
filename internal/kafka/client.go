package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"philoking/internal/config"
	"philoking/internal/types"

	"github.com/segmentio/kafka-go"
)

type Client struct {
	producer *kafka.Writer
	config   config.KafkaConfig
}

func NewClient(cfg config.KafkaConfig) (*Client, error) {
	// Create producer
	producer := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Balancer:  &kafka.LeastBytes{},
		BatchSize: 1,
	}

	return &Client{
		producer: producer,
		config:   cfg,
	}, nil
}

// PublishMessage publishes a message to the chat topic
func (c *Client) PublishMessage(ctx context.Context, message *types.ChatMessage) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	log.Printf("Publishing message to Kafka topic %s: %s (type: %s, agent: %s)", c.config.Topics.ChatMessages, message.Content, message.Type, message.AgentID)

	return c.producer.WriteMessages(ctx, kafka.Message{
		Topic: c.config.Topics.ChatMessages,
		Value: data,
	})
}

// SubscribeToMessages subscribes to chat messages with a specific consumer group
func (c *Client) SubscribeToMessages(ctx context.Context, groupID string, handler func(*types.ChatMessage) error) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.config.Brokers,
		Topic:    c.config.Topics.ChatMessages,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var chatMsg types.ChatMessage
			if err := chatMsg.FromJSON(msg.Value); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			log.Printf("Kafka consumed message in group %s: %s (type: %s, agent: %s)", groupID, chatMsg.Content, chatMsg.Type, chatMsg.AgentID)

			if err := handler(&chatMsg); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

// Close closes the Kafka client
func (c *Client) Close() error {
	return c.producer.Close()
}
