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

	// Initialize agent manager
	agentManager := agent.NewManager(kafkaClient, cfg.Agents)

	// Create natural conversation agents with different personalities
	curiousAgent := agent.NewNaturalAgent(
		"curious-agent",
		"Curious Agent",
		kafkaClient,
		convManager,
		"curious",
		[]string{"questions", "learning", "discovery", "science", "philosophy"},
	)

	helpfulAgent := agent.NewNaturalAgent(
		"helpful-agent",
		"Helpful Agent",
		kafkaClient,
		convManager,
		"helpful",
		[]string{"help", "support", "guidance", "problem-solving", "assistance"},
	)

	technicalAgent := agent.NewNaturalAgent(
		"technical-agent",
		"Technical Agent",
		kafkaClient,
		convManager,
		"technical",
		[]string{"programming", "technology", "software", "engineering", "code"},
	)

	philosophicalAgent := agent.NewNaturalAgent(
		"philosophical-agent",
		"Philosophical Agent",
		kafkaClient,
		convManager,
		"philosophical",
		[]string{"philosophy", "meaning", "existence", "truth", "reality", "ethics"},
	)

	// Register all agents
	agents := []agent.Agent{curiousAgent, helpfulAgent, technicalAgent, philosophicalAgent}
	for _, agent := range agents {
		if err := agentManager.RegisterAgent(agent); err != nil {
			log.Fatalf("Failed to register agent %s: %v", agent.ID(), err)
		}
	}

	// Register agents in conversation flow
	flowManager.RegisterParticipant("curious-agent", "Curious Agent", "agent",
		[]string{"questions", "learning", "discovery", "science", "philosophy"}, "curious")
	flowManager.RegisterParticipant("helpful-agent", "Helpful Agent", "agent",
		[]string{"help", "support", "guidance", "problem-solving", "assistance"}, "helpful")
	flowManager.RegisterParticipant("technical-agent", "Technical Agent", "agent",
		[]string{"programming", "technology", "software", "engineering", "code"}, "technical")
	flowManager.RegisterParticipant("philosophical-agent", "Philosophical Agent", "agent",
		[]string{"philosophy", "meaning", "existence", "truth", "reality", "ethics"}, "philosophical")

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

	log.Println("🎉 Natural Conversation System Started!")
	log.Println("📝 Conversation ID:", conversationID)
	log.Println("🤖 Active Agents:")
	log.Println("   - Curious Agent (curious, learning-focused)")
	log.Println("   - Helpful Agent (helpful, problem-solving)")
	log.Println("   - Technical Agent (technical, programming-focused)")
	log.Println("   - Philosophical Agent (philosophical, deep thinking)")
	log.Println("🌐 Web Interface: http://localhost:8080")
	log.Println("💬 Start chatting and watch the natural conversation flow!")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
