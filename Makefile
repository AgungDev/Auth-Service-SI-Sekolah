.PHONY: help build run test docker-build docker-up docker-down docker-clean lint fmt migrate

help:
	@echo "Auth Service - Makefile"
	@echo ""
	@echo "Available commands:"
	@echo "  make build              - Build the application"
	@echo "  make run                - Run the application locally"
	@echo "  make test               - Run tests"
	@echo "  make lint               - Run linter"
	@echo "  make fmt                - Format code"
	@echo "  make docker-build       - Build Docker image"
	@echo "  make docker-up          - Start Docker Compose"
	@echo "  make docker-down        - Stop Docker Compose"
	@echo "  make docker-clean       - Remove Docker containers and volumes"
	@echo "  make migrate            - Run database migrations"
	@echo "  make deps               - Download dependencies"
	@echo "  make clean              - Clean build artifacts"

# Build
build:
	CGO_ENABLED=0 go build -o bin/auth-service ./cmd/api

run: build
	./bin/auth-service

# Dependencies
deps:
	go mod download
	go mod tidy

# Code Quality
test:
	go test -v -cover ./...

lint:
	golangci-lint run

fmt:
	go fmt ./...

# Docker
docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

docker-clean:
	docker compose down -v

docker-logs:
	docker compose logs -f

docker-ps:
	docker compose ps

# Database
migrate:
	@echo "Running database migrations..."
	@docker compose exec auth-db psql -U postgres -d auth_db < database/migrations/001_init.sql

# Cleanup
clean:
	rm -rf bin/
	rm -rf dist/
	go clean

# Development
dev:
	docker compose up auth-db -d
	go run ./cmd/api

# Production build
prod-build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/auth-service ./cmd/api

# Check
check: fmt lint test

.DEFAULT_GOAL := help
