package edge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gorilla/websocket"
)

// SyncClient manages real-time synchronization with the central hub
type SyncClient struct {
	hubURL     string
	apiKey     string
	config     SyncConfig
	conn       *websocket.Conn
	cache      *FlagCache
	connected  bool
	lastSync   time.Time
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	updateChan chan *FlagUpdate
}

// SyncMessage represents a message from the hub
type SyncMessage struct {
	Type      string          `json:"type"`      // "flag_update", "api_key_update", "bulk_sync"
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
	EventID   string          `json:"event_id"`
}

// BulkSyncResponse represents bulk sync data from hub
type BulkSyncResponse struct {
	Flags      []*types.Flag    `json:"flags"`
	APIKeys    []*APIKeyInfo    `json:"api_keys"`
	Timestamp  time.Time        `json:"timestamp"`
	TotalCount int              `json:"total_count"`
}

// NewSyncClient creates a new synchronization client
func NewSyncClient(hubURL, apiKey string) *SyncClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SyncClient{
		hubURL:     hubURL,
		apiKey:     apiKey,
		connected:  false,
		lastSync:   time.Time{},
		ctx:        ctx,
		cancel:     cancel,
		updateChan: make(chan *FlagUpdate, 1000),
	}
}

// SetCache sets the flag cache for the sync client
func (sc *SyncClient) SetCache(cache *FlagCache) {
	sc.cache = cache
}

// SetConfig sets the sync configuration
func (sc *SyncClient) SetConfig(config SyncConfig) {
	sc.config = config
}

// StartRealtimeSync starts real-time synchronization with callback
func (sc *SyncClient) StartRealtimeSync(updateCallback func(*FlagUpdate)) {
	go sc.syncLoop(updateCallback)
	go sc.processUpdates(updateCallback)
}

// FullSync performs a complete synchronization with the hub
func (sc *SyncClient) FullSync() error {
	// Use HTTP for bulk sync as it's more reliable
	httpURL := sc.getHTTPURL()
	
	req, err := http.NewRequestWithContext(sc.ctx, "GET", httpURL+"/api/v1/edge/sync", nil)
	if err != nil {
		return fmt.Errorf("failed to create sync request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("X-API-Key", sc.apiKey)
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("sync request failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sync failed with status: %d", resp.StatusCode)
	}
	
	var syncResp BulkSyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return fmt.Errorf("failed to decode sync response: %w", err)
	}
	
	// Update cache with all flags
	flagsByEnv := make(map[string][]*types.Flag)
	for _, flag := range syncResp.Flags {
		env := "production" // Default, should be extracted from flag context
		flagsByEnv[env] = append(flagsByEnv[env], flag)
	}
	
	// Update cache for each environment
	for env, flags := range flagsByEnv {
		sc.cache.BulkUpdateFlags(flags, env)
	}
	
	// Update API keys
	for _, keyInfo := range syncResp.APIKeys {
		sc.cache.UpdateAPIKey("", keyInfo) // Key hash would be needed here
	}
	
	sc.mu.Lock()
	sc.lastSync = time.Now()
	sc.mu.Unlock()
	
	log.Printf("Full sync completed: %d flags, %d API keys", len(syncResp.Flags), len(syncResp.APIKeys))
	return nil
}

// syncLoop maintains WebSocket connection and handles real-time updates
func (sc *SyncClient) syncLoop(updateCallback func(*FlagUpdate)) {
	for {
		select {
		case <-sc.ctx.Done():
			return
		default:
			if err := sc.connect(); err != nil {
				log.Printf("Failed to connect to hub: %v", err)
				time.Sleep(sc.config.ReconnectInterval)
				continue
			}
			
			sc.handleMessages()
		}
	}
}

// connect establishes WebSocket connection to the hub
func (sc *SyncClient) connect() error {
	wsURL := sc.getWebSocketURL()
	
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}
	
	header := http.Header{}
	header.Set("Authorization", "Bearer "+sc.apiKey)
	header.Set("X-API-Key", sc.apiKey)
	
	conn, _, err := dialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("WebSocket dial failed: %w", err)
	}
	
	sc.mu.Lock()
	sc.conn = conn
	sc.connected = true
	sc.mu.Unlock()
	
	log.Printf("Connected to hub: %s", wsURL)
	return nil
}

// handleMessages processes incoming WebSocket messages
func (sc *SyncClient) handleMessages() {
	defer sc.disconnect()
	
	// Set up ping/pong handlers
	sc.conn.SetPongHandler(func(string) error {
		_ = sc.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Start ping ticker
	ticker := time.NewTicker(sc.config.HeartbeatInterval)
	defer ticker.Stop()
	
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := sc.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					log.Printf("Ping failed: %v", err)
					return
				}
			case <-sc.ctx.Done():
				return
			}
		}
	}()
	
	// Read messages
	for {
		_, messageBytes, err := sc.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}
		
		var msg SyncMessage
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		
		sc.handleSyncMessage(&msg)
	}
}

// handleSyncMessage processes different types of sync messages
func (sc *SyncClient) handleSyncMessage(msg *SyncMessage) {
	switch msg.Type {
	case "flag_update":
		var update FlagUpdate
		if err := json.Unmarshal(msg.Data, &update); err != nil {
			log.Printf("Failed to unmarshal flag update: %v", err)
			return
		}
		
		select {
		case sc.updateChan <- &update:
		default:
			log.Printf("Update channel full, dropping update")
		}
		
	case "api_key_update":
		var keyUpdate struct {
			APIKey  string       `json:"api_key"`
			KeyInfo *APIKeyInfo  `json:"key_info"`
			Action  string       `json:"action"` // "create", "update", "delete"
		}
		
		if err := json.Unmarshal(msg.Data, &keyUpdate); err != nil {
			log.Printf("Failed to unmarshal API key update: %v", err)
			return
		}
		
		if keyUpdate.Action == "delete" {
			// Handle API key deletion - would need cache method
		} else {
			sc.cache.UpdateAPIKey(keyUpdate.APIKey, keyUpdate.KeyInfo)
		}
		
	case "bulk_sync":
		// Trigger full sync
		go func() {
			if err := sc.FullSync(); err != nil {
				log.Printf("Bulk sync failed: %v", err)
			}
		}()
	}
}

// processUpdates handles flag updates from the channel
func (sc *SyncClient) processUpdates(updateCallback func(*FlagUpdate)) {
	for {
		select {
		case update := <-sc.updateChan:
			if sc.cache != nil {
				sc.cache.UpdateFlag(update)
			}
			if updateCallback != nil {
				updateCallback(update)
			}
		case <-sc.ctx.Done():
			return
		}
	}
}

// disconnect closes the WebSocket connection
func (sc *SyncClient) disconnect() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.conn != nil {
		sc.conn.Close()
		sc.conn = nil
	}
	sc.connected = false
}

// IsConnected returns the connection status
func (sc *SyncClient) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected
}

// LastSyncTime returns the last successful sync time
func (sc *SyncClient) LastSyncTime() time.Time {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.lastSync
}

// Close stops the sync client
func (sc *SyncClient) Close() {
	sc.cancel()
	sc.disconnect()
	close(sc.updateChan)
}

// Helper methods

func (sc *SyncClient) getWebSocketURL() string {
	u, _ := url.Parse(sc.hubURL)
	
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}
	
	u.Path = "/api/v1/edge/sync/ws"
	return u.String()
}

func (sc *SyncClient) getHTTPURL() string {
	u, _ := url.Parse(sc.hubURL)
	
	if u.Scheme == "ws" {
		u.Scheme = "http"
	} else if u.Scheme == "wss" {
		u.Scheme = "https"
	}
	
	return u.String()
}