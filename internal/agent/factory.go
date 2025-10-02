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

// CreateAgents creates agents from configuration based on their type
func (f *Factory) CreateAgents(agentConfigs []config.AgentConfig, agentsConfig config.AgentsConfig) []Agent {
	var agents []Agent

	for _, agentConfig := range agentConfigs {
		if !agentConfig.IsEnabled {
			log.Printf("Skipping disabled agent: %s", agentConfig.Name)
			continue
		}

		agent := f.createAgent(agentConfig, agentsConfig)
		if agent != nil {
			agents = append(agents, agent)
			log.Printf("Created %s agent: %s - %s", agentConfig.Type, agentConfig.Name, agentConfig.Description)
		}
	}

	return agents
}

// createAgent creates a single agent from configuration based on its type
func (f *Factory) createAgent(agentConfig config.AgentConfig, agentsConfig config.AgentsConfig) Agent {
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

	// Create agent based on type
	switch agentConfig.Type {
	case "llm":
		return f.createLLMAgent(agentConfig, agentsConfig)
	case "echo":
		return f.createEchoAgent(agentConfig)
	default:
		log.Printf("Warning: Unknown agent type '%s' for agent %s, skipping", agentConfig.Type, agentConfig.ID)
		return nil
	}
}

// createLLMAgent creates an LLM agent
func (f *Factory) createLLMAgent(agentConfig config.AgentConfig, agentsConfig config.AgentsConfig) Agent {
	return NewLLMAgent(agentConfig.ID, agentConfig.Name, f.kafkaClient, agentsConfig, agentConfig.ResponseChance, f.conversationManager)
}

// createEchoAgent creates an echo agent
func (f *Factory) createEchoAgent(agentConfig config.AgentConfig) Agent {
	return NewEchoAgent(agentConfig.ID, agentConfig.Name, f.kafkaClient, agentConfig.ResponseChance, f.conversationManager)
}

// RegisterAgentsInConversationFlow registers agents in the conversation flow
func (f *Factory) RegisterAgentsInConversationFlow(flowManager *conversation.FlowManager, agentConfigs []config.AgentConfig) {
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
