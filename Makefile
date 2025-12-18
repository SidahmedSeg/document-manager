.PHONY: help setup up down restart logs ps clean test db-migrate db-rollback db-reset health check-env

# =============================================================================
# VARIABLES
# =============================================================================
COMPOSE := docker-compose
COMPOSE_DEV := $(COMPOSE) --profile development
PROJECT_NAME := document-manager

# Colors for output
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m
COLOR_RED := \033[31m

# =============================================================================
# HELP
# =============================================================================
help: ## Show this help message
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Document Manager - Development Commands$(COLOR_RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(COLOR_BOLD)Common Workflows:$(COLOR_RESET)"
	@echo "  1. First time setup:  make setup && make up"
	@echo "  2. Start services:    make up"
	@echo "  3. View logs:         make logs"
	@echo "  4. Check health:      make health"
	@echo "  5. Stop services:     make down"
	@echo ""

# =============================================================================
# ENVIRONMENT SETUP
# =============================================================================
setup: ## Initial project setup (copy .env, create networks)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Setting up Document Manager...$(COLOR_RESET)"
	@if [ ! -f .env ]; then \
		echo "$(COLOR_YELLOW)Creating .env file from .env.example...$(COLOR_RESET)"; \
		cp .env.example .env; \
		echo "$(COLOR_GREEN)✓ .env file created. Please edit it with your configuration.$(COLOR_RESET)"; \
	else \
		echo "$(COLOR_YELLOW).env file already exists. Skipping...$(COLOR_RESET)"; \
	fi
	@echo "$(COLOR_YELLOW)Creating Docker networks...$(COLOR_RESET)"
	@docker network create shared-auth-network 2>/dev/null || echo "$(COLOR_YELLOW)Network 'shared-auth-network' already exists$(COLOR_RESET)"
	@echo "$(COLOR_GREEN)✓ Setup complete!$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Next steps:$(COLOR_RESET)"
	@echo "  1. Edit .env file with your configuration"
	@echo "  2. Run 'make up' to start all services"
	@echo ""

check-env: ## Check if .env file exists
	@if [ ! -f .env ]; then \
		echo "$(COLOR_RED)Error: .env file not found!$(COLOR_RESET)"; \
		echo "$(COLOR_YELLOW)Run 'make setup' first$(COLOR_RESET)"; \
		exit 1; \
	fi

# =============================================================================
# DOCKER COMPOSE OPERATIONS
# =============================================================================
up: check-env ## Start all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting Document Manager services...$(COLOR_RESET)"
	$(COMPOSE_DEV) up -d
	@echo "$(COLOR_GREEN)✓ All services started!$(COLOR_RESET)"
	@echo ""
	@make ps
	@echo ""
	@echo "$(COLOR_BOLD)Service URLs:$(COLOR_RESET)"
	@echo "  Oathkeeper Proxy:    http://localhost:14455"
	@echo "  Oathkeeper API:      http://localhost:14456"
	@echo "  PostgreSQL:          localhost:15432"
	@echo "  Redis:               localhost:16379"
	@echo "  MinIO Console:       http://localhost:19001"
	@echo "  MinIO API:           http://localhost:19000"
	@echo "  Meilisearch:         http://localhost:17700"
	@echo "  NATS Monitor:        http://localhost:18222"
	@echo "  ClickHouse:          http://localhost:18123"
	@echo "  MailSlurper:         http://localhost:14437"
	@echo ""

down: ## Stop all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Stopping Document Manager services...$(COLOR_RESET)"
	$(COMPOSE) down
	@echo "$(COLOR_GREEN)✓ All services stopped!$(COLOR_RESET)"

restart: ## Restart all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Restarting Document Manager services...$(COLOR_RESET)"
	$(COMPOSE_DEV) restart
	@echo "$(COLOR_GREEN)✓ All services restarted!$(COLOR_RESET)"

stop: ## Stop all services (alias for down)
	@make down

start: ## Start all services (alias for up)
	@make up

# =============================================================================
# LOGS & MONITORING
# =============================================================================
logs: ## View logs from all services (usage: make logs service=postgres)
	@if [ -z "$(service)" ]; then \
		echo "$(COLOR_BOLD)$(COLOR_BLUE)Showing logs for all services...$(COLOR_RESET)"; \
		$(COMPOSE) logs -f --tail=100; \
	else \
		echo "$(COLOR_BOLD)$(COLOR_BLUE)Showing logs for $(service)...$(COLOR_RESET)"; \
		$(COMPOSE) logs -f --tail=100 $(service); \
	fi

ps: ## Show status of all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Service Status:$(COLOR_RESET)"
	@$(COMPOSE) ps

health: ## Check health of all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Checking service health...$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Oathkeeper:$(COLOR_RESET)"
	@curl -f -s http://localhost:14456/health/ready > /dev/null && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)PostgreSQL:$(COLOR_RESET)"
	@docker exec docmanager-postgres pg_isready -U postgres > /dev/null 2>&1 && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Redis:$(COLOR_RESET)"
	@docker exec docmanager-redis redis-cli ping > /dev/null 2>&1 && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)MinIO:$(COLOR_RESET)"
	@curl -f -s http://localhost:19000/minio/health/live > /dev/null && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Meilisearch:$(COLOR_RESET)"
	@curl -f -s http://localhost:17700/health > /dev/null && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)NATS:$(COLOR_RESET)"
	@curl -f -s http://localhost:18222/healthz > /dev/null && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)ClickHouse:$(COLOR_RESET)"
	@curl -f -s http://localhost:18123/ping > /dev/null && echo "  $(COLOR_GREEN)✓ Healthy$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Unhealthy$(COLOR_RESET)"
	@echo ""

stats: ## Show resource usage statistics
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Resource Usage:$(COLOR_RESET)"
	@$(COMPOSE) stats --no-stream

top: ## Show running processes in services
	@$(COMPOSE) top

# =============================================================================
# DATABASE OPERATIONS
# =============================================================================
db-migrate: ## Run database migrations
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running database migrations...$(COLOR_RESET)"
	@docker exec -it docmanager-postgres psql -U postgres -d docmanager -f /migrations/000001_create_extensions_and_types.up.sql
	@echo "$(COLOR_GREEN)✓ Migrations completed!$(COLOR_RESET)"

db-rollback: ## Rollback last migration
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Rolling back last migration...$(COLOR_RESET)"
	@docker exec -it docmanager-postgres psql -U postgres -d docmanager -f /migrations/000001_create_extensions_and_types.down.sql
	@echo "$(COLOR_GREEN)✓ Rollback completed!$(COLOR_RESET)"

db-reset: ## Reset database (WARNING: deletes all data)
	@echo "$(COLOR_RED)$(COLOR_BOLD)WARNING: This will delete ALL data!$(COLOR_RESET)"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@echo "$(COLOR_YELLOW)Resetting database...$(COLOR_RESET)"
	@$(COMPOSE) down -v
	@$(COMPOSE) up -d postgres
	@sleep 5
	@make db-migrate
	@echo "$(COLOR_GREEN)✓ Database reset complete!$(COLOR_RESET)"

db-shell: ## Open PostgreSQL shell
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Opening PostgreSQL shell...$(COLOR_RESET)"
	@docker exec -it docmanager-postgres psql -U postgres -d docmanager

db-backup: ## Backup database to ./backups/
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Backing up database...$(COLOR_RESET)"
	@mkdir -p backups
	@docker exec docmanager-postgres pg_dump -U postgres docmanager | gzip > backups/docmanager_$$(date +%Y%m%d_%H%M%S).sql.gz
	@echo "$(COLOR_GREEN)✓ Backup created in backups/ directory$(COLOR_RESET)"

db-restore: ## Restore database from backup (usage: make db-restore file=backups/docmanager_20240101_120000.sql.gz)
	@if [ -z "$(file)" ]; then \
		echo "$(COLOR_RED)Error: Please specify backup file$(COLOR_RESET)"; \
		echo "Usage: make db-restore file=backups/docmanager_20240101_120000.sql.gz"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_YELLOW)Restoring database from $(file)...$(COLOR_RESET)"
	@gunzip -c $(file) | docker exec -i docmanager-postgres psql -U postgres -d docmanager
	@echo "$(COLOR_GREEN)✓ Database restored!$(COLOR_RESET)"

# =============================================================================
# REDIS OPERATIONS
# =============================================================================
redis-cli: ## Open Redis CLI
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Opening Redis CLI...$(COLOR_RESET)"
	@docker exec -it docmanager-redis redis-cli

redis-flush: ## Flush all Redis data (WARNING: clears cache)
	@echo "$(COLOR_RED)$(COLOR_BOLD)WARNING: This will clear ALL cached data!$(COLOR_RESET)"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@docker exec docmanager-redis redis-cli FLUSHALL
	@echo "$(COLOR_GREEN)✓ Redis cache cleared!$(COLOR_RESET)"

# =============================================================================
# MINIO OPERATIONS
# =============================================================================
minio-console: ## Open MinIO console in browser
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Opening MinIO console...$(COLOR_RESET)"
	@echo "URL: http://localhost:19001"
	@echo "Username: $(shell grep MINIO_ROOT_USER .env | cut -d '=' -f2)"
	@echo "Password: $(shell grep MINIO_ROOT_PASSWORD .env | cut -d '=' -f2)"

minio-buckets: ## List MinIO buckets
	@docker exec docmanager-minio-init mc ls minio/

# =============================================================================
# TESTING
# =============================================================================
test-auth: ## Test connectivity to shared Kratos/Hydra
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Testing shared authentication...$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Testing Kratos:$(COLOR_RESET)"
	@curl -f -s $(shell grep SHARED_KRATOS_PUBLIC_URL .env | cut -d '=' -f2)/health/ready > /dev/null && echo "  $(COLOR_GREEN)✓ Kratos is reachable$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Kratos is NOT reachable$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Testing Hydra:$(COLOR_RESET)"
	@curl -f -s $(shell grep SHARED_HYDRA_PUBLIC_URL .env | cut -d '=' -f2)/.well-known/jwks.json > /dev/null && echo "  $(COLOR_GREEN)✓ Hydra is reachable$(COLOR_RESET)" || echo "  $(COLOR_RED)✗ Hydra is NOT reachable$(COLOR_RESET)"
	@echo ""

test: ## Run all tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running tests...$(COLOR_RESET)"
	@cd backend && go test ./... -v

test-integration: ## Run integration tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running integration tests...$(COLOR_RESET)"
	@cd backend && go test ./... -tags=integration -v

test-e2e: ## Run end-to-end tests
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running E2E tests...$(COLOR_RESET)"
	@cd frontend/user-app && npm run test:e2e

# =============================================================================
# DEVELOPMENT
# =============================================================================
dev-backend: ## Start backend service in development mode (usage: make dev-backend service=tenant)
	@if [ -z "$(service)" ]; then \
		echo "$(COLOR_RED)Error: Please specify service name$(COLOR_RESET)"; \
		echo "Usage: make dev-backend service=tenant"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting $(service)-service in development mode...$(COLOR_RESET)"
	@cd backend/services/$(service)-service && go run cmd/main.go

dev-frontend: ## Start frontend in development mode
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting frontend in development mode...$(COLOR_RESET)"
	@cd frontend/user-app && npm run dev

dev-admin: ## Start admin frontend in development mode
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Starting admin frontend in development mode...$(COLOR_RESET)"
	@cd frontend/admin-app && npm run dev

# =============================================================================
# BUILD OPERATIONS
# =============================================================================
build: ## Build all backend services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Building backend services...$(COLOR_RESET)"
	@cd backend && $(MAKE) build

build-service: ## Build specific service (usage: make build-service service=tenant)
	@if [ -z "$(service)" ]; then \
		echo "$(COLOR_RED)Error: Please specify service name$(COLOR_RESET)"; \
		echo "Usage: make build-service service=tenant"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Building $(service)-service...$(COLOR_RESET)"
	@cd backend/services/$(service)-service && go build -o bin/$(service)-service cmd/main.go
	@echo "$(COLOR_GREEN)✓ Built bin/$(service)-service$(COLOR_RESET)"

build-docker: ## Build Docker images for all services
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Building Docker images...$(COLOR_RESET)"
	$(COMPOSE) build

build-docker-service: ## Build Docker image for specific service (usage: make build-docker-service service=tenant)
	@if [ -z "$(service)" ]; then \
		echo "$(COLOR_RED)Error: Please specify service name$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Building Docker image for $(service)-service...$(COLOR_RESET)"
	$(COMPOSE) build $(service)-service

# =============================================================================
# CLEAN OPERATIONS
# =============================================================================
clean: ## Stop services and remove containers, networks, volumes
	@echo "$(COLOR_RED)$(COLOR_BOLD)WARNING: This will remove all containers, networks, and volumes!$(COLOR_RESET)"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@echo "$(COLOR_YELLOW)Cleaning up...$(COLOR_RESET)"
	$(COMPOSE) down -v --remove-orphans
	@echo "$(COLOR_GREEN)✓ Cleanup complete!$(COLOR_RESET)"

clean-logs: ## Remove all log files
	@echo "$(COLOR_YELLOW)Removing log files...$(COLOR_RESET)"
	@find . -name "*.log" -type f -delete
	@echo "$(COLOR_GREEN)✓ Log files removed!$(COLOR_RESET)"

clean-cache: ## Clean Go build cache
	@echo "$(COLOR_YELLOW)Cleaning Go build cache...$(COLOR_RESET)"
	@go clean -cache -modcache -testcache
	@echo "$(COLOR_GREEN)✓ Cache cleaned!$(COLOR_RESET)"

prune: ## Remove all unused Docker resources (careful!)
	@echo "$(COLOR_RED)$(COLOR_BOLD)WARNING: This will remove all unused Docker resources!$(COLOR_RESET)"
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	@docker system prune -af --volumes
	@echo "$(COLOR_GREEN)✓ Docker resources pruned!$(COLOR_RESET)"

# =============================================================================
# DEPLOYMENT
# =============================================================================
deploy-staging: ## Deploy to staging environment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Deploying to staging...$(COLOR_RESET)"
	@kubectl apply -f k8s/staging/
	@echo "$(COLOR_GREEN)✓ Deployed to staging!$(COLOR_RESET)"

deploy-prod: ## Deploy to production environment
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Deploying to production...$(COLOR_RESET)"
	@kubectl apply -f k8s/production/
	@echo "$(COLOR_GREEN)✓ Deployed to production!$(COLOR_RESET)"

# =============================================================================
# UTILITIES
# =============================================================================
install-tools: ## Install required development tools
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Installing development tools...$(COLOR_RESET)"
	@command -v go >/dev/null 2>&1 || { echo "$(COLOR_YELLOW)Installing Go...$(COLOR_RESET)"; brew install go; }
	@command -v node >/dev/null 2>&1 || { echo "$(COLOR_YELLOW)Installing Node.js...$(COLOR_RESET)"; brew install node; }
	@command -v docker >/dev/null 2>&1 || { echo "$(COLOR_YELLOW)Installing Docker...$(COLOR_RESET)"; brew install docker; }
	@echo "$(COLOR_GREEN)✓ All tools installed!$(COLOR_RESET)"

version: ## Show version information
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Document Manager Version Information$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Application:$(COLOR_RESET)"
	@grep APP_VERSION .env | cut -d '=' -f2
	@echo ""
	@echo "$(COLOR_BOLD)Tools:$(COLOR_RESET)"
	@echo "  Go:         $(shell go version)"
	@echo "  Node:       $(shell node --version)"
	@echo "  Docker:     $(shell docker --version)"
	@echo "  Compose:    $(shell docker-compose --version)"
	@echo ""

lint: ## Run linters on Go code
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Running linters...$(COLOR_RESET)"
	@cd backend && golangci-lint run ./...

fmt: ## Format Go code
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Formatting Go code...$(COLOR_RESET)"
	@cd backend && go fmt ./...
	@echo "$(COLOR_GREEN)✓ Code formatted!$(COLOR_RESET)"

generate: ## Generate code (mocks, protobuf, etc.)
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Generating code...$(COLOR_RESET)"
	@cd backend && go generate ./...
	@echo "$(COLOR_GREEN)✓ Code generated!$(COLOR_RESET)"

docs: ## Generate API documentation
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Generating documentation...$(COLOR_RESET)"
	@cd backend && swag init
	@echo "$(COLOR_GREEN)✓ Documentation generated!$(COLOR_RESET)"

# =============================================================================
# MONITORING
# =============================================================================
monitor: ## Open all monitoring UIs in browser
	@echo "$(COLOR_BOLD)$(COLOR_BLUE)Opening monitoring dashboards...$(COLOR_RESET)"
	@open http://localhost:19001  # MinIO Console
	@open http://localhost:17700  # Meilisearch
	@open http://localhost:18222  # NATS Monitor
	@open http://localhost:14437  # MailSlurper

# =============================================================================
# DEFAULT
# =============================================================================
.DEFAULT_GOAL := help
