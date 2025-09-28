package unit

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/handlers"
	"github.com/telman03/ai-backend-generator/internal/middleware"
	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDB *gorm.DB
var testApp *fiber.App

func setupTestEnvironment(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Auto-migrate the schema
	err = db.AutoMigrate(&models.User{}, &models.ProjectHistory{})
	require.NoError(t, err)

	// Set global database for handlers
	database.DB = db
	testDB = db

	// Create Fiber app with middleware
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Add authentication middleware
	app.Use(func(c *fiber.Ctx) error {
		// Skip auth for test endpoints
		if c.Path() == "/test/login" {
			return c.Next()
		}

		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authorization header required",
			})
		}

		// Extract token (remove "Bearer " prefix)
		tokenString := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}

		// Parse JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("test-secret"), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract user ID from claims
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if userID, ok := claims["user_id"].(float64); ok {
				// Set both user_id (for authorization middleware) and validated_user_id (for handlers)
				c.Locals("user_id", userID) // float64 for authorization middleware
				c.Locals("validated_user_id", uint(userID)) // uint for handlers
				return c.Next()
			}
		}

		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	})

	// Register routes
	api := app.Group("/api")
	history := api.Group("/history")

	// Routes without project ID parameter
	history.Get("/", handlers.GetProjectHistory)
	history.Post("/duplicate", handlers.DuplicateProject)
	history.Get("/stats", handlers.GetProjectStats)
	history.Get("/dashboard", handlers.GetDashboardData)
	
	// Routes with project ID parameter (need ownership validation)
	history.Get("/:id", middleware.ProjectOwnershipValidator(), handlers.GetProjectDetails)
	history.Delete("/:id", middleware.ProjectOwnershipValidator(), handlers.DeleteProject)
	history.Get("/:id/download", middleware.ProjectOwnershipValidator(), handlers.DownloadProject)
	history.Post("/:id/regenerate", middleware.ProjectOwnershipValidator(), handlers.RegenerateProject)

	// Test endpoint for creating JWT tokens
	app.Post("/test/login", func(c *fiber.Ctx) error {
		var req struct {
			UserID uint `json:"user_id"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": req.UserID,
			"exp":     time.Now().Add(time.Hour).Unix(),
		})

		tokenString, err := token.SignedString([]byte("test-secret"))
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create token"})
		}

		return c.JSON(fiber.Map{"token": tokenString})
	})

	testApp = app
}

func createTestUserForHandler(t *testing.T) *models.User {
	user := &models.User{
		Email:      fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		Password:   "hashedpassword",
		IsVerified: true,
	}
	err := testDB.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestProjectForHandler(t *testing.T, userID uint, projectName string, framework string, features []string) *models.ProjectHistory {
	featuresJSON, err := json.Marshal(features)
	require.NoError(t, err)

	adjustedFeaturesJSON, err := json.Marshal(features)
	require.NoError(t, err)

	zipPath := fmt.Sprintf("./test_files/project_%d.zip", time.Now().UnixNano())
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
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = testDB.Create(project).Error
	require.NoError(t, err)
	return project
}

func getAuthToken(t *testing.T, userID uint) string {
	reqBody := map[string]uint{"user_id": userID}
	jsonBody, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/test/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := testApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var tokenResp struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	require.NoError(t, err)

	return tokenResp.Token
}

func TestGetProjectHistory(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Create test projects
	createTestProjectForHandler(t, user.ID, "project-1", "gin", []string{"auth", "database"})
	createTestProjectForHandler(t, user.ID, "project-2", "echo", []string{"auth", "redis"})
	createTestProjectForHandler(t, user.ID, "project-3", "fiber", []string{"websocket"})

	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "get all projects",
			queryParams:    "",
			expectedStatus: 200,
			expectedCount:  3,
		},
		{
			name:           "paginated request",
			queryParams:    "?page=1&page_size=2",
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "filter by framework",
			queryParams:    "?framework=gin",
			expectedStatus: 200,
			expectedCount:  1,
		},
		{
			name:           "filter by multiple frameworks",
			queryParams:    "?frameworks=gin,echo",
			expectedStatus: 200,
			expectedCount:  2,
		},
		{
			name:           "filter by status",
			queryParams:    "?status=available",
			expectedStatus: 200,
			expectedCount:  3,
		},
		{
			name:           "sort by project name",
			queryParams:    "?sort_by=project_name&sort_order=asc",
			expectedStatus: 200,
			expectedCount:  3,
		},
		{
			name:           "invalid page parameter",
			queryParams:    "?page=invalid",
			expectedStatus: 400,
			expectedCount:  0,
		},
		{
			name:           "invalid page size parameter",
			queryParams:    "?page_size=invalid",
			expectedStatus: 400,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/history"+tt.queryParams, nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var response models.ProjectHistoryListResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Len(t, response.Projects, tt.expectedCount)
				assert.Equal(t, tt.expectedCount, len(response.Projects))

				// Verify response structure
				if len(response.Projects) > 0 {
					project := response.Projects[0]
					assert.NotZero(t, project.ID)
					assert.NotEmpty(t, project.ProjectName)
					assert.NotEmpty(t, project.Framework)
					assert.NotEmpty(t, project.Features)
					assert.NotZero(t, project.CreatedAt)
				}
			}
		})
	}
}

func TestGetProjectHistory_Unauthorized(t *testing.T) {
	setupTestEnvironment(t)

	req := httptest.NewRequest("GET", "/api/history", nil)
	// No Authorization header

	resp, err := testApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 401, resp.StatusCode)
}

func TestGetProjectDetails(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	otherUser := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)
	otherToken := getAuthToken(t, otherUser.ID)

	project := createTestProjectForHandler(t, user.ID, "test-project", "gin", []string{"auth", "database"})

	tests := []struct {
		name           string
		projectID      string
		token          string
		expectedStatus int
	}{
		{
			name:           "get own project details",
			projectID:      strconv.Itoa(int(project.ID)),
			token:          token,
			expectedStatus: 200,
		},
		{
			name:           "get other user's project",
			projectID:      strconv.Itoa(int(project.ID)),
			token:          otherToken,
			expectedStatus: 404, // Should not find due to ownership validation
		},
		{
			name:           "invalid project ID",
			projectID:      "invalid",
			token:          token,
			expectedStatus: 400,
		},
		{
			name:           "non-existent project",
			projectID:      "999999",
			token:          token,
			expectedStatus: 404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/history/"+tt.projectID, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, float64(project.ID), response["id"])
				assert.Equal(t, project.ProjectName, response["project_name"])
				assert.Equal(t, project.Framework, response["framework"])
			}
		})
	}
}

func TestDeleteProject(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	otherUser := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)
	otherToken := getAuthToken(t, otherUser.ID)

	project := createTestProjectForHandler(t, user.ID, "test-project", "gin", []string{"auth", "database"})
	otherProject := createTestProjectForHandler(t, user.ID, "other-project", "echo", []string{"auth"})

	tests := []struct {
		name           string
		projectID      string
		token          string
		expectedStatus int
	}{
		{
			name:           "delete own project",
			projectID:      strconv.Itoa(int(project.ID)),
			token:          token,
			expectedStatus: 200,
		},
		{
			name:           "delete other user's project",
			projectID:      strconv.Itoa(int(otherProject.ID)),
			token:          otherToken,
			expectedStatus: 404, // Should fail due to ownership validation
		},
		{
			name:           "invalid project ID",
			projectID:      "invalid",
			token:          token,
			expectedStatus: 400,
		},
		{
			name:           "non-existent project",
			projectID:      "999999",
			token:          token,
			expectedStatus: 404, // Middleware will return 404 for non-existent project
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/history/"+tt.projectID, nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "Project deleted successfully", response["message"])
				assert.NotNil(t, response["project_id"])
				assert.NotNil(t, response["timestamp"])

				// Verify project is actually deleted
				var count int64
				testDB.Model(&models.ProjectHistory{}).Where("id = ?", project.ID).Count(&count)
				assert.Equal(t, int64(0), count)
			}
		})
	}
}

func TestDownloadProject(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Create test file in the output directory (which is what the FileService expects)
	testDir := "./output"
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	testFilePath := filepath.Join(testDir, "test_project.zip")
	testContent := []byte("test zip content")
	err := os.WriteFile(testFilePath, testContent, 0644)
	require.NoError(t, err)

	// Create project with existing file
	project := createTestProjectForHandler(t, user.ID, "test-project", "gin", []string{"auth"})
	project.ZipFilePath = &testFilePath
	testDB.Save(project)

	// Create project without file
	projectNoFile := createTestProjectForHandler(t, user.ID, "no-file-project", "echo", []string{"auth"})
	nonExistentPath := "./output/non_existent_file.zip"
	projectNoFile.ZipFilePath = &nonExistentPath
	testDB.Save(projectNoFile)

	tests := []struct {
		name           string
		projectID      string
		expectedStatus int
		checkContent   bool
	}{
		{
			name:           "download existing file",
			projectID:      strconv.Itoa(int(project.ID)),
			expectedStatus: 200,
			checkContent:   true,
		},
		{
			name:           "download non-existent file",
			projectID:      strconv.Itoa(int(projectNoFile.ID)),
			expectedStatus: 404,
			checkContent:   false,
		},
		{
			name:           "invalid project ID",
			projectID:      "invalid",
			expectedStatus: 400,
			checkContent:   false,
		},
		{
			name:           "non-existent project",
			projectID:      "999999",
			expectedStatus: 404,
			checkContent:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/history/"+tt.projectID+"/download", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.checkContent && tt.expectedStatus == 200 {
				// Check content type
				assert.Equal(t, "application/zip", resp.Header.Get("Content-Type"))
				
				// Check content disposition
				contentDisposition := resp.Header.Get("Content-Disposition")
				assert.Contains(t, contentDisposition, "attachment")
				assert.Contains(t, contentDisposition, "filename=")

				// Check file content
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				assert.Equal(t, testContent, body)
			}
		})
	}
}

func TestDuplicateProject(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	otherUser := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)
	otherToken := getAuthToken(t, otherUser.ID)

	project := createTestProjectForHandler(t, user.ID, "original-project", "gin", []string{"auth", "database"})

	tests := []struct {
		name           string
		requestBody    models.DuplicateProjectRequest
		token          string
		expectedStatus int
	}{
		{
			name: "duplicate own project",
			requestBody: models.DuplicateProjectRequest{
				OriginalProjectID: project.ID,
				NewProjectName:    "duplicated-project",
			},
			token:          token,
			expectedStatus: 200,
		},
		{
			name: "duplicate other user's project",
			requestBody: models.DuplicateProjectRequest{
				OriginalProjectID: project.ID,
				NewProjectName:    "duplicated-project",
			},
			token:          otherToken,
			expectedStatus: 404, // Should fail due to ownership validation
		},
		{
			name: "invalid project ID",
			requestBody: models.DuplicateProjectRequest{
				OriginalProjectID: 999999,
				NewProjectName:    "duplicated-project",
			},
			token:          token,
			expectedStatus: 404,
		},
		{
			name: "empty project name",
			requestBody: models.DuplicateProjectRequest{
				OriginalProjectID: project.ID,
				NewProjectName:    "",
			},
			token:          token,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/history/duplicate", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+tt.token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == 200 {
				var response map[string]interface{}
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.True(t, response["success"].(bool))
				assert.NotNil(t, response["duplicate_config"])
				assert.NotNil(t, response["form_data"])

				duplicateConfig := response["duplicate_config"].(map[string]interface{})
				assert.Equal(t, "gin", duplicateConfig["framework"])
				assert.NotEmpty(t, duplicateConfig["features"])
			}
		})
	}
}

func TestGetProjectStats(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Create test projects with different frameworks and features
	createTestProjectForHandler(t, user.ID, "gin-project-1", "gin", []string{"auth", "database"})
	createTestProjectForHandler(t, user.ID, "gin-project-2", "gin", []string{"auth", "redis"})
	createTestProjectForHandler(t, user.ID, "echo-project", "echo", []string{"auth", "websocket"})

	req := httptest.NewRequest("GET", "/api/history/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var stats models.ProjectStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&stats)
	require.NoError(t, err)

	assert.Equal(t, 3, stats.TotalProjects)
	assert.Equal(t, "gin", stats.MostUsedFramework) // gin appears 2 times
	assert.Contains(t, stats.MostUsedFeatures, "auth") // auth appears in all projects
	assert.Equal(t, 2, stats.FrameworkDistribution["gin"])
	assert.Equal(t, 1, stats.FrameworkDistribution["echo"])
	assert.NotEmpty(t, stats.RecentActivity)
}

func TestGetProjectStats_EmptyHistory(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	req := httptest.NewRequest("GET", "/api/history/stats", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var stats models.ProjectStatsResponse
	err = json.NewDecoder(resp.Body).Decode(&stats)
	require.NoError(t, err)

	assert.Equal(t, 0, stats.TotalProjects)
	assert.Empty(t, stats.MostUsedFramework)
	assert.Empty(t, stats.MostUsedFeatures)
	assert.Empty(t, stats.FrameworkDistribution)
	assert.NotEmpty(t, stats.RecentActivity) // Should still have activity array with zeros
}

func TestGetDashboardData(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Create some test data
	createTestProjectForHandler(t, user.ID, "recent-1", "gin", []string{"auth"})
	createTestProjectForHandler(t, user.ID, "recent-2", "echo", []string{"database"})

	req := httptest.NewRequest("GET", "/api/history/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := testApp.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var dashboardData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&dashboardData)
	require.NoError(t, err)

	assert.NotNil(t, dashboardData["overview"])
	assert.NotNil(t, dashboardData["framework_distribution"])
	assert.NotNil(t, dashboardData["most_used_features"])
	assert.NotNil(t, dashboardData["recent_activity"])
	assert.NotNil(t, dashboardData["cache_info"])

	overview := dashboardData["overview"].(map[string]interface{})
	assert.Equal(t, float64(2), overview["total_projects"])
	assert.Equal(t, float64(2), overview["total_frameworks"]) // gin and echo
}

func TestAuthenticationMiddleware(t *testing.T) {
	setupTestEnvironment(t)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "no auth header",
			authHeader:     "",
			expectedStatus: 401,
		},
		{
			name:           "invalid token",
			authHeader:     "Bearer invalid-token",
			expectedStatus: 401,
		},
		{
			name:           "malformed token",
			authHeader:     "Bearer",
			expectedStatus: 401,
		},
		{
			name:           "expired token",
			authHeader:     createExpiredToken(t),
			expectedStatus: 401,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/history", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestErrorHandling(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Test various error scenarios
	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
	}{
		{
			name:           "invalid JSON in duplicate request",
			method:         "POST",
			path:           "/api/history/duplicate",
			body:           "invalid json",
			expectedStatus: 400,
		},
		{
			name:           "missing required fields in duplicate request",
			method:         "POST",
			path:           "/api/history/duplicate",
			body:           map[string]interface{}{},
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody io.Reader
			if tt.body != nil {
				if str, ok := tt.body.(string); ok {
					reqBody = bytes.NewReader([]byte(str))
				} else {
					jsonBody, err := json.Marshal(tt.body)
					require.NoError(t, err)
					reqBody = bytes.NewReader(jsonBody)
				}
			}

			req := httptest.NewRequest(tt.method, tt.path, reqBody)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := testApp.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestConcurrentRequests(t *testing.T) {
	setupTestEnvironment(t)
	user := createTestUserForHandler(t)
	token := getAuthToken(t, user.ID)

	// Create test projects
	for i := 0; i < 10; i++ {
		createTestProjectForHandler(t, user.ID, fmt.Sprintf("project-%d", i), "gin", []string{"auth"})
	}

	// Test concurrent requests
	const numRequests = 10
	results := make(chan int, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/history", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			resp, err := testApp.Test(req)
			if err != nil {
				results <- 500
				return
			}
			defer resp.Body.Close()

			results <- resp.StatusCode
		}()
	}

	// Collect results
	successCount := 0
	for i := 0; i < numRequests; i++ {
		status := <-results
		if status == 200 {
			successCount++
		}
	}

	// All requests should succeed
	assert.Equal(t, numRequests, successCount)
}

// Helper function to create an expired token
func createExpiredToken(t *testing.T) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": 1,
		"exp":     time.Now().Add(-time.Hour).Unix(), // Expired 1 hour ago
	})

	tokenString, err := token.SignedString([]byte("test-secret"))
	require.NoError(t, err)

	return "Bearer " + tokenString
}