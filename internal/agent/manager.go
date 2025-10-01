package agent

import (
	"context"
	"fmt"
	"log"
	"sync"

	"philoking/internal/config"
	"philoking/internal/kafka"
)

// Manager manages all agents in the system
type Manager struct {
	agents      map[string]Agent
	kafkaClient *kafka.Client
	config      config.AgentsConfig
	mu          sync.RWMutex
}

// NewManager creates a new agent manager
func NewManager(kafkaClient *kafka.Client, config config.AgentsConfig) *Manager {
	return &Manager{
		agents:      make(map[string]Agent),
		kafkaClient: kafkaClient,
		config:      config,
	}
}

// RegisterAgent registers a new agent with the manager
func (m *Manager) RegisterAgent(agent Agent) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.agents[agent.ID()]; exists {
		return fmt.Errorf("agent with ID %s already registered", agent.ID())
	}

	m.agents[agent.ID()] = agent
	log.Printf("Registered agent: %s (%s)", agent.ID(), agent.Name())
	return nil
}

// Start starts all registered agents
func (m *Manager) Start(ctx context.Context) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(m.agents))

	for _, agent := range m.agents {
		wg.Add(1)
		go func(a Agent) {
			defer wg.Done()
			if err := a.Start(ctx); err != nil {
				errors <- fmt.Errorf("failed to start agent %s: %w", a.ID(), err)
			}
		}(agent)
	}

	// Wait for all agents to start
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Check for errors
	for err := range errors {
		if err != nil {
			return err
		}
	}

	log.Printf("All %d agents started successfully", len(m.agents))
	return nil
}

// Stop stops all registered agents
func (m *Manager) Stop() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var wg sync.WaitGroup
	errors := make(chan error, len(m.agents))

	for _, agent := range m.agents {
		wg.Add(1)
		go func(a Agent) {
			defer wg.Done()
			if err := a.Stop(); err != nil {
				errors <- fmt.Errorf("failed to stop agent %s: %w", a.ID(), err)
			}
		}(agent)
	}

	// Wait for all agents to stop
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Check for errors
	for err := range errors {
		if err != nil {
			log.Printf("Error stopping agent: %v", err)
		}
	}

	log.Printf("All agents stopped")
	return nil
}

// GetAgent returns an agent by ID
func (m *Manager) GetAgent(id string) (Agent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	agent, exists := m.agents[id]
	return agent, exists
}

// ListAgents returns a list of all registered agents
func (m *Manager) ListAgents() []Agent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	agents := make([]Agent, 0, len(m.agents))
	for _, agent := range m.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetConfig returns the agent configuration
func (m *Manager) GetConfig() config.AgentsConfig {
	return m.config
}

