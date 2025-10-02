package agent

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"philoking/internal/conversation"
	"philoking/internal/kafka"
	"philoking/internal/types"

	"github.com/google/uuid"
)

// BaseAgent provides common functionality for all agents
type BaseAgent struct {
	id             string
	name           string
	kafkaClient    *kafka.Client
	handler        MessageHandler
	running        bool
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	responseChance float64
	convManager    *conversation.Manager
}

// NewBaseAgent creates a new base agent
func NewBaseAgent(id, name string, kafkaClient *kafka.Client, responseChance float64, convManager *conversation.Manager) *BaseAgent {
	ctx, cancel := context.WithCancel(context.Background())
	return &BaseAgent{
		id:             id,
		name:           name,
		kafkaClient:    kafkaClient,
		ctx:            ctx,
		cancel:         cancel,
		responseChance: responseChance,
		convManager:    convManager,
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
	responseChance := a.responseChance
	a.mu.RUnlock()

	if handler == nil {
		return nil // No handler set
	}

	// Don't respond to our own messages
	if message.AgentID == a.id {
		return nil
	}

	// Add message to conversation history
	if a.convManager != nil {
		a.convManager.AddMessage(message.Metadata.ConversationID, message)
	}

	// Check response chance
	if !a.shouldRespond(responseChance) {
		log.Printf("Agent %s decided not to respond (chance: %.2f)", a.name, responseChance)
		return nil
	}

	// Wait 20 seconds before responding to allow more messages to accumulate
	log.Printf("Agent %s will respond in 20 seconds...", a.name)
	time.Sleep(20 * time.Second)

	// Check if we're still running after the delay
	a.mu.RLock()
	running := a.running
	a.mu.RUnlock()

	if !running {
		log.Printf("Agent %s stopped during delay, not responding", a.name)
		return nil
	}

	// Process the message with full conversation context
	return handler.HandleMessage(ctx, message)
}

// shouldRespond determines if this agent should respond based on response chance
func (a *BaseAgent) shouldRespond(responseChance float64) bool {
	if responseChance <= 0 {
		return false
	}
	if responseChance >= 1 {
		return true
	}

	// Use time-based randomness for more natural distribution
	seed := time.Now().UnixNano() + int64(len(a.id))
	rand.Seed(seed)
	return rand.Float64() < responseChance
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
