package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
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

// MockRolloutRepositorySimple mocks the rollout repository for simple tests
type MockRolloutRepositorySimple struct {
	mock.Mock
}

func (m *MockRolloutRepositorySimple) GetActiveRollouts(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	args := m.Called(ctx, flagID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepositorySimple) GetByFlag(ctx context.Context, flagID, environment string) ([]*types.Rollout, error) {
	args := m.Called(ctx, flagID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Rollout), args.Error(1)
}

func (m *MockRolloutRepositorySimple) GetStickyAssignment(ctx context.Context, flagID, environment, userKey string) (*types.StickyAssignment, error) {
	args := m.Called(ctx, flagID, environment, userKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.StickyAssignment), args.Error(1)
}

func (m *MockRolloutRepositorySimple) CreateStickyAssignment(ctx context.Context, assignment *types.StickyAssignment) error {
	args := m.Called(ctx, assignment)
	return args.Error(0)
}

func TestEvaluationHandler_EvaluateSimple(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	mockRolloutRepo := new(MockRolloutRepositorySimple)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)

	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "test-flag",
		Name:        "Test Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}

	mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(flag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, flag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, flag.ID, "production").Return([]*types.Rollout{}, nil)

	requestBody := map[string]interface{}{
		"flag_key": "test-flag",
		"user_id":  "user123",
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

	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_EvaluateString(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	mockRolloutRepo := new(MockRolloutRepositorySimple)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)

	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "theme-flag",
		Name:        "Theme Flag",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"dark"`),
		Environment: "production",
		ProjectID:   "proj_123",
	}

	mockRepo.On("GetByKey", mock.Anything, "theme-flag", "production").Return(flag, nil)
	mockRolloutRepo.On("GetActiveRollouts", mock.Anything, flag.ID, "production").Return([]*types.Rollout{}, nil)
	mockRolloutRepo.On("GetByFlag", mock.Anything, flag.ID, "production").Return([]*types.Rollout{}, nil)

	requestBody := map[string]interface{}{
		"flag_key": "theme-flag",
		"user_id":  "user123",
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
	assert.Equal(t, "theme-flag", response["flag_key"])
	assert.Equal(t, "dark", response["value"])

	mockRepo.AssertExpectations(t)
	mockRolloutRepo.AssertExpectations(t)
}

func TestEvaluationHandler_EvaluateDisabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	mockRolloutRepo := new(MockRolloutRepositorySimple)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)

	flag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "disabled-flag",
		Name:        "Disabled Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     false, // Flag is disabled
		Default:     json.RawMessage(`false`),
		Environment: "production",
		ProjectID:   "proj_123",
	}

	mockRepo.On("GetByKey", mock.Anything, "disabled-flag", "production").Return(flag, nil)

	requestBody := map[string]interface{}{
		"flag_key": "disabled-flag",
		"user_id":  "user123",
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
	assert.Equal(t, false, response["value"])
	assert.Equal(t, true, response["default"])
	assert.Equal(t, "flag_disabled", response["reason"])

	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_EvaluateNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	mockRolloutRepo := new(MockRolloutRepositorySimple)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)

	mockRepo.On("GetByKey", mock.Anything, "non-existent", "production").Return(nil, assert.AnError)

	requestBody := map[string]interface{}{
		"flag_key": "non-existent",
		"user_id":  "user123",
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

	mockRepo.AssertExpectations(t)
}

func TestEvaluationHandler_BatchEvaluate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	mockRolloutRepo := new(MockRolloutRepositorySimple)
	handler := handlers.NewEvaluationHandler(mockRepo, mockRolloutRepo)

	flag1 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "flag1",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		ProjectID:   "proj_123",
	}

	flag2 := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "flag2",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"value2"`),
		Environment: "production",
		ProjectID:   "proj_123",
	}

	mockRepo.On("GetByKey", mock.Anything, "flag1", "production").Return(flag1, nil)
	mockRepo.On("GetByKey", mock.Anything, "flag2", "production").Return(flag2, nil)

	requestBody := map[string]interface{}{
		"flag_keys": []string{"flag1", "flag2"},
		"user_id":   "batch-user",
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
	assert.Len(t, evaluations, 2)

	flag1Result := evaluations["flag1"].(map[string]interface{})
	assert.Equal(t, true, flag1Result["value"])

	flag2Result := evaluations["flag2"].(map[string]interface{})
	assert.Equal(t, "value2", flag2Result["value"])

	mockRepo.AssertExpectations(t)
}