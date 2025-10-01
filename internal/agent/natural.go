package agent

import (
	"context"
	"log"
	"math/rand"

	"philoking/internal/conversation"
	"philoking/internal/kafka"
	"philoking/internal/types"
)

// NaturalAgent is a simple agent that participates in conversations
type NaturalAgent struct {
	*BaseAgent
	conversationManager *conversation.Manager
	responseChance      float64
}

// NewNaturalAgent creates a new natural conversation agent
func NewNaturalAgent(id, name string, kafkaClient *kafka.Client, convManager *conversation.Manager, responseChance float64) *NaturalAgent {
	base := NewBaseAgent(id, name, kafkaClient)
	agent := &NaturalAgent{
		BaseAgent:           base,
		conversationManager: convManager,
		responseChance:      responseChance,
	}

	// Set up message handlers
	agent.SetHandler(types.MessageTypeUser, agent)
	agent.SetHandler(types.MessageTypeAgent, agent)
	agent.SetHandler(types.MessageTypeSystem, agent)

	return agent
}

// HandleUserMessage handles user messages
func (n *NaturalAgent) HandleUserMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("NaturalAgent %s received user message: %s", n.ID(), message.Content)

	// Check if this agent should respond
	if !n.shouldRespond(message) {
		log.Printf("NaturalAgent %s decided not to respond to: %s", n.ID(), message.Content)
		return nil
	}

	// Generate a simple response
	response := n.generateResponse(message)
	if response == "" {
		return nil // No response generated
	}

	log.Printf("NaturalAgent %s sending response: %s", n.ID(), response)

	// Send response
	return n.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// HandleAgentChatMessage handles agent messages that come as ChatMessage
func (n *NaturalAgent) HandleAgentChatMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("NaturalAgent %s received agent chat message from %s: %s", n.ID(), message.AgentID, message.Content)

	// Check if this agent should respond
	if !n.shouldRespond(message) {
		log.Printf("NaturalAgent %s decided not to respond to agent message: %s", n.ID(), message.Content)
		return nil
	}

	// Generate a simple response
	response := n.generateResponse(message)
	if response == "" {
		return nil // No response generated
	}

	log.Printf("NaturalAgent %s responding to agent %s: %s", n.ID(), message.AgentID, response)

	// Send response
	return n.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// HandleAgentMessage handles messages from other agents (legacy)
func (n *NaturalAgent) HandleAgentMessage(ctx context.Context, message *types.AgentMessage) error {
	log.Printf("NaturalAgent %s received agent message from %s: %s", n.ID(), message.FromAgent, message.Type)

	// Convert agent message to chat message for processing
	chatMsg := &types.ChatMessage{
		ID:        message.ID,
		Type:      types.MessageTypeAgent,
		Content:   message.Type,
		AgentID:   message.FromAgent,
		Timestamp: message.Timestamp,
		Metadata: types.Metadata{
			ConversationID: message.ConversationID,
		},
	}

	// Check if this agent should respond
	if !n.shouldRespond(chatMsg) {
		return nil
	}

	// Generate response to agent message
	response := n.generateResponse(chatMsg)
	if response == "" {
		return nil
	}

	log.Printf("NaturalAgent %s responding to agent %s: %s", n.ID(), message.FromAgent, response)

	return n.SendChatMessage(ctx, response, message.ConversationID)
}

// HandleSystemMessage handles system messages
func (n *NaturalAgent) HandleSystemMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("NaturalAgent %s received system message: %s", n.ID(), message.Content)

	// System messages are always relevant
	response := n.generateResponse(message)
	if response == "" {
		return nil
	}

	log.Printf("NaturalAgent %s responding to system: %s", n.ID(), response)

	return n.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// shouldRespond determines if this agent should respond to a message
func (n *NaturalAgent) shouldRespond(message *types.ChatMessage) bool {
	// Don't respond to our own messages
	if message.AgentID == n.ID() {
		return false
	}

	// Apply response chance (makes conversation more natural)
	if rand.Float64() > n.responseChance {
		return false
	}

	// Check if we've responded recently (avoid spam)
	recentMessages := n.conversationManager.GetRecentMessages(message.Metadata.ConversationID, 5)
	ourRecentResponses := 0
	for _, msg := range recentMessages {
		if msg.AgentID == n.ID() {
			ourRecentResponses++
		}
	}

	// Don't respond if we've responded to 2+ of the last 5 messages
	if ourRecentResponses >= 2 {
		return false
	}

	return true
}

// generateResponse generates a simple response to a message
func (n *NaturalAgent) generateResponse(message *types.ChatMessage) string {
	// Simple response templates
	responses := []string{
		"That's interesting!",
		"I see what you mean.",
		"That's a good point.",
		"I agree with that.",
		"That makes sense.",
		"I hadn't thought of it that way.",
		"That's worth considering.",
		"I can relate to that.",
		"That's a valid perspective.",
		"I understand what you're saying.",
	}

	// Return a random response
	return responses[rand.Intn(len(responses))]
}

// SetResponseChance sets the response chance for this agent
func (n *NaturalAgent) SetResponseChance(chance float64) {
	n.responseChance = chance
}

// GetResponseChance returns the current response chance
func (n *NaturalAgent) GetResponseChance() float64 {
	return n.responseChance
}
