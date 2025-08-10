package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectRepo *postgres.ProjectRepository
	flagRepo    storage.FlagRepository
	segmentRepo *postgres.SegmentRepository
	rolloutRepo *postgres.RolloutRepository
}

func NewProjectHandler(projectRepo *postgres.ProjectRepository, flagRepo storage.FlagRepository, segmentRepo *postgres.SegmentRepository, rolloutRepo *postgres.RolloutRepository) *ProjectHandler {
	return &ProjectHandler{
		projectRepo: projectRepo,
		flagRepo:    flagRepo,
		segmentRepo: segmentRepo,
		rolloutRepo: rolloutRepo,
	}
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	// Get user from context
	userInterface, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	claims, ok := userInterface.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user context"})
		return
	}

	var req types.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if slug already exists
	exists, err := h.projectRepo.SlugExists(c.Request.Context(), req.Slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check slug existence"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Project slug already exists"})
		return
	}

	// Create project
	project := &types.Project{
		Name:        req.Name,
		Description: req.Description,
		Slug:        req.Slug,
		IsActive:    true,
		Settings:    req.Settings,
		CreatedBy:   claims.UserID,
	}

	if err := h.projectRepo.Create(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects lists all projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
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

	projects, err := h.projectRepo.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list projects"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
		"limit":    limit,
		"offset":   offset,
	})
}

// GetProject retrieves a project by slug
func (h *ProjectHandler) GetProject(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	var req types.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get existing project
	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		project.Name = req.Name
	}
	if req.Description != "" {
		project.Description = req.Description
	}
	if req.IsActive != nil {
		project.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		project.Settings = req.Settings
	}

	if err := h.projectRepo.Update(c.Request.Context(), project); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// DeleteProject soft deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	// Get existing project first to get the ID
	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	if err := h.projectRepo.Delete(c.Request.Context(), project.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// CreateEnvironment creates a new environment for a project
func (h *ProjectHandler) CreateEnvironment(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	var req types.CreateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project to ensure it exists
	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	// Create environment
	env := &types.Environment{
		ProjectID:   project.ID,
		Name:        req.Name,
		Key:         req.Key,
		Description: req.Description,
		IsActive:    true,
		SortOrder:   req.SortOrder,
		Settings:    req.Settings,
	}

	if err := h.projectRepo.CreateEnvironment(c.Request.Context(), env); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create environment"})
		return
	}

	c.JSON(http.StatusCreated, env)
}

// GetEnvironments retrieves all environments for a project
func (h *ProjectHandler) GetEnvironments(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug is required"})
		return
	}

	// Get project to ensure it exists
	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}

	environments, err := h.projectRepo.GetEnvironmentsByProject(c.Request.Context(), project.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get environments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environments": environments})
}

// GetProjectStats gets statistics for a project
func (h *ProjectHandler) GetProjectStats(c *gin.Context) {
	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	// Count flags
	flagCount := 0
	if h.flagRepo != nil {
		// Count flags across all environments for this project
		environments := []string{"production", "staging", "development"}
		for _, env := range environments {
			flags, err := h.flagRepo.List(c.Request.Context(), env)
			if err == nil {
				for _, flag := range flags {
					if flag.ProjectID == projectID {
						flagCount++
					}
				}
			}
		}
	}

	// Count segments across all environments
	segmentCount := 0
	if h.segmentRepo != nil {
		environments := []string{"production", "staging", "development"}
		for _, env := range environments {
			segments, err := h.segmentRepo.List(c.Request.Context(), env, 1000, 0) // Large limit to get all
			if err == nil {
				segmentCount += len(segments)
			}
		}
	}

	// Count rollouts
	rolloutCount := 0
	if h.rolloutRepo != nil {
		// Count across all environments
		environments := []string{"production", "staging", "development"}
		for _, env := range environments {
			rollouts, err := h.rolloutRepo.GetByProject(c.Request.Context(), projectID, env)
			if err == nil {
				rolloutCount += len(rollouts)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"flags":    flagCount,
		"segments": segmentCount,
		"rollouts": rolloutCount,
	})
}