package handlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/flexflag/flexflag/internal/storage"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// ConfigHandler serves remote config values — all non-boolean flags returned as a flat key-value map.
// Designed for mobile SDK startup: one request to get all config values.
type ConfigHandler struct {
	repo storage.FlagRepository
}

func NewConfigHandler(repo storage.FlagRepository) *ConfigHandler {
	return &ConfigHandler{repo: repo}
}

// ConfigMeta holds lightweight metadata about each config entry.
type ConfigMeta struct {
	Type              types.FlagType         `json:"type"`
	Enabled           bool                   `json:"enabled"`
	UpdatedAt         time.Time              `json:"updated_at"`
	Platforms         []string               `json:"platforms,omitempty"`          // nil = all platforms
	RolloutPercentage *int                   `json:"rollout_percentage,omitempty"` // nil = 100%
	PlatformValues    map[string]interface{} `json:"platform_values,omitempty"`    // per-platform overrides
}

// ConfigResponse is the response shape for GET /config and POST /config/evaluate.
type ConfigResponse struct {
	Config      map[string]interface{} `json:"config"`
	Metadata    map[string]ConfigMeta  `json:"metadata"`
	Environment string                 `json:"environment"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ConfigEvaluateRequest is the request shape for POST /config/evaluate.
// Platform and UserID are used for platform targeting and sticky percentage rollout.
type ConfigEvaluateRequest struct {
	UserID     string                 `json:"user_id"`
	UserKey    string                 `json:"user_key"`
	// Platform identifies the client platform: "ios", "android", or "web".
	// When set, flags restricted to specific platforms are filtered accordingly.
	Platform   string                 `json:"platform"`
	Attributes map[string]interface{} `json:"attributes"`
}

// GetConfig returns all enabled remote config values (string, number, json flag types)
// as a flat map. Boolean flags are excluded — use /evaluate for feature gates.
// Platform filtering and rollout percentages are NOT applied here (no user context).
//
// @Summary     Get all remote config values
// @Tags        config
// @Produce     json
// @Param       environment  query  string  false  "Environment (default: production)"
// @Param       project_id   query  string  false  "Project ID filter"
// @Success     200  {object}  ConfigResponse
// @Failure     500  {object}  map[string]string
// @Router      /config [get]
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")

	if apiEnv, exists := c.Get("api_key_environment"); exists {
		if env, ok := apiEnv.(string); ok && env != "" {
			environment = env
		}
	}
	if apiProject, exists := c.Get("api_key_project_id"); exists {
		if proj, ok := apiProject.(string); ok && proj != "" {
			projectID = proj
		}
	}

	flags, err := h.listFlags(c, projectID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch config"})
		return
	}

	config := make(map[string]interface{})
	metadata := make(map[string]ConfigMeta)

	for _, flag := range flags {
		if flag.Type == types.FlagTypeBoolean || flag.Type == types.FlagTypeVariant {
			continue
		}

		platforms, rolloutPct := extractConfigSettings(flag.Metadata)
		meta := ConfigMeta{
			Type:              flag.Type,
			Enabled:           flag.Enabled,
			UpdatedAt:         flag.UpdatedAt,
			Platforms:         platforms,
			RolloutPercentage: rolloutPct,
			PlatformValues:    extractPlatformValues(flag.Metadata),
		}
		metadata[flag.Key] = meta

		if !flag.Enabled {
			continue
		}

		var value interface{}
		if jsonErr := json.Unmarshal(flag.Default, &value); jsonErr == nil {
			config[flag.Key] = value
		}
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Config:      config,
		Metadata:    metadata,
		Environment: environment,
		Timestamp:   time.Now().UTC(),
	})
}

// EvaluateConfig returns config values evaluated for a specific user context.
// Applies in order:
//  1. Platform filtering — skips flags not targeting the client's platform
//  2. Percentage rollout — uses MD5(flagKey:userID) % 100 for sticky bucketing
//  3. Targeting rules — returns variation value for the first matching rule
//
// @Summary     Evaluate remote config for a user context
// @Tags        config
// @Accept      json
// @Produce     json
// @Param       environment  query  string                 false  "Environment (default: production)"
// @Param       project_id   query  string                 false  "Project ID"
// @Param       request      body   ConfigEvaluateRequest  true   "User context"
// @Success     200  {object}  ConfigResponse
// @Failure     500  {object}  map[string]string
// @Router      /config/evaluate [post]
func (h *ConfigHandler) EvaluateConfig(c *gin.Context) {
	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")

	if apiEnv, exists := c.Get("api_key_environment"); exists {
		if env, ok := apiEnv.(string); ok && env != "" {
			environment = env
		}
	}
	if apiProject, exists := c.Get("api_key_project_id"); exists {
		if proj, ok := apiProject.(string); ok && proj != "" {
			projectID = proj
		}
	}

	var req ConfigEvaluateRequest
	_ = c.ShouldBindJSON(&req)

	flags, err := h.listFlags(c, projectID, environment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch config"})
		return
	}

	config := make(map[string]interface{})
	metadata := make(map[string]ConfigMeta)

	for _, flag := range flags {
		if flag.Type == types.FlagTypeBoolean || flag.Type == types.FlagTypeVariant {
			continue
		}

		platforms, rolloutPct := extractConfigSettings(flag.Metadata)
		metadata[flag.Key] = ConfigMeta{
			Type:              flag.Type,
			Enabled:           flag.Enabled,
			UpdatedAt:         flag.UpdatedAt,
			Platforms:         platforms,
			RolloutPercentage: rolloutPct,
		}

		if !flag.Enabled {
			continue
		}

		// 1. Platform filtering
		if !platformAllowed(platforms, req.Platform) {
			continue
		}

		// 2. Percentage rollout — consistent per user via MD5 bucketing
		userID := req.UserID
		if userID == "" {
			userID = req.UserKey
		}
		if rolloutPct != nil && userID != "" {
			bucket := configHashToPercentage(flag.Key, userID)
			if bucket >= *rolloutPct {
				// User is outside the rollout — serve default
				var def interface{}
				if err := json.Unmarshal(flag.Default, &def); err == nil {
					config[flag.Key] = def
				}
				continue
			}
		}

		// 3. Targeting rules + default
		config[flag.Key] = h.resolveConfigValue(flag, req)
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Config:      config,
		Metadata:    metadata,
		Environment: environment,
		Timestamp:   time.Now().UTC(),
	})
}

// listFlags fetches flags scoped by project when a project ID is provided.
func (h *ConfigHandler) listFlags(c *gin.Context, projectID, environment string) ([]*types.Flag, error) {
	if projectID != "" {
		return h.repo.ListByProject(c.Request.Context(), projectID, environment)
	}
	return h.repo.List(c.Request.Context(), environment)
}

// extractPlatformValues reads per-platform value overrides from metadata.
func extractPlatformValues(meta map[string]interface{}) map[string]interface{} {
	if meta == nil {
		return nil
	}
	if raw, ok := meta["platform_values"]; ok {
		if pvMap, ok := raw.(map[string]interface{}); ok && len(pvMap) > 0 {
			return pvMap
		}
	}
	return nil
}

// extractConfigSettings reads platform list and rollout percentage from flag metadata.
// These are stored as: metadata["platforms"] = ["ios","android"] and metadata["rollout_percentage"] = 75
func extractConfigSettings(meta map[string]interface{}) (platforms []string, rolloutPct *int) {
	if meta == nil {
		return nil, nil
	}

	if raw, ok := meta["platforms"]; ok {
		switch v := raw.(type) {
		case []interface{}:
			for _, p := range v {
				if s, ok := p.(string); ok {
					platforms = append(platforms, s)
				}
			}
		case []string:
			platforms = v
		}
	}

	if raw, ok := meta["rollout_percentage"]; ok {
		switch v := raw.(type) {
		case float64:
			n := int(v)
			rolloutPct = &n
		case int:
			rolloutPct = &v
		}
	}

	return platforms, rolloutPct
}

// platformAllowed returns true when the flag has no platform restriction,
// or when the client's platform is in the flag's allowed list.
func platformAllowed(platforms []string, clientPlatform string) bool {
	if len(platforms) == 0 {
		return true // no restriction — all platforms
	}
	if clientPlatform == "" {
		return true // client didn't specify platform — serve it
	}
	for _, p := range platforms {
		if p == clientPlatform {
			return true
		}
	}
	return false
}

// configHashToPercentage returns a deterministic bucket (0–99) for the given
// flagKey + userID pair using MD5, matching the existing rollout evaluator approach.
func configHashToPercentage(flagKey, userID string) int {
	key := fmt.Sprintf("%s:%s", flagKey, userID)
	hash := md5.Sum([]byte(key))
	var hashInt uint32
	for i := 0; i < 4; i++ {
		hashInt = hashInt<<8 + uint32(hash[i])
	}
	return int(hashInt % 100)
}

// resolveConfigValue returns the value for a flag given the request context.
// Resolution order:
//  1. Per-platform value from metadata.platform_values (e.g. {"ios": "dark", "android": "light"})
//  2. Targeting rules (attribute-based)
//  3. Flag default value
func (h *ConfigHandler) resolveConfigValue(flag *types.Flag, req ConfigEvaluateRequest) interface{} {
	var defaultValue interface{}
	if err := json.Unmarshal(flag.Default, &defaultValue); err != nil {
		return nil
	}

	// 1. Per-platform value — check metadata.platform_values first
	if req.Platform != "" && flag.Metadata != nil {
		if rawPV, ok := flag.Metadata["platform_values"]; ok {
			if pvMap, ok := rawPV.(map[string]interface{}); ok {
				if platformVal, ok := pvMap[req.Platform]; ok {
					return platformVal
				}
			}
		}
	}

	// 2. Targeting rules
	if flag.Targeting == nil || len(flag.Targeting.Rules) == 0 {
		return defaultValue
	}

	attrs := map[string]interface{}{}
	if req.UserID != "" {
		attrs["user_id"] = req.UserID
	}
	if req.UserKey != "" {
		attrs["user_key"] = req.UserKey
	}
	if req.Platform != "" {
		attrs["platform"] = req.Platform
	}
	for k, v := range req.Attributes {
		attrs[k] = v
	}

	for _, rule := range flag.Targeting.Rules {
		attrVal, ok := attrs[rule.Attribute]
		if !ok {
			continue
		}
		if configMatchesRule(configValToString(attrVal), rule.Operator, rule.Values) {
			for _, variation := range flag.Variations {
				if variation.ID == rule.Variation || variation.Name == rule.Variation {
					var varValue interface{}
					if err := json.Unmarshal(variation.Value, &varValue); err == nil {
						return varValue
					}
				}
			}
		}
	}

	// 3. Default
	return defaultValue
}

func configValToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func configMatchesRule(attrVal, operator string, values []string) bool {
	switch operator {
	case "eq":
		for _, v := range values {
			if attrVal == v {
				return true
			}
		}
	case "ne":
		for _, v := range values {
			if attrVal == v {
				return false
			}
		}
		return len(values) > 0
	case "in":
		for _, v := range values {
			if attrVal == v {
				return true
			}
		}
	case "nin":
		for _, v := range values {
			if attrVal == v {
				return false
			}
		}
		return true
	case "contains":
		for _, v := range values {
			if len(v) > 0 && len(attrVal) >= len(v) {
				for i := 0; i <= len(attrVal)-len(v); i++ {
					if attrVal[i:i+len(v)] == v {
						return true
					}
				}
			}
		}
	}
	return false
}
