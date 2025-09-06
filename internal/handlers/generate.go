package handlers

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/builder"
	"github.com/telman03/ai-backend-generator/internal/models"
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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request format"})
	}

	// Optionally log userID or track usage
	fmt.Printf("Generating project for user: %v\n", userID)

	zipPath, err := builder.GenerateProject(req.Features)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Download(zipPath)
}
