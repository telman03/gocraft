package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupPerformanceTestDB(t testing.TB) (*gorm.DB, *ProjectHistoryService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.ProjectHistory{})
	require.NoError(t, err)

	service := NewProjectHistoryService(db)
	return db, service
}

func createTestUserForPerformance(t testing.TB, db *gorm.DB, email string) *models.User {
	user := &models.User{
		Email:      email,
		Password:   "hashedpassword",
		IsVerified: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createBulkTestProjects(t testing.TB, db *gorm.DB, userID uint, count int) {
	frameworks := []string{"gin", "echo", "fiber"}
	features := [][]string{
		{"auth", "database"},
		{"auth", "redis"},
		{"websocket", "grpc"},
		{"auth", "database", "redis"},
		{"grpc", "websocket", "auth"},
	}

	projects := make([]models.ProjectHistory, count)
	for i := 0; i < count; i++ {
		framework := frameworks[i%len(frameworks)]
		projectFeatures := features[i%len(features)]
		
		featuresJSON, err := json.Marshal(projectFeatures)
		require.NoError(t, err)

		adjustedFeaturesJSON, err := json.Marshal(projectFeatures)
		require.NoError(t, err)

		zipPath := fmt.Sprintf("/path/to/project_%d.zip", i)
		zipSize := int64(1024 + i*100) // Varying file sizes
		duration := 1000 + i*10        // Varying durations

		projects[i] = models.ProjectHistory{
			UserID:               userID,
			ProjectName:          fmt.Sprintf("project-%d", i),
			Framework:            framework,
			Features:             featuresJSON,
			AdjustedFeatures:     adjustedFeaturesJSON,
			ZipFilePath:          &zipPath,
			ZipFileSize:          &zipSize,
			ZipFileStatus:        models.ZipFileStatusAvailable,
			GenerationDurationMs: &duration,
			CreatedAt:            time.Now().Add(-time.Duration(i) * time.Hour), // Spread over time
			UpdatedAt:            time.Now().Add(-time.Duration(i) * time.Hour),
		}
	}

	// Bulk insert for better performance
	err := db.CreateInBatches(projects, 100).Error
	require.NoError(t, err)
}

func BenchmarkProjectHistoryService_GetUserHistory(b *testing.B) {
	db, service := setupPerformanceTestDB(b)
	user := createTestUserForPerformance(b, db, "bench@example.com")

	// Create test data
	createBulkTestProjects(b, db, user.ID, 1000)

	filters := models.HistoryFilters{
		Page:     1,
		PageSize: 20,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetUserHistory(user.ID, filters)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProjectHistoryService_GetUserHistoryWithFilters(b *testing.B) {
	db, service := setupPerformanceTestDB(b)
	user := createTestUserForPerformance(b, db, "bench@example.com")

	// Create test data
	createBulkTestProjects(b, db, user.ID, 1000)

	filters := models.HistoryFilters{
		Page:      1,
		PageSize:  20,
		Framework: "gin",
		SortBy:    "created_at",
		SortOrder: "desc",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetUserHistory(user.ID, filters)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProjectHistoryService_GetProjectStats(b *testing.B) {
	db, service := setupPerformanceTestDB(b)
	user := createTestUserForPerformance(b, db, "bench@example.com")

	// Create test data
	createBulkTestProjects(b, db, user.ID, 1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := service.GetProjectStats(user.ID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkProjectHistoryService_CreateProjectRecord(b *testing.B) {
	db, service := setupPerformanceTestDB(b)
	user := createTestUserForPerformance(b, db, "bench@example.com")

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := models.CreateProjectRecordRequest{
			UserID:           user.ID,
			ProjectName:      fmt.Sprintf("bench-project-%d", i),
			Framework:        "gin",
			Features:         []string{"auth", "database"},
			AdjustedFeatures: []string{"auth", "database"},
		}

		_, err := service.CreateProjectRecord(req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestProjectHistoryService_LargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	db, service := setupPerformanceTestDB(t)
	user := createTestUserForPerformance(t, db, "large@example.com")

	// Create a large dataset
	const projectCount = 10000
	t.Logf("Creating %d test projects...", projectCount)
	
	start := time.Now()
	createBulkTestProjects(t, db, user.ID, projectCount)
	createDuration := time.Since(start)
	t.Logf("Created %d projects in %v", projectCount, createDuration)

	// Test query performance with large dataset
	tests := []struct {
		name    string
		filters models.HistoryFilters
	}{
		{
			name: "basic pagination",
			filters: models.HistoryFilters{
				Page:     1,
				PageSize: 50,
			},
		},
		{
			name: "framework filter",
			filters: models.HistoryFilters{
				Page:      1,
				PageSize:  50,
				Framework: "gin",
			},
		},
		{
			name: "date range filter",
			filters: models.HistoryFilters{
				Page:     1,
				PageSize: 50,
				DateFrom: time.Now().Add(-100 * time.Hour).Format("2006-01-02"),
				DateTo:   time.Now().Format("2006-01-02"),
			},
		},
		{
			name: "sorting by size",
			filters: models.HistoryFilters{
				Page:      1,
				PageSize:  50,
				SortBy:    "zip_file_size",
				SortOrder: "desc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			response, err := service.GetUserHistory(user.ID, tt.filters)
			duration := time.Since(start)

			assert.NoError(t, err)
			assert.NotNil(t, response)
			assert.Equal(t, projectCount, response.Total)
			assert.LessOrEqual(t, len(response.Projects), tt.filters.PageSize)

			// Performance assertion - queries should complete within reasonable time
			assert.Less(t, duration, 500*time.Millisecond, "Query took too long: %v", duration)
			t.Logf("Query '%s' completed in %v", tt.name, duration)
		})
	}

	// Test statistics performance with large dataset
	t.Run("statistics performance", func(t *testing.T) {
		start := time.Now()
		stats, err := service.GetProjectStats(user.ID)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, projectCount, stats.TotalProjects)
		assert.NotEmpty(t, stats.FrameworkDistribution)
		assert.NotEmpty(t, stats.MostUsedFeatures)

		// Performance assertion
		assert.Less(t, duration, 1*time.Second, "Statistics query took too long: %v", duration)
		t.Logf("Statistics query completed in %v", duration)
	})
}

func TestProjectHistoryService_ConcurrentOperations(t *testing.T) {
	db, service := setupPerformanceTestDB(t)
	user := createTestUserForPerformance(t, db, "concurrent@example.com")

	// Create some initial data
	createBulkTestProjects(t, db, user.ID, 100)

	const numGoroutines = 10
	const operationsPerGoroutine = 20

	t.Run("concurrent reads", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*operationsPerGoroutine)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				filters := models.HistoryFilters{
					Page:     1,
					PageSize: 10,
				}

				for j := 0; j < operationsPerGoroutine; j++ {
					_, err := service.GetUserHistory(user.ID, filters)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d, operation %d: %w", goroutineID, j, err)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		duration := time.Since(start)

		// Check for errors
		var errorCount int
		for err := range errors {
			t.Error(err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "Expected no errors in concurrent reads")
		t.Logf("Completed %d concurrent read operations in %v", numGoroutines*operationsPerGoroutine, duration)
	})

	t.Run("concurrent writes", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*operationsPerGoroutine)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					req := models.CreateProjectRecordRequest{
						UserID:           user.ID,
						ProjectName:      fmt.Sprintf("concurrent-project-%d-%d", goroutineID, j),
						Framework:        "gin",
						Features:         []string{"auth"},
						AdjustedFeatures: []string{"auth"},
					}

					_, err := service.CreateProjectRecord(req)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d, operation %d: %w", goroutineID, j, err)
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		duration := time.Since(start)

		// Check for errors
		var errorCount int
		for err := range errors {
			t.Error(err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "Expected no errors in concurrent writes")
		t.Logf("Completed %d concurrent write operations in %v", numGoroutines*operationsPerGoroutine, duration)

		// Verify all records were created
		filters := models.HistoryFilters{Page: 1, PageSize: 1000}
		response, err := service.GetUserHistory(user.ID, filters)
		assert.NoError(t, err)
		expectedTotal := 100 + (numGoroutines * operationsPerGoroutine) // Initial + concurrent writes
		assert.Equal(t, expectedTotal, response.Total)
	})

	t.Run("mixed concurrent operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*operationsPerGoroutine)

		start := time.Now()

		// Half goroutines do reads, half do writes
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			if i%2 == 0 {
				// Read operations
				go func(goroutineID int) {
					defer wg.Done()
					
					filters := models.HistoryFilters{
						Page:     1,
						PageSize: 10,
					}

					for j := 0; j < operationsPerGoroutine; j++ {
						_, err := service.GetUserHistory(user.ID, filters)
						if err != nil {
							errors <- fmt.Errorf("read goroutine %d, operation %d: %w", goroutineID, j, err)
						}
					}
				}(i)
			} else {
				// Write operations
				go func(goroutineID int) {
					defer wg.Done()

					for j := 0; j < operationsPerGoroutine; j++ {
						req := models.CreateProjectRecordRequest{
							UserID:           user.ID,
							ProjectName:      fmt.Sprintf("mixed-project-%d-%d", goroutineID, j),
							Framework:        "echo",
							Features:         []string{"redis"},
							AdjustedFeatures: []string{"redis"},
						}

						_, err := service.CreateProjectRecord(req)
						if err != nil {
							errors <- fmt.Errorf("write goroutine %d, operation %d: %w", goroutineID, j, err)
						}
					}
				}(i)
			}
		}

		wg.Wait()
		close(errors)
		duration := time.Since(start)

		// Check for errors
		var errorCount int
		for err := range errors {
			t.Error(err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "Expected no errors in mixed concurrent operations")
		t.Logf("Completed %d mixed concurrent operations in %v", numGoroutines*operationsPerGoroutine, duration)
	})
}

func TestProjectHistoryService_CachePerformance(t *testing.T) {
	db, service := setupPerformanceTestDB(t)
	user := createTestUserForPerformance(t, db, "cache@example.com")

	// Create test data
	createBulkTestProjects(t, db, user.ID, 1000)

	// First call should calculate stats (cache miss)
	start := time.Now()
	stats1, err := service.GetProjectStats(user.ID)
	firstCallDuration := time.Since(start)
	assert.NoError(t, err)
	assert.NotNil(t, stats1)

	// Second call should use cache (cache hit)
	start = time.Now()
	stats2, err := service.GetProjectStats(user.ID)
	secondCallDuration := time.Since(start)
	assert.NoError(t, err)
	assert.NotNil(t, stats2)

	// Cache hit should be significantly faster
	assert.Less(t, secondCallDuration, firstCallDuration/2, 
		"Cache hit (%v) should be much faster than cache miss (%v)", 
		secondCallDuration, firstCallDuration)

	// Results should be identical
	assert.Equal(t, stats1.TotalProjects, stats2.TotalProjects)
	assert.Equal(t, stats1.MostUsedFramework, stats2.MostUsedFramework)
	assert.Equal(t, stats1.FrameworkDistribution, stats2.FrameworkDistribution)

	t.Logf("Cache miss: %v, Cache hit: %v (%.2fx faster)", 
		firstCallDuration, secondCallDuration, 
		float64(firstCallDuration)/float64(secondCallDuration))
}

func TestFileService_ConcurrentFileOperations(t *testing.T) {
	tempDir := t.TempDir()
	fileService := NewFileService(tempDir, 1*time.Hour)

	const numGoroutines = 10
	const operationsPerGoroutine = 20

	t.Run("concurrent file operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*operationsPerGoroutine)

		start := time.Now()

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				for j := 0; j < operationsPerGoroutine; j++ {
					projectID := uint(goroutineID*operationsPerGoroutine + j)
					filename := fmt.Sprintf("test-file-%d-%d.zip", goroutineID, j)
					
					// Generate file path
					filePath := fileService.GetFilePath(projectID, filename)
					
					// Validate file path
					err := fileService.ValidateFilePath(filePath)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d, op %d: validate failed: %w", goroutineID, j, err)
						continue
					}

					// Check file existence (should be false)
					exists := fileService.FileExists(filePath)
					if exists {
						errors <- fmt.Errorf("goroutine %d, op %d: file should not exist", goroutineID, j)
						continue
					}

					// Generate secure file path
					securePath := fileService.GenerateSecureFilePath(uint(goroutineID), projectID, filename)
					err = fileService.ValidateFilePath(securePath)
					if err != nil {
						errors <- fmt.Errorf("goroutine %d, op %d: secure path validate failed: %w", goroutineID, j, err)
						continue
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)
		duration := time.Since(start)

		// Check for errors
		var errorCount int
		for err := range errors {
			t.Error(err)
			errorCount++
		}

		assert.Equal(t, 0, errorCount, "Expected no errors in concurrent file operations")
		t.Logf("Completed %d concurrent file operations in %v", numGoroutines*operationsPerGoroutine, duration)
	})
}

func TestCleanupService_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cleanup performance test in short mode")
	}

	db, _ := setupPerformanceTestDB(t)
	user := createTestUserForPerformance(t, db, "cleanup@example.com")

	// Create projects with expired status
	const expiredCount = 1000
	projects := make([]models.ProjectHistory, expiredCount)
	for i := 0; i < expiredCount; i++ {
		featuresJSON, _ := json.Marshal([]string{"auth"})
		zipPath := fmt.Sprintf("/expired/project_%d.zip", i)
		zipSize := int64(1024)

		projects[i] = models.ProjectHistory{
			UserID:           user.ID,
			ProjectName:      fmt.Sprintf("expired-project-%d", i),
			Framework:        "gin",
			Features:         featuresJSON,
			AdjustedFeatures: featuresJSON,
			ZipFilePath:      &zipPath,
			ZipFileSize:      &zipSize,
			ZipFileStatus:    models.ZipFileStatusExpired,
			CreatedAt:        time.Now().Add(-48 * time.Hour), // Old enough for cleanup
			UpdatedAt:        time.Now().Add(-48 * time.Hour),
		}
	}

	err := db.CreateInBatches(projects, 100).Error
	require.NoError(t, err)

	// Create cleanup service
	tempDir := t.TempDir()
	fileService := NewFileService(tempDir, 24*time.Hour)
	cleanupService := NewFileCleanupService(db, fileService, CleanupConfig{
		RetentionPeriod: 24 * time.Hour,
		CleanupInterval: time.Hour,
		BatchSize:       100,
	})

	// Test cleanup performance
	start := time.Now()
	ctx := context.Background()
	stats, err := cleanupService.RunCleanup(ctx)
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Less(t, duration, 5*time.Second, "Cleanup took too long: %v", duration)
	t.Logf("Cleanup completed in %v, processed %d files", duration, stats.FilesProcessed)
}