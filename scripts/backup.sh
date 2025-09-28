#!/bin/bash

# 🗄️ GoCraft Backup Script
# This script creates backups of the database and important files

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
BACKUP_TYPE="full"
RETENTION_DAYS=30

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--type)
            BACKUP_TYPE="$2"
            shift 2
            ;;
        -r|--retention)
            RETENTION_DAYS="$2"
            shift 2
            ;;
        -h|--help)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  -t, --type TYPE         Backup type: full, db, files (default: full)"
            echo "  -r, --retention DAYS    Retention period in days (default: 30)"
            echo "  -h, --help              Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                      # Full backup with 30 days retention"
            echo "  $0 -t db               # Database only backup"
            echo "  $0 -t files            # Files only backup"
            echo "  $0 -r 7                # Full backup with 7 days retention"
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Create backup directory
BACKUP_DIR="backups"
mkdir -p $BACKUP_DIR

# Timestamp for backup files
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

echo "🗄️ Starting GoCraft backup..."
echo "Backup type: $BACKUP_TYPE"
echo "Retention: $RETENTION_DAYS days"
echo ""

# Load environment variables
if [ -f ".env.local" ]; then
    export $(cat .env.local | grep -v '^#' | xargs)
    print_status "Environment variables loaded from .env.local"
elif [ -f ".env" ]; then
    export $(cat .env | grep -v '^#' | xargs)
    print_status "Environment variables loaded from .env"
else
    print_warning "No environment file found, using system environment"
fi

# Database backup function
backup_database() {
    print_status "Creating database backup..."
    
    if ! command -v pg_dump &> /dev/null; then
        print_error "pg_dump is not installed. Please install PostgreSQL client tools."
        return 1
    fi
    
    # Set default values if not provided
    DB_HOST=${DB_HOST:-localhost}
    DB_PORT=${DB_PORT:-5432}
    DB_USER=${DB_USER:-gocraft}
    DB_NAME=${DB_NAME:-gocraft_db}
    
    if [ -z "$DB_PASSWORD" ]; then
        print_error "DB_PASSWORD environment variable is required"
        return 1
    fi
    
    BACKUP_FILE="$BACKUP_DIR/db_backup_$TIMESTAMP.sql"
    
    # Set PGPASSWORD for non-interactive backup
    export PGPASSWORD=$DB_PASSWORD
    
    # Create database backup
    if pg_dump -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME > $BACKUP_FILE; then
        print_success "Database backup created: $BACKUP_FILE"
        
        # Compress the backup
        gzip $BACKUP_FILE
        print_success "Database backup compressed: $BACKUP_FILE.gz"
        
        # Get backup size
        BACKUP_SIZE=$(du -h "$BACKUP_FILE.gz" | cut -f1)
        print_status "Backup size: $BACKUP_SIZE"
    else
        print_error "Database backup failed"
        return 1
    fi
    
    # Unset PGPASSWORD
    unset PGPASSWORD
}

# Files backup function
backup_files() {
    print_status "Creating files backup..."
    
    BACKUP_FILE="$BACKUP_DIR/files_backup_$TIMESTAMP.tar.gz"
    
    # Files and directories to backup
    BACKUP_ITEMS=(
        "output"
        "internal/templates"
        "docs"
        ".env.example"
        "go.mod"
        "go.sum"
        "Makefile"
        "docker-compose.yaml"
        "Dockerfile"
        "README.md"
    )
    
    # Create tar archive
    tar -czf $BACKUP_FILE "${BACKUP_ITEMS[@]}" 2>/dev/null || {
        print_warning "Some files may not exist, continuing with available files"
    }
    
    if [ -f "$BACKUP_FILE" ]; then
        print_success "Files backup created: $BACKUP_FILE"
        
        # Get backup size
        BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
        print_status "Backup size: $BACKUP_SIZE"
    else
        print_error "Files backup failed"
        return 1
    fi
}

# Configuration backup function
backup_config() {
    print_status "Creating configuration backup..."
    
    CONFIG_BACKUP_DIR="$BACKUP_DIR/config_$TIMESTAMP"
    mkdir -p $CONFIG_BACKUP_DIR
    
    # Copy configuration files
    [ -f ".env.local" ] && cp .env.local $CONFIG_BACKUP_DIR/
    [ -f ".env.example" ] && cp .env.example $CONFIG_BACKUP_DIR/
    [ -f "docker-compose.yaml" ] && cp docker-compose.yaml $CONFIG_BACKUP_DIR/
    [ -f ".air.toml" ] && cp .air.toml $CONFIG_BACKUP_DIR/
    [ -f "Makefile" ] && cp Makefile $CONFIG_BACKUP_DIR/
    
    # Create archive
    tar -czf "$CONFIG_BACKUP_DIR.tar.gz" -C $BACKUP_DIR "config_$TIMESTAMP"
    rm -rf $CONFIG_BACKUP_DIR
    
    print_success "Configuration backup created: $CONFIG_BACKUP_DIR.tar.gz"
}

# Perform backup based on type
case $BACKUP_TYPE in
    "full")
        backup_database
        backup_files
        backup_config
        ;;
    "db")
        backup_database
        ;;
    "files")
        backup_files
        backup_config
        ;;
    *)
        print_error "Invalid backup type: $BACKUP_TYPE"
        exit 1
        ;;
esac

# Clean up old backups
print_status "Cleaning up old backups (older than $RETENTION_DAYS days)..."
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
find $BACKUP_DIR -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete 2>/dev/null || true
print_success "Old backups cleaned up"

# Create backup manifest
MANIFEST_FILE="$BACKUP_DIR/backup_manifest_$TIMESTAMP.txt"
cat > $MANIFEST_FILE << EOF
GoCraft Backup Manifest
======================
Timestamp: $(date)
Backup Type: $BACKUP_TYPE
Retention: $RETENTION_DAYS days

Files Created:
EOF

# List created backup files
find $BACKUP_DIR -name "*_$TIMESTAMP*" -type f >> $MANIFEST_FILE

print_success "Backup manifest created: $MANIFEST_FILE"

echo ""
print_success "🎉 Backup completed successfully!"
echo ""
echo "📋 Backup Summary:"
echo "  Type: $BACKUP_TYPE"
echo "  Timestamp: $TIMESTAMP"
echo "  Location: $BACKUP_DIR/"
echo "  Manifest: $MANIFEST_FILE"
echo ""
echo "🔧 Restore commands:"
echo "  Database: gunzip -c $BACKUP_DIR/db_backup_$TIMESTAMP.sql.gz | psql -h \$DB_HOST -U \$DB_USER -d \$DB_NAME"
echo "  Files: tar -xzf $BACKUP_DIR/files_backup_$TIMESTAMP.tar.gz"
echo ""
print_success "Backup complete! 🗄️"