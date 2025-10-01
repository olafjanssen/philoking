# PhiloKing - Multi-Agent Chat System

A distributed chat system built in Go with multiple loosely coupled agents that participate asynchronously in conversations using an LLM API and Kafka message bus.

## Features

- **Multi-Agent Architecture**: Multiple AI agents can participate in conversations asynchronously
- **Kafka Message Bus**: Robust message passing between agents and the web interface
- **Real-time Web Interface**: WebSocket-based chat interface for real-time communication
- **LLM Integration**: Support for OpenAI-compatible LLM APIs
- **Docker Support**: Easy deployment with Docker Compose
- **Extensible**: Easy to add new agents with different capabilities

## Architecture

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
                                    │   Echo Agent      │ │ LLM Agent │
                                    │   (Simple Echo)   │ │ (OpenAI)  │
                                    └───────────────────┘ └───────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (for local development)
- Optional: OpenAI API key for LLM functionality
- Optional: Ollama for local LLM models

### Using Docker Compose with Ollama (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd philoking
```

2. Set up Ollama (includes automatic model download):
```bash
# Linux/Mac
chmod +x setup-ollama.sh
./setup-ollama.sh

# Windows
setup-ollama.bat
```

3. Open your browser and go to `http://localhost:8080`

### Using Docker Compose with OpenAI

1. Clone the repository:
```bash
git clone <repository-url>
cd philoking
```

2. Set up environment variables:
```bash
cp env.example .env
# Edit .env and add your OpenAI API key
```

3. Update config.yaml to use OpenAI:
```yaml
agents:
  provider: "openai"
  llm_api_key: "your-api-key"
```

4. Start the system:
```bash
docker-compose up -d
```

5. Open your browser and go to `http://localhost:8080`

### Local Development

1. Install dependencies:
```bash
go mod download
```

2. Start Kafka and Ollama (using Docker):
```bash
docker-compose up -d zookeeper kafka ollama
```

3. Pull a model (if using Ollama):
```bash
docker exec ollama ollama pull llama2
```

4. Set environment variables:
```bash
# For Ollama (default)
export PROVIDER="ollama"
export MODEL="llama2"
export OLLAMA_URL="http://localhost:11434"

# For OpenAI (alternative)
export PROVIDER="openai"
export LLM_API_KEY="your-openai-api-key"
```

5. Run the application:
```bash
go run main.go
```

6. Open your browser and go to `http://localhost:8080`

## Configuration

The system can be configured via:

1. **config.yaml**: Main configuration file
2. **Environment variables**: Override config values
3. **Command line flags**: Future enhancement

### Key Configuration Options

- `kafka.brokers`: Kafka broker addresses
- `kafka.topics`: Topic names for different message types
- `web.host`/`web.port`: Web server configuration
- `agents.provider`: LLM provider ("ollama" or "openai")
- `agents.model`: Model name (e.g., "llama2", "codellama", "mistral")
- `agents.ollama_url`: Ollama server URL
- `agents.llm_api_key`: OpenAI API key for LLM functionality

## Agents

### Echo Agent
A simple agent that echoes user messages with some basic pattern recognition.

**Capabilities:**
- Echo responses
- Simple response patterns
- Greeting detection

### LLM Agent
An intelligent agent that uses either OpenAI's API or local Ollama models to generate contextual responses.

**Capabilities:**
- LLM-powered responses
- Conversational AI
- Context awareness
- Support for multiple model providers

**Supported Models:**
- **Ollama Models**: llama2, codellama, mistral, phi, neural-chat
- **OpenAI Models**: gpt-3.5-turbo, gpt-4, gpt-4-turbo

## Adding New Agents

To add a new agent:

1. Create a new agent file in `internal/agent/`
2. Implement the `Agent` interface
3. Register the agent in `main.go`

Example:
```go
type MyAgent struct {
    *BaseAgent
}

func NewMyAgent(kafkaClient *kafka.Client) *MyAgent {
    base := NewBaseAgent("my-agent", "My Agent", kafkaClient)
    agent := &MyAgent{BaseAgent: base}
    
    agent.SetHandler(types.MessageTypeUser, agent)
    agent.AddCapability("my_capability")
    
    return agent
}

func (a *MyAgent) HandleUserMessage(ctx context.Context, message *types.ChatMessage) error {
    // Your agent logic here
    return a.SendChatMessage(ctx, "Response", message.Metadata.ConversationID)
}
```

## API Endpoints

- `GET /`: Main chat interface
- `GET /ws`: WebSocket connection
- `POST /api/message`: Send message via HTTP
- `GET /api/agents`: Get agent information

## Message Types

### ChatMessage
- `user`: Messages from users
- `agent`: Messages from agents
- `system`: System messages
- `context`: Contextual information

### AgentMessage
- Inter-agent communication
- Broadcast or targeted messages
- Custom payload types

## Development

### Project Structure
```
philoking/
├── internal/
│   ├── agent/          # Agent implementations
│   ├── config/         # Configuration management
│   ├── kafka/          # Kafka client
│   ├── types/          # Data structures
│   └── web/            # Web server
├── web/
│   ├── static/         # CSS, JS assets
│   └── templates/      # HTML templates
├── config.yaml         # Configuration
├── docker-compose.yml  # Docker setup
└── main.go            # Application entry point
```

### Building
```bash
go build -o philoking main.go
```

### Testing
```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
