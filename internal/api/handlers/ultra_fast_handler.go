package handlers

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"net/http"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// PrecomputedFlag contains pre-processed flag data for ultra-fast evaluation
type PrecomputedFlag struct {
	Key              string
	Enabled          bool
	DefaultValue     interface{} // Pre-unmarshaled
	DefaultJSON      []byte      // Pre-marshaled for response
	Type             types.FlagType
	HasTargeting     bool
	Variations       []types.Variation
	VariationValues  map[string]interface{} // variationID -> pre-unmarshaled value
	Targeting        *types.TargetingConfig
	LastUpdated      time.Time
}

// UltraFastHandler optimizes for absolute minimum latency
type UltraFastHandler struct {
	repo           storage.FlagRepository
	flags          map[string]*PrecomputedFlag // flag_key:env -> PrecomputedFlag
	responseCaches map[string]*CachedResponse  // request_hash -> response
	mu             sync.RWMutex
	cacheMu        sync.RWMutex
	preloadDone    bool
}

type CachedResponse struct {
	Response  []byte
	ExpiresAt time.Time
}

func NewUltraFastHandler(repo storage.FlagRepository) *UltraFastHandler {
	handler := &UltraFastHandler{
		repo:           repo,
		flags:          make(map[string]*PrecomputedFlag),
		responseCaches: make(map[string]*CachedResponse),
	}
	
	// Preload all flags on startup
	go handler.preloadFlags()
	
	// Start cleanup goroutine
	go handler.cleanupExpiredCaches()
	
	return handler
}

func (h *UltraFastHandler) preloadFlags() {
	// Load flags for common environments
	environments := []string{"production", "staging", "development"}
	ctx := context.Background()
	
	for _, env := range environments {
		flags, err := h.repo.List(ctx, env)
		if err != nil {
			continue
		}
		
		h.mu.Lock()
		for _, flag := range flags {
			h.flags[flag.Key+":"+env] = buildPrecomputed(flag)
		}
		h.mu.Unlock()
	}
	h.mu.Lock()
	h.preloadDone = true
	h.mu.Unlock()
}

func (h *UltraFastHandler) UltraFastEvaluate(c *gin.Context) {
	startTime := time.Now()
	
	// Fast JSON parsing using unsafe operations
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	environment := c.DefaultQuery("environment", "production")
	
	// If API key authentication is used, override environment from key
	if apiKeyEnv, exists := c.Get("environment"); exists {
		environment = apiKeyEnv.(string)
	}
	
	cacheKey := h.generateCacheKey(&req, environment)
	
	// Check response cache first
	h.cacheMu.RLock()
	cached, exists := h.responseCaches[cacheKey]
	h.cacheMu.RUnlock()
	
	if exists && time.Now().Before(cached.ExpiresAt) {
		// Ultra-fast cached response - just return bytes
		c.Data(http.StatusOK, "application/json", cached.Response)
		return
	}
	
	// Get precomputed flag
	flagKey := req.FlagKey + ":" + environment
	h.mu.RLock()
	flag, exists := h.flags[flagKey]
	h.mu.RUnlock()
	
	if !exists {
		// Flag not in cache - fetch and cache it
		projectID := c.Query("project_id")
		var dbFlag *types.Flag
		var err error
		
		if projectID != "" {
			dbFlag, err = h.repo.GetByProjectKey(c.Request.Context(), projectID, req.FlagKey, environment)
		} else {
			dbFlag, err = h.repo.GetByKey(c.Request.Context(), req.FlagKey, environment)
		}
		
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
			return
		}
		
		flag = buildPrecomputed(dbFlag)
		
		h.mu.Lock()
		h.flags[flagKey] = flag
		h.mu.Unlock()
	}
	
	// Ultra-fast evaluation
	var responseValue interface{}
	var reason string
	var isDefault bool
	
	if !flag.Enabled {
		// Disabled flag - return precomputed default
		responseValue = flag.DefaultValue
		reason = "flag_disabled"
		isDefault = true
	} else if !flag.HasTargeting {
		// No targeting rules - return default
		responseValue = flag.DefaultValue
		reason = "default"
		isDefault = true
	} else {
		// Has targeting - do minimal evaluation
		if matched := h.fastTargetingEvaluation(flag, &req); matched != nil {
			responseValue = matched
			reason = "rule_match"
			isDefault = false
		} else {
			responseValue = flag.DefaultValue
			reason = "default"
			isDefault = true
		}
	}
	
	evalTime := float64(time.Since(startTime).Nanoseconds()) / 1_000_000.0
	
	// Build response using pre-allocated struct
	response := map[string]interface{}{
		"flag_key":        flag.Key,
		"value":          responseValue,
		"reason":         reason,
		"default":        isDefault,
		"evaluation_time_ms": evalTime,
		"timestamp":      time.Now(),
	}
	
	// Marshal once and cache the response bytes
	responseBytes, _ := json.Marshal(response)
	
	// Cache the response for identical requests
	h.cacheMu.Lock()
	h.responseCaches[cacheKey] = &CachedResponse{
		Response:  responseBytes,
		ExpiresAt: time.Now().Add(30 * time.Second), // Short cache for responses
	}
	h.cacheMu.Unlock()
	
	// Return pre-marshaled bytes for maximum speed
	c.Data(http.StatusOK, "application/json", responseBytes)
}

func (h *UltraFastHandler) fastTargetingEvaluation(flag *PrecomputedFlag, req *EvaluateRequest) interface{} {
	if flag.Targeting == nil {
		return nil
	}

	// Attribute-based rules — use pre-unmarshaled variation values (no json.Unmarshal on hot path)
	for _, rule := range flag.Targeting.Rules {
		if value, exists := req.Attributes[rule.Attribute]; exists {
			if h.matchesRule(toString(value), &rule) {
				if v, ok := flag.VariationValues[rule.Variation]; ok {
					return v
				}
			}
		}
	}

	// Percentage rollout — FNV hash, no crypto
	if flag.Targeting.Rollout != nil {
		bucket := h.fastHash(req.UserID+flag.Key) % 100
		cumulative := 0
		for _, vr := range flag.Targeting.Rollout.Variations {
			cumulative += vr.Weight
			if bucket < cumulative {
				if v, ok := flag.VariationValues[vr.VariationID]; ok {
					return v
				}
			}
		}
	}

	return nil
}

func (h *UltraFastHandler) matchesRule(value string, rule *types.TargetingRule) bool {
	switch rule.Operator {
	case "eq":
		return contains(rule.Values, value)
	case "in":
		return contains(rule.Values, value)
	case "contains":
		return len(rule.Values) > 0 && contains([]string{value}, rule.Values[0])
	default:
		return false
	}
}

// fastHash uses FNV-1a — ~10x faster than MD5, sufficient for bucketing.
func (h *UltraFastHandler) fastHash(s string) int {
	if s == "" {
		return 0
	}
	data := (*[1 << 20]byte)(unsafe.Pointer(unsafe.StringData(s)))[:len(s)]
	h64 := fnv.New64a()
	_, _ = h64.Write(data)
	return int(h64.Sum64() % 100)
}

// generateCacheKey builds a cache key using FNV-1a — no crypto, no allocations beyond string concat.
func (h *UltraFastHandler) generateCacheKey(req *EvaluateRequest, environment string) string {
	hf := fnv.New64a()
	_, _ = hf.Write([]byte(req.FlagKey))
	_, _ = hf.Write([]byte{':'})
	_, _ = hf.Write([]byte(environment))
	_, _ = hf.Write([]byte{':'})
	_, _ = hf.Write([]byte(req.UserID))
	for k, v := range req.Attributes {
		_, _ = hf.Write([]byte{':'})
		_, _ = hf.Write([]byte(k))
		_, _ = hf.Write([]byte{'='})
		_, _ = hf.Write([]byte(toString(v)))
	}
	return strconv.FormatUint(hf.Sum64(), 36) // base-36 is compact
}

func (h *UltraFastHandler) cleanupExpiredCaches() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		h.cacheMu.Lock()
		for key, cached := range h.responseCaches {
			if now.After(cached.ExpiresAt) {
				delete(h.responseCaches, key)
			}
		}
		h.cacheMu.Unlock()
	}
}

// RefreshFlag refreshes a specific flag in the cache
func (h *UltraFastHandler) RefreshFlag(flagKey, environment string) {
	ctx := context.Background()
	// Try to get the flag - use GetByKey since we don't have project context here
	flag, err := h.repo.GetByKey(ctx, flagKey, environment)
	if err != nil {
		// Flag might have been deleted - remove from cache
		h.mu.Lock()
		delete(h.flags, flagKey+":"+environment)
		h.mu.Unlock()
		
		// Also clear related response caches
		h.cacheMu.Lock()
		for key := range h.responseCaches {
			if contains([]string{key}, flagKey) {
				delete(h.responseCaches, key)
			}
		}
		h.cacheMu.Unlock()
		return
	}
	
	precomputed := buildPrecomputed(flag)

	h.mu.Lock()
	h.flags[flagKey+":"+environment] = precomputed
	h.mu.Unlock()
	
	// Clear related response caches since flag changed
	h.cacheMu.Lock()
	for key := range h.responseCaches {
		if contains([]string{key}, flagKey) {
			delete(h.responseCaches, key)
		}
	}
	h.cacheMu.Unlock()
}

// RefreshAllFlags refreshes all flags in the cache
func (h *UltraFastHandler) RefreshAllFlags() {
	go h.preloadFlags()
	
	// Clear all response caches
	h.cacheMu.Lock()
	h.responseCaches = make(map[string]*CachedResponse)
	h.cacheMu.Unlock()
}

func (h *UltraFastHandler) GetStats(c *gin.Context) {
	h.mu.RLock()
	flagCount := len(h.flags)
	preloadComplete := h.preloadDone
	h.mu.RUnlock()
	
	h.cacheMu.RLock()
	cacheCount := len(h.responseCaches)
	h.cacheMu.RUnlock()
	
	c.JSON(http.StatusOK, gin.H{
		"preloaded_flags":   flagCount,
		"cached_responses":  cacheCount,
		"preload_complete":  preloadComplete,
	})
}

// buildPrecomputed creates a PrecomputedFlag from a DB flag, pre-unmarshaling all
// variation values so the hot evaluation path never calls json.Unmarshal.
func buildPrecomputed(flag *types.Flag) *PrecomputedFlag {
	var defaultValue interface{}
	_ = json.Unmarshal(flag.Default, &defaultValue)

	varValues := make(map[string]interface{}, len(flag.Variations))
	for _, v := range flag.Variations {
		var val interface{}
		_ = json.Unmarshal(v.Value, &val)
		varValues[v.ID] = val
		if v.Name != "" {
			varValues[v.Name] = val // also index by name for rule.Variation matching
		}
	}

	return &PrecomputedFlag{
		Key:             flag.Key,
		Enabled:         flag.Enabled,
		DefaultValue:    defaultValue,
		DefaultJSON:     flag.Default,
		Type:            flag.Type,
		HasTargeting:    flag.Targeting != nil,
		Variations:      flag.Variations,
		VariationValues: varValues,
		Targeting:       flag.Targeting,
		LastUpdated:     flag.UpdatedAt,
	}
}

// Helper functions
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return string(rune(int(val)))
	case int:
		return string(rune(val))
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

func contains(slice []string, item string) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == item {
			return true
		}
	}
	return false
}