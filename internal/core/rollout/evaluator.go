package rollout

import (
	"crypto/md5"
	"fmt"
	"math"

	"github.com/flexflag/flexflag/pkg/types"
)

// Evaluator handles rollout evaluation with sticky assignments
type Evaluator struct{}

// NewEvaluator creates a new rollout evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// EvaluateRollout evaluates a rollout configuration for a user
func (e *Evaluator) EvaluateRollout(rollout *types.Rollout, userKey string, stickyAssignment *types.StickyAssignment) (*types.RolloutResult, error) {
	result := &types.RolloutResult{
		RolloutID:   rollout.ID,
		UserKey:     userKey,
		Matched:     false,
		VariationID: "",
		Reason:      "No rollout rules matched",
		IsSticky:    false,
	}

	// Check if we have a sticky assignment first
	if stickyAssignment != nil {
		result.VariationID = stickyAssignment.VariationID
		result.Matched = true
		result.IsSticky = true
		result.Reason = "Sticky assignment found"
		return result, nil
	}

	// Evaluate based on rollout type
	switch rollout.Type {
	case types.RolloutTypePercentage:
		return e.evaluatePercentageRollout(rollout, userKey)
	case types.RolloutTypeExperiment:
		return e.evaluateExperimentRollout(rollout, userKey)
	case types.RolloutTypeSegment:
		return e.evaluateSegmentRollout(rollout, userKey)
	default:
		result.Reason = fmt.Sprintf("Unknown rollout type: %s", rollout.Type)
		return result, nil
	}
}

// evaluatePercentageRollout handles percentage-based rollouts
func (e *Evaluator) evaluatePercentageRollout(rollout *types.Rollout, userKey string) (*types.RolloutResult, error) {
	result := &types.RolloutResult{
		RolloutID:   rollout.ID,
		UserKey:     userKey,
		Matched:     false,
		VariationID: "",
		Reason:      "User not in rollout percentage",
		IsSticky:    false,
	}

	config := rollout.Config

	// Generate bucket key for consistent assignment
	bucketKey := e.generateBucketKey(rollout.FlagID, rollout.Environment, userKey, config.BucketBy)
	
	// Calculate hash-based percentage (0-99)
	userPercentage := e.hashToPercentage(bucketKey)

	// Check traffic allocation first (if specified)
	trafficAllocation := 100
	if config.TrafficAllocation != nil {
		trafficAllocation = *config.TrafficAllocation
	}

	if userPercentage >= trafficAllocation {
		result.Reason = fmt.Sprintf("User percentage %d >= traffic allocation %d", userPercentage, trafficAllocation)
		return result, nil
	}

	// Handle single percentage rollout
	if config.Percentage != nil {
		if userPercentage < *config.Percentage {
			// Find the first variation or use default
			if len(config.Variations) > 0 {
				result.VariationID = config.Variations[0].VariationID
			} else {
				result.VariationID = "on"
			}
			result.Matched = true
			result.Reason = fmt.Sprintf("User percentage %d < rollout percentage %d", userPercentage, *config.Percentage)
		}
		return result, nil
	}

	// Handle weighted variations rollout
	if len(config.Variations) > 0 {
		variationID := e.selectVariationByWeight(config.Variations, userPercentage)
		if variationID != "" {
			result.VariationID = variationID
			result.Matched = true
			result.Reason = fmt.Sprintf("User percentage %d assigned to variation %s", userPercentage, variationID)
		}
		return result, nil
	}

	return result, nil
}

// evaluateExperimentRollout handles A/B test experiments
func (e *Evaluator) evaluateExperimentRollout(rollout *types.Rollout, userKey string) (*types.RolloutResult, error) {
	// For experiments, we use the same logic as percentage rollouts but with additional tracking
	return e.evaluatePercentageRollout(rollout, userKey)
}

// evaluateSegmentRollout handles segment-based rollouts
func (e *Evaluator) evaluateSegmentRollout(rollout *types.Rollout, userKey string) (*types.RolloutResult, error) {
	result := &types.RolloutResult{
		RolloutID:   rollout.ID,
		UserKey:     userKey,
		Matched:     false,
		VariationID: "",
		Reason:      "Segment rollout not implemented yet",
		IsSticky:    false,
	}

	// TODO: Implement segment-based rollouts
	// This would require integration with the segment evaluator

	return result, nil
}

// generateBucketKey creates a consistent key for user bucketing
func (e *Evaluator) generateBucketKey(flagID, environment, userKey, bucketBy string) string {
	if bucketBy == "" {
		bucketBy = "user_key"
	}

	// Create a consistent key for bucketing
	return fmt.Sprintf("%s:%s:%s:%s", flagID, environment, bucketBy, userKey)
}

// GenerateBucketKey exposes the bucket key generation for external use
func (e *Evaluator) GenerateBucketKey(flagID, environment, userKey, bucketBy string) string {
	return e.generateBucketKey(flagID, environment, userKey, bucketBy)
}

// hashToPercentage converts a string to a percentage (0-99)
func (e *Evaluator) hashToPercentage(key string) int {
	hash := md5.Sum([]byte(key))
	
	// Use first 4 bytes to create an integer
	var hashInt uint32
	for i := 0; i < 4; i++ {
		hashInt = hashInt<<8 + uint32(hash[i])
	}
	
	// Convert to percentage (0-99)
	return int(hashInt % 100)
}

// selectVariationByWeight selects a variation based on weighted distribution
func (e *Evaluator) selectVariationByWeight(variations []types.VariationAllocation, userPercentage int) string {
	totalWeight := 0
	for _, v := range variations {
		totalWeight += v.Weight
	}

	if totalWeight == 0 {
		return ""
	}

	// Normalize user percentage to total weight
	scaledPercentage := int(float64(userPercentage) * float64(totalWeight) / 100.0)

	// Find which variation the user falls into
	cumulative := 0
	for _, v := range variations {
		cumulative += v.Weight
		if scaledPercentage < cumulative {
			return v.VariationID
		}
	}

	return ""
}

// CalculateRequiredSampleSize calculates the required sample size for an experiment
func (e *Evaluator) CalculateRequiredSampleSize(baselineRate, minimumDetectableEffect, alpha, power float64) int {
	if baselineRate <= 0 || baselineRate >= 1 {
		return 0
	}

	// Using formula for two-proportion z-test
	// This is a simplified calculation
	z_alpha := 1.96  // For 95% confidence (alpha = 0.05)
	z_beta := 0.84   // For 80% power (beta = 0.20)

	p1 := baselineRate
	p2 := baselineRate * (1 + minimumDetectableEffect)

	if p2 > 1 {
		p2 = 1
	}

	pooled := (p1 + p2) / 2
	
	numerator := math.Pow(z_alpha*math.Sqrt(2*pooled*(1-pooled)) + z_beta*math.Sqrt(p1*(1-p1)+p2*(1-p2)), 2)
	denominator := math.Pow(p2-p1, 2)

	sampleSize := numerator / denominator

	return int(math.Ceil(sampleSize))
}

