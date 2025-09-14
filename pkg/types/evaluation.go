package types

import (
	"encoding/json"
	"time"
)

// EvaluationRequest represents a request to evaluate a flag
type EvaluationRequest struct {
	ProjectKey  string                 `json:"project_key,omitempty"`
	FlagKey     string                 `json:"flag_key"`
	UserID      string                 `json:"user_id"`
	UserKey     string                 `json:"user_key"`
	Attributes  map[string]interface{} `json:"attributes"`
	Environment string                 `json:"environment"`
}

// EvaluationResponse represents the response from a flag evaluation
type EvaluationResponse struct {
	FlagKey        string          `json:"flag_key"`
	Value          json.RawMessage `json:"value"`
	Variation      string          `json:"variation"`
	Reason         string          `json:"reason"`
	RuleID         string          `json:"rule_id,omitempty"`
	Default        bool            `json:"default"`
	Timestamp      time.Time       `json:"timestamp"`
	EvaluationTime float64         `json:"evaluation_time,omitempty"` // in milliseconds
	Source         string          `json:"source,omitempty"`          // e.g., "edge-cache", "central-hub"
}