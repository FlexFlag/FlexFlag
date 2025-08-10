package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/core/evaluation"
	"github.com/flexflag/flexflag/internal/core/rollout"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type EvaluationHandler struct {
	repo         storage.FlagRepository
	rolloutRepo  *postgres.RolloutRepository
	engine       *evaluation.Engine
	rolloutEval  *rollout.Evaluator
}

func NewEvaluationHandler(repo storage.FlagRepository, rolloutRepo *postgres.RolloutRepository) *EvaluationHandler {
	engine := evaluation.NewEngine()
	rolloutEval := rollout.NewEvaluator()
	return &EvaluationHandler{
		repo:        repo,
		rolloutRepo: rolloutRepo,
		engine:      engine,
		rolloutEval: rolloutEval,
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
	projectID := c.Query("project_id")

	// Fetch flag from database - need to use project-specific method
	var flag *types.Flag
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

	// Check for active rollouts first
	rollouts, err := h.rolloutRepo.GetActiveRollouts(c.Request.Context(), flag.ID, environment)
	fmt.Printf("DEBUG: Flag ID: %s, Environment: %s, Rollouts found: %d, Error: %v\n", flag.ID, environment, len(rollouts), err)
	
	// Debug: Also check ALL rollouts for this flag (not just active)
	allRollouts, allErr := h.rolloutRepo.GetByFlag(c.Request.Context(), flag.ID, environment)
	fmt.Printf("DEBUG: ALL rollouts for flag (any status): %d, Error: %v\n", len(allRollouts), allErr)
	if len(allRollouts) > 0 {
		for i, rollout := range allRollouts {
			fmt.Printf("DEBUG: Rollout %d: ID=%s, Status=%s, Type=%s, FlagID=%s\n", i+1, rollout.ID, rollout.Status, rollout.Type, rollout.FlagID)
		}
	}
	if err == nil && len(rollouts) > 0 {
		// Evaluate rollouts
		userKey := req.UserKey
		if userKey == "" {
			userKey = req.UserID
		}

		for _, rollout := range rollouts {
			// Get sticky assignment if enabled
			var stickyAssignment *types.StickyAssignment
			if rollout.Config.StickyBucketing {
				stickyAssignment, _ = h.rolloutRepo.GetStickyAssignment(c.Request.Context(), rollout.FlagID, rollout.Environment, userKey)
			}

			// Evaluate rollout
			result, rolloutErr := h.rolloutEval.EvaluateRollout(rollout, userKey, stickyAssignment)
			if rolloutErr == nil && result.Matched {
				var value interface{}
				json.Unmarshal(flag.Default, &value)

				// For percentage rollouts, return enabled/disabled based on match
				if rollout.Type == "percentage" && result.Matched {
					// Return the flag value for matched users
					evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
					c.JSON(http.StatusOK, EvaluateResponse{
						FlagKey:        flag.Key,
						Value:          value,
						Reason:         "rollout_match",
						Default:        false,
						EvaluationTime: evalTime,
						Timestamp:      time.Now(),
					})
					return
				}

				// For experiment rollouts, return variation value if available
				if rollout.Type == "experiment" && result.VariationID != "" {
					// Find the variation in the flag
					for _, variation := range flag.Variations {
						if variation.ID == result.VariationID {
							evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
							c.JSON(http.StatusOK, EvaluateResponse{
								FlagKey:        flag.Key,
								Value:          variation.Value,
								Variation:      variation.ID,
								Reason:         "rollout_experiment",
								Default:        false,
								EvaluationTime: evalTime,
								Timestamp:      time.Now(),
							})
							return
						}
					}
				}

				// Create sticky assignment if needed and user matched
				if rollout.Config.StickyBucketing && result.Matched && !result.IsSticky {
					assignment := &types.StickyAssignment{
						FlagID:      rollout.FlagID,
						Environment: rollout.Environment,
						UserKey:     userKey,
						VariationID: result.VariationID,
						BucketKey:   h.rolloutEval.GenerateBucketKey(rollout.FlagID, rollout.Environment, userKey, rollout.Config.BucketBy),
						ExpiresAt:   rollout.EndDate,
					}
					h.rolloutRepo.CreateStickyAssignment(c.Request.Context(), assignment)
				}
			}
		}
	}

	// If no rollout matched, return disabled/default value for percentage rollouts
	if len(rollouts) > 0 {
		for _, rollout := range rollouts {
			if rollout.Type == "percentage" {
				// For percentage rollouts, if user doesn't match, return false/disabled
				var defaultVal interface{} = false
				if flag.Type == "string" {
					defaultVal = ""
				} else if flag.Type == "number" {
					defaultVal = 0
				}

				evalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
				c.JSON(http.StatusOK, EvaluateResponse{
					FlagKey:        flag.Key,
					Value:          defaultVal,
					Reason:         "rollout_no_match",
					Default:        true,
					EvaluationTime: evalTime,
					Timestamp:      time.Now(),
				})
				return
			}
		}
	}

	// Fall back to normal flag evaluation if no rollouts
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