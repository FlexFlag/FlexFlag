package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/internal/services"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockFlagRepository mocks storage.FlagRepository
type MockFlagRepository struct {
	mock.Mock
}

func (m *MockFlagRepository) Create(ctx context.Context, flag *types.Flag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFlagRepository) GetByKey(ctx context.Context, key, environment string) (*types.Flag, error) {
	args := m.Called(ctx, key, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Flag), args.Error(1)
}

func (m *MockFlagRepository) GetByProjectKey(ctx context.Context, projectID, key, environment string) (*types.Flag, error) {
	args := m.Called(ctx, projectID, key, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Flag), args.Error(1)
}

func (m *MockFlagRepository) Update(ctx context.Context, flag *types.Flag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFlagRepository) Delete(ctx context.Context, key, environment string) error {
	args := m.Called(ctx, key, environment)
	return args.Error(0)
}

func (m *MockFlagRepository) List(ctx context.Context, environment string) ([]*types.Flag, error) {
	args := m.Called(ctx, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

func (m *MockFlagRepository) ListByProject(ctx context.Context, projectID, environment string) ([]*types.Flag, error) {
	args := m.Called(ctx, projectID, environment)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

func (m *MockFlagRepository) ListByTags(ctx context.Context, environment string, tags []string) ([]*types.Flag, error) {
	args := m.Called(ctx, environment, tags)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.Flag), args.Error(1)
}

// MockAuditService mocks the audit service
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(ctx context.Context, action, resourceType, resourceID string, details map[string]interface{}) error {
	args := m.Called(ctx, action, resourceType, resourceID, details)
	return args.Error(0)
}

// MockProjectRepository mocks the project repository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id string) (*types.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Project), args.Error(1)
}

func TestFlagHandler_CreateFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_creation", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, &services.AuditService{}, nil, &postgres.ProjectRepository{})

		// Mock the Create call
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.Flag")).Return(nil)

		requestBody := map[string]interface{}{
			"key":         "test-flag",
			"name":        "Test Flag",
			"description": "Test description",
			"type":        "boolean",
			"enabled":     true,
			"default":     true,
			"environment": "production",
			"project_id":  "proj_123",
			"tags":        []string{"test"},
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateFlag(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "flag")
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid_request_body", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		// Invalid request body (missing required fields)
		requestBody := map[string]interface{}{
			"name": "Test Flag",
			// Missing required "key" field
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateFlag(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertNotCalled(t, "Create")
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.Flag")).Return(fmt.Errorf("database error"))

		requestBody := map[string]interface{}{
			"key":         "test-flag",
			"name":        "Test Flag",
			"type":        "boolean",
			"enabled":     true,
			"default":     true,
			"environment": "production",
			"project_id":  "proj_123",
		}

		body, err := json.Marshal(requestBody)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req

		handler.CreateFlag(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestFlagHandler_GetFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_retrieval", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		expectedFlag := &types.Flag{
			ID:          uuid.New().String(),
			Key:         "test-flag",
			Name:        "Test Flag",
			Type:        types.FlagTypeBoolean,
			Enabled:     true,
			Default:     json.RawMessage(`true`),
			Environment: "production",
			ProjectID:   "proj_123",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(expectedFlag, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodGet, "/api/v1/flags/test-flag?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "test-flag"}}

		handler.GetFlag(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "flag")
		
		flag := response["flag"].(map[string]interface{})
		assert.Equal(t, "test-flag", flag["key"])
		assert.Equal(t, "Test Flag", flag["name"])
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("flag_not_found", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("GetByKey", mock.Anything, "non-existent", "production").Return(nil, fmt.Errorf("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodGet, "/api/v1/flags/non-existent?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "non-existent"}}

		handler.GetFlag(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestFlagHandler_ListFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_list", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		expectedFlags := []*types.Flag{
			{
				ID:          uuid.New().String(),
				Key:         "flag1",
				Name:        "Flag 1",
				Type:        types.FlagTypeBoolean,
				Enabled:     true,
				Environment: "production",
				ProjectID:   "proj_123",
			},
			{
				ID:          uuid.New().String(),
				Key:         "flag2",
				Name:        "Flag 2",
				Type:        types.FlagTypeString,
				Enabled:     false,
				Environment: "production",
				ProjectID:   "proj_123",
			},
		}

		mockRepo.On("List", mock.Anything, "production").Return(expectedFlags, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodGet, "/api/v1/flags?environment=production", nil)
		c.Request = req

		handler.ListFlags(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "flags")
		
		flags := response["flags"].([]interface{})
		assert.Len(t, flags, 2)
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty_list", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("List", mock.Anything, "production").Return([]*types.Flag{}, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodGet, "/api/v1/flags?environment=production", nil)
		c.Request = req

		handler.ListFlags(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "flags")
		
		flags := response["flags"].([]interface{})
		assert.Len(t, flags, 0)
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("repository_error", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("List", mock.Anything, "production").Return(nil, fmt.Errorf("database error"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodGet, "/api/v1/flags?environment=production", nil)
		c.Request = req

		handler.ListFlags(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestFlagHandler_UpdateFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_update", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		existingFlag := &types.Flag{
			ID:          "flag_123",
			Key:         "test-flag",
			Name:        "Test Flag",
			Type:        types.FlagTypeBoolean,
			Enabled:     true,
			Environment: "production",
			ProjectID:   "proj_123",
		}

		mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(existingFlag, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Flag")).Return(nil)

		updateRequest := map[string]interface{}{
			"key":         "test-flag",
			"type":        "boolean",
			"name":        "Updated Flag",
			"description": "Updated description",
			"enabled":     false,
		}

		body, err := json.Marshal(updateRequest)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPut, "/api/v1/flags/test-flag?environment=production", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "test-flag"}}

		handler.UpdateFlag(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("flag_not_found", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("GetByKey", mock.Anything, "non-existent", "production").Return(nil, fmt.Errorf("not found"))

		updateRequest := map[string]interface{}{
			"key":  "non-existent",
			"type": "boolean",
			"name": "Updated Flag",
		}

		body, err := json.Marshal(updateRequest)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPut, "/api/v1/flags/non-existent?environment=production", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "non-existent"}}

		handler.UpdateFlag(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid_request", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		// Missing required fields
		updateRequest := map[string]interface{}{
			"name": "Updated Flag",
			// Missing key and type
		}

		body, err := json.Marshal(updateRequest)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPut, "/api/v1/flags/test-flag?environment=production", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "test-flag"}}

		handler.UpdateFlag(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertNotCalled(t, "GetByKey")
	})
}

func TestFlagHandler_DeleteFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_deletion", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("Delete", mock.Anything, "test-flag", "production").Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/flags/test-flag?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "test-flag"}}

		handler.DeleteFlag(c)

		assert.Equal(t, http.StatusNoContent, w.Code)
		mockRepo.AssertExpectations(t)
	})

	t.Run("flag_not_found", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("Delete", mock.Anything, "non-existent", "production").Return(fmt.Errorf("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/flags/non-existent?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "non-existent"}}

		handler.DeleteFlag(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestFlagHandler_ToggleFlag(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("successful_toggle", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		existingFlag := &types.Flag{
			ID:          "flag_123",
			Key:         "test-flag",
			Name:        "Test Flag",
			Type:        types.FlagTypeBoolean,
			Enabled:     true,
			Environment: "production",
			ProjectID:   "proj_123",
		}

		mockRepo.On("GetByKey", mock.Anything, "test-flag", "production").Return(existingFlag, nil)
		mockRepo.On("Update", mock.Anything, mock.AnythingOfType("*types.Flag")).Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/flags/test-flag/toggle?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "test-flag"}}

		handler.ToggleFlag(c)

		assert.Equal(t, http.StatusOK, w.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Flag toggled successfully", response["message"])
		assert.Equal(t, false, response["enabled"]) // Was true, now false
		
		mockRepo.AssertExpectations(t)
	})

	t.Run("flag_not_found", func(t *testing.T) {
		mockRepo := new(MockFlagRepository)
		handler := handlers.NewFlagHandler(mockRepo, nil, nil, nil)

		mockRepo.On("GetByKey", mock.Anything, "non-existent", "production").Return(nil, fmt.Errorf("not found"))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		
		req := httptest.NewRequest(http.MethodPost, "/api/v1/flags/non-existent/toggle?environment=production", nil)
		c.Request = req
		c.Params = gin.Params{{Key: "key", Value: "non-existent"}}

		handler.ToggleFlag(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}