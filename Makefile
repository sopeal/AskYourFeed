# Makefile for AskYourFeed Development Database
# Automates PostgreSQL database setup via Docker

# Database configuration
DB_NAME := askyourfeed
DB_USER := postgres
DB_PASSWORD := postgres
DB_PORT := 5432
DB_CONTAINER := askyourfeed_postgres_dev
POSTGRES_VERSION := 16-alpine

# SQL migration file
MIGRATION_FILE := db/20251023214826_create_secure_schema.sql

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: help db-start db-stop db-restart db-logs db-clean db-status db-connect db-init db-reset

help: ## Show this help message
	@echo "$(GREEN)AskYourFeed Development Database Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

db-start: ## Start PostgreSQL container
	@echo "$(GREEN)Starting PostgreSQL container...$(NC)"
	@docker run -d \
		--name $(DB_CONTAINER) \
		-e POSTGRES_DB=$(DB_NAME) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-p $(DB_PORT):5432 \
		-v askyourfeed_pgdata:/var/lib/postgresql/data \
		postgres:$(POSTGRES_VERSION) || \
		(echo "$(YELLOW)Container already exists, starting it...$(NC)" && docker start $(DB_CONTAINER))
	@echo "$(GREEN)Waiting for PostgreSQL to be ready...$(NC)"
	@sleep 3
	@docker exec $(DB_CONTAINER) pg_isready -U $(DB_USER) || sleep 2
	@echo "$(GREEN)PostgreSQL is ready!$(NC)"

db-init: db-start ## Initialize database with schema migration
	@echo "$(GREEN)Initializing database schema...$(NC)"
	@sleep 2
	@docker exec -i $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME) < $(MIGRATION_FILE)
	@echo "$(GREEN)Database schema initialized successfully!$(NC)"

db-stop: ## Stop PostgreSQL container
	@echo "$(YELLOW)Stopping PostgreSQL container...$(NC)"
	@docker stop $(DB_CONTAINER) || echo "$(RED)Container not running$(NC)"
	@echo "$(GREEN)PostgreSQL container stopped$(NC)"

db-restart: db-stop db-start ## Restart PostgreSQL container
	@echo "$(GREEN)PostgreSQL container restarted$(NC)"

db-logs: ## Show PostgreSQL container logs
	@docker logs -f $(DB_CONTAINER)

db-clean: db-stop ## Stop and remove PostgreSQL container and volume
	@echo "$(RED)Removing PostgreSQL container and data volume...$(NC)"
	@docker rm $(DB_CONTAINER) || echo "$(YELLOW)Container already removed$(NC)"
	@docker volume rm askyourfeed_pgdata || echo "$(YELLOW)Volume already removed$(NC)"
	@echo "$(GREEN)Cleanup complete$(NC)"

db-reset: db-clean db-init ## Clean and reinitialize database
	@echo "$(GREEN)Database reset complete$(NC)"

db-status: ## Check PostgreSQL container status
	@echo "$(GREEN)PostgreSQL Container Status:$(NC)"
	@docker ps -a --filter name=$(DB_CONTAINER) --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "$(RED)Container not found$(NC)"

db-connect: ## Connect to PostgreSQL via psql
	@echo "$(GREEN)Connecting to PostgreSQL...$(NC)"
	@docker exec -it $(DB_CONTAINER) psql -U $(DB_USER) -d $(DB_NAME)

# Connection string for application use
db-url: ## Print database connection URL
	@echo "postgresql://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(DB_NAME)?sslmode=disable"
