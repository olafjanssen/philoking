package conversation

import (
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
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
	mu           sync.RWMutex
}

// Participant represents a conversation participant
type Participant struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Type     string    `json:"type"` // "user", "agent", "system"
	IsActive bool      `json:"is_active"`
	LastSeen time.Time `json:"last_seen"`
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
func (m *Manager) AddParticipant(conversationID, participantID, name, participantType string) {
	conv := m.GetOrCreateConversation(conversationID)

	conv.mu.Lock()
	defer conv.mu.Unlock()

	conv.Participants[participantID] = &Participant{
		ID:       participantID,
		Name:     name,
		Type:     participantType,
		IsActive: true,
		LastSeen: time.Now(),
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
// Simplified version - all messages are potentially relevant
func (m *Manager) IsRelevantToAgent(message *types.ChatMessage, agentID string, capabilities []string, personality string) bool {
	// System messages are always relevant
	if message.Type == types.MessageTypeSystem {
		return true
	}

	// Check if message is a direct reply to this agent
	if message.Metadata.ReplyTo == agentID {
		return true
	}

	// All other messages are potentially relevant
	// The agent's response chance will determine if it actually responds
	return true
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

// GetConversationContext gets the current conversation context
func (m *Manager) GetConversationContext(conversationID string) *Conversation {
	return m.GetOrCreateConversation(conversationID)
}
