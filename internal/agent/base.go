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
	id           string
	name         string
	kafkaClient  *kafka.Client
	handlers     map[types.MessageType]MessageHandler
	capabilities []string
	running      bool
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name string, kafkaClient *kafka.Client) *BaseAgent {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseAgent{
		id:           id,
		name:         name,
		kafkaClient:  kafkaClient,
		handlers:     make(map[types.MessageType]MessageHandler),
		capabilities: []string{},
		ctx:          ctx,
		cancel:       cancel,
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

// GetCapabilities returns the agent's capabilities
func (a *BaseAgent) GetCapabilities() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return append([]string{}, a.capabilities...)
}

// AddCapability adds a capability to the agent
func (a *BaseAgent) AddCapability(capability string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.capabilities = append(a.capabilities, capability)
}

// SetHandler sets a message handler for a specific message type
func (a *BaseAgent) SetHandler(msgType types.MessageType, handler MessageHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.handlers[msgType] = handler
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

	// Start listening for chat messages
	go func() {
		if err := a.kafkaClient.SubscribeToChatMessagesWithGroup(ctx, "philoking-agent-"+a.id, func(msg *types.ChatMessage) error {
			return a.ProcessMessage(ctx, msg)
		}); err != nil {
			log.Printf("Agent %s error subscribing to chat messages: %v", a.id, err)
		}
	}()

	// Start listening for agent messages
	go func() {
		if err := a.kafkaClient.SubscribeToAgentMessages(ctx, "agent-input", func(msg *types.AgentMessage) error {
			return a.ProcessAgentMessage(ctx, msg)
		}); err != nil {
			log.Printf("Agent %s error subscribing to agent messages: %v", a.id, err)
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
	handler, exists := a.handlers[message.Type]
	a.mu.RUnlock()

	if !exists {
		return nil // Agent doesn't handle this message type
	}

	switch message.Type {
	case types.MessageTypeUser:
		return handler.HandleUserMessage(ctx, message)
	case types.MessageTypeSystem:
		return handler.HandleSystemMessage(ctx, message)
	default:
		return nil
	}
}

// ProcessAgentMessage handles messages from other agents
func (a *BaseAgent) ProcessAgentMessage(ctx context.Context, message *types.AgentMessage) error {
	a.mu.RLock()
	handler, exists := a.handlers[types.MessageTypeAgent]
	a.mu.RUnlock()

	if !exists {
		return nil
	}

	return handler.HandleAgentMessage(ctx, message)
}

// SendChatMessage sends a chat message to the conversation
func (a *BaseAgent) SendChatMessage(ctx context.Context, content string, conversationID string) error {
	message := &types.ChatMessage{
		ID:        uuid.New().String(),
		Type:      types.MessageTypeAgent,
		Content:   content,
		AgentID:   a.id,
		Timestamp: time.Now(),
		Metadata: types.Metadata{
			ConversationID: conversationID,
		},
	}

	log.Printf("Agent %s publishing message to Kafka: %s", a.id, content)
	return a.kafkaClient.PublishChatResponse(ctx, message)
}

// SendAgentMessage sends a message to another agent
func (a *BaseAgent) SendAgentMessage(ctx context.Context, toAgent, msgType string, payload interface{}, conversationID string) error {
	message := &types.AgentMessage{
		ID:             uuid.New().String(),
		FromAgent:      a.id,
		ToAgent:        toAgent,
		Type:           msgType,
		Payload:        payload,
		Timestamp:      time.Now(),
		ConversationID: conversationID,
	}

	return a.kafkaClient.PublishAgentMessage(ctx, message)
}

// IsRunning returns whether the agent is currently running
func (a *BaseAgent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.running
}
