package main

import (
	"messanger/internal/app"
	"messanger/pkg/config"
)

const (
	pathToConfig = "../configs/local.json"
)

func main() {
	config, err := config.InitConfig(pathToConfig)
	if err != nil {
		panic(err)
	}
	app.Run(config)
}
