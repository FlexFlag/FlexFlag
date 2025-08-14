package handlers

import (
	"net/http"

	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

type ApiKeyHandler struct {
	repo *postgres.ApiKeyRepository
}

func NewApiKeyHandler(repo *postgres.ApiKeyRepository) *ApiKeyHandler {
	return &ApiKeyHandler{repo: repo}
}

// CreateApiKey generates a new API key for a project and environment
func (h *ApiKeyHandler) CreateApiKey(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	var req types.CreateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	apiKey, err := h.repo.GenerateApiKey(c.Request.Context(), &req, projectID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"api_key": apiKey})
}

// GetApiKeys returns all API keys for a project
func (h *ApiKeyHandler) GetApiKeys(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	apiKeys, err := h.repo.GetApiKeysByProject(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_keys": apiKeys})
}

// DeleteApiKey removes an API key
func (h *ApiKeyHandler) DeleteApiKey(c *gin.Context) {
	projectID := c.Param("projectId")
	keyID := c.Param("keyId")
	
	if projectID == "" || keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID and Key ID are required"})
		return
	}

	err := h.repo.DeleteApiKey(c.Request.Context(), keyID, projectID)
	if err != nil {
		if err.Error() == "API key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// UpdateApiKey modifies an existing API key
func (h *ApiKeyHandler) UpdateApiKey(c *gin.Context) {
	keyID := c.Param("keyId")
	if keyID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Key ID is required"})
		return
	}

	var req types.UpdateApiKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.repo.UpdateApiKey(c.Request.Context(), keyID, &req)
	if err != nil {
		if err.Error() == "API key not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
			return
		}
		if err.Error() == "no fields to update" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update API key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key updated successfully"})
}

// AuthenticateApiKey validates an API key (used internally by evaluation endpoints)
func (h *ApiKeyHandler) AuthenticateApiKey(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "API key is required"})
		return
	}

	keyInfo, err := h.repo.AuthenticateApiKey(c.Request.Context(), apiKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"api_key": keyInfo})
}