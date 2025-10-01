package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"philoking/internal/agent"
	"philoking/internal/config"
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

	// Initialize agent manager
	agentManager := agent.NewManager(kafkaClient, cfg.Agents)

	// Register agents
	echoAgent := agent.NewEchoAgent(kafkaClient)
	llmAgent := agent.NewLLMAgent(kafkaClient, cfg.Agents)

	if err := agentManager.RegisterAgent(echoAgent); err != nil {
		log.Fatalf("Failed to register echo agent: %v", err)
	}

	if err := agentManager.RegisterAgent(llmAgent); err != nil {
		log.Fatalf("Failed to register LLM agent: %v", err)
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

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	cancel()
}
