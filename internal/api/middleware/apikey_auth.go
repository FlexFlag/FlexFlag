package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// apiKeyCache is a process-level in-memory cache for validated API keys.
// This eliminates the DB round-trip on the hot evaluation path.
// Keys are stored by their raw value (never logged); cache entries expire after TTL.
type apiKeyCache struct {
	mu      sync.RWMutex
	entries map[string]*apiKeyCacheEntry
}

type apiKeyCacheEntry struct {
	keyInfo   *types.ApiKey
	expiresAt time.Time
}

const apiKeyCacheTTL = 5 * time.Minute

var globalApiKeyCache = &apiKeyCache{
	entries: make(map[string]*apiKeyCacheEntry),
}

func init() {
	// Background goroutine to evict expired entries every minute.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			globalApiKeyCache.evictExpired()
		}
	}()
}

func (c *apiKeyCache) get(rawKey string) (*types.ApiKey, bool) {
	c.mu.RLock()
	entry, ok := c.entries[rawKey]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.keyInfo, true
}

func (c *apiKeyCache) set(rawKey string, keyInfo *types.ApiKey) {
	c.mu.Lock()
	c.entries[rawKey] = &apiKeyCacheEntry{
		keyInfo:   keyInfo,
		expiresAt: time.Now().Add(apiKeyCacheTTL),
	}
	c.mu.Unlock()
}

// InvalidateApiKey removes a key from the cache immediately (call on key deletion/rotation).
func InvalidateApiKey(rawKey string) {
	globalApiKeyCache.mu.Lock()
	delete(globalApiKeyCache.entries, rawKey)
	globalApiKeyCache.mu.Unlock()
}

func (c *apiKeyCache) evictExpired() {
	now := time.Now()
	c.mu.Lock()
	for k, e := range c.entries {
		if now.After(e.expiresAt) {
			delete(c.entries, k)
		}
	}
	c.mu.Unlock()
}

// ApiKeyAuth creates a middleware for API key authentication
func ApiKeyAuth(apiKeyRepo *postgres.ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		keyInfo, err := authenticateWithCache(c, apiKeyRepo, apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		setApiKeyContext(c, keyInfo)
		c.Next()
	}
}

// OptionalApiKeyAuth creates a middleware for optional API key authentication
// If API key is provided, it validates it; if not, request continues without auth
func OptionalApiKeyAuth(apiKeyRepo *postgres.ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.Next()
			return
		}

		keyInfo, err := authenticateWithCache(c, apiKeyRepo, apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		setApiKeyContext(c, keyInfo)
		c.Next()
	}
}

func authenticateWithCache(c *gin.Context, repo *postgres.ApiKeyRepository, rawKey string) (*types.ApiKey, error) {
	if cached, ok := globalApiKeyCache.get(rawKey); ok {
		return cached, nil
	}
	// Cache miss — validate against DB.
	keyInfo, err := repo.AuthenticateApiKey(c.Request.Context(), rawKey)
	if err != nil {
		return nil, err
	}
	globalApiKeyCache.set(rawKey, keyInfo)
	return keyInfo, nil
}

func setApiKeyContext(c *gin.Context, keyInfo *types.ApiKey) {
	c.Set("apiKey", keyInfo)
	c.Set("projectID", keyInfo.ProjectID)
	c.Set("environmentID", keyInfo.EnvironmentID)
	c.Set("environment", keyInfo.Environment.Key)
	c.Set("permissions", keyInfo.Permissions)
}

// RequirePermission creates a middleware that checks if the API key has the required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		perms, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		hasPermission := false
		for _, perm := range perms {
			if perm == "admin" || perm == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}
