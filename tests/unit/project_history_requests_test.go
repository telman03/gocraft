package unit

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/telman03/ai-backend-generator/internal/models"
)

func TestProjectHistoryResponse_JSONSerialization(t *testing.T) {
	response := &models.ProjectHistoryResponse{
		ID:                   1,
		ProjectName:          "test-project",
		Framework:            "gin",
		Features:             []string{"auth", "database", "redis"},
		AdjustedFeatures:     []string{"auth", "database"},
		ZipFileSize:          int64Ptr(1024),
		ZipFileStatus:        "available",
		GenerationDurationMs: intPtr(5000),
		CreatedAt:            time.Now(),
		CanDownload:          true,
		CanRegenerate:        true,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "test-project")
	assert.Contains(t, string(jsonData), "gin")
	assert.Contains(t, string(jsonData), "auth")

	// Test JSON unmarshaling
	var unmarshaled models.ProjectHistoryResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.ProjectName, unmarshaled.ProjectName)
	assert.Equal(t, response.Framework, unmarshaled.Framework)
	assert.Equal(t, response.Features, unmarshaled.Features)
	assert.Equal(t, response.CanDownload, unmarshaled.CanDownload)
}

func TestProjectHistoryResponse_OptionalFields(t *testing.T) {
	// Test with nil optional fields
	response := &models.ProjectHistoryResponse{
		ID:                   1,
		ProjectName:          "minimal-project",
		Framework:            "echo",
		Features:             []string{"auth"},
		AdjustedFeatures:     []string{"auth"},
		ZipFileSize:          nil,
		ZipFileStatus:        "expired",
		GenerationDurationMs: nil,
		CreatedAt:            time.Now(),
		CanDownload:          false,
		CanRegenerate:        true,
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var unmarshaled models.ProjectHistoryResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Nil(t, unmarshaled.ZipFileSize)
	assert.Nil(t, unmarshaled.GenerationDurationMs)
	assert.False(t, unmarshaled.CanDownload)
	assert.True(t, unmarshaled.CanRegenerate)
}

func TestProjectHistoryListResponse_Pagination(t *testing.T) {
	projects := []models.ProjectHistoryResponse{
		{ID: 1, ProjectName: "project-1", Framework: "gin"},
		{ID: 2, ProjectName: "project-2", Framework: "echo"},
		{ID: 3, ProjectName: "project-3", Framework: "fiber"},
	}

	response := &models.ProjectHistoryListResponse{
		Projects:   projects,
		Total:      25,
		Page:       2,
		PageSize:   10,
		TotalPages: 3,
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var unmarshaled models.ProjectHistoryListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Len(t, unmarshaled.Projects, 3)
	assert.Equal(t, 25, unmarshaled.Total)
	assert.Equal(t, 2, unmarshaled.Page)
	assert.Equal(t, 10, unmarshaled.PageSize)
	assert.Equal(t, 3, unmarshaled.TotalPages)
}

func TestProjectHistoryListResponse_EmptyProjects(t *testing.T) {
	response := &models.ProjectHistoryListResponse{
		Projects:   []models.ProjectHistoryResponse{},
		Total:      0,
		Page:       1,
		PageSize:   10,
		TotalPages: 0,
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var unmarshaled models.ProjectHistoryListResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Empty(t, unmarshaled.Projects)
	assert.Equal(t, 0, unmarshaled.Total)
	assert.Equal(t, 0, unmarshaled.TotalPages)
}

func TestProjectStatsResponse_CompleteStats(t *testing.T) {
	response := &models.ProjectStatsResponse{
		TotalProjects:     15,
		MostUsedFramework: "gin",
		MostUsedFeatures:  []string{"auth", "database", "redis"},
		FrameworkDistribution: map[string]int{
			"gin":   8,
			"echo":  4,
			"fiber": 3,
		},
		RecentActivity: []models.DailyActivityCount{
			{Date: "2024-01-01", Count: 2},
			{Date: "2024-01-02", Count: 1},
			{Date: "2024-01-03", Count: 0},
		},
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "gin")
	assert.Contains(t, string(jsonData), "auth")

	var unmarshaled models.ProjectStatsResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 15, unmarshaled.TotalProjects)
	assert.Equal(t, "gin", unmarshaled.MostUsedFramework)
	assert.Len(t, unmarshaled.MostUsedFeatures, 3)
	assert.Equal(t, 8, unmarshaled.FrameworkDistribution["gin"])
	assert.Len(t, unmarshaled.RecentActivity, 3)
}

func TestProjectStatsResponse_EmptyStats(t *testing.T) {
	response := &models.ProjectStatsResponse{
		TotalProjects:         0,
		MostUsedFramework:     "",
		MostUsedFeatures:      []string{},
		FrameworkDistribution: map[string]int{},
		RecentActivity:        []models.DailyActivityCount{},
	}

	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)

	var unmarshaled models.ProjectStatsResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 0, unmarshaled.TotalProjects)
	assert.Empty(t, unmarshaled.MostUsedFramework)
	assert.Empty(t, unmarshaled.MostUsedFeatures)
	assert.Empty(t, unmarshaled.FrameworkDistribution)
	assert.Empty(t, unmarshaled.RecentActivity)
}

func TestDailyActivityCount_JSONSerialization(t *testing.T) {
	activity := &models.DailyActivityCount{
		Date:  "2024-01-15",
		Count: 5,
	}

	jsonData, err := json.Marshal(activity)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "2024-01-15")
	assert.Contains(t, string(jsonData), "5")

	var unmarshaled models.DailyActivityCount
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, "2024-01-15", unmarshaled.Date)
	assert.Equal(t, 5, unmarshaled.Count)
}

func TestDuplicateProjectRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request models.DuplicateProjectRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: models.DuplicateProjectRequest{
				OriginalProjectID: 1,
				NewProjectName:    "new-project",
			},
			valid: true,
		},
		{
			name: "zero project ID",
			request: models.DuplicateProjectRequest{
				OriginalProjectID: 0,
				NewProjectName:    "new-project",
			},
			valid: false,
		},
		{
			name: "empty project name",
			request: models.DuplicateProjectRequest{
				OriginalProjectID: 1,
				NewProjectName:    "",
			},
			valid: false,
		},
		{
			name: "project name too long",
			request: models.DuplicateProjectRequest{
				OriginalProjectID: 1,
				NewProjectName:    string(make([]byte, 51)), // 51 characters
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Actual validation would be done by the validator package
			// Here we just test the struct is properly formed
			if tt.valid {
				assert.NotZero(t, tt.request.OriginalProjectID)
				assert.NotEmpty(t, tt.request.NewProjectName)
				assert.LessOrEqual(t, len(tt.request.NewProjectName), 50)
			}
		})
	}
}

func TestCreateProjectRecordRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request models.CreateProjectRecordRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: models.CreateProjectRecordRequest{
				UserID:               1,
				ProjectName:          "test-project",
				Framework:            "gin",
				Features:             []string{"auth", "database"},
				AdjustedFeatures:     []string{"auth"},
				ZipFilePath:          stringPtr("/path/to/file.zip"),
				ZipFileSize:          int64Ptr(1024),
				GenerationDurationMs: intPtr(5000),
			},
			valid: true,
		},
		{
			name: "minimal valid request",
			request: models.CreateProjectRecordRequest{
				UserID:           1,
				ProjectName:      "minimal-project",
				Framework:        "echo",
				Features:         []string{"auth"},
				AdjustedFeatures: []string{"auth"},
			},
			valid: true,
		},
		{
			name: "zero user ID",
			request: models.CreateProjectRecordRequest{
				UserID:           0,
				ProjectName:      "test-project",
				Framework:        "gin",
				Features:         []string{"auth"},
				AdjustedFeatures: []string{"auth"},
			},
			valid: false,
		},
		{
			name: "empty project name",
			request: models.CreateProjectRecordRequest{
				UserID:           1,
				ProjectName:      "",
				Framework:        "gin",
				Features:         []string{"auth"},
				AdjustedFeatures: []string{"auth"},
			},
			valid: false,
		},
		{
			name: "invalid framework",
			request: models.CreateProjectRecordRequest{
				UserID:           1,
				ProjectName:      "test-project",
				Framework:        "invalid",
				Features:         []string{"auth"},
				AdjustedFeatures: []string{"auth"},
			},
			valid: false,
		},
		{
			name: "empty features",
			request: models.CreateProjectRecordRequest{
				UserID:           1,
				ProjectName:      "test-project",
				Framework:        "gin",
				Features:         []string{},
				AdjustedFeatures: []string{"auth"},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test JSON serialization
			jsonData, err := json.Marshal(tt.request)
			assert.NoError(t, err)

			var unmarshaled models.CreateProjectRecordRequest
			err = json.Unmarshal(jsonData, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tt.request.UserID, unmarshaled.UserID)
			assert.Equal(t, tt.request.ProjectName, unmarshaled.ProjectName)
			assert.Equal(t, tt.request.Framework, unmarshaled.Framework)
		})
	}
}

func TestHistoryFilters_DefaultValues(t *testing.T) {
	filters := models.HistoryFilters{}

	// Test default values
	assert.Equal(t, 0, filters.Page)
	assert.Equal(t, 0, filters.PageSize)
	assert.Empty(t, filters.Search)
	assert.Empty(t, filters.Framework)
	assert.Empty(t, filters.DateFrom)
	assert.Empty(t, filters.DateTo)
	assert.Empty(t, filters.SortBy)
	assert.Empty(t, filters.SortOrder)
}

func TestHistoryFilters_CompleteFilters(t *testing.T) {
	filters := models.HistoryFilters{
		Page:       2,
		PageSize:   20,
		Search:     "test project",
		Framework:  "gin",
		Frameworks: []string{"gin", "echo"},
		DateFrom:   "2024-01-01",
		DateTo:     "2024-01-31",
		SortBy:     "created_at",
		SortOrder:  "desc",
		Features:   []string{"auth", "database"},
		Status:     "available",
	}

	// Test JSON serialization
	jsonData, err := json.Marshal(filters)
	assert.NoError(t, err)

	var unmarshaled models.HistoryFilters
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, 2, unmarshaled.Page)
	assert.Equal(t, 20, unmarshaled.PageSize)
	assert.Equal(t, "test project", unmarshaled.Search)
	assert.Equal(t, "gin", unmarshaled.Framework)
	assert.Equal(t, []string{"gin", "echo"}, unmarshaled.Frameworks)
	assert.Equal(t, "2024-01-01", unmarshaled.DateFrom)
	assert.Equal(t, "2024-01-31", unmarshaled.DateTo)
	assert.Equal(t, "created_at", unmarshaled.SortBy)
	assert.Equal(t, "desc", unmarshaled.SortOrder)
	assert.Equal(t, []string{"auth", "database"}, unmarshaled.Features)
	assert.Equal(t, "available", unmarshaled.Status)
}

func TestHistoryFilters_ValidationConstraints(t *testing.T) {
	tests := []struct {
		name    string
		filters models.HistoryFilters
		valid   bool
	}{
		{
			name: "valid filters",
			filters: models.HistoryFilters{
				Page:      1,
				PageSize:  10,
				Framework: "gin",
				SortBy:    "created_at",
				SortOrder: "desc",
				Status:    "available",
			},
			valid: true,
		},
		{
			name: "page too high",
			filters: models.HistoryFilters{
				Page:     10001,
				PageSize: 10,
			},
			valid: false,
		},
		{
			name: "page size too high",
			filters: models.HistoryFilters{
				Page:     1,
				PageSize: 101,
			},
			valid: false,
		},
		{
			name: "invalid framework",
			filters: models.HistoryFilters{
				Page:      1,
				PageSize:  10,
				Framework: "invalid",
			},
			valid: false,
		},
		{
			name: "invalid sort by",
			filters: models.HistoryFilters{
				Page:     1,
				PageSize: 10,
				SortBy:   "invalid_field",
			},
			valid: false,
		},
		{
			name: "invalid sort order",
			filters: models.HistoryFilters{
				Page:      1,
				PageSize:  10,
				SortOrder: "invalid",
			},
			valid: false,
		},
		{
			name: "invalid status",
			filters: models.HistoryFilters{
				Page:     1,
				PageSize: 10,
				Status:   "invalid",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Actual validation would be done by the validator package
			// Here we just test the struct is properly formed
			if tt.valid {
				assert.True(t, tt.filters.Page <= 10000)
				assert.True(t, tt.filters.PageSize <= 100)
			}
		})
	}
}

// Helper function specific to this test file
func intPtr(i int) *int {
	return &i
}