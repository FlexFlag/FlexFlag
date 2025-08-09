package evaluation

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

type Engine struct {
	flags    map[string]*types.Flag
	segments map[string]*types.Segment
	mu       sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		flags:    make(map[string]*types.Flag),
		segments: make(map[string]*types.Segment),
	}
}

func (e *Engine) EvaluateFlag(ctx context.Context, req *types.EvaluationRequest) (*types.EvaluationResponse, error) {
	e.mu.RLock()
	flag, exists := e.flags[req.FlagKey]
	e.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("flag not found: %s", req.FlagKey)
	}

	response := &types.EvaluationResponse{
		FlagKey:   req.FlagKey,
		Timestamp: time.Now(),
	}

	if !flag.Enabled {
		response.Value = flag.Default
		response.Reason = "flag_disabled"
		response.Default = true
		return response, nil
	}

	if flag.Targeting != nil {
		if variation, reason, ruleID := e.evaluateTargeting(req, flag); variation != nil {
			response.Value = variation.Value
			response.Variation = variation.ID
			response.Reason = reason
			response.RuleID = ruleID
			return response, nil
		}
	}

	response.Value = flag.Default
	response.Reason = "default"
	response.Default = true
	return response, nil
}

func (e *Engine) evaluateTargeting(req *types.EvaluationRequest, flag *types.Flag) (*types.Variation, string, string) {
	for _, rule := range flag.Targeting.Rules {
		if e.evaluateRule(req, &rule) {
			variation := e.findVariation(flag, rule.Variation)
			if variation != nil {
				return variation, "rule_match", rule.ID
			}
		}
	}

	if flag.Targeting.Rollout != nil {
		variation := e.evaluateRollout(req, flag)
		if variation != nil {
			return variation, "rollout", ""
		}
	}

	return nil, "", ""
}

func (e *Engine) evaluateRule(req *types.EvaluationRequest, rule *types.TargetingRule) bool {
	value, exists := req.Attributes[rule.Attribute]
	if !exists {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)

	switch rule.Operator {
	case "eq":
		return contains(rule.Values, valueStr)
	case "ne":
		return !contains(rule.Values, valueStr)
	case "in":
		return contains(rule.Values, valueStr)
	case "not_in":
		return !contains(rule.Values, valueStr)
	case "gt":
		return e.compareNumeric(valueStr, rule.Values[0], func(a, b float64) bool { return a > b })
	case "gte":
		return e.compareNumeric(valueStr, rule.Values[0], func(a, b float64) bool { return a >= b })
	case "lt":
		return e.compareNumeric(valueStr, rule.Values[0], func(a, b float64) bool { return a < b })
	case "lte":
		return e.compareNumeric(valueStr, rule.Values[0], func(a, b float64) bool { return a <= b })
	case "contains":
		return strings.Contains(valueStr, rule.Values[0])
	case "starts_with":
		return strings.HasPrefix(valueStr, rule.Values[0])
	case "ends_with":
		return strings.HasSuffix(valueStr, rule.Values[0])
	}

	return false
}

func (e *Engine) compareNumeric(valueStr, targetStr string, compareFn func(float64, float64) bool) bool {
	value, err1 := strconv.ParseFloat(valueStr, 64)
	target, err2 := strconv.ParseFloat(targetStr, 64)
	if err1 != nil || err2 != nil {
		return false
	}
	return compareFn(value, target)
}

func (e *Engine) evaluateRollout(req *types.EvaluationRequest, flag *types.Flag) *types.Variation {
	bucketKey := req.UserID
	if flag.Targeting.Rollout.BucketBy != "" {
		if val, exists := req.Attributes[flag.Targeting.Rollout.BucketBy]; exists {
			bucketKey = fmt.Sprintf("%v", val)
		}
	}

	hashInput := fmt.Sprintf("%s:%s:%d", flag.Key, bucketKey, flag.Targeting.Rollout.Seed)
	hash := md5.Sum([]byte(hashInput))
	hashStr := hex.EncodeToString(hash[:])
	
	bucket := 0
	for i := 0; i < 8; i++ {
		bucket = bucket*16 + int(hashStr[i])
		if hashStr[i] >= '0' && hashStr[i] <= '9' {
			bucket = bucket - int('0')
		} else {
			bucket = bucket - int('a') + 10
		}
	}
	bucket = bucket % 100000

	cumulative := 0
	for _, vr := range flag.Targeting.Rollout.Variations {
		cumulative += vr.Weight
		if bucket < cumulative {
			return e.findVariation(flag, vr.VariationID)
		}
	}

	return nil
}

func (e *Engine) findVariation(flag *types.Flag, variationID string) *types.Variation {
	for _, v := range flag.Variations {
		if v.ID == variationID {
			return &v
		}
	}
	return nil
}

func (e *Engine) UpdateFlag(flag *types.Flag) {
	e.mu.Lock()
	e.flags[flag.Key] = flag
	e.mu.Unlock()
}

func (e *Engine) UpdateSegment(segment *types.Segment) {
	e.mu.Lock()
	e.segments[segment.Key] = segment
	e.mu.Unlock()
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}