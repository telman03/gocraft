package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"github.com/telman03/ai-backend-generator/internal/api"
	"github.com/telman03/ai-backend-generator/internal/config"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/services"

	fiberSwagger "github.com/swaggo/fiber-swagger"
	_ "github.com/telman03/ai-backend-generator/docs"
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

	// Load maintenance configuration
	maintenanceConfig := config.LoadMaintenanceConfig()

	// Initialize services
	historyService := services.NewProjectHistoryService(database.DB)
	
	// Initialize file service
	fileService := services.NewFileService(maintenanceConfig.StorageBasePath, maintenanceConfig.FileRetentionPeriod)
	
	// Initialize maintenance services
	fileCleanupConfig := services.CleanupConfig{
		CleanupInterval:  maintenanceConfig.FileCleanupInterval,
		BatchSize:        maintenanceConfig.FileCleanupBatchSize,
		MaxConcurrency:   maintenanceConfig.FileMaxConcurrency,
		RetentionPeriod:  maintenanceConfig.FileRetentionPeriod,
		EnableScheduling: maintenanceConfig.FileCleanupEnabled,
	}
	fileCleanupService := services.NewFileCleanupService(database.DB, fileService, fileCleanupConfig)
	
	dbMaintenanceConfig := services.MaintenanceConfig{
		MaintenanceInterval: maintenanceConfig.DBMaintenanceInterval,
		ArchivalThreshold:   maintenanceConfig.DBArchivalThreshold,
		CleanupBatchSize:    maintenanceConfig.DBCleanupBatchSize,
		EnableScheduling:    maintenanceConfig.DBMaintenanceEnabled,
	}
	dbMaintenanceService := services.NewDatabaseMaintenanceService(database.DB, dbMaintenanceConfig)

	// Start maintenance services if enabled
	ctx := context.Background()
	if maintenanceConfig.FileCleanupEnabled {
		if err := fileCleanupService.Start(ctx); err != nil {
			log.Printf("Warning: Failed to start file cleanup service: %v", err)
		} else {
			log.Println("File cleanup service started")
		}
	}
	
	if maintenanceConfig.DBMaintenanceEnabled {
		if err := dbMaintenanceService.Start(ctx); err != nil {
			log.Printf("Warning: Failed to start database maintenance service: %v", err)
		} else {
			log.Println("Database maintenance service started")
		}
	}

	app := fiber.New()

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:  "*",
		AllowMethods:  "GET,POST,PUT,DELETE,OPTIONS",
		ExposeHeaders: "Content-Disposition",
		AllowHeaders:  "Content-Type,Authorization",
	}))

	// Swagger route
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	// Debug page for testing downloads
	app.Static("/debug", "./debug_download.html")

	api.SetupRoutes(app, historyService, dbMaintenanceService, fileCleanupService)

	log.Fatal(app.Listen(":8081"))
}
