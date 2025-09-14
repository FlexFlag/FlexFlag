package types

import (
	"time"
)

// Segment represents a user segment for targeting
type Segment struct {
	ID          string           `json:"id" db:"id"`
	ProjectID   string           `json:"project_id" db:"project_id"`
	Key         string           `json:"key" db:"key"`
	Name        string           `json:"name" db:"name"`
	Description string           `json:"description" db:"description"`
	Rules       []TargetingRule  `json:"rules" db:"rules"`
	Environment string           `json:"environment" db:"environment"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}

// CreateSegmentRequest represents a request to create a segment
type CreateSegmentRequest struct {
	ProjectID   string          `json:"project_id" binding:"required"`
	Key         string          `json:"key" binding:"required"`
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description"`
	Rules       []TargetingRule `json:"rules" binding:"required"`
	Environment string          `json:"environment" binding:"required"`
}

// UpdateSegmentRequest represents a request to update a segment
type UpdateSegmentRequest struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Rules       []TargetingRule `json:"rules"`
}

// SegmentMatchResult represents the result of evaluating a user against a segment
type SegmentMatchResult struct {
	Matched     bool                    `json:"matched"`
	SegmentKey  string                  `json:"segment_key"`
	UserKey     string                  `json:"user_key"`
	Reason      string                  `json:"reason"`
	RuleResults []RuleEvaluationResult  `json:"rule_results,omitempty"`
}

// SegmentEvaluationRequest represents a request to evaluate a segment
type SegmentEvaluationRequest struct {
	ProjectID   string                 `json:"project_id" binding:"required"`
	SegmentKey  string                 `json:"segment_key" binding:"required"`
	UserKey     string                 `json:"user_key" binding:"required"`
	UserID      string                 `json:"user_id"`
	Attributes  map[string]interface{} `json:"attributes"`
	Environment string                 `json:"environment" binding:"required"`
}

// SegmentEvaluationResponse represents the response of a segment evaluation
type SegmentEvaluationResponse struct {
	SegmentKey string `json:"segment_key"`
	UserKey    string `json:"user_key"`
	Matched    bool   `json:"matched"`
	Reason     string `json:"reason,omitempty"`
}