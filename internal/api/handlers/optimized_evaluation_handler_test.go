package handlers_test

import (
	"bytes"
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

func TestOptimizedEvaluationHandler_FastEvaluate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

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
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "test-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	assert.Contains(t, response, "evaluation_time_ms")
	assert.Contains(t, response, "reason")

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastEvaluate_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	// Missing required flag_key
	requestBody := map[string]interface{}{
		"user_id": "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastEvaluate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "FlagKey")
}

func TestOptimizedEvaluationHandler_FastEvaluate_FlagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	mockRepo.On("GetByKey", mock.Anything, "nonexistent-flag", "production").Return(nil, fmt.Errorf("flag not found"))

	requestBody := map[string]interface{}{
		"flag_key": "nonexistent-flag",
		"user_id":  "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastEvaluate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag not found", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastBatchEvaluate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	// Setup flags
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
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastBatchEvaluate(c)

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
	assert.Contains(t, response, "cache_enabled")

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastBatchEvaluate_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	// Missing required flag_keys
	requestBody := map[string]interface{}{
		"user_id": "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast/batch", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastBatchEvaluate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "FlagKeys")
}

func TestOptimizedEvaluationHandler_GetCacheStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/evaluate/fast/stats", nil)
	c.Request = req

	handler.GetCacheStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	// Check that cache_stats object exists and has expected fields
	cacheStats := response["cache_stats"].(map[string]interface{})
	assert.Contains(t, cacheStats, "cached_flags")
	assert.Contains(t, cacheStats, "ttl_seconds")
}

func TestOptimizedEvaluationHandler_ClearCache_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast/cache/clear", nil)
	c.Request = req

	handler.ClearCache(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "cache cleared", response["message"])
}

func TestOptimizedEvaluationHandler_FastEvaluate_WithProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "project-flag",
		Name:        "Project Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "staging",
		ProjectID:   "proj_456",
	}

	mockRepo.On("GetByProjectKey", mock.Anything, "proj_456", "project-flag", "staging").Return(expectedFlag, nil)

	requestBody := map[string]interface{}{
		"flag_key": "project-flag",
		"user_id":  "user_123",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast?environment=staging&project_id=proj_456", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "project-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastEvaluate_DisabledFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "disabled-flag",
		Name:        "Disabled Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     false,
		Default:     json.RawMessage(`false`),
		Environment: "production",
	}

	mockRepo.On("GetByKey", mock.Anything, "disabled-flag", "production").Return(expectedFlag, nil)

	requestBody := map[string]interface{}{
		"flag_key": "disabled-flag",
		"user_id":  "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "disabled-flag", response["flag_key"])
	assert.Equal(t, false, response["value"])
	assert.Equal(t, "flag_disabled", response["reason"])
	assert.Equal(t, true, response["default"])

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastBatchEvaluate_WithProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "batch-project-flag",
		Name:        "Batch Project Flag",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "staging",
		ProjectID:   "batch_proj_789",
	}

	mockRepo.On("GetByProjectKey", mock.Anything, "batch_proj_789", "batch-project-flag", "staging").Return(expectedFlag, nil)

	requestBody := map[string]interface{}{
		"flag_keys": []string{"batch-project-flag"},
		"user_id":   "user_456",
		"attributes": map[string]interface{}{
			"region": "US",
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast/batch?environment=staging&project_id=batch_proj_789", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastBatchEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	evaluations := response["evaluations"].(map[string]interface{})
	assert.Contains(t, evaluations, "batch-project-flag")
	result := evaluations["batch-project-flag"].(map[string]interface{})
	assert.Equal(t, "default", result["value"])

	mockRepo.AssertExpectations(t)
}

func TestOptimizedEvaluationHandler_FastBatchEvaluate_DisabledFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	handler := handlers.NewOptimizedEvaluationHandler(mockRepo)

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
		"user_id":   "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/fast/batch?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.FastBatchEvaluate(c)

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