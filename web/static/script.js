class ChatApp {
    constructor() {
        this.ws = null;
        this.isConnected = false;
        this.messageInput = document.getElementById('message-input');
        this.sendButton = document.getElementById('send-button');
        this.messagesContainer = document.getElementById('messages');
        this.connectionStatus = document.getElementById('connection-status');
        
        this.init();
    }

    init() {
        this.connectWebSocket();
        this.setupEventListeners();
    }

    connectWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws`;
        
        this.ws = new WebSocket(wsUrl);
        
        this.ws.onopen = () => {
            this.isConnected = true;
            this.updateConnectionStatus('connected', 'Connected');
            console.log('WebSocket connected');
        };
        
        this.ws.onclose = () => {
            this.isConnected = false;
            this.updateConnectionStatus('disconnected', 'Disconnected');
            console.log('WebSocket disconnected');
            
            // Attempt to reconnect after 3 seconds
            setTimeout(() => {
                if (!this.isConnected) {
                    this.connectWebSocket();
                }
            }, 3000);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus('disconnected', 'Connection Error');
        };
        
        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error('Error parsing message:', error);
            }
        };
    }

    setupEventListeners() {
        this.sendButton.addEventListener('click', () => {
            this.sendMessage();
        });
        
        this.messageInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
        
        // Auto-focus input
        this.messageInput.focus();
    }

    sendMessage() {
        const content = this.messageInput.value.trim();
        if (!content || !this.isConnected) {
            return;
        }
        
        // Add user message to UI immediately
        this.addMessage({
            type: 'user',
            content: content,
            timestamp: new Date().toISOString()
        });
        
        // Send to server
        this.ws.send(JSON.stringify({
            type: 'message',
            content: content
        }));
        
        // Clear input
        this.messageInput.value = '';
    }

    handleMessage(message) {
        console.log('Received WebSocket message:', message);
        
        if (message.type === 'pong') {
            return; // Handle ping/pong
        }
        
        this.addMessage(message);
    }

    addMessage(message) {
        console.log('Adding message to UI:', message);
        
        const messageElement = document.createElement('div');
        messageElement.className = `message ${message.type}-message`;
        
        const contentElement = document.createElement('div');
        contentElement.className = 'message-content';
        contentElement.textContent = message.content;
        
        messageElement.appendChild(contentElement);
        
        // Add metadata if available
        if (message.agent_id || message.user_id) {
            const metaElement = document.createElement('div');
            metaElement.className = 'message-meta';
            metaElement.textContent = message.agent_id || message.user_id;
            messageElement.appendChild(metaElement);
        }
        
        this.messagesContainer.appendChild(messageElement);
        this.scrollToBottom();
        
        console.log('Message added to UI successfully');
    }

    scrollToBottom() {
        this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
    }

    updateConnectionStatus(status, text) {
        this.connectionStatus.className = `status-indicator ${status}`;
        this.connectionStatus.textContent = text;
    }
}

// Initialize the chat app when the page loads
document.addEventListener('DOMContentLoaded', () => {
    new ChatApp();
});

