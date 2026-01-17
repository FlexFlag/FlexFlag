package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ClientSSEHandler manages Server-Sent Events connections for client SDKs
type ClientSSEHandler struct {
	clients      map[string]*ClientSSEConnection
	mu           sync.RWMutex
	pingTicker   *time.Ticker
	done         chan bool
	apiKeyRepo   *postgres.ApiKeyRepository
}

// ClientSSEConnection represents a connected client SDK via SSE
type ClientSSEConnection struct {
	ID            string    `json:"id"`
	ProjectID     string    `json:"project_id"`
	Environment   string    `json:"environment"`
	EnvironmentID string    `json:"environment_id"`
	Writer        http.ResponseWriter
	Flusher       http.Flusher
	Done          chan bool
	LastPing      time.Time `json:"last_ping"`
	ConnectedAt   time.Time `json:"connected_at"`
}

// ClientSSEEvent represents an event sent to client SDKs via SSE
type ClientSSEEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// ClientFlagUpdateEvent represents a flag update event for clients
type ClientFlagUpdateEvent struct {
	FlagKey     string    `json:"flag_key"`
	Action      string    `json:"action"` // created, updated, deleted, toggled
	Environment string    `json:"environment"`
	Timestamp   time.Time `json:"timestamp"`
}

// NewClientSSEHandler creates a new SSE handler for client SDKs
func NewClientSSEHandler(apiKeyRepo *postgres.ApiKeyRepository) *ClientSSEHandler {
	h := &ClientSSEHandler{
		clients:    make(map[string]*ClientSSEConnection),
		pingTicker: time.NewTicker(30 * time.Second),
		done:       make(chan bool),
		apiKeyRepo: apiKeyRepo,
	}

	// Start ping routine
	go h.pingClients()

	return h
}

// HandleClientSSE handles SSE connections from client SDKs
func (h *ClientSSEHandler) HandleClientSSE(c *gin.Context) {
	// Get API key from query parameter (EventSource doesn't support custom headers)
	apiKey := c.Query("api_key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "api_key query parameter is required",
		})
		return
	}

	// Authenticate API key
	keyInfo, err := h.apiKeyRepo.AuthenticateApiKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired API key",
		})
		return
	}

	// Get environment from query parameter or use default from API key
	environment := c.Query("environment")
	if environment == "" {
		environment = keyInfo.Environment.Key
	}

	// Setup SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	// Get writer and flusher
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "streaming not supported",
		})
		return
	}

	// Create client connection
	client := &ClientSSEConnection{
		ID:            uuid.New().String(),
		ProjectID:     keyInfo.ProjectID,
		Environment:   environment,
		EnvironmentID: keyInfo.EnvironmentID,
		Writer:        writer,
		Flusher:       flusher,
		Done:          make(chan bool),
		LastPing:      time.Now(),
		ConnectedAt:   time.Now(),
	}

	// Register client
	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()

	// Send initial connection event
	h.sendEvent(client, ClientSSEEvent{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id":      client.ID,
			"project_id":     client.ProjectID,
			"environment":    client.Environment,
			"environment_id": client.EnvironmentID,
			"message":        "Connected to FlexFlag SSE stream",
		},
		Timestamp: time.Now(),
	})

	// Handle client disconnection
	defer func() {
		h.mu.Lock()
		delete(h.clients, client.ID)
		h.mu.Unlock()
		close(client.Done)
	}()

	// Keep connection alive until client disconnects
	select {
	case <-c.Request.Context().Done():
		return
	case <-client.Done:
		return
	}
}

// BroadcastFlagUpdate sends flag update to all connected clients for specific project/environment
func (h *ClientSSEHandler) BroadcastFlagUpdate(flag *types.Flag, action string) {
	event := ClientSSEEvent{
		Type: "flag_update",
		Data: ClientFlagUpdateEvent{
			FlagKey:     flag.Key,
			Action:      action,
			Environment: flag.Environment,
			Timestamp:   time.Now(),
		},
		Timestamp: time.Now(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	fmt.Printf("[ClientSSE] Broadcasting flag update: flag=%s, action=%s, env=%s, projectID=%s, total_clients=%d\n",
		flag.Key, action, flag.Environment, flag.ProjectID, len(h.clients))

	// Send to all clients in the same project and environment
	sentCount := 0
	for _, client := range h.clients {
		fmt.Printf("[ClientSSE] Checking client: clientID=%s, projectID=%s, env=%s, match=%v\n",
			client.ID, client.ProjectID, client.Environment,
			client.ProjectID == flag.ProjectID && client.Environment == flag.Environment)
		if client.ProjectID == flag.ProjectID && client.Environment == flag.Environment {
			h.sendEvent(client, event)
			sentCount++
		}
	}
	fmt.Printf("[ClientSSE] Sent flag update to %d clients\n", sentCount)
}

// BroadcastToEnvironment sends event to all clients in specific environment
func (h *ClientSSEHandler) BroadcastToEnvironment(projectID, environment string, event ClientSSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		if client.ProjectID == projectID && client.Environment == environment {
			h.sendEvent(client, event)
		}
	}
}

// GetConnectedClients returns list of connected client SDKs
func (h *ClientSSEHandler) GetConnectedClients() []*ClientSSEConnection {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := make([]*ClientSSEConnection, 0, len(h.clients))
	for _, client := range h.clients {
		// Create copy without Writer/Flusher for JSON serialization
		clientCopy := &ClientSSEConnection{
			ID:            client.ID,
			ProjectID:     client.ProjectID,
			Environment:   client.Environment,
			EnvironmentID: client.EnvironmentID,
			LastPing:      client.LastPing,
			ConnectedAt:   client.ConnectedAt,
		}
		clients = append(clients, clientCopy)
	}

	return clients
}

// sendEvent sends an SSE event to a specific client
func (h *ClientSSEHandler) sendEvent(client *ClientSSEConnection, event ClientSSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	// SSE format with event type: "event: flag_update\ndata: {json}\n\n"
	fmt.Fprintf(client.Writer, "event: %s\ndata: %s\n\n", event.Type, data)
	client.Flusher.Flush()
}

// pingClients sends periodic ping events to keep connections alive
func (h *ClientSSEHandler) pingClients() {
	for {
		select {
		case <-h.pingTicker.C:
			pingEvent := ClientSSEEvent{
				Type:      "ping",
				Data:      map[string]interface{}{"timestamp": time.Now()},
				Timestamp: time.Now(),
			}

			h.mu.Lock()
			for _, client := range h.clients {
				client.LastPing = time.Now()
				h.sendEvent(client, pingEvent)
			}
			h.mu.Unlock()

		case <-h.done:
			return
		}
	}
}

// Close shuts down the SSE handler
func (h *ClientSSEHandler) Close() {
	h.pingTicker.Stop()
	close(h.done)

	h.mu.Lock()
	for _, client := range h.clients {
		close(client.Done)
	}
	h.clients = make(map[string]*ClientSSEConnection)
	h.mu.Unlock()
}

// HandleClientStatus returns status of connected client SDKs
func (h *ClientSSEHandler) HandleClientStatus(c *gin.Context) {
	clients := h.GetConnectedClients()

	// Transform to response format
	clientsList := make([]map[string]interface{}, len(clients))
	for i, client := range clients {
		uptime := time.Since(client.ConnectedAt)
		clientsList[i] = map[string]interface{}{
			"id":               client.ID,
			"project_id":       client.ProjectID,
			"environment":      client.Environment,
			"environment_id":   client.EnvironmentID,
			"status":           "connected",
			"connected_at":     client.ConnectedAt.Format(time.RFC3339),
			"last_ping":        client.LastPing.Format(time.RFC3339),
			"uptime_seconds":   int(uptime.Seconds()),
			"uptime_human":     uptime.String(),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"clients": clientsList,
		"total":   len(clientsList),
	})
}
