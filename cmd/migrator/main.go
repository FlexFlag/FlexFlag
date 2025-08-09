package main

import (
	"database/sql"
	"flag"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	var databaseURL = flag.String("database-url", "postgres://localhost/flexflag?sslmode=disable", "Database URL")
	var migrationsPath = flag.String("migrations-path", "file://migrations", "Path to migrations directory")
	var direction = flag.String("direction", "up", "Migration direction: up or down")
	flag.Parse()

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