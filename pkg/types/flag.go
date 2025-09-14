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
	FlagTypeVariant FlagType = "variant" // A/B testing with multiple variants
)

type Variation struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Value       json.RawMessage `json:"value"`
	Weight      int             `json:"weight"`
}

