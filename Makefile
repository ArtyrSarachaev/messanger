.PHONY: up down logs infra migrate run test lint

# Поднять всё окружение (Postgres, Kafka, миграции, приложение)
up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f app

# Только инфраструктура — для запуска приложения из IDE
infra:
	docker compose up -d postgres zookeeper kafka

migrate:
	go run ./migrations

run:
	JWT_SECRET=$${JWT_SECRET:-local-dev-secret-change-me} go run ./cmd

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
