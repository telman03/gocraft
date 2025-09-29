package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/monitoring"
)

// MetricsMiddleware creates a middleware that collects request metrics
func MetricsMiddleware() fiber.Handler {
	metrics := monitoring.GetMetrics()

	return func(c *fiber.Ctx) error {
		// Record connection
		metrics.IncrementActiveConnections()
		defer metrics.DecrementActiveConnections()

		// Record start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get endpoint (method + path)
		endpoint := c.Method() + " " + c.Route().Path

		// Get status code
		statusCode := c.Response().StatusCode()

		// Record the request metrics
		metrics.RecordRequest(endpoint, duration, statusCode)

		// Record error if present
		if err != nil {
			metrics.RecordError("request_error", endpoint)
		}

		return err
	}
}

// SecurityEventMiddleware creates a middleware that records security events
func SecurityEventMiddleware() fiber.Handler {
	metrics := monitoring.GetMetrics()

	return func(c *fiber.Ctx) error {
		// Process request
		err := c.Next()

		// Check for security-related status codes
		statusCode := c.Response().StatusCode()
		
		var userID *uint
		if uid := c.Locals("user_id"); uid != nil {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Record security events based on status codes
		switch statusCode {
		case 401:
			metrics.RecordSecurityEvent(monitoring.SecurityEvent{
				Timestamp: time.Now(),
				Event:     "UNAUTHORIZED_ACCESS",
				Severity:  "medium",
				UserID:    userID,
				IP:        c.IP(),
				UserAgent: c.Get("User-Agent"),
				Details:   "Unauthorized access attempt",
				RequestID: c.Get("X-Request-ID", ""),
			})
		case 403:
			metrics.RecordSecurityEvent(monitoring.SecurityEvent{
				Timestamp: time.Now(),
				Event:     "FORBIDDEN_ACCESS",
				Severity:  "high",
				UserID:    userID,
				IP:        c.IP(),
				UserAgent: c.Get("User-Agent"),
				Details:   "Forbidden access attempt",
				RequestID: c.Get("X-Request-ID", ""),
			})
		case 429:
			metrics.RecordSecurityEvent(monitoring.SecurityEvent{
				Timestamp: time.Now(),
				Event:     "RATE_LIMIT_EXCEEDED",
				Severity:  "medium",
				UserID:    userID,
				IP:        c.IP(),
				UserAgent: c.Get("User-Agent"),
				Details:   "Rate limit exceeded",
				RequestID: c.Get("X-Request-ID", ""),
			})
		}

		return err
	}
}

// PerformanceMonitoringMiddleware creates a middleware that monitors performance
func PerformanceMonitoringMiddleware() fiber.Handler {
	metrics := monitoring.GetMetrics()

	return func(c *fiber.Ctx) error {
		// Skip monitoring for health check endpoints
		if c.Path() == "/health" || c.Path() == "/ping" {
			return c.Next()
		}

		start := time.Now()
		
		// Process request
		err := c.Next()
		
		duration := time.Since(start)
		
		// Get user ID if available
		var userID *uint
		if uid := c.Locals("user_id"); uid != nil {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Record performance metric for slow requests (>500ms)
		if duration > 500*time.Millisecond {
			operation := c.Method() + " " + c.Route().Path
			
			metadata := map[string]interface{}{
				"status_code":   c.Response().StatusCode(),
				"response_size": len(c.Response().Body()),
				"ip":           c.IP(),
				"user_agent":   c.Get("User-Agent"),
			}

			metrics.RecordPerformanceMetric(monitoring.PerformanceMetric{
				Timestamp: time.Now(),
				Operation: operation,
				Duration:  duration,
				Success:   err == nil && c.Response().StatusCode() < 400,
				UserID:    userID,
				Metadata:  metadata,
			})
		}

		return err
	}
}