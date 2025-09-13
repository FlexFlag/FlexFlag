package sdk

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	config := &Config{
		BaseURL:         "http://localhost:8080",
		APIKey:          "test-api-key",
		RefreshInterval: time.Minute,
		Timeout:         5 * time.Second,
		Environment:     "production",
	}

	client := NewClient(config)
	
	assert.NotNil(t, client)
	assert.Equal(t, config.BaseURL, client.baseURL)
	assert.Equal(t, config.APIKey, client.apiKey)
	assert.Equal(t, config, client.config)
	assert.NotNil(t, client.httpClient)
	assert.NotNil(t, client.cache)
	assert.NotNil(t, client.stopCh)
}

func TestNewClient_Defaults(t *testing.T) {
	config := &Config{
		BaseURL: "http://localhost:8080",
		APIKey:  "test-api-key",
	}

	client := NewClient(config)
	
	assert.Equal(t, 30*time.Second, client.config.RefreshInterval)
	assert.Equal(t, 10*time.Second, client.config.Timeout)
	assert.Equal(t, "production", client.config.Environment)
}

func TestClient_EvaluateFlag(t *testing.T) {
	// Create test server that handles both refresh and evaluate endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/api/v1/flags" {
			// Handle refresh endpoint
			resp := map[string]interface{}{
				"flags": []*types.Flag{},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.URL.Path == "/api/v1/evaluate" {
			// Handle evaluate endpoint
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req types.EvaluationRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test-flag", req.FlagKey)
			assert.Equal(t, "user_123", req.UserID)
			assert.Equal(t, "production", req.Environment)

			resp := &types.EvaluationResponse{
				FlagKey:   "test-flag",
				Value:     json.RawMessage(`true`),
				Variation: "default",
				Reason:    "default_value",
				Default:   true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value, err := client.EvaluateFlag(ctx, "test-flag", userContext)

	require.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, true, value)
}

func TestClient_EvaluateFlag_FromCache(t *testing.T) {
	config := &Config{
		BaseURL:     "http://localhost:8080",
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	// Pre-populate cache
	cachedFlag := &types.Flag{
		ID:          "cached_123",
		Key:         "cached-flag",
		Type:        types.FlagTypeString,
		Enabled:     true,
		Default:     json.RawMessage(`"cached"`),
		Environment: "production",
	}
	
	client.cacheMu.Lock()
	client.cache["cached-flag"] = cachedFlag
	client.cacheMu.Unlock()

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value, err := client.EvaluateFlag(ctx, "cached-flag", userContext)

	require.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, "cached", value)
}

func TestClient_EvaluateFlag_FromAPI(t *testing.T) {
	// Create test server that handles both refresh and evaluate endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/api/v1/flags" {
			// Handle refresh endpoint
			resp := map[string]interface{}{
				"flags": []*types.Flag{},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.URL.Path == "/api/v1/evaluate" {
			// Handle evaluate endpoint
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req types.EvaluationRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, "test-flag", req.FlagKey)
			assert.Equal(t, "user_123", req.UserID)
			assert.Equal(t, "production", req.Environment)

			resp := &types.EvaluationResponse{
				FlagKey:   "test-flag",
				Value:     json.RawMessage(`true`),
				Variation: "default",
				Reason:    "default_value",
				Default:   true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value, err := client.EvaluateFlag(ctx, "test-flag", userContext)

	require.NoError(t, err)
	assert.NotNil(t, value)
	assert.Equal(t, true, value)
}

func TestClient_BoolValue(t *testing.T) {
	// Create test server that handles both refresh and evaluate endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/api/v1/flags" {
			// Handle refresh endpoint
			resp := map[string]interface{}{
				"flags": []*types.Flag{},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.URL.Path == "/api/v1/evaluate" {
			// Handle evaluate endpoint
			resp := &types.EvaluationResponse{
				FlagKey:   "bool-flag",
				Value:     json.RawMessage(`true`),
				Variation: "default",
				Reason:    "default_value",
				Default:   true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value := client.BoolValue(ctx, "bool-flag", userContext, false)

	assert.True(t, value)
}

func TestClient_BoolValue_Error(t *testing.T) {
	// Create test server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value := client.BoolValue(ctx, "bool-flag", userContext, false)

	// Should return default value on error
	assert.False(t, value)
}

func TestClient_StringValue(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := &types.EvaluationResponse{
			FlagKey:   "string-flag",
			Value:     json.RawMessage(`"test-value"`),
			Variation: "default",
			Reason:    "default_value",
			Default:   true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value := client.StringValue(ctx, "string-flag", userContext, "default")

	assert.Equal(t, "test-value", value)
}

func TestClient_NumberValue(t *testing.T) {
	// Create test server that handles both refresh and evaluate endpoints
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		
		if r.URL.Path == "/api/v1/flags" {
			// Handle refresh endpoint
			resp := map[string]interface{}{
				"flags": []*types.Flag{},
			}
			_ = json.NewEncoder(w).Encode(resp)
		} else if r.URL.Path == "/api/v1/evaluate" {
			// Handle evaluate endpoint
			resp := &types.EvaluationResponse{
				FlagKey:   "number-flag",
				Value:     json.RawMessage(`42.5`),
				Variation: "default",
				Reason:    "default_value",
				Default:   true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	value := client.NumberValue(ctx, "number-flag", userContext, 0)

	assert.Equal(t, 42.5, value)
}


func TestClient_RefreshFlags(t *testing.T) {
	callCount := 0
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		assert.Equal(t, "/api/v1/flags", r.URL.Path)
		assert.Equal(t, "production", r.URL.Query().Get("environment"))

		flags := []*types.Flag{
			{
				ID:          "flag_1",
				Key:         "flag-1",
				Type:        types.FlagTypeBoolean,
				Enabled:     true,
				Default:     json.RawMessage(`true`),
				Environment: "production",
			},
			{
				ID:          "flag_2",
				Key:         "flag-2",
				Type:        types.FlagTypeString,
				Enabled:     false,
				Default:     json.RawMessage(`"test"`),
				Environment: "production",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"flags": flags,
		})
	}))
	defer server.Close()

	config := &Config{
		BaseURL:         server.URL,
		APIKey:          "test-api-key",
		Environment:     "production",
		RefreshInterval: time.Millisecond * 100,
	}
	client := NewClient(config)

	// Wait for automatic refresh to happen (client starts refreshing automatically)
	time.Sleep(time.Millisecond * 150)

	// Check cache is populated
	client.cacheMu.RLock()
	assert.Len(t, client.cache, 2)
	assert.NotNil(t, client.cache["flag-1"])
	assert.NotNil(t, client.cache["flag-2"])
	client.cacheMu.RUnlock()
	
	// Verify refresh was called
	assert.Greater(t, callCount, 0)
}

func TestClient_Close(t *testing.T) {
	config := &Config{
		BaseURL:     "http://localhost:8080",
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)
	
	// Close should not panic
	assert.NotPanics(t, func() {
		client.Close()
	})
}

func TestConfig_Structure(t *testing.T) {
	config := Config{
		BaseURL:         "http://api.example.com",
		APIKey:          "api-key-123",
		RefreshInterval: 5 * time.Minute,
		Timeout:         30 * time.Second,
		Environment:     "staging",
	}

	assert.Equal(t, "http://api.example.com", config.BaseURL)
	assert.Equal(t, "api-key-123", config.APIKey)
	assert.Equal(t, 5*time.Minute, config.RefreshInterval)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "staging", config.Environment)
}

func BenchmarkClient_EvaluateFlag_FromCache(b *testing.B) {
	config := &Config{
		BaseURL:     "http://localhost:8080",
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	// Pre-populate cache
	cachedFlag := &types.Flag{
		ID:          "bench_123",
		Key:         "bench-flag",
		Type:        types.FlagTypeBoolean,
		Enabled:     true,
		Default:     json.RawMessage(`true`),
		Environment: "production",
	}
	
	client.cacheMu.Lock()
	client.cache["bench-flag"] = cachedFlag
	client.cacheMu.Unlock()

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.EvaluateFlag(ctx, "bench-flag", userContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkClient_BoolValue(b *testing.B) {
	// Create test server that mimics flag endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flag := &types.Flag{
			ID:          "flag_123",
			Key:         "bool-flag",
			Type:        types.FlagTypeBoolean,
			Enabled:     true,
			Default:     json.RawMessage(`true`),
			Environment: "production",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(flag)
	}))
	defer server.Close()

	config := &Config{
		BaseURL:     server.URL,
		APIKey:      "test-api-key",
		Environment: "production",
	}
	client := NewClient(config)

	ctx := context.Background()
	userContext := NewUserContext("user_123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.BoolValue(ctx, "bool-flag", userContext, false)
	}
}