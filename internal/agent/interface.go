package agent

import (
	"context"
	"philoking/internal/types"
)

// Agent represents a chat agent that can process messages
type Agent interface {
	// ID returns the unique identifier for this agent
	ID() string

	// Name returns a human-readable name for this agent
	Name() string

	// Start begins the agent's processing loop
	Start(ctx context.Context) error

	// Stop gracefully stops the agent
	Stop() error

	// ProcessMessage handles incoming chat messages
	ProcessMessage(ctx context.Context, message *types.ChatMessage) error
}

// MessageHandler defines how agents handle messages
type MessageHandler interface {
	HandleMessage(ctx context.Context, message *types.ChatMessage) error
}

// LLMProvider defines the interface for LLM services
type LLMProvider interface {
	GenerateResponse(ctx context.Context, prompt string, conversation []types.ChatMessage) (string, error)
}
