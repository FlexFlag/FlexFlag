package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flexflag/flexflag/internal/services"
	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FlagHandler struct {
	repo             storage.FlagRepository
	auditService     *services.AuditService
	ultraFastHandler *UltraFastHandler
}

func NewFlagHandler(repo storage.FlagRepository, auditService *services.AuditService, ultraFastHandler *UltraFastHandler) *FlagHandler {
	return &FlagHandler{
		repo:             repo,
		auditService:     auditService,
		ultraFastHandler: ultraFastHandler,
	}
}

type CreateFlagRequest struct {
	Key         string                 `json:"key" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Type        types.FlagType         `json:"type" binding:"required"`
	Enabled     bool                   `json:"enabled"`
	Default     interface{}            `json:"default" binding:"required"`
	Environment string                 `json:"environment"`
	ProjectID   string                 `json:"project_id"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

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
	
	// Define all environments where flags should be created
	environments := []string{"production", "staging", "development"}
	
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
			Variations:  []types.Variation{},
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