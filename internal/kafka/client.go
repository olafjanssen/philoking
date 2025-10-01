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
	reader   *kafka.Reader
	config   config.KafkaConfig
}

func NewClient(cfg config.KafkaConfig) (*Client, error) {
	// Create producer
	producer := &kafka.Writer{
		Addr:      kafka.TCP(cfg.Brokers...),
		Balancer:  &kafka.LeastBytes{},
		BatchSize: 1,
	}

	// Create reader for chat messages
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  cfg.Brokers,
		Topic:    cfg.Topics.ChatMessages,
		GroupID:  "philoking-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	return &Client{
		producer: producer,
		reader:   reader,
		config:   cfg,
	}, nil
}

// PublishChatMessage publishes a chat message to Kafka
func (c *Client) PublishChatMessage(ctx context.Context, message *types.ChatMessage) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := kafka.Message{
		Topic: c.config.Topics.ChatMessages,
		Key:   []byte(message.ID),
		Value: data,
	}

	log.Printf("Publishing message to Kafka topic %s: %s (type: %s, agent: %s)", c.config.Topics.ChatMessages, message.Content, message.Type, message.AgentID)

	return c.producer.WriteMessages(ctx, kafkaMsg)
}

// PublishChatResponse publishes a chat response to Kafka
func (c *Client) PublishChatResponse(ctx context.Context, message *types.ChatMessage) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := kafka.Message{
		Topic: c.config.Topics.ChatResponses,
		Key:   []byte(message.ID),
		Value: data,
	}

	log.Printf("Publishing response to Kafka topic %s: %s (type: %s, agent: %s)", c.config.Topics.ChatResponses, message.Content, message.Type, message.AgentID)

	return c.producer.WriteMessages(ctx, kafkaMsg)
}

// PublishAgentMessage publishes an agent message to Kafka
func (c *Client) PublishAgentMessage(ctx context.Context, message *types.AgentMessage) error {
	data, err := message.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal agent message: %w", err)
	}

	topic := c.config.Topics.AgentInput
	if message.ToAgent != "" {
		topic = c.config.Topics.AgentOutput
	}

	kafkaMsg := kafka.Message{
		Topic: topic,
		Key:   []byte(message.ID),
		Value: data,
	}

	return c.producer.WriteMessages(ctx, kafkaMsg)
}

// SubscribeToChatMessages subscribes to chat messages
func (c *Client) SubscribeToChatMessages(ctx context.Context, handler func(*types.ChatMessage) error) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.config.Brokers,
		Topic:    c.config.Topics.ChatMessages,
		GroupID:  "philoking-agents",
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

			log.Printf("Kafka consumed message: %s (type: %s, agent: %s)", chatMsg.Content, chatMsg.Type, chatMsg.AgentID)

			if err := handler(&chatMsg); err != nil {
				log.Printf("Error handling message: %v", err)
			}
		}
	}
}

// SubscribeToChatMessagesWithGroup subscribes to chat messages with a specific consumer group
func (c *Client) SubscribeToChatMessagesWithGroup(ctx context.Context, groupID string, handler func(*types.ChatMessage) error) error {
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

// SubscribeToChatResponses subscribes to chat responses
func (c *Client) SubscribeToChatResponses(ctx context.Context, handler func(*types.ChatMessage) error) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.config.Brokers,
		Topic:    c.config.Topics.ChatResponses,
		GroupID:  "philoking-web-group",
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
				log.Printf("Error reading response message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var chatMsg types.ChatMessage
			if err := chatMsg.FromJSON(msg.Value); err != nil {
				log.Printf("Error unmarshaling response message: %v", err)
				continue
			}

			log.Printf("Kafka consumed response: %s (type: %s, agent: %s)", chatMsg.Content, chatMsg.Type, chatMsg.AgentID)

			if err := handler(&chatMsg); err != nil {
				log.Printf("Error handling response message: %v", err)
			}
		}
	}
}

// SubscribeToAgentMessages subscribes to agent messages
func (c *Client) SubscribeToAgentMessages(ctx context.Context, topic string, handler func(*types.AgentMessage) error) error {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  c.config.Brokers,
		Topic:    topic,
		GroupID:  "philoking-agents",
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				log.Printf("Error reading agent message: %v", err)
				time.Sleep(time.Second)
				continue
			}

			var agentMsg types.AgentMessage
			if err := agentMsg.FromJSON(msg.Value); err != nil {
				log.Printf("Error unmarshaling agent message: %v", err)
				continue
			}

			if err := handler(&agentMsg); err != nil {
				log.Printf("Error handling agent message: %v", err)
			}
		}
	}
}

func (c *Client) Close() error {
	if err := c.producer.Close(); err != nil {
		log.Printf("Error closing producer: %v", err)
	}
	if err := c.reader.Close(); err != nil {
		log.Printf("Error closing reader: %v", err)
	}
	return nil
}
