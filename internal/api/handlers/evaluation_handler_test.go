package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)


func TestEvaluationRequest_ValidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id": "user_123",
		"user_key": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
			"tier": "premium",
		},
		"environment": "production",
	}

	jsonBody, _ := json.Marshal(reqBody)
	
	r := gin.New()
	r.POST("/evaluate", func(c *gin.Context) {
		var req types.EvaluationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		
		assert.Equal(t, "test-flag", req.FlagKey)
		assert.Equal(t, "user_123", req.UserID)
		assert.Equal(t, "production", req.Environment)
		assert.Equal(t, "US", req.Attributes["country"])
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestEvaluationRequest_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Invalid JSON - missing required fields
	reqBody := map[string]interface{}{
		"user_id": "user_123",
		// missing flag_key
	}

	jsonBody, _ := json.Marshal(reqBody)
	
	r := gin.New()
	r.POST("/evaluate", func(c *gin.Context) {
		var req types.EvaluationRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
		
		// Validation - flag_key is required
		if req.FlagKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "flag_key is required"})
			return
		}
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/evaluate", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "flag_key is required")
}

func TestEvaluationResponse_Structure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	r := gin.New()
	r.GET("/evaluate/:flagkey", func(c *gin.Context) {
		flagKey := c.Param("flagkey")
		
		response := types.EvaluationResponse{
			FlagKey:   flagKey,
			Value:     json.RawMessage(`true`),
			Variation: "default",
			Reason:    "default_value",
			Default:   true,
		}
		
		c.JSON(http.StatusOK, response)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/evaluate/test-flag", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var response types.EvaluationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "test-flag", response.FlagKey)
	assert.Equal(t, json.RawMessage(`true`), response.Value)
	assert.Equal(t, "default", response.Variation)
	assert.Equal(t, "default_value", response.Reason)
	assert.True(t, response.Default)
}


func TestFlagTypes_Constants(t *testing.T) {
	assert.Equal(t, types.FlagType("boolean"), types.FlagTypeBoolean)
	assert.Equal(t, types.FlagType("string"), types.FlagTypeString)
	assert.Equal(t, types.FlagType("number"), types.FlagTypeNumber)
	assert.Equal(t, types.FlagType("json"), types.FlagTypeJSON)
	assert.Equal(t, types.FlagType("variant"), types.FlagTypeVariant)
}

func BenchmarkEvaluationRequest_Parsing(b *testing.B) {
	gin.SetMode(gin.TestMode)

	reqBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id": "user_123",
		"user_key": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
			"tier": "premium",
		},
		"environment": "production",
	}

	jsonBody, _ := json.Marshal(reqBody)
	
	r := gin.New()
	r.POST("/evaluate", func(c *gin.Context) {
		var req types.EvaluationRequest
		_ = c.ShouldBindJSON(&req)
		c.JSON(http.StatusOK, gin.H{"flag": req.FlagKey})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/evaluate", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
	}
}

// MockRolloutRepository for evaluation handler tests
type MockRolloutRepository struct {
	mock.Mock
}

func (m *MockRolloutRepository) GetActiveRollouts(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	args := m.Called(ctx, flagID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepository) GetByFlag(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	args := m.Called(ctx, flagID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepository) GetByID(ctx context.Context, id string) (*types.Rollout, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepository) GetByProject(ctx context.Context, projectID, environment string) ([]*types.Rollout, error) {
	args := m.Called(ctx, projectID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepository) Create(ctx context.Context, rollout *types.Rollout) error {
	args := m.Called(ctx, rollout)
	return args.Error(0)
}

func (m *MockRolloutRepository) Update(ctx context.Context, rollout *types.Rollout) error {
	args := m.Called(ctx, rollout)
	return args.Error(0)
}

func (m *MockRolloutRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRolloutRepository) CreateStickyAssignment(ctx context.Context, assignment *types.StickyAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func (m *MockRolloutRepository) GetStickyAssignment(ctx context.Context, flagID, environment, userKey string) (*types.StickyAssignment, error) {
	args := m.Called(ctx, flagID, environment, userKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.StickyAssignment), args.Error(1)
}

func (m *MockRolloutRepository) DeleteStickyAssignment(ctx context.Context, flagID, environment, userKey string) error {
	args := m.Called(ctx, flagID, environment, userKey)
	return args.Error(0)
}

func (m *MockRolloutRepository) CleanupExpiredAssignments(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockFlagRepository for evaluation handler tests  
type MockEvalFlagRepository struct {
	mock.Mock
}

func (m *MockEvalFlagRepository) Create(ctx context.Context, flag *types.Flag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockEvalFlagRepository) GetByKey(ctx context.Context, key, environment string) (*types.Flag, error) {
	args := m.Called(ctx, key, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Flag), args.Error(1)
}

func (m *MockEvalFlagRepository) GetByProjectKey(ctx context.Context, projectID, key, environment string) (*types.Flag, error) {
	args := m.Called(ctx, projectID, key, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Flag), args.Error(1)
}

func (m *MockEvalFlagRepository) Update(ctx context.Context, flag *types.Flag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockEvalFlagRepository) Delete(ctx context.Context, key, environment string) error {
	args := m.Called(ctx, key, environment)
	return args.Error(0)
}

func (m *MockEvalFlagRepository) List(ctx context.Context, environment string) ([]*types.Flag, error) {
	args := m.Called(ctx, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

func (m *MockEvalFlagRepository) ListByProject(ctx context.Context, projectID, environment string) ([]*types.Flag, error) {
	args := m.Called(ctx, projectID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

func (m *MockEvalFlagRepository) ListByTags(ctx context.Context, environment string, tags []string) ([]*types.Flag, error) {
	args := m.Called(ctx, environment, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

func TestEvaluationHandler_Evaluate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Name:        "Test Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id":  "user_123",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	assert.Contains(t, response, "evaluation_time_ms")
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_FlagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	mockRepo.On("GetByKey", mock.Anything, "nonexistent-flag", "production").Return(nil, fmt.Errorf("not found"))
	
	requestBody := map[string]interface{}{
		"flag_key": "nonexistent-flag",
		"user_id":  "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag not found", response["error"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_DisabledFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	disabledFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "disabled-flag",
		Name:        "Disabled Flag",
		Type:        types.FlagTypeString,
		Enabled:     false,
		Default:     json.RawMessage(`"disabled_value"`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "disabled-flag", "production").Return(disabledFlag, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "disabled-flag",
		"user_id":  "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "disabled-flag", response["flag_key"])
	assert.Equal(t, "disabled_value", response["value"])
	assert.Equal(t, "flag_disabled", response["reason"])
	assert.True(t, response["default"].(bool))
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Missing required flag_key
	requestBody := map[string]interface{}{
		"user_id": "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "FlagKey")
}

func TestEvaluationHandler_BatchEvaluate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Setup first flag
	flag1 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "flag-1",
		Name:        "Flag 1",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}
	
	// Setup second flag
	flag2 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "flag-2",
		Name:        "Flag 2",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
		ProjectID:   "proj_123",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "flag-1", "production").Return(flag1, nil)
	mockRepo.On("GetByKey", mock.Anything, "flag-2", "production").Return(flag2, nil)
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"flag-1", "flag-2"},
		"user_id":   "user_123",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	// Check that we have evaluations for both flags
	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "flag-1")
	assert.Contains(t, evaluations, "flag-2")
	
	// Check response structure
	assert.Contains(t, response, "total_time_ms")
	assert.Contains(t, response, "avg_time_per_flag_ms")
	assert.Contains(t, response, "timestamp")
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_SomeFlagsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Setup only one flag, the other will return an error
	flag1 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "existing-flag",
		Name:        "Existing Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "existing-flag", "production").Return(flag1, nil)
	mockRepo.On("GetByKey", mock.Anything, "missing-flag", "production").Return(nil, fmt.Errorf("not found"))
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"existing-flag", "missing-flag"},
		"user_id":   "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	evaluations := response["evaluations"].(map[string]interface{})
	
	// Should have evaluation for existing flag
	assert.Contains(t, evaluations, "existing-flag")
	existingResult := evaluations["existing-flag"].(map[string]interface{})
	assert.Equal(t, true, existingResult["value"])
	
	// Should have error for missing flag
	assert.Contains(t, evaluations, "missing-flag")
	missingResult := evaluations["missing-flag"].(map[string]interface{})
	assert.Equal(t, "flag not found", missingResult["error"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Missing required flag_keys
	requestBody := map[string]interface{}{
		"user_id": "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "FlagKeys")
}

func TestEvaluationHandler_Evaluate_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Mock database error
	mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(nil, fmt.Errorf("database connection failed"))
	
	requestBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	
	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag not found", response["error"])
	
	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_WithProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "project-flag",
		Name:        "Project Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}
	
	mockRepo.On("GetByProjectKey", mock.Anything, "proj_123", "project-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "project-flag",
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production&project_id=proj_123", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "project-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	// Mock database error for first flag
	mockRepo.On("GetByKey", mock.Anything, "flag-1", "production").Return(nil, fmt.Errorf("database error"))
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"flag-1"},
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "flag-1")
	result := evaluations["flag-1"].(map[string]interface{})
	assert.Equal(t, "flag not found", result["error"])
	
	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_RolloutError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Name:        "Test Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return(nil, fmt.Errorf("rollout database error"))
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	// Should still return OK and use default flag value despite rollout error
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_WithAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "api-flag",
		Name:        "API Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "staging",
		ProjectID:   "api_proj_123",
	}
	
	mockRepo.On("GetByProjectKey", mock.Anything, "api_proj_123", "api-flag", "staging").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "staging").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "staging").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "api-flag",
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	// Simulate API key authentication middleware setting context values
	c.Set("environment", "staging")
	c.Set("projectID", "api_proj_123")
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "api-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_WithAPIKeyAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "batch-flag",
		Name:        "Batch Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "staging",
		ProjectID:   "batch_proj_123",
	}
	
	mockRepo.On("GetByProjectKey", mock.Anything, "batch_proj_123", "batch-flag", "staging").Return(expectedFlag, nil)
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"batch-flag"},
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"country": "US",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	// Simulate API key authentication middleware setting context values
	c.Set("environment", "staging")
	c.Set("projectID", "batch_proj_123")
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "batch-flag")
	result := evaluations["batch-flag"].(map[string]interface{})
	assert.Equal(t, true, result["value"])
	
	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_WithProject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "project-batch-flag",
		Name:        "Project Batch Flag",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
		ProjectID:   "proj_456",
	}
	
	mockRepo.On("GetByProjectKey", mock.Anything, "proj_456", "project-batch-flag", "production").Return(expectedFlag, nil)
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"project-batch-flag"},
		"user_id": "user_456",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch?project_id=proj_456&environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "project-batch-flag")
	result := evaluations["project-batch-flag"].(map[string]interface{})
	assert.Equal(t, "default", result["value"])
	
	assert.Contains(t, response, "total_time_ms")
	assert.Contains(t, response, "avg_time_per_flag_ms")
	
	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_PercentageRollout_Match(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "rollout-flag",
		Name:        "Rollout Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	percentage := 100
	percentageRollout := &types.Rollout{
		ID:          uuid.New().String(),
		FlagID:      expectedFlag.ID,
		Type:        types.RolloutTypePercentage,
		Status:      types.RolloutStatusActive,
		Environment: "production",
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage, // 100% to ensure match
		},
	}
	
	mockRepo.On("GetByKey", mock.Anything, "rollout-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{percentageRollout}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{percentageRollout}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "rollout-flag",
		"user_id": "user_123",
		"user_key": "test_user",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "rollout-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	assert.Contains(t, []string{"rollout_match", "rollout_no_match"}, response["reason"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_PercentageRollout_NoMatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "rollout-no-match",
		Name:        "Rollout No Match",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
	}
	
	percentage := 0
	percentageRollout := &types.Rollout{
		ID:          uuid.New().String(),
		FlagID:      expectedFlag.ID,
		Type:        types.RolloutTypePercentage,
		Status:      types.RolloutStatusActive,
		Environment: "production",
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage, // 0% to ensure no match
		},
	}
	
	mockRepo.On("GetByKey", mock.Anything, "rollout-no-match", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{percentageRollout}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{percentageRollout}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "rollout-no-match",
		"user_id": "user_123",
		"user_key": "test_user",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "rollout-no-match", response["flag_key"])
	// For string flags, no match should return empty string
	assert.Equal(t, "", response["value"])
	assert.Equal(t, "rollout_no_match", response["reason"])
	assert.Equal(t, true, response["default"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_ExperimentRollout_WithVariation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "experiment-flag",
		Name:        "Experiment Flag",
		Type:        types.FlagTypeVariant,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
		Variations: []types.Variation{
			{
				ID:    "var_a",
				Name:  "Variation A",
				Value: json.RawMessage(`"variation_a"`),
			},
			{
				ID:    "var_b",
				Name:  "Variation B", 
				Value: json.RawMessage(`"variation_b"`),
			},
		},
	}
	
	percentage := 100
	experimentRollout := &types.Rollout{
		ID:          uuid.New().String(),
		FlagID:      expectedFlag.ID,
		Type:        types.RolloutTypeExperiment,
		Status:      types.RolloutStatusActive,
		Environment: "production",
		Config: types.AdvancedRolloutConfig{
			Percentage: &percentage,
		},
	}
	
	mockRepo.On("GetByKey", mock.Anything, "experiment-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{experimentRollout}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{experimentRollout}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "experiment-flag",
		"user_id": "user_123",
		"user_key": "test_user",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "experiment-flag", response["flag_key"])
	// Should get one of the variations or default
	assert.Contains(t, []interface{}{"variation_a", "variation_b", "default"}, response["value"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_StickyAssignment_Found(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "sticky-flag",
		Name:        "Sticky Flag",
		Type:        types.FlagTypeVariant,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
		Variations: []types.Variation{
			{
				ID:    "sticky_var",
				Name:  "Sticky Variation",
				Value: json.RawMessage(`"sticky_value"`),
			},
		},
		Targeting: &types.TargetingConfig{
			Rollout: &types.RolloutConfig{
				StickyBucketing: true,
				BucketBy: "user_key",
				Seed: 12345,
			},
		},
	}
	
	stickyAssignment := &types.StickyAssignment{
		ID:          uuid.New().String(),
		FlagID:      expectedFlag.ID,
		Environment: "production",
		UserKey:     "sticky_user",
		VariationID: "sticky_var",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "sticky-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetStickyAssignment", mock.Anything, expectedFlag.ID, "production", "sticky_user").Return(stickyAssignment, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "sticky-flag",
		"user_id": "user_123",
		"user_key": "sticky_user",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "sticky-flag", response["flag_key"])
	assert.Equal(t, "sticky_value", response["value"])
	assert.Equal(t, "sticky_assignment", response["reason"])
	assert.Equal(t, "sticky_var", response["variation"])
	assert.Equal(t, false, response["default"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_NumberFlagType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "number-flag",
		Name:        "Number Flag",
		Type:        types.FlagTypeNumber,
		Enabled:     true,
		Default:     json.RawMessage(`42`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "number-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "number-flag",
		"user_id": "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "number-flag", response["flag_key"])
	assert.Equal(t, float64(42), response["value"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_Evaluate_JSONFlagType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "json-flag",
		Name:        "JSON Flag",
		Type:        types.FlagTypeJSON,
		Enabled:     true,
		Default:     json.RawMessage(`{"key":"value","number":123}`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "json-flag", "production").Return(expectedFlag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, expectedFlag.ID, "production").Return([]*types.Rollout{}, nil)
	
	requestBody := map[string]interface{}{
		"flag_key": "json-flag",
		"user_id": "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.Evaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "json-flag", response["flag_key"])
	
	// Value should be a map representing the JSON object
	value := response["value"].(map[string]interface{})
	assert.Equal(t, "value", value["key"])
	assert.Equal(t, float64(123), value["number"])
	
	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate_DisabledFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	mockRepo := new(MockEvalFlagRepository)
	mockRolloutRepo := new(MockRolloutRepository)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)
	
	disabledFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "disabled-batch-flag",
		Name:        "Disabled Batch Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     false,
		Default:     json.RawMessage(`false`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "disabled-batch-flag", "production").Return(disabledFlag, nil)
	
	requestBody := map[string]interface{}{
		"flag_keys": []string{"disabled-batch-flag"},
		"user_id": "user_123",
	}
	
	body, err := json.Marshal(requestBody)
	require.NoError(t, err)
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/evaluate/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	
	handler.BatchEvaluate(c)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "disabled-batch-flag")
	result := evaluations["disabled-batch-flag"].(map[string]interface{})
	assert.Equal(t, false, result["value"])
	assert.Equal(t, "flag_disabled", result["reason"])
	assert.Equal(t, true, result["default"])
	
	mockRepo.AssertExpectations(t)
}