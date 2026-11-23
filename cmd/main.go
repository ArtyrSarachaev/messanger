package main

import (
	"os"

	"messanger/internal/app"
	"messanger/pkg/config"
)

func main() {
	log := app.NewLogger()

	cfg, err := config.Load()
	if err != nil {
		log.Error("load config", "error", err)
		os.Exit(1)
	}

	if err := app.Run(cfg, log); err != nil {
		log.Error("application stopped with error", "error", err)
		os.Exit(1)
	}
}
