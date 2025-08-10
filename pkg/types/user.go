package types

import (
	"encoding/json"
	"time"
)

// User represents a system user
type User struct {
	ID           string    `json:"id" db:"id"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	FullName     string    `json:"full_name" db:"full_name"`
	Role         UserRole  `json:"role" db:"role"`
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// UserRole defines user permission levels
type UserRole string

const (
	UserRoleAdmin  UserRole = "admin"
	UserRoleEditor UserRole = "editor" 
	UserRoleViewer UserRole = "viewer"
)

// Project role constants
const (
	ProjectRoleAdmin  = "admin"
	ProjectRoleEditor = "editor"
	ProjectRoleViewer = "viewer"
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

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID           string                 `json:"id" db:"id"`
	ProjectID    *string                `json:"project_id,omitempty" db:"project_id"`
	UserID       *string                `json:"user_id,omitempty" db:"user_id"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   string                 `json:"resource_id" db:"resource_id"`
	Action       string                 `json:"action" db:"action"`
	OldValues    json.RawMessage        `json:"old_values,omitempty" db:"old_values"`
	NewValues    json.RawMessage        `json:"new_values,omitempty" db:"new_values"`
	Metadata     map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	IPAddress    string                 `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	FullName string   `json:"full_name" binding:"required"`
	Role     UserRole `json:"role"`
}

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

// AddProjectMemberRequest represents a request to add a member to a project
type AddProjectMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}