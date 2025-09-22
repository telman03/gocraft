package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/builder"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/utils"
	"github.com/telman03/ai-backend-generator/internal/validation"
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
