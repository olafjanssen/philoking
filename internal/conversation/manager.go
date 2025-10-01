package conversation

import (
	"math/rand"
	"strings"
	"sync"
	"time"

	"philoking/internal/types"
)

// Manager manages conversation state and context
type Manager struct {
	conversations map[string]*Conversation
	mu            sync.RWMutex
}

// Conversation represents a conversation session
type Conversation struct {
	ID           string                  `json:"id"`
	Participants map[string]*Participant `json:"participants"`
	Messages     []*types.ChatMessage    `json:"messages"`
	Topic        string                  `json:"topic,omitempty"`
	Mood         string                  `json:"mood,omitempty"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	mu           sync.RWMutex
}

// Participant represents a conversation participant
type Participant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Type         string    `json:"type"` // "user", "agent", "system"
	IsActive     bool      `json:"is_active"`
	LastSeen     time.Time `json:"last_seen"`
	Capabilities []string  `json:"capabilities,omitempty"`
	Personality  string    `json:"personality,omitempty"`
}

// NewManager creates a new conversation manager
func NewManager() *Manager {
	return &Manager{
		conversations: make(map[string]*Conversation),
	}
}

// GetOrCreateConversation gets an existing conversation or creates a new one
func (m *Manager) GetOrCreateConversation(conversationID string) *Conversation {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conv, exists := m.conversations[conversationID]; exists {
		return conv
	}

	conv := &Conversation{
		ID:           conversationID,
		Participants: make(map[string]*Participant),
		Messages:     make([]*types.ChatMessage, 0),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	m.conversations[conversationID] = conv
	return conv
}

// AddMessage adds a message to a conversation
func (m *Manager) AddMessage(conversationID string, message *types.ChatMessage) {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.Lock()
	defer conv.mu.Unlock()

	conv.Messages = append(conv.Messages, message)
	conv.UpdatedAt = time.Now()

	// Update participant last seen
	participantID := message.AgentID
	if participantID == "" {
		participantID = message.UserID
	}

	if participantID != "" {
		if participant, exists := conv.Participants[participantID]; exists {
			participant.LastSeen = time.Now()
		}
	}
}

// AddParticipant adds a participant to a conversation
func (m *Manager) AddParticipant(conversationID, participantID, name, participantType string, capabilities []string, personality string) {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.Lock()
	defer conv.mu.Unlock()

	conv.Participants[participantID] = &Participant{
		ID:           participantID,
		Name:         name,
		Type:         participantType,
		IsActive:     true,
		LastSeen:     time.Now(),
		Capabilities: capabilities,
		Personality:  personality,
	}
}

// GetRecentMessages gets recent messages from a conversation
func (m *Manager) GetRecentMessages(conversationID string, limit int) []*types.ChatMessage {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.RLock()
	defer conv.mu.RUnlock()

	if len(conv.Messages) <= limit {
		return conv.Messages
	}

	return conv.Messages[len(conv.Messages)-limit:]
}

// IsRelevantToAgent checks if a message is relevant to a specific agent
func (m *Manager) IsRelevantToAgent(message *types.ChatMessage, agentID string, capabilities []string, personality string) bool {
	// System messages are always relevant
	if message.Type == types.MessageTypeSystem {
		return true
	}

	// Check if message is a direct reply to this agent
	if message.Metadata.ReplyTo == agentID {
		return true
	}

	// Check if message contains keywords from agent capabilities
	content := strings.ToLower(message.Content)
	for _, capability := range capabilities {
		if strings.Contains(content, strings.ToLower(capability)) {
			return true
		}
	}

	// Check relevance score if available (using Custom field for now)
	if relevance, exists := message.Metadata.Custom["relevance"]; exists {
		if relevance == "high" {
			return true
		}
	}

	// Personality-based relevance
	if personality == "curious" && (strings.Contains(content, "?") || strings.Contains(content, "what") || strings.Contains(content, "how")) {
		return true
	}

	if personality == "helpful" && (strings.Contains(content, "help") || strings.Contains(content, "problem") || strings.Contains(content, "issue")) {
		return true
	}

	if personality == "social" && (strings.Contains(content, "hello") || strings.Contains(content, "hi") || strings.Contains(content, "greeting")) {
		return true
	}

	// Random chance for agents to participate (makes conversation more natural)
	// This simulates agents "overhearing" conversations
	if rand.Float64() < 0.3 { // 30% chance
		return true
	}

	return false
}

// GetActiveParticipants gets active participants in a conversation
func (m *Manager) GetActiveParticipants(conversationID string) []*Participant {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.RLock()
	defer conv.mu.RUnlock()

	var active []*Participant
	for _, participant := range conv.Participants {
		if participant.IsActive {
			active = append(active, participant)
		}
	}

	return active
}

// SetConversationTopic sets the topic of a conversation
func (m *Manager) SetConversationTopic(conversationID, topic string) {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.Lock()
	defer conv.mu.Unlock()

	conv.Topic = topic
	conv.UpdatedAt = time.Now()
}

// SetConversationMood sets the mood of a conversation
func (m *Manager) SetConversationMood(conversationID, mood string) {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.Lock()
	defer conv.mu.Unlock()

	conv.Mood = mood
	conv.UpdatedAt = time.Now()
}

// GetConversationContext gets the current conversation context
func (m *Manager) GetConversationContext(conversationID string) *Conversation {
	return m.GetOrCreateConversation(conversationID)
}
