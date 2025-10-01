package types

import (
	"encoding/json"
	"time"
)

// MessageType represents the type of message in the chat system
type MessageType string

const (
	MessageTypeUser    MessageType = "user"
	MessageTypeAgent   MessageType = "agent"
	MessageTypeSystem  MessageType = "system"
	MessageTypeContext MessageType = "context"
)

// ChatMessage represents a message in the chat system
type ChatMessage struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Content   string      `json:"content"`
	AgentID   string      `json:"agent_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Metadata  Metadata    `json:"metadata,omitempty"`
}

// Metadata contains additional information about the message
type Metadata struct {
	ConversationID string            `json:"conversation_id,omitempty"`
	ReplyTo        string            `json:"reply_to,omitempty"`
	FromAgent      string            `json:"from_agent,omitempty"` // Human-readable agent name
	Tags           []string          `json:"tags,omitempty"`
	Custom         map[string]string `json:"custom,omitempty"`
}

// AgentMessage represents a message sent between agents
type AgentMessage struct {
	ID             string      `json:"id"`
	FromAgent      string      `json:"from_agent"`
	ToAgent        string      `json:"to_agent,omitempty"` // Empty means broadcast
	Type           string      `json:"type"`
	Payload        interface{} `json:"payload"`
	Timestamp      time.Time   `json:"timestamp"`
	ConversationID string      `json:"conversation_id"`
}

// KafkaMessage wraps messages for Kafka transport
type KafkaMessage struct {
	Topic   string          `json:"topic"`
	Key     string          `json:"key"`
	Payload json.RawMessage `json:"payload"`
}

// ToJSON converts a message to JSON bytes
func (m *ChatMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates a ChatMessage from JSON bytes
func (m *ChatMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// ToJSON converts an AgentMessage to JSON bytes
func (m *AgentMessage) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON creates an AgentMessage from JSON bytes
func (m *AgentMessage) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}
