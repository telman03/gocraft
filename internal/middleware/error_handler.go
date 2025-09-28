package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/errors"
	"github.com/telman03/ai-backend-generator/internal/logging"
)

// ErrorHandlerConfig contains configuration for the error handler middleware
type ErrorHandlerConfig struct {
	// EnableStackTrace determines if stack traces should be included in error responses
	EnableStackTrace bool
	// EnableDetailedErrors determines if detailed error information should be exposed
	EnableDetailedErrors bool
	// Logger is the logger instance to use for error logging
	Logger *logging.Logger
}

// DefaultErrorHandlerConfig returns the default configuration
func DefaultErrorHandlerConfig() ErrorHandlerConfig {
	return ErrorHandlerConfig{
		EnableStackTrace:     false, // Disable in production
		EnableDetailedErrors: false, // Disable in production
		Logger:              logging.GetLogger().WithComponent("error_handler"),
	}
}

// ErrorHandlerMiddleware creates a comprehensive error handling middleware
func ErrorHandlerMiddleware(config ...ErrorHandlerConfig) fiber.Handler {
	cfg := DefaultErrorHandlerConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		// Create request logger for this request
		requestLogger := cfg.Logger.NewRequestLogger(c)

		// Process the request and capture any errors
		err := c.Next()

		// If no error occurred, continue normally
		if err == nil {
			return nil
		}

		// Handle the error
		return handleError(c, err, cfg, requestLogger)
	}
}

// handleError processes and responds to errors
func handleError(c *fiber.Ctx, err error, config ErrorHandlerConfig, logger *logging.RequestLogger) error {
	// Extract user ID if available
	var userID *uint
	if uid := c.Locals("user_id"); uid != nil {
		if id, ok := uid.(uint); ok {
			userID = &id
		}
	}

	// Check if it's already an AppError
	if appErr, ok := errors.IsAppError(err); ok {
		// Log the error with context
		logAppError(logger, appErr, c, userID)
		
		// Send the structured error response
		return errors.SendError(c, appErr)
	}

	// Check if it's a Fiber error
	if fiberErr, ok := err.(*fiber.Error); ok {
		appErr := errors.FromFiberError(fiberErr)
		if userID != nil {
			appErr = appErr.WithUserID(*userID)
		}
		
		// Add request context
		appErr = appErr.WithRequestID(c.Get("X-Request-ID", ""))
		
		// Log the error
		logAppError(logger, appErr, c, userID)
		
		return errors.SendError(c, appErr)
	}

	// Handle unknown errors
	appErr := errors.NewInternalError("An unexpected error occurred")
	if userID != nil {
		appErr = appErr.WithUserID(*userID)
	}
	appErr = appErr.WithRequestID(c.Get("X-Request-ID", "")).WithInternalError(err)

	// Log the unexpected error with more details
	logUnexpectedError(logger, err, c, userID)

	// In development, include more details
	if config.EnableDetailedErrors {
		appErr = appErr.WithDetails(err.Error())
	}

	return errors.SendError(c, appErr)
}

// logAppError logs an AppError with appropriate context
func logAppError(logger *logging.RequestLogger, appErr *errors.AppError, c *fiber.Ctx, userID *uint) {
	fields := map[string]interface{}{
		"error_code":    string(appErr.Code),
		"status_code":   appErr.StatusCode,
		"method":        c.Method(),
		"path":          c.Path(),
		"ip":            c.IP(),
		"user_agent":    c.Get("User-Agent"),
		"request_id":    c.Get("X-Request-ID", ""),
	}

	if userID != nil {
		fields["user_id"] = *userID
	}

	if appErr.Context != nil {
		fields["error_context"] = appErr.Context
	}

	if appErr.InternalErr != nil {
		fields["internal_error"] = appErr.InternalErr.Error()
	}

	// Log based on error severity
	if appErr.StatusCode >= 500 {
		logger.Error(fmt.Sprintf("Server error: %s", appErr.Message), appErr.InternalErr, fields)
	} else if appErr.StatusCode >= 400 {
		logger.Warn(fmt.Sprintf("Client error: %s", appErr.Message), fields)
	} else {
		logger.Info(fmt.Sprintf("Error handled: %s", appErr.Message), fields)
	}

	// Log security events for specific error types
	if isSecurityRelatedError(appErr) {
		severity := "medium"
		if appErr.StatusCode == 401 || appErr.StatusCode == 403 {
			severity = "high"
		}
		
		logger.LogSecurityEvent(
			string(appErr.Code),
			severity,
			fmt.Sprintf("Security error: %s from IP %s", appErr.Message, c.IP()),
		)
	}
}

// logUnexpectedError logs unexpected errors with full context
func logUnexpectedError(logger *logging.RequestLogger, err error, c *fiber.Ctx, userID *uint) {
	fields := map[string]interface{}{
		"error_type":    fmt.Sprintf("%T", err),
		"method":        c.Method(),
		"path":          c.Path(),
		"ip":            c.IP(),
		"user_agent":    c.Get("User-Agent"),
		"request_id":    c.Get("X-Request-ID", ""),
		"query_params":  c.Queries(),
		"body_size":     len(c.Body()),
	}

	if userID != nil {
		fields["user_id"] = *userID
	}

	// Add request headers (filtered)
	headers := make(map[string]string)
	// Iterate through headers using the new approach
	for key, values := range c.GetReqHeaders() {
		keyStr := strings.ToLower(key)
		// Skip sensitive headers
		if !isSensitiveHeader(keyStr) && len(values) > 0 {
			headers[keyStr] = values[0]
		}
	}
	fields["headers"] = headers

	logger.Error("Unexpected error occurred", err, fields)
}

// isSecurityRelatedError determines if an error is security-related
func isSecurityRelatedError(appErr *errors.AppError) bool {
	securityCodes := map[errors.ErrorCode]bool{
		errors.ErrCodeAuthFailed:         true,
		errors.ErrCodeInvalidToken:       true,
		errors.ErrCodeTokenExpired:       true,
		errors.ErrCodeInsufficientPerms:  true,
		errors.ErrCodeAccessDenied:       true,
		errors.ErrCodeRateLimitExceeded:  true,
		errors.ErrCodeTooManyRequests:    true,
	}
	
	return securityCodes[appErr.Code]
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
		"x-csrf-token":  true,
		"x-session-id":  true,
	}
	
	return sensitiveHeaders[header]
}

// PanicRecoveryMiddleware recovers from panics and converts them to errors
func PanicRecoveryMiddleware() fiber.Handler {
	logger := logging.GetLogger().WithComponent("panic_recovery")

	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				// Create request logger
				requestLogger := logger.NewRequestLogger(c)

				// Log the panic with full context
				fields := map[string]interface{}{
					"panic_value": fmt.Sprintf("%v", r),
					"method":      c.Method(),
					"path":        c.Path(),
					"ip":          c.IP(),
					"user_agent":  c.Get("User-Agent"),
					"request_id":  c.Get("X-Request-ID", ""),
				}

				if uid := c.Locals("user_id"); uid != nil {
					fields["user_id"] = uid
				}

				requestLogger.Error("Panic recovered", fmt.Errorf("panic: %v", r), fields)

				// Log as critical security event
				requestLogger.LogSecurityEvent(
					"PANIC_RECOVERED",
					"critical",
					fmt.Sprintf("Application panic recovered: %v", r),
				)

				// Create and send error response
				appErr := errors.NewInternalError("Internal server error").
					WithRequestID(c.Get("X-Request-ID", "")).
					WithDetails("An unexpected error occurred")

				errors.SendError(c, appErr)
			}
		}()

		return c.Next()
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check if request ID already exists
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = c.Get("X-Trace-Id")
		}
		
		// Generate new request ID if none exists
		if requestID == "" {
			requestID = generateRequestID()
			c.Set("X-Request-ID", requestID)
		}

		// Store in locals for easy access
		c.Locals("request_id", requestID)

		return c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple timestamp-based ID (in production, consider using UUID)
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// ValidationErrorMiddleware handles validation errors specifically
func ValidationErrorMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		
		if err == nil {
			return nil
		}

		// Check if it's a validation error that we can parse
		if strings.Contains(err.Error(), "validation failed") {
			// Parse validation errors and send structured response
			fieldErrors := parseValidationErrors(err.Error())
			return errors.SendValidationError(c, "Request validation failed", fieldErrors)
		}

		// Let other middleware handle non-validation errors
		return err
	}
}

// parseValidationErrors parses validation error messages into field-specific errors
func parseValidationErrors(errorMsg string) map[string][]string {
	// This is a simplified parser - in a real implementation, you'd want
	// to integrate with your validation library (e.g., go-playground/validator)
	fieldErrors := make(map[string][]string)
	
	// For now, return a generic validation error
	fieldErrors["general"] = []string{errorMsg}
	
	return fieldErrors
}

// TimeoutMiddleware handles request timeouts
func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create a channel to signal completion
		done := make(chan error, 1)
		
		// Run the request in a goroutine
		go func() {
			done <- c.Next()
		}()
		
		// Wait for completion or timeout
		select {
		case err := <-done:
			return err
		case <-time.After(timeout):
			// Log timeout
			logger := logging.GetLogger().WithComponent("timeout")
			requestLogger := logger.NewRequestLogger(c)
			
			requestLogger.Warn("Request timeout", map[string]interface{}{
				"timeout_duration": timeout.String(),
				"method":          c.Method(),
				"path":            c.Path(),
			})
			
			// Return timeout error
			appErr := errors.NewTimeoutError("request processing").
				WithRequestID(c.Get("X-Request-ID", "")).
				WithContext("timeout_duration", timeout.String())
			
			return errors.SendError(c, appErr)
		}
	}
}