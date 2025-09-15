package middleware

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/services"
)

// HistoryTrackingMiddleware captures generation requests and records them in history
func HistoryTrackingMiddleware(historyService *services.ProjectHistoryService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only track POST requests to generation endpoints
		if c.Method() != "POST" || !strings.Contains(c.Path(), "/generate") {
			return c.Next()
		}

		// Skip if this is not the main generate endpoint
		if c.Path() != "/generate/" && c.Path() != "/generate" {
			return c.Next()
		}

		// Get user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			return c.Next() // Let auth middleware handle this
		}

		userIDUint, ok := userID.(uint)
		if !ok {
			return c.Next() // Invalid user ID format
		}

		// Read and preserve the request body
		bodyBytes := c.Body()
		if len(bodyBytes) == 0 {
			return c.Next() // Continue without tracking if no body
		}

		// Parse the generation request
		var genRequest models.GenerateRequest
		if err := json.Unmarshal(bodyBytes, &genRequest); err != nil {
			return c.Next() // Continue without tracking on parse error
		}

		// Record start time for duration measurement
		startTime := time.Now()

		// Store request data in context for post-processing
		c.Locals("history_tracking", map[string]interface{}{
			"user_id":      userIDUint,
			"request":      genRequest,
			"start_time":   startTime,
			"should_track": true,
		})

		// Continue to the actual handler
		handlerErr := c.Next()

		// Post-process: Record history if generation was successful
		go func() {
			if err := recordHistoryAsync(c, historyService, handlerErr); err != nil {
				fmt.Printf("Warning: Failed to record project history: %v\n", err)
			}
		}()

		return handlerErr
	}
}

// recordHistoryAsync handles the asynchronous recording of project history
func recordHistoryAsync(c *fiber.Ctx, historyService *services.ProjectHistoryService, handlerErr error) error {
	// Get tracking data from context
	trackingData := c.Locals("history_tracking")
	if trackingData == nil {
		return nil // No tracking data available
	}

	trackingMap, ok := trackingData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid tracking data format")
	}

	shouldTrack, exists := trackingMap["should_track"].(bool)
	if !exists || !shouldTrack {
		return nil // Tracking disabled
	}

	// Only record history if the generation was successful (no error and status 200)
	if handlerErr != nil || c.Response().StatusCode() != 200 {
		return nil // Don't record failed generations
	}

	// Extract tracking data
	userID, ok := trackingMap["user_id"].(uint)
	if !ok {
		return fmt.Errorf("invalid user ID in tracking data")
	}

	genRequest, ok := trackingMap["request"].(models.GenerateRequest)
	if !ok {
		return fmt.Errorf("invalid request data in tracking data")
	}

	startTime, ok := trackingMap["start_time"].(time.Time)
	if !ok {
		return fmt.Errorf("invalid start time in tracking data")
	}

	// Calculate generation duration
	duration := time.Since(startTime)
	durationMs := int(duration.Milliseconds())

	// Determine framework from request
	framework := genRequest.Framework
	if framework == "" && len(genRequest.Features) > 0 {
		// Try to extract framework from features
		for _, feature := range genRequest.Features {
			lowerFeature := strings.ToLower(feature)
			if lowerFeature == "gin" || lowerFeature == "echo" || lowerFeature == "fiber" {
				framework = lowerFeature
				break
			}
		}
	}

	// Merge framework into features if provided separately
	allFeatures := genRequest.Features
	if framework != "" {
		// Check if framework is already in features
		frameworkExists := false
		for _, feature := range genRequest.Features {
			if strings.ToLower(feature) == strings.ToLower(framework) {
				frameworkExists = true
				break
			}
		}
		// Add framework to features if not already present
		if !frameworkExists {
			allFeatures = append([]string{framework}, genRequest.Features...)
		}
	}

	// For now, use the same features for adjusted features
	// In a real implementation, this would come from the validation result
	adjustedFeatures := allFeatures

	// Try to determine the ZIP file path and size
	// This is a best-effort attempt based on common patterns
	var zipFilePath *string
	var zipFileSize *int64

	// Try to get the file path from the response or generate expected path
	expectedZipPath := fmt.Sprintf("output/%s.zip", genRequest.ProjectName)
	if fileInfo, err := os.Stat(expectedZipPath); err == nil {
		cleanPath := filepath.Clean(expectedZipPath)
		zipFilePath = &cleanPath
		size := fileInfo.Size()
		zipFileSize = &size
	}

	// Create the history record
	createRequest := models.CreateProjectRecordRequest{
		UserID:               userID,
		ProjectName:          genRequest.ProjectName,
		Framework:            framework,
		Features:             allFeatures,
		AdjustedFeatures:     adjustedFeatures,
		ZipFilePath:          zipFilePath,
		ZipFileSize:          zipFileSize,
		GenerationDurationMs: &durationMs,
	}

	// Record the project in history
	_, err := historyService.CreateProjectRecord(createRequest)
	if err != nil {
		return fmt.Errorf("failed to create project history record: %w", err)
	}

	fmt.Printf("Successfully recorded project '%s' in history for user %d (duration: %dms)\n", 
		genRequest.ProjectName, userID, durationMs)

	return nil
}