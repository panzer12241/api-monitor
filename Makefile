# API Monitor - Simple Development Makefile

# Variables
BACKEND_DIR = backend
FRONTEND_DIR = frontend

# Colors for output
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

.PHONY: dev help

# Development targets (sequential execution - recommended)
dev:
	@echo "$(GREEN)Starting development environment (sequential)...$(NC)"
	@echo "🔄 Starting frontend first, then backend after delay..."
	@$(MAKE) dev-frontend & 
	@sleep 3
	@$(MAKE) dev-backend

# Development targets (parallel execution - for advanced users)
dev-parallel:
	@echo "$(GREEN)Starting development environment (parallel)...$(NC)"
	@echo "⚠️  Warning: Backend logs may interfere with frontend startup"
	@$(MAKE) -j2 dev-frontend dev-backend

# Frontend development  
dev-frontend:
	@echo "$(GREEN)Starting Frontend (Vue.js)...$(NC)"
	@cd $(FRONTEND_DIR) && \
	if [ ! -d node_modules ]; then \
		echo "📦 Installing frontend dependencies..."; \
		bun install; \
	fi && \
	echo "🚀 Frontend starting..." && \
	bun dev

# Backend development
dev-backend:
	@echo "$(GREEN)Starting Backend (Go)...$(NC)"
	@cd $(BACKEND_DIR) && \
	if [ ! -f .env ]; then \
		echo "❌ .env file not found in backend directory"; \
		echo "Please create .env file with database configuration"; \
		exit 1; \
	fi && \
	echo "🚀 Backend starting..." && \
	go run main.go

# Help
help:
	@echo "$(BLUE)API Monitor Development$(NC)"
	@echo ""
	@echo "$(GREEN)Usage:$(NC)"
	@echo "  make dev    - Start both backend and frontend"
	@echo "  make help   - Show this help"
	@echo ""
	@echo "$(GREEN)Services:$(NC)"
	@echo "  Backend:  http://localhost:8080"
	@echo "  Frontend: http://localhost:3000"
	@echo ""
	@echo "$(GREEN)Requirements:$(NC)"
	@echo "  - Go 1.23+"
	@echo "  - Node.js 18+"
	@echo "  - .env file in backend directory"
