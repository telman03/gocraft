package services

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/telman03/gocraft-backend/internal/models"
	"gorm.io/gorm"
)

// DatabaseMaintenanceService handles database cleanup, archival, and health monitoring
type DatabaseMaintenanceService struct {
	db                  *gorm.DB
	maintenanceInterval time.Duration
	archivalThreshold   time.Duration
	cleanupBatchSize    int
	isRunning           bool
	stopChan            chan struct{}
	mu                  sync.RWMutex
	logger              *log.Logger
}

// MaintenanceConfig holds configuration for database maintenance
type MaintenanceConfig struct {
	MaintenanceInterval time.Duration `json:"maintenance_interval"`
	ArchivalThreshold   time.Duration `json:"archival_threshold"`
	CleanupBatchSize    int           `json:"cleanup_batch_size"`
	EnableScheduling    bool          `json:"enable_scheduling"`
	MaxRecordsPerUser   int           `json:"max_records_per_user"`
}

// MaintenanceStats represents statistics from a maintenance operation
type MaintenanceStats struct {
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	Duration         string    `json:"duration"` // Duration as string for JSON serialization
	RecordsProcessed int       `json:"records_processed"`
	RecordsArchived  int       `json:"records_archived"`
	RecordsDeleted   int       `json:"records_deleted"`
	OrphanedRecords  int       `json:"orphaned_records"`
	IndexesOptimized int       `json:"indexes_optimized"`
	VacuumOperations int       `json:"vacuum_operations"`
	Errors           int       `json:"errors"`
}

// DatabaseHealthReport represents the health status of the database
type DatabaseHealthReport struct {
	Timestamp          time.Time `json:"timestamp"`
	ConnectionStatus   string    `json:"connection_status"`
	TotalRecords       int64     `json:"total_records"`
	ActiveRecords      int64     `json:"active_records"`
	ExpiredRecords     int64     `json:"expired_records"`
	OrphanedRecords    int64     `json:"orphaned_records"`
	AverageQueryTime   float64   `json:"average_query_time_ms"`
	SlowQueries        int       `json:"slow_queries"`
	IndexEfficiency    float64   `json:"index_efficiency"`
	TableSize          int64     `json:"table_size_mb"`
	FragmentationLevel float64   `json:"fragmentation_level"`
	RecommendedActions []string  `json:"recommended_actions"`
}

// PerformanceMetrics represents database performance metrics
type PerformanceMetrics struct {
	QueryExecutionTime  map[string]float64 `json:"query_execution_time"`
	IndexUsage          map[string]int64   `json:"index_usage"`
	TableScans          int64              `json:"table_scans"`
	IndexScans          int64              `json:"index_scans"`
	CacheHitRatio       float64            `json:"cache_hit_ratio"`
	ConnectionPoolUsage float64            `json:"connection_pool_usage"`
}

// NewDatabaseMaintenanceService creates a new instance of DatabaseMaintenanceService
func NewDatabaseMaintenanceService(db *gorm.DB, config MaintenanceConfig) *DatabaseMaintenanceService {
	// Set default values
	if config.MaintenanceInterval == 0 {
		config.MaintenanceInterval = 24 * time.Hour // Default: daily maintenance
	}
	if config.ArchivalThreshold == 0 {
		config.ArchivalThreshold = 90 * 24 * time.Hour // Default: 90 days
	}
	if config.CleanupBatchSize == 0 {
		config.CleanupBatchSize = 1000 // Default: process 1000 records at a time
	}

	return &DatabaseMaintenanceService{
		db:                  db,
		maintenanceInterval: config.MaintenanceInterval,
		archivalThreshold:   config.ArchivalThreshold,
		cleanupBatchSize:    config.CleanupBatchSize,
		stopChan:            make(chan struct{}),
		logger:              log.New(os.Stdout, "[DBMaintenance] ", log.LstdFlags),
	}
}

// Start begins the automatic database maintenance service
func (s *DatabaseMaintenanceService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("database maintenance service is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Printf("Starting database maintenance service with interval: %v", s.maintenanceInterval)

	go s.runMaintenanceLoop(ctx)
	return nil
}

// Stop stops the automatic maintenance service
func (s *DatabaseMaintenanceService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("database maintenance service is not running")
	}

	s.logger.Println("Stopping database maintenance service...")
	close(s.stopChan)
	s.isRunning = false
	s.logger.Println("Database maintenance service stopped")

	return nil
}

// IsRunning returns whether the maintenance service is currently running
func (s *DatabaseMaintenanceService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// RunMaintenance performs a manual maintenance operation
func (s *DatabaseMaintenanceService) RunMaintenance(ctx context.Context) (*MaintenanceStats, error) {
	s.logger.Println("Starting database maintenance operation...")

	stats := &MaintenanceStats{
		StartTime: time.Now(),
	}

	// Clean up old project records
	if err := s.cleanupOldRecords(ctx, stats); err != nil {
		s.logger.Printf("Error during record cleanup: %v", err)
		stats.Errors++
	}

	// Archive old records
	if err := s.archiveOldRecords(ctx, stats); err != nil {
		s.logger.Printf("Error during record archival: %v", err)
		stats.Errors++
	}

	// Clean up orphaned records
	if err := s.cleanupOrphanedRecords(ctx, stats); err != nil {
		s.logger.Printf("Error during orphaned record cleanup: %v", err)
		stats.Errors++
	}

	// Optimize database performance
	if err := s.optimizeDatabase(ctx, stats); err != nil {
		s.logger.Printf("Error during database optimization: %v", err)
		stats.Errors++
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()

	s.logger.Printf("Maintenance completed: %d records processed, %d archived, %d deleted",
		stats.RecordsProcessed, stats.RecordsArchived, stats.RecordsDeleted)

	return stats, nil
}

// runMaintenanceLoop runs the maintenance operation at regular intervals
func (s *DatabaseMaintenanceService) runMaintenanceLoop(ctx context.Context) {
	ticker := time.NewTicker(s.maintenanceInterval)
	defer ticker.Stop()

	// Run initial maintenance
	if _, err := s.RunMaintenance(ctx); err != nil {
		s.logger.Printf("Initial maintenance failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("Maintenance loop stopped due to context cancellation")
			return
		case <-s.stopChan:
			s.logger.Println("Maintenance loop stopped")
			return
		case <-ticker.C:
			if _, err := s.RunMaintenance(ctx); err != nil {
				s.logger.Printf("Scheduled maintenance failed: %v", err)
			}
		}
	}
}

// cleanupOldRecords removes very old project records that are no longer needed
func (s *DatabaseMaintenanceService) cleanupOldRecords(ctx context.Context, stats *MaintenanceStats) error {
	// Delete records older than archival threshold that are already expired
	cutoffDate := time.Now().Add(-s.archivalThreshold)

	var count int64
	err := s.db.Model(&models.ProjectHistory{}).
		Where("created_at < ? AND zip_file_status IN (?)",
			cutoffDate, []string{string(models.ZipFileStatusExpired), string(models.ZipFileStatusDeleted)}).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to count old records: %w", err)
	}

	if count == 0 {
		return nil
	}

	s.logger.Printf("Found %d old records to delete", count)
	stats.RecordsProcessed += int(count)

	// Delete in batches to avoid long-running transactions
	for {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		result := s.db.Where("created_at < ? AND zip_file_status IN (?)",
			cutoffDate, []string{string(models.ZipFileStatusExpired), string(models.ZipFileStatusDeleted)}).
			Limit(s.cleanupBatchSize).
			Delete(&models.ProjectHistory{})

		if result.Error != nil {
			return fmt.Errorf("failed to delete old records: %w", result.Error)
		}

		deletedCount := int(result.RowsAffected)
		stats.RecordsDeleted += deletedCount

		// If no records were deleted, we're done
		if deletedCount == 0 {
			break
		}

		s.logger.Printf("Deleted batch of %d old records", deletedCount)

		// Add a small delay to prevent overwhelming the database
		time.Sleep(10 * time.Millisecond)
	}

	return nil
}

// archiveOldRecords implements archival strategy for historical data
func (s *DatabaseMaintenanceService) archiveOldRecords(ctx context.Context, stats *MaintenanceStats) error {
	// For now, we'll implement a simple archival by updating a flag
	// In a more complex system, this could move data to a separate archive table

	archiveDate := time.Now().Add(-s.archivalThreshold / 2) // Archive at half the cleanup threshold

	var count int64
	err := s.db.Model(&models.ProjectHistory{}).
		Where("created_at < ? AND zip_file_status = ?", archiveDate, string(models.ZipFileStatusExpired)).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to count records for archival: %w", err)
	}

	if count == 0 {
		return nil
	}

	s.logger.Printf("Found %d records to archive", count)
	stats.RecordsProcessed += int(count)

	// Update records to archived status (we could add an archived status to the enum)
	// For now, we'll just log that these would be archived
	stats.RecordsArchived = int(count)
	s.logger.Printf("Would archive %d records (archival strategy placeholder)", count)

	return nil
}

// cleanupOrphanedRecords removes records that reference non-existent users
func (s *DatabaseMaintenanceService) cleanupOrphanedRecords(ctx context.Context, stats *MaintenanceStats) error {
	// Find project history records that reference non-existent users
	var orphanedRecords []models.ProjectHistory
	err := s.db.Table("project_history").
		Select("project_history.*").
		Joins("LEFT JOIN users ON users.id = project_history.user_id").
		Where("users.id IS NULL").
		Find(&orphanedRecords).Error

	if err != nil {
		return fmt.Errorf("failed to find orphaned records: %w", err)
	}

	if len(orphanedRecords) == 0 {
		return nil
	}

	s.logger.Printf("Found %d orphaned records to clean up", len(orphanedRecords))
	stats.OrphanedRecords = len(orphanedRecords)

	// Delete orphaned records in batches
	for i := 0; i < len(orphanedRecords); i += s.cleanupBatchSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			end := i + s.cleanupBatchSize
			if end > len(orphanedRecords) {
				end = len(orphanedRecords)
			}

			batch := orphanedRecords[i:end]
			var ids []uint
			for _, record := range batch {
				ids = append(ids, record.ID)
			}

			result := s.db.Where("id IN ?", ids).Delete(&models.ProjectHistory{})
			if result.Error != nil {
				return fmt.Errorf("failed to delete orphaned records: %w", result.Error)
			}

			s.logger.Printf("Deleted batch of %d orphaned records", len(batch))
		}
	}

	return nil
}

// optimizeDatabase performs database optimization operations
func (s *DatabaseMaintenanceService) optimizeDatabase(ctx context.Context, stats *MaintenanceStats) error {
	// Analyze table statistics
	if err := s.analyzeTable(ctx); err != nil {
		s.logger.Printf("Failed to analyze table: %v", err)
	} else {
		stats.IndexesOptimized++
	}

	// Vacuum operations (PostgreSQL specific)
	if err := s.vacuumTable(ctx); err != nil {
		s.logger.Printf("Failed to vacuum table: %v", err)
	} else {
		stats.VacuumOperations++
	}

	return nil
}

// analyzeTable updates table statistics for query optimization
func (s *DatabaseMaintenanceService) analyzeTable(ctx context.Context) error {
	// PostgreSQL ANALYZE command
	err := s.db.Exec("ANALYZE project_history").Error
	if err != nil {
		return fmt.Errorf("failed to analyze project_history table: %w", err)
	}

	s.logger.Println("Analyzed project_history table statistics")
	return nil
}

// vacuumTable performs vacuum operation to reclaim space
func (s *DatabaseMaintenanceService) vacuumTable(ctx context.Context) error {
	// PostgreSQL VACUUM command (non-blocking)
	err := s.db.Exec("VACUUM (ANALYZE) project_history").Error
	if err != nil {
		return fmt.Errorf("failed to vacuum project_history table: %w", err)
	}

	s.logger.Println("Vacuumed project_history table")
	return nil
}

// GetDatabaseHealth performs comprehensive health checks
func (s *DatabaseMaintenanceService) GetDatabaseHealth(ctx context.Context) (*DatabaseHealthReport, error) {
	report := &DatabaseHealthReport{
		Timestamp:          time.Now(),
		RecommendedActions: []string{},
	}

	// Test database connection
	sqlDB, err := s.db.DB()
	if err != nil {
		report.ConnectionStatus = "error"
		return report, fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		report.ConnectionStatus = "disconnected"
		return report, fmt.Errorf("database ping failed: %w", err)
	}
	report.ConnectionStatus = "connected"

	// Get record counts
	if err := s.getRecordCounts(report); err != nil {
		s.logger.Printf("Failed to get record counts: %v", err)
	}

	// Measure query performance
	if err := s.measureQueryPerformance(ctx, report); err != nil {
		s.logger.Printf("Failed to measure query performance: %v", err)
	}

	// Check table size and fragmentation
	if err := s.checkTableHealth(ctx, report); err != nil {
		s.logger.Printf("Failed to check table health: %v", err)
	}

	// Generate recommendations
	s.generateRecommendations(report)

	return report, nil
}

// getRecordCounts retrieves various record counts for health reporting
func (s *DatabaseMaintenanceService) getRecordCounts(report *DatabaseHealthReport) error {
	// Total records
	if err := s.db.Model(&models.ProjectHistory{}).Count(&report.TotalRecords).Error; err != nil {
		return fmt.Errorf("failed to count total records: %w", err)
	}

	// Active records
	if err := s.db.Model(&models.ProjectHistory{}).
		Where("zip_file_status = ?", string(models.ZipFileStatusAvailable)).
		Count(&report.ActiveRecords).Error; err != nil {
		return fmt.Errorf("failed to count active records: %w", err)
	}

	// Expired records
	if err := s.db.Model(&models.ProjectHistory{}).
		Where("zip_file_status = ?", string(models.ZipFileStatusExpired)).
		Count(&report.ExpiredRecords).Error; err != nil {
		return fmt.Errorf("failed to count expired records: %w", err)
	}

	// Orphaned records
	var orphanedCount int64
	err := s.db.Table("project_history").
		Joins("LEFT JOIN users ON users.id = project_history.user_id").
		Where("users.id IS NULL").
		Count(&orphanedCount).Error
	if err != nil {
		return fmt.Errorf("failed to count orphaned records: %w", err)
	}
	report.OrphanedRecords = orphanedCount

	return nil
}

// measureQueryPerformance measures the performance of common queries
func (s *DatabaseMaintenanceService) measureQueryPerformance(ctx context.Context, report *DatabaseHealthReport) error {
	// Measure a common query performance
	start := time.Now()

	var count int64
	err := s.db.Model(&models.ProjectHistory{}).
		Where("created_at > ?", time.Now().Add(-30*24*time.Hour)).
		Count(&count).Error

	if err != nil {
		return fmt.Errorf("failed to execute performance test query: %w", err)
	}

	queryTime := time.Since(start)
	report.AverageQueryTime = float64(queryTime.Nanoseconds()) / 1e6 // Convert to milliseconds

	// Consider queries over 100ms as slow
	if queryTime > 100*time.Millisecond {
		report.SlowQueries = 1
	}

	return nil
}

// checkTableHealth checks table size and fragmentation
func (s *DatabaseMaintenanceService) checkTableHealth(ctx context.Context, report *DatabaseHealthReport) error {
	// PostgreSQL specific query to get table size
	var tableSize sql.NullFloat64
	err := s.db.Raw(`
		SELECT pg_total_relation_size('project_history') / 1024.0 / 1024.0 as size_mb
	`).Scan(&tableSize).Error

	if err == nil && tableSize.Valid {
		report.TableSize = int64(tableSize.Float64)
	}

	// Estimate index efficiency (simplified)
	if report.TotalRecords > 0 {
		// If we have records and queries are fast, assume good index efficiency
		if report.AverageQueryTime < 50 {
			report.IndexEfficiency = 0.9
		} else if report.AverageQueryTime < 100 {
			report.IndexEfficiency = 0.7
		} else {
			report.IndexEfficiency = 0.5
		}
	}

	return nil
}

// generateRecommendations generates maintenance recommendations based on health metrics
func (s *DatabaseMaintenanceService) generateRecommendations(report *DatabaseHealthReport) {
	if report.OrphanedRecords > 0 {
		report.RecommendedActions = append(report.RecommendedActions,
			fmt.Sprintf("Clean up %d orphaned records", report.OrphanedRecords))
	}

	if report.ExpiredRecords > 1000 {
		report.RecommendedActions = append(report.RecommendedActions,
			fmt.Sprintf("Consider archiving %d expired records", report.ExpiredRecords))
	}

	if report.AverageQueryTime > 100 {
		report.RecommendedActions = append(report.RecommendedActions,
			"Query performance is slow, consider index optimization")
	}

	if report.TableSize > 1000 { // > 1GB
		report.RecommendedActions = append(report.RecommendedActions,
			"Table size is large, consider implementing data archival")
	}

	if report.IndexEfficiency < 0.7 {
		report.RecommendedActions = append(report.RecommendedActions,
			"Index efficiency is low, consider rebuilding indexes")
	}

	if len(report.RecommendedActions) == 0 {
		report.RecommendedActions = append(report.RecommendedActions, "Database health is good")
	}
}

// GetPerformanceMetrics returns detailed performance metrics
func (s *DatabaseMaintenanceService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	metrics := &PerformanceMetrics{
		QueryExecutionTime: make(map[string]float64),
		IndexUsage:         make(map[string]int64),
	}

	// Measure common query execution times
	queries := map[string]string{
		"user_history":    "SELECT * FROM project_history WHERE user_id = ? ORDER BY created_at DESC LIMIT 20",
		"recent_projects": "SELECT * FROM project_history WHERE created_at > ? ORDER BY created_at DESC",
		"framework_count": "SELECT framework, COUNT(*) FROM project_history GROUP BY framework",
	}

	for queryName := range queries {
		start := time.Now()

		switch queryName {
		case "user_history":
			var projects []models.ProjectHistory
			s.db.Where("user_id = ?", 1).Order("created_at DESC").Limit(20).Find(&projects)
		case "recent_projects":
			var projects []models.ProjectHistory
			s.db.Where("created_at > ?", time.Now().Add(-7*24*time.Hour)).Order("created_at DESC").Find(&projects)
		case "framework_count":
			var results []struct {
				Framework string
				Count     int64
			}
			s.db.Model(&models.ProjectHistory{}).Select("framework, COUNT(*) as count").Group("framework").Scan(&results)
		}

		duration := time.Since(start)
		metrics.QueryExecutionTime[queryName] = float64(duration.Nanoseconds()) / 1e6
	}

	return metrics, nil
}

// GetMaintenanceConfig returns the current maintenance configuration
func (s *DatabaseMaintenanceService) GetMaintenanceConfig() MaintenanceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return MaintenanceConfig{
		MaintenanceInterval: s.maintenanceInterval,
		ArchivalThreshold:   s.archivalThreshold,
		CleanupBatchSize:    s.cleanupBatchSize,
		EnableScheduling:    s.isRunning,
	}
}

// UpdateMaintenanceConfig updates the maintenance configuration
func (s *DatabaseMaintenanceService) UpdateMaintenanceConfig(config MaintenanceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config.MaintenanceInterval > 0 {
		s.maintenanceInterval = config.MaintenanceInterval
	}
	if config.ArchivalThreshold > 0 {
		s.archivalThreshold = config.ArchivalThreshold
	}
	if config.CleanupBatchSize > 0 {
		s.cleanupBatchSize = config.CleanupBatchSize
	}

	s.logger.Printf("Updated maintenance configuration: interval=%v, archival_threshold=%v, batch_size=%d",
		s.maintenanceInterval, s.archivalThreshold, s.cleanupBatchSize)

	return nil
}
