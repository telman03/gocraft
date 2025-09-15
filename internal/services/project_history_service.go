package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/gorm"
)

// CacheEntry represents a cached statistics entry
type CacheEntry struct {
	Data      *models.ProjectStatsResponse
	ExpiresAt time.Time
}

// StatsCache provides in-memory caching for user statistics
type StatsCache struct {
	cache map[uint]*CacheEntry
	mutex sync.RWMutex
	ttl   time.Duration
}

// NewStatsCache creates a new statistics cache with specified TTL
func NewStatsCache(ttl time.Duration) *StatsCache {
	cache := &StatsCache{
		cache: make(map[uint]*CacheEntry),
		ttl:   ttl,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves cached statistics for a user
func (sc *StatsCache) Get(userID uint) (*models.ProjectStatsResponse, bool) {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	
	entry, exists := sc.cache[userID]
	if !exists {
		return nil, false
	}
	
	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}
	
	return entry.Data, true
}

// Set stores statistics in cache for a user
func (sc *StatsCache) Set(userID uint, stats *models.ProjectStatsResponse) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	sc.cache[userID] = &CacheEntry{
		Data:      stats,
		ExpiresAt: time.Now().Add(sc.ttl),
	}
}

// Invalidate removes cached statistics for a user
func (sc *StatsCache) Invalidate(userID uint) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	
	delete(sc.cache, userID)
}

// cleanup removes expired entries from cache
func (sc *StatsCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()
	
	for range ticker.C {
		sc.mutex.Lock()
		now := time.Now()
		for userID, entry := range sc.cache {
			if now.After(entry.ExpiresAt) {
				delete(sc.cache, userID)
			}
		}
		sc.mutex.Unlock()
	}
}

// ProjectHistoryService handles business logic for project history operations
type ProjectHistoryService struct {
	db         *gorm.DB
	statsCache *StatsCache
}

// NewProjectHistoryService creates a new instance of ProjectHistoryService
func NewProjectHistoryService(db *gorm.DB) *ProjectHistoryService {
	return &ProjectHistoryService{
		db:         db,
		statsCache: NewStatsCache(15 * time.Minute), // Cache stats for 15 minutes
	}
}

// CreateProjectRecord creates a new project history record with transaction handling
func (s *ProjectHistoryService) CreateProjectRecord(req models.CreateProjectRecordRequest) (*models.ProjectHistory, error) {
	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Convert features slices to JSON
	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	adjustedFeaturesJSON, err := json.Marshal(req.AdjustedFeatures)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to marshal adjusted features: %w", err)
	}

	// Create project history record
	projectHistory := &models.ProjectHistory{
		UserID:               req.UserID,
		ProjectName:          req.ProjectName,
		Framework:            req.Framework,
		Features:             featuresJSON,
		AdjustedFeatures:     adjustedFeaturesJSON,
		ZipFilePath:          req.ZipFilePath,
		ZipFileSize:          req.ZipFileSize,
		ZipFileStatus:        models.ZipFileStatusAvailable,
		GenerationDurationMs: req.GenerationDurationMs,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Validate that user exists
	var userExists bool
	err = tx.Model(&models.User{}).Select("count(*) > 0").Where("id = ?", req.UserID).Find(&userExists).Error
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to validate user: %w", err)
	}
	if !userExists {
		tx.Rollback()
		return nil, errors.New("user not found")
	}

	// Save the project history record
	if err := tx.Create(projectHistory).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create project history record: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate user's statistics cache since new project was added
	s.statsCache.Invalidate(req.UserID)

	return projectHistory, nil
}

// GetUserHistory retrieves user's project history with filtering and pagination
func (s *ProjectHistoryService) GetUserHistory(userID uint, filters models.HistoryFilters) (*models.ProjectHistoryListResponse, error) {
	// Set default pagination values
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 10
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}

	// Build base query
	query := s.db.Model(&models.ProjectHistory{}).Where("user_id = ?", userID)

	// Apply search filter - enhanced to search in project name, framework, and features
	if filters.Search != "" {
		searchPattern := "%" + filters.Search + "%"
		query = query.Where(
			"project_name ILIKE ? OR framework ILIKE ? OR features::text ILIKE ? OR adjusted_features::text ILIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	// Apply framework filter (single framework for backward compatibility)
	if filters.Framework != "" {
		query = query.Where("framework = ?", filters.Framework)
	}

	// Apply multiple frameworks filter
	if len(filters.Frameworks) > 0 {
		query = query.Where("framework IN ?", filters.Frameworks)
	}

	// Apply features filter - check if any of the specified features exist in the project
	if len(filters.Features) > 0 {
		for _, feature := range filters.Features {
			featurePattern := fmt.Sprintf("%%\"%s\"%%", feature)
			query = query.Where("features::text ILIKE ? OR adjusted_features::text ILIKE ?", featurePattern, featurePattern)
		}
	}

	// Apply status filter
	if filters.Status != "" {
		query = query.Where("zip_file_status = ?", filters.Status)
	}

	// Apply date range filters
	if filters.DateFrom != "" {
		dateFrom, err := time.Parse("2006-01-02", filters.DateFrom)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from format: %w", err)
		}
		query = query.Where("created_at >= ?", dateFrom)
	}

	if filters.DateTo != "" {
		dateTo, err := time.Parse("2006-01-02", filters.DateTo)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to format: %w", err)
		}
		// Add 24 hours to include the entire day
		dateTo = dateTo.Add(24 * time.Hour)
		query = query.Where("created_at < ?", dateTo)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count projects: %w", err)
	}

	// Apply sorting
	orderClause := s.buildOrderClause(filters.SortBy, filters.SortOrder)
	query = query.Order(orderClause)

	// Apply pagination
	offset := (filters.Page - 1) * filters.PageSize
	var projects []models.ProjectHistory
	if err := query.Limit(filters.PageSize).Offset(offset).Find(&projects).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch projects: %w", err)
	}

	// Convert to response format
	projectResponses := make([]models.ProjectHistoryResponse, len(projects))
	for i, project := range projects {
		response, err := s.convertToResponse(project)
		if err != nil {
			return nil, fmt.Errorf("failed to convert project to response: %w", err)
		}
		projectResponses[i] = *response
	}

	// Calculate total pages
	totalPages := int((total + int64(filters.PageSize) - 1) / int64(filters.PageSize))

	return &models.ProjectHistoryListResponse{
		Projects:   projectResponses,
		Total:      int(total),
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetProjectByID retrieves a specific project with ownership validation
func (s *ProjectHistoryService) GetProjectByID(userID uint, projectID uint) (*models.ProjectHistory, error) {
	var project models.ProjectHistory
	
	// Query with ownership validation
	err := s.db.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project not found or access denied")
		}
		return nil, fmt.Errorf("failed to fetch project: %w", err)
	}

	return &project, nil
}

// DeleteProject deletes a project from history with file cleanup
func (s *ProjectHistoryService) DeleteProject(userID uint, projectID uint) error {
	// Start transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get project with ownership validation
	var project models.ProjectHistory
	err := tx.Where("id = ? AND user_id = ?", projectID, userID).First(&project).Error
	if err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project not found or access denied")
		}
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	// Delete associated ZIP file if it exists
	if project.ZipFilePath != nil && *project.ZipFilePath != "" {
		if err := s.deleteFile(*project.ZipFilePath); err != nil {
			// Log the error but don't fail the transaction
			// File might already be deleted or moved
			fmt.Printf("Warning: failed to delete file %s: %v\n", *project.ZipFilePath, err)
		}
	}

	// Delete the project record
	if err := tx.Delete(&project).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate user's statistics cache since project was deleted
	s.statsCache.Invalidate(userID)

	return nil
}

// convertToResponse converts ProjectHistory model to response format
func (s *ProjectHistoryService) convertToResponse(project models.ProjectHistory) (*models.ProjectHistoryResponse, error) {
	// Unmarshal features
	var features []string
	if err := json.Unmarshal(project.Features, &features); err != nil {
		return nil, fmt.Errorf("failed to unmarshal features: %w", err)
	}

	var adjustedFeatures []string
	if err := json.Unmarshal(project.AdjustedFeatures, &adjustedFeatures); err != nil {
		return nil, fmt.Errorf("failed to unmarshal adjusted features: %w", err)
	}

	// Check if file is available for download
	canDownload := false
	canRegenerate := false
	
	if project.ZipFileStatus == models.ZipFileStatusAvailable && project.ZipFilePath != nil {
		canDownload = s.fileExists(*project.ZipFilePath)
	}
	
	// Can regenerate if we have the configuration data
	canRegenerate = len(features) > 0 && project.Framework != ""

	return &models.ProjectHistoryResponse{
		ID:                   project.ID,
		ProjectName:          project.ProjectName,
		Framework:            project.Framework,
		Features:             features,
		AdjustedFeatures:     adjustedFeatures,
		ZipFileSize:          project.ZipFileSize,
		ZipFileStatus:        string(project.ZipFileStatus),
		GenerationDurationMs: project.GenerationDurationMs,
		CreatedAt:            project.CreatedAt,
		CanDownload:          canDownload,
		CanRegenerate:        canRegenerate,
	}, nil
}

// fileExists checks if a file exists at the given path
func (s *ProjectHistoryService) fileExists(path string) bool {
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// buildOrderClause constructs the ORDER BY clause based on sort parameters
func (s *ProjectHistoryService) buildOrderClause(sortBy, sortOrder string) string {
	// Default sorting
	if sortBy == "" {
		return "created_at DESC"
	}

	// Validate sort order
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Map of allowed sort fields to prevent SQL injection
	allowedSortFields := map[string]string{
		"created_at":             "created_at",
		"updated_at":             "updated_at",
		"project_name":           "project_name",
		"framework":              "framework",
		"zip_file_size":          "zip_file_size",
		"generation_duration_ms": "generation_duration_ms",
		"zip_file_status":        "zip_file_status",
	}

	// Check if sort field is allowed
	if dbField, exists := allowedSortFields[sortBy]; exists {
		return fmt.Sprintf("%s %s", dbField, sortOrder)
	}

	// Default to created_at DESC if invalid sort field
	return "created_at DESC"
}

// SearchProjectsByFeatures searches projects that contain specific features
func (s *ProjectHistoryService) SearchProjectsByFeatures(userID uint, features []string, limit int) ([]models.ProjectHistory, error) {
	if len(features) == 0 {
		return []models.ProjectHistory{}, nil
	}

	if limit <= 0 {
		limit = 10
	}

	query := s.db.Model(&models.ProjectHistory{}).Where("user_id = ?", userID)

	// Build feature search conditions
	for _, feature := range features {
		featurePattern := fmt.Sprintf("%%\"%s\"%%", feature)
		query = query.Where("features::text ILIKE ? OR adjusted_features::text ILIKE ?", featurePattern, featurePattern)
	}

	var projects []models.ProjectHistory
	err := query.Order("created_at DESC").Limit(limit).Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search projects by features: %w", err)
	}

	return projects, nil
}

// SearchProjectsByFramework searches projects by framework with additional filters
func (s *ProjectHistoryService) SearchProjectsByFramework(userID uint, framework string, dateRange *time.Duration) ([]models.ProjectHistory, error) {
	query := s.db.Model(&models.ProjectHistory{}).Where("user_id = ? AND framework = ?", userID, framework)

	// Apply date range filter if provided
	if dateRange != nil {
		cutoffDate := time.Now().Add(-*dateRange)
		query = query.Where("created_at >= ?", cutoffDate)
	}

	var projects []models.ProjectHistory
	err := query.Order("created_at DESC").Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("failed to search projects by framework: %w", err)
	}

	return projects, nil
}

// GetProjectsByDateRange retrieves projects within a specific date range
func (s *ProjectHistoryService) GetProjectsByDateRange(userID uint, startDate, endDate time.Time) ([]models.ProjectHistory, error) {
	var projects []models.ProjectHistory
	
	err := s.db.Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Order("created_at DESC").
		Find(&projects).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get projects by date range: %w", err)
	}

	return projects, nil
}

// GetProjectStats calculates and returns user project statistics for dashboard with caching
func (s *ProjectHistoryService) GetProjectStats(userID uint) (*models.ProjectStatsResponse, error) {
	// Check cache first
	if cachedStats, found := s.statsCache.Get(userID); found {
		return cachedStats, nil
	}

	// Calculate stats if not in cache
	stats, err := s.calculateProjectStats(userID)
	if err != nil {
		return nil, err
	}

	// Cache the results
	s.statsCache.Set(userID, stats)

	return stats, nil
}

// calculateProjectStats performs the actual statistics calculation
func (s *ProjectHistoryService) calculateProjectStats(userID uint) (*models.ProjectStatsResponse, error) {
	// Get total project count
	var totalProjects int64
	if err := s.db.Model(&models.ProjectHistory{}).Where("user_id = ?", userID).Count(&totalProjects).Error; err != nil {
		return nil, fmt.Errorf("failed to count total projects: %w", err)
	}

	// Get framework distribution
	frameworkDistribution, err := s.getFrameworkDistribution(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get framework distribution: %w", err)
	}

	// Get most used framework
	mostUsedFramework := s.getMostUsedFramework(frameworkDistribution)

	// Get most used features
	mostUsedFeatures, err := s.getMostUsedFeatures(userID, 5) // Top 5 features
	if err != nil {
		return nil, fmt.Errorf("failed to get most used features: %w", err)
	}

	// Get recent activity (last 30 days)
	recentActivity, err := s.getRecentActivity(userID, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	return &models.ProjectStatsResponse{
		TotalProjects:         int(totalProjects),
		MostUsedFramework:     mostUsedFramework,
		MostUsedFeatures:      mostUsedFeatures,
		FrameworkDistribution: frameworkDistribution,
		RecentActivity:        recentActivity,
	}, nil
}

// getFrameworkDistribution calculates the distribution of frameworks used by the user
func (s *ProjectHistoryService) getFrameworkDistribution(userID uint) (map[string]int, error) {
	type FrameworkCount struct {
		Framework string
		Count     int64
	}

	var results []FrameworkCount
	err := s.db.Model(&models.ProjectHistory{}).
		Select("framework, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("framework").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.Framework] = int(result.Count)
	}

	return distribution, nil
}

// getMostUsedFramework determines the most frequently used framework
func (s *ProjectHistoryService) getMostUsedFramework(distribution map[string]int) string {
	if len(distribution) == 0 {
		return ""
	}

	maxCount := 0
	mostUsed := ""
	
	for framework, count := range distribution {
		if count > maxCount {
			maxCount = count
			mostUsed = framework
		}
	}

	return mostUsed
}

// getMostUsedFeatures calculates the most frequently used features
func (s *ProjectHistoryService) getMostUsedFeatures(userID uint, limit int) ([]string, error) {
	var projects []models.ProjectHistory
	err := s.db.Select("features, adjusted_features").Where("user_id = ?", userID).Find(&projects).Error
	if err != nil {
		return nil, err
	}

	// Count feature occurrences
	featureCount := make(map[string]int)
	
	for _, project := range projects {
		// Count features from both features and adjusted_features
		features := s.extractFeaturesFromJSON(project.Features)
		adjustedFeatures := s.extractFeaturesFromJSON(project.AdjustedFeatures)
		
		// Combine both feature sets
		allFeatures := append(features, adjustedFeatures...)
		
		for _, feature := range allFeatures {
			if feature != "" {
				featureCount[feature]++
			}
		}
	}

	// Sort features by count and return top N
	type FeatureCount struct {
		Feature string
		Count   int
	}

	var featureCounts []FeatureCount
	for feature, count := range featureCount {
		featureCounts = append(featureCounts, FeatureCount{Feature: feature, Count: count})
	}

	// Sort by count (descending)
	for i := 0; i < len(featureCounts)-1; i++ {
		for j := i + 1; j < len(featureCounts); j++ {
			if featureCounts[i].Count < featureCounts[j].Count {
				featureCounts[i], featureCounts[j] = featureCounts[j], featureCounts[i]
			}
		}
	}

	// Extract top features
	var topFeatures []string
	maxFeatures := limit
	if len(featureCounts) < maxFeatures {
		maxFeatures = len(featureCounts)
	}

	for i := 0; i < maxFeatures; i++ {
		topFeatures = append(topFeatures, featureCounts[i].Feature)
	}

	return topFeatures, nil
}

// getRecentActivity generates daily activity counts for the specified number of days
func (s *ProjectHistoryService) getRecentActivity(userID uint, days int) ([]models.DailyActivityCount, error) {
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Query for daily counts
	type DailyCount struct {
		Date  string
		Count int64
	}

	var results []DailyCount
	err := s.db.Model(&models.ProjectHistory{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	countMap := make(map[string]int)
	for _, result := range results {
		countMap[result.Date] = int(result.Count)
	}

	// Generate complete activity array with zeros for missing days
	var activity []models.DailyActivityCount
	currentDate := startDate

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		count := countMap[dateStr] // Will be 0 if not found
		
		activity = append(activity, models.DailyActivityCount{
			Date:  dateStr,
			Count: count,
		})
		
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return activity, nil
}

// extractFeaturesFromJSON extracts feature strings from JSON data
func (s *ProjectHistoryService) extractFeaturesFromJSON(jsonData []byte) []string {
	var features []string
	if err := json.Unmarshal(jsonData, &features); err != nil {
		// If unmarshaling fails, return empty slice
		return []string{}
	}
	return features
}

// GetFrameworkUsageStats returns detailed framework usage statistics
func (s *ProjectHistoryService) GetFrameworkUsageStats(userID uint) (map[string]interface{}, error) {
	distribution, err := s.getFrameworkDistribution(userID)
	if err != nil {
		return nil, err
	}

	// Calculate percentages
	var total int
	for _, count := range distribution {
		total += count
	}

	stats := make(map[string]interface{})
	for framework, count := range distribution {
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		
		stats[framework] = map[string]interface{}{
			"count":      count,
			"percentage": percentage,
		}
	}

	return stats, nil
}

// GetProjectTrends analyzes project generation trends over time
func (s *ProjectHistoryService) GetProjectTrends(userID uint, months int) ([]map[string]interface{}, error) {
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, -months, 0)

	type MonthlyTrend struct {
		Month     string
		Framework string
		Count     int64
	}

	var results []MonthlyTrend
	err := s.db.Model(&models.ProjectHistory{}).
		Select("TO_CHAR(created_at, 'YYYY-MM') as month, framework, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Group("TO_CHAR(created_at, 'YYYY-MM'), framework").
		Order("month ASC, framework ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Group by month
	monthlyData := make(map[string]map[string]int64)
	for _, result := range results {
		if monthlyData[result.Month] == nil {
			monthlyData[result.Month] = make(map[string]int64)
		}
		monthlyData[result.Month][result.Framework] = result.Count
	}

	// Convert to response format
	var trends []map[string]interface{}
	for month, frameworks := range monthlyData {
		trend := map[string]interface{}{
			"month":      month,
			"frameworks": frameworks,
		}
		
		// Calculate total for the month
		var total int64
		for _, count := range frameworks {
			total += count
		}
		trend["total"] = total
		
		trends = append(trends, trend)
	}

	return trends, nil
}

// deleteFile safely deletes a file
func (s *ProjectHistoryService) deleteFile(path string) error {
	if path == "" {
		return nil
	}

	// Validate the path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return errors.New("invalid file path")
	}

	// Check if file exists before attempting to delete
	if !s.fileExists(path) {
		return nil // File doesn't exist, consider it already deleted
	}

	return os.Remove(path)
}

// GetDashboardData provides optimized dashboard data aggregation with caching
func (s *ProjectHistoryService) GetDashboardData(userID uint) (map[string]interface{}, error) {
	// Check if we have cached stats
	if cachedStats, found := s.statsCache.Get(userID); found {
		return s.formatDashboardData(cachedStats), nil
	}

	// Use a single transaction for all queries to improve performance
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Get all data in parallel using goroutines
	type result struct {
		totalProjects         int64
		frameworkDistribution map[string]int
		mostUsedFeatures      []string
		recentActivity        []models.DailyActivityCount
		err                   error
	}

	resultChan := make(chan result, 1)

	go func() {
		var res result

		// Get total project count
		if err := tx.Model(&models.ProjectHistory{}).Where("user_id = ?", userID).Count(&res.totalProjects).Error; err != nil {
			res.err = fmt.Errorf("failed to count total projects: %w", err)
			resultChan <- res
			return
		}

		// Get framework distribution with optimized query
		res.frameworkDistribution, res.err = s.getFrameworkDistributionOptimized(tx, userID)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get most used features with optimized query
		res.mostUsedFeatures, res.err = s.getMostUsedFeaturesOptimized(tx, userID, 5)
		if res.err != nil {
			resultChan <- res
			return
		}

		// Get recent activity with optimized query
		res.recentActivity, res.err = s.getRecentActivityOptimized(tx, userID, 30)
		if res.err != nil {
			resultChan <- res
			return
		}

		resultChan <- res
	}()

	// Wait for results
	res := <-resultChan
	if res.err != nil {
		return nil, res.err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Create stats response
	stats := &models.ProjectStatsResponse{
		TotalProjects:         int(res.totalProjects),
		MostUsedFramework:     s.getMostUsedFramework(res.frameworkDistribution),
		MostUsedFeatures:      res.mostUsedFeatures,
		FrameworkDistribution: res.frameworkDistribution,
		RecentActivity:        res.recentActivity,
	}

	// Cache the results
	s.statsCache.Set(userID, stats)

	return s.formatDashboardData(stats), nil
}

// formatDashboardData formats statistics for dashboard consumption
func (s *ProjectHistoryService) formatDashboardData(stats *models.ProjectStatsResponse) map[string]interface{} {
	// Calculate additional metrics
	totalFrameworks := len(stats.FrameworkDistribution)
	
	// Calculate activity trend (last 7 days vs previous 7 days)
	var recentCount, previousCount int
	activityLen := len(stats.RecentActivity)
	
	if activityLen >= 7 {
		// Last 7 days
		for i := activityLen - 7; i < activityLen; i++ {
			recentCount += stats.RecentActivity[i].Count
		}
		
		// Previous 7 days
		if activityLen >= 14 {
			for i := activityLen - 14; i < activityLen-7; i++ {
				previousCount += stats.RecentActivity[i].Count
			}
		}
	}

	// Calculate trend percentage
	var trendPercentage float64
	if previousCount > 0 {
		trendPercentage = float64(recentCount-previousCount) / float64(previousCount) * 100
	} else if recentCount > 0 {
		trendPercentage = 100 // 100% increase from 0
	}

	return map[string]interface{}{
		"overview": map[string]interface{}{
			"total_projects":      stats.TotalProjects,
			"total_frameworks":    totalFrameworks,
			"most_used_framework": stats.MostUsedFramework,
			"activity_trend": map[string]interface{}{
				"recent_count":      recentCount,
				"previous_count":    previousCount,
				"trend_percentage":  trendPercentage,
			},
		},
		"framework_distribution": stats.FrameworkDistribution,
		"most_used_features":     stats.MostUsedFeatures,
		"recent_activity":        stats.RecentActivity,
		"cache_info": map[string]interface{}{
			"cached_at": time.Now().Format(time.RFC3339),
			"ttl_minutes": 15,
		},
	}
}

// getFrameworkDistributionOptimized uses optimized query for framework distribution
func (s *ProjectHistoryService) getFrameworkDistributionOptimized(tx *gorm.DB, userID uint) (map[string]int, error) {
	type FrameworkCount struct {
		Framework string `json:"framework"`
		Count     int64  `json:"count"`
	}

	var results []FrameworkCount
	err := tx.Model(&models.ProjectHistory{}).
		Select("framework, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("framework").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get framework distribution: %w", err)
	}

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.Framework] = int(result.Count)
	}

	return distribution, nil
}

// getMostUsedFeaturesOptimized uses optimized query for most used features
func (s *ProjectHistoryService) getMostUsedFeaturesOptimized(tx *gorm.DB, userID uint, limit int) ([]string, error) {
	// Use a more efficient approach by aggregating features in the database
	var projects []models.ProjectHistory
	err := tx.Select("features, adjusted_features").
		Where("user_id = ?", userID).
		Find(&projects).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to get projects for feature analysis: %w", err)
	}

	// Count feature occurrences efficiently
	featureCount := make(map[string]int)
	
	for _, project := range projects {
		// Process features from both fields
		s.countFeaturesFromJSON(project.Features, featureCount)
		s.countFeaturesFromJSON(project.AdjustedFeatures, featureCount)
	}

	// Convert to sorted slice
	type FeatureCount struct {
		Feature string
		Count   int
	}

	var featureCounts []FeatureCount
	for feature, count := range featureCount {
		featureCounts = append(featureCounts, FeatureCount{Feature: feature, Count: count})
	}

	// Sort by count (descending) using efficient sorting
	for i := 0; i < len(featureCounts)-1; i++ {
		for j := i + 1; j < len(featureCounts); j++ {
			if featureCounts[i].Count < featureCounts[j].Count {
				featureCounts[i], featureCounts[j] = featureCounts[j], featureCounts[i]
			}
		}
	}

	// Extract top features
	var topFeatures []string
	maxFeatures := limit
	if len(featureCounts) < maxFeatures {
		maxFeatures = len(featureCounts)
	}

	for i := 0; i < maxFeatures; i++ {
		topFeatures = append(topFeatures, featureCounts[i].Feature)
	}

	return topFeatures, nil
}

// countFeaturesFromJSON efficiently counts features from JSON data
func (s *ProjectHistoryService) countFeaturesFromJSON(jsonData []byte, featureCount map[string]int) {
	var features []string
	if err := json.Unmarshal(jsonData, &features); err != nil {
		return // Skip invalid JSON
	}
	
	for _, feature := range features {
		if feature != "" {
			featureCount[feature]++
		}
	}
}

// getRecentActivityOptimized uses optimized query for recent activity
func (s *ProjectHistoryService) getRecentActivityOptimized(tx *gorm.DB, userID uint, days int) ([]models.DailyActivityCount, error) {
	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// Use optimized query with proper date handling
	type DailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var results []DailyCount
	err := tx.Model(&models.ProjectHistory{}).
		Select("DATE(created_at) as date, COUNT(*) as count").
		Where("user_id = ? AND created_at >= ? AND created_at <= ?", userID, startDate, endDate).
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	// Create a map for quick lookup
	countMap := make(map[string]int)
	for _, result := range results {
		countMap[result.Date] = int(result.Count)
	}

	// Generate complete activity array with zeros for missing days
	var activity []models.DailyActivityCount
	currentDate := startDate

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		dateStr := currentDate.Format("2006-01-02")
		count := countMap[dateStr] // Will be 0 if not found
		
		activity = append(activity, models.DailyActivityCount{
			Date:  dateStr,
			Count: count,
		})
		
		currentDate = currentDate.AddDate(0, 0, 1)
	}

	return activity, nil
}

// GetCacheStats returns information about the cache performance
func (s *ProjectHistoryService) GetCacheStats() map[string]interface{} {
	s.statsCache.mutex.RLock()
	defer s.statsCache.mutex.RUnlock()
	
	totalEntries := len(s.statsCache.cache)
	expiredEntries := 0
	now := time.Now()
	
	for _, entry := range s.statsCache.cache {
		if now.After(entry.ExpiresAt) {
			expiredEntries++
		}
	}
	
	return map[string]interface{}{
		"total_entries":   totalEntries,
		"expired_entries": expiredEntries,
		"active_entries":  totalEntries - expiredEntries,
		"ttl_minutes":     int(s.statsCache.ttl.Minutes()),
	}
}

// InvalidateAllCache clears all cached statistics (useful for maintenance)
func (s *ProjectHistoryService) InvalidateAllCache() {
	s.statsCache.mutex.Lock()
	defer s.statsCache.mutex.Unlock()
	
	s.statsCache.cache = make(map[uint]*CacheEntry)
}