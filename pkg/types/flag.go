package types

import (
	"encoding/json"
	"time"
)

type Flag struct {
	ID               string                 `json:"id" db:"id"`
	ProjectID        string                 `json:"project_id" db:"project_id"`
	Key              string                 `json:"key" db:"key"`
	Name             string                 `json:"name" db:"name"`
	Description      string                 `json:"description" db:"description"`
	Type             FlagType               `json:"type" db:"type"`
	Enabled          bool                   `json:"enabled" db:"enabled"`
	Default          json.RawMessage        `json:"default" db:"default_value"`
	Variations       []Variation            `json:"variations" db:"variations"`
	Targeting        *TargetingConfig       `json:"targeting" db:"targeting"`
	RolloutConfig    json.RawMessage        `json:"rollout_config,omitempty" db:"rollout_config"`
	ExperimentConfig json.RawMessage        `json:"experiment_config,omitempty" db:"experiment_config"`
	Environment      string                 `json:"environment" db:"environment"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
	Tags             []string               `json:"tags" db:"tags"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
}

type FlagType string

const (
	FlagTypeBoolean FlagType = "boolean"
	FlagTypeString  FlagType = "string"
	FlagTypeNumber  FlagType = "number"
	FlagTypeJSON    FlagType = "json"
)

type Variation struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Value       json.RawMessage `json:"value"`
	Weight      int             `json:"weight"`
}

type TargetingConfig struct {
	Rules    []TargetingRule `json:"rules"`
	Rollout  *RolloutConfig  `json:"rollout,omitempty"`
	Segments []string        `json:"segments,omitempty"`
}

type TargetingRule struct {
	ID          string      `json:"id"`
	Attribute   string      `json:"attribute"`
	Operator    string      `json:"operator"`
	Values      []string    `json:"values"`
	Variation   string      `json:"variation,omitempty"`
	Description string      `json:"description"`
}

type RolloutConfig struct {
	Type        string             `json:"type"`
	Variations  []VariationRollout `json:"variations"`
	BucketBy    string             `json:"bucket_by"`
	Seed        int64              `json:"seed"`
}

type VariationRollout struct {
	VariationID string `json:"variation_id"`
	Weight      int    `json:"weight"`
}

type EvaluationRequest struct {
	ProjectKey  string                 `json:"project_key,omitempty"`
	FlagKey     string                 `json:"flag_key"`
	UserID      string                 `json:"user_id"`
	UserKey     string                 `json:"user_key"`
	Attributes  map[string]interface{} `json:"attributes"`
	Environment string                 `json:"environment"`
}

type EvaluationResponse struct {
	FlagKey     string          `json:"flag_key"`
	Value       json.RawMessage `json:"value"`
	Variation   string          `json:"variation"`
	Reason      string          `json:"reason"`
	RuleID      string          `json:"rule_id,omitempty"`
	Default     bool            `json:"default"`
	Timestamp   time.Time       `json:"timestamp"`
}

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

// RuleEvaluationResult represents the result of evaluating a single rule
type RuleEvaluationResult struct {
	RuleID         string        `json:"rule_id"`
	Matched        bool          `json:"matched"`
	Attribute      string        `json:"attribute"`
	Operator       string        `json:"operator"`
	ActualValue    interface{}   `json:"actual_value"`
	ExpectedValues []interface{} `json:"expected_values"`
	Reason         string        `json:"reason"`
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