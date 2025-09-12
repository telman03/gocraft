package handlers

import (
	"fmt"

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

	// Validate template conflicts
	validator := validation.NewTemplateValidator()
	validationResult := validator.ValidateFeatures(req.Features)
	
	if !validationResult.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Feature validation failed",
			"validation_result": validationResult,
			"message": "Please resolve the conflicts and try again",
		})
	}

	// Use adjusted features (with dependencies added)
	adjustedFeatures := validationResult.AdjustedFeatures

	// Optionally log userID or track usage
	fmt.Printf("Generating project '%s' for user: %v with features: %v\n", req.ProjectName, userID, adjustedFeatures)

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
