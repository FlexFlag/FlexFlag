package middleware

import (
	"net/http"

	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/gin-gonic/gin"
)

// ApiKeyAuth creates a middleware for API key authentication
func ApiKeyAuth(apiKeyRepo *postgres.ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		keyInfo, err := apiKeyRepo.AuthenticateApiKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Set API key information in context for use by handlers
		c.Set("apiKey", keyInfo)
		c.Set("projectID", keyInfo.ProjectID)
		c.Set("environmentID", keyInfo.EnvironmentID)
		c.Set("environment", keyInfo.Environment.Key)
		c.Set("permissions", keyInfo.Permissions)

		c.Next()
	}
}

// OptionalApiKeyAuth creates a middleware for optional API key authentication
// If API key is provided, it validates it; if not, request continues without auth
func OptionalApiKeyAuth(apiKeyRepo *postgres.ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// No API key provided, continue without authentication
			c.Next()
			return
		}

		keyInfo, err := apiKeyRepo.AuthenticateApiKey(c.Request.Context(), apiKey)
		if err != nil {
			// Invalid API key provided, return error
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Set API key information in context for use by handlers
		c.Set("apiKey", keyInfo)
		c.Set("projectID", keyInfo.ProjectID)
		c.Set("environmentID", keyInfo.EnvironmentID)
		c.Set("environment", keyInfo.Environment.Key)
		c.Set("permissions", keyInfo.Permissions)

		c.Next()
	}
}

// RequirePermission creates a middleware that checks if the API key has the required permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No permissions found"})
			c.Abort()
			return
		}

		perms, ok := permissions.([]string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid permissions format"})
			c.Abort()
			return
		}

		// Check if user has admin permission or the specific required permission
		hasPermission := false
		for _, perm := range perms {
			if perm == "admin" || perm == permission {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}