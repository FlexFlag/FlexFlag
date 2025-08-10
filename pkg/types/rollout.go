package types

import (
	"time"
)

// Rollout represents a rollout configuration for A/B testing and percentage rollouts
type Rollout struct {
	ID          string          `json:"id" db:"id"`
	FlagID      string          `json:"flag_id" db:"flag_id"`
	Environment string          `json:"environment" db:"environment"`
	Type        RolloutType     `json:"type" db:"type"`
	Name        string          `json:"name" db:"name"`
	Description string          `json:"description" db:"description"`
	Config      AdvancedRolloutConfig   `json:"config" db:"config"`
	Status      RolloutStatus   `json:"status" db:"status"`
	StartDate   *time.Time      `json:"start_date,omitempty" db:"start_date"`
	EndDate     *time.Time      `json:"end_date,omitempty" db:"end_date"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// RolloutType defines the type of rollout
type RolloutType string

const (
	RolloutTypePercentage RolloutType = "percentage"
	RolloutTypeExperiment RolloutType = "experiment"
	RolloutTypeSegment    RolloutType = "segment"
)

// RolloutStatus defines the status of a rollout
type RolloutStatus string

const (
	RolloutStatusActive    RolloutStatus = "active"
	RolloutStatusPaused    RolloutStatus = "paused"
	RolloutStatusCompleted RolloutStatus = "completed"
	RolloutStatusDraft     RolloutStatus = "draft"
)

// AdvancedRolloutConfig contains configuration for different rollout types
type AdvancedRolloutConfig struct {
	// For percentage rollouts
	Percentage   *int                   `json:"percentage,omitempty"`
	Variations   []VariationAllocation  `json:"variations,omitempty"`
	
	// For experiments
	HypothesisDescription string                 `json:"hypothesis_description,omitempty"`
	SuccessMetrics       []string               `json:"success_metrics,omitempty"`
	RequiredSampleSize   *int                   `json:"required_sample_size,omitempty"`
	SignificanceLevel    *float64              `json:"significance_level,omitempty"`
	
	// For segment rollouts
	TargetSegments []string `json:"target_segments,omitempty"`
	
	// Common settings
	BucketBy         string `json:"bucket_by"`         // user_id, user_key, custom_attribute
	StickyBucketing  bool   `json:"sticky_bucketing"`  // whether to maintain consistent assignments
	TrafficAllocation *int   `json:"traffic_allocation,omitempty"` // percentage of traffic to include in rollout
	
	// Advanced settings
	Seed            *int64                 `json:"seed,omitempty"`
	CustomRules     []CustomRolloutRule    `json:"custom_rules,omitempty"`
	ExcludedUsers   []string               `json:"excluded_users,omitempty"`
	IncludedUsers   []string               `json:"included_users,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// VariationAllocation represents how traffic is allocated to variations
type VariationAllocation struct {
	VariationID string `json:"variation_id"`
	Weight      int    `json:"weight"` // out of 100
}

// CustomRolloutRule allows for complex rollout logic
type CustomRolloutRule struct {
	ID          string                 `json:"id"`
	Description string                 `json:"description"`
	Condition   RuleCondition          `json:"condition"`
	Allocation  []VariationAllocation  `json:"allocation"`
}

// RuleCondition defines conditions for custom rollout rules
type RuleCondition struct {
	Attribute string      `json:"attribute"`
	Operator  string      `json:"operator"` // equals, not_equals, in, not_in, greater_than, less_than, contains, regex
	Values    []string    `json:"values"`
	LogicType string      `json:"logic_type"` // and, or
	SubRules  []RuleCondition `json:"sub_rules,omitempty"`
}

// StickyAssignment represents a user's sticky assignment to a variation
type StickyAssignment struct {
	ID          string     `json:"id" db:"id"`
	FlagID      string     `json:"flag_id" db:"flag_id"`
	Environment string     `json:"environment" db:"environment"`
	UserID      string     `json:"user_id" db:"user_id"`
	UserKey     string     `json:"user_key" db:"user_key"`
	VariationID string     `json:"variation_id" db:"variation_id"`
	BucketKey   string     `json:"bucket_key" db:"bucket_key"`
	AssignedAt  time.Time  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

// CreateRolloutRequest represents a request to create a rollout
type CreateRolloutRequest struct {
	FlagID      string        `json:"flag_id" binding:"required"`
	Environment string        `json:"environment" binding:"required"`
	Type        RolloutType   `json:"type" binding:"required"`
	Name        string        `json:"name" binding:"required"`
	Description string        `json:"description"`
	Config      AdvancedRolloutConfig `json:"config" binding:"required"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
}

// UpdateRolloutRequest represents a request to update a rollout
type UpdateRolloutRequest struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Config      AdvancedRolloutConfig `json:"config"`
	Status      RolloutStatus `json:"status"`
	StartDate   *time.Time    `json:"start_date,omitempty"`
	EndDate     *time.Time    `json:"end_date,omitempty"`
}

// RolloutAnalytics represents analytics data for a rollout
type RolloutAnalytics struct {
	RolloutID      string                    `json:"rollout_id"`
	TotalUsers     int                       `json:"total_users"`
	VariationStats []VariationAnalytics      `json:"variation_stats"`
	ConversionRate map[string]float64        `json:"conversion_rate,omitempty"`
	StatisticalSignificance *float64         `json:"statistical_significance,omitempty"`
	Confidence     *float64                  `json:"confidence,omitempty"`
	TimeSeriesData []RolloutTimeSeriesPoint  `json:"time_series_data,omitempty"`
}

// VariationAnalytics represents analytics for a specific variation
type VariationAnalytics struct {
	VariationID   string  `json:"variation_id"`
	UserCount     int     `json:"user_count"`
	Percentage    float64 `json:"percentage"`
	Conversions   int     `json:"conversions,omitempty"`
	ConversionRate *float64 `json:"conversion_rate,omitempty"`
}

// RolloutTimeSeriesPoint represents a data point in rollout time series
type RolloutTimeSeriesPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	UserCount     int       `json:"user_count"`
	VariationData map[string]int `json:"variation_data"`
}

// RolloutResult represents the result of a rollout evaluation
type RolloutResult struct {
	RolloutID   string `json:"rollout_id"`
	UserKey     string `json:"user_key"`
	Matched     bool   `json:"matched"`
	VariationID string `json:"variation_id"`
	Reason      string `json:"reason"`
	IsSticky    bool   `json:"is_sticky"`
	BucketKey   string `json:"bucket_key,omitempty"`
}