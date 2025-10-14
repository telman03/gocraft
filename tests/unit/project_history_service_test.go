package unit

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telman03/gocraft-backend/internal/models"
	"github.com/telman03/gocraft-backend/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDBForService(t *testing.T) (*gorm.DB, *services.ProjectHistoryService) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.ProjectHistory{})
	require.NoError(t, err)

	service := services.NewProjectHistoryService(db)
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
	db, service := setupTestDBForService(t)
	user := createTestUserForService(t, db)

	stats, err := service.GetProjectStats(user.ID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats.TotalProjects)
	assert.Empty(t, stats.MostUsedFramework)
	assert.Empty(t, stats.MostUsedFeatures)
	assert.Empty(t, stats.FrameworkDistribution)
	assert.NotEmpty(t, stats.RecentActivity) // Should still have activity array with zeros
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