# =============================================================================
# DevFlow — Makefile
# =============================================================================
# Usage:
#   make run              Start the development server
#   make build            Build the binary
#   make test             Run all tests
#   make lint             Run linter (requires golangci-lint)
#   make tidy             Clean up go.mod and go.sum
#   make migrate-up       Run all pending migrations
#   make migrate-down     Roll back last migration
#   make migrate-create   Create new migration (usage: make migrate-create name=create_tasks)
#   make help             Show available commands

include .env
export

.PHONY: run build test lint tidy clean migrate-up migrate-down migrate-create swagger help

.DEFAULT_GOAL := help

## run: Start the development server
run:
	go run cmd/server/main.go

## build: Build the binary to bin/server
build:
	@mkdir -p bin
	go build -o bin/server cmd/server/main.go

## test: Run all tests with verbose output
test:
	go test ./... -v -count=1

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## tidy: Clean up module dependencies
tidy:
	go mod tidy

## clean: Remove build artifacts
clean:
	@rm -rf bin/
	@echo "Cleaned build artifacts"

MIGRATE_BIN := $(shell go env GOPATH)/bin/migrate
SWAG_BIN := $(shell go env GOPATH)/bin/swag

## swagger: Generate Swagger / OpenAPI documentation from doc comments
swagger:
	$(SWAG_BIN) init -g cmd/server/main.go -o docs

## migrate-up: Run all pending database migrations
migrate-up:
	$(MIGRATE_BIN) -path migrations -database "$(DATABASE_URL)" -verbose up

## migrate-down: Roll back the last database migration
migrate-down:
	$(MIGRATE_BIN) -path migrations -database "$(DATABASE_URL)" -verbose down 1

## migrate-create: Create a new migration file (usage: make migrate-create name=create_tasks)
migrate-create:
	$(MIGRATE_BIN) create -ext sql -dir migrations -seq $(name)

## help: Show this help message
help:
	@echo "DevFlow — Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
	@echo ""
