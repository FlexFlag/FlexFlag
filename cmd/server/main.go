// Package main FlexFlag API Server
// @title           FlexFlag API
// @version         1.0
// @description     A high-performance feature flag management system with project-based multi-tenancy
// @termsOfService  http://swagger.io/terms/

// @contact.name   FlexFlag Support
// @contact.url    http://www.flexflag.com/support
// @contact.email  support@flexflag.com

// @license.name  MIT
// @license.url   http://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Bearer token authentication

// @securityDefinitions.apikey ApiKeyHeader
// @in header
// @name X-API-Key
// @description API Key authentication

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flexflag/flexflag/internal/analytics"
	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/internal/api/middleware"
	"github.com/flexflag/flexflag/internal/auth"
	"github.com/flexflag/flexflag/internal/config"
	"github.com/flexflag/flexflag/internal/services"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/gin-gonic/gin"
	
	// Swagger imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "github.com/flexflag/flexflag/api" // Generated swagger docs
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config: %v. Using defaults.", err)
		cfg = &config.Config{
			Server: config.ServerConfig{
				Host: "0.0.0.0",
				Port: 8080,
			},
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5433,
				Username: "flexflag",
				Password: "flexflag",
				Database: "flexflag",
				SSLMode:  "disable",
				MaxConns: 10,
				MinConns: 2,
			},
		}
	}

	db, err := postgres.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	flagRepo := postgres.NewFlagRepository(db)
	userRepo := postgres.NewUserRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	segmentRepo := postgres.NewSegmentRepository(db)
	rolloutRepo := postgres.NewRolloutRepository(db)
	auditRepo := postgres.NewAuditRepository(db)
	apiKeyRepo := postgres.NewApiKeyRepository(db)
	
	// Initialize services
	auditService := services.NewAuditService(auditRepo)
	
	// Initialize auth
	jwtManager := auth.NewJWTManager("your-secret-key-here", 24*time.Hour)
	
	// Initialize analytics counter (shared across all eval handlers)
	analyticsCounter := analytics.NewCounter()

	// Initialize handlers
	ultraFastHandler := handlers.NewUltraFastHandler(flagRepo)
	edgeSyncHandler := handlers.NewEdgeSyncHandler(flagRepo, apiKeyRepo)
	sseHandler := handlers.NewSSEHandler()
	clientSSEHandler := handlers.NewClientSSEHandler(apiKeyRepo)
	flagHandler := handlers.NewFlagHandler(flagRepo, auditService, ultraFastHandler, projectRepo)
	flagHandler.SetEdgeSyncHandler(edgeSyncHandler)
	flagHandler.SetSSEHandler(sseHandler)
	flagHandler.SetClientSSEHandler(clientSSEHandler)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)
	userHandler := handlers.NewUserHandler(userRepo)
	oauthHandler := handlers.NewOAuthHandler(
		userRepo,
		jwtManager,
		cfg.Auth.OAuth.Google.ClientID,
		cfg.Auth.OAuth.Google.ClientSecret,
		cfg.Auth.OAuth.Google.RedirectURL,
	)
	projectHandler := handlers.NewProjectHandler(projectRepo, flagRepo, segmentRepo, rolloutRepo)
	segmentHandler := handlers.NewSegmentHandler(segmentRepo)
	rolloutHandler := handlers.NewRolloutHandler(rolloutRepo)
	auditHandler := handlers.NewAuditHandler(auditRepo)
	evaluationHandler := handlers.NewEvaluationHandler(flagRepo, rolloutRepo, analyticsCounter)
	optimizedEvalHandler := handlers.NewOptimizedEvaluationHandler(flagRepo, analyticsCounter)
	apiKeyHandler := handlers.NewApiKeyHandler(apiKeyRepo)
	configHandler := handlers.NewConfigHandler(flagRepo)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsCounter)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "flexflag-server",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := r.Group("/api/v1")
	{
		// Authentication endpoints (public)
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/refresh", authHandler.RefreshToken)

		// OAuth endpoints (public)
		api.GET("/auth/google/login", oauthHandler.GoogleLogin)
		api.GET("/auth/google/callback", oauthHandler.GoogleCallback)
		
		// Protected authentication endpoints
		authGroup := api.Group("/auth")
		authGroup.Use(auth.AuthMiddleware(jwtManager))
		{
			authGroup.GET("/profile", authHandler.GetProfile)
			authGroup.PUT("/profile", authHandler.UpdateProfile)
			authGroup.POST("/change-password", authHandler.ChangePassword)
			authGroup.POST("/logout", authHandler.Logout)
		}
		
		// Project statistics and members endpoints (must come before projects group to avoid conflicts)
		api.GET("/project-stats/:id", auth.AuthMiddleware(jwtManager), projectHandler.GetProjectStats)
		api.GET("/project-members/:id", auth.AuthMiddleware(jwtManager), projectHandler.GetProjectMembers)
		api.POST("/project-members/:id", auth.AuthMiddleware(jwtManager), auth.RequireEditorOrAdmin(), projectHandler.AddProjectMember)
		api.DELETE("/project-members/:id/:userId", auth.AuthMiddleware(jwtManager), auth.RequireEditorOrAdmin(), projectHandler.RemoveProjectMember)
		api.PUT("/project-members/:id/:userId", auth.AuthMiddleware(jwtManager), auth.RequireEditorOrAdmin(), projectHandler.UpdateProjectMemberRole)

		// Project management (require authentication)
		projects := api.Group("/projects")
		projects.Use(auth.AuthMiddleware(jwtManager))
		{
			projects.POST("", auth.RequireEditorOrAdmin(), projectHandler.CreateProject)
			projects.GET("", projectHandler.ListProjects)
			projects.GET("/:slug", projectHandler.GetProject)
			projects.PUT("/:slug", auth.RequireEditorOrAdmin(), projectHandler.UpdateProject)
			projects.DELETE("/:slug", auth.RequireAdmin(), projectHandler.DeleteProject)

			// Environment endpoints
			projects.POST("/:slug/environments", auth.RequireEditorOrAdmin(), projectHandler.CreateEnvironment)
			projects.GET("/:slug/environments", projectHandler.GetEnvironments)
		}
		
		// API Key endpoints (separate group to avoid route conflicts)
		apiKeys := api.Group("/project-api-keys")
		apiKeys.Use(auth.AuthMiddleware(jwtManager))
		{
			apiKeys.POST("/:projectId", auth.RequireEditorOrAdmin(), apiKeyHandler.CreateApiKey)
			apiKeys.GET("/:projectId", apiKeyHandler.GetApiKeys)
			apiKeys.PUT("/:projectId/:keyId", auth.RequireEditorOrAdmin(), apiKeyHandler.UpdateApiKey)
			apiKeys.DELETE("/:projectId/:keyId", auth.RequireEditorOrAdmin(), apiKeyHandler.DeleteApiKey)
		}
		
		// Environment management (require authentication) - separate group for individual environment operations
		environments := api.Group("/environments")
		environments.Use(auth.AuthMiddleware(jwtManager))
		{
			environments.PUT("/:id", auth.RequireEditorOrAdmin(), projectHandler.UpdateEnvironment)
			environments.DELETE("/:id", auth.RequireEditorOrAdmin(), projectHandler.DeleteEnvironment)
		}
		
		// Segment management (require authentication)
		segments := api.Group("/segments")
		segments.Use(auth.AuthMiddleware(jwtManager))
		{
			segments.POST("", auth.RequireEditorOrAdmin(), segmentHandler.CreateSegment)
			segments.GET("", segmentHandler.ListSegments)
			segments.GET("/:key", segmentHandler.GetSegment)
			segments.PUT("/:key", auth.RequireEditorOrAdmin(), segmentHandler.UpdateSegment)
			segments.DELETE("/:key", auth.RequireEditorOrAdmin(), segmentHandler.DeleteSegment)
			segments.POST("/evaluate", segmentHandler.EvaluateSegment)
		}
		
		// Rollout management (require authentication)
		rollouts := api.Group("/rollouts")
		rollouts.Use(auth.AuthMiddleware(jwtManager))
		{
			rollouts.POST("", auth.RequireEditorOrAdmin(), rolloutHandler.CreateRollout)
			rollouts.GET("", func(c *gin.Context) {
				// Check if flag_id is provided to determine which handler to use
				flagID := c.Query("flag_id")
				if flagID != "" {
					rolloutHandler.GetRolloutsByFlag(c)
				} else {
					rolloutHandler.GetAllRollouts(c)
				}
			})
			rollouts.GET("/:id", rolloutHandler.GetRollout)
			rollouts.PUT("/:id", auth.RequireEditorOrAdmin(), rolloutHandler.UpdateRollout)
			rollouts.DELETE("/:id", auth.RequireEditorOrAdmin(), rolloutHandler.DeleteRollout)
			rollouts.POST("/:id/activate", auth.RequireEditorOrAdmin(), rolloutHandler.ActivateRollout)
			rollouts.POST("/:id/pause", auth.RequireEditorOrAdmin(), rolloutHandler.PauseRollout)
			rollouts.POST("/:id/complete", auth.RequireEditorOrAdmin(), rolloutHandler.CompleteRollout)
			rollouts.POST("/:id/evaluate", rolloutHandler.EvaluateRollout)
		}
		
		// Sticky assignment management (require authentication)
		assignments := api.Group("/assignments")
		assignments.Use(auth.AuthMiddleware(jwtManager))
		{
			assignments.GET("/sticky", rolloutHandler.GetStickyAssignments)
			assignments.DELETE("/sticky", auth.RequireEditorOrAdmin(), rolloutHandler.DeleteStickyAssignment)
			assignments.POST("/cleanup", auth.RequireEditorOrAdmin(), rolloutHandler.CleanupExpiredAssignments)
		}
		
		// Flag management (require authentication)
		flags := api.Group("/flags")
		flags.Use(auth.AuthMiddleware(jwtManager))
		{
			flags.POST("", auth.RequireEditorOrAdmin(), flagHandler.CreateFlag)
			flags.GET("", flagHandler.ListFlags)
			flags.GET("/:key", flagHandler.GetFlag)
			flags.PUT("/:key", auth.RequireEditorOrAdmin(), flagHandler.UpdateFlag)
			flags.DELETE("/:key", auth.RequireEditorOrAdmin(), flagHandler.DeleteFlag)
			flags.POST("/:key/toggle", auth.RequireEditorOrAdmin(), flagHandler.ToggleFlag)
		}
		
		// User management (require authentication and admin role)
		users := api.Group("/users")
		users.Use(auth.AuthMiddleware(jwtManager))
		{
			users.POST("", auth.RequireAdmin(), userHandler.CreateUser)
			users.GET("", auth.RequireAdmin(), userHandler.ListUsers)
			users.GET("/:id", auth.RequireAdmin(), userHandler.GetUser)
			users.PUT("/:id", auth.RequireAdmin(), userHandler.UpdateUser)
			users.DELETE("/:id", auth.RequireAdmin(), userHandler.DeleteUser)
			users.POST("/:id/reset-password", auth.RequireAdmin(), userHandler.ResetPassword)
		}

		// Audit logs (require authentication)
		audit := api.Group("/audit")
		audit.Use(auth.AuthMiddleware(jwtManager))
		{
			audit.GET("/logs", auditHandler.ListAuditLogs)
			audit.GET("/logs/resource", auditHandler.ListAuditLogsByResource)
		}
		
		// Flag evaluation (supports both JWT and API key authentication)
		api.POST("/evaluate", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), evaluationHandler.Evaluate)
		api.POST("/evaluate/batch", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), evaluationHandler.BatchEvaluate)
		
		// Optimized flag evaluation
		api.POST("/evaluate/fast", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), optimizedEvalHandler.FastEvaluate)
		api.POST("/evaluate/fast/batch", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), optimizedEvalHandler.FastBatchEvaluate)
		
		// Cache management (require authentication)
		cache := api.Group("/evaluate/cache")
		cache.Use(auth.AuthMiddleware(jwtManager))
		{
			cache.GET("/stats", optimizedEvalHandler.GetCacheStats)
			cache.POST("/clear", auth.RequireEditorOrAdmin(), optimizedEvalHandler.ClearCache)
		}
		
		// Ultra-fast flag evaluation
		api.POST("/evaluate/ultra", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), ultraFastHandler.UltraFastEvaluate)
		api.GET("/evaluate/ultra/stats", auth.AuthMiddleware(jwtManager), ultraFastHandler.GetStats)
		
		// Edge server synchronization endpoints
		edge := api.Group("/edge")
		{
			edge.GET("/sync", edgeSyncHandler.BulkSync)
			edge.GET("/sync/ws", edgeSyncHandler.WebSocketSync)
			edge.GET("/sync/sse", sseHandler.HandleSSE)
			edge.POST("/auth", edgeSyncHandler.AuthenticateAPIKey)
			edge.GET("/servers", auth.AuthMiddleware(jwtManager), sseHandler.HandleEdgeServerStatus)
		}

		// Remote Config endpoints (supports both JWT and API key authentication)
		// GET  /config          — all config values as flat map (for SDK startup)
		// POST /config/evaluate — config values evaluated for a specific user context
		api.GET("/config", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), configHandler.GetConfig)
		api.POST("/config/evaluate", middleware.OptionalApiKeyAuth(apiKeyRepo), auth.OptionalAuth(jwtManager), configHandler.EvaluateConfig)

		// Analytics — in-memory evaluation counters (resets on restart)
		api.GET("/analytics/flags", auth.AuthMiddleware(jwtManager), analyticsHandler.GetFlagAnalytics)

		// Client SDK SSE stream (uses API key authentication via query parameter)
		api.GET("/stream", clientSSEHandler.HandleClientSSE)
		api.GET("/stream/clients", auth.AuthMiddleware(jwtManager), clientSSEHandler.HandleClientStatus)
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	log.Printf("Server started on %s", srv.Addr)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}