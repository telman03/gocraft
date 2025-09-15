package models

import "time"

// ProjectHistoryResponse represents the response format for project history items
type ProjectHistoryResponse struct {
	ID                   uint      `json:"id"`
	ProjectName          string    `json:"project_name"`
	Framework            string    `json:"framework"`
	Features             []string  `json:"features"`
	AdjustedFeatures     []string  `json:"adjusted_features"`
	ZipFileSize          *int64    `json:"zip_file_size,omitempty"`
	ZipFileStatus        string    `json:"zip_file_status"`
	GenerationDurationMs *int      `json:"generation_duration_ms,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	CanDownload          bool      `json:"can_download"`
	CanRegenerate        bool      `json:"can_regenerate"`
}

// ProjectHistoryListResponse represents the paginated response for project history
type ProjectHistoryListResponse struct {
	Projects   []ProjectHistoryResponse `json:"projects"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

// ProjectStatsResponse represents user project statistics
type ProjectStatsResponse struct {
	TotalProjects         int                  `json:"total_projects"`
	MostUsedFramework     string               `json:"most_used_framework"`
	MostUsedFeatures      []string             `json:"most_used_features"`
	FrameworkDistribution map[string]int       `json:"framework_distribution"`
	RecentActivity        []DailyActivityCount `json:"recent_activity"`
}

// DailyActivityCount represents daily project generation activity
type DailyActivityCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// DuplicateProjectRequest represents the request to duplicate a project
type DuplicateProjectRequest struct {
	OriginalProjectID uint   `json:"original_project_id" validate:"required,min=1"`
	NewProjectName    string `json:"new_project_name" validate:"required,min=1,max=50,project_name,safe_string"`
}

// CreateProjectRecordRequest represents the request to create a project history record
type CreateProjectRecordRequest struct {
	UserID               uint     `json:"user_id" validate:"required,min=1"`
	ProjectName          string   `json:"project_name" validate:"required,min=1,max=100,safe_string,no_html"`
	Framework            string   `json:"framework" validate:"required,oneof=gin echo fiber"`
	Features             []string `json:"features" validate:"required,min=1,dive,safe_string"`
	AdjustedFeatures     []string `json:"adjusted_features" validate:"required,min=1,dive,safe_string"`
	ZipFilePath          *string  `json:"zip_file_path,omitempty" validate:"omitempty,max=500"`
	ZipFileSize          *int64   `json:"zip_file_size,omitempty" validate:"omitempty,min=0"`
	GenerationDurationMs *int     `json:"generation_duration_ms,omitempty" validate:"omitempty,min=0"`
}

// HistoryFilters represents filters for project history queries
type HistoryFilters struct {
	Page        int      `json:"page" validate:"min=1,max=10000"`
	PageSize    int      `json:"page_size" validate:"min=1,max=100"`
	Search      string   `json:"search,omitempty" validate:"omitempty,max=100,safe_string"`
	Framework   string   `json:"framework,omitempty" validate:"omitempty,oneof=gin echo fiber"`
	Frameworks  []string `json:"frameworks,omitempty" validate:"omitempty,dive,oneof=gin echo fiber"`
	DateFrom    string   `json:"date_from,omitempty" validate:"omitempty,datetime=2006-01-02"`
	DateTo      string   `json:"date_to,omitempty" validate:"omitempty,datetime=2006-01-02"`
	SortBy      string   `json:"sort_by,omitempty" validate:"omitempty,oneof=created_at project_name framework zip_file_size generation_duration_ms"`
	SortOrder   string   `json:"sort_order,omitempty" validate:"omitempty,oneof=asc desc"`
	Features    []string `json:"features,omitempty" validate:"omitempty,dive,safe_string"`
	Status      string   `json:"status,omitempty" validate:"omitempty,oneof=available expired deleted error"`
}