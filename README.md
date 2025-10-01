# PhiloKing - Configurable Multi-Agent Chat System

A **YAML-configurable** multi-agent chat system built in Go with loosely coupled agents that participate asynchronously in natural conversations using LLM APIs and Kafka message bus.

## 🌟 Features

- **🤖 YAML-Configurable Agents** - Easy agent management through configuration files
- **💬 Natural Conversation Flow** - Agents respond selectively based on relevance and personality
- **🔄 Asynchronous Communication** - Kafka message bus for robust agent communication
- **🌐 Real-time Web Interface** - WebSocket-based chat interface
- **🧠 LLM Integration** - Support for OpenAI and local Ollama models
- **⚙️ Flexible Configuration** - Customize agent personalities, interests, and behavior
- **🐳 Docker Support** - Easy deployment with Docker Compose

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+ (for local development)
- Optional: OpenAI API key for LLM functionality
- Optional: Ollama for local LLM models

### 1. Start the System
```bash
# Start Kafka
docker-compose up -d zookeeper kafka

# Run the application
test.bat
```

### 2. Access the Interface
Open `http://localhost:8080` in your browser and start chatting!

## 🎯 Agent Configuration

The system comes with 4 pre-configured agents that you can customize in `config.yaml`:

### **Curious Agent** 🔍
- **Personality**: Inquisitive, learning-focused
- **Interests**: Questions, learning, discovery, science, philosophy
- **Response Rate**: 80%

### **Helpful Agent** 🛠️
- **Personality**: Supportive, problem-solving oriented
- **Interests**: Help, support, guidance, problem-solving, assistance
- **Response Rate**: 90%

### **Technical Agent** 💻
- **Personality**: Engineering-minded, detail-oriented
- **Interests**: Programming, technology, software, engineering, code
- **Response Rate**: 70%

### **Philosophical Agent** 🤔
- **Personality**: Deep-thinking, contemplative
- **Interests**: Philosophy, meaning, existence, truth, reality, ethics
- **Response Rate**: 60%

## ⚙️ Configuration

### Agent Configuration Example
```yaml
natural_agents:
  - id: "my-agent"
    name: "My Custom Agent"
    personality: "helpful"
    interests:
      - "help"
      - "support"
      - "guidance"
    response_chance: 0.8
    enabled: true
    description: "My custom helpful agent"
```

### Configuration Parameters
| Parameter | Type | Description | Default |
|-----------|------|-------------|---------|
| `id` | string | Unique agent identifier | Required |
| `name` | string | Human-readable agent name | Uses ID if empty |
| `personality` | string | Agent personality type | "default" |
| `interests` | array | List of topics agent cares about | [] |
| `response_chance` | float | Probability of responding (0.0-1.0) | 0.7 |
| `enabled` | boolean | Whether agent is active | true |
| `description` | string | Agent description | "" |

## 🎨 Customization Examples

### Add a New Agent
```yaml
natural_agents:
  - id: "creative-agent"
    name: "Creative Agent"
    personality: "creative"
    interests:
      - "art"
      - "creativity"
      - "design"
      - "music"
    response_chance: 0.7
    enabled: true
    description: "An artistic agent focused on creative expression"
```

### Adjust Response Rates
```yaml
# High participation (agents respond often)
response_chance: 0.9

# Medium participation (balanced)
response_chance: 0.7

# Low participation (agents respond rarely)
response_chance: 0.3
```

### Enable/Disable Agents
```yaml
# Disable an agent
- id: "philosophical-agent"
  enabled: false

# Enable an agent
- id: "technical-agent"
  enabled: true
```

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Client    │    │   Web Server    │    │  Kafka Broker  │
│   (WebSocket)   │◄──►│   (Gin + WS)    │◄──►│   (Message     │
└─────────────────┘    └─────────────────┘    │    Bus)         │
                                              └─────────────────┘
                                                       ▲
                                                       │
                                              ┌────────┴────────┐
                                              │                 │
                                    ┌─────────▼─────────┐ ┌─────▼─────┐
                                    │  Curious Agent    │ │ Helpful   │
                                    │  (Learning)       │ │ Agent     │
                                    └───────────────────┘ └───────────┘
                                              │
                                    ┌─────────▼─────────┐
                                    │  Technical Agent  │
                                    │  (Engineering)    │
                                    └───────────────────┘
```

## 🔧 Development

### Project Structure
```
philoking/
├── internal/
│   ├── agent/          # Agent implementations
│   ├── config/         # Configuration management
│   ├── conversation/   # Conversation flow management
│   ├── kafka/          # Kafka client
│   ├── types/          # Data structures
│   └── web/            # Web server
├── web/
│   ├── static/         # CSS, JS assets
│   └── templates/      # HTML templates
├── config.yaml         # Main configuration
├── docker-compose.yml  # Docker setup
├── test.bat           # Test script
└── main.go            # Application entry point
```

### Building
```bash
go build -o philoking main.go
```

### Testing
```bash
test.bat
```

## 🎯 Key Benefits

1. **🎯 Easy Customization** - No code changes needed for agent configuration
2. **🔄 Quick Iteration** - Test different agent combinations easily
3. **👥 Team Collaboration** - Non-developers can configure agents
4. **📊 A/B Testing** - Compare different agent setups
5. **🔧 Environment-specific** - Different configs for different environments
6. **📝 Self-Documenting** - Configuration serves as documentation

## 🚀 Advanced Usage

### Using OpenAI Instead of Ollama
```yaml
agents:
  provider: "openai"
  model: "gpt-3.5-turbo"
  llm_api_key: "your-api-key"
```

### Using Different Ollama Models
```yaml
agents:
  provider: "ollama"
  model: "codellama"  # or "mistral", "phi", etc.
  ollama_url: "http://localhost:11434"
```

### Custom Agent Personalities
You can extend the system by adding new personality types in the agent code and using them in your configuration.

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Configure your perfect conversation ecosystem and start chatting!** 🎉