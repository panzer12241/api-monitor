# API Monitor - Makefile

.PHONY: help migrate-fresh migrate-fresh-seed migrate-run seed dev build clean

help: ## Show this help message
	@echo "API Monitor - Available commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-fresh: ## Drop all tables and run fresh migrations (like Laravel migrate:fresh)
	@echo "🔄 Running fresh migration..."
	@go run cmd/migrate/main.go fresh

migrate-fresh-seed: ## Drop all tables, run migrations, and seed data (like Laravel migrate:fresh --seed)
	@echo "🔄 Running fresh migration with seeding..."
	@go run cmd/migrate/main.go fresh-seed

migrate-run: ## Run pending migrations only
	@echo "🔄 Running migrations..."
	@go run cmd/migrate/main.go run

seed: ## Run database seeders only
	@echo "🌱 Running seeders..."
	@go run cmd/migrate/main.go seed

kill-port: ## Kill process using port 8080
	@echo "🔪 Killing process on port 8080..."
	@-lsof -ti:8080 | xargs kill -9 2>/dev/null || echo "No process found on port 8080"

dev: ## Start development server
	@echo "🚀 Starting development server..."
	@make kill-port
	@go run main.go

build: ## Build the application
	@echo "🔨 Building application..."
	@go build -o bin/api-monitor main.go

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	@rm -rf bin/

test: ## Run tests
	@echo "🧪 Running tests..."
	@go test ./...

.DEFAULT_GOAL := help
