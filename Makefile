# 🚀 GoCraft Makefile

# Variables
BINARY_NAME=gocraft
MAIN_PATH=cmd/gocraft/main.go
BUILD_DIR=bin
DOCKER_IMAGE=gocraft/generator
VERSION?=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

.PHONY: all build clean test deps dev docker help

# Default target
all: clean deps test build

## 🏗️ Build Commands

# Build the application
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "🌍 Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	@echo "✅ Multi-platform build complete"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf output/*
	@echo "✅ Clean complete"

## 🧪 Testing Commands

# Run all tests
test:
	@echo "🧪 Running tests..."
	$(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run integration tests
test-integration:
	@echo "🔗 Running integration tests..."
	$(GOTEST) -v -tags=integration ./tests/integration/...

# Run benchmarks
benchmark:
	@echo "⚡ Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## 📦 Dependencies

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Update dependencies
deps-update:
	@echo "🔄 Updating dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

# Verify dependencies
deps-verify:
	@echo "🔍 Verifying dependencies..."
	$(GOMOD) verify
	@echo "✅ Dependencies verified"

## 🚀 Development Commands

# Start development server with hot reload
dev:
	@echo "🔥 Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air -c .air.toml

# Start development server (simple)
run:
	@echo "🚀 Starting server..."
	$(GOCMD) run $(MAIN_PATH)

# Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...
	@echo "✅ Code formatted"

# Run linter
lint:
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run
	@echo "✅ Linting complete"

# Security scan
security:
	@echo "🔒 Running security scan..."
	@which gosec > /dev/null || (echo "Installing gosec..." && go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest)
	gosec ./...
	@echo "✅ Security scan complete"

## 🐳 Docker Commands

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(VERSION) .
	docker tag $(DOCKER_IMAGE):$(VERSION) $(DOCKER_IMAGE):latest
	@echo "✅ Docker image built: $(DOCKER_IMAGE):$(VERSION)"

# Run with Docker Compose
docker-up:
	@echo "🚀 Starting services with Docker Compose..."
	docker-compose up -d
	@echo "✅ Services started"

# Stop Docker Compose services
docker-down:
	@echo "🛑 Stopping Docker Compose services..."
	docker-compose down
	@echo "✅ Services stopped"

# View Docker Compose logs
docker-logs:
	docker-compose logs -f

## 📚 Documentation Commands

# Generate API documentation
docs:
	@echo "📚 Generating API documentation..."
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g $(MAIN_PATH) -o docs/api
	@echo "✅ API documentation generated"

# Serve documentation locally
docs-serve:
	@echo "📖 Serving documentation..."
	@which godoc > /dev/null || (echo "Installing godoc..." && go install golang.org/x/tools/cmd/godoc@latest)
	godoc -http=:6060
	@echo "📖 Documentation available at http://localhost:6060"

## 🗄️ Database Commands

# Run database migrations
migrate:
	@echo "🗄️ Running database migrations..."
	$(GOCMD) run scripts/migrate.go
	@echo "✅ Migrations complete"

# Add user roles and create admin
setup-admin:
	@echo "👑 Setting up admin user..."
	$(GOCMD) run scripts/add_user_roles.go
	@echo "✅ Admin setup complete"

# Optimize database performance
optimize-db:
	@echo "⚡ Optimizing database..."
	$(GOCMD) run scripts/optimize_database.go
	@echo "✅ Database optimization complete"

## 🔧 Utility Commands

# Install development tools
install-tools:
	@echo "🛠️ Installing development tools..."
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	go install github.com/swaggo/swag/cmd/swag@latest
	go install golang.org/x/tools/cmd/godoc@latest
	@echo "✅ Development tools installed"

# Setup development environment
setup-dev:
	@echo "🚀 Setting up development environment..."
	./scripts/setup-dev.sh
	@echo "✅ Development environment setup complete"

# Create backup
backup:
	@echo "🗄️ Creating backup..."
	./scripts/backup.sh
	@echo "✅ Backup complete"

# Create database backup only
backup-db:
	@echo "🗄️ Creating database backup..."
	./scripts/backup.sh -t db
	@echo "✅ Database backup complete"

# Deploy to production
deploy:
	@echo "🚀 Deploying to production..."
	./scripts/deploy.sh
	@echo "✅ Deployment complete"

# Deploy to staging
deploy-staging:
	@echo "🚀 Deploying to staging..."
	./scripts/deploy.sh -e staging
	@echo "✅ Staging deployment complete"

# Check project health
health-check:
	@echo "🏥 Checking project health..."
	@echo "Go version: $$(go version)"
	@echo "Dependencies: $$(go list -m all | wc -l) modules"
	@echo "Test coverage: $$(go test -coverprofile=/tmp/coverage.out ./... > /dev/null 2>&1 && go tool cover -func=/tmp/coverage.out | tail -1 | awk '{print $$3}' || echo 'N/A')"
	@echo "Build status: $$(make build > /dev/null 2>&1 && echo '✅ OK' || echo '❌ Failed')"
	@echo "✅ Health check complete"

# Show project statistics
stats:
	@echo "📊 Project Statistics:"
	@echo "Lines of code: $$(find . -name '*.go' -not -path './vendor/*' | xargs wc -l | tail -1 | awk '{print $$1}')"
	@echo "Go files: $$(find . -name '*.go' -not -path './vendor/*' | wc -l)"
	@echo "Packages: $$(go list ./... | wc -l)"
	@echo "Dependencies: $$(go list -m all | wc -l)"

## 📋 CI/CD Commands

# Run all CI checks
ci: deps fmt lint security test build
	@echo "✅ All CI checks passed"

# Prepare for release
release: clean deps test build-all docs
	@echo "🎉 Release preparation complete"

## ❓ Help

# Show available commands
help:
	@echo "🚀 GoCraft Makefile Commands:"
	@echo ""
	@echo "🏗️  Build Commands:"
	@echo "  build         Build the application"
	@echo "  build-all     Build for multiple platforms"
	@echo "  clean         Clean build artifacts"
	@echo ""
	@echo "🧪 Testing Commands:"
	@echo "  test          Run all tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo "  benchmark     Run performance benchmarks"
	@echo ""
	@echo "🚀 Development Commands:"
	@echo "  dev           Start development server with hot reload"
	@echo "  run           Start server (simple)"
	@echo "  fmt           Format code"
	@echo "  lint          Run linter"
	@echo "  security      Run security scan"
	@echo "  setup-dev     Setup development environment"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  docker-build  Build Docker image"
	@echo "  docker-up     Start with Docker Compose"
	@echo "  docker-down   Stop Docker Compose"
	@echo ""
	@echo "📚 Documentation Commands:"
	@echo "  docs          Generate API documentation"
	@echo "  docs-serve    Serve documentation locally"
	@echo ""
	@echo "🗄️  Database Commands:"
	@echo "  migrate       Run database migrations"
	@echo "  setup-admin   Create admin user"
	@echo "  optimize-db   Optimize database performance"
	@echo ""
	@echo "🔧 Utility Commands:"
	@echo "  install-tools Install development tools"
	@echo "  health-check  Check project health"
	@echo "  stats         Show project statistics"
	@echo "  ci            Run all CI checks"
	@echo "  backup        Create full backup"
	@echo "  backup-db     Create database backup only"
	@echo "  deploy        Deploy to production"
	@echo "  deploy-staging Deploy to staging"
	@echo "  help          Show this help message"