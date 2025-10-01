package conversation

import (
	"context"
	"log"
	"strings"
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
func (f *FlowManager) RegisterParticipant(participantID, name, participantType string, capabilities []string, personality string) {
	f.participants[participantID] = &Participant{
		ID:           participantID,
		Name:         name,
		Type:         participantType,
		IsActive:     true,
		LastSeen:     time.Now(),
		Capabilities: capabilities,
		Personality:  personality,
	}

	log.Printf("Registered participant: %s (%s) - %s", name, participantType, personality)
}

// StartConversationFlow starts the natural conversation flow
func (f *FlowManager) StartConversationFlow(ctx context.Context, conversationID string) error {
	// Register the user as a participant
	f.RegisterParticipant("user", "User", "user", []string{"general", "conversation"}, "social")

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

	// Update conversation topic if this is a new topic
	f.updateConversationTopic(conversationID, message)

	// Update conversation mood
	f.updateConversationMood(conversationID, message)

	log.Printf("Conversation flow handled message: %s (type: %s, from: %s)",
		message.Content, message.Type, f.getParticipantID(message))

	return nil
}

// updateConversationTopic updates the conversation topic based on message content
func (f *FlowManager) updateConversationTopic(conversationID string, message *types.ChatMessage) {
	content := message.Content
	topic := f.detectTopic(content)

	if topic != "" {
		f.conversationManager.SetConversationTopic(conversationID, topic)
		log.Printf("Updated conversation topic to: %s", topic)
	}
}

// updateConversationMood updates the conversation mood based on message content
func (f *FlowManager) updateConversationMood(conversationID string, message *types.ChatMessage) {
	content := message.Content
	mood := f.detectMood(content)

	if mood != "" {
		f.conversationManager.SetConversationMood(conversationID, mood)
		log.Printf("Updated conversation mood to: %s", mood)
	}
}

// detectTopic detects the topic of a message
func (f *FlowManager) detectTopic(content string) string {
	content = strings.ToLower(content)

	topics := map[string][]string{
		"technology": {"code", "programming", "software", "computer", "tech", "ai", "machine learning"},
		"philosophy": {"think", "believe", "meaning", "purpose", "existence", "truth", "reality"},
		"science":    {"research", "study", "experiment", "theory", "hypothesis", "data"},
		"art":        {"creative", "artistic", "design", "beautiful", "aesthetic", "music"},
		"politics":   {"government", "policy", "election", "democracy", "rights", "law"},
		"health":     {"health", "medical", "doctor", "medicine", "wellness", "fitness"},
		"travel":     {"travel", "trip", "vacation", "journey", "adventure", "explore"},
		"food":       {"food", "cooking", "recipe", "restaurant", "meal", "taste"},
	}

	for topic, keywords := range topics {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return topic
			}
		}
	}

	return ""
}

// detectMood detects the mood of a message
func (f *FlowManager) detectMood(content string) string {
	content = strings.ToLower(content)

	moods := map[string][]string{
		"excited":   {"excited", "amazing", "awesome", "fantastic", "wow", "incredible"},
		"curious":   {"wonder", "curious", "interesting", "fascinating", "intriguing"},
		"concerned": {"worried", "concerned", "problem", "issue", "trouble", "difficult"},
		"happy":     {"happy", "joy", "pleased", "delighted", "cheerful", "glad"},
		"serious":   {"serious", "important", "critical", "urgent", "matter"},
		"casual":    {"casual", "relaxed", "easy", "simple", "chill"},
	}

	for mood, keywords := range moods {
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				return mood
			}
		}
	}

	return ""
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
		"topic":        conv.Topic,
		"mood":         conv.Mood,
		"created_at":   conv.CreatedAt,
		"updated_at":   conv.UpdatedAt,
	}

	return stats
}
