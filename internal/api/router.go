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
	auth.Post("/register", handlers.Register)
	auth.Post("/verify-otp", handlers.VerifyOTP)
	auth.Post("/resend-otp", handlers.ResendOTP)
	auth.Post("/login", handlers.Login)
	auth.Get("/me", middleware.RequireAuth, handlers.GetCurrentUser)

	// Feature validation (public endpoint)
	app.Get("/features", handlers.GetSupportedFeatures)
	
	// Project generation and validation
	secure := app.Group("/generate", middleware.RequireAuth)
	secure.Post("/", handlers.Generate)
	secure.Post("/verify", handlers.VerifyGeneration)
	secure.Post("/validate", handlers.ValidateFeatures)
}
