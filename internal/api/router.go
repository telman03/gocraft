package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/handlers"
	"github.com/telman03/ai-backend-generator/internal/middleware"
	"github.com/telman03/ai-backend-generator/internal/services"
)

func SetupRoutes(app *fiber.App, historyService *services.ProjectHistoryService, dbMaintenanceService *services.DatabaseMaintenanceService, fileCleanupService *services.FileCleanupService) {
	// Create input sanitizer and rate limiter
	sanitizer := middleware.NewInputSanitizer()
	rateLimiter := middleware.NewRateLimiter(100, time.Minute) // 100 requests per minute

	// Apply global middleware
	app.Use(sanitizer.ValidateAndSanitizeQueryParams())
	app.Use(sanitizer.ValidateJSONBody())
	app.Use(rateLimiter.RateLimit())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "Welcome to GoCraft API 🚀"})
	})

	// Health check endpoints for deployment systems
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "gocraft-api",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		// Check database connection
		sqlDB, err := database.DB.DB()
		if err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status": "not ready",
				"error": "database connection failed",
			})
		}

		if err := sqlDB.Ping(); err != nil {
			return c.Status(503).JSON(fiber.Map{
				"status": "not ready", 
				"error": "database ping failed",
			})
		}

		return c.JSON(fiber.Map{
			"status": "ready",
			"service": "gocraft-api",
			"database": "connected",
			"timestamp": time.Now().Format(time.RFC3339),
		})
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
	
	// Debug endpoints (for troubleshooting)
	debug := app.Group("/debug")
	debug.Post("/request", handlers.DebugRequest)
	debug.Post("/download", handlers.DebugDownload)
	
	// Project generation and validation
	secure := app.Group("/generate", middleware.RequireAuth, middleware.HistoryTrackingMiddleware(historyService))
	secure.Post("/", handlers.Generate)
	secure.Post("/verify", handlers.VerifyGeneration)
	secure.Post("/validate", handlers.ValidateFeatures)

	// Project history endpoints
	api := app.Group("/api", middleware.RequireAuth)
	
	// General history endpoints (user context validation only)
	api.Get("/history", middleware.UserContextValidator(), sanitizer.ValidateHistoryFilters(), handlers.GetProjectHistory)
	api.Get("/history/stats", middleware.UserContextValidator(), handlers.GetProjectStats)
	api.Get("/history/dashboard", middleware.UserContextValidator(), handlers.GetDashboardData)
	api.Get("/history/cache-info", middleware.UserContextValidator(), handlers.GetCacheInfo)
	api.Post("/history/duplicate", middleware.UserContextValidator(), handlers.DuplicateProject)
	
	// Project-specific endpoints (ownership validation required)
	api.Get("/history/:id", sanitizer.ValidateProjectID(), middleware.ProjectOwnershipValidator(), handlers.GetProjectDetails)
	api.Delete("/history/:id", sanitizer.ValidateProjectID(), middleware.ProjectOwnershipValidator(), handlers.DeleteProject)
	api.Get("/history/:id/download", sanitizer.ValidateProjectID(), middleware.ProjectOwnershipValidator(), handlers.DownloadProject)
	api.Post("/history/:id/regenerate", sanitizer.ValidateProjectID(), middleware.ProjectOwnershipValidator(), handlers.RegenerateProject)

	// Admin endpoints (admin access required)
	admin := api.Group("/admin", middleware.RequireAuth, middleware.RequireAdmin)
	admin.Get("/users", handlers.GetAllUsers)
	admin.Put("/users/:id/role", handlers.UpdateUserRole)
	admin.Get("/stats", handlers.GetUserStats)

	// Maintenance endpoints (admin access required)
	maintenanceHandler := handlers.NewMaintenanceHandler(dbMaintenanceService, fileCleanupService)
	maintenance := api.Group("/maintenance", middleware.RequireAuth, middleware.RequireAdmin)
	maintenance.Get("/health", maintenanceHandler.GetDatabaseHealth)
	maintenance.Get("/performance", maintenanceHandler.GetPerformanceMetrics)
	maintenance.Get("/status", maintenanceHandler.GetMaintenanceStatus)
	maintenance.Get("/storage", maintenanceHandler.GetStorageStats)
	maintenance.Post("/database/run", maintenanceHandler.RunDatabaseMaintenance)
	maintenance.Post("/files/run", maintenanceHandler.RunFileCleanup)
	maintenance.Post("/integrity", maintenanceHandler.ValidateFileIntegrity)
	maintenance.Post("/start", maintenanceHandler.StartMaintenanceServices)
	maintenance.Post("/stop", maintenanceHandler.StopMaintenanceServices)
	maintenance.Put("/config", maintenanceHandler.UpdateMaintenanceConfig)
}
