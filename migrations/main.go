package main

import (
	"embed"
	"log"
	"os"

	"messanger/pkg/config"
	"messanger/pkg/postgres"

	"github.com/pressly/goose/v3"
)

const (
	migrationsActionEnvKey = "MIGRATIONS_ACTION"
	migrationsDir          = "migration_requests"
)

// SQL-файлы вшиваются в бинарник, поэтому мигратор не зависит от рабочей директории.
//
//go:embed migration_requests/*.sql
var migrationsFS embed.FS

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := postgres.NewDB(cfg.Postgres)
	if err != nil {
		log.Fatalf("connect to postgres: %v", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrationsFS)
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}

	action := os.Getenv(migrationsActionEnvKey)
	if action == "" {
		action = "up"
	}

	switch action {
	case "up":
		err = goose.Up(db, migrationsDir)
	case "down":
		err = goose.Down(db, migrationsDir)
	case "status":
		err = goose.Status(db, migrationsDir)
	default:
		log.Fatalf("unknown %s=%q, expected up|down|status", migrationsActionEnvKey, action)
	}
	if err != nil {
		log.Fatalf("migrations %s failed: %v", action, err)
	}

	log.Printf("migrations %s completed", action)
}
