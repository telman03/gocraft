#!/bin/bash

# 🚀 GoCraft Development Setup Script
# This script sets up the development environment for GoCraft

set -e

echo "🚀 Setting up GoCraft development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.23.4 or later."
    exit 1
fi

print_status "Go version: $(go version)"

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_warning "Docker is not installed. Some features may not work."
else
    print_status "Docker version: $(docker --version)"
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    print_warning "Docker Compose is not installed. Some features may not work."
else
    print_status "Docker Compose version: $(docker-compose --version)"
fi

# Create .env.local if it doesn't exist
if [ ! -f ".env.local" ]; then
    print_status "Creating .env.local from .env.example..."
    if [ -f ".env.example" ]; then
        cp .env.example .env.local
        print_success ".env.local created from .env.example"
    else
        print_warning ".env.example not found. Creating basic .env.local..."
        cat > .env.local << EOF
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=gocraft
DB_PASSWORD=change_this_password
DB_NAME=gocraft_db
DB_SSLMODE=disable

# Server Configuration
PORT=8080
GO_ENV=development

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production-min-32-chars

# Admin Configuration
ADMIN_EMAIL=admin@gocraft.dev
ADMIN_PASSWORD=change_this_admin_password

# File Storage Configuration
STORAGE_BASE_PATH=./output
FILE_RETENTION_PERIOD=24h

# Maintenance Configuration
FILE_CLEANUP_ENABLED=true
FILE_CLEANUP_INTERVAL=1h
FILE_CLEANUP_BATCH_SIZE=100
FILE_MAX_CONCURRENCY=5

DB_MAINTENANCE_ENABLED=true
DB_MAINTENANCE_INTERVAL=6h
DB_ARCHIVAL_THRESHOLD=168h
DB_CLEANUP_BATCH_SIZE=1000
EOF
        print_success "Basic .env.local created"
    fi
else
    print_status ".env.local already exists"
fi

# Install development tools
print_status "Installing development tools..."

# Install air for hot reload
if ! command -v air &> /dev/null; then
    print_status "Installing air for hot reload..."
    go install github.com/cosmtrek/air@latest
    print_success "Air installed"
else
    print_status "Air already installed"
fi

# Install golangci-lint
if ! command -v golangci-lint &> /dev/null; then
    print_status "Installing golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    print_success "golangci-lint installed"
else
    print_status "golangci-lint already installed"
fi

# Install gosec
if ! command -v gosec &> /dev/null; then
    print_status "Installing gosec for security scanning..."
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    print_success "gosec installed"
else
    print_status "gosec already installed"
fi

# Install swag for API documentation
if ! command -v swag &> /dev/null; then
    print_status "Installing swag for API documentation..."
    go install github.com/swaggo/swag/cmd/swag@latest
    print_success "swag installed"
else
    print_status "swag already installed"
fi

# Download Go dependencies
print_status "Downloading Go dependencies..."
go mod download
go mod tidy
print_success "Dependencies downloaded"

# Create necessary directories
print_status "Creating necessary directories..."
mkdir -p output
mkdir -p tmp
mkdir -p bin
mkdir -p tests/fixtures
print_success "Directories created"

# Generate API documentation
print_status "Generating API documentation..."
if command -v swag &> /dev/null; then
    swag init -g cmd/gocraft/main.go -o docs/api
    print_success "API documentation generated"
else
    print_warning "Swag not available, skipping API documentation generation"
fi

# Start database with Docker Compose (if available)
if command -v docker-compose &> /dev/null; then
    print_status "Starting database with Docker Compose..."
    docker-compose up -d db
    
    # Wait for database to be ready
    print_status "Waiting for database to be ready..."
    sleep 10
    
    # Run migrations
    print_status "Running database migrations..."
    go run scripts/migrate/main.go
    print_success "Database migrations completed"
    
    # Setup admin user
    print_status "Setting up admin user..."
    go run scripts/add-user-roles/main.go
    print_success "Admin user setup completed"
    
else
    print_warning "Docker Compose not available. Please start the database manually."
fi

echo ""
print_success "🎉 Development environment setup complete!"
echo ""
echo "📋 Next steps:"
echo "  1. Review and update .env.local with your configuration"
echo "  2. Start development server: make dev"
echo "  3. Visit http://localhost:8080/swagger/ for API documentation"
echo "  4. Visit http://localhost:8080/debug for debug page"
echo ""
echo "🔧 Available commands:"
echo "  make dev          - Start development server with hot reload"
echo "  make test         - Run tests"
echo "  make lint         - Run linter"
echo "  make docker-up    - Start all services with Docker"
echo "  make help         - Show all available commands"
echo ""
print_success "Happy coding! 🚀"