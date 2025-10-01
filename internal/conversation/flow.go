package conversation

import (
	"context"
	"log"
	"time"

	"philoking/internal/kafka"
	"philoking/internal/types"
)

// FlowManager manages the natural conversation flow
type FlowManager struct {
	kafkaClient         *kafka.Client
	conversationManager *Manager
	participants        map[string]*Participant
}

// NewFlowManager creates a new conversation flow manager
func NewFlowManager(kafkaClient *kafka.Client, convManager *Manager) *FlowManager {
	return &FlowManager{
		kafkaClient:         kafkaClient,
		conversationManager: convManager,
		participants:        make(map[string]*Participant),
	}
}

// RegisterParticipant registers a participant in the conversation
func (f *FlowManager) RegisterParticipant(participantID, name, participantType string) {
	f.participants[participantID] = &Participant{
		ID:       participantID,
		Name:     name,
		Type:     participantType,
		IsActive: true,
		LastSeen: time.Now(),
	}

	log.Printf("Registered participant: %s (%s)", name, participantType)
}

// StartConversationFlow starts the natural conversation flow
func (f *FlowManager) StartConversationFlow(ctx context.Context, conversationID string) error {
	// Register the user as a participant
	f.RegisterParticipant("user", "User", "user")

	// Start listening to the unified conversation topic
	go func() {
		err := f.kafkaClient.SubscribeToChatMessages(ctx, func(message *types.ChatMessage) error {
			return f.handleMessage(ctx, message, conversationID)
		})
		if err != nil {
			log.Printf("Error in conversation flow: %v", err)
		}
	}()

	log.Printf("Started conversation flow for conversation: %s", conversationID)
	return nil
}

// handleMessage handles incoming messages in the conversation flow
func (f *FlowManager) handleMessage(ctx context.Context, message *types.ChatMessage, conversationID string) error {
	// Add message to conversation history
	f.conversationManager.AddMessage(conversationID, message)

	log.Printf("Conversation flow handled message: %s (type: %s, from: %s)",
		message.Content, message.Type, f.getParticipantID(message))

	return nil
}

// getParticipantID gets the participant ID from a message
func (f *FlowManager) getParticipantID(message *types.ChatMessage) string {
	if message.AgentID != "" {
		return message.AgentID
	}
	if message.UserID != "" {
		return message.UserID
	}
	return "unknown"
}

// GetConversationStats returns statistics about the conversation
func (f *FlowManager) GetConversationStats(conversationID string) map[string]interface{} {
	conv := f.conversationManager.GetConversationContext(conversationID)

	stats := map[string]interface{}{
		"id":           conv.ID,
		"participants": len(conv.Participants),
		"messages":     len(conv.Messages),
		"created_at":   conv.CreatedAt,
		"updated_at":   conv.UpdatedAt,
	}

	return stats
}
