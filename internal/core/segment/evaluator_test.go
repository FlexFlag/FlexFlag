package segment

import (
	"testing"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	assert.NotNil(t, evaluator)
}

func TestEvaluateSegment_NoRules(t *testing.T) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:   "test-segment",
		Name:  "Test Segment",
		Rules: []types.TargetingRule{}, // No rules
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan": "premium",
		},
	}

	result, err := evaluator.EvaluateSegment(segment, userContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-segment", result.SegmentKey)
	assert.Equal(t, "user123", result.UserKey)
	assert.True(t, result.Matched)
	assert.Equal(t, "No rules defined (matches all)", result.Reason)
	assert.Empty(t, result.RuleResults)
}

func TestEvaluateSegment_SingleRule_Match(t *testing.T) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:  "premium-segment",
		Name: "Premium Users",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
		},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan": "premium",
		},
	}

	result, err := evaluator.EvaluateSegment(segment, userContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Matched)
	assert.Equal(t, "All rules matched", result.Reason)
	assert.Len(t, result.RuleResults, 1)
	assert.True(t, result.RuleResults[0].Matched)
	assert.Equal(t, "rule1", result.RuleResults[0].RuleID)
	assert.Equal(t, "premium", result.RuleResults[0].ActualValue)
}

func TestEvaluateSegment_SingleRule_NoMatch(t *testing.T) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:  "premium-segment",
		Name: "Premium Users",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
		},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan": "basic",
		},
	}

	result, err := evaluator.EvaluateSegment(segment, userContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Matched)
	assert.Equal(t, "One or more rules did not match", result.Reason)
	assert.Len(t, result.RuleResults, 1)
	assert.False(t, result.RuleResults[0].Matched)
}

func TestEvaluateSegment_MultipleRules_AllMatch(t *testing.T) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:  "power-users",
		Name: "Power Users",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
			{
				ID:        "rule2",
				Attribute: "region",
				Operator:  "equals",
				Values:    []string{"us-east", "us-west"},
			},
		},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan":   "premium",
			"region": "us-east",
		},
	}

	result, err := evaluator.EvaluateSegment(segment, userContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Matched)
	assert.Equal(t, "All rules matched", result.Reason)
	assert.Len(t, result.RuleResults, 2)
	assert.True(t, result.RuleResults[0].Matched)
	assert.True(t, result.RuleResults[1].Matched)
}

func TestEvaluateSegment_MultipleRules_PartialMatch(t *testing.T) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:  "power-users",
		Name: "Power Users",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
			{
				ID:        "rule2",
				Attribute: "region",
				Operator:  "equals",
				Values:    []string{"us-east", "us-west"},
			},
		},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan":   "premium",
			"region": "europe", // Doesn't match rule2
		},
	}

	result, err := evaluator.EvaluateSegment(segment, userContext)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Matched)
	assert.Equal(t, "One or more rules did not match", result.Reason)
	assert.Len(t, result.RuleResults, 2)
	assert.True(t, result.RuleResults[0].Matched)   // Rule1 matches
	assert.False(t, result.RuleResults[1].Matched) // Rule2 doesn't match
}

func TestEvaluateRule_StringOperators(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name       string
		operator   string
		ruleValues []string
		userValue  interface{}
		expected   bool
	}{
		{"equals_match", "equals", []string{"premium"}, "premium", true},
		{"equals_no_match", "equals", []string{"premium"}, "basic", false},
		{"equals_multiple_values", "equals", []string{"premium", "enterprise"}, "premium", true},
		{"not_equals_match", "not_equals", []string{"premium"}, "basic", true},
		{"not_equals_no_match", "not_equals", []string{"premium"}, "premium", false},
		{"contains_match", "contains", []string{"prem"}, "premium", true},
		{"contains_no_match", "contains", []string{"prem"}, "basic", false},
		{"not_contains_match", "not_contains", []string{"prem"}, "basic", true},
		{"not_contains_no_match", "not_contains", []string{"prem"}, "premium", false},
		{"starts_with_match", "starts_with", []string{"prem"}, "premium", true},
		{"starts_with_no_match", "starts_with", []string{"prem"}, "basic", false},
		{"ends_with_match", "ends_with", []string{"ium"}, "premium", true},
		{"ends_with_no_match", "ends_with", []string{"ium"}, "basic", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := types.TargetingRule{
				ID:        "test-rule",
				Attribute: "plan",
				Operator:  tt.operator,
				Values:    tt.ruleValues,
			}

			userContext := &types.SegmentEvaluationRequest{
				UserKey: "user123",
				Attributes: map[string]interface{}{
					"plan": tt.userValue,
				},
			}

			result := evaluator.evaluateRule(rule, userContext)
			assert.Equal(t, tt.expected, result.Matched, "operator: %s, values: %v, user: %v", tt.operator, tt.ruleValues, tt.userValue)
			assert.Equal(t, "test-rule", result.RuleID)
			assert.Equal(t, "plan", result.Attribute)
			assert.Equal(t, tt.operator, result.Operator)
			assert.Equal(t, tt.userValue, result.ActualValue)
		})
	}
}

func TestEvaluateRule_NumericOperators(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		operator  string
		ruleValue string
		userValue interface{}
		expected  bool
	}{
		{"greater_than_match", "greater_than", "100", 150, true},
		{"greater_than_no_match", "greater_than", "100", 50, false},
		{"greater_than_equal", "greater_than", "100", 100, false}, // Should be false for equal
		{"less_than_match", "less_than", "100", 50, true},
		{"less_than_no_match", "less_than", "100", 150, false},
		{"greater_than_or_equal_match", "greater_than_or_equal", "100", 100, true},
		{"greater_than_or_equal_greater", "greater_than_or_equal", "100", 150, true},
		{"greater_than_or_equal_no_match", "greater_than_or_equal", "100", 50, false},
		{"less_than_or_equal_match", "less_than_or_equal", "100", 100, true},
		{"less_than_or_equal_less", "less_than_or_equal", "100", 50, true},
		{"less_than_or_equal_no_match", "less_than_or_equal", "100", 150, false},
		{"numeric_string_comparison", "greater_than", "100", "150", true},
		{"float_comparison", "greater_than", "10.5", 11.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := types.TargetingRule{
				ID:        "test-rule",
				Attribute: "score",
				Operator:  tt.operator,
				Values:    []string{tt.ruleValue},
			}

			userContext := &types.SegmentEvaluationRequest{
				UserKey: "user123",
				Attributes: map[string]interface{}{
					"score": tt.userValue,
				},
			}

			result := evaluator.evaluateRule(rule, userContext)
			assert.Equal(t, tt.expected, result.Matched, "operator: %s, ruleValue: %s, userValue: %v", tt.operator, tt.ruleValue, tt.userValue)
		})
	}
}

func TestEvaluateRule_RegexOperator(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name      string
		pattern   string
		userValue interface{}
		expected  bool
	}{
		{"email_pattern_match", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "user@example.com", true},
		{"email_pattern_no_match", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`, "invalid-email", false},
		{"phone_pattern", `^\+1\d{10}$`, "+12345678901", true},
		{"phone_pattern_no_match", `^\+1\d{10}$`, "123-456-7890", false},
		{"simple_pattern", `^test.*`, "test123", true},
		{"simple_pattern_no_match", `^test.*`, "hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := types.TargetingRule{
				ID:        "test-rule",
				Attribute: "field",
				Operator:  "regex",
				Values:    []string{tt.pattern},
			}

			userContext := &types.SegmentEvaluationRequest{
				UserKey: "user123",
				Attributes: map[string]interface{}{
					"field": tt.userValue,
				},
			}

			result := evaluator.evaluateRule(rule, userContext)
			assert.Equal(t, tt.expected, result.Matched, "pattern: %s, value: %v", tt.pattern, tt.userValue)
		})
	}
}

func TestEvaluateRule_MissingAttribute(t *testing.T) {
	evaluator := NewEvaluator()

	rule := types.TargetingRule{
		ID:        "test-rule",
		Attribute: "missing_attr",
		Operator:  "equals",
		Values:    []string{"value"},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"other_attr": "value",
		},
	}

	result := evaluator.evaluateRule(rule, userContext)
	assert.False(t, result.Matched)
	assert.Nil(t, result.ActualValue)
	assert.Contains(t, result.Reason, "Attribute 'missing_attr' not found")
}

func TestGetUserAttribute(t *testing.T) {
	evaluator := NewEvaluator()

	userContext := &types.SegmentEvaluationRequest{
		UserID:  "user123",
		UserKey: "userkey456",
		Attributes: map[string]interface{}{
			"plan":   "premium",
			"score":  100,
			"active": true,
		},
	}

	tests := []struct {
		attribute string
		expected  interface{}
	}{
		{"user_id", "user123"},
		{"user_key", "userkey456"},
		{"plan", "premium"},
		{"score", 100},
		{"active", true},
		{"missing", nil},
	}

	for _, tt := range tests {
		t.Run(tt.attribute, func(t *testing.T) {
			result := evaluator.getUserAttribute(tt.attribute, userContext)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestToString(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"float", 123.45, "123.45"},
		{"bool_true", true, "true"},
		{"bool_false", false, "false"},
		{"nil", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.toString(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEvaluateEquals(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		actual   string
		expected []string
		result   bool
	}{
		{"single_match", "premium", []string{"premium"}, true},
		{"single_no_match", "premium", []string{"basic"}, false},
		{"multiple_match", "premium", []string{"basic", "premium", "enterprise"}, true},
		{"multiple_no_match", "premium", []string{"basic", "enterprise"}, false},
		{"empty_values", "premium", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.evaluateEquals(tt.actual, tt.expected)
			assert.Equal(t, tt.result, result)
		})
	}
}

func TestEvaluateContains(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name     string
		actual   string
		values   []string
		expected bool
	}{
		{"contains_match", "premium plan", []string{"prem"}, true},
		{"contains_no_match", "basic plan", []string{"prem"}, false},
		{"multiple_values_match", "premium plan", []string{"basic", "prem"}, true},
		{"multiple_values_no_match", "enterprise plan", []string{"basic", "prem"}, false},
		{"empty_string", "", []string{"test"}, false},
		{"empty_values", "premium", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.evaluateContains(tt.actual, tt.values)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func BenchmarkEvaluateSegment(b *testing.B) {
	evaluator := NewEvaluator()

	segment := &types.Segment{
		Key:  "benchmark-segment",
		Name: "Benchmark Segment",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
			{
				ID:        "rule2",
				Attribute: "region",
				Operator:  "contains",
				Values:    []string{"us"},
			},
		},
	}

	userContext := &types.SegmentEvaluationRequest{
		UserKey: "user123",
		Attributes: map[string]interface{}{
			"plan":   "premium",
			"region": "us-east",
		},
	}

	for i := 0; i < b.N; i++ {
		evaluator.EvaluateSegment(segment, userContext)
	}
}