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

	// ProcessAgentMessage handles messages from other agents
	ProcessAgentMessage(ctx context.Context, message *types.AgentMessage) error

	// GetCapabilities returns what this agent can do
	GetCapabilities() []string
}

// MessageHandler defines how agents handle different types of messages
type MessageHandler interface {
	HandleUserMessage(ctx context.Context, message *types.ChatMessage) error
	HandleAgentMessage(ctx context.Context, message *types.AgentMessage) error
	HandleSystemMessage(ctx context.Context, message *types.ChatMessage) error
}

// LLMProvider defines the interface for LLM services
type LLMProvider interface {
	GenerateResponse(ctx context.Context, prompt string, conversation []types.ChatMessage) (string, error)
	GenerateAgentResponse(ctx context.Context, prompt string, context []types.ChatMessage) (string, error)
}

