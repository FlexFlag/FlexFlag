package handlers

import (
	"encoding/json"
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
	Type      types.FlagType `json:"type"`
	Enabled   bool           `json:"enabled"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// ConfigResponse is the response shape for GET /config and POST /config/evaluate.
type ConfigResponse struct {
	Config      map[string]interface{} `json:"config"`
	Metadata    map[string]ConfigMeta  `json:"metadata"`
	Environment string                 `json:"environment"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ConfigEvaluateRequest is the request shape for POST /config/evaluate.
type ConfigEvaluateRequest struct {
	UserID     string                 `json:"user_id"`
	UserKey    string                 `json:"user_key"`
	Attributes map[string]interface{} `json:"attributes"`
}

// GetConfig returns all enabled remote config values (string, number, json flag types)
// as a flat map. Boolean flags are excluded — use /evaluate for feature gates.
//
// @Summary     Get all remote config values
// @Tags        config
// @Produce     json
// @Param       environment  query     string  false  "Environment (default: production)"
// @Param       project_id   query     string  false  "Project ID filter"
// @Success     200  {object}  ConfigResponse
// @Failure     500  {object}  map[string]string
// @Router      /config [get]
func (h *ConfigHandler) GetConfig(c *gin.Context) {
	environment := c.DefaultQuery("environment", "production")
	projectID := c.Query("project_id")

	// API key middleware may have already resolved environment/project
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
		// Remote config = non-boolean flag types only
		if flag.Type == types.FlagTypeBoolean || flag.Type == types.FlagTypeVariant {
			continue
		}

		var value interface{}
		if jsonErr := json.Unmarshal(flag.Default, &value); jsonErr != nil {
			value = nil
		}

		// Only include enabled flags in the config map; disabled ones appear in metadata only
		if flag.Enabled {
			config[flag.Key] = value
		}

		metadata[flag.Key] = ConfigMeta{
			Type:      flag.Type,
			Enabled:   flag.Enabled,
			UpdatedAt: flag.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, ConfigResponse{
		Config:      config,
		Metadata:    metadata,
		Environment: environment,
		Timestamp:   time.Now().UTC(),
	})
}

// EvaluateConfig returns config values evaluated for a specific user context,
// applying targeting rules (segments, attributes) if configured on the flag.
//
// @Summary     Evaluate remote config for a user context
// @Tags        config
// @Accept      json
// @Produce     json
// @Param       environment  query     string                 false  "Environment (default: production)"
// @Param       project_id   query     string                 false  "Project ID"
// @Param       request      body      ConfigEvaluateRequest  true   "User context"
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
	// No body is fine — evaluate with no user context
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

		metadata[flag.Key] = ConfigMeta{
			Type:      flag.Type,
			Enabled:   flag.Enabled,
			UpdatedAt: flag.UpdatedAt,
		}

		if !flag.Enabled {
			continue
		}

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

// resolveConfigValue applies targeting rules for a flag and returns the appropriate value.
// Falls back to the flag default if no rule matches or targeting is not configured.
func (h *ConfigHandler) resolveConfigValue(flag *types.Flag, req ConfigEvaluateRequest) interface{} {
	var defaultValue interface{}
	if err := json.Unmarshal(flag.Default, &defaultValue); err != nil {
		return nil
	}

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
