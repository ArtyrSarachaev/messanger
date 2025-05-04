package main

import (
	"context"
	"messanger/internal/app"
	"messanger/pkg/config"
	"path"
)

func main() {
	pathToConfig := path.Join("..", "configs", "local.json")
	ctx := context.Background()
	config, err := config.InitConfig(ctx, pathToConfig)
	if err != nil {
		panic(err)
	}
	app.Run(ctx, config)
}
