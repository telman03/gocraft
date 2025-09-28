# 🚀 GoCraft Development Guide

This guide covers the development workflow, build processes, and automation tools for the GoCraft project.

## Table of Contents

- [Quick Start](#quick-start)
- [Development Environment](#development-environment)
- [Build System](#build-system)
- [Docker Development](#docker-development)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Database Management](#database-management)
- [Deployment](#deployment)
- [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Go 1.23.4 or later
- Docker and Docker Compose (optional but recommended)
- PostgreSQL client tools (for database operations)

### Setup Development Environment

1. **Automated Setup** (Recommended):
   ```bash
   ./scripts/setup-dev.sh
   ```

2. **Manual Setup**:
   ```bash
   # Clone and navigate to the project
   git clone <repository-url>
   cd gocraft

   # Copy environment file
   cp .env.example .env.local

   # Install dependencies
   go mod download

   # Install development tools
   make install-tools

   # Start database
   make docker-up

   # Run migrations
   make migrate

   # Setup admin user
   make setup-admin
   ```

3. **Start Development Server**:
   ```bash
   make dev  # Hot reload development server
   # or
   make run  # Simple server without hot reload
   ```

## Development Environment

### Environment Configuration

The project uses environment-specific configuration files:

- `.env.example` - Template with all available options
- `.env.local` - Local development configuration (not tracked in git)
- `.env.production` - Production configuration (not tracked in git)
- `.env.staging` - Staging configuration (not tracked in git)

### Key Environment Variables

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=gocraft
DB_PASSWORD=your_secure_password
DB_NAME=gocraft_db
DB_SSLMODE=disable

# Server
PORT=8080
GO_ENV=development

# JWT
JWT_SECRET=your-super-secret-jwt-key

# Admin
ADMIN_EMAIL=admin@gocraft.dev
ADMIN_PASSWORD=admin123

# File Storage
STORAGE_BASE_PATH=./output
FILE_RETENTION_PERIOD=24h

# Maintenance
FILE_CLEANUP_ENABLED=true
FILE_CLEANUP_INTERVAL=1h
DB_MAINTENANCE_ENABLED=true
DB_MAINTENANCE_INTERVAL=6h
```

### Hot Reload Development

The project uses [Air](https://github.com/cosmtrek/air) for hot reload during development:

```bash
# Start with hot reload
make dev

# Or directly with air
air -c .air.toml
```

Configuration is in `.air.toml` and watches for changes in:
- Go files (`.go`)
- Template files (`.tpl`, `.tmpl`, `.html`)
- Excludes test files and temporary directories

## Build System

### Makefile Targets

The project includes a comprehensive Makefile with the following categories:

#### Build Commands
```bash
make build          # Build the application
make build-all      # Build for multiple platforms
make clean          # Clean build artifacts
```

#### Testing Commands
```bash
make test           # Run all tests
make test-coverage  # Run tests with coverage report
make test-integration # Run integration tests
make benchmark      # Run performance benchmarks
```

#### Development Commands
```bash
make dev            # Start development server with hot reload
make run            # Start server (simple)
make fmt            # Format code
make lint           # Run linter
make security       # Run security scan
```

#### Docker Commands
```bash
make docker-build   # Build Docker image
make docker-up      # Start with Docker Compose
make docker-down    # Stop Docker Compose
make docker-logs    # View Docker Compose logs
```

#### Documentation Commands
```bash
make docs           # Generate API documentation
make docs-serve     # Serve documentation locally
```

#### Database Commands
```bash
make migrate        # Run database migrations
make setup-admin    # Create admin user
make optimize-db    # Optimize database performance
```

#### Utility Commands
```bash
make install-tools  # Install development tools
make health-check   # Check project health
make stats          # Show project statistics
make ci             # Run all CI checks
make help           # Show all available commands
```

### Build Configuration

The build system supports:

- **Cross-platform builds**: Linux, macOS, Windows
- **Version embedding**: Build time and version information
- **Optimized binaries**: CGO disabled for static linking
- **Multi-stage Docker builds**: Minimal production images

## Docker Development

### Services

The `docker-compose.yaml` defines three services:

1. **Database** (`db`): PostgreSQL 15 with health checks
2. **Application** (`app`): Production build with health checks
3. **Development** (`dev`): Development build with hot reload (profile: dev)

### Development with Docker

```bash
# Start database only
docker-compose up -d db

# Start all services (production mode)
docker-compose up -d

# Start development mode with hot reload
docker-compose --profile dev up -d

# View logs
docker-compose logs -f app
docker-compose logs -f db

# Stop services
docker-compose down
```

### Docker Images

- **Production**: Multi-stage build with Alpine Linux
- **Development**: Go development image with Air for hot reload
- **Security**: Non-root user, minimal attack surface

## Testing

### Test Structure

```
tests/
├── unit/           # Unit tests
├── integration/    # Integration tests
└── fixtures/       # Test data and fixtures
```

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Integration tests only
make test-integration

# Benchmarks
make benchmark

# Specific package
go test ./internal/handlers/...
```

### Test Utilities

The project includes test utilities in `tests/unit/test_utils.go`:
- Database setup/teardown
- Mock data generation
- Common test helpers

## Code Quality

### Linting

The project uses `golangci-lint` with comprehensive rules:

```bash
# Run linter
make lint

# Install linter
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Security Scanning

Security scanning with `gosec`:

```bash
# Run security scan
make security

# Install gosec
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
```

### Code Formatting

```bash
# Format code
make fmt

# This runs: go fmt ./...
```

### Pre-commit Checks

Run all quality checks:

```bash
make ci  # Runs: deps, fmt, lint, security, test, build
```

## Database Management

### Migrations

Database migrations are handled by GORM's AutoMigrate:

```bash
# Run migrations
make migrate

# Or directly
go run scripts/migrate/main.go
```

### Admin User Setup

```bash
# Create admin user
make setup-admin

# Or directly
go run scripts/add-user-roles/main.go
```

### Database Optimization

```bash
# Optimize database performance
make optimize-db

# Or directly
go run scripts/optimize_database.go
```

### Backup and Restore

```bash
# Full backup
./scripts/backup.sh

# Database only
./scripts/backup.sh -t db

# Files only
./scripts/backup.sh -t files

# Custom retention (7 days)
./scripts/backup.sh -r 7
```

## Deployment

### Production Deployment

```bash
# Full deployment
./scripts/deploy.sh

# Staging deployment
./scripts/deploy.sh -e staging -v v1.2.3

# Skip tests (not recommended)
./scripts/deploy.sh --skip-tests

# Quick deployment
./scripts/deploy.sh --skip-tests --skip-build
```

### Deployment Process

1. **Pre-deployment checks**: Tools, environment files
2. **Testing**: Unit tests, security scan
3. **Building**: Application binary, Docker image
4. **Database**: Migrations, backup (production)
5. **Deployment**: Docker Compose deployment
6. **Health checks**: Service availability
7. **Post-deployment**: Database optimization, documentation

### Health Checks

The application includes health check endpoints:

- **Application**: `GET /api/health`
- **Docker**: Built-in health checks with curl
- **Database**: PostgreSQL health checks

## Troubleshooting

### Common Issues

#### Database Connection Issues
```bash
# Check database status
docker-compose ps db

# View database logs
docker-compose logs db

# Test connection
psql -h localhost -p 5432 -U gocraft -d gocraft_db
```

#### Build Issues
```bash
# Clean and rebuild
make clean
make build

# Check Go version
go version

# Update dependencies
make deps-update
```

#### Hot Reload Not Working
```bash
# Check air installation
which air

# Reinstall air
go install github.com/cosmtrek/air@latest

# Check .air.toml configuration
cat .air.toml
```

#### Permission Issues
```bash
# Fix script permissions
chmod +x scripts/*.sh

# Check Docker permissions
docker ps
```

### Debug Mode

Enable debug logging:

```bash
# Set environment variable
export GO_ENV=development

# Or in .env.local
GO_ENV=development
```

### Performance Issues

```bash
# Run benchmarks
make benchmark

# Check project health
make health-check

# Optimize database
make optimize-db
```

### Getting Help

1. **Check logs**: `docker-compose logs -f app`
2. **Run health check**: `make health-check`
3. **View project stats**: `make stats`
4. **Check documentation**: `make docs-serve`
5. **Review Makefile**: `make help`

## Development Workflow

### Recommended Workflow

1. **Start development environment**:
   ```bash
   ./scripts/setup-dev.sh
   make dev
   ```

2. **Make changes** to code

3. **Run tests**:
   ```bash
   make test
   ```

4. **Check code quality**:
   ```bash
   make lint
   make security
   ```

5. **Commit changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

6. **Before pushing**:
   ```bash
   make ci  # Run all checks
   ```

### Git Hooks (Optional)

Consider setting up pre-commit hooks:

```bash
# .git/hooks/pre-commit
#!/bin/bash
make ci
```

## Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Fiber Framework](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)

---

For more information, see the [API Documentation](../docs/api/) or run `make help` for available commands.