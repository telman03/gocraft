package unit

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/services"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBMaintenance(t *testing.T) (*gorm.DB, *services.DatabaseMaintenanceService) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.ProjectHistory{})
	require.NoError(t, err)

	// Create maintenance service with test configuration
	config := services.MaintenanceConfig{
		MaintenanceInterval: time.Hour,
		ArchivalThreshold:   7 * 24 * time.Hour, // 7 days for testing
		CleanupBatchSize:    10,
		EnableScheduling:    false, // Don't start automatic scheduling in tests
	}

	service := services.NewDatabaseMaintenanceService(db, config)
	return db, service
}

func createTestProjectHistory(t *testing.T, db *gorm.DB, userID uint, createdAt time.Time, status models.ZipFileStatus) *models.ProjectHistory {
	features := `["auth", "database"]`
	project := &models.ProjectHistory{
		UserID:           userID,
		ProjectName:      "test-project",
		Framework:        "gin",
		Features:         datatypes.JSON(features),
		AdjustedFeatures: datatypes.JSON(features),
		ZipFilePath:      stringPtr("/path/to/file.zip"),
		ZipFileSize:      int64Ptr(1024),
		ZipFileStatus:    status,
		CreatedAt:        createdAt,
		UpdatedAt:        createdAt,
	}
	err := db.Create(project).Error
	require.NoError(t, err)
	return project
}

func TestDatabaseMaintenanceService_GetDatabaseHealth(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create some test data
	now := time.Now()
	createTestProjectHistory(t, db, user.ID, now, models.ZipFileStatusAvailable)
	createTestProjectHistory(t, db, user.ID, now.Add(-time.Hour), models.ZipFileStatusExpired)

	ctx := context.Background()
	health, err := service.GetDatabaseHealth(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "connected", health.ConnectionStatus)
	assert.Equal(t, int64(2), health.TotalRecords)
	assert.Equal(t, int64(1), health.ActiveRecords)
	assert.Equal(t, int64(1), health.ExpiredRecords)
	assert.Equal(t, int64(0), health.OrphanedRecords)
	assert.NotEmpty(t, health.RecommendedActions)
}

func TestDatabaseMaintenanceService_CleanupOldRecords(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create old expired records that should be cleaned up
	oldDate := time.Now().Add(-10 * 24 * time.Hour) // 10 days old
	createTestProjectHistory(t, db, user.ID, oldDate, models.ZipFileStatusExpired)
	createTestProjectHistory(t, db, user.ID, oldDate, models.ZipFileStatusDeleted)

	// Create recent records that should not be cleaned up
	recentDate := time.Now().Add(-1 * time.Hour)
	createTestProjectHistory(t, db, user.ID, recentDate, models.ZipFileStatusAvailable)

	ctx := context.Background()
	stats, err := service.RunMaintenance(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats.RecordsDeleted) // Should delete the 2 old expired records

	// Verify records were actually deleted
	var count int64
	err = db.Model(&models.ProjectHistory{}).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count) // Only the recent record should remain
}

func TestDatabaseMaintenanceService_CleanupOrphanedRecords(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create a project history record
	project := createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusAvailable)

	// Delete the user to create an orphaned record
	err := db.Delete(user).Error
	require.NoError(t, err)

	ctx := context.Background()
	stats, err := service.RunMaintenance(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 1, stats.OrphanedRecords)

	// Verify the orphaned record was cleaned up
	var count int64
	err = db.Model(&models.ProjectHistory{}).Where("id = ?", project.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDatabaseMaintenanceService_GetPerformanceMetrics(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create some test data
	createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusAvailable)
	createTestProjectHistory(t, db, user.ID, time.Now().Add(-time.Hour), models.ZipFileStatusExpired)

	ctx := context.Background()
	metrics, err := service.GetPerformanceMetrics(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.NotEmpty(t, metrics.QueryExecutionTime)
	assert.Contains(t, metrics.QueryExecutionTime, "user_history")
	assert.Contains(t, metrics.QueryExecutionTime, "recent_projects")
	assert.Contains(t, metrics.QueryExecutionTime, "framework_count")
}

func TestDatabaseMaintenanceService_StartStop(t *testing.T) {
	_, service := setupTestDBMaintenance(t)

	// Initially should not be running
	assert.False(t, service.IsRunning())

	// Start the service
	ctx := context.Background()
	err := service.Start(ctx)
	assert.NoError(t, err)
	assert.True(t, service.IsRunning())

	// Starting again should return an error
	err = service.Start(ctx)
	assert.Error(t, err)

	// Stop the service
	err = service.Stop()
	assert.NoError(t, err)
	assert.False(t, service.IsRunning())

	// Stopping again should return an error
	err = service.Stop()
	assert.Error(t, err)
}

func TestDatabaseMaintenanceService_UpdateConfig(t *testing.T) {
	_, service := setupTestDBMaintenance(t)

	// Get initial config
	initialConfig := service.GetMaintenanceConfig()
	assert.Equal(t, time.Hour, initialConfig.MaintenanceInterval)

	// Update config
	newConfig := services.MaintenanceConfig{
		MaintenanceInterval: 2 * time.Hour,
		ArchivalThreshold:   14 * 24 * time.Hour,
		CleanupBatchSize:    20,
	}

	err := service.UpdateMaintenanceConfig(newConfig)
	assert.NoError(t, err)

	// Verify config was updated
	updatedConfig := service.GetMaintenanceConfig()
	assert.Equal(t, 2*time.Hour, updatedConfig.MaintenanceInterval)
	assert.Equal(t, 14*24*time.Hour, updatedConfig.ArchivalThreshold)
	assert.Equal(t, 20, updatedConfig.CleanupBatchSize)
}

func TestDatabaseMaintenanceService_GenerateRecommendations(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create many expired records to trigger recommendations
	for i := 0; i < 1500; i++ {
		createTestProjectHistory(t, db, user.ID, time.Now().Add(-time.Hour), models.ZipFileStatusExpired)
	}

	// Delete the user to create orphaned records
	err := db.Delete(user).Error
	require.NoError(t, err)

	ctx := context.Background()
	health, err := service.GetDatabaseHealth(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.NotEmpty(t, health.RecommendedActions)

	// Should recommend cleaning up orphaned records
	foundOrphanedRecommendation := false
	foundExpiredRecommendation := false
	for _, action := range health.RecommendedActions {
		if contains(action, "orphaned") {
			foundOrphanedRecommendation = true
		}
		if contains(action, "expired") {
			foundExpiredRecommendation = true
		}
	}

	assert.True(t, foundOrphanedRecommendation, "Should recommend cleaning orphaned records")
	assert.True(t, foundExpiredRecommendation, "Should recommend archiving expired records")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		 containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDatabaseMaintenanceService_RecordCounts(t *testing.T) {
	db, service := setupTestDBMaintenance(t)
	user := createTestUser(t, db)

	// Create records with different statuses
	createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusAvailable)
	createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusAvailable)
	createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusExpired)
	createTestProjectHistory(t, db, user.ID, time.Now(), models.ZipFileStatusDeleted)

	ctx := context.Background()
	health, err := service.GetDatabaseHealth(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int64(4), health.TotalRecords)
	assert.Equal(t, int64(2), health.ActiveRecords)
	assert.Equal(t, int64(1), health.ExpiredRecords)
	assert.Equal(t, int64(0), health.OrphanedRecords)
}