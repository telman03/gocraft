package config

import (
	"os"
	"strconv"
	"time"
)

// MaintenanceConfig holds all maintenance-related configuration
type MaintenanceConfig struct {
	// File cleanup configuration
	FileCleanupInterval  time.Duration
	FileRetentionPeriod  time.Duration
	FileCleanupBatchSize int
	FileMaxConcurrency   int
	FileCleanupEnabled   bool

	// Database maintenance configuration
	DBMaintenanceInterval   time.Duration
	DBArchivalThreshold     time.Duration
	DBCleanupBatchSize      int
	DBMaintenanceEnabled    bool

	// Storage configuration
	StorageBasePath string
}

// LoadMaintenanceConfig loads maintenance configuration from environment variables
func LoadMaintenanceConfig() *MaintenanceConfig {
	config := &MaintenanceConfig{
		// Default values
		FileCleanupInterval:     24 * time.Hour,
		FileRetentionPeriod:     30 * 24 * time.Hour,
		FileCleanupBatchSize:    100,
		FileMaxConcurrency:      5,
		FileCleanupEnabled:      true,
		DBMaintenanceInterval:   24 * time.Hour,
		DBArchivalThreshold:     90 * 24 * time.Hour,
		DBCleanupBatchSize:      1000,
		DBMaintenanceEnabled:    true,
		StorageBasePath:         "./output",
	}

	// Load from environment variables
	if val := os.Getenv("FILE_CLEANUP_INTERVAL_HOURS"); val != "" {
		if hours, err := strconv.Atoi(val); err == nil {
			config.FileCleanupInterval = time.Duration(hours) * time.Hour
		}
	}

	if val := os.Getenv("FILE_RETENTION_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			config.FileRetentionPeriod = time.Duration(days) * 24 * time.Hour
		}
	}

	if val := os.Getenv("FILE_CLEANUP_BATCH_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.FileCleanupBatchSize = size
		}
	}

	if val := os.Getenv("FILE_MAX_CONCURRENCY"); val != "" {
		if concurrency, err := strconv.Atoi(val); err == nil {
			config.FileMaxConcurrency = concurrency
		}
	}

	if val := os.Getenv("FILE_CLEANUP_ENABLED"); val != "" {
		config.FileCleanupEnabled = val == "true"
	}

	if val := os.Getenv("DB_MAINTENANCE_INTERVAL_HOURS"); val != "" {
		if hours, err := strconv.Atoi(val); err == nil {
			config.DBMaintenanceInterval = time.Duration(hours) * time.Hour
		}
	}

	if val := os.Getenv("DB_ARCHIVAL_THRESHOLD_DAYS"); val != "" {
		if days, err := strconv.Atoi(val); err == nil {
			config.DBArchivalThreshold = time.Duration(days) * 24 * time.Hour
		}
	}

	if val := os.Getenv("DB_CLEANUP_BATCH_SIZE"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			config.DBCleanupBatchSize = size
		}
	}

	if val := os.Getenv("DB_MAINTENANCE_ENABLED"); val != "" {
		config.DBMaintenanceEnabled = val == "true"
	}

	if val := os.Getenv("STORAGE_BASE_PATH"); val != "" {
		config.StorageBasePath = val
	}

	return config
}