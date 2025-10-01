package agent

import (
	"log"

	"philoking/internal/config"
	"philoking/internal/conversation"
	"philoking/internal/kafka"
)

// Factory creates agents from configuration
type Factory struct {
	kafkaClient         *kafka.Client
	conversationManager *conversation.Manager
}

// NewFactory creates a new agent factory
func NewFactory(kafkaClient *kafka.Client, convManager *conversation.Manager) *Factory {
	return &Factory{
		kafkaClient:         kafkaClient,
		conversationManager: convManager,
	}
}

// CreateNaturalAgents creates natural conversation agents from configuration
func (f *Factory) CreateNaturalAgents(agentConfigs []config.NaturalAgentConfig) []Agent {
	var agents []Agent

	for _, agentConfig := range agentConfigs {
		if !agentConfig.IsEnabled {
			log.Printf("Skipping disabled agent: %s", agentConfig.Name)
			continue
		}

		agent := f.createNaturalAgent(agentConfig)
		if agent != nil {
			agents = append(agents, agent)
			log.Printf("Created agent: %s - %s", agentConfig.Name, agentConfig.Description)
		}
	}

	return agents
}

// createNaturalAgent creates a single natural agent from configuration
func (f *Factory) createNaturalAgent(agentConfig config.NaturalAgentConfig) Agent {
	// Validate required fields
	if agentConfig.ID == "" {
		log.Printf("Warning: Agent missing ID, skipping")
		return nil
	}

	if agentConfig.Name == "" {
		log.Printf("Warning: Agent %s missing name, using ID", agentConfig.ID)
		agentConfig.Name = agentConfig.ID
	}

	// Set default response chance if not specified
	if agentConfig.ResponseChance == 0 {
		agentConfig.ResponseChance = 0.7
	}

	// Create the natural agent
	naturalAgent := NewNaturalAgent(
		agentConfig.ID,
		agentConfig.Name,
		f.kafkaClient,
		f.conversationManager,
		agentConfig.ResponseChance,
	)

	return naturalAgent
}

// CreateLegacyAgents creates the original echo and LLM agents
func (f *Factory) CreateLegacyAgents(agentsConfig config.AgentsConfig) []Agent {
	var agents []Agent

	// Create Echo Agent
	echoAgent := NewEchoAgent(f.kafkaClient)
	agents = append(agents, echoAgent)
	log.Printf("Created legacy agent: Echo Agent")

	// Create LLM Agent
	llmAgent := NewLLMAgent(f.kafkaClient, agentsConfig)
	agents = append(agents, llmAgent)
	log.Printf("Created legacy agent: LLM Agent")

	return agents
}

// RegisterAgentsInConversationFlow registers agents in the conversation flow
func (f *Factory) RegisterAgentsInConversationFlow(flowManager *conversation.FlowManager, agentConfigs []config.NaturalAgentConfig) {
	for _, agentConfig := range agentConfigs {
		if !agentConfig.IsEnabled {
			continue
		}

		flowManager.RegisterParticipant(
			agentConfig.ID,
			agentConfig.Name,
			"agent",
		)
	}
}
