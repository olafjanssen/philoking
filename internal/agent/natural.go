package agent

import (
	"context"
	"log"
	"math/rand"
	"strings"

	"philoking/internal/conversation"
	"philoking/internal/kafka"
	"philoking/internal/types"
)

// NaturalAgent is an agent that participates naturally in conversations
type NaturalAgent struct {
	*BaseAgent
	conversationManager *conversation.Manager
	personality         string
	interests           []string
	responseChance      float64
}

// NewNaturalAgent creates a new natural conversation agent
func NewNaturalAgent(id, name string, kafkaClient *kafka.Client, convManager *conversation.Manager, personality string, interests []string) *NaturalAgent {
	base := NewBaseAgent(id, name, kafkaClient)
	agent := &NaturalAgent{
		BaseAgent:           base,
		conversationManager: convManager,
		personality:         personality,
		interests:           interests,
		responseChance:      0.7, // 70% chance to respond when relevant
	}

	// Set up message handlers
	agent.SetHandler(types.MessageTypeUser, agent)
	agent.SetHandler(types.MessageTypeAgent, agent)
	agent.SetHandler(types.MessageTypeSystem, agent)

	// Add capabilities based on interests
	for _, interest := range interests {
		agent.AddCapability(interest)
	}

	return agent
}

// HandleUserMessage handles user messages with natural conversation flow
func (n *NaturalAgent) HandleUserMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("NaturalAgent %s received user message: %s", n.ID(), message.Content)

	// Check if this message is relevant to this agent
	if !n.shouldRespond(message) {
		log.Printf("NaturalAgent %s decided not to respond to: %s", n.ID(), message.Content)
		return nil
	}

	// Generate a natural response
	response := n.generateNaturalResponse(message)
	if response == "" {
		return nil // No response generated
	}

	log.Printf("NaturalAgent %s sending response: %s", n.ID(), response)

	// Send response
	return n.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// HandleAgentMessage handles messages from other agents
func (n *NaturalAgent) HandleAgentMessage(ctx context.Context, message *types.AgentMessage) error {
	log.Printf("NaturalAgent %s received agent message from %s: %s", n.ID(), message.FromAgent, message.Type)

	// Convert agent message to chat message for processing
	chatMsg := &types.ChatMessage{
		ID:        message.ID,
		Type:      types.MessageTypeAgent,
		Content:   message.Type, // Use message type as content for now
		AgentID:   message.FromAgent,
		Timestamp: message.Timestamp,
		Metadata: types.Metadata{
			ConversationID: message.ConversationID,
		},
	}

	// Check if this agent message is relevant
	if !n.shouldRespond(chatMsg) {
		return nil
	}

	// Generate response to agent message
	response := n.generateAgentResponse(message)
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
	response := n.generateSystemResponse(message)
	if response == "" {
		return nil
	}

	log.Printf("NaturalAgent %s responding to system: %s", n.ID(), response)

	return n.SendChatMessage(ctx, response, message.Metadata.ConversationID)
}

// shouldRespond determines if this agent should respond to a message
func (n *NaturalAgent) shouldRespond(message *types.ChatMessage) bool {
	// Check relevance based on interests and personality
	relevant := n.conversationManager.IsRelevantToAgent(message, n.ID(), n.interests, n.personality)

	if !relevant {
		return false
	}

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

// generateNaturalResponse generates a natural response to a message
func (n *NaturalAgent) generateNaturalResponse(message *types.ChatMessage) string {
	// Personality-based responses
	switch n.personality {
	case "curious":
		return n.generateCuriousResponse(message)
	case "helpful":
		return n.generateHelpfulResponse(message)
	case "social":
		return n.generateSocialResponse(message)
	case "technical":
		return n.generateTechnicalResponse(message)
	case "philosophical":
		return n.generatePhilosophicalResponse(message)
	default:
		return n.generateDefaultResponse(message)
	}
}

// generateCuriousResponse generates responses for curious agents
func (n *NaturalAgent) generateCuriousResponse(message *types.ChatMessage) string {
	content := message.Content

	curiousResponses := []string{
		"That's interesting! Can you tell me more about that?",
		"I'm curious about that. What made you think of it?",
		"Fascinating! I'd love to hear more details.",
		"That's a great point. How did you come to that conclusion?",
		"I'm intrigued by that. Could you elaborate?",
	}

	// Check for questions
	if strings.Contains(content, "?") {
		return "That's a great question! " + curiousResponses[rand.Intn(len(curiousResponses))]
	}

	// Check for statements
	if strings.Contains(content, "I") || strings.Contains(content, "we") || strings.Contains(content, "my") {
		return "I find that really interesting! " + curiousResponses[rand.Intn(len(curiousResponses))]
	}

	return curiousResponses[rand.Intn(len(curiousResponses))]
}

// generateHelpfulResponse generates responses for helpful agents
func (n *NaturalAgent) generateHelpfulResponse(message *types.ChatMessage) string {
	content := strings.ToLower(message.Content)

	helpfulResponses := []string{
		"I'd be happy to help with that!",
		"That sounds like something I can assist with.",
		"Let me see if I can provide some guidance on that.",
		"I'm here to help! What specifically would you like to know?",
		"That's a common challenge. I might have some suggestions.",
	}

	// Check for help requests
	if strings.Contains(content, "help") || strings.Contains(content, "problem") || strings.Contains(content, "issue") {
		return helpfulResponses[rand.Intn(len(helpfulResponses))]
	}

	// Check for questions
	if strings.Contains(content, "?") {
		return "I'd be glad to help answer that! " + helpfulResponses[rand.Intn(len(helpfulResponses))]
	}

	return helpfulResponses[rand.Intn(len(helpfulResponses))]
}

// generateSocialResponse generates responses for social agents
func (n *NaturalAgent) generateSocialResponse(message *types.ChatMessage) string {
	content := strings.ToLower(message.Content)

	socialResponses := []string{
		"That's really cool! I love hearing about that kind of thing.",
		"Nice! I'm always interested in what others are thinking about.",
		"That sounds awesome! I'd love to chat more about it.",
		"I'm really enjoying this conversation!",
		"That's a great perspective! I hadn't thought of it that way.",
	}

	// Check for greetings
	if strings.Contains(content, "hello") || strings.Contains(content, "hi") || strings.Contains(content, "hey") {
		return "Hello there! " + socialResponses[rand.Intn(len(socialResponses))]
	}

	return socialResponses[rand.Intn(len(socialResponses))]
}

// generateTechnicalResponse generates responses for technical agents
func (n *NaturalAgent) generateTechnicalResponse(message *types.ChatMessage) string {
	content := strings.ToLower(message.Content)

	technicalResponses := []string{
		"From a technical perspective, that's quite interesting.",
		"I can see the technical implications of what you're saying.",
		"That raises some interesting technical questions.",
		"I'd like to explore the technical aspects of that.",
		"From an engineering standpoint, that's worth considering.",
	}

	// Check for technical keywords
	technicalKeywords := []string{"code", "programming", "software", "system", "algorithm", "data", "function", "method"}
	for _, keyword := range technicalKeywords {
		if strings.Contains(content, keyword) {
			return technicalResponses[rand.Intn(len(technicalResponses))]
		}
	}

	return technicalResponses[rand.Intn(len(technicalResponses))]
}

// generatePhilosophicalResponse generates responses for philosophical agents
func (n *NaturalAgent) generatePhilosophicalResponse(message *types.ChatMessage) string {
	content := strings.ToLower(message.Content)

	philosophicalResponses := []string{
		"That's a profound observation. It makes me think about the deeper meaning.",
		"I find myself pondering the philosophical implications of what you've said.",
		"That touches on some fundamental questions about existence and purpose.",
		"I'm intrigued by the philosophical dimensions of your statement.",
		"That raises some interesting questions about the nature of reality.",
	}

	// Check for philosophical keywords
	philosophicalKeywords := []string{"think", "believe", "feel", "meaning", "purpose", "life", "existence", "truth", "reality"}
	for _, keyword := range philosophicalKeywords {
		if strings.Contains(content, keyword) {
			return philosophicalResponses[rand.Intn(len(philosophicalResponses))]
		}
	}

	return philosophicalResponses[rand.Intn(len(philosophicalResponses))]
}

// generateDefaultResponse generates a default response
func (n *NaturalAgent) generateDefaultResponse(message *types.ChatMessage) string {
	defaultResponses := []string{
		"That's interesting!",
		"I see what you mean.",
		"That's a good point.",
		"I hadn't thought of it that way.",
		"That's worth considering.",
	}

	return defaultResponses[rand.Intn(len(defaultResponses))]
}

// generateAgentResponse generates a response to another agent
func (n *NaturalAgent) generateAgentResponse(message *types.AgentMessage) string {
	agentResponses := []string{
		"I agree with that perspective.",
		"That's a good point from " + message.FromAgent + ".",
		"I'd like to add to what " + message.FromAgent + " said.",
		"That's interesting, " + message.FromAgent + ".",
		"I have a different take on that.",
	}

	return agentResponses[rand.Intn(len(agentResponses))]
}

// generateSystemResponse generates a response to system messages
func (n *NaturalAgent) generateSystemResponse(message *types.ChatMessage) string {
	systemResponses := []string{
		"I understand the system message.",
		"Got it, thanks for the update.",
		"I'll keep that in mind.",
		"Understood.",
		"Noted.",
	}

	return systemResponses[rand.Intn(len(systemResponses))]
}

// SetResponseChance sets the probability that this agent will respond to relevant messages
func (n *NaturalAgent) SetResponseChance(chance float64) {
	if chance >= 0.0 && chance <= 1.0 {
		n.responseChance = chance
	}
}

// GetPersonality returns the agent's personality
func (n *NaturalAgent) GetPersonality() string {
	return n.personality
}

// GetInterests returns the agent's interests
func (n *NaturalAgent) GetInterests() []string {
	return n.interests
}
