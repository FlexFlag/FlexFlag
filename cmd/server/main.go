package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flexflag/flexflag/internal/api/handlers"
	"github.com/flexflag/flexflag/internal/config"
	"github.com/flexflag/flexflag/internal/storage/postgres"
	"github.com/gin-gonic/gin"
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

	flagRepo := postgres.NewFlagRepository(db)
	flagHandler := handlers.NewFlagHandler(flagRepo)
	evaluationHandler := handlers.NewEvaluationHandler(flagRepo)
	optimizedEvalHandler := handlers.NewOptimizedEvaluationHandler(flagRepo)
	ultraFastHandler := handlers.NewUltraFastHandler(flagRepo)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "flexflag-server",
		})
	})

	api := r.Group("/api/v1")
	{
		// Flag management
		api.POST("/flags", flagHandler.CreateFlag)
		api.GET("/flags", flagHandler.ListFlags)
		api.GET("/flags/:key", flagHandler.GetFlag)
		api.PUT("/flags/:key", flagHandler.UpdateFlag)
		api.DELETE("/flags/:key", flagHandler.DeleteFlag)
		api.POST("/flags/:key/toggle", flagHandler.ToggleFlag)
		
		// Standard flag evaluation
		api.POST("/evaluate", evaluationHandler.Evaluate)
		api.POST("/evaluate/batch", evaluationHandler.BatchEvaluate)
		
		// Optimized flag evaluation (with caching)
		api.POST("/evaluate/fast", optimizedEvalHandler.FastEvaluate)
		api.POST("/evaluate/fast/batch", optimizedEvalHandler.FastBatchEvaluate)
		api.GET("/evaluate/cache/stats", optimizedEvalHandler.GetCacheStats)
		api.POST("/evaluate/cache/clear", optimizedEvalHandler.ClearCache)
		
		// Ultra-fast flag evaluation (with preloading and response caching)
		api.POST("/evaluate/ultra", ultraFastHandler.UltraFastEvaluate)
		api.GET("/evaluate/ultra/stats", ultraFastHandler.GetStats)
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