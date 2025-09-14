package types

// TargetingConfig represents the targeting configuration for a flag
type TargetingConfig struct {
	Rules    []TargetingRule `json:"rules"`
	Rollout  *RolloutConfig  `json:"rollout,omitempty"`
	Segments []string        `json:"segments,omitempty"`
}

// TargetingRule represents a single targeting rule
type TargetingRule struct {
	ID          string      `json:"id"`
	Attribute   string      `json:"attribute"`
	Operator    string      `json:"operator"`
	Values      []string    `json:"values"`
	Variation   string      `json:"variation,omitempty"`
	Description string      `json:"description"`
}

// RolloutConfig represents the rollout configuration for gradual releases
type RolloutConfig struct {
	Type            string             `json:"type"`
	Variations      []VariationRollout `json:"variations"`
	BucketBy        string             `json:"bucket_by"`
	Seed            int64              `json:"seed"`
	StickyBucketing bool               `json:"sticky_bucketing,omitempty"`
}

// VariationRollout represents a variation's weight in a rollout
type VariationRollout struct {
	VariationID string `json:"variation_id"`
	Weight      int    `json:"weight"`
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