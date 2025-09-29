# Makefile for test-full-stack

# Variables
BINARY_NAME=test-full-stack
BUILD_DIR=build
DOCKER_IMAGE=test-full-stack
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_UNIX=$(BINARY_NAME)_unix

# Default target
.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: build
build: ## Build the application
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/test-full-stack/main.go

.PHONY: build-linux
build-linux: ## Build for Linux
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) ./cmd/test-full-stack/main.go

.PHONY: clean
clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

.PHONY: test
test: ## Run tests
	$(GOTEST) -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

.PHONY: run
run: ## Run the application
	$(GOCMD) run ./cmd/test-full-stack/main.go

.PHONY: run-docker
run-docker: ## Run with Docker
	docker-compose up --build

.PHONY: stop-docker
stop-docker: ## Stop Docker containers
	docker-compose down

.PHONY: docker-build
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## Run Docker container
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-push
docker-push: ## Push Docker image
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: deps
deps: ## Download dependencies
	$(GOGET) -v -t -d ./...

.PHONY: deps-update
deps-update: ## Update dependencies
	$(GOMOD) tidy
	$(GOMOD) download

.PHONY: lint
lint: ## Run linter
	golangci-lint run

.PHONY: lint-install
lint-install: ## Install linter
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

.PHONY: swagger
swagger: ## Generate Swagger docs
	swag init -g cmd/test-full-stack/main.go

.PHONY: swagger-install
swagger-install: ## Install Swagger
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: migrate-up
migrate-up: ## Run database migrations up
	migrate -path db/migrations -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down: ## Run database migrations down
	migrate -path db/migrations -database "$(DATABASE_URL)" down

.PHONY: migrate-create
migrate-create: ## Create new migration file
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

.PHONY: dev
dev: ## Start development environment
	docker-compose -f docker-compose.dev.yml up --build

.PHONY: prod
prod: ## Start production environment
	docker-compose -f docker-compose.prod.yml up --build

.PHONY: logs
logs: ## Show application logs
	docker-compose logs -f app

.PHONY: shell
shell: ## Access application container shell
	docker-compose exec app sh

.PHONY: db-shell
db-shell: ## Access database shell
	docker-compose exec postgres psql -U postgres -d test-full-stack

.PHONY: redis-cli
redis-cli: ## Access Redis CLI
	docker-compose exec redis redis-cli

.PHONY: mongo-shell
mongo-shell: ## Access MongoDB shell
	docker-compose exec mongo mongosh test-full-stack

.PHONY: backup
backup: ## Backup database
	docker-compose exec postgres pg_dump -U postgres test-full-stack > backup_$(shell date +%Y%m%d_%H%M%S).sql

.PHONY: restore
restore: ## Restore database from backup
	@read -p "Enter backup file name: " file; \
	docker-compose exec -T postgres psql -U postgres test-full-stack < $$file

.PHONY: security-scan
security-scan: ## Run security scan
	trivy fs .

.PHONY: security-scan-install
security-scan-install: ## Install security scanner
	go install github.com/aquasecurity/trivy/cmd/trivy@latest

.PHONY: format
format: ## Format code
	gofmt -s -w .

.PHONY: vet
vet: ## Run go vet
	$(GOCMD) vet ./...

.PHONY: mod-verify
mod-verify: ## Verify dependencies
	$(GOMOD) verify

.PHONY: generate
generate: ## Generate code
	$(GOCMD) generate ./...

.PHONY: install
install: ## Install the application
	$(GOBUILD) -o $(GOPATH)/bin/$(BINARY_NAME) ./cmd/test-full-stack/main.go

.PHONY: uninstall
uninstall: ## Uninstall the application
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

.PHONY: release
release: ## Create a release build
	@echo "Creating release build..."
	$(MAKE) clean
	$(MAKE) build-linux
	@echo "Release build created in $(BUILD_DIR)/$(BINARY_UNIX)"

.PHONY: setup
setup: ## Initial setup
	@echo "Setting up test-full-stack..."
	$(MAKE) deps
	$(MAKE) lint-install
	$(MAKE) swagger-install
	$(MAKE) security-scan-install
	@echo "Setup complete!" 