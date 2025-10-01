package web

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"philoking/internal/config"
	"philoking/internal/kafka"
	"philoking/internal/types"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ClientInfo stores information about a WebSocket client
type ClientInfo struct {
	Conn   *websocket.Conn
	UserID string
	Name   string
}

// Server handles web requests and WebSocket connections
type Server struct {
	config      config.WebConfig
	kafkaClient *kafka.Client
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]*ClientInfo
	clientsMu   sync.RWMutex
}

// NewServer creates a new web server
func NewServer(cfg config.WebConfig, kafkaClient *kafka.Client) *Server {
	return &Server{
		config:      cfg,
		kafkaClient: kafkaClient,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
		clients: make(map[*websocket.Conn]*ClientInfo),
	}
}

// Start starts the web server
func (s *Server) Start() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Serve static files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// Routes
	r.GET("/", s.handleIndex)
	r.GET("/ws", s.handleWebSocket)
	r.POST("/api/message", s.handleSendMessage)
	r.GET("/api/agents", s.handleGetAgents)

	// Start Kafka message consumer for WebSocket broadcasting
	go s.startMessageConsumer()

	addr := s.config.Host + ":" + s.config.Port
	log.Printf("Web server starting on %s", addr)
	return r.Run(addr)
}

// handleIndex serves the main chat page
func (s *Server) handleIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "PhiloKing Chat",
	})
}

// handleWebSocket handles WebSocket connections
func (s *Server) handleWebSocket(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Create unique user agent for this connection
	userID := uuid.New().String()
	userName := "User-" + userID[:8] // Short ID for display

	// Register client with user info
	s.clientsMu.Lock()
	s.clients[conn] = &ClientInfo{
		Conn:   conn,
		UserID: userID,
		Name:   userName,
	}
	s.clientsMu.Unlock()

	log.Printf("WebSocket client connected as %s (ID: %s). Total clients: %d", userName, userID, len(s.clients))

	// Handle client messages
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Handle different message types
		switch msg["type"] {
		case "ping":
			conn.WriteJSON(map[string]string{"type": "pong"})
		case "message":
			// Forward to Kafka with user info
			if content, ok := msg["content"].(string); ok {
				s.sendUserMessage(content, userID, userName)
			}
		}
	}

	// Unregister client
	s.clientsMu.Lock()
	delete(s.clients, conn)
	s.clientsMu.Unlock()
	log.Printf("WebSocket client disconnected. Total clients: %d", len(s.clients))
}

// handleSendMessage handles HTTP POST requests to send messages
func (s *Server) handleSendMessage(c *gin.Context) {
	var req struct {
		Content string `json:"content"`
		UserID  string `json:"user_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate user ID and name if not provided
	userID := req.UserID
	if userID == "" {
		userID = uuid.New().String()
	}
	userName := "User-" + userID[:8]

	if err := s.sendUserMessage(req.Content, userID, userName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "message sent"})
}

// handleGetAgents returns information about available agents
func (s *Server) handleGetAgents(c *gin.Context) {
	// This would typically query the agent manager
	agents := []map[string]string{
		{"id": "echo-agent", "name": "Echo Agent", "status": "active"},
		{"id": "llm-agent", "name": "LLM Agent", "status": "active"},
	}
	c.JSON(http.StatusOK, gin.H{"agents": agents})
}

// sendUserMessage sends a user message to Kafka
func (s *Server) sendUserMessage(content, userID, userName string) error {
	message := &types.ChatMessage{
		ID:        generateID(),
		Type:      types.MessageTypeUser,
		Content:   content,
		AgentID:   userID, // Treat user as an agent
		UserID:    userID,
		Timestamp: time.Now(),
		Metadata: types.Metadata{
			ConversationID: "main-conversation",
			FromAgent:      userName, // Human-readable name
		},
	}

	log.Printf("User %s (%s) sending message: %s", userName, userID, content)
	return s.kafkaClient.PublishChatMessage(context.Background(), message)
}

// startMessageConsumer starts consuming messages from Kafka and broadcasting to WebSocket clients
func (s *Server) startMessageConsumer() {
	ctx := context.Background()

	// Subscribe to user messages
	go func() {
		err := s.kafkaClient.SubscribeToChatMessages(ctx, func(message *types.ChatMessage) error {
			s.broadcastMessage(message)
			return nil
		})
		if err != nil {
			log.Printf("Error in user message consumer: %v", err)
		}
	}()

	// Subscribe to agent responses
	go func() {
		err := s.kafkaClient.SubscribeToChatResponses(ctx, func(message *types.ChatMessage) error {
			s.broadcastMessage(message)
			return nil
		})
		if err != nil {
			log.Printf("Error in response message consumer: %v", err)
		}
	}()
}

// broadcastMessage broadcasts a message to all connected WebSocket clients
func (s *Server) broadcastMessage(message *types.ChatMessage) {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()

	// Get sender name for display
	senderName := message.AgentID
	if message.Metadata.FromAgent != "" {
		senderName = message.Metadata.FromAgent
	}

	log.Printf("Broadcasting message: %s (type: %s, from: %s)", message.Content, message.Type, senderName)

	// Convert message to JSON
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message for broadcast: %v", err)
		return
	}

	// Broadcast to all clients
	for conn, clientInfo := range s.clients {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("Error broadcasting to client %s: %v", clientInfo.Name, err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// generateID generates a simple ID (in production, use a proper UUID library)
func generateID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
