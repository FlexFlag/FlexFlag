package handlers

import (
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/core/rollout"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type RolloutHandler struct {
	rolloutRepo storage.RolloutRepository
	evaluator   *rollout.Evaluator
}

func NewRolloutHandler(rolloutRepo storage.RolloutRepository) *RolloutHandler {
	return &RolloutHandler{
		rolloutRepo: rolloutRepo,
		evaluator:   rollout.NewEvaluator(),
	}
}

// CreateRollout creates a new rollout
func (h *RolloutHandler) CreateRollout(c *gin.Context) {
	var req types.CreateRolloutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rollout := &types.Rollout{
		FlagID:      req.FlagID,
		Environment: req.Environment,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Config:      req.Config,
		Status:      types.RolloutStatusDraft,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	if err := h.rolloutRepo.Create(c.Request.Context(), rollout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rollout)
}

// GetRollout retrieves a rollout by ID
func (h *RolloutHandler) GetRollout(c *gin.Context) {
	id := c.Param("id")

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rollout)
}

// GetRolloutsByFlag retrieves all rollouts for a flag
func (h *RolloutHandler) GetRolloutsByFlag(c *gin.Context) {
	flagID := c.Query("flag_id")
	environment := c.Query("environment")

	if flagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_id is required"})
		return
	}

	if environment == "" {
		environment = "production"
	}

	rollouts, err := h.rolloutRepo.GetByFlag(c.Request.Context(), flagID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rollouts": rollouts})
}

// GetAllRollouts retrieves all rollouts for a project and environment
func (h *RolloutHandler) GetAllRollouts(c *gin.Context) {
	projectID := c.Query("project_id")
	environment := c.Query("environment")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	if environment == "" {
		environment = "production"
	}

	rollouts, err := h.rolloutRepo.GetByProject(c.Request.Context(), projectID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rollouts": rollouts})
}

// UpdateRollout updates an existing rollout
func (h *RolloutHandler) UpdateRollout(c *gin.Context) {
	id := c.Param("id")

	var req types.UpdateRolloutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	rollout.Name = req.Name
	rollout.Description = req.Description
	rollout.Config = req.Config
	rollout.Status = req.Status
	rollout.StartDate = req.StartDate
	rollout.EndDate = req.EndDate

	if err := h.rolloutRepo.Update(c.Request.Context(), rollout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rollout)
}

// DeleteRollout deletes a rollout
func (h *RolloutHandler) DeleteRollout(c *gin.Context) {
	id := c.Param("id")

	if err := h.rolloutRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rollout deleted successfully"})
}

// EvaluateRollout evaluates a rollout for a specific user
func (h *RolloutHandler) EvaluateRollout(c *gin.Context) {
	id := c.Param("id")
	userKey := c.Query("user_key")

	if userKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_key is required"})
		return
	}

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check for existing sticky assignment
	var stickyAssignment *types.StickyAssignment
	if rollout.Config.StickyBucketing {
		stickyAssignment, _ = h.rolloutRepo.GetStickyAssignment(c.Request.Context(), rollout.FlagID, rollout.Environment, userKey)
	}

	result, err := h.evaluator.EvaluateRollout(rollout, userKey, stickyAssignment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create sticky assignment if needed and user matched
	if rollout.Config.StickyBucketing && result.Matched && !result.IsSticky {
		assignment := &types.StickyAssignment{
			FlagID:      rollout.FlagID,
			Environment: rollout.Environment,
			UserKey:     userKey,
			VariationID: result.VariationID,
			BucketKey:   h.evaluator.GenerateBucketKey(rollout.FlagID, rollout.Environment, userKey, rollout.Config.BucketBy),
			ExpiresAt:   rollout.EndDate,
		}

		h.rolloutRepo.CreateStickyAssignment(c.Request.Context(), assignment)
	}

	c.JSON(http.StatusOK, result)
}

// GetStickyAssignments retrieves sticky assignments for a flag
func (h *RolloutHandler) GetStickyAssignments(c *gin.Context) {
	flagID := c.Query("flag_id")
	environment := c.Query("environment")
	userKey := c.Query("user_key")

	if flagID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_id is required"})
		return
	}

	if environment == "" {
		environment = "production"
	}

	if userKey != "" {
		// Get specific assignment
		assignment, err := h.rolloutRepo.GetStickyAssignment(c.Request.Context(), flagID, environment, userKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if assignment == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "assignment not found"})
			return
		}

		c.JSON(http.StatusOK, assignment)
		return
	}

	// For now, just return empty array - in production you'd implement pagination
	c.JSON(http.StatusOK, gin.H{"assignments": []interface{}{}})
}

// DeleteStickyAssignment removes a sticky assignment
func (h *RolloutHandler) DeleteStickyAssignment(c *gin.Context) {
	flagID := c.Query("flag_id")
	environment := c.Query("environment")
	userKey := c.Query("user_key")

	if flagID == "" || userKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flag_id and user_key are required"})
		return
	}

	if environment == "" {
		environment = "production"
	}

	if err := h.rolloutRepo.DeleteStickyAssignment(c.Request.Context(), flagID, environment, userKey); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sticky assignment deleted successfully"})
}

// ActivateRollout activates a rollout
func (h *RolloutHandler) ActivateRollout(c *gin.Context) {
	id := c.Param("id")

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	rollout.Status = types.RolloutStatusActive
	if rollout.StartDate == nil {
		now := time.Now()
		rollout.StartDate = &now
	}

	if err := h.rolloutRepo.Update(c.Request.Context(), rollout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rollout)
}

// PauseRollout pauses a rollout
func (h *RolloutHandler) PauseRollout(c *gin.Context) {
	id := c.Param("id")

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	rollout.Status = types.RolloutStatusPaused

	if err := h.rolloutRepo.Update(c.Request.Context(), rollout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rollout)
}

// CompleteRollout completes a rollout
func (h *RolloutHandler) CompleteRollout(c *gin.Context) {
	id := c.Param("id")

	rollout, err := h.rolloutRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	rollout.Status = types.RolloutStatusCompleted
	if rollout.EndDate == nil {
		now := time.Now()
		rollout.EndDate = &now
	}

	if err := h.rolloutRepo.Update(c.Request.Context(), rollout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rollout)
}

// CleanupExpiredAssignments removes expired sticky assignments
func (h *RolloutHandler) CleanupExpiredAssignments(c *gin.Context) {
	if err := h.rolloutRepo.CleanupExpiredAssignments(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Expired assignments cleaned up successfully"})
}