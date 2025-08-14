package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flexflag/flexflag/internal/edge"
	"github.com/flexflag/flexflag/pkg/types"
	"github.com/gin-gonic/gin"
)

// EdgeServer represents a distributed edge server for ultra-fast flag evaluation
type EdgeServer struct {
	cache      *edge.FlagCache
	syncClient *edge.SyncClient
	config     *edge.Config
	router     *gin.Engine
}

func main() {
	// Load configuration
	config := edge.LoadConfig()
	
	// Initialize edge server components
	cache := edge.NewFlagCache(config.CacheConfig)
	syncClient := edge.NewSyncClient(config.HubURL, config.APIKey)
	syncClient.SetCache(cache)
	syncClient.SetConfig(config.SyncConfig)
	
	server := &EdgeServer{
		cache:      cache,
		syncClient: syncClient,
		config:     config,
	}
	
	// Setup HTTP router
	server.setupRouter()
	
	// Start synchronization with central hub
	go server.startSync()
	
	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Port),
		Handler: server.router,
	}
	
	go func() {
		log.Printf("Edge server starting on port %d", config.Port)
		log.Printf("Connected to hub: %s", config.HubURL)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down edge server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	server.syncClient.Close()
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	
	log.Println("Edge server stopped")
}

func (s *EdgeServer) setupRouter() {
	if s.config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(s.corsMiddleware())
	
	// Health check endpoints (no auth required)
	r.GET("/health", s.healthCheck)
	r.GET("/ready", s.readinessCheck)
	
	// API endpoints (auth required)
	authenticated := r.Group("/")
	authenticated.Use(s.authMiddleware())
	
	// Edge evaluation endpoints (ultra-fast) - require auth
	api := authenticated.Group("/api/v1")
	{
		api.POST("/evaluate", s.evaluateFlag)
		api.POST("/evaluate/batch", s.batchEvaluate)
		api.GET("/cache/stats", s.getCacheStats)
		api.POST("/cache/refresh", s.refreshCache) // Admin endpoint
	}
	
	s.router = r
}

func (s *EdgeServer) evaluateFlag(c *gin.Context) {
	startTime := time.Now()
	
	var req types.EvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Extract environment from API key context (set by auth middleware)
	environment := c.GetString("environment")
	if environment == "" {
		environment = "production" // Default fallback
	}
	
	// Try local cache first (L1 - ultra fast)
	flag := s.cache.GetFlag(req.FlagKey, environment)
	if flag == nil {
		// Cache miss - this should be rare with proper sync
		c.JSON(http.StatusNotFound, gin.H{
			"error": "flag not found",
			"cache": "miss",
		})
		return
	}
	
	// Evaluate flag locally
	result := s.evaluateLocally(flag, &req)
	result.EvaluationTime = float64(time.Since(startTime).Microseconds()) / 1000.0
	result.Source = "edge-cache"
	
	c.JSON(http.StatusOK, result)
}

func (s *EdgeServer) batchEvaluate(c *gin.Context) {
	startTime := time.Now()
	
	var req struct {
		FlagKeys   []string               `json:"flag_keys" binding:"required"`
		UserID     string                 `json:"user_id"`
		UserKey    string                 `json:"user_key"`
		Attributes map[string]interface{} `json:"attributes"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	environment := c.GetString("environment")
	if environment == "" {
		environment = "production"
	}
	
	results := make(map[string]interface{})
	
	for _, flagKey := range req.FlagKeys {
		flag := s.cache.GetFlag(flagKey, environment)
		if flag == nil {
			results[flagKey] = map[string]interface{}{
				"error": "flag not found",
			}
			continue
		}
		
		evalReq := &types.EvaluationRequest{
			FlagKey:    flagKey,
			UserID:     req.UserID,
			UserKey:    req.UserKey,
			Attributes: req.Attributes,
		}
		
		result := s.evaluateLocally(flag, evalReq)
		results[flagKey] = result
	}
	
	totalTime := float64(time.Since(startTime).Microseconds()) / 1000.0
	
	c.JSON(http.StatusOK, gin.H{
		"results":          results,
		"evaluation_time":  totalTime,
		"flags_evaluated":  len(req.FlagKeys),
		"source":          "edge-cache",
	})
}

func (s *EdgeServer) evaluateLocally(flag *types.Flag, req *types.EvaluationRequest) *types.EvaluationResponse {
	// Simple evaluation logic - in production, use the full evaluation engine
	value := flag.Default
	if len(value) == 0 {
		value = json.RawMessage(`null`)
	}
	
	return &types.EvaluationResponse{
		FlagKey:   flag.Key,
		Value:     value,
		Reason:    "default_value",
		Default:   true,
		Timestamp: time.Now(),
	}
}

func (s *EdgeServer) getCacheStats(c *gin.Context) {
	stats := s.cache.GetStats()
	c.JSON(http.StatusOK, stats)
}

func (s *EdgeServer) refreshCache(c *gin.Context) {
	// Admin endpoint to force cache refresh
	if err := s.syncClient.FullSync(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Cache refreshed successfully"})
}

func (s *EdgeServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "flexflag-edge",
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
	})
}

func (s *EdgeServer) readinessCheck(c *gin.Context) {
	// Check if cache is populated and sync is working
	stats := s.cache.GetStats()
	isReady := stats.FlagCount > 0 && s.syncClient.IsConnected()
	
	status := http.StatusOK
	if !isReady {
		status = http.StatusServiceUnavailable
	}
	
	c.JSON(status, gin.H{
		"ready":      isReady,
		"cache_size": stats.FlagCount,
		"sync_status": map[string]interface{}{
			"connected":  s.syncClient.IsConnected(),
			"last_sync":  s.syncClient.LastSyncTime(),
		},
	})
}

func (s *EdgeServer) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

func (s *EdgeServer) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}
		
		log.Printf("DEBUG: Received API key: %s", apiKey[:16]+"****")
		
		// First try local cache
		keyInfo := s.cache.ValidateAPIKey(apiKey)
		log.Printf("DEBUG: Cache lookup result: %v", keyInfo != nil)
		
		if keyInfo == nil {
			log.Printf("DEBUG: Cache miss, trying fallback authentication")
			// Cache miss - authenticate with central hub
			keyInfo = s.authenticateWithHub(apiKey)
			log.Printf("DEBUG: Fallback authentication result: %v", keyInfo != nil)
			if keyInfo == nil {
				log.Printf("DEBUG: API key authentication failed")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
				c.Abort()
				return
			}
			// Cache the validated key for future use
			log.Printf("DEBUG: Caching validated API key")
			s.cache.UpdateAPIKey(apiKey, keyInfo)
		}
		
		log.Printf("DEBUG: API key authenticated successfully, environment: %s", keyInfo.Environment)
		
		// Set context for evaluation
		c.Set("environment", keyInfo.Environment)
		c.Set("project_id", keyInfo.ProjectID)
		c.Set("permissions", keyInfo.Permissions)
		
		c.Next()
	}
}

// authenticateWithHub validates API key with the central hub
func (s *EdgeServer) authenticateWithHub(apiKey string) *edge.APIKeyInfo {
	log.Printf("DEBUG: Starting fallback authentication for API key")
	
	// Use the main FlexFlag server's existing API key authentication
	client := &http.Client{Timeout: 5 * time.Second}
	
	// Create authentication request to main server  
	hubURL := s.config.HubURL
	if hubURL == "" {
		hubURL = "http://localhost:8080"
	}
	
	// Convert WebSocket URL to HTTP URL for fallback authentication
	if strings.HasPrefix(hubURL, "ws://") {
		hubURL = strings.Replace(hubURL, "ws://", "http://", 1)
	} else if strings.HasPrefix(hubURL, "wss://") {
		hubURL = strings.Replace(hubURL, "wss://", "https://", 1)
	}
	
	log.Printf("DEBUG: Hub URL (converted to HTTP): %s", hubURL)
	
	// Use the existing evaluation endpoint which already handles API key auth
	reqBody := `{"flag_key":"__health_check__","user_id":"edge-server"}`
	req, err := http.NewRequest("POST", hubURL+"/api/v1/evaluate", strings.NewReader(reqBody))
	if err != nil {
		log.Printf("DEBUG: Failed to create request: %v", err)
		return nil
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", apiKey)
	
	log.Printf("DEBUG: Making fallback request to %s", hubURL+"/api/v1/evaluate")
	
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("DEBUG: Fallback request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()
	
	log.Printf("DEBUG: Fallback response status: %d", resp.StatusCode)
	
	if resp.StatusCode == 401 {
		log.Printf("DEBUG: API key rejected by main server")
		return nil // Invalid API key
	}
	
	if resp.StatusCode == 404 || resp.StatusCode == 200 {
		// API key is valid (404 just means flag not found, but auth worked)
		// For now, return a basic key info - in production this would extract more details
		log.Printf("DEBUG: API key accepted by main server")
		return &edge.APIKeyInfo{
			ProjectID:   "unknown", // Would extract from response or separate call
			Environment: "production", // Would extract from API key prefix  
			Permissions: []string{"read"},
		}
	}
	
	log.Printf("DEBUG: Unexpected response status: %d", resp.StatusCode)
	return nil
}

func (s *EdgeServer) startSync() {
	// Initial full synchronization
	if err := s.syncClient.FullSync(); err != nil {
		log.Printf("Initial sync failed: %v", err)
	}
	
	// Start real-time sync
	s.syncClient.StartRealtimeSync(func(update *edge.FlagUpdate) {
		s.cache.UpdateFlag(update)
		log.Printf("Flag updated: %s in %s", update.FlagKey, update.Environment)
	})
}