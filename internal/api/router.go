package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/handlers"
	"github.com/telman03/ai-backend-generator/internal/middleware"
)

func SetupRoutes(app *fiber.App) {
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome to GoCraft API 🚀"})
	})

	// Auth routes
	auth := app.Group("/auth")
	auth.Post("/register", handlers.Register) // You’ll build this
	auth.Post("/login", handlers.Login)       // You’ll build this too

	// Protected route
	secure := app.Group("/generate", middleware.RequireAuth)
	secure.Post("/", handlers.Generate)
}