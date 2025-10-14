package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/errors"
	"github.com/telman03/gocraft-backend/internal/logging"
)

// HistoryErrorHandler provides error handling utilities for history handlers
type HistoryErrorHandler struct {
	logger *logging.Logger
}

// NewHistoryErrorHandler creates a new history error handler
func NewHistoryErrorHandler() *HistoryErrorHandler {
	return &HistoryErrorHandler{
		logger: logging.GetLogger().WithComponent("history_handler"),
	}
}

// HandleAuthError handles authentication-related errors
func (h *HistoryErrorHandler) HandleAuthError(c *fiber.Ctx, err error, operation string) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewAuthError("User authentication failed").
		WithDetails(err.Error()).
		WithRequestID(c.Get("X-Request-ID", ""))

	// Log the authentication failure
	requestLogger.LogUserAction(operation, "project_history", false, err.Error())
	
	// Log as security event
	requestLogger.LogSecurityEvent(
		"AUTH_FAILURE",
		"high",
		fmt.Sprintf("Authentication failed for operation %s: %s", operation, err.Error()),
	)

	return errors.SendError(c, appErr)
}

// HandleProjectIDError handles invalid project ID errors
func (h *HistoryErrorHandler) HandleProjectIDError(c *fiber.Ctx, err error, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewInvalidInputError("project_id").
		WithDetails(err.Error()).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID)

	// Log the invalid project ID access attempt
	requestLogger.LogProjectAccess(0, "invalid_project_id", false, "INVALID_PROJECT_ID")

	return errors.SendError(c, appErr)
}

// HandleProjectNotFoundError handles project not found errors
func (h *HistoryErrorHandler) HandleProjectNotFoundError(c *fiber.Ctx, err error, userID, projectID uint, operation string) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewProjectNotFoundError().
		WithDetails(err.Error()).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID)

	// Log the project access attempt
	requestLogger.LogProjectAccess(projectID, operation, false, "PROJECT_NOT_FOUND")

	return errors.SendError(c, appErr)
}

// HandleFileNotFoundError handles file not found errors
func (h *HistoryErrorHandler) HandleFileNotFoundError(c *fiber.Ctx, projectID uint, userID uint, reason string) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewFileNotFoundError("project file").
		WithDetails(reason).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID).
		WithContext("can_regenerate", true)

	// Log the file access attempt
	requestLogger.LogProjectAccess(projectID, "download_project", false, "FILE_NOT_FOUND")

	return errors.SendError(c, appErr)
}

// HandleFileExpiredError handles expired file errors
func (h *HistoryErrorHandler) HandleFileExpiredError(c *fiber.Ctx, projectID uint, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewFileExpiredError().
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID)

	// Log the expired file access attempt
	requestLogger.LogProjectAccess(projectID, "download_project", false, "FILE_EXPIRED")

	return errors.SendError(c, appErr)
}

// HandleInvalidFilePathError handles invalid file path errors
func (h *HistoryErrorHandler) HandleInvalidFilePathError(c *fiber.Ctx, err error, projectID uint, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewInvalidFilePathError().
		WithDetails(err.Error()).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID)

	// Log the invalid file path attempt as a security event
	requestLogger.LogSecurityEvent(
		"INVALID_FILE_PATH",
		"medium",
		fmt.Sprintf("Invalid file path access attempt for project %d: %s", projectID, err.Error()),
	)

	return errors.SendError(c, appErr)
}

// HandleDatabaseError handles database-related errors
func (h *HistoryErrorHandler) HandleDatabaseError(c *fiber.Ctx, err error, operation string, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewDatabaseError(operation).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithInternalError(err)

	// Log the database error
	requestLogger.Error(fmt.Sprintf("Database error during %s", operation), err, map[string]interface{}{
		"operation": operation,
		"user_id":   userID,
	})

	return errors.SendError(c, appErr)
}

// HandleGenerationError handles project generation errors
func (h *HistoryErrorHandler) HandleGenerationError(c *fiber.Ctx, err error, userID uint, projectName string) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewGenerationFailedError(err.Error()).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_name", projectName)

	// Log the generation failure
	requestLogger.Error("Project generation failed", err, map[string]interface{}{
		"project_name": projectName,
		"user_id":      userID,
	})

	return errors.SendError(c, appErr)
}

// HandleInsufficientDataError handles insufficient data errors
func (h *HistoryErrorHandler) HandleInsufficientDataError(c *fiber.Ctx, missing string, userID uint, projectID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewInsufficientDataError(missing).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID)

	// Log the insufficient data error
	requestLogger.LogProjectAccess(projectID, "insufficient_data", false, missing)

	return errors.SendError(c, appErr)
}

// HandleValidationError handles request validation errors
func (h *HistoryErrorHandler) HandleValidationError(c *fiber.Ctx, validationErr error, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	// Parse validation errors if possible
	fieldErrors := parseValidationErrors(validationErr.Error())
	
	// Log the validation error
	requestLogger.LogUserAction("validation_failed", "request", false, validationErr.Error())

	return errors.SendValidationError(c, "Request validation failed", fieldErrors)
}

// HandleFileOperationError handles file operation errors
func (h *HistoryErrorHandler) HandleFileOperationError(c *fiber.Ctx, err error, operation string, userID uint, projectID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewAppError(
		errors.ErrCodeFileOperationFailed,
		fmt.Sprintf("File operation failed: %s", operation),
		500,
	).WithDetails(err.Error()).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID).
		WithContext("project_id", projectID).
		WithContext("operation", operation)

	// Log the file operation error
	requestLogger.Error(fmt.Sprintf("File operation failed: %s", operation), err, map[string]interface{}{
		"operation":  operation,
		"project_id": projectID,
		"user_id":    userID,
	})

	return errors.SendError(c, appErr)
}

// HandleRateLimitError handles rate limiting errors
func (h *HistoryErrorHandler) HandleRateLimitError(c *fiber.Ctx, retryAfter string, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewRateLimitExceededError(retryAfter).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID)

	// Log the rate limit event
	requestLogger.LogSecurityEvent(
		"RATE_LIMIT_EXCEEDED",
		"medium",
		fmt.Sprintf("Rate limit exceeded for user %d", userID),
	)

	return errors.SendError(c, appErr)
}

// HandleServiceUnavailableError handles service unavailable errors
func (h *HistoryErrorHandler) HandleServiceUnavailableError(c *fiber.Ctx, service string, userID uint) error {
	requestLogger := h.logger.NewRequestLogger(c)
	
	appErr := errors.NewServiceUnavailableError(service).
		WithRequestID(c.Get("X-Request-ID", "")).
		WithUserID(userID)

	// Log the service unavailable event
	requestLogger.Error(fmt.Sprintf("Service unavailable: %s", service), nil, map[string]interface{}{
		"service": service,
		"user_id": userID,
	})

	return errors.SendError(c, appErr)
}

// LogSuccessfulOperation logs successful operations
func (h *HistoryErrorHandler) LogSuccessfulOperation(c *fiber.Ctx, operation string, userID uint, projectID *uint, metadata map[string]interface{}) {
	requestLogger := h.logger.NewRequestLogger(c)
	
	if projectID != nil {
		requestLogger.LogProjectAccess(*projectID, operation, true, "")
	} else {
		requestLogger.LogUserAction(operation, "project_history", true, "")
	}

	// Log performance metrics if duration is provided
	if duration, ok := metadata["duration"]; ok {
		if d, ok := duration.(int64); ok {
			requestLogger.LogPerformanceMetric(operation, 
				time.Duration(d)*time.Millisecond, 
				metadata)
		}
	}
}

// parseValidationErrors parses validation error messages into field-specific errors
func parseValidationErrors(errorMsg string) map[string][]string {
	fieldErrors := make(map[string][]string)
	
	// This is a simplified parser - in a real implementation, you'd want
	// to integrate more closely with your validation library
	if errorMsg != "" {
		fieldErrors["general"] = []string{errorMsg}
	}
	
	return fieldErrors
}

// GetUserIDFromContext safely extracts user ID from Fiber context
func GetUserIDFromContext(c *fiber.Ctx) (uint, error) {
	if uid := c.Locals("user_id"); uid != nil {
		if userID, ok := uid.(uint); ok {
			return userID, nil
		}
	}
	return 0, fmt.Errorf("user ID not found in context")
}

// GetProjectIDFromParams safely extracts project ID from URL parameters
func GetProjectIDFromParams(c *fiber.Ctx) (uint, error) {
	idStr := c.Params("id")
	if idStr == "" {
		return 0, fmt.Errorf("project ID parameter is required")
	}

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid project ID format: %s", idStr)
	}

	if id == 0 {
		return 0, fmt.Errorf("project ID must be greater than 0")
	}

	return uint(id), nil
}

// ValidateRequestBody validates and parses request body
func ValidateRequestBody(c *fiber.Ctx, target interface{}) error {
	if err := c.BodyParser(target); err != nil {
		return fmt.Errorf("invalid request body format: %w", err)
	}

	// Additional validation can be added here using the validation package
	return nil
}