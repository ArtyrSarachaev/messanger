package config

import (
	"context"
	"encoding/json"
	"os"
)

const (
	envKey = "env"
)

func InitConfig(ctx context.Context, pathToConfig string) (Config, error) {
	config := Config{}
	data, err := os.ReadFile(pathToConfig)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	ctx = context.WithValue(ctx, envKey, os.Getenv("ENV"))

	return config, nil
}
