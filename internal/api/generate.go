package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/builder"
)

type GenerateRequest struct {
	Features []string `json:"features"`
}

func GenerateHandler(c *fiber.Ctx) error {
	var req GenerateRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	zipPath, err := builder.GenerateProject(req.Features)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Download(zipPath)
}