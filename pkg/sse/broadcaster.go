package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/danghamo/life/internal/api/jsonrpcx"
	"github.com/danghamo/life/internal/api/middleware"
	"github.com/danghamo/life/pkg/logger"
)

// SSEClient represents a connected SSE client
type SSEClient struct {
	ID       string
	UserID   string
	Writer   http.ResponseWriter
	Flusher  http.Flusher
	Done     chan bool
	LastSeen time.Time
	mutex    sync.Mutex // Protects concurrent writes to this client
}

// UserMessage represents a message targeted to a specific user
type UserMessage struct {
	UserID       string
	Notification jsonrpcx.JsonRpcNotification
}

// SSEBroadcaster manages SSE connections and broadcasts
type SSEBroadcaster struct {
	logger        *logger.Logger
	clients       map[string]*SSEClient
	userClients   map[string][]*SSEClient // Map userID to their clients
	mutex         sync.RWMutex
	broadcast     chan []byte
	userBroadcast chan UserMessage
	cleanup       *time.Ticker
	shutdown      chan struct{} // Global shutdown signal
}

// NewSSEBroadcaster creates a new SSE broadcaster
func NewSSEBroadcaster(logger *logger.Logger) *SSEBroadcaster {
	broadcaster := &SSEBroadcaster{
		logger:        logger.WithComponent("sse-broadcaster"),
		clients:       make(map[string]*SSEClient),
		userClients:   make(map[string][]*SSEClient),
		broadcast:     make(chan []byte, 1000),
		userBroadcast: make(chan UserMessage, 1000),
		cleanup:       time.NewTicker(30 * time.Second), // Cleanup every 30 seconds
		shutdown:      make(chan struct{}),
	}

	// Start background goroutines
	go broadcaster.broadcastLoop()
	go broadcaster.userBroadcastLoop()
	go broadcaster.cleanupLoop()

	return broadcaster
}

// AddClient adds a new SSE client
func (b *SSEBroadcaster) AddClient(client *SSEClient) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.clients[client.ID] = client
	
	// Add to user clients map
	if b.userClients[client.UserID] == nil {
		b.userClients[client.UserID] = make([]*SSEClient, 0)
	}
	b.userClients[client.UserID] = append(b.userClients[client.UserID], client)
	
	b.logger.Debug("SSE client connected",
		zap.String("clientId", client.ID),
		zap.String("userId", client.UserID))
}

// RemoveClient removes an SSE client
func (b *SSEBroadcaster) RemoveClient(clientID string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if client, exists := b.clients[clientID]; exists {
		// Safely close the Done channel
		select {
		case <-client.Done:
			// Channel already closed
		default:
			close(client.Done)
		}
		delete(b.clients, clientID)
		
		// Remove from user clients map
		if userClients := b.userClients[client.UserID]; userClients != nil {
			for i, uc := range userClients {
				if uc.ID == clientID {
					b.userClients[client.UserID] = append(userClients[:i], userClients[i+1:]...)
					break
				}
			}
			// Clean up empty user client slice
			if len(b.userClients[client.UserID]) == 0 {
				delete(b.userClients, client.UserID)
			}
		}
		
		b.logger.Debug("SSE client disconnected",
			zap.String("clientId", clientID),
			zap.String("userId", client.UserID))
	}
}

// broadcastToUser sends a JSON-RPC notification to a specific user (internal helper)
func (b *SSEBroadcaster) broadcastToUser(userID string, notification jsonrpcx.JsonRpcNotification) {
	msg := UserMessage{
		UserID:       userID,
		Notification: notification,
	}
	
	select {
	case b.userBroadcast <- msg:
	default:
		b.logger.Warn("User broadcast channel full, dropping message",
			zap.String("userId", userID))
	}
}

// BroadcastToAll sends a JSON-RPC notification to all connected clients
func (b *SSEBroadcaster) BroadcastToAll(notification jsonrpcx.JsonRpcNotification) {
	data, err := json.Marshal(notification)
	if err != nil {
		b.logger.Error("Failed to marshal JSON-RPC notification", zap.Error(err))
		return
	}

	select {
	case b.broadcast <- data:
	default:
		b.logger.Warn("Broadcast channel full, dropping message")
	}
}

// BroadcastToUsers sends a JSON-RPC notification to specific users (only if they are connected to this server)
func (b *SSEBroadcaster) BroadcastToUsers(targetUsers []string, notification jsonrpcx.JsonRpcNotification) {
	if len(targetUsers) == 0 {
		return
	}

	b.mutex.RLock()
	// Find which target users are connected to this server
	localTargetUsers := make([]string, 0)
	for _, userID := range targetUsers {
		if clients := b.userClients[userID]; clients != nil && len(clients) > 0 {
			localTargetUsers = append(localTargetUsers, userID)
		}
	}
	b.mutex.RUnlock()

	if len(localTargetUsers) == 0 {
		b.logger.Debug("No target users connected to this server", 
			zap.Strings("targetUsers", targetUsers))
		return
	}

	// Send to each connected target user
	for _, userID := range localTargetUsers {
		b.broadcastToUser(userID, notification)
	}

	b.logger.Debug("Broadcast sent to local target users",
		zap.Strings("targetUsers", targetUsers),
		zap.Strings("localTargetUsers", localTargetUsers))
}

// userBroadcastLoop handles broadcasting messages to specific users
func (b *SSEBroadcaster) userBroadcastLoop() {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Recovered from panic in userBroadcastLoop", zap.Any("panic", r))
			// Restart the loop
			go b.userBroadcastLoop()
		}
	}()
	
	for {
		select {
		case <-b.shutdown:
			b.logger.Info("User broadcast loop shutting down")
			return
		case msg := <-b.userBroadcast:
			b.mutex.RLock()
			userClients := make([]*SSEClient, 0)
			if clients := b.userClients[msg.UserID]; clients != nil {
				userClients = append(userClients, clients...)
			}
			b.mutex.RUnlock()

			if len(userClients) == 0 {
				b.logger.Debug("No clients found for user", zap.String("userId", msg.UserID))
				continue
			}

			data, err := json.Marshal(msg.Notification)
			if err != nil {
				b.logger.Error("Failed to marshal user notification", zap.Error(err))
				continue
			}

			// Create a list of clients to remove (to avoid modifying during iteration)
			var toRemove []string
			
			for _, client := range userClients {
				// Skip nil clients
				if client == nil {
					continue
				}
				
				select {
				case <-client.Done:
					toRemove = append(toRemove, client.ID)
				default:
					if err := b.sendToClient(client, data); err != nil {
						b.logger.Warn("Failed to send to user client",
							zap.String("clientId", client.ID),
							zap.String("userId", client.UserID),
							zap.Error(err))
						toRemove = append(toRemove, client.ID)
					}
				}
			}
			
			// Remove failed clients
			for _, clientID := range toRemove {
				b.RemoveClient(clientID)
			}
		}
	}
}

// broadcastLoop handles broadcasting messages to all connected clients
func (b *SSEBroadcaster) broadcastLoop() {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Recovered from panic in broadcastLoop", zap.Any("panic", r))
			// Restart the loop
			go b.broadcastLoop()
		}
	}()
	
	for {
		select {
		case <-b.shutdown:
			b.logger.Info("Broadcast loop shutting down")
			return
		case data := <-b.broadcast:
			b.mutex.RLock()
			clients := make([]*SSEClient, 0, len(b.clients))
			for _, client := range b.clients {
				clients = append(clients, client)
			}
			b.mutex.RUnlock()

			for _, client := range clients {
				select {
				case <-client.Done:
					b.RemoveClient(client.ID)
				default:
					if err := b.sendToClient(client, data); err != nil {
						b.logger.Warn("Failed to send to client",
							zap.String("clientId", client.ID),
							zap.Error(err))
						b.RemoveClient(client.ID)
					}
				}
			}
		}
	}
}

// sendToClient sends data to a specific SSE client
func (b *SSEBroadcaster) sendToClient(client *SSEClient, data []byte) (err error) {
	// Recover from any panic
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Recovered from panic in sendToClient", 
				zap.Any("panic", r),
				zap.String("clientId", func() string {
					if client != nil {
						return client.ID
					}
					return "unknown"
				}()))
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	// Check if client or its components are nil
	if client == nil {
		return fmt.Errorf("client is nil")
	}
	if client.Writer == nil {
		return fmt.Errorf("client writer is nil")
	}
	if client.Flusher == nil {
		return fmt.Errorf("client flusher is nil")
	}

	// Use client-specific mutex to prevent concurrent writes
	client.mutex.Lock()
	defer client.mutex.Unlock()
	
	// Check if client is still connected before writing
	select {
	case <-client.Done:
		return fmt.Errorf("client connection closed")
	default:
	}
	
	// Write SSE data with proper formatting
	// Use a single write operation to reduce chunking issues
	sseData := fmt.Sprintf("data: %s\n\n", data)
	n, err := client.Writer.Write([]byte(sseData))
	if err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	if n != len(sseData) {
		return fmt.Errorf("incomplete write: wrote %d/%d bytes", n, len(sseData))
	}

	// Force flush immediately
	client.Flusher.Flush()
	client.LastSeen = time.Now()
	return nil
}

// cleanupLoop removes stale connections
func (b *SSEBroadcaster) cleanupLoop() {
	defer func() {
		if r := recover(); r != nil {
			b.logger.Error("Recovered from panic in cleanupLoop", zap.Any("panic", r))
			// Restart the loop
			go b.cleanupLoop()
		}
	}()
	
	for {
		select {
		case <-b.shutdown:
			b.logger.Info("Cleanup loop shutting down")
			return
		case <-b.cleanup.C:
			b.mutex.Lock()
			now := time.Now()
			for clientID, client := range b.clients {
				if now.Sub(client.LastSeen) > 60*time.Second {
					b.logger.Debug("Removing stale SSE client",
						zap.String("clientId", clientID))
					close(client.Done)
					delete(b.clients, clientID)
				}
			}
			b.mutex.Unlock()
		}
	}
}

// GetClientCount returns the number of connected clients
func (b *SSEBroadcaster) GetClientCount() int {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return len(b.clients)
}

// Close shuts down the broadcaster
func (b *SSEBroadcaster) Close() {
	b.logger.Debug("Shutting down SSE broadcaster")
	
	// Signal all goroutines to stop
	close(b.shutdown)
	
	b.cleanup.Stop()
	close(b.broadcast)
	close(b.userBroadcast)

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Close all client connections immediately
	for clientID, client := range b.clients {
		b.logger.Debug("Force closing SSE client", zap.String("clientId", clientID))
		close(client.Done)
	}
	b.clients = make(map[string]*SSEClient)
	b.userClients = make(map[string][]*SSEClient)
	
	b.logger.Debug("SSE broadcaster shutdown complete")
}

// HandleSSE handles SSE connections for real-time position updates
func (b *SSEBroadcaster) HandleSSE(w http.ResponseWriter, r *http.Request) {
	b.logger.Debug("SSE connection attempt", zap.String("method", r.Method), zap.String("url", r.URL.String()))
	
	// Get user info from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		b.logger.Error("SSE: Authentication failed - no user ID in context")
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	
	b.logger.Debug("SSE: User authenticated", zap.String("userID", userID))

	// Check if client supports SSE
	flusher, ok := w.(http.Flusher)
	if !ok {
		b.logger.Error("SSE: Client does not support flusher interface")
		http.Error(w, "Server-Sent Events not supported", http.StatusInternalServerError)
		return
	}
	
	b.logger.Debug("SSE: Client supports flusher interface")

	// Set SSE headers with improved chunked encoding handling
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering
	w.Header().Set("Transfer-Encoding", "chunked") // Explicit chunked encoding
	
	b.logger.Debug("SSE: Headers set successfully")

	// Create client
	clientID := fmt.Sprintf("%s-%d", userID, time.Now().UnixNano())
	client := &SSEClient{
		ID:       clientID,
		UserID:   userID,
		Writer:   w,
		Flusher:  flusher,
		Done:     make(chan bool),
		LastSeen: time.Now(),
	}
	
	b.logger.Debug("SSE: Client created", zap.String("clientID", clientID))

	// Add client to broadcaster
	b.AddClient(client)
	defer b.RemoveClient(clientID)
	
	b.logger.Debug("SSE: Client added to broadcaster")

	// Send initial connection message
	initialMsg := fmt.Sprintf("data: {\"type\":\"connected\",\"client_id\":\"%s\"}\n\n", clientID)
	w.Write([]byte(initialMsg))
	flusher.Flush()
	
	b.logger.Debug("SSE: Initial message sent and flushed")
	
	// TODO: Send initial sync data (current online trainers)
	// This will be implemented after adding MovementBroadcaster reference

	// Keep connection alive
	heartbeat := time.NewTicker(30 * time.Second)
	defer heartbeat.Stop()
	
	b.logger.Debug("SSE: Starting connection loop")

	for {
		select {
		case <-client.Done:
			b.logger.Debug("SSE client done signal received", zap.String("clientId", clientID))
			return
		case <-r.Context().Done():
			b.logger.Info("SSE request context cancelled", zap.String("clientId", clientID))
			return
		case <-b.shutdown:
			b.logger.Debug("SSE broadcaster shutdown signal received", zap.String("clientId", clientID))
			return
		case <-heartbeat.C:
			if err := b.sendHeartbeat(w, flusher); err != nil {
				b.logger.Warn("Failed to send heartbeat", 
					zap.String("clientId", clientID),
					zap.Error(err))
				return
			}
		}
	}
}

// sendHeartbeat sends a heartbeat message to the SSE client
func (b *SSEBroadcaster) sendHeartbeat(w http.ResponseWriter, flusher http.Flusher) error {
	heartbeatData := fmt.Sprintf("data: {\"type\":\"heartbeat\",\"timestamp\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
	n, err := w.Write([]byte(heartbeatData))
	if err != nil {
		return fmt.Errorf("heartbeat write failed: %w", err)
	}
	if n != len(heartbeatData) {
		return fmt.Errorf("incomplete heartbeat write: wrote %d/%d bytes", n, len(heartbeatData))
	}
	flusher.Flush()
	return nil
}