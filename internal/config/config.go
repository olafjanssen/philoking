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
		ChatMessages  string `mapstructure:"chat_messages"`
		ChatResponses string `mapstructure:"chat_responses"`
		AgentInput    string `mapstructure:"agent_input"`
		AgentOutput   string `mapstructure:"agent_output"`
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
	// Natural conversation agents configuration
	NaturalAgents []NaturalAgentConfig `mapstructure:"natural_agents"`
}

// NaturalAgentConfig defines the configuration for a natural conversation agent
type NaturalAgentConfig struct {
	ID             string   `mapstructure:"id"`
	Name           string   `mapstructure:"name"`
	Personality    string   `mapstructure:"personality"`
	Interests      []string `mapstructure:"interests"`
	ResponseChance float64  `mapstructure:"response_chance"`
	IsEnabled      bool     `mapstructure:"enabled"`
	Description    string   `mapstructure:"description,omitempty"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	viper.SetDefault("kafka.brokers", []string{"localhost:9092"})
	viper.SetDefault("kafka.topics.chat_messages", "chat-messages")
	viper.SetDefault("kafka.topics.chat_responses", "chat-responses")
	viper.SetDefault("kafka.topics.agent_input", "agent-input")
	viper.SetDefault("kafka.topics.agent_output", "agent-output")
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

// GetEnabledNaturalAgents returns only the enabled natural agents
func (c *Config) GetEnabledNaturalAgents() []NaturalAgentConfig {
	var enabled []NaturalAgentConfig
	for _, agent := range c.Agents.NaturalAgents {
		if agent.IsEnabled {
			enabled = append(enabled, agent)
		}
	}
	return enabled
}
