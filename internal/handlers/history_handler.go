package handlers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/builder"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/middleware"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/services"
	"github.com/telman03/ai-backend-generator/internal/utils"
)

// GetProjectHistory godoc
// @Summary Get user's project history
// @Description Retrieves paginated list of user's generated projects with filtering and search capabilities
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Items per page (default: 10, max: 100)"
// @Param search query string false "Search in project name, framework, or features"
// @Param framework query string false "Filter by framework (gin, echo, fiber)"
// @Param frameworks query string false "Filter by multiple frameworks (comma-separated)"
// @Param features query string false "Filter by features (comma-separated)"
// @Param status query string false "Filter by ZIP file status (available, expired, deleted, error)"
// @Param date_from query string false "Filter from date (YYYY-MM-DD format)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD format)"
// @Param sort_by query string false "Sort field (created_at, project_name, framework, zip_file_size, generation_duration_ms)"
// @Param sort_order query string false "Sort order (asc, desc, default: desc)"
// @Success 200 {object} models.ProjectHistoryListResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history [get]
func GetProjectHistory(c *fiber.Ctx) error {
	// Get validated user ID from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "get_history", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	// Parse query parameters
	filters := models.HistoryFilters{}

	// Parse pagination parameters
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Page = page
		} else {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid page parameter")
		}
	} else {
		filters.Page = 1 // Default page
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			if pageSize > 100 {
				pageSize = 100 // Cap at 100
			}
			filters.PageSize = pageSize
		} else {
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid page_size parameter")
		}
	} else {
		filters.PageSize = 10 // Default page size
	}

	// Parse search parameter
	filters.Search = strings.TrimSpace(c.Query("search"))

	// Parse framework filters
	filters.Framework = strings.TrimSpace(c.Query("framework"))
	if frameworksStr := c.Query("frameworks"); frameworksStr != "" {
		frameworks := strings.Split(frameworksStr, ",")
		for i, fw := range frameworks {
			frameworks[i] = strings.TrimSpace(fw)
		}
		// Filter out empty strings
		var validFrameworks []string
		for _, fw := range frameworks {
			if fw != "" {
				validFrameworks = append(validFrameworks, fw)
			}
		}
		filters.Frameworks = validFrameworks
	}

	// Parse features filter
	if featuresStr := c.Query("features"); featuresStr != "" {
		features := strings.Split(featuresStr, ",")
		for i, feature := range features {
			features[i] = strings.TrimSpace(feature)
		}
		// Filter out empty strings
		var validFeatures []string
		for _, feature := range features {
			if feature != "" {
				validFeatures = append(validFeatures, feature)
			}
		}
		filters.Features = validFeatures
	}

	// Parse status filter
	filters.Status = strings.TrimSpace(c.Query("status"))

	// Parse date filters
	filters.DateFrom = strings.TrimSpace(c.Query("date_from"))
	filters.DateTo = strings.TrimSpace(c.Query("date_to"))

	// Parse sorting parameters
	filters.SortBy = strings.TrimSpace(c.Query("sort_by"))
	filters.SortOrder = strings.ToLower(strings.TrimSpace(c.Query("sort_order")))

	// Validate filters
	if validationErr := utils.ValidateStruct(&filters); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get user history
	response, err := historyService.GetUserHistory(userIDUint, filters)
	if err != nil {
		middleware.LogUserAction(c, userIDUint, "get_project_history", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve project history", map[string]string{
			"code": "HISTORY_RETRIEVAL_FAILED",
			"details": err.Error(),
		})
	}

	// Log successful history retrieval
	middleware.LogUserAction(c, userIDUint, "get_project_history", "project_history", true, "")

	return c.JSON(response)
}

// GetProjectDetails godoc
// @Summary Get specific project details
// @Description Retrieves detailed information about a specific project with ownership validation
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} models.ProjectHistoryResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/{id} [get]
func GetProjectDetails(c *fiber.Ctx) error {
	// Get validated user and project IDs from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "get_project_details", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	projectID, err := middleware.GetValidatedProjectID(c)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, 0, "get_project_details", false, "INVALID_PROJECT_ID")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid project ID", map[string]string{
			"code": "INVALID_PROJECT_ID",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get project (ownership already validated by middleware)
	project, err := historyService.GetProjectByID(userIDUint, projectID)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, projectID, "get_project_details", false, "PROJECT_NOT_FOUND")
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Project not found", map[string]string{
			"code": "PROJECT_NOT_FOUND",
			"details": err.Error(),
		})
	}

	// Log successful access
	middleware.LogProjectAccess(c, userIDUint, projectID, "get_project_details", true, "")

	// Create file service to check file availability
	fileService := services.NewFileService("./output", 30*24*time.Hour) // 30 days retention

	// Convert to response format with file availability check
	response, err := convertProjectToDetailedResponse(project, fileService)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to process project details", map[string]string{
			"details": err.Error(),
		})
	}

	return c.JSON(response)
}

// DeleteProject godoc
// @Summary Delete project from history
// @Description Deletes a project from user's history with ownership validation and file cleanup
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Project ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/{id} [delete]
func DeleteProject(c *fiber.Ctx) error {
	// Get validated user and project IDs from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "delete_project", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	projectID, err := middleware.GetValidatedProjectID(c)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, 0, "delete_project", false, "INVALID_PROJECT_ID")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid project ID", map[string]string{
			"code": "INVALID_PROJECT_ID",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Delete project (ownership already validated by middleware)
	err = historyService.DeleteProject(userIDUint, projectID)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, projectID, "delete_project", false, "DELETE_FAILED")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to delete project", map[string]string{
			"code": "DELETE_FAILED",
			"details": err.Error(),
		})
	}

	// Log successful deletion
	middleware.LogProjectAccess(c, userIDUint, projectID, "delete_project", true, "")

	return c.JSON(fiber.Map{
		"message":    "Project deleted successfully",
		"project_id": projectID,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// DownloadProject godoc
// @Summary Download project ZIP file
// @Description Downloads the ZIP file for a specific project with ownership validation and security checks
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce application/zip
// @Param id path int true "Project ID"
// @Success 200 {file} file "Project ZIP file"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 410 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/{id}/download [get]
func DownloadProject(c *fiber.Ctx) error {
	// Get validated user and project IDs from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "download_project", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	projectID, err := middleware.GetValidatedProjectID(c)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, 0, "download_project", false, "INVALID_PROJECT_ID")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid project ID", map[string]string{
			"code": "INVALID_PROJECT_ID",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get project (ownership already validated by middleware)
	project, err := historyService.GetProjectByID(userIDUint, projectID)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, projectID, "download_project", false, "PROJECT_NOT_FOUND")
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Project not found", map[string]string{
			"code": "PROJECT_NOT_FOUND",
			"details": err.Error(),
		})
	}

	// Check if ZIP file path exists
	if project.ZipFilePath == nil || *project.ZipFilePath == "" {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Project file not available", map[string]string{
			"code": "FILE_NOT_AVAILABLE",
			"details": "No file path associated with this project",
		})
	}

	// Create file service to validate file
	fileService := services.NewFileService("./output", 30*24*time.Hour) // 30 days retention

	// Validate file path and check existence
	if err := fileService.ValidateFilePath(*project.ZipFilePath); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid file path", map[string]string{
			"code": "INVALID_FILE_PATH",
			"details": err.Error(),
		})
	}

	// Check if file exists
	if !fileService.FileExists(*project.ZipFilePath) {
		// Update project status to expired if file doesn't exist
		go func() {
			// Update status in background to avoid blocking the response
			project.ZipFileStatus = models.ZipFileStatusExpired
			database.DB.Save(project)
		}()

		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Project file no longer available", map[string]string{
			"code": "FILE_EXPIRED",
			"details": "The file may have been deleted or expired",
			"can_regenerate": "true",
		})
	}

	// Check if file is expired
	if fileService.IsFileExpired(*project.ZipFilePath) {
		// Update project status to expired
		go func() {
			project.ZipFileStatus = models.ZipFileStatusExpired
			database.DB.Save(project)
		}()

		return utils.SendErrorResponse(c, fiber.StatusGone, "Project file has expired", map[string]string{
			"code": "FILE_EXPIRED",
			"details": "File has exceeded the retention period",
			"can_regenerate": "true",
		})
	}

	// Get file info for headers
	fileInfo, err := fileService.GetFileInfo(*project.ZipFilePath)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to get file information", map[string]string{
			"details": err.Error(),
		})
	}

	// Log download activity (for tracking and analytics)
	middleware.LogProjectAccess(c, userIDUint, projectID, "download_project", true, "")
	go func() {
		// Log download in background - you could extend this to track download statistics
		fmt.Printf("User %d downloaded project %d (%s) at %s\n", 
			userIDUint, project.ID, project.ProjectName, time.Now().Format(time.RFC3339))
	}()

	// Set appropriate headers for file download
	filename := fmt.Sprintf("%s_%s.zip", project.ProjectName, project.Framework)
	// Sanitize filename for download
	filename = sanitizeDownloadFilename(filename)
	
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	// Send the file
	return c.SendFile(*project.ZipFilePath)
}

// RegenerateProject godoc
// @Summary Regenerate project with same configuration
// @Description Regenerates a project using the same configuration as a previous project with ownership validation
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce application/zip
// @Param id path int true "Project ID"
// @Success 200 {file} file "Regenerated project ZIP file"
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/{id}/regenerate [post]
func RegenerateProject(c *fiber.Ctx) error {
	// Get validated user and project IDs from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "regenerate_project", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	projectID, err := middleware.GetValidatedProjectID(c)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, 0, "regenerate_project", false, "INVALID_PROJECT_ID")
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid project ID", map[string]string{
			"code": "INVALID_PROJECT_ID",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get original project (ownership already validated by middleware)
	originalProject, err := historyService.GetProjectByID(userIDUint, projectID)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, projectID, "regenerate_project", false, "PROJECT_NOT_FOUND")
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Project not found", map[string]string{
			"code": "PROJECT_NOT_FOUND",
			"details": err.Error(),
		})
	}

	// Extract features from the original project
	var originalFeatures []string
	if err := json.Unmarshal(originalProject.Features, &originalFeatures); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to parse original project features", map[string]string{
			"details": err.Error(),
		})
	}

	var adjustedFeatures []string
	if err := json.Unmarshal(originalProject.AdjustedFeatures, &adjustedFeatures); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to parse adjusted project features", map[string]string{
			"details": err.Error(),
		})
	}

	// Validate that we have the necessary data for regeneration
	if len(adjustedFeatures) == 0 || originalProject.Framework == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Cannot regenerate project: insufficient configuration data", map[string]string{
			"code": "INSUFFICIENT_DATA",
			"details": "Original project lacks necessary configuration information",
		})
	}

	// Generate a new project name with timestamp to avoid conflicts
	timestamp := time.Now().Format("20060102_150405")
	newProjectName := fmt.Sprintf("%s_regenerated_%s", originalProject.ProjectName, timestamp)

	// Log regeneration activity
	fmt.Printf("User %d regenerating project %d (%s) as '%s' with features: %v\n", 
		userIDUint, originalProject.ID, originalProject.ProjectName, newProjectName, adjustedFeatures)

	// Start timing the generation
	startTime := time.Now()

	// Generate the project using the builder
	zipPath, err := builder.GenerateProject(newProjectName, adjustedFeatures)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to regenerate project", map[string]string{
			"details": err.Error(),
			"original_project": originalProject.ProjectName,
		})
	}

	// Calculate generation duration
	generationDuration := time.Since(startTime)
	generationDurationMs := int(generationDuration.Milliseconds())

	// Get file size
	fileService := services.NewFileService("./output", 30*24*time.Hour)
	fileInfo, err := fileService.GetFileInfo(zipPath)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to get file info for regenerated project: %v\n", err)
	}

	// Create new project history record for the regenerated project
	go func() {
		// Create history record in background to avoid blocking the download
		createReq := models.CreateProjectRecordRequest{
			UserID:               userIDUint,
			ProjectName:          newProjectName,
			Framework:            originalProject.Framework,
			Features:             originalFeatures,
			AdjustedFeatures:     adjustedFeatures,
			ZipFilePath:          &zipPath,
			GenerationDurationMs: &generationDurationMs,
		}

		if fileInfo != nil {
			createReq.ZipFileSize = &fileInfo.Size
		}

		_, err := historyService.CreateProjectRecord(createReq)
		if err != nil {
			fmt.Printf("Warning: failed to create history record for regenerated project: %v\n", err)
		} else {
			fmt.Printf("Successfully created history record for regenerated project: %s\n", newProjectName)
		}
	}()

	// Set proper headers for download
	filename := fmt.Sprintf("%s_%s.zip", newProjectName, originalProject.Framework)
	filename = sanitizeDownloadFilename(filename)
	
	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	if fileInfo != nil {
		c.Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size))
	}
	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Pragma", "no-cache")
	c.Set("Expires", "0")

	// Log successful regeneration
	middleware.LogProjectAccess(c, userIDUint, projectID, "regenerate_project", true, "")

	// Send the regenerated file
	return c.SendFile(zipPath)
}

// DuplicateProject godoc
// @Summary Duplicate project configuration
// @Description Creates a duplicate configuration from an existing project for pre-populating the generator form
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param data body models.DuplicateProjectRequest true "Duplicate Project Request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/duplicate [post]
func DuplicateProject(c *fiber.Ctx) error {
	// Get validated user ID from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "duplicate_project", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	// Parse request body
	var req models.DuplicateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format", map[string]string{
			"details": err.Error(),
		})
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Validate ownership of the original project
	if err := middleware.ValidateProjectOwnership(userIDUint, req.OriginalProjectID); err != nil {
		middleware.LogProjectAccess(c, userIDUint, req.OriginalProjectID, "duplicate_project", false, "ACCESS_DENIED")
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Original project not found or access denied", map[string]string{
			"code": "PROJECT_NOT_FOUND",
			"details": "You don't have permission to access this project",
		})
	}

	// Get original project
	originalProject, err := historyService.GetProjectByID(userIDUint, req.OriginalProjectID)
	if err != nil {
		middleware.LogProjectAccess(c, userIDUint, req.OriginalProjectID, "duplicate_project", false, "PROJECT_NOT_FOUND")
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "Original project not found", map[string]string{
			"code": "PROJECT_NOT_FOUND",
			"details": err.Error(),
		})
	}

	// Extract features from the original project
	var originalFeatures []string
	if err := json.Unmarshal(originalProject.Features, &originalFeatures); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to parse original project features", map[string]string{
			"details": err.Error(),
		})
	}

	var adjustedFeatures []string
	if err := json.Unmarshal(originalProject.AdjustedFeatures, &adjustedFeatures); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to parse adjusted project features", map[string]string{
			"details": err.Error(),
		})
	}

	// Validate that we have the necessary data for duplication
	if len(originalFeatures) == 0 || originalProject.Framework == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Cannot duplicate project: insufficient configuration data", map[string]string{
			"code": "INSUFFICIENT_DATA",
			"details": "Original project lacks necessary configuration information",
		})
	}

	// Validate new project name
	if strings.TrimSpace(req.NewProjectName) == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "New project name is required")
	}

	// Sanitize the new project name
	sanitizedProjectName := sanitizeProjectName(req.NewProjectName)
	if sanitizedProjectName == "" {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid project name format")
	}

	// Check if the new project name is different from the original
	if sanitizedProjectName == originalProject.ProjectName {
		// Suggest a new name by appending a suffix
		timestamp := time.Now().Format("20060102_150405")
		sanitizedProjectName = fmt.Sprintf("%s_copy_%s", originalProject.ProjectName, timestamp)
	}

	// Log duplication activity
	fmt.Printf("User %d duplicating project %d (%s) as '%s'\n", 
		userIDUint, originalProject.ID, originalProject.ProjectName, sanitizedProjectName)

	// Prepare the response with configuration data for pre-populating the form
	response := map[string]interface{}{
		"success": true,
		"message": "Project configuration duplicated successfully",
		"duplicate_config": map[string]interface{}{
			"project_name":       sanitizedProjectName,
			"framework":          originalProject.Framework,
			"features":           originalFeatures,
			"adjusted_features":  adjustedFeatures,
			"original_project": map[string]interface{}{
				"id":           originalProject.ID,
				"name":         originalProject.ProjectName,
				"created_at":   originalProject.CreatedAt,
			},
		},
		"form_data": map[string]interface{}{
			"projectName": sanitizedProjectName,
			"framework":   originalProject.Framework,
			"features":    originalFeatures,
		},
		"suggestions": map[string]interface{}{
			"alternative_names": generateAlternativeNames(originalProject.ProjectName),
		},
	}

	// Log successful duplication
	middleware.LogProjectAccess(c, userIDUint, req.OriginalProjectID, "duplicate_project", true, "")

	return c.JSON(response)
}

// GetProjectStats godoc
// @Summary Get user's project statistics
// @Description Retrieves comprehensive statistics about user's project generation patterns including framework usage, most used features, and recent activity
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.ProjectStatsResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/stats [get]
func GetProjectStats(c *fiber.Ctx) error {
	// Get validated user ID from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "get_project_stats", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get user statistics
	stats, err := historyService.GetProjectStats(userIDUint)
	if err != nil {
		middleware.LogUserAction(c, userIDUint, "get_project_stats", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve project statistics", map[string]string{
			"code": "STATS_RETRIEVAL_FAILED",
			"details": err.Error(),
		})
	}

	// Log successful stats retrieval
	middleware.LogUserAction(c, userIDUint, "get_project_stats", "project_history", true, "")

	return c.JSON(stats)
}

// GetDashboardData godoc
// @Summary Get optimized dashboard data
// @Description Retrieves comprehensive dashboard data with caching and performance optimizations for large datasets
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/history/dashboard [get]
func GetDashboardData(c *fiber.Ctx) error {
	// Get validated user ID from middleware
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "get_dashboard_data", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get optimized dashboard data
	dashboardData, err := historyService.GetDashboardData(userIDUint)
	if err != nil {
		middleware.LogUserAction(c, userIDUint, "get_dashboard_data", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve dashboard data", map[string]string{
			"code": "DASHBOARD_RETRIEVAL_FAILED",
			"details": err.Error(),
		})
	}

	// Log successful dashboard data retrieval
	middleware.LogUserAction(c, userIDUint, "get_dashboard_data", "project_history", true, "")

	return c.JSON(dashboardData)
}

// GetCacheInfo godoc
// @Summary Get cache performance information
// @Description Retrieves information about the statistics cache performance and status
// @Tags History
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Router /api/history/cache-info [get]
func GetCacheInfo(c *fiber.Ctx) error {
	// Get validated user ID from middleware (for authentication)
	userIDUint, err := middleware.GetValidatedUserID(c)
	if err != nil {
		middleware.LogUserAction(c, 0, "get_cache_info", "project_history", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User authentication failed", map[string]string{
			"code": "AUTH_FAILED",
			"details": err.Error(),
		})
	}

	// Create service instance
	historyService := services.NewProjectHistoryService(database.DB)

	// Get cache statistics
	cacheStats := historyService.GetCacheStats()

	// Log successful cache info retrieval
	middleware.LogUserAction(c, userIDUint, "get_cache_info", "project_history", true, "")

	return c.JSON(map[string]interface{}{
		"cache_stats": cacheStats,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// sanitizeProjectName sanitizes project name for safe usage
func sanitizeProjectName(projectName string) string {
	// Remove leading/trailing spaces
	sanitized := strings.TrimSpace(projectName)
	
	// Replace spaces and special characters with hyphens
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	
	// Remove or replace dangerous characters
	dangerous := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t", "@", "#", "$", "%", "^", "&", "(", ")", "+", "=", "[", "]", "{", "}", ";", "'", ","}
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	// Remove multiple consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}
	
	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")
	
	// Ensure name is not empty and within length limits
	if sanitized == "" {
		sanitized = "project"
	}
	
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
		sanitized = strings.Trim(sanitized, "-")
	}
	
	return sanitized
}

// generateAlternativeNames generates alternative project names
func generateAlternativeNames(originalName string) []string {
	alternatives := []string{}
	baseName := sanitizeProjectName(originalName)
	
	// Generate timestamp-based alternatives
	timestamp := time.Now().Format("20060102")
	shortTimestamp := time.Now().Format("0102")
	
	alternatives = append(alternatives, fmt.Sprintf("%s-copy", baseName))
	alternatives = append(alternatives, fmt.Sprintf("%s-v2", baseName))
	alternatives = append(alternatives, fmt.Sprintf("%s-%s", baseName, timestamp))
	alternatives = append(alternatives, fmt.Sprintf("%s-copy-%s", baseName, shortTimestamp))
	alternatives = append(alternatives, fmt.Sprintf("new-%s", baseName))
	
	// Remove duplicates and filter out names that are too long
	seen := make(map[string]bool)
	var filtered []string
	for _, name := range alternatives {
		if len(name) <= 50 && !seen[name] {
			seen[name] = true
			filtered = append(filtered, name)
		}
	}
	
	return filtered
}

// sanitizeDownloadFilename sanitizes filename for safe download
func sanitizeDownloadFilename(filename string) string {
	// Remove or replace dangerous characters
	dangerous := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	sanitized := filename
	
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "project.zip"
	}
	
	// Limit filename length
	if len(sanitized) > 100 {
		ext := ".zip"
		name := strings.TrimSuffix(sanitized, ext)
		if len(name) > 96 {
			name = name[:96]
		}
		sanitized = name + ext
	}
	
	return sanitized
}

// convertProjectToDetailedResponse converts a ProjectHistory model to a detailed response format
func convertProjectToDetailedResponse(project *models.ProjectHistory, fileService *services.FileService) (*models.ProjectHistoryResponse, error) {
	// Unmarshal features
	var features []string
	if err := json.Unmarshal(project.Features, &features); err != nil {
		return nil, err
	}

	var adjustedFeatures []string
	if err := json.Unmarshal(project.AdjustedFeatures, &adjustedFeatures); err != nil {
		return nil, err
	}

	// Check if file is available for download
	canDownload := false
	canRegenerate := false
	
	if project.ZipFileStatus == models.ZipFileStatusAvailable && project.ZipFilePath != nil {
		canDownload = fileService.FileExists(*project.ZipFilePath)
		
		// Update status if file doesn't exist but status says it's available
		if !canDownload && project.ZipFileStatus == models.ZipFileStatusAvailable {
			// Note: In a production system, you might want to update the database status here
			// For now, we'll just return the correct status in the response
		}
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