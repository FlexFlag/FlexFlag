package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuditRepository for audit handler tests
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) Create(ctx context.Context, log *types.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditRepository) List(ctx context.Context, projectID string, limit, offset int) ([]*types.AuditLog, error) {
	args := m.Called(ctx, projectID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*types.AuditLog, error) {
	args := m.Called(ctx, resourceType, resourceID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*types.AuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.AuditLog), args.Error(1)
}

func TestAuditHandler_ListAuditLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	expectedLogs := []*types.AuditLog{
		{
			ID:           "log1",
			ResourceType: "flag",
			ResourceID:   "flag1",
			Action:       "update",
		},
		{
			ID:           "log2",
			ResourceType: "flag",
			ResourceID:   "flag2",
			Action:       "create",
		},
	}

	mockRepo.On("List", mock.Anything, "proj_123", 50, 0).Return(expectedLogs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs?project_id=proj_123", nil)
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "logs")

	logs := response["logs"].([]interface{})
	assert.Len(t, logs, 2)

	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_MissingProjectID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs", nil) // Missing project_id
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "project_id is required", response["error"])
}

func TestAuditHandler_ListAuditLogs_WithPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	expectedLogs := []*types.AuditLog{
		{
			ID:           "log3",
			ResourceType: "flag",
			ResourceID:   "flag3",
			Action:       "delete",
		},
	}

	mockRepo.On("List", mock.Anything, "proj_123", 10, 20).Return(expectedLogs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs?project_id=proj_123&limit=10&offset=20", nil)
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "logs")

	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	expectedLogs := []*types.AuditLog{
		{
			ID:           "log1",
			ResourceType: "flag",
			ResourceID:   "flag_123",
			Action:       "update",
		},
	}

	mockRepo.On("ListByResource", mock.Anything, "flag", "flag_123", 50, 0).Return(expectedLogs, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_type=flag&resource_id=flag_123", nil)
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response, "logs")

	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_MissingParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_type=flag", nil) // Missing resource_id
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "resource_type and resource_id are required", response["error"])
}

func TestAuditHandler_ListAuditLogs_InvalidPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	// Mock should still be called with default values when parsing fails
	mockRepo.On("List", mock.Anything, "proj_123", 50, 0).Return([]*types.AuditLog{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs?project_id=proj_123&limit=invalid&offset=negative", nil)
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_LimitTooHigh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	// Should cap limit at 50 when > 1000
	mockRepo.On("List", mock.Anything, "proj_456", 50, 10).Return([]*types.AuditLog{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs?project_id=proj_456&limit=5000&offset=10", nil)
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogs_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	mockRepo.On("List", mock.Anything, "proj_123", 25, 0).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs?project_id=proj_123&limit=25", nil)
	c.Request = req

	handler.ListAuditLogs(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")

	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_InvalidPaginationParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	// Should use defaults when parsing fails
	mockRepo.On("ListByResource", mock.Anything, "flag", "flag_123", 50, 0).Return([]*types.AuditLog{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_type=flag&resource_id=flag_123&limit=abc&offset=-5", nil)
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_LimitTooHigh(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	// Should cap limit at 50 when > 1000  
	mockRepo.On("ListByResource", mock.Anything, "project", "proj_789", 50, 20).Return([]*types.AuditLog{}, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_type=project&resource_id=proj_789&limit=2000&offset=20", nil)
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_DatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	mockRepo.On("ListByResource", mock.Anything, "segment", "seg_456", 30, 5).Return(nil, assert.AnError)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_type=segment&resource_id=seg_456&limit=30&offset=5", nil)
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response["error"], "assert.AnError")

	mockRepo.AssertExpectations(t)
}

func TestAuditHandler_ListAuditLogsByResource_MissingResourceType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockAuditRepository)
	handler := handlers.NewAuditHandler(mockRepo)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/audit/logs/resource?resource_id=some_id", nil) // Missing resource_type
	c.Request = req

	handler.ListAuditLogsByResource(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "resource_type and resource_id are required", response["error"])
}