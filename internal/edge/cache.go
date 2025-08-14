package edge

import (
	"fmt"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

// FlagCache provides ultra-fast in-memory flag caching for edge servers
type FlagCache struct {
	flags    map[string]*types.Flag // key: "flag_key:environment"
	apiKeys  map[string]*APIKeyInfo  // key: api_key_hash
	stats    *CacheStats
	mu       sync.RWMutex
	config   CacheConfig
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	TTL               time.Duration `json:"ttl"`
	MaxFlags          int           `json:"max_flags"`
	MaxAPIKeys        int           `json:"max_api_keys"`
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	EnableMetrics     bool          `json:"enable_metrics"`
}

// CacheStats tracks cache performance metrics
type CacheStats struct {
	FlagCount     int64     `json:"flag_count"`
	APIKeyCount   int64     `json:"api_key_count"`
	HitRate       float64   `json:"hit_rate"`
	TotalRequests int64     `json:"total_requests"`
	CacheHits     int64     `json:"cache_hits"`
	CacheMisses   int64     `json:"cache_misses"`
	LastUpdated   time.Time `json:"last_updated"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// APIKeyInfo contains cached API key information
type APIKeyInfo struct {
	ProjectID     string    `json:"project_id"`
	Environment   string    `json:"environment"`
	Permissions   []string  `json:"permissions"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	LastUsed      time.Time  `json:"last_used"`
}

// FlagUpdate represents a flag change event
type FlagUpdate struct {
	FlagKey     string           `json:"flag_key"`
	Environment string           `json:"environment"`
	Flag        *types.Flag      `json:"flag,omitempty"`
	Operation   string           `json:"operation"` // create, update, delete
	Timestamp   time.Time        `json:"timestamp"`
}

// NewFlagCache creates a new flag cache instance
func NewFlagCache(config CacheConfig) *FlagCache {
	cache := &FlagCache{
		flags:   make(map[string]*types.Flag),
		apiKeys: make(map[string]*APIKeyInfo),
		stats: &CacheStats{
			LastUpdated: time.Now(),
		},
		config: config,
	}
	
	// Start cleanup goroutine
	if config.CleanupInterval > 0 {
		go cache.startCleanup()
	}
	
	return cache
}

// GetFlag retrieves a flag from cache (ultra-fast lookup)
func (c *FlagCache) GetFlag(flagKey, environment string) *types.Flag {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	key := c.cacheKey(flagKey, environment)
	flag, exists := c.flags[key]
	
	// Update stats
	if c.config.EnableMetrics {
		c.stats.TotalRequests++
		if exists {
			c.stats.CacheHits++
		} else {
			c.stats.CacheMisses++
		}
		c.updateHitRate()
	}
	
	if !exists {
		return nil
	}
	
	return flag
}

// UpdateFlag updates or creates a flag in cache
func (c *FlagCache) UpdateFlag(update *FlagUpdate) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	key := c.cacheKey(update.FlagKey, update.Environment)
	
	switch update.Operation {
	case "create", "update":
		if update.Flag != nil {
			c.flags[key] = update.Flag
			c.stats.FlagCount = int64(len(c.flags))
		}
	case "delete":
		delete(c.flags, key)
		c.stats.FlagCount = int64(len(c.flags))
	}
	
	c.stats.LastUpdated = time.Now()
}

// BulkUpdateFlags updates multiple flags at once (for initial sync)
func (c *FlagCache) BulkUpdateFlags(flags []*types.Flag, environment string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Clear existing flags for this environment
	for key := range c.flags {
		if c.extractEnvironment(key) == environment {
			delete(c.flags, key)
		}
	}
	
	// Add new flags
	for _, flag := range flags {
		key := c.cacheKey(flag.Key, environment)
		c.flags[key] = flag
	}
	
	c.stats.FlagCount = int64(len(c.flags))
	c.stats.LastUpdated = time.Now()
}

// ValidateAPIKey validates an API key and returns cached info
func (c *FlagCache) ValidateAPIKey(apiKey string) *APIKeyInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	keyInfo, exists := c.apiKeys[apiKey]
	if !exists {
		return nil
	}
	
	// Check expiration
	if keyInfo.ExpiresAt != nil && keyInfo.ExpiresAt.Before(time.Now()) {
		return nil
	}
	
	// Update last used (in a separate goroutine to avoid blocking)
	go func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		keyInfo.LastUsed = time.Now()
	}()
	
	return keyInfo
}

// UpdateAPIKey updates API key information
func (c *FlagCache) UpdateAPIKey(apiKey string, keyInfo *APIKeyInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.apiKeys[apiKey] = keyInfo
	c.stats.APIKeyCount = int64(len(c.apiKeys))
}

// GetStats returns current cache statistics
func (c *FlagCache) GetStats() *CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	// Calculate uptime
	uptime := time.Since(c.stats.LastUpdated).Seconds()
	
	return &CacheStats{
		FlagCount:     c.stats.FlagCount,
		APIKeyCount:   c.stats.APIKeyCount,
		HitRate:       c.stats.HitRate,
		TotalRequests: c.stats.TotalRequests,
		CacheHits:     c.stats.CacheHits,
		CacheMisses:   c.stats.CacheMisses,
		LastUpdated:   c.stats.LastUpdated,
		UptimeSeconds: int64(uptime),
	}
}

// Clear removes all cached data
func (c *FlagCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.flags = make(map[string]*types.Flag)
	c.apiKeys = make(map[string]*APIKeyInfo)
	c.stats.FlagCount = 0
	c.stats.APIKeyCount = 0
	c.stats.LastUpdated = time.Now()
}

// Helper methods

func (c *FlagCache) cacheKey(flagKey, environment string) string {
	return fmt.Sprintf("%s:%s", flagKey, environment)
}

func (c *FlagCache) extractEnvironment(cacheKey string) string {
	// Extract environment from cache key format "flag_key:environment"
	for i := len(cacheKey) - 1; i >= 0; i-- {
		if cacheKey[i] == ':' {
			return cacheKey[i+1:]
		}
	}
	return ""
}

func (c *FlagCache) updateHitRate() {
	if c.stats.TotalRequests > 0 {
		c.stats.HitRate = float64(c.stats.CacheHits) / float64(c.stats.TotalRequests)
	}
}

func (c *FlagCache) startCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		c.cleanupExpiredKeys()
	}
}

func (c *FlagCache) cleanupExpiredKeys() {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	
	// Clean up expired API keys
	for key, keyInfo := range c.apiKeys {
		if keyInfo.ExpiresAt != nil && keyInfo.ExpiresAt.Before(now) {
			delete(c.apiKeys, key)
		}
	}
	
	c.stats.APIKeyCount = int64(len(c.apiKeys))
}