package types

import (
	"encoding/json"
	"time"
)

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