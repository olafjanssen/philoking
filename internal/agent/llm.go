package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"philoking/internal/config"
	"philoking/internal/kafka"
	"philoking/internal/types"
)

// LLMAgent is an agent that uses an LLM API to generate responses
type LLMAgent struct {
	*BaseAgent
	config config.AgentsConfig
	client *http.Client
}

// LLMRequest represents a request to the LLM API (OpenAI format)
type LLMRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

// Message represents a message in the LLM conversation
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMResponse represents the response from the LLM API (OpenAI format)
type LLMResponse struct {
	Choices []Choice `json:"choices"`
}

// Choice represents a choice in the LLM response
type Choice struct {
	Message Message `json:"message"`
}

// OllamaRequest represents a request to the Ollama API
type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []Message     `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  OllamaOptions `json:"options,omitempty"`
}

// OllamaOptions represents options for Ollama requests
type OllamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	TopK        int     `json:"top_k,omitempty"`
}

// OllamaResponse represents the response from the Ollama API
type OllamaResponse struct {
	Model     string  `json:"model"`
	Message   Message `json:"message"`
	Done      bool    `json:"done"`
	CreatedAt string  `json:"created_at"`
}

// NewLLMAgent creates a new LLM agent
func NewLLMAgent(id, name string, kafkaClient *kafka.Client, config config.AgentsConfig) *LLMAgent {
	base := NewBaseAgent(id, name, kafkaClient)
	agent := &LLMAgent{
		BaseAgent: base,
		config:    config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Set the message handler
	agent.SetHandler(agent)

	return agent
}

// HandleMessage handles all incoming messages (unified)
func (l *LLMAgent) HandleMessage(ctx context.Context, message *types.ChatMessage) error {
	log.Printf("LLMAgent received message from %s: %s", message.AgentID, message.Content)

	// For now, we'll use a simple response since we don't have a real LLM API key
	// In a real implementation, you would call the LLM API here
	response, err := l.generateResponse(ctx, message.Content, message.Metadata.ConversationID)
	if err != nil {
		log.Printf("Error generating LLM response: %v", err)
		response = "I'm sorry, I'm having trouble processing your request right now. Please try again later."
	}

	log.Printf("LLMAgent sending response: %s", response)

	// Send response
	return l.SendMessage(ctx, response, message.Metadata.ConversationID)
}

// generateResponse generates a response using the configured LLM provider
func (l *LLMAgent) generateResponse(ctx context.Context, userMessage, conversationID string) (string, error) {
	// Determine which provider to use
	provider := l.config.Provider
	if provider == "" {
		provider = "ollama" // Default to Ollama
	}

	switch provider {
	case "ollama":
		return l.generateOllamaResponse(ctx, userMessage)
	case "openai":
		return l.generateOpenAIResponse(ctx, userMessage)
	default:
		// Fallback to mock response
		return l.generateMockResponse(userMessage), nil
	}
}

// generateOllamaResponse generates a response using Ollama
func (l *LLMAgent) generateOllamaResponse(ctx context.Context, userMessage string) (string, error) {
	// Prepare the request
	reqBody := OllamaRequest{
		Model: l.config.Model,
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful AI assistant participating in a multi-agent chat system. Be conversational and helpful.",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		Stream: false,
		Options: OllamaOptions{
			Temperature: 0.7,
			TopP:        0.9,
			TopK:        40,
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal Ollama request: %w", err)
	}

	// Create HTTP request
	url := l.config.OllamaURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create Ollama request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Make the request
	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make Ollama request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	return ollamaResp.Message.Content, nil
}

// generateOpenAIResponse generates a response using OpenAI API
func (l *LLMAgent) generateOpenAIResponse(ctx context.Context, userMessage string) (string, error) {
	// If no API key is configured, return a mock response
	if l.config.LLMAPIKey == "" {
		return l.generateMockResponse(userMessage), nil
	}

	// Prepare the request
	reqBody := LLMRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a helpful AI assistant participating in a multi-agent chat system. Be conversational and helpful.",
			},
			{
				Role:    "user",
				Content: userMessage,
			},
		},
		MaxTokens:   150,
		Temperature: 0.7,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OpenAI request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", l.config.LLMURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create OpenAI request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.config.LLMAPIKey)

	// Make the request
	resp, err := l.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make OpenAI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var llmResp LLMResponse
	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("failed to decode OpenAI response: %w", err)
	}

	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI response")
	}

	return llmResp.Choices[0].Message.Content, nil
}

// generateMockResponse generates a mock response when no LLM API is available
func (l *LLMAgent) generateMockResponse(userMessage string) string {
	responses := []string{
		"That's an interesting point! I'm the LLM Agent, and I'd love to discuss this further.",
		"I understand what you're saying. From an AI perspective, this is quite fascinating.",
		"That's a great question! While I'm currently in mock mode, I'd normally provide a more detailed response.",
		"I appreciate you sharing that with me. In a real implementation, I'd use advanced language models to respond.",
		"That's thought-provoking! I'm designed to engage in meaningful conversations about various topics.",
	}

	// Simple hash-based selection for consistency
	hash := 0
	for _, char := range userMessage {
		hash += int(char)
	}

	return responses[hash%len(responses)] + " (Mock response - configure LLM_API_KEY for real AI responses)"
}
