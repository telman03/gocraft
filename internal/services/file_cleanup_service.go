package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/gorm"
)

// FileCleanupService handles automatic cleanup and maintenance of project files
type FileCleanupService struct {
	db                *gorm.DB
	fileService       *FileService
	cleanupInterval   time.Duration
	batchSize         int
	maxConcurrency    int
	retentionPeriod   time.Duration
	isRunning         bool
	stopChan          chan struct{}
	mu                sync.RWMutex
	logger            *log.Logger
}

// CleanupConfig holds configuration for the cleanup service
type CleanupConfig struct {
	CleanupInterval   time.Duration `json:"cleanup_interval"`
	BatchSize         int           `json:"batch_size"`
	MaxConcurrency    int           `json:"max_concurrency"`
	RetentionPeriod   time.Duration `json:"retention_period"`
	EnableScheduling  bool          `json:"enable_scheduling"`
}

// CleanupStats represents statistics from a cleanup operation
type CleanupStats struct {
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	Duration          string    `json:"duration"` // Duration as string for JSON serialization
	FilesProcessed    int       `json:"files_processed"`
	FilesDeleted      int       `json:"files_deleted"`
	FilesSkipped      int       `json:"files_skipped"`
	Errors            int       `json:"errors"`
	BytesFreed        int64     `json:"bytes_freed"`
	DatabaseUpdates   int       `json:"database_updates"`
}

// NewFileCleanupService creates a new instance of FileCleanupService
func NewFileCleanupService(db *gorm.DB, fileService *FileService, config CleanupConfig) *FileCleanupService {
	// Set default values
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 24 * time.Hour // Default: daily cleanup
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100 // Default: process 100 files at a time
	}
	if config.MaxConcurrency == 0 {
		config.MaxConcurrency = 5 // Default: 5 concurrent operations
	}
	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 30 * 24 * time.Hour // Default: 30 days
	}

	return &FileCleanupService{
		db:              db,
		fileService:     fileService,
		cleanupInterval: config.CleanupInterval,
		batchSize:       config.BatchSize,
		maxConcurrency:  config.MaxConcurrency,
		retentionPeriod: config.RetentionPeriod,
		stopChan:        make(chan struct{}),
		logger:          log.New(os.Stdout, "[FileCleanup] ", log.LstdFlags),
	}
}

// Start begins the automatic cleanup service
func (s *FileCleanupService) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return fmt.Errorf("cleanup service is already running")
	}
	s.isRunning = true
	s.mu.Unlock()

	s.logger.Printf("Starting file cleanup service with interval: %v", s.cleanupInterval)

	go s.runCleanupLoop(ctx)
	return nil
}

// Stop stops the automatic cleanup service
func (s *FileCleanupService) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.isRunning {
		return fmt.Errorf("cleanup service is not running")
	}

	s.logger.Println("Stopping file cleanup service...")
	close(s.stopChan)
	s.isRunning = false
	s.logger.Println("File cleanup service stopped")

	return nil
}

// IsRunning returns whether the cleanup service is currently running
func (s *FileCleanupService) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// RunCleanup performs a manual cleanup operation
func (s *FileCleanupService) RunCleanup(ctx context.Context) (*CleanupStats, error) {
	s.logger.Println("Starting manual cleanup operation...")
	
	stats := &CleanupStats{
		StartTime: time.Now(),
	}

	// Clean up expired files
	expiredStats, err := s.cleanupExpiredFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to cleanup expired files: %w", err)
	}

	// Clean up orphaned files (files that exist but have no database record)
	orphanedStats, err := s.cleanupOrphanedFiles(ctx)
	if err != nil {
		s.logger.Printf("Warning: failed to cleanup orphaned files: %v", err)
		// Don't fail the entire operation for orphaned file cleanup
	}

	// Merge statistics
	stats.FilesProcessed = expiredStats.FilesProcessed + orphanedStats.FilesProcessed
	stats.FilesDeleted = expiredStats.FilesDeleted + orphanedStats.FilesDeleted
	stats.FilesSkipped = expiredStats.FilesSkipped + orphanedStats.FilesSkipped
	stats.Errors = expiredStats.Errors + orphanedStats.Errors
	stats.BytesFreed = expiredStats.BytesFreed + orphanedStats.BytesFreed
	stats.DatabaseUpdates = expiredStats.DatabaseUpdates + orphanedStats.DatabaseUpdates

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()

	s.logger.Printf("Cleanup completed: %d files processed, %d deleted, %d MB freed", 
		stats.FilesProcessed, stats.FilesDeleted, stats.BytesFreed/(1024*1024))

	return stats, nil
}

// runCleanupLoop runs the cleanup operation at regular intervals
func (s *FileCleanupService) runCleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	// Run initial cleanup
	if _, err := s.RunCleanup(ctx); err != nil {
		s.logger.Printf("Initial cleanup failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			s.logger.Println("Cleanup loop stopped due to context cancellation")
			return
		case <-s.stopChan:
			s.logger.Println("Cleanup loop stopped")
			return
		case <-ticker.C:
			if _, err := s.RunCleanup(ctx); err != nil {
				s.logger.Printf("Scheduled cleanup failed: %v", err)
			}
		}
	}
}

// cleanupExpiredFiles removes files that have exceeded the retention period
func (s *FileCleanupService) cleanupExpiredFiles(ctx context.Context) (*CleanupStats, error) {
	stats := &CleanupStats{
		StartTime: time.Now(),
	}

	// Find projects with files that should be expired
	cutoffDate := time.Now().Add(-s.retentionPeriod)
	
	var projects []models.ProjectHistory
	err := s.db.Where("zip_file_path IS NOT NULL AND zip_file_path != '' AND zip_file_status = ? AND created_at < ?", 
		models.ZipFileStatusAvailable, cutoffDate).
		Find(&projects).Error
	
	if err != nil {
		return stats, fmt.Errorf("failed to query expired projects: %w", err)
	}

	stats.FilesProcessed = len(projects)

	if len(projects) == 0 {
		s.logger.Println("No expired files found")
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime).String()
		return stats, nil
	}

	s.logger.Printf("Found %d expired files to process", len(projects))

	// Process files in batches with concurrency control
	semaphore := make(chan struct{}, s.maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < len(projects); i += s.batchSize {
		end := i + s.batchSize
		if end > len(projects) {
			end = len(projects)
		}

		batch := projects[i:end]
		
		wg.Add(1)
		go func(batch []models.ProjectHistory) {
			defer wg.Done()
			
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
				
				batchStats := s.processBatch(ctx, batch)
				
				mu.Lock()
				stats.FilesDeleted += batchStats.FilesDeleted
				stats.FilesSkipped += batchStats.FilesSkipped
				stats.Errors += batchStats.Errors
				stats.BytesFreed += batchStats.BytesFreed
				stats.DatabaseUpdates += batchStats.DatabaseUpdates
				mu.Unlock()
				
			case <-ctx.Done():
				return
			}
		}(batch)
	}

	wg.Wait()

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()

	return stats, nil
}

// cleanupOrphanedFiles removes files that exist on disk but have no database record
func (s *FileCleanupService) cleanupOrphanedFiles(ctx context.Context) (*CleanupStats, error) {
	stats := &CleanupStats{
		StartTime: time.Now(),
	}

	basePath := s.fileService.GetBasePath()
	if basePath == "" {
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime).String()
		return stats, nil
	}

	// Walk through the file system to find all ZIP files
	var orphanedFiles []string
	
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Check if this is a ZIP file
		if filepath.Ext(path) != ".zip" {
			return nil
		}

		// Check if this file is referenced in the database
		var count int64
		dbErr := s.db.Model(&models.ProjectHistory{}).Where("zip_file_path = ?", path).Count(&count).Error
		if dbErr != nil {
			s.logger.Printf("Warning: failed to check database for file %s: %v", path, dbErr)
			return nil
		}

		if count == 0 {
			orphanedFiles = append(orphanedFiles, path)
		}

		return nil
	})

	if err != nil {
		return stats, fmt.Errorf("failed to walk file system: %w", err)
	}

	stats.FilesProcessed = len(orphanedFiles)

	if len(orphanedFiles) == 0 {
		s.logger.Println("No orphaned files found")
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime).String()
		return stats, nil
	}

	s.logger.Printf("Found %d orphaned files to remove", len(orphanedFiles))

	// Remove orphaned files
	for _, filePath := range orphanedFiles {
		select {
		case <-ctx.Done():
			goto cleanup_done
		default:
			// Get file size before deletion
			if size, err := s.fileService.GetFileSize(filePath); err == nil {
				stats.BytesFreed += size
			}

			// Delete the file
			if err := s.fileService.DeleteFile(filePath); err != nil {
				s.logger.Printf("Failed to delete orphaned file %s: %v", filePath, err)
				stats.Errors++
			} else {
				stats.FilesDeleted++
				s.logger.Printf("Deleted orphaned file: %s", filePath)
			}
		}
	}

cleanup_done:
	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime).String()

	return stats, nil
}

// processBatch processes a batch of projects for cleanup
func (s *FileCleanupService) processBatch(ctx context.Context, projects []models.ProjectHistory) *CleanupStats {
	stats := &CleanupStats{}

	for _, project := range projects {
		select {
		case <-ctx.Done():
			return stats
		default:
			if s.processProject(project, stats) {
				stats.DatabaseUpdates++
			}
		}
	}

	return stats
}

// processProject processes a single project for cleanup
func (s *FileCleanupService) processProject(project models.ProjectHistory, stats *CleanupStats) bool {
	if project.ZipFilePath == nil || *project.ZipFilePath == "" {
		stats.FilesSkipped++
		return false
	}

	filePath := *project.ZipFilePath

	// Get file size before deletion for statistics
	if size, err := s.fileService.GetFileSize(filePath); err == nil {
		stats.BytesFreed += size
	}

	// Delete the file
	if err := s.fileService.DeleteFile(filePath); err != nil {
		s.logger.Printf("Failed to delete file %s for project %d: %v", filePath, project.ID, err)
		stats.Errors++
		return false
	}

	// Update database record
	err := s.db.Model(&project).Updates(map[string]interface{}{
		"zip_file_status": models.ZipFileStatusExpired,
		"updated_at":      time.Now(),
	}).Error

	if err != nil {
		s.logger.Printf("Failed to update project %d status: %v", project.ID, err)
		stats.Errors++
		return false
	}

	stats.FilesDeleted++
	s.logger.Printf("Cleaned up file for project %d: %s", project.ID, filePath)
	return true
}

// ValidateFileIntegrity checks if files referenced in the database actually exist
func (s *FileCleanupService) ValidateFileIntegrity(ctx context.Context) (*IntegrityReport, error) {
	report := &IntegrityReport{
		StartTime: time.Now(),
	}

	// Get all projects with file paths
	var projects []models.ProjectHistory
	err := s.db.Where("zip_file_path IS NOT NULL AND zip_file_path != ''").Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}

	report.TotalRecords = len(projects)

	for _, project := range projects {
		select {
		case <-ctx.Done():
			return report, ctx.Err()
		default:
			s.validateProjectFile(project, report)
		}
	}

	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime).String()

	return report, nil
}

// IntegrityReport represents the results of a file integrity check
type IntegrityReport struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Duration        string    `json:"duration"` // Duration as string for JSON serialization
	TotalRecords    int       `json:"total_records"`
	ValidFiles      int       `json:"valid_files"`
	MissingFiles    int       `json:"missing_files"`
	InvalidPaths    int       `json:"invalid_paths"`
	StatusMismatches int      `json:"status_mismatches"`
	FixedRecords    int       `json:"fixed_records"`
}

// validateProjectFile validates a single project's file integrity
func (s *FileCleanupService) validateProjectFile(project models.ProjectHistory, report *IntegrityReport) {
	if project.ZipFilePath == nil || *project.ZipFilePath == "" {
		return
	}

	filePath := *project.ZipFilePath

	// Validate file path
	if err := s.fileService.ValidateFilePath(filePath); err != nil {
		report.InvalidPaths++
		s.logger.Printf("Invalid file path for project %d: %s", project.ID, filePath)
		return
	}

	// Check if file exists
	exists := s.fileService.FileExists(filePath)
	
	if exists {
		report.ValidFiles++
		
		// Check if status matches reality
		if project.ZipFileStatus != models.ZipFileStatusAvailable {
			report.StatusMismatches++
			
			// Fix the status
			if err := s.db.Model(&project).Update("zip_file_status", models.ZipFileStatusAvailable).Error; err != nil {
				s.logger.Printf("Failed to fix status for project %d: %v", project.ID, err)
			} else {
				report.FixedRecords++
				s.logger.Printf("Fixed status for project %d", project.ID)
			}
		}
	} else {
		report.MissingFiles++
		
		// Update status to expired if file is missing
		if project.ZipFileStatus == models.ZipFileStatusAvailable {
			if err := s.db.Model(&project).Update("zip_file_status", models.ZipFileStatusExpired).Error; err != nil {
				s.logger.Printf("Failed to update status for missing file, project %d: %v", project.ID, err)
			} else {
				report.FixedRecords++
				s.logger.Printf("Updated status for missing file, project %d", project.ID)
			}
		}
	}
}

// GetCleanupConfig returns the current cleanup configuration
func (s *FileCleanupService) GetCleanupConfig() CleanupConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return CleanupConfig{
		CleanupInterval:  s.cleanupInterval,
		BatchSize:        s.batchSize,
		MaxConcurrency:   s.maxConcurrency,
		RetentionPeriod:  s.retentionPeriod,
		EnableScheduling: s.isRunning,
	}
}

// UpdateCleanupConfig updates the cleanup configuration
func (s *FileCleanupService) UpdateCleanupConfig(config CleanupConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if config.CleanupInterval > 0 {
		s.cleanupInterval = config.CleanupInterval
	}
	if config.BatchSize > 0 {
		s.batchSize = config.BatchSize
	}
	if config.MaxConcurrency > 0 {
		s.maxConcurrency = config.MaxConcurrency
	}
	if config.RetentionPeriod > 0 {
		s.retentionPeriod = config.RetentionPeriod
	}
	
	s.logger.Printf("Updated cleanup configuration: interval=%v, batch_size=%d, max_concurrency=%d, retention=%v", 
		s.cleanupInterval, s.batchSize, s.maxConcurrency, s.retentionPeriod)
	
	return nil
}

// GetStorageStats returns statistics about file storage usage
func (s *FileCleanupService) GetStorageStats() (*StorageStats, error) {
	stats := &StorageStats{}
	
	// Count database records by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	
	err := s.db.Model(&models.ProjectHistory{}).
		Select("zip_file_status as status, COUNT(*) as count").
		Where("zip_file_path IS NOT NULL AND zip_file_path != ''").
		Group("zip_file_status").
		Scan(&statusCounts).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get status counts: %w", err)
	}
	
	for _, sc := range statusCounts {
		switch sc.Status {
		case string(models.ZipFileStatusAvailable):
			stats.AvailableFiles = int(sc.Count)
		case string(models.ZipFileStatusExpired):
			stats.ExpiredFiles = int(sc.Count)
		case string(models.ZipFileStatusDeleted):
			stats.DeletedFiles = int(sc.Count)
		}
	}
	
	// Calculate total storage usage by walking the file system
	basePath := s.fileService.GetBasePath()
	if basePath != "" {
		err = filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip files we can't access
			}
			
			if !info.IsDir() && filepath.Ext(path) == ".zip" {
				stats.TotalFiles++
				stats.TotalSize += info.Size()
			}
			
			return nil
		})
		
		if err != nil {
			s.logger.Printf("Warning: failed to calculate storage usage: %v", err)
		}
	}
	
	return stats, nil
}

// StorageStats represents storage usage statistics
type StorageStats struct {
	TotalFiles     int   `json:"total_files"`
	TotalSize      int64 `json:"total_size"`
	AvailableFiles int   `json:"available_files"`
	ExpiredFiles   int   `json:"expired_files"`
	DeletedFiles   int   `json:"deleted_files"`
}