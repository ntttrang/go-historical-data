.PHONY: help build run test lint clean docker-build docker-up docker-down migrate-up migrate-down

# Variables
APP_NAME=historical-data-api
DOCKER_IMAGE=historical-data-api
DOCKER_TAG=latest
GO_VERSION=1.21

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	@go build -o bin/$(APP_NAME) cmd/api/main.go
	@echo "Build completed: bin/$(APP_NAME)"

run: ## Run the application
	@echo "Running $(APP_NAME)..."
	@go run cmd/api/main.go

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v -race ./internal/...

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@go test -v -race ./tests/integration/...

test-csv: ## Test CSV upload endpoint
	@echo "Testing CSV upload endpoint..."
	@./scripts/test_csv_upload.sh

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run --timeout=5m

lint-fix: ## Run linter and fix issues
	@echo "Running linter with auto-fix..."
	@golangci-lint run --fix --timeout=5m

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean completed"

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

docker-build: ## Build Docker image (local platform only)
	@echo "Building Docker image for local platform..."
	@docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	@echo "Docker image built: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-build-multiplatform: ## Build multi-platform Docker image (amd64 + arm64)
	@echo "Setting up Docker buildx..."
	@docker buildx create --name multiplatform-builder --use 2>/dev/null || docker buildx use multiplatform-builder
	@docker buildx inspect --bootstrap
	@echo "Building multi-platform Docker image (amd64 + arm64)..."
	@docker buildx build \
		--platform linux/amd64,linux/arm64 \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--tag $(DOCKER_IMAGE):latest \
		--push \
		.
	@echo "Multi-platform Docker image built and pushed: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-build-amd64: ## Build Docker image for linux/amd64 (Render compatible)
	@echo "Setting up Docker buildx..."
	@docker buildx create --name multiplatform-builder --use 2>/dev/null || docker buildx use multiplatform-builder
	@docker buildx inspect --bootstrap
	@echo "Building Docker image for linux/amd64..."
	@docker buildx build \
		--platform linux/amd64 \
		--tag $(DOCKER_IMAGE):$(DOCKER_TAG) \
		--tag $(DOCKER_IMAGE):latest \
		--load \
		.
	@echo "Docker image built for linux/amd64: $(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-push-amd64: ## Build and push Docker image for linux/amd64 to Docker Hub
	@echo "Setting up Docker buildx..."
	@docker buildx create --name multiplatform-builder --use 2>/dev/null || docker buildx use multiplatform-builder
	@docker buildx inspect --bootstrap
	@echo "Building and pushing Docker image for linux/amd64..."
	@docker buildx build \
		--platform linux/amd64 \
		--tag minhtrang2106/$(DOCKER_IMAGE):$(DOCKER_TAG) \
		--tag minhtrang2106/$(DOCKER_IMAGE):latest \
		--push \
		.
	@echo "Docker image built and pushed for linux/amd64: minhtrang2106/$(DOCKER_IMAGE):$(DOCKER_TAG)"

docker-up: ## Start Docker Compose services
	@echo "Starting Docker Compose services..."
	@docker-compose up -d
	@echo "Services started"

docker-down: ## Stop Docker Compose services
	@echo "Stopping Docker Compose services..."
	@docker-compose down
	@echo "Services stopped"

docker-jenkins: ## Start Docker Compose services with Jenkins
	@echo "Starting Docker Compose services with Jenkins..."
	@docker-compose -f docker-compose-ci.yml up -d
	@echo "Services started with Jenkins"

docker-kill: ## Kill Docker Compose services

docker-logs: ## View Docker Compose logs
	@docker-compose logs -f

docker-restart: ## Restart Docker Compose services
	@echo "Restarting Docker Compose services..."
	@docker-compose restart
	@echo "Services restarted"

verify-metrics: ## Verify Prometheus and Grafana setup
	@echo "Verifying metrics setup..."
	@./scripts/verify_metrics.sh

verify-tracing: ## Verify Jaeger tracing setup
	@echo "Verifying tracing setup..."
	@./scripts/verify_tracing.sh

setup-elk: ## Set up Elasticsearch ILM and index template
	@echo "Setting up Elasticsearch..."
	@./scripts/setup_elasticsearch.sh

logs-api: ## View API logs
	@docker-compose logs -f api

logs-elk: ## View ELK stack logs
	@docker-compose logs -f elasticsearch logstash kibana

logs-all: ## View all service logs
	@docker-compose logs -f

elk-status: ## Check ELK stack health
	@echo "Checking Elasticsearch..."
	@curl -s http://localhost:9200/_cluster/health?pretty || echo "❌ Elasticsearch not available"
	@echo "\nChecking Logstash..."
	@curl -s http://localhost:9600/_node/stats?pretty | head -20 || echo "❌ Logstash not available"
	@echo "\nChecking Kibana..."
	@curl -s http://localhost:5601/api/status | head -10 || echo "❌ Kibana not available"

elk-indices: ## List Elasticsearch indices
	@curl -s http://localhost:9200/_cat/indices?v

elk-count: ## Count logs in Elasticsearch
	@curl -s http://localhost:9200/historical-data-api-*/_count?pretty

elk-test: ## Send test log to Logstash
	@echo '{"message":"Test log from Makefile","level":"INFO","timestamp":"'$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")'","service":"historical-data-api"}' | nc localhost 5000
	@echo "✅ Test log sent. Check Kibana in a few seconds."

elk-verify: ## Run comprehensive ELK stack tests
	@./scripts/test_elk_stack.sh

observability-urls: ## Display all observability service URLs
	@echo "📊 Observability Stack URLs:"
	@echo "  Kibana (Logs):      http://localhost:5601"
	@echo "  Grafana (Metrics):  http://localhost:3000 (admin/admin)"
	@echo "  Jaeger (Traces):    http://localhost:16686"
	@echo "  Prometheus:         http://localhost:9090"
	@echo "  Elasticsearch:      http://localhost:9200"
	@echo "  Logstash:           http://localhost:9600"
	@echo "  API:                http://localhost:8080"

migrate-create: ## Create a new migration (usage: make migrate-create NAME=migration_name)
	@echo "Creating migration: $(NAME)"
	@touch database/migrations/$(shell date +%Y%m%d%H%M%S)_$(NAME).up.sql
	@touch database/migrations/$(shell date +%Y%m%d%H%M%S)_$(NAME).down.sql

dev: docker-up run ## Start development environment

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install go.uber.org/mock/mockgen@latest
	@echo "Tools installed"

jenkins-up: ## Start Jenkins services
	@echo "Starting Jenkins services..."
	@docker-compose -f docker-compose-ci.yml up -d
	@echo "Jenkins services started"
	@echo "Waiting for Jenkins to be ready..."
	@sleep 10
	@echo "Jenkins is available at http://localhost:8089"

jenkins-down: ## Stop Jenkins services
	@echo "Stopping Jenkins services..."
	@docker-compose -f docker-compose-ci.yml down
	@echo "Jenkins services stopped"
	
jenkins-initpassword: ## Get Jenkins initial admin password
	@echo "Fetching Jenkins initial admin password..."
	@docker exec jenkins-ci cat /var/jenkins_home/secrets/initialAdminPassword || echo "Jenkins already configured"

jenkins-login: ## Login to Jenkins
	@echo "Logging in to Jenkins..."
	@open http://localhost:8089

jenkins-rebuild: ## Rebuild Jenkins with Go and Docker
	@echo "Rebuilding Jenkins..."
	@./scripts/rebuild_jenkins.sh

jenkins-verify: ## Verify Jenkins tools installation
	@echo "Verifying Jenkins tools..."
	@docker-compose -f docker-compose-ci.yml exec -T jenkins bash -c "go version && docker --version && docker compose version"

.DEFAULT_GOAL := help

