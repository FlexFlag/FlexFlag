package types

import (
	"time"
)

// Environment represents a deployment environment
type Environment struct {
	ID          string                 `json:"id" db:"id"`
	ProjectID   string                 `json:"project_id" db:"project_id"`
	Name        string                 `json:"name" db:"name"`
	Key         string                 `json:"key" db:"key"` // e.g., "dev", "staging", "prod"
	Description string                 `json:"description" db:"description"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	SortOrder   int                    `json:"sort_order" db:"sort_order"`
	Settings    map[string]interface{} `json:"settings,omitempty" db:"settings"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// CreateEnvironmentRequest represents a request to create an environment
type CreateEnvironmentRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Key         string                 `json:"key" binding:"required"`
	Description string                 `json:"description"`
	SortOrder   int                    `json:"sort_order"`
	Settings    map[string]interface{} `json:"settings"`
}

// UpdateEnvironmentRequest represents a request to update an environment
type UpdateEnvironmentRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsActive    *bool                  `json:"is_active"`
	SortOrder   *int                   `json:"sort_order"`
	Settings    map[string]interface{} `json:"settings"`
}