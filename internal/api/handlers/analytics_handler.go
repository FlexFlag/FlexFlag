package handlers

import (
	"net/http"

	"github.com/flexflag/flexflag/internal/analytics"
	"github.com/gin-gonic/gin"
)

// AnalyticsHandler exposes in-memory flag evaluation metrics.
type AnalyticsHandler struct {
	counter *analytics.Counter
}

func NewAnalyticsHandler(counter *analytics.Counter) *AnalyticsHandler {
	return &AnalyticsHandler{counter: counter}
}

// GetFlagAnalytics returns evaluation counts, unique user counts, and
// variation distribution for all flags in the requested environment.
//
// @Summary     Get flag evaluation analytics
// @Tags        analytics
// @Produce     json
// @Param       environment  query  string  false  "Environment (default: production)"
// @Success     200  {object}  map[string]analytics.FlagStats
// @Router      /analytics/flags [get]
func (h *AnalyticsHandler) GetFlagAnalytics(c *gin.Context) {
	environment := c.DefaultQuery("environment", "production")
	snapshot := h.counter.Snapshot(environment)
	c.JSON(http.StatusOK, gin.H{
		"flags":       snapshot,
		"environment": environment,
	})
}
