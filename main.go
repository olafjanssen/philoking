package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"philoking/internal/agent"
	"philoking/internal/config"
	"philoking/internal/conversation"
	"philoking/internal/kafka"
	"philoking/internal/web"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Kafka producer and consumer
	kafkaClient, err := kafka.NewClient(cfg.Kafka)
	if err != nil {
		log.Fatalf("Failed to initialize Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	// Initialize conversation manager
	convManager := conversation.NewManager()
	flowManager := conversation.NewFlowManager(kafkaClient, convManager)

	// Start conversation flow
	conversationID := "main-conversation"
	if err := flowManager.StartConversationFlow(ctx, conversationID); err != nil {
		log.Fatalf("Failed to start conversation flow: %v", err)
	}

	// Initialize agent factory
	agentFactory := agent.NewFactory(kafkaClient, convManager)

	// Create agents from configuration
	allAgents := agentFactory.CreateAgents(cfg.GetEnabledAgents(), cfg.Agents)

	// Register agents in conversation flow
	agentFactory.RegisterAgentsInConversationFlow(flowManager, cfg.Agents.Agents)

	// Initialize agent manager
	agentManager := agent.NewManager(kafkaClient, cfg.Agents)

	// Register all agents
	for _, agent := range allAgents {
		if err := agentManager.RegisterAgent(agent); err != nil {
			log.Fatalf("Failed to register agent %s: %v", agent.ID(), err)
		}
	}

	// Start agents
	if err := agentManager.Start(ctx); err != nil {
		log.Fatalf("Failed to start agents: %v", err)
	}

	// Start web server
	webServer := web.NewServer(cfg.Web, kafkaClient)
	go func() {
		if err := webServer.Start(); err != nil {
			log.Fatalf("Failed to start web server: %v", err)
		}
	}()

	// Display startup information
	log.Println("🎉 Multi-Agent Conversation System Started!")
	log.Println("📝 Conversation ID:", conversationID)
	log.Println("🤖 Active Agents:")

	enabledAgents := cfg.GetEnabledAgents()
	for _, agentConfig := range enabledAgents {
		log.Printf("   - %s (%s) - %s", agentConfig.Name, agentConfig.Type, agentConfig.Description)
	}

	log.Printf("📊 Total Agents: %d", len(allAgents))
	log.Println("🌐 Web Interface: http://localhost:8080")
	log.Println("💬 Start chatting and watch the multi-agent conversation!")
	log.Println("⚙️  Configure agents in config.yaml")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
