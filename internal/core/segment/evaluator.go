package segment

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/flexflag/flexflag/pkg/types"
)

// Evaluator handles segment rule evaluation
type Evaluator struct{}

// NewEvaluator creates a new segment evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// EvaluateSegment evaluates if a user context matches a segment
func (e *Evaluator) EvaluateSegment(segment *types.Segment, userContext *types.SegmentEvaluationRequest) (*types.SegmentMatchResult, error) {
	result := &types.SegmentMatchResult{
		SegmentKey: segment.Key,
		UserKey:    userContext.UserKey,
		Matched:    false,
		Reason:     "No rules matched",
	}

	if len(segment.Rules) == 0 {
		result.Matched = true
		result.Reason = "No rules defined (matches all)"
		return result, nil
	}

	// Evaluate all rules - all must match for the segment to match
	allMatched := true
	var ruleResults []types.RuleEvaluationResult

	for _, rule := range segment.Rules {
		ruleResult := e.evaluateRule(rule, userContext)
		ruleResults = append(ruleResults, ruleResult)

		if !ruleResult.Matched {
			allMatched = false
		}
	}

	result.RuleResults = ruleResults
	result.Matched = allMatched

	if allMatched {
		result.Reason = "All rules matched"
	} else {
		result.Reason = "One or more rules did not match"
	}

	return result, nil
}

// evaluateRule evaluates a single targeting rule against user context
func (e *Evaluator) evaluateRule(rule types.TargetingRule, userContext *types.SegmentEvaluationRequest) types.RuleEvaluationResult {
	result := types.RuleEvaluationResult{
		RuleID:         rule.ID,
		Matched:        false,
		Attribute:      rule.Attribute,
		Operator:       rule.Operator,
		ExpectedValues: make([]interface{}, len(rule.Values)),
	}

	// Convert string values to interfaces for the result
	for i, v := range rule.Values {
		result.ExpectedValues[i] = v
	}

	// Get the actual value from user context
	actualValue := e.getUserAttribute(rule.Attribute, userContext)
	result.ActualValue = actualValue

	if actualValue == nil {
		result.Reason = fmt.Sprintf("Attribute '%s' not found in user context", rule.Attribute)
		return result
	}

	// Convert actual value to string for comparison
	actualStr := e.toString(actualValue)

	// Evaluate based on operator
	switch rule.Operator {
	case "equals":
		result.Matched = e.evaluateEquals(actualStr, rule.Values)
	case "not_equals":
		result.Matched = !e.evaluateEquals(actualStr, rule.Values)
	case "contains":
		result.Matched = e.evaluateContains(actualStr, rule.Values)
	case "not_contains":
		result.Matched = !e.evaluateContains(actualStr, rule.Values)
	case "starts_with":
		result.Matched = e.evaluateStartsWith(actualStr, rule.Values)
	case "ends_with":
		result.Matched = e.evaluateEndsWith(actualStr, rule.Values)
	case "in":
		result.Matched = e.evaluateIn(actualStr, rule.Values)
	case "not_in":
		result.Matched = !e.evaluateIn(actualStr, rule.Values)
	case "greater_than":
		result.Matched = e.evaluateGreaterThan(actualStr, rule.Values)
	case "greater_than_or_equal":
		result.Matched = e.evaluateGreaterThanOrEqual(actualStr, rule.Values)
	case "less_than":
		result.Matched = e.evaluateLessThan(actualStr, rule.Values)
	case "less_than_or_equal":
		result.Matched = e.evaluateLessThanOrEqual(actualStr, rule.Values)
	case "regex":
		result.Matched = e.evaluateRegex(actualStr, rule.Values)
	case "exists":
		result.Matched = true // If we got here, the attribute exists
	case "not_exists":
		result.Matched = false // If we got here, the attribute exists
	default:
		result.Reason = fmt.Sprintf("Unknown operator: %s", rule.Operator)
		return result
	}

	if result.Matched {
		result.Reason = fmt.Sprintf("Rule matched: %s %s %v", rule.Attribute, rule.Operator, rule.Values)
	} else {
		result.Reason = fmt.Sprintf("Rule did not match: %s (%v) %s %v", rule.Attribute, actualValue, rule.Operator, rule.Values)
	}

	return result
}

// getUserAttribute extracts an attribute value from user context
func (e *Evaluator) getUserAttribute(attribute string, userContext *types.SegmentEvaluationRequest) interface{} {
	switch attribute {
	case "user_id":
		return userContext.UserID
	case "user_key":
		return userContext.UserKey
	case "environment":
		return userContext.Environment
	default:
		// Check in attributes map
		if userContext.Attributes != nil {
			if value, exists := userContext.Attributes[attribute]; exists {
				return value
			}
		}
		return nil
	}
}

// toString converts an interface{} value to string
func (e *Evaluator) toString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// Evaluation methods for different operators

func (e *Evaluator) evaluateEquals(actual string, expected []string) bool {
	for _, exp := range expected {
		if actual == exp {
			return true
		}
	}
	return false
}

func (e *Evaluator) evaluateContains(actual string, expected []string) bool {
	for _, exp := range expected {
		if strings.Contains(actual, exp) {
			return true
		}
	}
	return false
}

func (e *Evaluator) evaluateStartsWith(actual string, expected []string) bool {
	for _, exp := range expected {
		if strings.HasPrefix(actual, exp) {
			return true
		}
	}
	return false
}

func (e *Evaluator) evaluateEndsWith(actual string, expected []string) bool {
	for _, exp := range expected {
		if strings.HasSuffix(actual, exp) {
			return true
		}
	}
	return false
}

func (e *Evaluator) evaluateIn(actual string, expected []string) bool {
	return e.evaluateEquals(actual, expected)
}

func (e *Evaluator) evaluateGreaterThan(actual string, expected []string) bool {
	if len(expected) == 0 {
		return false
	}
	
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected[0], 64)
	
	if err1 != nil || err2 != nil {
		return actual > expected[0] // String comparison as fallback
	}
	
	return actualNum > expectedNum
}

func (e *Evaluator) evaluateGreaterThanOrEqual(actual string, expected []string) bool {
	if len(expected) == 0 {
		return false
	}
	
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected[0], 64)
	
	if err1 != nil || err2 != nil {
		return actual >= expected[0] // String comparison as fallback
	}
	
	return actualNum >= expectedNum
}

func (e *Evaluator) evaluateLessThan(actual string, expected []string) bool {
	if len(expected) == 0 {
		return false
	}
	
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected[0], 64)
	
	if err1 != nil || err2 != nil {
		return actual < expected[0] // String comparison as fallback
	}
	
	return actualNum < expectedNum
}

func (e *Evaluator) evaluateLessThanOrEqual(actual string, expected []string) bool {
	if len(expected) == 0 {
		return false
	}
	
	actualNum, err1 := strconv.ParseFloat(actual, 64)
	expectedNum, err2 := strconv.ParseFloat(expected[0], 64)
	
	if err1 != nil || err2 != nil {
		return actual <= expected[0] // String comparison as fallback
	}
	
	return actualNum <= expectedNum
}

func (e *Evaluator) evaluateRegex(actual string, expected []string) bool {
	for _, pattern := range expected {
		if regex, err := regexp.Compile(pattern); err == nil {
			if regex.MatchString(actual) {
				return true
			}
		}
	}
	return false
}