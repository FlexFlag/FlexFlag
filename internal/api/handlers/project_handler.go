package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

// CreateProject godoc
// @Summary Create a new project
// @Description Create a new project with specified configuration
// @Tags projects
// @Accept json
// @Produce json
// @Param project body types.CreateProjectRequest true "Project creation request"
// @Success 201 {object} map[string]interface{} "Created project"
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /projects [post]
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

	// Create default environments for the new project
	defaultEnvironments := []types.Environment{
		{
			ProjectID:   project.ID,
			Key:         "production",
			Name:        "Production",
			Description: "Production environment",
			IsActive:    true,
			SortOrder:   0,
		},
		{
			ProjectID:   project.ID,
			Key:         "staging",
			Name:        "Staging",
			Description: "Staging environment",
			IsActive:    true,
			SortOrder:   1,
		},
		{
			ProjectID:   project.ID,
			Key:         "development",
			Name:        "Development",
			Description: "Development environment",
			IsActive:    true,
			SortOrder:   2,
		},
	}

	for _, env := range defaultEnvironments {
		if err := h.projectRepo.CreateEnvironment(c.Request.Context(), &env); err != nil {
			// Log the error but don't fail the project creation
			fmt.Printf("Warning: Failed to create default environment %s for project %s: %v\n", env.Key, project.ID, err)
		} else {
			fmt.Printf("Successfully created default environment %s for project %s\n", env.Key, project.ID)
		}
	}

	c.JSON(http.StatusCreated, project)
}

// ListProjects godoc
// @Summary List projects
// @Description List all projects with pagination
// @Tags projects
// @Produce json
// @Param limit query int false "Number of projects to return" default(20)
// @Param offset query int false "Number of projects to skip" default(0)
// @Success 200 {object} map[string]interface{} "Response with projects array"
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /projects [get]
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

	// Copy existing flags from other environments to the new environment
	if h.flagRepo != nil {
		// Get flags from production environment as the source (or any existing environment)
		sourceEnvs := []string{"production", "staging", "development"}
		var sourceFlags []*types.Flag
		
		for _, sourceEnv := range sourceEnvs {
			flags, err := h.flagRepo.ListByProject(c.Request.Context(), project.ID, sourceEnv)
			if err == nil && len(flags) > 0 {
				sourceFlags = flags
				fmt.Printf("Found %d flags in %s environment to copy to new environment %s\n", len(flags), sourceEnv, env.Key)
				break
			}
		}
		
		// Copy each flag to the new environment
		for _, sourceFlag := range sourceFlags {
			newFlag := &types.Flag{
				ID:          uuid.New().String(),
				Key:         sourceFlag.Key,
				Name:        sourceFlag.Name,
				Description: sourceFlag.Description,
				Type:        sourceFlag.Type,
				Enabled:     sourceFlag.Enabled,
				Default:     sourceFlag.Default,
				Variations:  sourceFlag.Variations,
				Targeting:   sourceFlag.Targeting,
				Environment: env.Key, // Set to the new environment
				Tags:        sourceFlag.Tags,
				Metadata:    sourceFlag.Metadata,
				ProjectID:   sourceFlag.ProjectID,
			}
			
			if err := h.flagRepo.Create(c.Request.Context(), newFlag); err != nil {
				// Log error but don't fail environment creation
				fmt.Printf("Warning: Failed to copy flag %s to new environment %s: %v\n", sourceFlag.Key, env.Key, err)
			} else {
				fmt.Printf("Successfully copied flag %s to new environment %s\n", sourceFlag.Key, env.Key)
			}
		}
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
	fmt.Printf("Looking for project with slug: %s\n", slug)
	project, err := h.projectRepo.GetBySlug(c.Request.Context(), slug)
	if err != nil {
		fmt.Printf("Error getting project with slug %s: %v\n", slug, err)
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get project"})
		return
	}
	fmt.Printf("Found project: %s (ID: %s)\n", project.Name, project.ID)

	environments, err := h.projectRepo.GetEnvironmentsByProject(c.Request.Context(), project.ID)
	if err != nil {
		// Add debug logging
		fmt.Printf("Error getting environments for project %s: %v\n", project.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get environments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"environments": environments})
}

// UpdateEnvironment updates an environment
func (h *ProjectHandler) UpdateEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment ID is required"})
		return
	}

	var req types.UpdateEnvironmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.projectRepo.UpdateEnvironment(c.Request.Context(), id, &req); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update environment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment updated successfully"})
}

// DeleteEnvironment deletes an environment
func (h *ProjectHandler) DeleteEnvironment(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Environment ID is required"})
		return
	}

	// Get environment first to check if it's a default environment
	env, err := h.projectRepo.GetEnvironmentByID(c.Request.Context(), id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Environment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get environment"})
		return
	}

	// Prevent deletion of default environments
	defaultEnvs := []string{"production", "staging", "development"}
	for _, defaultEnv := range defaultEnvs {
		if env.Key == defaultEnv {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default environment"})
			return
		}
	}

	if err := h.projectRepo.DeleteEnvironment(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete environment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Environment deleted successfully"})
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