-- Database initialization script for GoCraft
-- This script runs when the PostgreSQL container starts for the first time

-- Create extensions if they don't exist
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements";

-- Create indexes for better performance (these will be created by GORM as well, but having them here ensures they exist)
-- Note: GORM will handle the actual table creation, this is just for additional optimizations

-- Set some PostgreSQL configuration for better development experience
ALTER SYSTEM SET log_statement = 'all';
ALTER SYSTEM SET log_min_duration_statement = 0;
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';

-- Reload configuration
SELECT pg_reload_conf();