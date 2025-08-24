package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func TestRolloutHandler_CreateRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.Rollout")).Return(nil)

	requestBody := map[string]interface{}{
		"flag_id":     "flag_123",
		"environment": "production",
		"type":        "percentage",
		"name":        "Test Rollout",
		"description": "A test rollout",
		"config": map[string]interface{}{
			"percentage": 50,
		},
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/rollouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateRollout(c)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Test Rollout", response["name"])
	assert.Equal(t, "flag_123", response["flag_id"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_CreateRollout_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	// Missing required fields
	requestBody := map[string]interface{}{
		"name": "Test Rollout",
		// missing flag_id, environment, type
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/rollouts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.CreateRollout(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRolloutHandler_GetRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	expectedRollout := &types.Rollout{
		ID:          rolloutID,
		FlagID:      "flag_123",
		Environment: "production",
		Type:        "percentage",
		Name:        "Test Rollout",
		Description: "A test rollout",
		Status:      "active",
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(expectedRollout, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodGet, "/rollouts/"+rolloutID, nil)
	c.Request = req

	handler.GetRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, rolloutID, response["id"])
	assert.Equal(t, "Test Rollout", response["name"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetRollout_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(nil, fmt.Errorf("rollout not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodGet, "/rollouts/"+rolloutID, nil)
	c.Request = req

	handler.GetRollout(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "rollout not found", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetRolloutsByFlag_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	flagID := "flag_123"
	environment := "production"
	expectedRollouts := []*types.Rollout{
		{
			ID:          uuid.New().String(),
			FlagID:      flagID,
			Environment: environment,
			Type:        "percentage",
			Name:        "Rollout 1",
			Status:      "active",
		},
		{
			ID:          uuid.New().String(),
			FlagID:      flagID,
			Environment: environment,
			Type:        "experiment",
			Name:        "Rollout 2",
			Status:      "paused",
		},
	}

	mockRepo.On("GetByFlag", mock.Anything, flagID, environment).Return(expectedRollouts, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/rollouts?flag_id=%s&environment=%s", flagID, environment), nil)
	c.Request = req

	handler.GetRolloutsByFlag(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	rollouts := response["rollouts"].([]interface{})
	assert.Len(t, rollouts, 2)

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetRolloutsByFlag_MissingFlagID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts?environment=production", nil) // Missing flag_id
	c.Request = req

	handler.GetRolloutsByFlag(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag_id is required", response["error"])
}

func TestRolloutHandler_UpdateRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	existingRollout := &types.Rollout{
		ID:          rolloutID,
		FlagID:      "flag_123",
		Environment: "production",
		Type:        "percentage",
		Name:        "Old Name",
		Description: "Old Description",
		Status:      "active",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(existingRollout, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Rollout")).Return(nil)

	requestBody := map[string]interface{}{
		"name":        "Updated Name",
		"description": "Updated Description",
		"status":      "paused",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodPut, "/rollouts/"+rolloutID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	handler.UpdateRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, rolloutID, response["id"])
	assert.Equal(t, "Updated Name", response["name"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_DeleteRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	mockRepo.On("Delete", mock.Anything, rolloutID).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodDelete, "/rollouts/"+rolloutID, nil)
	c.Request = req

	handler.DeleteRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Rollout deleted successfully", response["message"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_DeleteRollout_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	mockRepo.On("Delete", mock.Anything, rolloutID).Return(fmt.Errorf("rollout not found"))

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodDelete, "/rollouts/"+rolloutID, nil)
	c.Request = req

	handler.DeleteRollout(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "rollout not found", response["error"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_ActivateRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	existingRollout := &types.Rollout{
		ID:          rolloutID,
		FlagID:      "flag_123",
		Environment: "production",
		Type:        "percentage",
		Name:        "Test Rollout",
		Status:      "draft",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(existingRollout, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Rollout")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodPost, "/rollouts/"+rolloutID+"/activate", nil)
	c.Request = req

	handler.ActivateRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, rolloutID, response["id"])
	assert.Equal(t, "active", response["status"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_PauseRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	existingRollout := &types.Rollout{
		ID:          rolloutID,
		FlagID:      "flag_123",
		Environment: "production",
		Type:        "percentage",
		Name:        "Test Rollout",
		Status:      "active",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(existingRollout, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Rollout")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodPost, "/rollouts/"+rolloutID+"/pause", nil)
	c.Request = req

	handler.PauseRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, rolloutID, response["id"])
	assert.Equal(t, "paused", response["status"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_CompleteRollout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	rolloutID := uuid.New().String()
	existingRollout := &types.Rollout{
		ID:          rolloutID,
		FlagID:      "flag_123",
		Environment: "production",
		Type:        "percentage",
		Name:        "Test Rollout",
		Status:      "active",
		CreatedAt:   time.Now().Add(-time.Hour),
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", mock.Anything, rolloutID).Return(existingRollout, nil)
	mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Rollout")).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: rolloutID}}
	req := httptest.NewRequest(http.MethodPost, "/rollouts/"+rolloutID+"/complete", nil)
	c.Request = req

	handler.CompleteRollout(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, rolloutID, response["id"])
	assert.Equal(t, "completed", response["status"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetAllRollouts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	expectedRollouts := []*types.Rollout{
		{
			ID:          uuid.New().String(),
			Name:        "Rollout 1",
			FlagID:      "flag_1",
			Environment: "production",
			Status:      types.RolloutStatusActive,
		},
		{
			ID:          uuid.New().String(),
			Name:        "Rollout 2", 
			FlagID:      "flag_2",
			Environment: "production",
			Status:      types.RolloutStatusPaused,
		},
	}

	mockRepo.On("GetByProject", mock.Anything, "proj_123", "production").Return(expectedRollouts, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts?project_id=proj_123&environment=production", nil)
	c.Request = req

	handler.GetAllRollouts(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "rollouts")

	rollouts := response["rollouts"].([]interface{})
	assert.Len(t, rollouts, 2)

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetAllRollouts_MissingProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts", nil)
	c.Request = req

	handler.GetAllRollouts(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "project_id is required", response["error"])
}

func TestRolloutHandler_GetStickyAssignments_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	expectedAssignment := &types.StickyAssignment{
		FlagID:      "flag_123",
		Environment: "production",
		UserKey:     "user_123",
		VariationID: "var_1",
	}

	mockRepo.On("GetStickyAssignment", mock.Anything, "flag_123", "production", "user_123").Return(expectedAssignment, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts/sticky?flag_id=flag_123&environment=production&user_key=user_123", nil)
	c.Request = req

	handler.GetStickyAssignments(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response types.StickyAssignment
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag_123", response.FlagID)
	assert.Equal(t, "user_123", response.UserKey)

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_GetStickyAssignments_AllAssignments(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts/sticky?flag_id=flag_123", nil)
	c.Request = req

	handler.GetStickyAssignments(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "assignments")
	assignments := response["assignments"].([]interface{})
	assert.Len(t, assignments, 0)
}

func TestRolloutHandler_GetStickyAssignments_MissingFlagID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/rollouts/sticky", nil)
	c.Request = req

	handler.GetStickyAssignments(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag_id is required", response["error"])
}

func TestRolloutHandler_DeleteStickyAssignment_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	mockRepo.On("DeleteStickyAssignment", mock.Anything, "flag_123", "production", "user_123").Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodDelete, "/rollouts/sticky?flag_id=flag_123&environment=production&user_key=user_123", nil)
	c.Request = req

	handler.DeleteStickyAssignment(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Sticky assignment deleted successfully", response["message"])

	mockRepo.AssertExpectations(t)
}

func TestRolloutHandler_DeleteStickyAssignment_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodDelete, "/rollouts/sticky?flag_id=flag_123", nil)
	c.Request = req

	handler.DeleteStickyAssignment(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "flag_id and user_key are required", response["error"])
}

func TestRolloutHandler_CleanupExpiredAssignments_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockRolloutRepository)
	handler := handlers.NewRolloutHandler(mockRepo)

	mockRepo.On("CleanupExpiredAssignments", mock.Anything).Return(nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/rollouts/cleanup", nil)
	c.Request = req

	handler.CleanupExpiredAssignments(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "Expired assignments cleaned up successfully", response["message"])

	mockRepo.AssertExpectations(t)
}