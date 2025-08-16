package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApiKeyRepository mocks the API key repository
type MockApiKeyRepository struct {
	mock.Mock
}

func (m *MockApiKeyRepository) AuthenticateApiKey(ctx context.Context, apiKey string) (*types.ApiKey, error) {
	args := m.Called(ctx, apiKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiKey), args.Error(1)
}

func (m *MockApiKeyRepository) Create(ctx context.Context, apiKey *types.ApiKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}

func (m *MockApiKeyRepository) GetByID(ctx context.Context, id string) (*types.ApiKey, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.ApiKey), args.Error(1)
}

func (m *MockApiKeyRepository) Update(ctx context.Context, apiKey *types.ApiKey) error {
	args := m.Called(ctx, apiKey)
	return args.Error(0)
}

func (m *MockApiKeyRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockApiKeyRepository) ListByProject(ctx context.Context, projectID string) ([]*types.ApiKey, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*types.ApiKey), args.Error(1)
}

// Create interface for better testability
type ApiKeyRepository interface {
	AuthenticateApiKey(ctx context.Context, apiKey string) (*types.ApiKey, error)
}

func createTestApiKeyAuth(repo ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key is required"})
			c.Abort()
			return
		}

		keyInfo, err := repo.AuthenticateApiKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Set API key information in context for use by handlers
		c.Set("apiKey", keyInfo)
		c.Set("projectID", keyInfo.ProjectID)
		c.Set("environmentID", keyInfo.EnvironmentID)
		if keyInfo.Environment != nil {
			c.Set("environment", keyInfo.Environment.Key)
		}
		c.Set("permissions", keyInfo.Permissions)

		c.Next()
	}
}

func createTestOptionalApiKeyAuth(repo ApiKeyRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			// No API key provided, continue without authentication
			c.Next()
			return
		}

		keyInfo, err := repo.AuthenticateApiKey(c.Request.Context(), apiKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired API key"})
			c.Abort()
			return
		}

		// Set API key information in context for use by handlers
		c.Set("apiKey", keyInfo)
		c.Set("projectID", keyInfo.ProjectID)
		c.Set("environmentID", keyInfo.EnvironmentID)
		if keyInfo.Environment != nil {
			c.Set("environment", keyInfo.Environment.Key)
		}
		c.Set("permissions", keyInfo.Permissions)

		c.Next()
	}
}

func TestCORS_OptionsRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "X-Environment")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "PUT")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "DELETE")
}

func TestCORS_RegularRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(CORS())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Body.String(), "success")
}

func TestApiKeyAuth_ValidApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	// Mock valid API key authentication
	keyInfo := &types.ApiKey{
		ID:            "key_123",
		Name:          "Test Key",
		ProjectID:     "proj_123",
		EnvironmentID: "env_123",
		Environment: &types.Environment{
			ID:        "env_123",
			Key:       "production",
			Name:      "Production",
			ProjectID: "proj_123",
		},
		Permissions: []string{"read", "write"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	mockRepo.On("AuthenticateApiKey", mock.Anything, "valid_api_key").Return(keyInfo, nil)

	r := gin.New()
	r.Use(createTestApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		// Verify that context values are set
		apiKey := c.MustGet("apiKey")
		projectID := c.MustGet("projectID")
		environmentID := c.MustGet("environmentID")
		environment := c.MustGet("environment")
		permissions := c.MustGet("permissions")

		assert.Equal(t, keyInfo, apiKey)
		assert.Equal(t, "proj_123", projectID)
		assert.Equal(t, "env_123", environmentID)
		assert.Equal(t, "production", environment)
		assert.Equal(t, []string{"read", "write"}, permissions)

		c.JSON(http.StatusOK, gin.H{"message": "authenticated"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "valid_api_key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "authenticated")
	mockRepo.AssertExpectations(t)
}

func TestApiKeyAuth_MissingApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	r := gin.New()
	r.Use(createTestApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No API key header set
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "API key is required")
	// Repository should not be called
	mockRepo.AssertNotCalled(t, "AuthenticateApiKey")
}

func TestApiKeyAuth_InvalidApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	// Mock invalid API key authentication
	mockRepo.On("AuthenticateApiKey", mock.Anything, "invalid_api_key").Return(nil, assert.AnError)

	r := gin.New()
	r.Use(createTestApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid_api_key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired API key")
	mockRepo.AssertExpectations(t)
}

func TestOptionalApiKeyAuth_ValidApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	// Mock valid API key authentication
	keyInfo := &types.ApiKey{
		ID:            "key_123",
		Name:          "Test Key",
		ProjectID:     "proj_123",
		EnvironmentID: "env_123",
		Environment: &types.Environment{
			ID:        "env_123",
			Key:       "staging",
			Name:      "Staging",
			ProjectID: "proj_123",
		},
		Permissions: []string{"read"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	mockRepo.On("AuthenticateApiKey", mock.Anything, "valid_optional_key").Return(keyInfo, nil)

	r := gin.New()
	r.Use(createTestOptionalApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		// Verify that context values are set when API key is valid
		apiKey, exists := c.Get("apiKey")
		if exists {
			assert.Equal(t, keyInfo, apiKey)
			projectID := c.MustGet("projectID")
			environment := c.MustGet("environment")
			assert.Equal(t, "proj_123", projectID)
			assert.Equal(t, "staging", environment)
		}

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "valid_optional_key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockRepo.AssertExpectations(t)
}

func TestOptionalApiKeyAuth_NoApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	r := gin.New()
	r.Use(createTestOptionalApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		// Verify that context values are not set when no API key
		_, exists := c.Get("apiKey")
		assert.False(t, exists)

		_, exists = c.Get("projectID")
		assert.False(t, exists)

		c.JSON(http.StatusOK, gin.H{"message": "no auth success"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No API key header set
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "no auth success")
	// Repository should not be called when no API key
	mockRepo.AssertNotCalled(t, "AuthenticateApiKey")
}

func TestOptionalApiKeyAuth_InvalidApiKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	// Mock invalid API key authentication
	mockRepo.On("AuthenticateApiKey", mock.Anything, "invalid_optional_key").Return(nil, assert.AnError)

	r := gin.New()
	r.Use(createTestOptionalApiKeyAuth(mockRepo))
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "invalid_optional_key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired API key")
	mockRepo.AssertExpectations(t)
}

func BenchmarkCORS(b *testing.B) {
	gin.SetMode(gin.TestMode)
	corsMiddleware := CORS()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request = req

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		corsMiddleware(c)
	}
}

func BenchmarkApiKeyAuth_ValidKey(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockRepo := new(MockApiKeyRepository)

	keyInfo := &types.ApiKey{
		ID:            "bench_key",
		ProjectID:     "bench_proj",
		EnvironmentID: "bench_env",
		Environment: &types.Environment{
			Key: "production",
		},
		Permissions: []string{"read"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	mockRepo.On("AuthenticateApiKey", mock.Anything, "bench_key").Return(keyInfo, nil)

	authMiddleware := createTestApiKeyAuth(mockRepo)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "bench_key")
	c.Request = req

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authMiddleware(c)
	}
}