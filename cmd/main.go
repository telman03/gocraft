package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/handlers"
	"github.com/telman03/ai-backend-generator/internal/middleware"

	_ "github.com/telman03/ai-backend-generator/docs" // 👈 required for swag
	"github.com/swaggo/fiber-swagger"                 // fiber-swagger middleware
)

// @title GoCraft API
// @version 1.0
// @description Backend generator microservice
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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



	app.Post("/auth/login", handlers.Login)
	app.Post("/auth/register", handlers.Register)

	app.Post("/generate", middleware.RequireAuth, handlers.Generate)



	log.Fatal(app.Listen(":8080"))
}