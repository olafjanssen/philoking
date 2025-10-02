package agent

import (
	"context"
	"log"
	"strings"

	"philoking/internal/kafka"
	"philoking/internal/types"
)

// EchoAgent is a simple agent that echoes messages
type EchoAgent struct {
	*BaseAgent
}

// NewEchoAgent creates a new echo agent
func NewEchoAgent(kafkaClient *kafka.Client) *EchoAgent {
	base := NewBaseAgent("echo-agent", "Echo Agent", kafkaClient)
	agent := &EchoAgent{BaseAgent: base}

	// Set the message handler
	agent.SetHandler(agent)

	return agent
}

// HandleMessage handles all incoming messages (unified)
func (e *EchoAgent) HandleMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("EchoAgent received message from %s: %s", message.AgentID, message.Content)

	// Simple echo with a twist
	response := "Echo: " + message.Content

	// Add some variety based on content
	if strings.Contains(strings.ToLower(message.Content), "hello") {
		response = "Hello! I'm the Echo Agent. You said: " + message.Content
	} else if strings.Contains(strings.ToLower(message.Content), "bye") {
		response = "Goodbye! Thanks for chatting. You said: " + message.Content
	} else if strings.Contains(strings.ToLower(message.Content), "help") {
		response = "I'm the Echo Agent! I simply echo back what you say. Try saying something!"
	}

	log.Printf("EchoAgent sending response: %s", response)

	// Send response
	return e.SendMessage(ctx, response, message.Metadata.ConversationID)
}
