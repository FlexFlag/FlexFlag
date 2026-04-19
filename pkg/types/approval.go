package types

import "time"

type ApprovalStatus string
type ChangeType string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"

	ChangeTypeToggle ChangeType = "toggle"
	ChangeTypeUpdate ChangeType = "update"
	ChangeTypeDelete ChangeType = "delete"
)

// ApprovalRequest represents a pending flag change waiting for human sign-off.
type ApprovalRequest struct {
	ID                string                 `json:"id"`
	ProjectID         string                 `json:"project_id"`
	FlagKey           string                 `json:"flag_key"`
	Environment       string                 `json:"environment"`
	ChangeType        ChangeType             `json:"change_type"`
	Payload           map[string]interface{} `json:"payload"` // the full pending change
	RequestedBy       string                 `json:"requested_by"`
	RequestedByName   string                 `json:"requested_by_name"`
	Status            ApprovalStatus         `json:"status"`
	ReviewNote        string                 `json:"review_note,omitempty"`
	ReviewedBy        string                 `json:"reviewed_by,omitempty"`
	ReviewedAt        *time.Time             `json:"reviewed_at,omitempty"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
}

type ReviewRequest struct {
	Note string `json:"note"`
}
