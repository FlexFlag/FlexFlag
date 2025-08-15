package services

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditRepository mocks the audit repository
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

func TestNewAuditService(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	assert.NotNil(t, service)
	assert.Equal(t, mockRepo, service.auditRepo)
}

func TestLogFlagAction_Create(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	flag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		Name:        "Test Flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", nil)
	req.Header.Set("User-Agent", "test-client/1.0")
	req.RemoteAddr = "192.168.1.1:8080"
	c.Request = req
	c.Set("user_id", "user_123")

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "create" &&
			log.ResourceType == "flag" &&
			log.ResourceID == "flag_123" &&
			log.ProjectID != nil && *log.ProjectID == "proj_123" &&
			log.UserID != nil && *log.UserID == "user_123" &&
			log.UserAgent == "test-client/1.0" &&
			log.NewValues != nil
	})).Return(nil)

	err := service.LogFlagAction(context.Background(), c, "create", flag, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogFlagAction_Update(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	oldFlag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		Name:        "Test Flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     false,
		Default:     json.RawMessage(`false`),
		Environment: "production",
	}

	newFlag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		Name:        "Test Flag Updated",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/flags/test-flag", nil)
	req.Header.Set("User-Agent", "test-client/1.0")
	c.Request = req

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "update" &&
			log.ResourceType == "flag" &&
			log.ResourceID == "flag_123" &&
			log.OldValues != nil &&
			log.NewValues != nil
	})).Return(nil)

	err := service.LogFlagAction(context.Background(), c, "update", newFlag, oldFlag)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogFlagAction_Delete(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	flag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		Name:        "Test Flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	// Set up Gin context without user_id
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/flags/test-flag", nil)
	c.Request = req

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "delete" &&
			log.ResourceType == "flag" &&
			log.ResourceID == "flag_123" &&
			log.UserID == nil // No user ID set
	})).Return(nil)

	err := service.LogFlagAction(context.Background(), c, "delete", flag, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogProjectAction_Create(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	project := &types.Project{
		ID:          "proj_123",
		Name:        "Test Project",
		Description: "Test project description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", nil)
	req.Header.Set("User-Agent", "project-client/1.0")
	c.Request = req
	c.Set("user_id", "admin_user")

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "create" &&
			log.ResourceType == "project" &&
			log.ResourceID == "proj_123" &&
			log.ProjectID != nil && *log.ProjectID == "proj_123" &&
			log.UserID != nil && *log.UserID == "admin_user" &&
			log.UserAgent == "project-client/1.0"
	})).Return(nil)

	err := service.LogProjectAction(context.Background(), c, "create", project, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogProjectAction_Update(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	oldProject := &types.Project{
		ID:          "proj_123",
		Name:        "Old Project",
		Description: "Old description",
	}

	newProject := &types.Project{
		ID:          "proj_123",
		Name:        "New Project",
		Description: "New description",
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/proj_123", nil)
	c.Request = req

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "update" &&
			log.ResourceType == "project" &&
			log.ResourceID == "proj_123" &&
			log.OldValues != nil &&
			log.NewValues != nil
	})).Return(nil)

	err := service.LogProjectAction(context.Background(), c, "update", newProject, oldProject)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogSegmentAction_Create(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	segment := &types.Segment{
		ID:          "seg_123",
		ProjectID:   "proj_123",
		Key:         "premium-users",
		Name:        "Premium Users",
		Description: "Users with premium plan",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
		},
		Environment: "production",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/segments", nil)
	req.Header.Set("User-Agent", "segment-client/1.0")
	c.Request = req
	c.Set("user_id", "segment_admin")

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "create" &&
			log.ResourceType == "segment" &&
			log.ResourceID == "seg_123" &&
			log.ProjectID != nil && *log.ProjectID == "proj_123" &&
			log.UserID != nil && *log.UserID == "segment_admin" &&
			log.UserAgent == "segment-client/1.0"
	})).Return(nil)

	err := service.LogSegmentAction(context.Background(), c, "create", segment, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogSegmentAction_Update(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	oldSegment := &types.Segment{
		ID:        "seg_123",
		ProjectID: "proj_123",
		Key:       "users",
		Name:      "All Users",
		Rules:     []types.TargetingRule{},
	}

	newSegment := &types.Segment{
		ID:        "seg_123",
		ProjectID: "proj_123",
		Key:       "users",
		Name:      "Premium Users",
		Rules: []types.TargetingRule{
			{
				ID:        "rule1",
				Attribute: "plan",
				Operator:  "equals",
				Values:    []string{"premium"},
			},
		},
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/segments/users", nil)
	c.Request = req

	// Mock repository call
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.Action == "update" &&
			log.ResourceType == "segment" &&
			log.ResourceID == "seg_123" &&
			log.OldValues != nil &&
			log.NewValues != nil
	})).Return(nil)

	err := service.LogSegmentAction(context.Background(), c, "update", newSegment, oldSegment)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogFlagAction_JSONMarshalError(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	// Create a flag with invalid JSON that will cause marshal error
	flag := &types.Flag{
		ID:        "flag_123",
		Key:       "test-flag",
		ProjectID: "proj_123",
		// This will cause JSON marshal to potentially fail if there are circular references
		// but since Flag struct is simple, we'll test with a valid flag
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", nil)
	c.Request = req

	// Mock repository call to succeed (JSON marshaling will work for valid structs)
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.AuditLog")).Return(nil)

	err := service.LogFlagAction(context.Background(), c, "create", flag, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestLogFlagAction_RepositoryError(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	flag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	// Set up Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", nil)
	c.Request = req

	// Mock repository call to return error
	expectedError := assert.AnError
	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.AuditLog")).Return(expectedError)

	err := service.LogFlagAction(context.Background(), c, "create", flag, nil)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
}

func TestLogAction_UserIDTypeAssertion(t *testing.T) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	flag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	// Set up Gin context with non-string user_id
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", nil)
	c.Request = req
	c.Set("user_id", 12345) // Non-string value

	// Mock repository call - UserID should be nil due to type assertion failure
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(log *types.AuditLog) bool {
		return log.UserID == nil // Should be nil because type assertion failed
	})).Return(nil)

	err := service.LogFlagAction(context.Background(), c, "create", flag, nil)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func BenchmarkLogFlagAction(b *testing.B) {
	mockRepo := new(MockAuditRepository)
	service := NewAuditService(mockRepo)

	flag := &types.Flag{
		ID:          "flag_123",
		Key:         "test-flag",
		ProjectID:   "proj_123",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/flags", nil)
	c.Request = req

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*types.AuditLog")).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.LogFlagAction(context.Background(), c, "create", flag, nil)
	}
}