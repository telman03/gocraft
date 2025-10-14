package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/builder"
	"github.com/telman03/gocraft-backend/internal/models"
	"github.com/telman03/gocraft-backend/internal/utils"
	"github.com/telman03/gocraft-backend/internal/validation"
)

// Generate GenerateHandler godoc
// @Summary Generate Go backend project
// @Description Accepts selected features and returns a downloadable zip file
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Security BearerAuth
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /generate [post]
func Generate(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Merge framework into features if provided separately
	allFeatures := req.Features
	if req.Framework != "" {
		// Check if framework is already in features
		frameworkExists := false
		for _, feature := range req.Features {
			if strings.EqualFold(feature, req.Framework) {
				frameworkExists = true
				break
			}
		}
		// Add framework to features if not already present
		if !frameworkExists {
			allFeatures = append([]string{req.Framework}, req.Features...)
		}
	}

	// Validate template conflicts
	validator := validation.NewTemplateValidator()
	validationResult := validator.ValidateFeatures(allFeatures)
	
	if !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Feature validation failed",
			"validation_result": validationResult,
			"message": "Please resolve the conflicts and try again",
		})
	}

	// Use adjusted features (with dependencies added)
	adjustedFeatures := validationResult.AdjustedFeatures

	// Generate unique request ID for tracking
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	// Log request with unique ID to help debug duplicate downloads
	fmt.Printf("[REQ:%s] Generating project '%s' for user: %v\n", requestID, req.ProjectName, userID)
	fmt.Printf("[REQ:%s] Original features: %v, Framework: %s\n", requestID, req.Features, req.Framework)
	fmt.Printf("[REQ:%s] Merged features: %v\n", requestID, allFeatures)
	fmt.Printf("[REQ:%s] Final adjusted features: %v\n", requestID, adjustedFeatures)

	zipPath, err := builder.GenerateProject(req.ProjectName, adjustedFeatures)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}

	// Set proper headers for download
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", req.ProjectName))
	c.Set("Content-Type", "application/zip")
	
	return c.Download(zipPath)
}

// GenerateWithOptionalAuth handles project generation with optional authentication
// If user is authenticated, it tracks the project in history. If not, it just generates the project.
// @Summary Generate Go backend project (with optional auth)
// @Description Accepts selected features and returns a downloadable zip file. Tracks history for authenticated users.
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /api/v1/generate [post]
func GenerateWithOptionalAuth(c *fiber.Ctx) error {
	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Check if user is authenticated (set by OptionalAuth middleware)
	userID := c.Locals("user_id")
	isAuthenticated, _ := c.Locals("is_authenticated").(bool)

	// Merge framework into features if provided separately
	allFeatures := req.Features
	if req.Framework != "" {
		// Check if framework is already in features
		frameworkExists := false
		for _, feature := range req.Features {
			if strings.EqualFold(feature, req.Framework) {
				frameworkExists = true
				break
			}
		}
		// Add framework to features if not already present
		if !frameworkExists {
			allFeatures = append([]string{req.Framework}, req.Features...)
		}
	}

	// Validate template conflicts
	validator := validation.NewTemplateValidator()
	validationResult := validator.ValidateFeatures(allFeatures)
	
	if !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Feature validation failed",
			"validation_result": validationResult,
			"message": "Please resolve the conflicts and try again",
		})
	}

	// Use adjusted features (with dependencies added)
	adjustedFeatures := validationResult.AdjustedFeatures

	// Generate unique request ID for tracking
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	// Log request with appropriate context
	if isAuthenticated {
		fmt.Printf("[AUTH-REQ:%s] Generating project '%s' for user: %v\n", requestID, req.ProjectName, userID)
	} else {
		fmt.Printf("[GUEST-REQ:%s] Generating project '%s' for guest user\n", requestID, req.ProjectName)
	}
	fmt.Printf("[REQ:%s] Original features: %v, Framework: %s\n", requestID, req.Features, req.Framework)
	fmt.Printf("[REQ:%s] Merged features: %v\n", requestID, allFeatures)
	fmt.Printf("[REQ:%s] Final adjusted features: %v\n", requestID, adjustedFeatures)

	zipPath, err := builder.GenerateProject(req.ProjectName, adjustedFeatures)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}

	// Set proper headers for download
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", req.ProjectName))
	c.Set("Content-Type", "application/zip")
	
	return c.Download(zipPath)
}

// GeneratePublic handles project generation for guest users (no authentication required)
// @Summary Generate Go backend project (public)
// @Description Accepts selected features and returns a downloadable zip file for guest users
// @Tags Generator
// @Accept json
// @Produce application/zip
// @Param data body models.GenerateRequest true "Selected Features"
// @Router /api/v1/generate [post]
func GeneratePublic(c *fiber.Ctx) error {
	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Merge framework into features if provided separately
	allFeatures := req.Features
	if req.Framework != "" {
		// Check if framework is already in features
		frameworkExists := false
		for _, feature := range req.Features {
			if strings.EqualFold(feature, req.Framework) {
				frameworkExists = true
				break
			}
		}
		// Add framework to features if not already present
		if !frameworkExists {
			allFeatures = append([]string{req.Framework}, req.Features...)
		}
	}

	// Validate template conflicts
	validator := validation.NewTemplateValidator()
	validationResult := validator.ValidateFeatures(allFeatures)
	
	if !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Feature validation failed",
			"validation_result": validationResult,
			"message": "Please resolve the conflicts and try again",
		})
	}

	// Use adjusted features (with dependencies added)
	adjustedFeatures := validationResult.AdjustedFeatures

	// Generate unique request ID for tracking
	requestID := c.Get("X-Request-ID")
	if requestID == "" {
		requestID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	// Log request with unique ID (no user ID for public requests)
	fmt.Printf("[PUBLIC-REQ:%s] Generating project '%s' for guest user\n", requestID, req.ProjectName)
	fmt.Printf("[PUBLIC-REQ:%s] Original features: %v, Framework: %s\n", requestID, req.Features, req.Framework)
	fmt.Printf("[PUBLIC-REQ:%s] Merged features: %v\n", requestID, allFeatures)
	fmt.Printf("[PUBLIC-REQ:%s] Final adjusted features: %v\n", requestID, adjustedFeatures)

	zipPath, err := builder.GenerateProject(req.ProjectName, adjustedFeatures)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate project", map[string]string{
			"details": err.Error(),
		})
	}

	// Set proper headers for download
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", req.ProjectName))
	c.Set("Content-Type", "application/zip")
	
	return c.Download(zipPath)
}
