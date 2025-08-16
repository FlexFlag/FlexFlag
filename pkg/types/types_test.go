package types

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlagType_Constants(t *testing.T) {
	assert.Equal(t, FlagType("boolean"), FlagTypeBoolean)
	assert.Equal(t, FlagType("string"), FlagTypeString)
	assert.Equal(t, FlagType("number"), FlagTypeNumber)
	assert.Equal(t, FlagType("json"), FlagTypeJSON)
	assert.Equal(t, FlagType("variant"), FlagTypeVariant)
}

func TestFlag_JSONMarshaling(t *testing.T) {
	flag := Flag{
		ID:          "flag_123",
		ProjectID:   "proj_456",
		Key:         "test-flag",
		Name:        "Test Flag",
		Description: "A test flag for unit testing",
		Type:        FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		CreatedAt:   time.Now().Truncate(time.Second),
		UpdatedAt:   time.Now().Truncate(time.Second),
		Tags:        []string{"testing", "api"},
		Metadata:    map[string]interface{}{"team": "backend"},
	}

	// Marshal to JSON
	data, err := json.Marshal(flag)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal from JSON
	var unmarshaled Flag
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, flag.ID, unmarshaled.ID)
	assert.Equal(t, flag.Key, unmarshaled.Key)
	assert.Equal(t, flag.Type, unmarshaled.Type)
	assert.Equal(t, flag.Enabled, unmarshaled.Enabled)
	assert.Equal(t, flag.Tags, unmarshaled.Tags)
}

func TestVariation_Structure(t *testing.T) {
	variation := Variation{
		ID:          "var_123",
		Name:        "Control",
		Description: "Control group for A/B test",
		Value:       json.RawMessage(`{"enabled": false}`),
		Weight:      50,
	}

	assert.Equal(t, "var_123", variation.ID)
	assert.Equal(t, "Control", variation.Name)
	assert.Equal(t, "Control group for A/B test", variation.Description)
	assert.Equal(t, 50, variation.Weight)
}

func TestTargetingRule_Structure(t *testing.T) {
	rule := TargetingRule{
		ID:          "rule_123",
		Attribute:   "country",
		Operator:    "eq",
		Values:      []string{"US", "CA"},
		Variation:   "var_123",
		Description: "North America targeting",
	}

	assert.Equal(t, "rule_123", rule.ID)
	assert.Equal(t, "country", rule.Attribute)
	assert.Equal(t, "eq", rule.Operator)
	assert.Equal(t, []string{"US", "CA"}, rule.Values)
	assert.Equal(t, "var_123", rule.Variation)
	assert.Equal(t, "North America targeting", rule.Description)
}

func TestEvaluationRequest_JSONMarshaling(t *testing.T) {
	req := EvaluationRequest{
		ProjectKey: "project-key",
		FlagKey:    "test-flag",
		UserID:     "user_123",
		UserKey:    "user_key_123",
		Attributes: map[string]interface{}{
			"country": "US",
			"tier":    "premium",
			"age":     25,
		},
		Environment: "production",
	}

	// Marshal to JSON
	data, err := json.Marshal(req)
	require.NoError(t, err)
	assert.Contains(t, string(data), "project-key")
	assert.Contains(t, string(data), "test-flag")
	assert.Contains(t, string(data), "user_123")

	// Unmarshal from JSON
	var unmarshaled EvaluationRequest
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, req.ProjectKey, unmarshaled.ProjectKey)
	assert.Equal(t, req.FlagKey, unmarshaled.FlagKey)
	assert.Equal(t, req.UserID, unmarshaled.UserID)
	// JSON unmarshaling converts numbers to float64
	assert.Equal(t, "US", unmarshaled.Attributes["country"])
	assert.Equal(t, "premium", unmarshaled.Attributes["tier"])
	assert.Equal(t, float64(25), unmarshaled.Attributes["age"])
}

func TestEvaluationResponse_Structure(t *testing.T) {
	timestamp := time.Now()
	resp := EvaluationResponse{
		FlagKey:        "test-flag",
		Value:          json.RawMessage(`true`),
		Variation:      "default",
		Reason:         "default_value",
		RuleID:         "rule_123",
		Default:        true,
		Timestamp:      timestamp,
		EvaluationTime: 1.5,
		Source:         "edge-cache",
	}

	assert.Equal(t, "test-flag", resp.FlagKey)
	assert.Equal(t, json.RawMessage(`true`), resp.Value)
	assert.Equal(t, "default", resp.Variation)
	assert.Equal(t, "default_value", resp.Reason)
	assert.Equal(t, "rule_123", resp.RuleID)
	assert.True(t, resp.Default)
	assert.Equal(t, timestamp, resp.Timestamp)
	assert.Equal(t, 1.5, resp.EvaluationTime)
	assert.Equal(t, "edge-cache", resp.Source)
}

func TestUser_Structure(t *testing.T) {
	now := time.Now()
	user := User{
		ID:           "user_123",
		Email:        "test@example.com",
		PasswordHash: "hashed_password",
		FullName:     "Test User",
		Role:         UserRoleAdmin,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	assert.Equal(t, "user_123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.FullName)
	assert.Equal(t, UserRoleAdmin, user.Role)
	assert.True(t, user.IsActive)
}

func TestUserRole_Constants(t *testing.T) {
	assert.Equal(t, UserRole("admin"), UserRoleAdmin)
	assert.Equal(t, UserRole("editor"), UserRoleEditor)
	assert.Equal(t, UserRole("viewer"), UserRoleViewer)
}

func TestProject_Structure(t *testing.T) {
	now := time.Now()
	project := Project{
		ID:          "proj_123",
		Name:        "Test Project",
		Description: "A project for testing",
		Slug:        "test-project",
		IsActive:    true,
		Settings:    map[string]interface{}{"theme": "dark"},
		CreatedBy:   "user_123",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "proj_123", project.ID)
	assert.Equal(t, "Test Project", project.Name)
	assert.Equal(t, "test-project", project.Slug)
	assert.True(t, project.IsActive)
	assert.Equal(t, "user_123", project.CreatedBy)
}

func TestEnvironment_Structure(t *testing.T) {
	now := time.Now()
	env := Environment{
		ID:          "env_123",
		ProjectID:   "proj_123",
		Name:        "Production",
		Key:         "production",
		Description: "Production environment",
		IsActive:    true,
		SortOrder:   1,
		Settings:    map[string]interface{}{"strict": true},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "env_123", env.ID)
	assert.Equal(t, "proj_123", env.ProjectID)
	assert.Equal(t, "Production", env.Name)
	assert.Equal(t, "production", env.Key)
	assert.True(t, env.IsActive)
	assert.Equal(t, 1, env.SortOrder)
}

func TestApiKey_Structure(t *testing.T) {
	now := time.Now()
	expiresAt := now.Add(24 * time.Hour)
	lastUsedAt := now.Add(-time.Hour)

	apiKey := ApiKey{
		ID:            "key_123",
		ProjectID:     "proj_123",
		EnvironmentID: "env_123",
		Name:          "Test API Key",
		KeyHash:       "hashed_key",
		KeyPrefix:     "ff_prod_1234****",
		FullKey:       "ff_prod_1234567890abcdef",
		Permissions:   []string{"read", "write"},
		CreatedBy:     "user_123",
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     &expiresAt,
		LastUsedAt:    &lastUsedAt,
		IsActive:      true,
	}

	assert.Equal(t, "key_123", apiKey.ID)
	assert.Equal(t, "Test API Key", apiKey.Name)
	assert.Equal(t, "ff_prod_1234****", apiKey.KeyPrefix)
	assert.Equal(t, []string{"read", "write"}, apiKey.Permissions)
	assert.True(t, apiKey.IsActive)
	assert.NotNil(t, apiKey.ExpiresAt)
	assert.NotNil(t, apiKey.LastUsedAt)
}

func TestAuditLog_Structure(t *testing.T) {
	now := time.Now()
	projectID := "proj_123"
	userID := "user_123"

	auditLog := AuditLog{
		ID:           "audit_123",
		ProjectID:    &projectID,
		UserID:       &userID,
		ResourceType: "flag",
		ResourceID:   "flag_123",
		Action:       "create",
		OldValues:    json.RawMessage(`{}`),
		NewValues:    json.RawMessage(`{"enabled": true}`),
		Metadata:     map[string]interface{}{"source": "api"},
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		CreatedAt:    now,
	}

	assert.Equal(t, "audit_123", auditLog.ID)
	assert.Equal(t, &projectID, auditLog.ProjectID)
	assert.Equal(t, &userID, auditLog.UserID)
	assert.Equal(t, "flag", auditLog.ResourceType)
	assert.Equal(t, "create", auditLog.Action)
	assert.Equal(t, "192.168.1.1", auditLog.IPAddress)
}

func TestCreateUserRequest_Structure(t *testing.T) {
	req := CreateUserRequest{
		Email:    "new@example.com",
		Password: "SecureP@ssw0rd123",
		FullName: "New User",
		Role:     UserRoleEditor,
	}

	assert.Equal(t, "new@example.com", req.Email)
	assert.Equal(t, "SecureP@ssw0rd123", req.Password)
	assert.Equal(t, "New User", req.FullName)
	assert.Equal(t, UserRoleEditor, req.Role)
}

func TestCreateProjectRequest_Structure(t *testing.T) {
	req := CreateProjectRequest{
		Name:        "New Project",
		Description: "A new test project",
		Slug:        "new-project",
		Settings:    map[string]interface{}{"enabled": true},
	}

	assert.Equal(t, "New Project", req.Name)
	assert.Equal(t, "A new test project", req.Description)
	assert.Equal(t, "new-project", req.Slug)
	assert.Equal(t, map[string]interface{}{"enabled": true}, req.Settings)
}

func TestSegment_Structure(t *testing.T) {
	now := time.Now()
	segment := Segment{
		ID:          "seg_123",
		ProjectID:   "proj_123",
		Key:         "premium-users",
		Name:        "Premium Users",
		Description: "Users with premium subscription",
		Rules: []TargetingRule{
			{
				ID:        "rule_1",
				Attribute: "tier",
				Operator:  "eq",
				Values:    []string{"premium"},
			},
		},
		Environment: "production",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "seg_123", segment.ID)
	assert.Equal(t, "premium-users", segment.Key)
	assert.Equal(t, "Premium Users", segment.Name)
	assert.Len(t, segment.Rules, 1)
	assert.Equal(t, "tier", segment.Rules[0].Attribute)
}

func TestSegmentMatchResult_Structure(t *testing.T) {
	result := SegmentMatchResult{
		Matched:    true,
		SegmentKey: "premium-users",
		UserKey:    "user_123",
		Reason:     "rule_matched",
		RuleResults: []RuleEvaluationResult{
			{
				RuleID:         "rule_1",
				Matched:        true,
				Attribute:      "tier",
				Operator:       "eq",
				ActualValue:    "premium",
				ExpectedValues: []interface{}{"premium"},
				Reason:         "exact_match",
			},
		},
	}

	assert.True(t, result.Matched)
	assert.Equal(t, "premium-users", result.SegmentKey)
	assert.Equal(t, "user_123", result.UserKey)
	assert.Len(t, result.RuleResults, 1)
	assert.True(t, result.RuleResults[0].Matched)
}

func TestRoleConstants(t *testing.T) {
	assert.Equal(t, "admin", ProjectRoleAdmin)
	assert.Equal(t, "editor", ProjectRoleEditor)
	assert.Equal(t, "viewer", ProjectRoleViewer)
}

func TestTargetingConfig_Structure(t *testing.T) {
	rollout := &RolloutConfig{
		Type: "percentage",
		Variations: []VariationRollout{
			{VariationID: "var_1", Weight: 50},
			{VariationID: "var_2", Weight: 50},
		},
		BucketBy:        "user_id",
		Seed:            12345,
		StickyBucketing: true,
	}

	config := TargetingConfig{
		Rules: []TargetingRule{
			{
				ID:        "rule_1",
				Attribute: "country",
				Operator:  "in",
				Values:    []string{"US", "CA"},
			},
		},
		Rollout:  rollout,
		Segments: []string{"premium-users", "beta-testers"},
	}

	assert.Len(t, config.Rules, 1)
	assert.NotNil(t, config.Rollout)
	assert.Equal(t, "percentage", config.Rollout.Type)
	assert.Len(t, config.Segments, 2)
}

func TestRolloutConfig_Structure(t *testing.T) {
	rollout := RolloutConfig{
		Type: "percentage",
		Variations: []VariationRollout{
			{VariationID: "control", Weight: 30},
			{VariationID: "treatment", Weight: 70},
		},
		BucketBy:        "user_key",
		Seed:            98765,
		StickyBucketing: false,
	}

	assert.Equal(t, "percentage", rollout.Type)
	assert.Len(t, rollout.Variations, 2)
	assert.Equal(t, "user_key", rollout.BucketBy)
	assert.Equal(t, int64(98765), rollout.Seed)
	assert.False(t, rollout.StickyBucketing)
}

func TestLoginRequest_Structure(t *testing.T) {
	req := LoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	}

	assert.Equal(t, "user@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
}

func TestLoginResponse_Structure(t *testing.T) {
	user := User{
		ID:       "user_123",
		Email:    "user@example.com",
		FullName: "Test User",
		Role:     UserRoleAdmin,
	}

	resp := LoginResponse{
		Token: "jwt_token_here",
		User:  user,
	}

	assert.Equal(t, "jwt_token_here", resp.Token)
	assert.Equal(t, "user_123", resp.User.ID)
	assert.Equal(t, "user@example.com", resp.User.Email)
}

func BenchmarkFlag_JSONMarshal(b *testing.B) {
	flag := Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		Type:        FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		Tags:        []string{"test"},
		Metadata:    map[string]interface{}{"test": true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(flag)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEvaluationRequest_JSONMarshal(b *testing.B) {
	req := EvaluationRequest{
		FlagKey:    "test-flag",
		UserID:     "user_123",
		Attributes: map[string]interface{}{"country": "US"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}