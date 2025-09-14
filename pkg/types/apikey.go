package types

import (
	"time"
)

// ApiKey represents an API key for programmatic access
type ApiKey struct {
	ID            string    `json:"id" db:"id"`
	ProjectID     string    `json:"project_id" db:"project_id"`
	EnvironmentID string    `json:"environment_id" db:"environment_id"`
	Name          string    `json:"name" db:"name"`
	KeyHash       string    `json:"-" db:"key_hash"` // Never expose the hash
	KeyPrefix     string    `json:"key_prefix" db:"key_prefix"`
	FullKey       string    `json:"full_key,omitempty" db:"-"` // Only shown once during creation
	Permissions   []string  `json:"permissions" db:"permissions"`
	CreatedBy     string    `json:"created_by" db:"created_by"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
	ExpiresAt     *time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt    *time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive      bool      `json:"is_active" db:"is_active"`
	
	// Virtual fields for API responses
	Environment   *Environment `json:"environment,omitempty" db:"-"`
	Project       *Project     `json:"project,omitempty" db:"-"`
}

// CreateApiKeyRequest represents a request to create an API key
type CreateApiKeyRequest struct {
	Name          string   `json:"name" binding:"required"`
	EnvironmentID string   `json:"environment_id" binding:"required"`
	Permissions   []string `json:"permissions" binding:"required"`
	ExpiresInDays *int     `json:"expires_in_days"` // nil means no expiration
}

// UpdateApiKeyRequest represents a request to update an API key
type UpdateApiKeyRequest struct {
	Name        string     `json:"name"`
	Permissions []string   `json:"permissions"`
	IsActive    *bool      `json:"is_active"`
	ExpiresAt   *time.Time `json:"expires_at"`
}