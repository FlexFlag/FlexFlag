package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type EdgeSyncHandler struct {
	flagRepo     *postgres.FlagRepository
	apiKeyRepo   *postgres.ApiKeyRepository
	upgrader     websocket.Upgrader
	clients      map[*EdgeClient]bool
	clientsMutex sync.RWMutex
	broadcast    chan interface{}
}

// EdgeClient represents a connected edge server
type EdgeClient struct {
	conn         *websocket.Conn
	send         chan interface{}
	apiKey       string
	projectID    string
	environment  string
	id           string
	connectedAt  time.Time
	lastPingTime time.Time
	region       string
	version      string
	remoteAddr   string
}

func NewEdgeSyncHandler(flagRepo *postgres.FlagRepository, apiKeyRepo *postgres.ApiKeyRepository) *EdgeSyncHandler {
	h := &EdgeSyncHandler{
		flagRepo:   flagRepo,
		apiKeyRepo: apiKeyRepo,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for edge servers
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:   make(map[*EdgeClient]bool),
		broadcast: make(chan interface{}, 256),
	}
	
	// Start broadcast processor
	go h.processBroadcasts()
	
	return h
}

// BulkSync returns all flags and API keys for edge server synchronization
func (h *EdgeSyncHandler) BulkSync(c *gin.Context) {
	// Get optional pagination parameters
	limitStr := c.DefaultQuery("limit", "10000")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10000
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}
	
	// Get all flags (simplified - in production you'd want to filter by projects the edge server has access to)
	flags, err := h.flagRepo.GetAllFlags(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch flags: " + err.Error()})
		return
	}
	
	// Get all API keys (simplified - in production you'd want better filtering)
	apiKeys, err := h.apiKeyRepo.GetAllApiKeys(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys: " + err.Error()})
		return
	}
	
	// Convert API keys to edge format
	edgeAPIKeys := make([]*EdgeAPIKeyInfo, len(apiKeys))
	for i, key := range apiKeys {
		var expiresAt *string
		if key.ExpiresAt != nil {
			expiry := key.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
			expiresAt = &expiry
		}
		edgeAPIKeys[i] = &EdgeAPIKeyInfo{
			KeyHash:     key.KeyHash,
			ProjectID:   key.ProjectID,
			Environment: key.EnvironmentID, // Note: This should be environment key, not ID
			Permissions: key.Permissions,
			ExpiresAt:   expiresAt,
		}
	}
	
	response := BulkSyncResponse{
		Flags:      flags,
		APIKeys:    edgeAPIKeys,
		TotalCount: len(flags),
	}
	
	c.JSON(http.StatusOK, response)
}

// AuthenticateAPIKey validates an API key for edge server use
func (h *EdgeSyncHandler) AuthenticateAPIKey(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required"})
		return
	}
	
	keyInfo, err := h.apiKeyRepo.AuthenticateApiKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	
	// Convert to edge format
	var expiresAt *string
	if keyInfo.ExpiresAt != nil {
		expiry := keyInfo.ExpiresAt.Format("2006-01-02T15:04:05Z07:00")
		expiresAt = &expiry
	}
	edgeKeyInfo := &EdgeAPIKeyInfo{
		KeyHash:     keyInfo.KeyHash,
		ProjectID:   keyInfo.ProjectID,
		Environment: keyInfo.EnvironmentID, // Should be environment key
		Permissions: keyInfo.Permissions,
		ExpiresAt:   expiresAt,
	}
	
	c.JSON(http.StatusOK, gin.H{"api_key": edgeKeyInfo})
}

// GetEdgeServersStatus returns the status of all connected edge servers
func (h *EdgeSyncHandler) GetEdgeServersStatus(c *gin.Context) {
	h.clientsMutex.RLock()
	defer h.clientsMutex.RUnlock()
	
	var servers []EdgeServerStatus
	regions := make(map[string]int)
	connected := 0
	
	now := time.Now()
	for client := range h.clients {
		// Determine status based on last ping time
		status := "connected"
		if now.Sub(client.lastPingTime) > 2*time.Minute {
			status = "unhealthy"
		}
		
		uptime := now.Sub(client.connectedAt)
		uptimeStr := fmt.Sprintf("%.0fs", uptime.Seconds())
		if uptime.Minutes() >= 1 {
			uptimeStr = fmt.Sprintf("%.0fm", uptime.Minutes())
		}
		if uptime.Hours() >= 1 {
			uptimeStr = fmt.Sprintf("%.1fh", uptime.Hours())
		}
		
		server := EdgeServerStatus{
			ID:           client.id,
			ProjectID:    client.projectID,
			Environment:  client.environment,
			Region:       client.region,
			Version:      client.version,
			RemoteAddr:   client.remoteAddr,
			ConnectedAt:  client.connectedAt,
			LastPingTime: client.lastPingTime,
			Uptime:       uptimeStr,
			Status:       status,
		}
		
		servers = append(servers, server)
		regions[client.region]++
		if status == "connected" {
			connected++
		}
	}
	
	response := EdgeServersStatusResponse{
		Servers:      servers,
		TotalCount:   len(servers),
		Connected:    connected,
		Disconnected: len(servers) - connected,
		Regions:      regions,
	}
	
	c.JSON(http.StatusOK, response)
}

// Edge-specific types
type EdgeAPIKeyInfo struct {
	KeyHash     string     `json:"key_hash"`
	ProjectID   string     `json:"project_id"`
	Environment string     `json:"environment"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *string    `json:"expires_at,omitempty"`
}

type BulkSyncResponse struct {
	Flags      []*types.Flag      `json:"flags"`
	APIKeys    []*EdgeAPIKeyInfo  `json:"api_keys"`
	TotalCount int                `json:"total_count"`
}

// EdgeServerStatus represents the status of an edge server
type EdgeServerStatus struct {
	ID           string    `json:"id"`
	ProjectID    string    `json:"project_id"`
	Environment  string    `json:"environment"`
	Region       string    `json:"region"`
	Version      string    `json:"version"`
	RemoteAddr   string    `json:"remote_addr"`
	ConnectedAt  time.Time `json:"connected_at"`
	LastPingTime time.Time `json:"last_ping_time"`
	Uptime       string    `json:"uptime"`
	Status       string    `json:"status"` // "connected", "disconnected", "unhealthy"
}

// EdgeServersStatusResponse represents the response for edge servers status
type EdgeServersStatusResponse struct {
	Servers     []EdgeServerStatus `json:"servers"`
	TotalCount  int                `json:"total_count"`
	Connected   int                `json:"connected"`
	Disconnected int               `json:"disconnected"`
	Regions     map[string]int     `json:"regions"`
}

// WebSocketSync handles WebSocket connections for real-time sync
func (h *EdgeSyncHandler) WebSocketSync(c *gin.Context) {
	// Validate API key
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
		return
	}
	
	// Authenticate the API key
	keyInfo, err := h.apiKeyRepo.AuthenticateApiKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}
	
	// Upgrade to WebSocket
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	
	// Generate unique ID for this edge server connection
	clientID := fmt.Sprintf("edge-%s-%d", keyInfo.ProjectID[:8], time.Now().Unix())
	
	// Extract region from headers or use default
	region := c.GetHeader("X-Edge-Region")
	if region == "" {
		region = "unknown"
	}
	
	version := c.GetHeader("X-Edge-Version")
	if version == "" {
		version = "1.0.0"
	}

	// Create client
	client := &EdgeClient{
		conn:         conn,
		send:         make(chan interface{}, 256),
		apiKey:       apiKey,
		projectID:    keyInfo.ProjectID,
		environment:  keyInfo.EnvironmentID,
		id:           clientID,
		connectedAt:  time.Now(),
		lastPingTime: time.Now(),
		region:       region,
		version:      version,
		remoteAddr:   conn.RemoteAddr().String(),
	}
	
	// Register client
	h.registerClient(client)
	
	// Start client goroutines
	go client.writePump()
	go client.readPump(h)
	
	log.Printf("Edge server connected via WebSocket from %s", conn.RemoteAddr())
}

// registerClient adds a new edge client
func (h *EdgeSyncHandler) registerClient(client *EdgeClient) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	h.clients[client] = true
}

// unregisterClient removes an edge client
func (h *EdgeSyncHandler) unregisterClient(client *EdgeClient) {
	h.clientsMutex.Lock()
	defer h.clientsMutex.Unlock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
}

// processBroadcasts sends updates to all connected clients
func (h *EdgeSyncHandler) processBroadcasts() {
	for {
		update := <-h.broadcast
		h.clientsMutex.RLock()
		for client := range h.clients {
			select {
			case client.send <- update:
			default:
				// Client's send channel is full, skip
			}
		}
		h.clientsMutex.RUnlock()
	}
}

// BroadcastFlagUpdate sends a flag update to all edge servers
func (h *EdgeSyncHandler) BroadcastFlagUpdate(flag *types.Flag, action string) {
	// Create FlagUpdate structure that matches what the edge server expects
	flagUpdate := map[string]interface{}{
		"flag_key":     flag.Key,
		"environment":  flag.Environment,
		"flag":         flag,
		"operation":    action, // "create", "update", "delete"
		"timestamp":    time.Now().Format(time.RFC3339),
	}
	
	update := map[string]interface{}{
		"type": "flag_update",
		"data": flagUpdate,
	}
	
	select {
	case h.broadcast <- update:
	default:
		log.Printf("Broadcast channel full, dropping update")
	}
}

// BroadcastAPIKeyUpdate sends an API key update to all edge servers
func (h *EdgeSyncHandler) BroadcastAPIKeyUpdate(apiKey *types.ApiKey, action string) {
	update := map[string]interface{}{
		"type": "api_key_update",
		"data": map[string]interface{}{
			"api_key": map[string]interface{}{
				"key_hash":     apiKey.KeyHash,
				"project_id":   apiKey.ProjectID,
				"environment":  apiKey.EnvironmentID,
				"permissions":  apiKey.Permissions,
				"expires_at":   apiKey.ExpiresAt,
			},
			"action":    action, // "create", "update", "delete"
			"timestamp": time.Now().Unix(),
		},
	}
	
	select {
	case h.broadcast <- update:
	default:
		log.Printf("Broadcast channel full, dropping update")
	}
}

// Client methods

// readPump handles incoming messages from the edge client
func (c *EdgeClient) readPump(h *EdgeSyncHandler) {
	defer func() {
		h.unregisterClient(c)
		c.conn.Close()
	}()
	
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.lastPingTime = time.Now()
		return nil
	})
	
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Handle incoming messages (heartbeat, requests, etc.)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		
		if msgType, ok := msg["type"].(string); ok {
			switch msgType {
			case "ping":
				// Respond with pong
				c.send <- map[string]string{"type": "pong"}
			case "request_sync":
				// Edge server requesting full sync - could implement this
				log.Printf("Edge server requested sync")
			}
		}
	}
}

// writePump handles sending messages to the edge client
func (c *EdgeClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.conn.WriteJSON(message); err != nil {
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}