package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testConfig = `{
	"kafka": {"brokers": ["127.0.0.1:9092"], "topic": "message_send", "group_id": "g"},
	"http_server": {"host": "0.0.0.0", "port": "8000"},
	"websocket_server": {"host": "0.0.0.0", "port": "8001"},
	"postgres": {"host": "127.0.0.1", "port": 5433, "user": "u", "password": "p", "name_db": "db"}
}`

func writeConfig(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(testConfig), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadFromFileWithEnvOverrides(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "postgres")
	t.Setenv("POSTGRES_PORT", "5432")
	t.Setenv("KAFKA_BROKERS", "kafka:29092,kafka2:29092")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_TTL", "30m")

	cfg, err := LoadFromFile(writeConfig(t))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if cfg.Postgres.Host != "postgres" || cfg.Postgres.Port != 5432 {
		t.Errorf("postgres = %s:%d", cfg.Postgres.Host, cfg.Postgres.Port)
	}
	if len(cfg.Kafka.Brokers) != 2 || cfg.Kafka.Brokers[0] != "kafka:29092" {
		t.Errorf("brokers = %v", cfg.Kafka.Brokers)
	}
	if cfg.JWT.Secret != "secret" || cfg.JWT.TTL != 30*time.Minute {
		t.Errorf("jwt = %+v", cfg.JWT)
	}
	if cfg.HttpServer.Address() != "0.0.0.0:8000" {
		t.Errorf("http address = %s", cfg.HttpServer.Address())
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("validate: %v", err)
	}
}

func TestValidateRequiresJWTSecret(t *testing.T) {
	t.Setenv("JWT_SECRET", "")

	cfg, err := LoadFromFile(writeConfig(t))
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	if err := cfg.Validate(); !errors.Is(err, ErrEmptyJWTSecret) {
		t.Fatalf("want ErrEmptyJWTSecret, got %v", err)
	}
}

func TestInvalidPort(t *testing.T) {
	t.Setenv("POSTGRES_PORT", "abc")

	if _, err := LoadFromFile(writeConfig(t)); err == nil {
		t.Fatal("want error for invalid POSTGRES_PORT")
	}
}
