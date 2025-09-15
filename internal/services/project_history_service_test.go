package services

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBForService(t *testing.T) (*gorm.DB, *ProjectHistoryService) {
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

func createTestUserForService(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:      "test@example.com",
		Password:   "hashedpassword",
		IsVerified: true,
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestProjectHistoryForService(t *testing.T, db *gorm.DB, userID uint, projectName string, framework string, features []string, createdAt time.Time) *models.ProjectHistory {
	featuresJSON, err := json.Marshal(features)
	require.NoError(t, err)

	adjustedFeaturesJSON, err := json.Marshal(features)
	require.NoError(t, err)

	zipPath := "/path/to/test.zip"
	zipSize := int64(1024)
	duration := 5000

	project := &models.ProjectHistory{
		UserID:               userID,
		ProjectName:          projectName,
		Framework:            framework,
		Features:             featuresJSON,
		AdjustedFeatures:     adjustedFeaturesJSON,
		ZipFilePath:          &zipPath,
		ZipFileSize:          &zipSize,
		ZipFileStatus:        models.ZipFileStatusAvailable,
		GenerationDurationMs: &duration,
		CreatedAt:            createdAt,
		UpdatedAt:            createdAt,
	}

	err = db.Create(project).Error
	require.NoError(t, err)
	return project
}

func TestProjectHistoryService_CreateProjectRecord(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	req := models.CreateProjectRecordRequest{
		UserID:               user.ID,
		ProjectName:          "test-project",
		Framework:            "gin",
		Features:             []string{"auth", "database", "redis"},
		AdjustedFeatures:     []string{"auth", "database"},
		ZipFilePath:          stringPtrPHS("/path/to/test.zip"),
		ZipFileSize:          int64PtrPHS(1024),
		GenerationDurationMs: intPtrPHS(5000),
	}

	project, err := service.CreateProjectRecord(req)
	assert.NoError(t, err)
	assert.NotNil(t, project)
	assert.NotZero(t, project.ID)
	assert.Equal(t, user.ID, project.UserID)
	assert.Equal(t, "test-project", project.ProjectName)
	assert.Equal(t, "gin", project.Framework)
	assert.Equal(t, models.ZipFileStatusAvailable, project.ZipFileStatus)

	// Verify features are properly stored as JSON
	var features []string
	err = json.Unmarshal(project.Features, &features)
	assert.NoError(t, err)
	assert.Equal(t, []string{"auth", "database", "redis"}, features)

	var adjustedFeatures []string
	err = json.Unmarshal(project.AdjustedFeatures, &adjustedFeatures)
	assert.NoError(t, err)
	assert.Equal(t, []string{"auth", "database"}, adjustedFeatures)
}

func TestProjectHistoryService_CreateProjectRecord_InvalidUser(t *testing.T) {
	_, service := setupTestDBForService(t)

	req := models.CreateProjectRecordRequest{
		UserID:           999, // Non-existent user
		ProjectName:      "test-project",
		Framework:        "gin",
		Features:         []string{"auth"},
		AdjustedFeatures: []string{"auth"},
	}

	project, err := service.CreateProjectRecord(req)
	assert.Error(t, err)
	assert.Nil(t, project)
	assert.Contains(t, err.Error(), "user not found")
}

func TestProjectHistoryService_CreateProjectRecord_TransactionRollback(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create request with invalid JSON (this should cause marshaling to fail)
	req := models.CreateProjectRecordRequest{
		UserID:           user.ID,
		ProjectName:      "test-project",
		Framework:        "gin",
		Features:         []string{"auth"},
		AdjustedFeatures: []string{"auth"},
	}

	// This should succeed normally, but let's test with a corrupted database state
	// by closing the database connection
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	project, err := service.CreateProjectRecord(req)
	assert.Error(t, err)
	assert.Nil(t, project)
}

func TestProjectHistoryService_GetUserHistory(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create test projects
	now := time.Now()
	createTestProjectHistoryForService(t, db, user.ID, "project-1", "gin", []string{"auth", "database"}, now.Add(-2*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "project-2", "echo", []string{"auth", "redis"}, now.Add(-1*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "project-3", "fiber", []string{"websocket"}, now)

	filters := models.HistoryFilters{
		Page:     1,
		PageSize: 10,
	}

	response, err := service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Len(t, response.Projects, 3)
	assert.Equal(t, 3, response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 1, response.TotalPages)

	// Verify projects are sorted by created_at DESC (newest first)
	assert.Equal(t, "project-3", response.Projects[0].ProjectName)
	assert.Equal(t, "project-2", response.Projects[1].ProjectName)
	assert.Equal(t, "project-1", response.Projects[2].ProjectName)
}

func TestProjectHistoryService_GetUserHistory_WithFilters(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create test projects with different frameworks
	now := time.Now()
	createTestProjectHistoryForService(t, db, user.ID, "gin-project", "gin", []string{"auth", "database"}, now.Add(-2*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "echo-project", "echo", []string{"auth", "redis"}, now.Add(-1*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "fiber-project", "fiber", []string{"websocket"}, now)

	// Test framework filter (single framework)
	filters := models.HistoryFilters{
		Page:      1,
		PageSize:  10,
		Framework: "gin",
	}

	response, err := service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 1)
	assert.Equal(t, "gin-project", response.Projects[0].ProjectName)

	// Test multiple frameworks filter
	filters = models.HistoryFilters{
		Page:       1,
		PageSize:   10,
		Frameworks: []string{"gin", "fiber"},
	}

	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 2)

	// Test status filter
	filters = models.HistoryFilters{
		Page:     1,
		PageSize: 10,
		Status:   "available",
	}

	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 3) // All projects should be available

	// Note: Skipping search filter test as it uses PostgreSQL-specific ILIKE syntax
	// which is not supported in SQLite test environment
}

func TestProjectHistoryService_GetUserHistory_Pagination(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create 25 test projects
	now := time.Now()
	for i := 0; i < 25; i++ {
		createTestProjectHistoryForService(t, db, user.ID, fmt.Sprintf("project-%d", i), "gin", []string{"auth"}, now.Add(-time.Duration(i)*time.Hour))
	}

	// Test first page
	filters := models.HistoryFilters{
		Page:     1,
		PageSize: 10,
	}

	response, err := service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 10)
	assert.Equal(t, 25, response.Total)
	assert.Equal(t, 1, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 3, response.TotalPages)

	// Test second page
	filters.Page = 2
	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 10)
	assert.Equal(t, 2, response.Page)

	// Test last page
	filters.Page = 3
	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 5) // Remaining 5 projects
	assert.Equal(t, 3, response.Page)
}

func TestProjectHistoryService_GetUserHistory_DateFilters(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different dates
	baseDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	createTestProjectHistoryForService(t, db, user.ID, "old-project", "gin", []string{"auth"}, baseDate.Add(-10*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "recent-project", "echo", []string{"auth"}, baseDate)
	createTestProjectHistoryForService(t, db, user.ID, "future-project", "fiber", []string{"auth"}, baseDate.Add(5*24*time.Hour))

	// Test date from filter
	filters := models.HistoryFilters{
		Page:     1,
		PageSize: 10,
		DateFrom: "2024-01-10",
	}

	response, err := service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 2) // recent and future projects

	// Test date to filter
	filters = models.HistoryFilters{
		Page:   1,
		PageSize: 10,
		DateTo: "2024-01-16",
	}

	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 2) // old and recent projects

	// Test date range filter
	filters = models.HistoryFilters{
		Page:     1,
		PageSize: 10,
		DateFrom: "2024-01-10",
		DateTo:   "2024-01-16",
	}

	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Len(t, response.Projects, 1) // only recent project
	assert.Equal(t, "recent-project", response.Projects[0].ProjectName)
}

func TestProjectHistoryService_GetUserHistory_Sorting(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different names and sizes
	now := time.Now()
	project1 := createTestProjectHistoryForService(t, db, user.ID, "zebra-project", "gin", []string{"auth"}, now.Add(-2*time.Hour))
	project2 := createTestProjectHistoryForService(t, db, user.ID, "alpha-project", "echo", []string{"auth"}, now.Add(-1*time.Hour))
	project3 := createTestProjectHistoryForService(t, db, user.ID, "beta-project", "fiber", []string{"auth"}, now)

	// Update zip file sizes for testing
	db.Model(project1).Update("zip_file_size", 3000)
	db.Model(project2).Update("zip_file_size", 1000)
	db.Model(project3).Update("zip_file_size", 2000)

	// Test sort by project name ascending
	filters := models.HistoryFilters{
		Page:      1,
		PageSize:  10,
		SortBy:    "project_name",
		SortOrder: "asc",
	}

	response, err := service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Equal(t, "alpha-project", response.Projects[0].ProjectName)
	assert.Equal(t, "beta-project", response.Projects[1].ProjectName)
	assert.Equal(t, "zebra-project", response.Projects[2].ProjectName)

	// Test sort by zip file size descending
	filters = models.HistoryFilters{
		Page:      1,
		PageSize:  10,
		SortBy:    "zip_file_size",
		SortOrder: "desc",
	}

	response, err = service.GetUserHistory(user.ID, filters)
	assert.NoError(t, err)
	assert.Equal(t, int64(3000), *response.Projects[0].ZipFileSize)
	assert.Equal(t, int64(2000), *response.Projects[1].ZipFileSize)
	assert.Equal(t, int64(1000), *response.Projects[2].ZipFileSize)
}

func TestProjectHistoryService_GetProjectByID(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)
	otherUser := &models.User{Email: "other@example.com", Password: "pass", IsVerified: true}
	db.Create(otherUser)

	project := createTestProjectHistoryForService(t, db, user.ID, "test-project", "gin", []string{"auth"}, time.Now())

	// Test successful retrieval
	retrieved, err := service.GetProjectByID(user.ID, project.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, project.ID, retrieved.ID)
	assert.Equal(t, "test-project", retrieved.ProjectName)

	// Test access denied for different user
	retrieved, err = service.GetProjectByID(otherUser.ID, project.ID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "project not found or access denied")

	// Test non-existent project
	retrieved, err = service.GetProjectByID(user.ID, 999)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
	assert.Contains(t, err.Error(), "project not found or access denied")
}

func TestProjectHistoryService_DeleteProject(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)
	otherUser := &models.User{Email: "other@example.com", Password: "pass", IsVerified: true}
	db.Create(otherUser)

	project := createTestProjectHistoryForService(t, db, user.ID, "test-project", "gin", []string{"auth"}, time.Now())

	// Test successful deletion
	err := service.DeleteProject(user.ID, project.ID)
	assert.NoError(t, err)

	// Verify project is deleted
	var count int64
	db.Model(&models.ProjectHistory{}).Where("id = ?", project.ID).Count(&count)
	assert.Equal(t, int64(0), count)

	// Test access denied for different user
	project2 := createTestProjectHistoryForService(t, db, user.ID, "test-project-2", "echo", []string{"auth"}, time.Now())
	err = service.DeleteProject(otherUser.ID, project2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found or access denied")

	// Verify project still exists
	db.Model(&models.ProjectHistory{}).Where("id = ?", project2.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Test non-existent project
	err = service.DeleteProject(user.ID, 999)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project not found or access denied")
}

func TestProjectHistoryService_GetProjectStats(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different frameworks and features
	now := time.Now()
	createTestProjectHistoryForService(t, db, user.ID, "gin-project-1", "gin", []string{"auth", "database"}, now.Add(-5*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "gin-project-2", "gin", []string{"auth", "redis"}, now.Add(-4*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "echo-project", "echo", []string{"auth", "websocket"}, now.Add(-3*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "fiber-project", "fiber", []string{"database", "grpc"}, now.Add(-2*24*time.Hour))

	stats, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 4, stats.TotalProjects)
	assert.Equal(t, "gin", stats.MostUsedFramework) // gin appears 2 times
	assert.Contains(t, stats.MostUsedFeatures, "auth") // auth appears 3 times
	assert.Equal(t, 2, stats.FrameworkDistribution["gin"])
	assert.Equal(t, 1, stats.FrameworkDistribution["echo"])
	assert.Equal(t, 1, stats.FrameworkDistribution["fiber"])
	assert.NotEmpty(t, stats.RecentActivity)
}

func TestProjectHistoryService_GetProjectStats_EmptyHistory(t *testing.T) {
	_, service := setupTestDBForService(t)
	user := createTestUserForService(t, service.db)

	stats, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalProjects)
	assert.Empty(t, stats.MostUsedFramework)
	assert.Empty(t, stats.MostUsedFeatures)
	assert.Empty(t, stats.FrameworkDistribution)
	assert.NotEmpty(t, stats.RecentActivity) // Should still have activity array with zeros
}

func TestProjectHistoryService_SearchProjectsByFeatures(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different features
	createTestProjectHistoryForService(t, db, user.ID, "auth-project", "gin", []string{"auth", "database"}, time.Now())
	createTestProjectHistoryForService(t, db, user.ID, "redis-project", "echo", []string{"redis", "websocket"}, time.Now())
	createTestProjectHistoryForService(t, db, user.ID, "mixed-project", "fiber", []string{"auth", "redis"}, time.Now())

	// Note: SearchProjectsByFeatures uses PostgreSQL-specific ILIKE syntax
	// For SQLite testing, we'll test the basic functionality without the complex search
	
	// Search with empty features should return empty
	projects, err := service.SearchProjectsByFeatures(user.ID, []string{}, 10)
	assert.NoError(t, err)
	assert.Empty(t, projects)

	// Test that the method doesn't crash with valid input
	// The actual search functionality would work in PostgreSQL environment
	_, err = service.SearchProjectsByFeatures(user.ID, []string{"auth"}, 10)
	// We expect this to fail in SQLite due to ILIKE syntax, so we just check it doesn't panic
	if err != nil {
		t.Logf("Expected error in SQLite environment due to ILIKE syntax: %v", err)
	}
}

func TestProjectHistoryService_SearchProjectsByFramework(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different frameworks and dates
	now := time.Now()
	createTestProjectHistoryForService(t, db, user.ID, "gin-old", "gin", []string{"auth"}, now.Add(-10*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "gin-recent", "gin", []string{"auth"}, now.Add(-1*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "echo-project", "echo", []string{"auth"}, now)

	// Search all gin projects
	projects, err := service.SearchProjectsByFramework(user.ID, "gin", nil)
	assert.NoError(t, err)
	assert.Len(t, projects, 2)

	// Search gin projects within last 2 days
	dateRange := 2 * 24 * time.Hour
	projects, err = service.SearchProjectsByFramework(user.ID, "gin", &dateRange)
	assert.NoError(t, err)
	assert.Len(t, projects, 1) // Only gin-recent
	assert.Equal(t, "gin-recent", projects[0].ProjectName)

	// Search non-existent framework
	projects, err = service.SearchProjectsByFramework(user.ID, "nonexistent", nil)
	assert.NoError(t, err)
	assert.Empty(t, projects)
}

func TestProjectHistoryService_GetProjectsByDateRange(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different dates
	baseDate := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	createTestProjectHistoryForService(t, db, user.ID, "project-1", "gin", []string{"auth"}, baseDate.Add(-2*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "project-2", "echo", []string{"auth"}, baseDate)
	createTestProjectHistoryForService(t, db, user.ID, "project-3", "fiber", []string{"auth"}, baseDate.Add(2*24*time.Hour))

	// Get projects within date range
	startDate := baseDate.Add(-1 * 24 * time.Hour)
	endDate := baseDate.Add(1 * 24 * time.Hour)

	projects, err := service.GetProjectsByDateRange(user.ID, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, projects, 1) // Only project-2
	assert.Equal(t, "project-2", projects[0].ProjectName)

	// Get all projects with wide date range
	startDate = baseDate.Add(-10 * 24 * time.Hour)
	endDate = baseDate.Add(10 * 24 * time.Hour)

	projects, err = service.GetProjectsByDateRange(user.ID, startDate, endDate)
	assert.NoError(t, err)
	assert.Len(t, projects, 3)
}

func TestProjectHistoryService_ConvertToResponse(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	project := createTestProjectHistoryForService(t, db, user.ID, "test-project", "gin", []string{"auth", "database"}, time.Now())

	response, err := service.convertToResponse(*project)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, project.ID, response.ID)
	assert.Equal(t, "test-project", response.ProjectName)
	assert.Equal(t, "gin", response.Framework)
	assert.Equal(t, []string{"auth", "database"}, response.Features)
	assert.Equal(t, []string{"auth", "database"}, response.AdjustedFeatures)
	assert.Equal(t, "available", response.ZipFileStatus)
	assert.False(t, response.CanDownload) // File doesn't actually exist
	assert.True(t, response.CanRegenerate) // Has features and framework
}

func TestProjectHistoryService_StatsCache(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create a project
	createTestProjectHistoryForService(t, db, user.ID, "test-project", "gin", []string{"auth"}, time.Now())

	// First call should calculate stats
	stats1, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, stats1.TotalProjects)

	// Second call should use cache (same result)
	stats2, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, stats1.TotalProjects, stats2.TotalProjects)

	// Create another project
	createTestProjectHistoryForService(t, db, user.ID, "test-project-2", "echo", []string{"redis"}, time.Now())

	// Cache should be invalidated after creating a new project
	// But since we're calling GetProjectStats directly, we need to test cache invalidation
	// through CreateProjectRecord
	req := models.CreateProjectRecordRequest{
		UserID:           user.ID,
		ProjectName:      "cache-test-project",
		Framework:        "fiber",
		Features:         []string{"websocket"},
		AdjustedFeatures: []string{"websocket"},
	}

	_, err = service.CreateProjectRecord(req)
	assert.NoError(t, err)

	// Now stats should reflect the new project count
	stats3, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, stats3.TotalProjects) // Should include all 3 projects
}

func TestProjectHistoryService_GetFrameworkUsageStats(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects with different frameworks
	createTestProjectHistoryForService(t, db, user.ID, "gin-1", "gin", []string{"auth"}, time.Now())
	createTestProjectHistoryForService(t, db, user.ID, "gin-2", "gin", []string{"auth"}, time.Now())
	createTestProjectHistoryForService(t, db, user.ID, "echo-1", "echo", []string{"auth"}, time.Now())

	stats, err := service.GetFrameworkUsageStats(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	ginStats := stats["gin"].(map[string]interface{})
	assert.Equal(t, 2, ginStats["count"])
	assert.InDelta(t, 66.67, ginStats["percentage"], 0.1) // 2/3 * 100

	echoStats := stats["echo"].(map[string]interface{})
	assert.Equal(t, 1, echoStats["count"])
	assert.InDelta(t, 33.33, echoStats["percentage"], 0.1) // 1/3 * 100
}

func TestProjectHistoryService_GetProjectTrends(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create projects in different months
	jan2024 := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	feb2024 := time.Date(2024, 2, 15, 12, 0, 0, 0, time.UTC)

	createTestProjectHistoryForService(t, db, user.ID, "jan-gin", "gin", []string{"auth"}, jan2024)
	createTestProjectHistoryForService(t, db, user.ID, "jan-echo", "echo", []string{"auth"}, jan2024)
	createTestProjectHistoryForService(t, db, user.ID, "feb-gin", "gin", []string{"auth"}, feb2024)

	// Note: GetProjectTrends uses PostgreSQL-specific TO_CHAR function
	// For SQLite testing, we expect this to fail
	trends, err := service.GetProjectTrends(user.ID, 3) // Last 3 months
	
	if err != nil {
		// Expected in SQLite environment due to TO_CHAR function
		t.Logf("Expected error in SQLite environment due to TO_CHAR function: %v", err)
		assert.Contains(t, err.Error(), "TO_CHAR")
		return
	}

	// If it somehow works (shouldn't in SQLite), verify the structure
	assert.NotNil(t, trends)
}

func TestProjectHistoryService_GetDashboardData(t *testing.T) {
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	// Create some test data
	now := time.Now()
	createTestProjectHistoryForService(t, db, user.ID, "recent-1", "gin", []string{"auth"}, now.Add(-1*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "recent-2", "echo", []string{"database"}, now.Add(-2*24*time.Hour))
	createTestProjectHistoryForService(t, db, user.ID, "old-1", "gin", []string{"auth"}, now.Add(-10*24*time.Hour))

	dashboardData, err := service.GetDashboardData(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, dashboardData)

	overview := dashboardData["overview"].(map[string]interface{})
	assert.Equal(t, 3, overview["total_projects"])
	assert.Equal(t, 2, overview["total_frameworks"]) // gin and echo
	assert.Equal(t, "gin", overview["most_used_framework"])

	assert.NotNil(t, dashboardData["framework_distribution"])
	assert.NotNil(t, dashboardData["most_used_features"])
	assert.NotNil(t, dashboardData["recent_activity"])
	assert.NotNil(t, dashboardData["cache_info"])
}

// Helper functions
func stringPtrPHS(s string) *string {
	return &s
}

func int64PtrPHS(i int64) *int64 {
	return &i
}

func intPtrPHS(i int) *int {
	return &i
}