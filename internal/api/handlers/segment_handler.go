package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/core/segment"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type SegmentHandler struct {
	segmentRepo *postgres.SegmentRepository
	evaluator   *segment.Evaluator
}

func NewSegmentHandler(segmentRepo *postgres.SegmentRepository) *SegmentHandler {
	return &SegmentHandler{
		segmentRepo: segmentRepo,
		evaluator:   segment.NewEvaluator(),
	}
}

// CreateSegment creates a new segment
func (h *SegmentHandler) CreateSegment(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	_, ok := userInterface.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	var req types.CreateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if key already exists in the project
	exists, err := h.segmentRepo.KeyExists(c.Request.Context(), req.ProjectID, req.Key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check key existence"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Segment key already exists in this project"})
		return
	}

	// Create segment
	segment := &types.Segment{
		ProjectID:   req.ProjectID,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Rules:       req.Rules,
	}

	if err := h.segmentRepo.Create(c.Request.Context(), segment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create segment"})
		return
	}

	c.JSON(http.StatusCreated, segment)
}

// ListSegments lists all segments for a project
func (h *SegmentHandler) ListSegments(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID parameter is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100 // Max limit
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	segments, err := h.segmentRepo.List(c.Request.Context(), projectID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list segments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"segments":   segments,
		"project_id": projectID,
		"limit":      limit,
		"offset":     offset,
	})
}

// GetSegment retrieves a segment by project ID and key
func (h *SegmentHandler) GetSegment(c *gin.Context) {
	key := c.Param("key")
	projectID := c.Query("project_id")
	
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Segment key is required"})
		return
	}
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID parameter is required"})
		return
	}

	segment, err := h.segmentRepo.GetByKey(c.Request.Context(), projectID, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get segment"})
		return
	}

	c.JSON(http.StatusOK, segment)
}

// UpdateSegment updates a segment
func (h *SegmentHandler) UpdateSegment(c *gin.Context) {
	key := c.Param("key")
	projectID := c.Query("project_id")
	
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Segment key is required"})
		return
	}
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID parameter is required"})
		return
	}

	var req types.UpdateSegmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing segment
	segment, err := h.segmentRepo.GetByKey(c.Request.Context(), projectID, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get segment"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		segment.Name = req.Name
	}
	if req.Description != "" {
		segment.Description = req.Description
	}
	if req.Rules != nil {
		segment.Rules = req.Rules
	}

	if err := h.segmentRepo.Update(c.Request.Context(), segment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update segment"})
		return
	}

	c.JSON(http.StatusOK, segment)
}

// DeleteSegment deletes a segment
func (h *SegmentHandler) DeleteSegment(c *gin.Context) {
	key := c.Param("key")
	projectID := c.Query("project_id")
	
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Segment key is required"})
		return
	}
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID parameter is required"})
		return
	}

	if err := h.segmentRepo.Delete(c.Request.Context(), projectID, key); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete segment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Segment deleted successfully"})
}

// EvaluateSegment evaluates if a user matches a segment
func (h *SegmentHandler) EvaluateSegment(c *gin.Context) {
	var req types.SegmentEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the segment - we need project_id from the request  
	segment, err := h.segmentRepo.GetByKey(c.Request.Context(), req.ProjectID, req.SegmentKey)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Segment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get segment"})
		return
	}

	// Evaluate the segment
	result, err := h.evaluator.EvaluateSegment(segment, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to evaluate segment"})
		return
	}

	response := types.SegmentEvaluationResponse{
		SegmentKey: result.SegmentKey,
		UserKey:    result.UserKey,
		Matched:    result.Matched,
		Reason:     result.Reason,
	}

	c.JSON(http.StatusOK, response)
}