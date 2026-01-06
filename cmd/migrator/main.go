package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func buildDatabaseURL() string {
	host := getEnv("FLEXFLAG_DATABASE_HOST", "localhost")
	port := getEnv("FLEXFLAG_DATABASE_PORT", "5432")
	user := getEnv("FLEXFLAG_DATABASE_USERNAME", "flexflag")
	password := getEnv("FLEXFLAG_DATABASE_PASSWORD", "flexflag")
	dbname := getEnv("FLEXFLAG_DATABASE_DATABASE", "flexflag")
	sslmode := getEnv("FLEXFLAG_DATABASE_SSL_MODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

func main() {
	defaultDatabaseURL := buildDatabaseURL()
	var databaseURL = flag.String("database-url", defaultDatabaseURL, "Database URL")
	var migrationsPath = flag.String("migrations-path", "file://migrations", "Path to migrations directory")
	var direction = flag.String("direction", "up", "Migration direction: up or down")
	var forceVersion = flag.Int("force-version", -1, "Force migration version (use to fix dirty state)")
	flag.Parse()

	log.Printf("Connecting to database at %s", *databaseURL)

	db, err := sql.Open("postgres", *databaseURL)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create database driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(*migrationsPath, "postgres", driver)
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	// Force version if specified (to fix dirty state)
	if *forceVersion >= 0 {
		if err := m.Force(*forceVersion); err != nil {
			log.Fatalf("Could not force version: %v", err)
		}
		log.Printf("Forced migration version to %d", *forceVersion)
		return
	}

	switch *direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Could not run migrations: %v", err)
		}
		log.Println("Migrations applied successfully")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Could not rollback migrations: %v", err)
		}
		log.Println("Migrations rolled back successfully")
	default:
		log.Fatalf("Invalid direction: %s. Use 'up' or 'down'", *direction)
	}
}