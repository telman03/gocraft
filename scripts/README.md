# 🛠️ GoCraft Scripts

This directory contains automation scripts for development, deployment, and maintenance tasks.

## Scripts Overview

### Development Scripts

#### `setup-dev.sh`
Automated development environment setup script.

**Usage:**
```bash
./scripts/setup-dev.sh
```

**What it does:**
- Checks for required tools (Go, Docker, Docker Compose)
- Creates `.env.local` from `.env.example` if it doesn't exist
- Installs development tools (air, golangci-lint, gosec, swag)
- Downloads Go dependencies
- Creates necessary directories
- Starts database and runs migrations
- Sets up admin user
- Generates API documentation

**Requirements:**
- Go 1.23.4+
- Docker and Docker Compose (optional)

### Deployment Scripts

#### `deploy.sh`
Production deployment automation script.

**Usage:**
```bash
# Deploy to production
./scripts/deploy.sh

# Deploy to staging with specific version
./scripts/deploy.sh -e staging -v v1.2.3

# Quick deployment (skip tests and build)
./scripts/deploy.sh --skip-tests --skip-build
```

**Options:**
- `-e, --environment ENV`: Deployment environment (default: production)
- `-v, --version VERSION`: Version tag (default: latest)
- `--skip-tests`: Skip running tests
- `--skip-build`: Skip building the application
- `-h, --help`: Show help message

**What it does:**
- Pre-deployment checks (tools, environment files)
- Runs tests and security scans
- Builds application and Docker image
- Runs database migrations
- Creates database backup (production only)
- Deploys with Docker Compose
- Performs health checks
- Runs post-deployment tasks

### Backup Scripts

#### `backup.sh`
Database and file backup automation script.

**Usage:**
```bash
# Full backup
./scripts/backup.sh

# Database only
./scripts/backup.sh -t db

# Files only
./scripts/backup.sh -t files

# Custom retention period
./scripts/backup.sh -r 7
```

**Options:**
- `-t, --type TYPE`: Backup type (full, db, files) - default: full
- `-r, --retention DAYS`: Retention period in days - default: 30
- `-h, --help`: Show help message

**What it does:**
- Creates compressed database backups
- Archives important files and configurations
- Manages backup retention
- Creates backup manifests
- Provides restore instructions

### Database Scripts

#### `migrate/main.go`
Database migration script using GORM AutoMigrate.

**Usage:**
```bash
go run scripts/migrate/main.go
# or
make migrate
```

**What it does:**
- Connects to the database
- Runs GORM AutoMigrate for all models
- Creates/updates database schema

#### `add-user-roles/main.go`
Admin user creation and role management script.

**Usage:**
```bash
# Set environment variables first
export ADMIN_EMAIL=admin@example.com
export ADMIN_PASSWORD=secure-password

go run scripts/add-user-roles/main.go
# or
make setup-admin
```

**What it does:**
- Adds role column to users table
- Creates admin user with specified credentials
- Sets up proper user roles

#### `optimize_database.go`
Database performance optimization script.

**Usage:**
```bash
go run scripts/optimize_database.go
# or
make optimize-db
```

**What it does:**
- Creates performance indexes
- Optimizes query performance
- Updates table statistics
- Analyzes database tables

### Configuration Files

#### `init-db.sql`
PostgreSQL initialization script for Docker containers.

**What it does:**
- Creates necessary PostgreSQL extensions
- Sets up development-friendly logging
- Configures performance settings

## Environment Variables

### Required for Database Scripts
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=gocraft
DB_PASSWORD=your-password
DB_NAME=gocraft_db
DB_SSLMODE=disable
```

### Required for Admin Setup
```bash
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=secure-password
```

## Script Permissions

Make sure scripts are executable:
```bash
chmod +x scripts/*.sh
```

## Integration with Makefile

All scripts are integrated with the project Makefile:

```bash
# Development
make setup-dev      # Run setup-dev.sh
make migrate        # Run migrate/main.go
make setup-admin    # Run add-user-roles/main.go
make optimize-db    # Run optimize_database.go

# Backup
make backup         # Run backup.sh (full)
make backup-db      # Run backup.sh -t db

# Deployment
make deploy         # Run deploy.sh
make deploy-staging # Run deploy.sh -e staging
```

## Docker Integration

Scripts work with Docker Compose services:

```bash
# Start database for scripts
docker-compose up -d db

# Run scripts
./scripts/setup-dev.sh

# Stop services
docker-compose down
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
chmod +x scripts/*.sh
```

#### Database Connection Failed
```bash
# Check if database is running
docker-compose ps db

# Check environment variables
cat .env.local

# Test connection
psql -h localhost -p 5432 -U gocraft -d gocraft_db
```

#### Missing Tools
```bash
# Install required tools
make install-tools

# Or manually
go install github.com/cosmtrek/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Script Logs

Scripts provide colored output:
- 🔵 **INFO**: General information
- 🟢 **SUCCESS**: Successful operations
- 🟡 **WARNING**: Non-critical issues
- 🔴 **ERROR**: Critical errors

### Getting Help

Each script supports the `--help` flag:
```bash
./scripts/setup-dev.sh --help
./scripts/deploy.sh --help
./scripts/backup.sh --help
```

## Best Practices

1. **Always test scripts in development first**
2. **Review environment variables before running**
3. **Keep backups before major operations**
4. **Use version tags for production deployments**
5. **Monitor script output for errors**
6. **Keep scripts executable with proper permissions**

## Contributing

When adding new scripts:

1. Follow the existing naming convention
2. Add proper error handling and colored output
3. Include help documentation (`--help` flag)
4. Add integration with Makefile
5. Update this README
6. Make scripts executable (`chmod +x`)
7. Test thoroughly in development environment

---

For more information, see the [Development Guide](../docs/DEVELOPMENT.md) or run `make help`.