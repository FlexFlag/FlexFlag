package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/cache"
	"github.com/flexflag/flexflag/internal/core/evaluation"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type OptimizedEvaluationHandler struct {
	repo   storage.FlagRepository
	engine *evaluation.Engine
	cache  *cache.MemoryCache
}

func NewOptimizedEvaluationHandler(repo storage.FlagRepository) *OptimizedEvaluationHandler {
	engine := evaluation.NewEngine()
	cache := cache.NewMemoryCache(5 * time.Minute) // 5 minute TTL
	
	return &OptimizedEvaluationHandler{
		repo:   repo,
		engine: engine,
		cache:  cache,
	}
}

func (h *OptimizedEvaluationHandler) FastEvaluate(c *gin.Context) {
	startTime := time.Now()
	
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")
	
	// If API key authentication is used, override environment and project from key
	if apiKeyEnv, exists := c.Get("environment"); exists {
		environment = apiKeyEnv.(string)
	}
	if apiKeyProjectID, exists := c.Get("projectID"); exists {
		projectID = apiKeyProjectID.(string)
	}

	// Try cache first
	flag, found := h.cache.Get(c.Request.Context(), req.FlagKey, environment)
	if !found {
		// Cache miss - fetch from database
		var err error
		if projectID != "" {
			flag, err = h.repo.GetByProjectKey(c.Request.Context(), projectID, req.FlagKey, environment)
		} else {
			flag, err = h.repo.GetByKey(c.Request.Context(), req.FlagKey, environment)
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
			return
		}
		
		// Store in cache for next time
		h.cache.Set(c.Request.Context(), req.FlagKey, environment, flag)
	}

	// Fast path for disabled flags
	if !flag.Enabled {
		var value interface{}
		_ = json.Unmarshal(flag.Default, &value)
		
		evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
		c.JSON(http.StatusOK, EvaluateResponse{
			FlagKey:        flag.Key,
			Value:          value,
			Reason:         "flag_disabled",
			Default:        true,
			EvaluationTime: evalTime,
			Timestamp:      time.Now(),
		})
		return
	}

	// Update engine with cached flag (very fast since it's in-memory)
	h.engine.UpdateFlag(flag)

	// Build evaluation request
	evalReq := &types.EvaluationRequest{
		FlagKey:     req.FlagKey,
		UserID:      req.UserID,
		UserKey:     req.UserKey,
		Attributes:  req.Attributes,
		Environment: environment,
	}

	// Perform in-memory evaluation
	evalResp, err := h.engine.EvaluateFlag(c.Request.Context(), evalReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate evaluation time
	evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	// Parse the value from JSON
	var value interface{}
	_ = json.Unmarshal(evalResp.Value, &value)

	response := EvaluateResponse{
		FlagKey:        evalResp.FlagKey,
		Value:          value,
		Variation:      evalResp.Variation,
		Reason:         evalResp.Reason,
		RuleID:         evalResp.RuleID,
		Default:        evalResp.Default,
		EvaluationTime: evalTime,
		Timestamp:      evalResp.Timestamp,
	}

	c.JSON(http.StatusOK, response)
}

func (h *OptimizedEvaluationHandler) FastBatchEvaluate(c *gin.Context) {
	startTime := time.Now()
	
	var req struct {
		FlagKeys   []string               `json:"flag_keys" binding:"required"`
		UserID     string                 `json:"user_id"`
		UserKey    string                 `json:"user_key"`
		Attributes map[string]interface{} `json:"attributes"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")
	
	// If API key authentication is used, override environment and project from key
	if apiKeyEnv, exists := c.Get("environment"); exists {
		environment = apiKeyEnv.(string)
	}
	if apiKeyProjectID, exists := c.Get("projectID"); exists {
		projectID = apiKeyProjectID.(string)
	}
	
	results := make(map[string]interface{})
	
	// Batch process all flags
	for _, flagKey := range req.FlagKeys {
		// Try cache first
		flag, found := h.cache.Get(c.Request.Context(), flagKey, environment)
		if !found {
			// Cache miss - fetch from database
			var err error
			if projectID != "" {
				flag, err = h.repo.GetByProjectKey(c.Request.Context(), projectID, flagKey, environment)
			} else {
				flag, err = h.repo.GetByKey(c.Request.Context(), flagKey, environment)
			}
			if err != nil {
				results[flagKey] = map[string]interface{}{
					"error": "flag not found",
				}
				continue
			}
			
			// Store in cache
			h.cache.Set(c.Request.Context(), flagKey, environment, flag)
		}

		// Fast path for disabled flags
		if !flag.Enabled {
			var value interface{}
			_ = json.Unmarshal(flag.Default, &value)
			results[flagKey] = map[string]interface{}{
				"value":   value,
				"reason":  "flag_disabled",
				"default": true,
			}
			continue
		}

		// Update engine and evaluate
		h.engine.UpdateFlag(flag)
		evalReq := &types.EvaluationRequest{
			FlagKey:     flagKey,
			UserID:      req.UserID,
			UserKey:     req.UserKey,
			Attributes:  req.Attributes,
			Environment: environment,
		}

		evalResp, err := h.engine.EvaluateFlag(c.Request.Context(), evalReq)
		if err != nil {
			results[flagKey] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}

		var value interface{}
		_ = json.Unmarshal(evalResp.Value, &value)

		results[flagKey] = map[string]interface{}{
			"value":     value,
			"variation": evalResp.Variation,
			"reason":    evalResp.Reason,
			"rule_id":   evalResp.RuleID,
			"default":   evalResp.Default,
		}
	}

	// Calculate total evaluation time
	totalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
	avgTime := totalTime / float64(len(req.FlagKeys))

	c.JSON(http.StatusOK, gin.H{
		"evaluations":           results,
		"total_time_ms":         totalTime,
		"avg_time_per_flag_ms":  avgTime,
		"cache_enabled":         true,
		"timestamp":             time.Now(),
	})
}

func (h *OptimizedEvaluationHandler) GetCacheStats(c *gin.Context) {
	stats := h.cache.Stats()
	c.JSON(http.StatusOK, gin.H{
		"cache_stats": stats,
	})
}

func (h *OptimizedEvaluationHandler) ClearCache(c *gin.Context) {
	h.cache.Clear(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{
		"message": "cache cleared",
	})
}