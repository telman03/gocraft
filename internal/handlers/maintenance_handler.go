package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/services"
)

// MaintenanceHandler handles maintenance-related API endpoints
type MaintenanceHandler struct {
	dbMaintenanceService   *services.DatabaseMaintenanceService
	fileCleanupService     *services.FileCleanupService
}

// NewMaintenanceHandler creates a new maintenance handler
func NewMaintenanceHandler(dbMaintenanceService *services.DatabaseMaintenanceService, fileCleanupService *services.FileCleanupService) *MaintenanceHandler {
	return &MaintenanceHandler{
		dbMaintenanceService: dbMaintenanceService,
		fileCleanupService:   fileCleanupService,
	}
}

// GetDatabaseHealth godoc
// @Summary Get database health status
// @Description Returns comprehensive database health metrics and recommendations
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.DatabaseHealthReport
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/health [get]
func (h *MaintenanceHandler) GetDatabaseHealth(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	health, err := h.dbMaintenanceService.GetDatabaseHealth(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get database health",
			"details": err.Error(),
		})
	}

	return c.JSON(health)
}

// GetPerformanceMetrics godoc
// @Summary Get database performance metrics
// @Description Returns detailed database performance metrics
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.PerformanceMetrics
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/performance [get]
func (h *MaintenanceHandler) GetPerformanceMetrics(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	metrics, err := h.dbMaintenanceService.GetPerformanceMetrics(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get performance metrics",
			"details": err.Error(),
		})
	}

	return c.JSON(metrics)
}

// RunDatabaseMaintenance godoc
// @Summary Run database maintenance
// @Description Manually trigger database maintenance operations
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.MaintenanceStats
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/database/run [post]
func (h *MaintenanceHandler) RunDatabaseMaintenance(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stats, err := h.dbMaintenanceService.RunMaintenance(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to run database maintenance",
			"details": err.Error(),
		})
	}

	return c.JSON(stats)
}

// RunFileCleanup godoc
// @Summary Run file cleanup
// @Description Manually trigger file cleanup operations
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.CleanupStats
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/files/run [post]
func (h *MaintenanceHandler) RunFileCleanup(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	stats, err := h.fileCleanupService.RunCleanup(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to run file cleanup",
			"details": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetMaintenanceStatus godoc
// @Summary Get maintenance service status
// @Description Returns the status of maintenance services
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/status [get]
func (h *MaintenanceHandler) GetMaintenanceStatus(c *fiber.Ctx) error {
	dbConfig := h.dbMaintenanceService.GetMaintenanceConfig()
	fileConfig := h.fileCleanupService.GetCleanupConfig()

	status := fiber.Map{
		"database_maintenance": fiber.Map{
			"running": h.dbMaintenanceService.IsRunning(),
			"config":  dbConfig,
		},
		"file_cleanup": fiber.Map{
			"running": h.fileCleanupService.IsRunning(),
			"config":  fileConfig,
		},
	}

	return c.JSON(status)
}

// UpdateMaintenanceConfig godoc
// @Summary Update maintenance configuration
// @Description Updates the configuration for maintenance services
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param config body MaintenanceConfigRequest true "Maintenance configuration"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/config [put]
func (h *MaintenanceHandler) UpdateMaintenanceConfig(c *fiber.Ctx) error {
	var req MaintenanceConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Update database maintenance config
	if req.DatabaseMaintenance != nil {
		dbConfig := services.MaintenanceConfig{
			MaintenanceInterval: time.Duration(req.DatabaseMaintenance.MaintenanceIntervalHours) * time.Hour,
			ArchivalThreshold:   time.Duration(req.DatabaseMaintenance.ArchivalThresholdDays) * 24 * time.Hour,
			CleanupBatchSize:    req.DatabaseMaintenance.CleanupBatchSize,
		}
		
		if err := h.dbMaintenanceService.UpdateMaintenanceConfig(dbConfig); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update database maintenance config",
				"details": err.Error(),
			})
		}
	}

	// Update file cleanup config
	if req.FileCleanup != nil {
		fileConfig := services.CleanupConfig{
			CleanupInterval:  time.Duration(req.FileCleanup.CleanupIntervalHours) * time.Hour,
			RetentionPeriod:  time.Duration(req.FileCleanup.RetentionPeriodDays) * 24 * time.Hour,
			BatchSize:        req.FileCleanup.BatchSize,
			MaxConcurrency:   req.FileCleanup.MaxConcurrency,
		}
		
		if err := h.fileCleanupService.UpdateCleanupConfig(fileConfig); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update file cleanup config",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"message": "Configuration updated successfully",
	})
}

// GetStorageStats godoc
// @Summary Get storage statistics
// @Description Returns file storage usage statistics
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.StorageStats
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/storage [get]
func (h *MaintenanceHandler) GetStorageStats(c *fiber.Ctx) error {
	stats, err := h.fileCleanupService.GetStorageStats()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get storage statistics",
			"details": err.Error(),
		})
	}

	return c.JSON(stats)
}

// ValidateFileIntegrity godoc
// @Summary Validate file integrity
// @Description Checks if files referenced in database actually exist
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.IntegrityReport
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/integrity [post]
func (h *MaintenanceHandler) ValidateFileIntegrity(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	report, err := h.fileCleanupService.ValidateFileIntegrity(ctx)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to validate file integrity",
			"details": err.Error(),
		})
	}

	return c.JSON(report)
}

// StartMaintenanceServices godoc
// @Summary Start maintenance services
// @Description Starts the automatic maintenance services
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/start [post]
func (h *MaintenanceHandler) StartMaintenanceServices(c *fiber.Ctx) error {
	ctx := context.Background()

	// Start database maintenance service
	if err := h.dbMaintenanceService.Start(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start database maintenance service",
			"details": err.Error(),
		})
	}

	// Start file cleanup service
	if err := h.fileCleanupService.Start(ctx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to start file cleanup service",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Maintenance services started successfully",
	})
}

// StopMaintenanceServices godoc
// @Summary Stop maintenance services
// @Description Stops the automatic maintenance services
// @Tags maintenance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/maintenance/stop [post]
func (h *MaintenanceHandler) StopMaintenanceServices(c *fiber.Ctx) error {
	// Stop database maintenance service
	if err := h.dbMaintenanceService.Stop(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stop database maintenance service",
			"details": err.Error(),
		})
	}

	// Stop file cleanup service
	if err := h.fileCleanupService.Stop(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to stop file cleanup service",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Maintenance services stopped successfully",
	})
}

// MaintenanceConfigRequest represents the request body for updating maintenance configuration
type MaintenanceConfigRequest struct {
	DatabaseMaintenance *DatabaseMaintenanceConfigRequest `json:"database_maintenance,omitempty"`
	FileCleanup         *FileCleanupConfigRequest         `json:"file_cleanup,omitempty"`
}

// DatabaseMaintenanceConfigRequest represents database maintenance configuration
type DatabaseMaintenanceConfigRequest struct {
	MaintenanceIntervalHours int `json:"maintenance_interval_hours" validate:"min=1,max=168"`
	ArchivalThresholdDays    int `json:"archival_threshold_days" validate:"min=1,max=365"`
	CleanupBatchSize         int `json:"cleanup_batch_size" validate:"min=100,max=10000"`
}

// FileCleanupConfigRequest represents file cleanup configuration
type FileCleanupConfigRequest struct {
	CleanupIntervalHours int `json:"cleanup_interval_hours" validate:"min=1,max=168"`
	RetentionPeriodDays  int `json:"retention_period_days" validate:"min=1,max=365"`
	BatchSize            int `json:"batch_size" validate:"min=10,max=1000"`
	MaxConcurrency       int `json:"max_concurrency" validate:"min=1,max=20"`
}