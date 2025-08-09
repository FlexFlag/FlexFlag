package sdk

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flexflag/flexflag/pkg/types"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cache      map[string]*types.Flag
	cacheMu    sync.RWMutex
	refreshing bool
	stopCh     chan struct{}
	config     *Config
}

type Config struct {
	BaseURL        string
	APIKey         string
	RefreshInterval time.Duration
	Timeout        time.Duration
	Environment    string
}

func NewClient(config *Config) *Client {
	if config.RefreshInterval == 0 {
		config.RefreshInterval = 30 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.Environment == "" {
		config.Environment = "production"
	}

	client := &Client{
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		cache:  make(map[string]*types.Flag),
		stopCh: make(chan struct{}),
		config: config,
	}

	go client.startRefreshing()
	return client
}

func (c *Client) Close() {
	close(c.stopCh)
}

func (c *Client) EvaluateFlag(ctx context.Context, flagKey string, userContext *UserContext) (interface{}, error) {
	req := &types.EvaluationRequest{
		FlagKey:     flagKey,
		UserID:      userContext.UserID,
		UserKey:     userContext.UserKey,
		Attributes:  userContext.Attributes,
		Environment: c.config.Environment,
	}

	c.cacheMu.RLock()
	flag, exists := c.cache[flagKey]
	c.cacheMu.RUnlock()

	if exists && flag != nil {
		return c.evaluateLocally(flag, req)
	}

	return c.evaluateRemotely(ctx, req)
}

func (c *Client) BoolValue(ctx context.Context, flagKey string, userContext *UserContext, defaultValue bool) bool {
	value, err := c.EvaluateFlag(ctx, flagKey, userContext)
	if err != nil {
		return defaultValue
	}

	if boolVal, ok := value.(bool); ok {
		return boolVal
	}
	return defaultValue
}

func (c *Client) StringValue(ctx context.Context, flagKey string, userContext *UserContext, defaultValue string) string {
	value, err := c.EvaluateFlag(ctx, flagKey, userContext)
	if err != nil {
		return defaultValue
	}

	if strVal, ok := value.(string); ok {
		return strVal
	}
	return defaultValue
}

func (c *Client) NumberValue(ctx context.Context, flagKey string, userContext *UserContext, defaultValue float64) float64 {
	value, err := c.EvaluateFlag(ctx, flagKey, userContext)
	if err != nil {
		return defaultValue
	}

	if numVal, ok := value.(float64); ok {
		return numVal
	}
	return defaultValue
}

func (c *Client) evaluateLocally(flag *types.Flag, req *types.EvaluationRequest) (interface{}, error) {
	// Simple local evaluation - just return default for now
	// Full evaluation engine would be implemented here
	var value interface{}
	if err := json.Unmarshal(flag.Default, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func (c *Client) evaluateRemotely(ctx context.Context, req *types.EvaluationRequest) (interface{}, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/api/v1/evaluate", c.baseURL), bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("evaluation failed with status: %d", resp.StatusCode)
	}

	var evalResp types.EvaluationResponse
	if err := json.NewDecoder(resp.Body).Decode(&evalResp); err != nil {
		return nil, err
	}

	var value interface{}
	if err := json.Unmarshal(evalResp.Value, &value); err != nil {
		return nil, err
	}

	return value, nil
}

func (c *Client) startRefreshing() {
	ticker := time.NewTicker(c.config.RefreshInterval)
	defer ticker.Stop()

	c.refresh()

	for {
		select {
		case <-ticker.C:
			c.refresh()
		case <-c.stopCh:
			return
		}
	}
}

func (c *Client) refresh() {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/v1/flags?environment=%s", c.baseURL, c.config.Environment), nil)
	if err != nil {
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	var result struct {
		Flags []*types.Flag `json:"flags"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return
	}

	c.cacheMu.Lock()
	for _, flag := range result.Flags {
		c.cache[flag.Key] = flag
	}
	c.cacheMu.Unlock()
}

type UserContext struct {
	UserID     string
	UserKey    string
	Attributes map[string]interface{}
}

func NewUserContext(userID string) *UserContext {
	return &UserContext{
		UserID:     userID,
		UserKey:    userID,
		Attributes: make(map[string]interface{}),
	}
}

func (uc *UserContext) WithAttribute(key string, value interface{}) *UserContext {
	uc.Attributes[key] = value
	return uc
}