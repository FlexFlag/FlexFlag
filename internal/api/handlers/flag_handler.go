package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flexflag/flexflag/internal/services"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FlagHandler struct {
	repo             storage.FlagRepository
	auditService     *services.AuditService
	ultraFastHandler *UltraFastHandler
	projectRepo      *postgres.ProjectRepository
	edgeSyncHandler  *EdgeSyncHandler
	sseHandler       *SSEHandler
}

func NewFlagHandler(repo storage.FlagRepository, auditService *services.AuditService, ultraFastHandler *UltraFastHandler, projectRepo *postgres.ProjectRepository) *FlagHandler {
	return &FlagHandler{
		repo:             repo,
		auditService:     auditService,
		ultraFastHandler: ultraFastHandler,
		projectRepo:      projectRepo,
	}
}

// SetEdgeSyncHandler sets the edge sync handler for broadcasting updates
func (h *FlagHandler) SetEdgeSyncHandler(edgeSyncHandler *EdgeSyncHandler) {
	h.edgeSyncHandler = edgeSyncHandler
}

// SetSSEHandler sets the SSE handler for broadcasting updates
func (h *FlagHandler) SetSSEHandler(sseHandler *SSEHandler) {
	h.sseHandler = sseHandler
}

// CreateFlagRequest represents the request to create a new flag
type CreateFlagRequest struct {
	Key         string                 `json:"key" binding:"required" example:"feature-toggle"`
	Name        string                 `json:"name" binding:"required" example:"Feature Toggle"`
	Description string                 `json:"description" example:"Description of the feature"`
	Type        types.FlagType         `json:"type" binding:"required" example:"boolean"`
	Enabled     bool                   `json:"enabled" example:"true"`
	Default     interface{}            `json:"default" binding:"required" swaggertype:"object"`
	Environment string                 `json:"environment" example:"production"`
	ProjectID   string                 `json:"project_id" example:"proj_123"`
	Tags        []string               `json:"tags" example:"feature,toggle"`
	Metadata    map[string]interface{} `json:"metadata" swaggertype:"object"`
	Variations  []SwaggerVariation     `json:"variations,omitempty"`
	Targeting   *types.TargetingConfig `json:"targeting,omitempty"`
}

// SwaggerVariation represents a variation for Swagger documentation
type SwaggerVariation struct {
	ID          string      `json:"id" example:"var_1"`
	Name        string      `json:"name" example:"Variation A"`
	Description string      `json:"description" example:"First variation"`
	Value       interface{} `json:"value" swaggertype:"object"`
	Weight      int         `json:"weight" example:"50"`
}

// CreateFlag godoc
// @Summary Create a new flag
// @Description Create a new feature flag with specified configuration
// @Tags flags
// @Accept json
// @Produce json
// @Param flag body CreateFlagRequest true "Flag creation request"
// @Success 201 {object} map[string]interface{} "Created flag"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /flags [post]
func (h *FlagHandler) CreateFlag(c *gin.Context) {
	var req CreateFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Environment == "" {
		req.Environment = "production"
	}

	defaultValue, err := json.Marshal(req.Default)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid default value"})
		return
	}

	// Use provided project_id or fall back to default
	projectID := req.ProjectID
	if projectID == "" {
		projectID = "5aa79fcc-7e77-46fd-be58-f151574e57a9" // Default project fallback
	}
	
	// Convert SwaggerVariation to types.Variation
	var variations []types.Variation
	for _, sv := range req.Variations {
		valueBytes, _ := json.Marshal(sv.Value)
		variations = append(variations, types.Variation{
			ID:          sv.ID,
			Name:        sv.Name,
			Description: sv.Description,
			Value:       valueBytes,
			Weight:      sv.Weight,
		})
	}

	// Get all environments for the project to create flags in each one
	var environments []string
	if h.projectRepo != nil {
		envs, err := h.projectRepo.GetEnvironmentsByProject(c.Request.Context(), projectID)
		if err != nil {
			// Fallback to default environments if we can't fetch project environments
			fmt.Printf("Warning: Failed to get environments for project %s: %v\n", projectID, err)
			environments = []string{"production", "staging", "development"}
		} else {
			// Extract environment keys from the environments
			for _, env := range envs {
				environments = append(environments, env.Key)
			}
		}
	} else {
		// Fallback to default environments if no project repo
		environments = []string{"production", "staging", "development"}
	}
	
	// Create flag in all environments
	var createdFlags []*types.Flag
	for _, env := range environments {
		flag := &types.Flag{
			ID:          uuid.New().String(),
			Key:         req.Key,
			Name:        req.Name,
			Description: req.Description,
			Type:        req.Type,
			Enabled:     req.Enabled,
			Default:     defaultValue,
			Variations:  variations,
			Targeting:   req.Targeting,
			Environment: env,
			Tags:        req.Tags,
			Metadata:    req.Metadata,
			ProjectID:   projectID,
		}

		if err := h.repo.Create(c.Request.Context(), flag); err != nil {
			// Check for duplicate key constraint violation
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("A flag with this key already exists in the %s environment", env)})
				return
			}
			
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Refresh ultra-fast cache for new flag
		if h.ultraFastHandler != nil {
			h.ultraFastHandler.RefreshFlag(flag.Key, env)
		}
		
		// Broadcast to edge servers via WebSocket
		if h.edgeSyncHandler != nil {
			h.edgeSyncHandler.BroadcastFlagUpdate(flag, "create")
		}
		
		// Broadcast to edge servers via SSE
		if h.sseHandler != nil {
			h.sseHandler.BroadcastFlagUpdate(flag, "create")
		}

		createdFlags = append(createdFlags, flag)

		// Log the create action for each environment
		if h.auditService != nil {
			if err := h.auditService.LogFlagAction(c.Request.Context(), c, "create", flag, nil); err != nil {
				// Log error but don't fail the request
				c.Header("X-Audit-Error", err.Error())
			} else {
				c.Header("X-Audit-Success", "true")
			}
		}
	}

	// Return the flag from the requested environment (or production as default)
	requestedEnv := req.Environment
	if requestedEnv == "" {
		requestedEnv = "production"
	}
	
	// Find and return the flag for the requested environment
	var responseFlag *types.Flag
	for _, flag := range createdFlags {
		if flag.Environment == requestedEnv {
			responseFlag = flag
			break
		}
	}
	
	if responseFlag == nil {
		responseFlag = createdFlags[0] // fallback to first created flag
	}

	c.JSON(http.StatusCreated, responseFlag)
}

// GetFlag godoc
// @Summary Get a flag by key
// @Description Get a feature flag by its key and environment
// @Tags flags
// @Produce json
// @Param key path string true "Flag key"
// @Param environment query string false "Environment" default(production)
// @Success 200 {object} map[string]interface{} "Flag details"
// @Failure 404 {object} map[string]string
// @Security ApiKeyAuth
// @Router /flags/{key} [get]
func (h *FlagHandler) GetFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")

	flag, err := h.repo.GetByKey(c.Request.Context(), key, environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	c.JSON(http.StatusOK, flag)
}

// ListFlags godoc
// @Summary List flags
// @Description List all feature flags for an environment, optionally filtered by project
// @Tags flags
// @Produce json
// @Param environment query string false "Environment" default(production)
// @Param project_id query string false "Project ID to filter by"
// @Success 200 {object} map[string]interface{} "Response with flags array"
// @Failure 500 {object} map[string]string
// @Security ApiKeyAuth
// @Router /flags [get]
func (h *FlagHandler) ListFlags(c *gin.Context) {
	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")

	var flags []*types.Flag
	var err error

	if projectID != "" {
		// Filter by project if project_id is provided
		flags, err = h.repo.ListByProject(c.Request.Context(), projectID, environment)
	} else {
		// Default behavior - list all flags for the environment
		flags, err = h.repo.List(c.Request.Context(), environment)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"flags": flags})
}

func (h *FlagHandler) UpdateFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")

	existingFlag, err := h.repo.GetByKey(c.Request.Context(), key, environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	var req CreateFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	defaultValue, err := json.Marshal(req.Default)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid default value"})
		return
	}

	existingFlag.Name = req.Name
	existingFlag.Description = req.Description
	existingFlag.Type = req.Type
	existingFlag.Enabled = req.Enabled
	existingFlag.Default = defaultValue
	existingFlag.Tags = req.Tags
	existingFlag.Metadata = req.Metadata

	if err := h.repo.Update(c.Request.Context(), existingFlag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh ultra-fast cache
	if h.ultraFastHandler != nil {
		h.ultraFastHandler.RefreshFlag(existingFlag.Key, environment)
	}
	
	// Broadcast to edge servers via WebSocket
	if h.edgeSyncHandler != nil {
		h.edgeSyncHandler.BroadcastFlagUpdate(existingFlag, "update")
	}
	
	// Broadcast to edge servers via SSE
	if h.sseHandler != nil {
		h.sseHandler.BroadcastFlagUpdate(existingFlag, "update")
	}

	c.JSON(http.StatusOK, existingFlag)
}

func (h *FlagHandler) DeleteFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")

	if err := h.repo.Delete(c.Request.Context(), key, environment); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	// Refresh ultra-fast cache (removes deleted flag)
	if h.ultraFastHandler != nil {
		h.ultraFastHandler.RefreshFlag(key, environment)
	}
	
	// Broadcast deletion to edge servers via WebSocket
	if h.edgeSyncHandler != nil {
		// Create a minimal flag object for the delete broadcast
		flag := &types.Flag{
			Key:         key,
			Environment: environment,
		}
		h.edgeSyncHandler.BroadcastFlagUpdate(flag, "delete")
	}
	
	// Broadcast deletion to edge servers via SSE
	if h.sseHandler != nil {
		// Create a minimal flag object for the delete broadcast
		flag := &types.Flag{
			Key:         key,
			Environment: environment,
		}
		h.sseHandler.BroadcastFlagUpdate(flag, "delete")
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *FlagHandler) ToggleFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")

	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project_id is required"})
		return
	}

	flag, err := h.repo.GetByProjectKey(c.Request.Context(), projectID, key, environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	// Store the old flag state for audit log
	oldFlag := *flag
	flag.Enabled = !flag.Enabled

	if err := h.repo.Update(c.Request.Context(), flag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Refresh ultra-fast cache
	if h.ultraFastHandler != nil {
		h.ultraFastHandler.RefreshFlag(flag.Key, environment)
	}
	
	// Broadcast to edge servers via WebSocket
	if h.edgeSyncHandler != nil {
		h.edgeSyncHandler.BroadcastFlagUpdate(flag, "update")
	}
	
	// Broadcast to edge servers via SSE
	if h.sseHandler != nil {
		h.sseHandler.BroadcastFlagUpdate(flag, "update")
	}

	// Log the toggle action
	if h.auditService != nil {
		if err := h.auditService.LogFlagAction(c.Request.Context(), c, "toggle", flag, &oldFlag); err != nil {
			// Log error but don't fail the request
			c.Header("X-Audit-Error", err.Error())
		} else {
			c.Header("X-Audit-Success", "true")
		}
	} else {
		c.Header("X-Audit-Service", "nil")
	}

	c.JSON(http.StatusOK, gin.H{
		"key":     flag.Key,
		"enabled": flag.Enabled,
	})
}