package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultConfigPath = "configs/local.json"
	defaultTokenTTL   = 24 * time.Hour

	envConfigPath = "CONFIG_PATH"
)

var ErrEmptyJWTSecret = errors.New("JWT_SECRET is not set")

// Load читает конфиг из JSON-файла (путь задаётся через CONFIG_PATH)
// и переопределяет значения переменными окружения.
// Секрет для JWT берётся только из окружения, чтобы не хранить его в репозитории.
func Load() (Config, error) {
	path := os.Getenv(envConfigPath)
	if path == "" {
		path = defaultConfigPath
	}

	return LoadFromFile(path)
}

func LoadFromFile(path string) (Config, error) {
	var cfg Config

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config %s: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config %s: %w", path, err)
	}

	if err := applyEnv(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func applyEnv(cfg *Config) error {
	setString(&cfg.HttpServer.Host, "HTTP_HOST")
	setString(&cfg.HttpServer.Port, "HTTP_PORT")
	setString(&cfg.WSServer.Host, "WS_HOST")
	setString(&cfg.WSServer.Port, "WS_PORT")

	setString(&cfg.Postgres.Host, "POSTGRES_HOST")
	setString(&cfg.Postgres.UserName, "POSTGRES_USER")
	setString(&cfg.Postgres.Password, "POSTGRES_PASSWORD")
	setString(&cfg.Postgres.NameDB, "POSTGRES_DB")
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		port, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid POSTGRES_PORT %q: %w", v, err)
		}
		cfg.Postgres.Port = port
	}

	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		cfg.Kafka.Brokers = strings.Split(v, ",")
	}
	setString(&cfg.Kafka.Topic, "KAFKA_TOPIC")
	setString(&cfg.Kafka.GroupID, "KAFKA_GROUP_ID")

	cfg.JWT.Secret = os.Getenv("JWT_SECRET")
	cfg.JWT.TTL = defaultTokenTTL
	if v := os.Getenv("JWT_TTL"); v != "" {
		ttl, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("invalid JWT_TTL %q: %w", v, err)
		}
		cfg.JWT.TTL = ttl
	}

	return nil
}

func setString(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

// Validate проверяет параметры, без которых приложение не может работать.
func (c Config) Validate() error {
	if c.JWT.Secret == "" {
		return ErrEmptyJWTSecret
	}
	if len(c.Kafka.Brokers) == 0 {
		return errors.New("kafka brokers are not set")
	}
	if c.Kafka.Topic == "" {
		return errors.New("kafka topic is not set")
	}
	return nil
}
