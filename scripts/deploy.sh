#!/bin/bash

# 🚀 GoCraft Production Deployment Script
# This script handles production deployment tasks

set -e

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

# Default values
ENVIRONMENT="production"
VERSION="latest"
SKIP_TESTS=false
SKIP_BUILD=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        --skip-tests)
            SKIP_TESTS=true
            shift
            ;;
        --skip-build)
            SKIP_BUILD=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -e, --environment ENV    Deployment environment (default: production)"
            echo "  -v, --version VERSION    Version tag (default: latest)"
            echo "  --skip-tests            Skip running tests"
            echo "  --skip-build            Skip building the application"
            echo "  -h, --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                                    # Deploy to production with latest version"
            echo "  $0 -e staging -v v1.2.3             # Deploy to staging with specific version"
            echo "  $0 --skip-tests --skip-build        # Quick deployment without tests/build"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

echo "🚀 Starting GoCraft deployment..."
echo "Environment: $ENVIRONMENT"
echo "Version: $VERSION"
echo ""

# Pre-deployment checks
print_status "Running pre-deployment checks..."

# Check if required tools are installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    exit 1
fi

# Check if .env file exists for the environment
ENV_FILE=".env.${ENVIRONMENT}"
if [ ! -f "$ENV_FILE" ]; then
    print_warning "$ENV_FILE not found, using .env.local"
    ENV_FILE=".env.local"
    if [ ! -f "$ENV_FILE" ]; then
        print_error "No environment file found. Please create $ENV_FILE"
        exit 1
    fi
fi

print_success "Pre-deployment checks passed"

# Run tests (unless skipped)
if [ "$SKIP_TESTS" = false ]; then
    print_status "Running tests..."
    make test
    print_success "All tests passed"
else
    print_warning "Skipping tests"
fi

# Run security scan
print_status "Running security scan..."
if command -v gosec &> /dev/null; then
    make security
    print_success "Security scan completed"
else
    print_warning "gosec not installed, skipping security scan"
fi

# Build application (unless skipped)
if [ "$SKIP_BUILD" = false ]; then
    print_status "Building application..."
    make clean
    make build
    print_success "Application built successfully"
else
    print_warning "Skipping build"
fi

# Build Docker image
print_status "Building Docker image..."
docker build -t gocraft/generator:$VERSION .
docker tag gocraft/generator:$VERSION gocraft/generator:latest
print_success "Docker image built: gocraft/generator:$VERSION"

# Database migrations
print_status "Running database migrations..."
if [ -f "$ENV_FILE" ]; then
    export $(cat $ENV_FILE | grep -v '^#' | xargs)
fi

# Check database connectivity
print_status "Checking database connectivity..."
go run scripts/migrate/main.go
print_success "Database migrations completed"

# Backup database (production only)
if [ "$ENVIRONMENT" = "production" ]; then
    print_status "Creating database backup..."
    BACKUP_FILE="backup_$(date +%Y%m%d_%H%M%S).sql"
    
    if command -v pg_dump &> /dev/null; then
        pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > "backups/$BACKUP_FILE"
        print_success "Database backup created: backups/$BACKUP_FILE"
    else
        print_warning "pg_dump not available, skipping database backup"
    fi
fi

# Deploy with Docker Compose
print_status "Deploying with Docker Compose..."

# Stop existing services
docker-compose down

# Start services
docker-compose up -d

# Wait for services to be healthy
print_status "Waiting for services to be healthy..."
sleep 30

# Health check
print_status "Performing health check..."
HEALTH_URL="http://localhost:8080/api/health"

for i in {1..10}; do
    if curl -f $HEALTH_URL > /dev/null 2>&1; then
        print_success "Health check passed"
        break
    else
        if [ $i -eq 10 ]; then
            print_error "Health check failed after 10 attempts"
            exit 1
        fi
        print_status "Health check attempt $i/10 failed, retrying in 10 seconds..."
        sleep 10
    fi
done

# Post-deployment tasks
print_status "Running post-deployment tasks..."

# Optimize database
print_status "Optimizing database..."
go run scripts/optimize_database.go
print_success "Database optimization completed"

# Generate API documentation
print_status "Generating API documentation..."
if command -v swag &> /dev/null; then
    make docs
    print_success "API documentation updated"
fi

# Clean up old Docker images
print_status "Cleaning up old Docker images..."
docker image prune -f
print_success "Docker cleanup completed"

echo ""
print_success "🎉 Deployment completed successfully!"
echo ""
echo "📋 Deployment Summary:"
echo "  Environment: $ENVIRONMENT"
echo "  Version: $VERSION"
echo "  Health Check: ✅ Passed"
echo "  API Documentation: http://localhost:8080/swagger/"
echo ""
echo "🔧 Useful commands:"
echo "  docker-compose logs -f app    # View application logs"
echo "  docker-compose logs -f db     # View database logs"
echo "  make docker-down              # Stop all services"
echo ""
print_success "Deployment complete! 🚀"