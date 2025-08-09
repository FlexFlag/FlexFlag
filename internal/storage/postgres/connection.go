package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/flexflag/flexflag/internal/config"
	_ "github.com/lib/pq"
)

func NewConnection(cfg *config.DatabaseConfig) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database,
		cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Optimized connection pool settings
	db.SetMaxOpenConns(cfg.MaxConns * 2) // Allow more connections for high load
	db.SetMaxIdleConns(cfg.MaxConns)     // Keep more idle connections ready
	db.SetConnMaxLifetime(5 * time.Minute) // Rotate connections every 5 minutes
	db.SetConnMaxIdleTime(30 * time.Second) // Close idle connections after 30 seconds

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}