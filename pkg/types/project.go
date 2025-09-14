package types

import (
	"time"
)

// Project represents a project containing flags
type Project struct {
	ID          string                 `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Slug        string                 `json:"slug" db:"key"` // URL-friendly identifier
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Settings    map[string]interface{} `json:"settings,omitempty" db:"settings"`
	CreatedBy   string                 `json:"created_by" db:"created_by"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// ProjectMember represents user membership in a project
type ProjectMember struct {
	ID        string    `json:"id" db:"id"`
	ProjectID string    `json:"project_id" db:"project_id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Role      UserRole  `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Project role constants
const (
	ProjectRoleAdmin  = "admin"
	ProjectRoleEditor = "editor"
	ProjectRoleViewer = "viewer"
)

// CreateProjectRequest represents a request to create a project
type CreateProjectRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Slug        string                 `json:"slug" binding:"required"`
	Settings    map[string]interface{} `json:"settings"`
}

// UpdateProjectRequest represents a request to update a project
type UpdateProjectRequest struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsActive    *bool                  `json:"is_active"`
	Settings    map[string]interface{} `json:"settings"`
}

// AddProjectMemberRequest represents a request to add a member to a project
type AddProjectMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}