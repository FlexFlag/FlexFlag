package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestEngine_NewEngine(t *testing.T) {
	engine := NewEngine()
	assert.NotNil(t, engine)
	assert.NotNil(t, engine.flags)
	assert.NotNil(t, engine.segments)
	assert.Len(t, engine.flags, 0)
	assert.Len(t, engine.segments, 0)
}

func TestEngine_EvaluateFlag_NotFound(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	req := &types.EvaluationRequest{
		FlagKey: "non-existent",
		UserID:  "user123",
	}

	result, err := engine.EvaluateFlag(ctx, req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "flag not found: non-existent")
}

func TestEngine_EvaluateFlag_Disabled(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	flag := &types.Flag{
		Key:     "disabled-flag",
		Name:    "Disabled Flag",
		Type:    types.FlagTypeBoolean,
		Enabled: false,
		Default: json.RawMessage(`false`),
	}

	engine.UpdateFlag(flag)

	req := &types.EvaluationRequest{
		FlagKey: "disabled-flag",
		UserID:  "user123",
	}

	result, err := engine.EvaluateFlag(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "disabled-flag", result.FlagKey)
	assert.Equal(t, flag.Default, result.Value)
	assert.Equal(t, "flag_disabled", result.Reason)
	assert.True(t, result.Default)
}

func TestEngine_EvaluateFlag_Enabled_Default(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	flag := &types.Flag{
		Key:     "simple-flag",
		Name:    "Simple Flag",
		Type:    types.FlagTypeBoolean,
		Enabled: true,
		Default: json.RawMessage(`true`),
	}

	engine.UpdateFlag(flag)

	req := &types.EvaluationRequest{
		FlagKey: "simple-flag",
		UserID:  "user123",
	}

	result, err := engine.EvaluateFlag(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "simple-flag", result.FlagKey)
	assert.Equal(t, flag.Default, result.Value)
	assert.Equal(t, "default", result.Reason)
	assert.True(t, result.Default)
}

func TestEngine_EvaluateFlag_WithTargeting_RuleMatch(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	flag := &types.Flag{
		Key:     "targeting-flag",
		Name:    "Targeting Flag",
		Type:    types.FlagTypeVariant,
		Enabled: true,
		Default: json.RawMessage(`"default"`),
		Variations: []types.Variation{
			{
				ID:    "var1",
				Name:  "Variation 1",
				Value: json.RawMessage(`"premium"`),
			},
			{
				ID:    "var2",
				Name:  "Variation 2",
				Value: json.RawMessage(`"basic"`),
			},
		},
		Targeting: &types.TargetingConfig{
			Rules: []types.TargetingRule{
				{
					ID:        "rule1",
					Attribute: "plan",
					Operator:  "eq",
					Values:    []string{"premium"},
					Variation: "var1",
				},
				{
					ID:        "rule2",
					Attribute: "region",
					Operator:  "in",
					Values:    []string{"us-east", "us-west"},
					Variation: "var2",
				},
			},
		},
	}

	engine.UpdateFlag(flag)

	// Test premium plan match
	req1 := &types.EvaluationRequest{
		FlagKey: "targeting-flag",
		UserID:  "user123",
		Attributes: map[string]interface{}{
			"plan": "premium",
		},
	}

	result1, err := engine.EvaluateFlag(ctx, req1)
	assert.NoError(t, err)
	assert.NotNil(t, result1)
	assert.Equal(t, "targeting-flag", result1.FlagKey)
	assert.Equal(t, json.RawMessage(`"premium"`), result1.Value)
	assert.Equal(t, "var1", result1.Variation)
	assert.Equal(t, "rule_match", result1.Reason)
	assert.Equal(t, "rule1", result1.RuleID)
	assert.False(t, result1.Default)

	// Test region match
	req2 := &types.EvaluationRequest{
		FlagKey: "targeting-flag",
		UserID:  "user456",
		Attributes: map[string]interface{}{
			"plan":   "basic",
			"region": "us-east",
		},
	}

	result2, err := engine.EvaluateFlag(ctx, req2)
	assert.NoError(t, err)
	assert.NotNil(t, result2)
	assert.Equal(t, json.RawMessage(`"basic"`), result2.Value)
	assert.Equal(t, "var2", result2.Variation)
	assert.Equal(t, "rule_match", result2.Reason)
	assert.Equal(t, "rule2", result2.RuleID)
}

func TestEngine_EvaluateRule_StringOperators(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name       string
		operator   string
		ruleValues []string
		userValue  interface{}
		expected   bool
	}{
		{"eq_match", "eq", []string{"premium"}, "premium", true},
		{"eq_no_match", "eq", []string{"premium"}, "basic", false},
		{"ne_match", "ne", []string{"premium"}, "basic", true},
		{"ne_no_match", "ne", []string{"premium"}, "premium", false},
		{"in_match", "in", []string{"premium", "enterprise"}, "premium", true},
		{"in_no_match", "in", []string{"premium", "enterprise"}, "basic", false},
		{"not_in_match", "not_in", []string{"premium", "enterprise"}, "basic", true},
		{"not_in_no_match", "not_in", []string{"premium", "enterprise"}, "premium", false},
		{"contains_match", "contains", []string{"prem"}, "premium", true},
		{"contains_no_match", "contains", []string{"prem"}, "basic", false},
		{"starts_with_match", "starts_with", []string{"prem"}, "premium", true},
		{"starts_with_no_match", "starts_with", []string{"prem"}, "basic", false},
		{"ends_with_match", "ends_with", []string{"ium"}, "premium", true},
		{"ends_with_no_match", "ends_with", []string{"ium"}, "basic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &types.TargetingRule{
				ID:        "test-rule",
				Attribute: "plan",
				Operator:  tt.operator,
				Values:    tt.ruleValues,
			}

			req := &types.EvaluationRequest{
				FlagKey: "test-flag",
				UserID:  "user123",
				Attributes: map[string]interface{}{
					"plan": tt.userValue,
				},
			}

			result := engine.evaluateRule(req, rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEngine_EvaluateRule_NumericOperators(t *testing.T) {
	engine := NewEngine()

	tests := []struct {
		name      string
		operator  string
		ruleValue string
		userValue interface{}
		expected  bool
	}{
		{"gt_match", "gt", "100", "150", true},
		{"gt_no_match", "gt", "100", "50", false},
		{"gte_match", "gte", "100", "100", true},
		{"gte_no_match", "gte", "100", "50", false},
		{"lt_match", "lt", "100", "50", true},
		{"lt_no_match", "lt", "100", "150", false},
		{"lte_match", "lte", "100", "100", true},
		{"lte_no_match", "lte", "100", "150", false},
		{"gt_float", "gt", "10.5", "11.0", true},
		{"invalid_user_value", "gt", "100", "invalid", false},
		{"invalid_rule_value", "gt", "invalid", "100", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &types.TargetingRule{
				ID:        "test-rule",
				Attribute: "score",
				Operator:  tt.operator,
				Values:    []string{tt.ruleValue},
			}

			req := &types.EvaluationRequest{
				FlagKey: "test-flag",
				UserID:  "user123",
				Attributes: map[string]interface{}{
					"score": tt.userValue,
				},
			}

			result := engine.evaluateRule(req, rule)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEngine_EvaluateRule_MissingAttribute(t *testing.T) {
	engine := NewEngine()

	rule := &types.TargetingRule{
		ID:        "test-rule",
		Attribute: "missing_attr",
		Operator:  "eq",
		Values:    []string{"value"},
	}

	req := &types.EvaluationRequest{
		FlagKey: "test-flag",
		UserID:  "user123",
		Attributes: map[string]interface{}{
			"other_attr": "value",
		},
	}

	result := engine.evaluateRule(req, rule)
	assert.False(t, result)
}

func TestEngine_EvaluateFlag_WithRollout(t *testing.T) {
	engine := NewEngine()
	ctx := context.Background()

	flag := &types.Flag{
		Key:     "rollout-flag",
		Name:    "Rollout Flag",
		Type:    types.FlagTypeVariant,
		Enabled: true,
		Default: json.RawMessage(`"default"`),
		Variations: []types.Variation{
			{
				ID:    "var1",
				Name:  "Variation 1",
				Value: json.RawMessage(`"A"`),
			},
			{
				ID:    "var2",
				Name:  "Variation 2", 
				Value: json.RawMessage(`"B"`),
			},
		},
		Targeting: &types.TargetingConfig{
			Rollout: &types.RolloutConfig{
				Type:     "weighted",
				BucketBy: "user_id",
				Seed:     12345,
				Variations: []types.VariationRollout{
					{VariationID: "var1", Weight: 50000}, // 50%
					{VariationID: "var2", Weight: 50000}, // 50%
				},
			},
		},
	}

	engine.UpdateFlag(flag)

	// Test multiple users to ensure consistent bucketing
	results := make(map[string]int)
	
	for i := 0; i < 100; i++ {
		req := &types.EvaluationRequest{
			FlagKey: "rollout-flag",
			UserID:  fmt.Sprintf("user%d", i),
		}

		result, err := engine.EvaluateFlag(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "rollout", result.Reason)
		
		var value string
		_ = json.Unmarshal(result.Value, &value)
		results[value]++
	}

	// Both variations should have some users (allowing for randomness)
	assert.Greater(t, results["A"], 0)
	assert.Greater(t, results["B"], 0)
}

func TestEngine_FindVariation(t *testing.T) {
	engine := NewEngine()

	flag := &types.Flag{
		Variations: []types.Variation{
			{ID: "var1", Value: json.RawMessage(`"A"`)},
			{ID: "var2", Value: json.RawMessage(`"B"`)},
		},
	}

	// Test finding existing variation
	variation := engine.findVariation(flag, "var1")
	assert.NotNil(t, variation)
	assert.Equal(t, "var1", variation.ID)

	// Test finding non-existent variation
	variation = engine.findVariation(flag, "non-existent")
	assert.Nil(t, variation)
}

func TestEngine_UpdateFlag(t *testing.T) {
	engine := NewEngine()

	flag := &types.Flag{
		Key:  "test-flag",
		Name: "Test Flag",
	}

	engine.UpdateFlag(flag)

	// Verify flag was added
	engine.mu.RLock()
	storedFlag, exists := engine.flags["test-flag"]
	engine.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, flag, storedFlag)
}

func TestEngine_UpdateSegment(t *testing.T) {
	engine := NewEngine()

	segment := &types.Segment{
		Key:  "test-segment",
		Name: "Test Segment",
	}

	engine.UpdateSegment(segment)

	// Verify segment was added
	engine.mu.RLock()
	storedSegment, exists := engine.segments["test-segment"]
	engine.mu.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, segment, storedSegment)
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	assert.True(t, contains(slice, "a"))
	assert.True(t, contains(slice, "b"))
	assert.True(t, contains(slice, "c"))
	assert.False(t, contains(slice, "d"))
	assert.False(t, contains([]string{}, "a"))
}