package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/core/evaluation"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type EvaluationHandler struct {
	repo   storage.FlagRepository
	engine *evaluation.Engine
}

func NewEvaluationHandler(repo storage.FlagRepository) *EvaluationHandler {
	engine := evaluation.NewEngine()
	return &EvaluationHandler{
		repo:   repo,
		engine: engine,
	}
}

type EvaluateRequest struct {
	FlagKey    string                 `json:"flag_key" binding:"required"`
	UserID     string                 `json:"user_id"`
	UserKey    string                 `json:"user_key"`
	Attributes map[string]interface{} `json:"attributes"`
}

type EvaluateResponse struct {
	FlagKey       string          `json:"flag_key"`
	Value         interface{}     `json:"value"`
	Variation     string          `json:"variation,omitempty"`
	Reason        string          `json:"reason"`
	RuleID        string          `json:"rule_id,omitempty"`
	Default       bool            `json:"default"`
	EvaluationTime float64        `json:"evaluation_time_ms"`
	Timestamp     time.Time       `json:"timestamp"`
}

func (h *EvaluationHandler) Evaluate(c *gin.Context) {
	startTime := time.Now()
	
	var req EvaluateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	environment := c.DefaultQuery("environment", "production")

	// Fetch flag from database
	flag, err := h.repo.GetByKey(c.Request.Context(), req.FlagKey, environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	// If flag is disabled, return default value immediately
	if !flag.Enabled {
		var value interface{}
		json.Unmarshal(flag.Default, &value)
		
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

	// Update engine with latest flag data
	h.engine.UpdateFlag(flag)

	// Build evaluation request
	evalReq := &types.EvaluationRequest{
		FlagKey:     req.FlagKey,
		UserID:      req.UserID,
		UserKey:     req.UserKey,
		Attributes:  req.Attributes,
		Environment: environment,
	}

	// Perform evaluation
	evalResp, err := h.engine.EvaluateFlag(c.Request.Context(), evalReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate evaluation time
	evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0

	// Parse the value from JSON
	var value interface{}
	json.Unmarshal(evalResp.Value, &value)

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

func (h *EvaluationHandler) BatchEvaluate(c *gin.Context) {
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
	results := make(map[string]interface{})
	
	for _, flagKey := range req.FlagKeys {
		// Fetch flag from database
		flag, err := h.repo.GetByKey(c.Request.Context(), flagKey, environment)
		if err != nil {
			results[flagKey] = map[string]interface{}{
				"error": "flag not found",
			}
			continue
		}

		// If flag is disabled, return default value
		if !flag.Enabled {
			var value interface{}
			json.Unmarshal(flag.Default, &value)
			results[flagKey] = map[string]interface{}{
				"value":   value,
				"reason":  "flag_disabled",
				"default": true,
			}
			continue
		}

		// Update engine with latest flag data
		h.engine.UpdateFlag(flag)

		// Build evaluation request
		evalReq := &types.EvaluationRequest{
			FlagKey:     flagKey,
			UserID:      req.UserID,
			UserKey:     req.UserKey,
			Attributes:  req.Attributes,
			Environment: environment,
		}

		// Perform evaluation
		evalResp, err := h.engine.EvaluateFlag(c.Request.Context(), evalReq)
		if err != nil {
			results[flagKey] = map[string]interface{}{
				"error": err.Error(),
			}
			continue
		}

		// Parse the value from JSON
		var value interface{}
		json.Unmarshal(evalResp.Value, &value)

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
		"evaluations":         results,
		"total_time_ms":       totalTime,
		"avg_time_per_flag_ms": avgTime,
		"timestamp":           time.Now(),
	})
}