package main

import (
	"context"
	"messanger/pkg/config"
	"messanger/pkg/postgres"
	"os"

	"github.com/pressly/goose/v3"

	"go.uber.org/zap"
)

const (
	migrationsActionEnvKey = "MIGRATIONS_ACTION"

	migrationsActionUp   = "up"
	migrationsActionDown = "down"

	pathToMigrationsRequests = "./migration_requests"
	pathToConfig             = "../configs/local.json"
)

func main() {
	log := zap.NewExample()
	config, err := config.InitConfig(pathToConfig)
	if err != nil {
		log.Sugar().DPanicf("cant init config: %v", err)
	}

	ctx := context.Background()

	db, err := postgres.NewDB(ctx, config)
	if err != nil {
		log.Sugar().Fatalf("cant create connection for postgres: %v", err)
	}
	defer db.Close()

	switch os.Getenv(migrationsActionEnvKey) {
	case migrationsActionUp:
		if err := goose.Up(db, pathToMigrationsRequests); err != nil {
			log.Sugar().Fatalf("migration up is failed: %v", err)
		}
	case migrationsActionDown:
		if err := goose.Down(db, pathToMigrationsRequests); err != nil {
			log.Sugar().Fatalf("migration down is failed: %v", err)
		}
	default:
		log.Error("migrations action is unknown")
		return
	}
	log.Info("migrations is complete")
}
