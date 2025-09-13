package rollout

import (
	"fmt"
	"testing"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewEvaluator(t *testing.T) {
	evaluator := NewEvaluator()
	assert.NotNil(t, evaluator)
}

func TestEvaluateRollout_StickyAssignment(t *testing.T) {
	evaluator := NewEvaluator()

	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypePercentage,
	}

	stickyAssignment := &types.StickyAssignment{
		FlagID:      "flag_1",
		Environment: "production",
		UserKey:     "user123",
		VariationID: "variation_a",
	}

	result, err := evaluator.EvaluateRollout(rollout, "user123", stickyAssignment)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rollout_1", result.RolloutID)
	assert.Equal(t, "user123", result.UserKey)
	assert.True(t, result.Matched)
	assert.Equal(t, "variation_a", result.VariationID)
	assert.True(t, result.IsSticky)
	assert.Equal(t, "Sticky assignment found", result.Reason)
}

func TestEvaluateRollout_UnknownType(t *testing.T) {
	evaluator := NewEvaluator()

	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        "unknown",
	}

	result, err := evaluator.EvaluateRollout(rollout, "user123", nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Matched)
	assert.Equal(t, "Unknown rollout type: unknown", result.Reason)
}

func TestEvaluatePercentageRollout_SinglePercentage(t *testing.T) {
	evaluator := NewEvaluator()

	percentage := 50
	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypePercentage,
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage,
		},
	}

	// Test multiple users to check consistency
	results := make(map[bool]int)
	
	for i := 0; i < 100; i++ {
		userKey := fmt.Sprintf("user%d", i)
		result, err := evaluator.evaluatePercentageRollout(rollout, userKey)
		
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "rollout_1", result.RolloutID)
		assert.Equal(t, userKey, result.UserKey)
		
		results[result.Matched]++
	}

	// Both matched and unmatched should have some users (allowing for randomness)
	assert.Greater(t, results[true], 10)   // At least 10% should match
	assert.Greater(t, results[false], 10)  // At least 10% should not match
}

func TestEvaluatePercentageRollout_WeightedVariations(t *testing.T) {
	evaluator := NewEvaluator()

	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypePercentage,
		Config: types.AdvancedRolloutConfig{
			Variations: []types.VariationAllocation{
				{VariationID: "var_a", Weight: 30},
				{VariationID: "var_b", Weight: 70},
			},
		},
	}

	results := make(map[string]int)
	
	for i := 0; i < 1000; i++ {
		userKey := fmt.Sprintf("user%d", i)
		result, err := evaluator.evaluatePercentageRollout(rollout, userKey)
		
		assert.NoError(t, err)
		if result.Matched {
			results[result.VariationID]++
		} else {
			results["unmatched"]++
		}
	}

	// Both variations should have some users
	assert.Greater(t, results["var_a"], 50)   // Should be roughly 30%
	assert.Greater(t, results["var_b"], 50)   // Should be roughly 70%
}

func TestEvaluatePercentageRollout_TrafficAllocation(t *testing.T) {
	evaluator := NewEvaluator()

	percentage := 100
	trafficAllocation := 20 // Only 20% of users see this test
	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypePercentage,
		Config: types.AdvancedRolloutConfig{
			Percentage:        &percentage,
			TrafficAllocation: &trafficAllocation,
		},
	}

	matched := 0
	for i := 0; i < 1000; i++ {
		userKey := fmt.Sprintf("user%d", i)
		result, err := evaluator.evaluatePercentageRollout(rollout, userKey)
		
		assert.NoError(t, err)
		if result.Matched {
			matched++
		}
	}

	// Should be roughly 20% matched due to traffic allocation
	assert.Greater(t, matched, 100)  // At least 10%
	assert.Less(t, matched, 400)     // At most 40%
}

func TestEvaluateExperimentRollout(t *testing.T) {
	evaluator := NewEvaluator()

	percentage := 50
	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypeExperiment,
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage,
		},
	}

	result, err := evaluator.evaluateExperimentRollout(rollout, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rollout_1", result.RolloutID)
	assert.Equal(t, "user123", result.UserKey)
}

func TestEvaluateSegmentRollout(t *testing.T) {
	evaluator := NewEvaluator()

	rollout := &types.Rollout{
		ID:          "rollout_1",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypeSegment,
	}

	result, err := evaluator.evaluateSegmentRollout(rollout, "user123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rollout_1", result.RolloutID)
	assert.Equal(t, "user123", result.UserKey)
	assert.False(t, result.Matched)
	assert.Equal(t, "Segment rollout not implemented yet", result.Reason)
}

func TestGenerateBucketKey(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name        string
		flagID      string
		environment string
		userKey     string
		bucketBy    string
		expected    string
	}{
		{
			name:        "default_bucket_by",
			flagID:      "flag1",
			environment: "prod",
			userKey:     "user123",
			bucketBy:    "",
			expected:    "flag1:prod:user_key:user123",
		},
		{
			name:        "custom_bucket_by",
			flagID:      "flag1",
			environment: "prod",
			userKey:     "user123",
			bucketBy:    "account_id",
			expected:    "flag1:prod:account_id:user123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.GenerateBucketKey(tt.flagID, tt.environment, tt.userKey, tt.bucketBy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHashToPercentage(t *testing.T) {
	evaluator := NewEvaluator()

	// Test that same input always produces same output
	key := "test-key"
	result1 := evaluator.hashToPercentage(key)
	result2 := evaluator.hashToPercentage(key)
	assert.Equal(t, result1, result2)

	// Test that result is in valid range
	assert.GreaterOrEqual(t, result1, 0)
	assert.Less(t, result1, 100)

	// Test different keys produce different results (mostly)
	key2 := "different-key"
	result3 := evaluator.hashToPercentage(key2)
	// While not guaranteed, extremely likely to be different
	assert.NotEqual(t, result1, result3)
}

func TestSelectVariationByWeight(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name            string
		variations      []types.VariationAllocation
		userPercentage  int
		expectedVariation string
	}{
		{
			name: "first_variation",
			variations: []types.VariationAllocation{
				{VariationID: "var_a", Weight: 50},
				{VariationID: "var_b", Weight: 50},
			},
			userPercentage:    10,
			expectedVariation: "var_a",
		},
		{
			name: "second_variation",
			variations: []types.VariationAllocation{
				{VariationID: "var_a", Weight: 30},
				{VariationID: "var_b", Weight: 70},
			},
			userPercentage:    80,
			expectedVariation: "var_b",
		},
		{
			name: "empty_variations",
			variations: []types.VariationAllocation{},
			userPercentage:    50,
			expectedVariation: "",
		},
		{
			name: "zero_weight",
			variations: []types.VariationAllocation{
				{VariationID: "var_a", Weight: 0},
				{VariationID: "var_b", Weight: 0},
			},
			userPercentage:    50,
			expectedVariation: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.selectVariationByWeight(tt.variations, tt.userPercentage)
			assert.Equal(t, tt.expectedVariation, result)
		})
	}
}

func TestCalculateRequiredSampleSize(t *testing.T) {
	evaluator := NewEvaluator()

	tests := []struct {
		name                     string
		baselineRate            float64
		minimumDetectableEffect float64
		alpha                   float64
		power                   float64
		expectedMin             int
		expectedMax             int
	}{
		{
			name:                     "typical_conversion_test",
			baselineRate:            0.10, // 10% baseline conversion
			minimumDetectableEffect: 0.20, // 20% relative improvement
			alpha:                   0.05,
			power:                   0.80,
			expectedMin:             1000,
			expectedMax:             5000,
		},
		{
			name:                     "invalid_baseline_rate_zero",
			baselineRate:            0.0,
			minimumDetectableEffect: 0.20,
			alpha:                   0.05,
			power:                   0.80,
			expectedMin:             0,
			expectedMax:             0,
		},
		{
			name:                     "invalid_baseline_rate_one",
			baselineRate:            1.0,
			minimumDetectableEffect: 0.20,
			alpha:                   0.05,
			power:                   0.80,
			expectedMin:             0,
			expectedMax:             0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := evaluator.CalculateRequiredSampleSize(
				tt.baselineRate,
				tt.minimumDetectableEffect,
				tt.alpha,
				tt.power,
			)
			
			assert.GreaterOrEqual(t, result, tt.expectedMin)
			assert.LessOrEqual(t, result, tt.expectedMax)
		})
	}
}

func BenchmarkHashToPercentage(b *testing.B) {
	evaluator := NewEvaluator()
	key := "benchmark-key"
	
	for i := 0; i < b.N; i++ {
		evaluator.hashToPercentage(key)
	}
}

func BenchmarkEvaluatePercentageRollout(b *testing.B) {
	evaluator := NewEvaluator()
	percentage := 50
	rollout := &types.Rollout{
		ID:          "benchmark_rollout",
		FlagID:      "flag_1",
		Environment: "production",
		Type:        types.RolloutTypePercentage,
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage,
		},
	}

	for i := 0; i < b.N; i++ {
		_, _ = evaluator.evaluatePercentageRollout(rollout, "user123")
	}
}