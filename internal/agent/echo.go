package agent

import (
	"context"
	"log"
	"strings"

	"philoking/internal/kafka"
	"philoking/internal/types"
)

// EchoAgent is a simple agent that echoes user messages
type EchoAgent struct {
	*BaseAgent
}

// NewEchoAgent creates a new echo agent
func NewEchoAgent(kafkaClient *kafka.Client) *EchoAgent {
	base := NewBaseAgent("echo-agent", "Echo Agent", kafkaClient)
	agent := &EchoAgent{BaseAgent: base}

	// Set up message handlers
	agent.SetHandler(types.MessageTypeUser, agent)
	agent.AddCapability("echo")
	agent.AddCapability("simple_response")

	return agent
}

// HandleUserMessage handles user messages by echoing them
func (e *EchoAgent) HandleUserMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("EchoAgent received user message: %s", message.Content)

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
	return e.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// HandleAgentMessage handles messages from other agents
func (e *EchoAgent) HandleAgentMessage(ctx context.Context, message *types.AgentMessage) error {
	log.Printf("EchoAgent received agent message from %s: %s", message.FromAgent, message.Type)

	// Echo agent doesn't typically respond to other agents
	// but could be extended to do so
	return nil
}

// HandleSystemMessage handles system messages
func (e *EchoAgent) HandleSystemMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("EchoAgent received system message: %s", message.Content)
	return nil
}
