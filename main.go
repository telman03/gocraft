package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/database"

	fiberSwagger "github.com/swaggo/fiber-swagger"
	_ "github.com/telman03/ai-backend-generator/docs"
	"github.com/telman03/ai-backend-generator/internal/api"
)

// @title GoCraft API
// @version 1.0
// @description Backend generator microservice
// @security BearerAuth
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token in the format: Bearer <token>
func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.InitDB()

	app := fiber.New()

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,POST,OPTIONS",
		ExposeHeaders: "Content-Disposition",
		AllowHeaders:  "Content-Type,Authorization",
	}))

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Debug page for testing downloads
	app.Static("/debug", "./debug_download.html")

	api.SetupRoutes(app)

	log.Fatal(app.Listen(":8081"))
}
