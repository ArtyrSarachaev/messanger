package main

import (
	"context"
	"messanger/internal/app"
	"messanger/pkg/config"
	"path"
)

func main() {
	// тут бы получить путь к конфигу из env
	// configPath := os.Getenv("CONFIG_PATH")
	// а лучше даже через флаги, это универсальная история для всех ОС
	// https://gobyexample.com/command-line-flags
	pathToConfig := path.Join("..", "configs", "local.json")
	ctx := context.Background()

	// переменная одинаково называется с именем импортируемого пакета
	config, err := config.InitConfig(ctx, pathToConfig)
	if err != nil {
		panic(err)
	}

	app.Run(ctx, config)
}
