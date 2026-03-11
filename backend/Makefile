# =============================================================================
# Storefront Backend Makefile
# Run all targets from backend/
# =============================================================================

.PHONY: help build run test lint \
        migrate-up migrate-down migrate-status migrate-reset migrate-version \
        migrate-create docker-up docker-down

# Auto-load .env if it exists
-include .env
export

# ── Defaults ─────────────────────────────────────────────────────────────────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-22s\033[0m %s\n", $$1, $$2}'

# ── Build & Run ───────────────────────────────────────────────────────────────

build: ## Compile the API server binary
	go build -o bin/api ./cmd/api

run: ## Run the API server (reads from .env)
	go run ./cmd/api

test: ## Run all tests
	go test ./... -race -count=1

lint: ## Run golangci-lint
	golangci-lint run ./...

# ── Migrations (goose) ────────────────────────────────────────────────────────
# Requires DATABASE_URL to be set (or present in .env — source it first).
# Example: source .env && make migrate-up

migrate-up: ## Apply all pending migrations
	go run ./cmd/migrate up

migrate-down: ## Roll back the last applied migration
	go run ./cmd/migrate down

migrate-status: ## Show migration status
	go run ./cmd/migrate status

migrate-version: ## Print the current schema version
	go run ./cmd/migrate version

migrate-reset: ## Roll back ALL migrations (destructive — dev only)
	go run ./cmd/migrate reset

migrate-create: ## Create a new migration: make migrate-create name=add_something
	@test -n "$(name)" || (echo "Error: provide a name — e.g. make migrate-create name=add_something" && exit 1)
	go run ./cmd/migrate create $(name)

# ── Docker ────────────────────────────────────────────────────────────────────

docker-up: ## Start local Postgres via docker-compose
	docker compose up -d postgres

docker-down: ## Stop and remove containers
	docker compose down
