package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SSEHandler manages Server-Sent Events connections for edge servers
type SSEHandler struct {
	clients    map[string]*SSEClient
	mu         sync.RWMutex
	pingTicker *time.Ticker
	done       chan bool
}

// SSEClient represents an connected edge server via SSE
type SSEClient struct {
	ID          string    `json:"id"`
	ProjectID   string    `json:"project_id"`
	Environment string    `json:"environment"`
	ServerID    string    `json:"server_id"`
	Writer      http.ResponseWriter
	Flusher     http.Flusher
	Done        chan bool
	LastPing    time.Time `json:"last_ping"`
	ConnectedAt time.Time `json:"connected_at"`
}

// SSEEvent represents an event sent via SSE
type SSEEvent struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// FlagUpdateEvent represents a flag update event
type FlagUpdateEvent struct {
	Type        string      `json:"type"`
	Action      string      `json:"action"` // create, update, delete
	Flag        *types.Flag `json:"flag"`
	ProjectID   string      `json:"project_id"`
	Environment string      `json:"environment"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewSSEHandler creates a new SSE handler
func NewSSEHandler() *SSEHandler {
	h := &SSEHandler{
		clients:    make(map[string]*SSEClient),
		pingTicker: time.NewTicker(30 * time.Second),
		done:       make(chan bool),
	}
	
	// Start ping routine
	go h.pingClients()
	
	return h
}

// HandleSSE handles SSE connections from edge servers
func (h *SSEHandler) HandleSSE(c *gin.Context) {
	// Validate required parameters - edge servers are global, so only server_id is required
	serverID := c.Query("server_id")
	
	if serverID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "server_id is required",
		})
		return
	}
	
	// Optional parameters for backwards compatibility
	projectID := c.Query("project_id")
	environment := c.Query("environment")
	
	// Set defaults for global edge servers
	if projectID == "" {
		projectID = "global-edge-server"
	}
	if environment == "" {
		environment = "all"
	}
	
	// Setup SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")
	
	// Get writer and flusher
	writer := c.Writer
	flusher, ok := writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "streaming not supported",
		})
		return
	}
	
	// Create client
	client := &SSEClient{
		ID:          uuid.New().String(),
		ProjectID:   projectID,
		Environment: environment,
		ServerID:    serverID,
		Writer:      writer,
		Flusher:     flusher,
		Done:        make(chan bool),
		LastPing:    time.Now(),
		ConnectedAt: time.Now(),
	}
	
	// Register client
	h.mu.Lock()
	h.clients[client.ID] = client
	h.mu.Unlock()
	
	// Send initial connection event
	h.sendEvent(client, SSEEvent{
		Type: "connected",
		Data: map[string]interface{}{
			"client_id":   client.ID,
			"server_id":   client.ServerID,
			"project_id":  client.ProjectID,
			"environment": client.Environment,
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

// BroadcastFlagUpdate sends flag update to ALL edge servers (they are project-independent)
func (h *SSEHandler) BroadcastFlagUpdate(flag *types.Flag, action string) {
	event := SSEEvent{
		Type: "flag_update",
		Data: FlagUpdateEvent{
			Type:        "flag_update",
			Action:      action,
			Flag:        flag,
			ProjectID:   flag.ProjectID,
			Environment: flag.Environment,
			Timestamp:   time.Now(),
		},
		Timestamp: time.Now(),
	}
	
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Send to ALL edge servers - they are global and should receive all flag updates
	for _, client := range h.clients {
		h.sendEvent(client, event)
	}
}

// BroadcastToProject sends event to all edge servers in a project
func (h *SSEHandler) BroadcastToProject(projectID string, event SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, client := range h.clients {
		if client.ProjectID == projectID {
			h.sendEvent(client, event)
		}
	}
}

// BroadcastToEnvironment sends event to all edge servers in an environment
func (h *SSEHandler) BroadcastToEnvironment(projectID, environment string, event SSEEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, client := range h.clients {
		if client.ProjectID == projectID && client.Environment == environment {
			h.sendEvent(client, event)
		}
	}
}

// GetConnectedClients returns list of connected edge servers
func (h *SSEHandler) GetConnectedClients() []*SSEClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	clients := make([]*SSEClient, 0, len(h.clients))
	for _, client := range h.clients {
		// Create copy without Writer/Flusher for JSON serialization
		clientCopy := &SSEClient{
			ID:          client.ID,
			ProjectID:   client.ProjectID,
			Environment: client.Environment,
			ServerID:    client.ServerID,
			LastPing:    client.LastPing,
			ConnectedAt: client.ConnectedAt,
		}
		clients = append(clients, clientCopy)
	}
	
	return clients
}

// GetClientsByProject returns connected clients for a specific project
func (h *SSEHandler) GetClientsByProject(projectID string) []*SSEClient {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	var clients []*SSEClient
	for _, client := range h.clients {
		if client.ProjectID == projectID {
			clientCopy := &SSEClient{
				ID:          client.ID,
				ProjectID:   client.ProjectID,
				Environment: client.Environment,
				ServerID:    client.ServerID,
				LastPing:    client.LastPing,
				ConnectedAt: client.ConnectedAt,
			}
			clients = append(clients, clientCopy)
		}
	}
	
	return clients
}

// sendEvent sends an SSE event to a specific client
func (h *SSEHandler) sendEvent(client *SSEClient, event SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	
	// SSE format: "data: {json}\n\n"
	fmt.Fprintf(client.Writer, "data: %s\n\n", data)
	client.Flusher.Flush()
}

// pingClients sends periodic ping events to keep connections alive
func (h *SSEHandler) pingClients() {
	for {
		select {
		case <-h.pingTicker.C:
			pingEvent := SSEEvent{
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
func (h *SSEHandler) Close() {
	h.pingTicker.Stop()
	close(h.done)
	
	h.mu.Lock()
	for _, client := range h.clients {
		close(client.Done)
	}
	h.clients = make(map[string]*SSEClient)
	h.mu.Unlock()
}

// HandleEdgeServerStatus returns status of connected edge servers
func (h *SSEHandler) HandleEdgeServerStatus(c *gin.Context) {
	projectID := c.Query("project_id")
	
	var clients []*SSEClient
	if projectID != "" {
		// Project-specific view
		clients = h.GetClientsByProject(projectID)
	} else {
		// Global view - all edge servers
		clients = h.GetConnectedClients()
	}
	
	// Transform to response format
	servers := make([]map[string]interface{}, len(clients))
	for i, client := range clients {
		uptime := time.Since(client.ConnectedAt)
		servers[i] = map[string]interface{}{
			"id":               client.ServerID,
			"client_id":        client.ID,
			"project_id":       client.ProjectID,
			"environment":      client.Environment,
			"status":          "connected",
			"connected_at":     client.ConnectedAt.Format(time.RFC3339),
			"last_ping":        client.LastPing.Format(time.RFC3339),
			"uptime_seconds":   int(uptime.Seconds()),
			"uptime_human":     uptime.String(),
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"servers": servers,
		"total":   len(servers),
	})
}