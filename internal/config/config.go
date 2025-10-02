package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Kafka  KafkaConfig  `mapstructure:"kafka"`
	Web    WebConfig    `mapstructure:"web"`
	Agents AgentsConfig `mapstructure:"agents"`
}

type KafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topics  struct {
		ChatMessages string `mapstructure:"chat_messages"`
	} `mapstructure:"topics"`
}

type WebConfig struct {
	Port string `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type AgentsConfig struct {
	LLMAPIKey string `mapstructure:"llm_api_key"`
	LLMURL    string `mapstructure:"llm_url"`
	OllamaURL string `mapstructure:"ollama_url"`
	Model     string `mapstructure:"model"`
	Provider  string `mapstructure:"provider"` // "openai" or "ollama"
	// Agents configuration
	Agents []AgentConfig `mapstructure:"agents"`
}

// AgentConfig defines the configuration for any agent
type AgentConfig struct {
	ID             string  `mapstructure:"id"`
	Name           string  `mapstructure:"name"`
	Type           string  `mapstructure:"type"` // "llm", "echo", "custom", etc.
	ResponseChance float64 `mapstructure:"response_chance"`
	IsEnabled      bool    `mapstructure:"enabled"`
	Description    string  `mapstructure:"description,omitempty"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topics.chat_messages", "chat-messages")
	viper.SetDefault("web.port", "8080")
	viper.SetDefault("web.host", "localhost")
	viper.SetDefault("agents.llm_url", "https://api.openai.com/v1/chat/completions")
	viper.SetDefault("agents.ollama_url", "http://localhost:11434")
	viper.SetDefault("agents.model", "llama2")
	viper.SetDefault("agents.provider", "ollama")

	// Allow environment variables to override config
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found, use defaults and env vars
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Override with environment variables if set
	if apiKey := os.Getenv("LLM_API_KEY"); apiKey != "" {
		config.Agents.LLMAPIKey = apiKey
	}

	return &config, nil
}

// GetEnabledAgents returns only the enabled agents
func (c *Config) GetEnabledAgents() []AgentConfig {
	var enabled []AgentConfig
	for _, agent := range c.Agents.Agents {
		if agent.IsEnabled {
			enabled = append(enabled, agent)
		}
	}
	return enabled
}
