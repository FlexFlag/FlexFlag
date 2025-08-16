package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestFlagRepository_GetByKey_NotFound(t *testing.T) {
	mockRepo := new(MockFlagRepository)
	
	// Mock flag not found
	mockRepo.On("GetByKey", mock.Anything, "nonexistent-flag", "production").Return(nil, assert.AnError)
	
	ctx := context.Background()
	flag, err := mockRepo.GetByKey(ctx, "nonexistent-flag", "production")
	
	assert.Error(t, err)
	assert.Nil(t, flag)
	mockRepo.AssertExpectations(t)
}

func TestFlagRepository_GetByKey_Success(t *testing.T) {
	mockRepo := new(MockFlagRepository)
	
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
	
	ctx := context.Background()
	flag, err := mockRepo.GetByKey(ctx, "test-flag", "production")
	
	assert.NoError(t, err)
	assert.NotNil(t, flag)
	assert.Equal(t, "test-flag", flag.Key)
	assert.Equal(t, types.FlagTypeBoolean, flag.Type)
	assert.True(t, flag.Enabled)
	mockRepo.AssertExpectations(t)
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
		c.ShouldBindJSON(&req)
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