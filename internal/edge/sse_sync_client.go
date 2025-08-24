package edge

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

// SSESyncClient manages real-time synchronization with the central hub using Server-Sent Events
type SSESyncClient struct {
	hubURL     string
	apiKey     string
	config     SyncConfig
	client     *http.Client
	cache      *FlagCache
	connected  bool
	lastSync   time.Time
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	updateChan chan *FlagUpdate
	resp       *http.Response
}

// SSEEvent represents an event received from SSE
type SSEEvent struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// SSEFlagUpdateEvent represents flag update data from SSE
type SSEFlagUpdateEvent struct {
	Type        string      `json:"type"`
	Action      string      `json:"action"` // create, update, delete
	Flag        *types.Flag `json:"flag"`
	ProjectID   string      `json:"project_id"`
	Environment string      `json:"environment"`
	Timestamp   time.Time   `json:"timestamp"`
}

// NewSSESyncClient creates a new SSE synchronization client
func NewSSESyncClient(hubURL, apiKey string) *SSESyncClient {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SSESyncClient{
		hubURL:     hubURL,
		apiKey:     apiKey,
		client:     &http.Client{Timeout: 0}, // No timeout for SSE connection
		connected:  false,
		lastSync:   time.Time{},
		ctx:        ctx,
		cancel:     cancel,
		updateChan: make(chan *FlagUpdate, 1000),
	}
}

// SetCache sets the flag cache for the sync client
func (sc *SSESyncClient) SetCache(cache *FlagCache) {
	sc.cache = cache
}

// SetConfig sets the sync configuration
func (sc *SSESyncClient) SetConfig(config SyncConfig) {
	sc.config = config
}

// StartRealtimeSync starts real-time synchronization with callback
func (sc *SSESyncClient) StartRealtimeSync(updateCallback func(*FlagUpdate)) {
	go sc.syncLoop(updateCallback)
	go sc.processUpdates(updateCallback)
}

// FullSync performs a complete synchronization with the hub
func (sc *SSESyncClient) FullSync() error {
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
		env := flag.Environment
		if env == "" {
			env = "production" // Default fallback
		}
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
	
	log.Printf("SSE Full sync completed: %d flags, %d API keys", len(syncResp.Flags), len(syncResp.APIKeys))
	return nil
}

// syncLoop maintains SSE connection and handles real-time updates
func (sc *SSESyncClient) syncLoop(updateCallback func(*FlagUpdate)) {
	for {
		select {
		case <-sc.ctx.Done():
			return
		default:
			if err := sc.connect(); err != nil {
				log.Printf("Failed to connect to hub via SSE: %v", err)
				time.Sleep(sc.config.ReconnectInterval)
				continue
			}
			
			sc.handleSSEEvents()
		}
	}
}

// connect establishes SSE connection to the hub
func (sc *SSESyncClient) connect() error {
	sseURL := sc.getSSEURL()
	
	req, err := http.NewRequestWithContext(sc.ctx, "GET", sseURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create SSE request: %w", err)
	}
	
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Authorization", "Bearer "+sc.apiKey)
	req.Header.Set("X-API-Key", sc.apiKey)
	
	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("SSE request failed: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return fmt.Errorf("SSE connection failed with status: %d", resp.StatusCode)
	}
	
	sc.mu.Lock()
	sc.resp = resp
	sc.connected = true
	sc.mu.Unlock()
	
	log.Printf("Connected to hub via SSE: %s", sseURL)
	return nil
}

// handleSSEEvents processes incoming SSE events
func (sc *SSESyncClient) handleSSEEvents() {
	defer sc.disconnect()
	
	scanner := bufio.NewScanner(sc.resp.Body)
	var eventType, eventData string
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Handle SSE event format
		if strings.HasPrefix(line, "event: ") {
			eventType = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			eventData = strings.TrimPrefix(line, "data: ")
		} else if line == "" && eventData != "" {
			// Empty line indicates end of event, process it
			sc.handleSSEEvent(eventType, eventData)
			eventType, eventData = "", ""
		}
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("SSE scanner error: %v", err)
	}
}

// handleSSEEvent processes individual SSE events
func (sc *SSESyncClient) handleSSEEvent(eventType, eventData string) {
	if eventData == "" {
		return
	}
	
	var event SSEEvent
	if err := json.Unmarshal([]byte(eventData), &event); err != nil {
		log.Printf("Failed to unmarshal SSE event: %v", err)
		return
	}
	
	switch event.Type {
	case "connected":
		log.Printf("SSE connection established: %s", eventData)
		
	case "ping":
		// Just log ping events
		log.Printf("Received SSE ping")
		
	case "flag_update":
		var flagUpdateEvent SSEFlagUpdateEvent
		if err := json.Unmarshal(event.Data, &flagUpdateEvent); err != nil {
			log.Printf("Failed to unmarshal flag update event: %v", err)
			return
		}
		
		// Convert to FlagUpdate format
		update := &FlagUpdate{
			FlagKey:     flagUpdateEvent.Flag.Key,
			Environment: flagUpdateEvent.Environment,
			Flag:        flagUpdateEvent.Flag,
			Operation:   flagUpdateEvent.Action,
			Timestamp:   flagUpdateEvent.Timestamp,
		}
		
		select {
		case sc.updateChan <- update:
		default:
			log.Printf("SSE update channel full, dropping update")
		}
		
	default:
		log.Printf("Unknown SSE event type: %s", event.Type)
	}
}

// processUpdates handles flag updates from the channel
func (sc *SSESyncClient) processUpdates(updateCallback func(*FlagUpdate)) {
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

// disconnect closes the SSE connection
func (sc *SSESyncClient) disconnect() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	
	if sc.resp != nil {
		sc.resp.Body.Close()
		sc.resp = nil
	}
	sc.connected = false
}

// IsConnected returns the connection status
func (sc *SSESyncClient) IsConnected() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.connected
}

// LastSyncTime returns the last successful sync time
func (sc *SSESyncClient) LastSyncTime() time.Time {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.lastSync
}

// Close stops the sync client
func (sc *SSESyncClient) Close() {
	sc.cancel()
	sc.disconnect()
	close(sc.updateChan)
}

// Helper methods

func (sc *SSESyncClient) getSSEURL() string {
	u, _ := url.Parse(sc.hubURL)
	
	// Convert WebSocket URLs to HTTP for SSE
	if u.Scheme == "ws" {
		u.Scheme = "http"
	} else if u.Scheme == "wss" {
		u.Scheme = "https"
	}
	
	u.Path = "/api/v1/edge/sync/sse"
	
	// Edge servers are global - only need server_id
	q := u.Query()
	q.Set("server_id", "edge-server-global") // Global edge server ID
	u.RawQuery = q.Encode()
	
	return u.String()
}

func (sc *SSESyncClient) getHTTPURL() string {
	u, _ := url.Parse(sc.hubURL)
	
	if u.Scheme == "ws" {
		u.Scheme = "http"
	} else if u.Scheme == "wss" {
		u.Scheme = "https"
	}
	
	return u.String()
}