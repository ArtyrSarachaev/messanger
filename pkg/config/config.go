package config

import (
	"encoding/json"
	"os"
)

const (
	localEnvType = "local"

	envKey = "ENV"
)

func InitConfig(configPath string) (Config, error) {
	config := Config{}
	switch os.Getenv(envKey) {
	case localEnvType:
		data, err := os.ReadFile(configPath)
		if err != nil {
			return config, err
		}

		err = json.Unmarshal(data, &config)
		if err != nil {
			return config, err
		}
	}

	return config, nil
}
