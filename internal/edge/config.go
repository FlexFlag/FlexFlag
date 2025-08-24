package edge

import (
	"os"
	"strconv"
	"time"
)

// Config represents edge server configuration
type Config struct {
	Port         int         `json:"port"`
	Environment  string      `json:"environment"`
	HubURL       string      `json:"hub_url"`
	APIKey       string      `json:"api_key"`
	CacheConfig  CacheConfig `json:"cache_config"`
	SyncConfig   SyncConfig  `json:"sync_config"`
}

// SyncConfig contains synchronization settings
type SyncConfig struct {
	Type              string        `json:"type"`              // "websocket" or "sse"
	ReconnectInterval time.Duration `json:"reconnect_interval"`
	HeartbeatInterval time.Duration `json:"heartbeat_interval"`
	MaxRetries        int           `json:"max_retries"`
	EnableSSL         bool          `json:"enable_ssl"`
	BufferSize        int           `json:"buffer_size"`
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	port, _ := strconv.Atoi(getEnv("FLEXFLAG_EDGE_PORT", "8081"))
	maxFlags, _ := strconv.Atoi(getEnv("FLEXFLAG_EDGE_MAX_FLAGS", "10000"))
	maxAPIKeys, _ := strconv.Atoi(getEnv("FLEXFLAG_EDGE_MAX_API_KEYS", "1000"))
	maxRetries, _ := strconv.Atoi(getEnv("FLEXFLAG_EDGE_MAX_RETRIES", "5"))
	bufferSize, _ := strconv.Atoi(getEnv("FLEXFLAG_EDGE_BUFFER_SIZE", "1000"))

	cacheTTL, _ := time.ParseDuration(getEnv("FLEXFLAG_EDGE_CACHE_TTL", "1h"))
	cleanupInterval, _ := time.ParseDuration(getEnv("FLEXFLAG_EDGE_CLEANUP_INTERVAL", "5m"))
	reconnectInterval, _ := time.ParseDuration(getEnv("FLEXFLAG_EDGE_RECONNECT_INTERVAL", "30s"))
	heartbeatInterval, _ := time.ParseDuration(getEnv("FLEXFLAG_EDGE_HEARTBEAT_INTERVAL", "30s"))

	return &Config{
		Port:        port,
		Environment: getEnv("FLEXFLAG_EDGE_ENVIRONMENT", "production"),
		HubURL:      getEnv("FLEXFLAG_HUB_URL", "ws://localhost:8080"),
		APIKey:      getEnv("FLEXFLAG_EDGE_API_KEY", ""),
		CacheConfig: CacheConfig{
			TTL:             cacheTTL,
			MaxFlags:        maxFlags,
			MaxAPIKeys:      maxAPIKeys,
			CleanupInterval: cleanupInterval,
			EnableMetrics:   getEnv("FLEXFLAG_EDGE_ENABLE_METRICS", "true") == "true",
		},
		SyncConfig: SyncConfig{
			Type:              getEnv("FLEXFLAG_EDGE_SYNC_TYPE", "websocket"),
			ReconnectInterval: reconnectInterval,
			HeartbeatInterval: heartbeatInterval,
			MaxRetries:        maxRetries,
			EnableSSL:         getEnv("FLEXFLAG_EDGE_ENABLE_SSL", "false") == "true",
			BufferSize:        bufferSize,
		},
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}