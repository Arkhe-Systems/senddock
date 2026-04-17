include .env
export

export PATH := $(PATH):$(shell go env GOPATH)/bin

.PHONY: dev db-up db-down db-restart migrate run build test sqlc

# Start PostgreSQL and Redis
db-up:
	docker compose -f ../docker-compose.yml up -d

# Stop PostgreSQL and Redis
db-down:
	docker compose -f ../docker-compose.yml down

# Restart Docker + containers (fixes port conflicts)
db-restart:
	sudo systemctl restart docker && docker compose -f ../docker-compose.yml up -d

# Run database migrations
migrate:
	goose -dir migrations postgres "$(DATABASE_URL)" up

# Run the server
run:
	go run cmd/server/main.go

# Build the binary
build:
	go build -o bin/senddock cmd/server/main.go

# Start everything: DB + migrations + server
test:
	go test ./internal/... -v

sqlc:
	sqlc generate

dev: db-up migrate run
