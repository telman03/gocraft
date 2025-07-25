package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/database"

	"github.com/telman03/ai-backend-generator/internal/api"
	_ "github.com/telman03/ai-backend-generator/docs" 
	"github.com/swaggo/fiber-swagger"                 
)

// @title GoCraft API
// @version 1.0
// @description Backend generator microservice
// @host localhost:8080
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



	api.SetupRoutes(app)


	log.Fatal(app.Listen(":8080"))
}