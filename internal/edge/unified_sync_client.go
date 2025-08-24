package edge

import (
	"log"
	"time"
)

// SyncClientInterface defines the interface for synchronization clients
type SyncClientInterface interface {
	SetCache(cache *FlagCache)
	SetConfig(config SyncConfig)
	StartRealtimeSync(updateCallback func(*FlagUpdate))
	FullSync() error
	IsConnected() bool
	LastSyncTime() time.Time
	Close()
}

// UnifiedSyncClient manages sync with hub using either WebSocket or SSE
type UnifiedSyncClient struct {
	syncType   string // "websocket" or "sse"
	wsClient   *SyncClient
	sseClient  *SSESyncClient
	activeClient SyncClientInterface
}

// NewUnifiedSyncClient creates a new unified sync client
func NewUnifiedSyncClient(hubURL, apiKey, syncType string) *UnifiedSyncClient {
	client := &UnifiedSyncClient{
		syncType: syncType,
	}
	
	switch syncType {
	case "sse":
		client.sseClient = NewSSESyncClient(hubURL, apiKey)
		client.activeClient = client.sseClient
		log.Printf("Using SSE for synchronization")
	case "websocket":
		fallthrough
	default:
		client.wsClient = NewSyncClient(hubURL, apiKey)
		client.activeClient = client.wsClient
		client.syncType = "websocket"
		log.Printf("Using WebSocket for synchronization")
	}
	
	return client
}

// SetCache sets the flag cache for the active sync client
func (uc *UnifiedSyncClient) SetCache(cache *FlagCache) {
	uc.activeClient.SetCache(cache)
}

// SetConfig sets the sync configuration for the active client
func (uc *UnifiedSyncClient) SetConfig(config SyncConfig) {
	uc.activeClient.SetConfig(config)
}

// StartRealtimeSync starts real-time synchronization with callback
func (uc *UnifiedSyncClient) StartRealtimeSync(updateCallback func(*FlagUpdate)) {
	uc.activeClient.StartRealtimeSync(updateCallback)
}

// FullSync performs a complete synchronization with the hub
func (uc *UnifiedSyncClient) FullSync() error {
	return uc.activeClient.FullSync()
}

// IsConnected returns the connection status
func (uc *UnifiedSyncClient) IsConnected() bool {
	return uc.activeClient.IsConnected()
}

// LastSyncTime returns the last successful sync time
func (uc *UnifiedSyncClient) LastSyncTime() time.Time {
	return uc.activeClient.LastSyncTime()
}

// Close stops the active sync client
func (uc *UnifiedSyncClient) Close() {
	uc.activeClient.Close()
}

// GetSyncType returns the current sync type being used
func (uc *UnifiedSyncClient) GetSyncType() string {
	return uc.syncType
}

// SwitchToSSE switches to SSE-based synchronization (requires restart)
func (uc *UnifiedSyncClient) SwitchToSSE(hubURL, apiKey string) error {
	if uc.syncType == "sse" {
		return nil // Already using SSE
	}
	
	// Stop current client
	uc.activeClient.Close()
	
	// Switch to SSE
	uc.sseClient = NewSSESyncClient(hubURL, apiKey)
	uc.activeClient = uc.sseClient
	uc.syncType = "sse"
	
	log.Printf("Switched to SSE synchronization")
	return nil
}

// SwitchToWebSocket switches to WebSocket-based synchronization (requires restart)
func (uc *UnifiedSyncClient) SwitchToWebSocket(hubURL, apiKey string) error {
	if uc.syncType == "websocket" {
		return nil // Already using WebSocket
	}
	
	// Stop current client
	uc.activeClient.Close()
	
	// Switch to WebSocket
	uc.wsClient = NewSyncClient(hubURL, apiKey)
	uc.activeClient = uc.wsClient
	uc.syncType = "websocket"
	
	log.Printf("Switched to WebSocket synchronization")
	return nil
}