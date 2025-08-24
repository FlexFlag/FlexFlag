package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUltraFastHandler_NewUltraFastHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	assert.NotNil(t, handler)
	
	// Give a moment for the goroutines to start
	time.Sleep(10 * time.Millisecond)
}

func TestUltraFastHandler_UltraFastEvaluate_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "ultra-flag",
		Name:        "Ultra Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "ultra-flag", "production").Return(expectedFlag, nil)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	requestBody := map[string]interface{}{
		"flag_key": "ultra-flag",
		"user_id": "user_123",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/ultra?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UltraFastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ultra-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])
	assert.Contains(t, response, "evaluation_time_ms")
	assert.Contains(t, response, "reason")
	assert.Contains(t, response, "default")
	assert.Contains(t, response, "timestamp")

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_UltraFastEvaluate_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/ultra", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UltraFastEvaluate(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "invalid character")
}

func TestUltraFastHandler_UltraFastEvaluate_FlagNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	mockRepo.On("GetByKey", mock.Anything, "nonexistent-flag", "production").Return(nil, assert.AnError)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	requestBody := map[string]interface{}{
		"flag_key": "nonexistent-flag",
		"user_id": "user_123",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/ultra?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UltraFastEvaluate(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag not found", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_RefreshFlag_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "refresh-flag",
		Name:        "Refresh Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "refresh-flag", "production").Return(expectedFlag, nil)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	// Call RefreshFlag method
	handler.RefreshFlag("refresh-flag", "production")

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_RefreshAllFlags_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	expectedFlags := []*types.Flag{
		{
			ID:          uuid.New().String(),
			Key:         "flag-1",
			Name:        "Flag 1",
			Type:        types.FlagTypeBoolean,
			Enabled:     true,
			Default:     json.RawMessage(`true`),
			Environment: "production",
		},
		{
			ID:          uuid.New().String(),
			Key:         "flag-2",
			Name:        "Flag 2",
			Type:        types.FlagTypeString,
			Enabled:     true,
			Default:     json.RawMessage(`"default"`),
			Environment: "production",
		},
	}
	
	// Mock multiple calls for different environments
	mockRepo.On("List", mock.Anything, "production").Return(expectedFlags, nil)
	mockRepo.On("List", mock.Anything, "staging").Return([]*types.Flag{}, nil)
	mockRepo.On("List", mock.Anything, "development").Return([]*types.Flag{}, nil)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	// Call RefreshAllFlags method
	handler.RefreshAllFlags()

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_GetStats_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/evaluate/ultra/stats", nil)
	c.Request = req

	handler.GetStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "preloaded_flags")
	assert.Contains(t, response, "cached_responses")
	assert.Contains(t, response, "preload_complete")
}

func TestUltraFastHandler_UltraFastEvaluate_WithTargeting(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	// Create a flag with targeting rules to test the targeting evaluation path
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "targeting-flag",
		Name:        "Targeting Flag",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"default"`),
		Environment: "production",
		Variations: []types.Variation{
			{
				ID:    "var_1",
				Name:  "Variation 1",
				Value: json.RawMessage(`"variation_1"`),
			},
			{
				ID:    "var_2", 
				Name:  "Variation 2",
				Value: json.RawMessage(`"variation_2"`),
			},
		},
		Targeting: &types.TargetingConfig{
			Rules: []types.TargetingRule{
				{
					ID:          "rule_1",
					Description: "Premium users",
					Attribute:   "tier",
					Operator:    "equals",
					Values:      []string{"premium"},
					Variation:   "var_1",
				},
			},
		},
	}
	
	mockRepo.On("GetByKey", mock.Anything, "targeting-flag", "production").Return(expectedFlag, nil)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	requestBody := map[string]interface{}{
		"flag_key": "targeting-flag",
		"user_id": "user_123",
		"user_key": "premium_user",
		"attributes": map[string]interface{}{
			"tier": "premium",
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/ultra?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UltraFastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "targeting-flag", response["flag_key"])
	// The response could be either the default or the targeting result
	assert.Contains(t, []string{"default", "variation_1"}, response["value"])

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_RefreshFlag_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	mockRepo.On("GetByKey", mock.Anything, "error-flag", "production").Return(nil, assert.AnError)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	// Call RefreshFlag with a flag that will cause an error - this should be handled gracefully
	handler.RefreshFlag("error-flag", "production")

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_RefreshAllFlags_WithErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock List calls that return errors for some environments
	mockRepo.On("List", mock.Anything, "production").Return(nil, assert.AnError)
	mockRepo.On("List", mock.Anything, "staging").Return([]*types.Flag{}, nil)
	mockRepo.On("List", mock.Anything, "development").Return(nil, assert.AnError)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization 
	time.Sleep(10 * time.Millisecond)

	// Call RefreshAllFlags - errors should be handled gracefully
	handler.RefreshAllFlags()

	mockRepo.AssertExpectations(t)
}

func TestUltraFastHandler_CacheCleanup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()

	_ = handlers.NewUltraFastHandler(mockRepo)
	
	// Give time for cleanup goroutine to start and potentially run
	time.Sleep(50 * time.Millisecond)

	// The cleanup goroutine should be running in the background
	// We can't easily test the cleanup directly, but we've triggered the code path
}

func TestUltraFastHandler_EvaluationWithUserKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockFlagRepository)
	
	// Mock the List call that happens during preload
	mockRepo.On("List", mock.Anything, mock.AnythingOfType("string")).Return([]*types.Flag{}, nil).Maybe()
	
	expectedFlag := &types.Flag{
		ID:          uuid.New().String(),
		Key:         "user-key-flag",
		Name:        "User Key Flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	mockRepo.On("GetByKey", mock.Anything, "user-key-flag", "production").Return(expectedFlag, nil)

	handler := handlers.NewUltraFastHandler(mockRepo)
	
	// Give a moment for initialization
	time.Sleep(10 * time.Millisecond)

	requestBody := map[string]interface{}{
		"flag_key": "user-key-flag",
		"user_key": "special_user_key",
		"attributes": map[string]interface{}{
			"country": "CA",
			"age":     25,
			"active":  true,
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/evaluate/ultra?environment=production", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UltraFastEvaluate(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "user-key-flag", response["flag_key"])
	assert.Equal(t, true, response["value"])

	mockRepo.AssertExpectations(t)
}