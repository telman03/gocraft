package main

import (

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/api"
	"github.com/telman03/ai-backend-generator/internal/middleware"
)

func main() {
		// ✅ Load .env first
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	

	app := fiber.New()
	api.InitSupabase()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,OPTIONS",
		ExposeHeaders: "Content-Disposition",
		AllowHeaders:  "Content-Type",
	}))

	app.Post("/generate", middleware.RequireAuth, api.GenerateHandler)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Listen(":8080")
}