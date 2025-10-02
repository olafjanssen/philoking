package agent

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"philoking/internal/kafka"
	"philoking/internal/types"

	"github.com/google/uuid"
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	id          string
	name        string
	kafkaClient *kafka.Client
	handler     MessageHandler
	running     bool
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name string, kafkaClient *kafka.Client) *BaseAgent {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseAgent{
		id:          id,
		name:        name,
		kafkaClient: kafkaClient,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// ID returns the agent's unique identifier
func (a *BaseAgent) ID() string {
	return a.id
}

// Name returns the agent's human-readable name
func (a *BaseAgent) Name() string {
	return a.name
}

// SetHandler sets the message handler for this agent
func (a *BaseAgent) SetHandler(handler MessageHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handler = handler
}

// Start begins the agent's processing loop
func (a *BaseAgent) Start(ctx context.Context) error {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return fmt.Errorf("agent %s is already running", a.id)
	}
	a.running = true
	a.mu.Unlock()

	// Start listening for all chat messages
	go func() {
		if err := a.kafkaClient.SubscribeToMessages(ctx, "philoking-agent-"+a.id, func(msg *types.ChatMessage) error {
			return a.ProcessMessage(ctx, msg)
		}); err != nil {
			log.Printf("Agent %s error subscribing to messages: %v", a.id, err)
		}
	}()

	log.Printf("Agent %s (%s) started", a.id, a.name)
	return nil
}

// Stop gracefully stops the agent
func (a *BaseAgent) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	a.cancel()
	a.running = false
	log.Printf("Agent %s stopped", a.id)
	return nil
}

// ProcessMessage handles incoming chat messages
func (a *BaseAgent) ProcessMessage(ctx context.Context, message *types.ChatMessage) error {
	a.mu.RLock()
	handler := a.handler
	a.mu.RUnlock()

	if handler == nil {
		return nil // No handler set
	}

	// All messages are processed the same way - no distinction between user and agent
	return handler.HandleMessage(ctx, message)
}

// SendMessage sends a message to the global conversation
func (a *BaseAgent) SendMessage(ctx context.Context, content string, conversationID string) error {
	message := &types.ChatMessage{
		ID:        uuid.New().String(),
		Type:      types.MessageTypeAgent,
		Content:   content,
		AgentID:   a.id,
		Timestamp: time.Now(),
		Metadata: types.Metadata{
			ConversationID: conversationID,
			FromAgent:      a.name, // Human-readable name
		},
	}

	log.Printf("Agent %s publishing message to Kafka: %s", a.id, content)
	return a.kafkaClient.PublishMessage(ctx, message)
}

// IsRunning returns whether the agent is currently running
func (a *BaseAgent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}
