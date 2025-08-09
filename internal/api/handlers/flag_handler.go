package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FlagHandler struct {
	repo storage.FlagRepository
}

func NewFlagHandler(repo storage.FlagRepository) *FlagHandler {
	return &FlagHandler{repo: repo}
}

type CreateFlagRequest struct {
	Key         string                 `json:"key" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Type        types.FlagType         `json:"type" binding:"required"`
	Enabled     bool                   `json:"enabled"`
	Default     interface{}            `json:"default" binding:"required"`
	Environment string                 `json:"environment"`
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

	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Enabled:     req.Enabled,
		Default:     defaultValue,
		Variations:  []types.Variation{},
		Environment: req.Environment,
		Tags:        req.Tags,
		Metadata:    req.Metadata,
	}

	if err := h.repo.Create(c.Request.Context(), flag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, flag)
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

	flags, err := h.repo.List(c.Request.Context(), environment)
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

	c.JSON(http.StatusOK, existingFlag)
}

func (h *FlagHandler) DeleteFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")

	if err := h.repo.Delete(c.Request.Context(), key, environment); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *FlagHandler) ToggleFlag(c *gin.Context) {
	key := c.Param("key")
	environment := c.DefaultQuery("environment", "production")

	flag, err := h.repo.GetByKey(c.Request.Context(), key, environment)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "flag not found"})
		return
	}

	flag.Enabled = !flag.Enabled

	if err := h.repo.Update(c.Request.Context(), flag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":     flag.Key,
		"enabled": flag.Enabled,
	})
}