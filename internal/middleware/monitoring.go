package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/logging"
	"github.com/telman03/ai-backend-generator/internal/monitoring"
)

// MonitoringConfig contains configuration for the monitoring middleware
type MonitoringConfig struct {
	// EnableMetrics determines if metrics collection is enabled
	EnableMetrics bool
	// EnableAuditLogging determines if audit logging is enabled
	EnableAuditLogging bool
	// EnablePerformanceMonitoring determines if performance monitoring is enabled
	EnablePerformanceMonitoring bool
	// SlowRequestThreshold defines the threshold for slow request logging
	SlowRequestThreshold time.Duration
	// SkipPaths defines paths to skip monitoring (e.g., health checks)
	SkipPaths []string
}

// DefaultMonitoringConfig returns the default monitoring configuration
func DefaultMonitoringConfig() MonitoringConfig {
	return MonitoringConfig{
		EnableMetrics:               true,
		EnableAuditLogging:          true,
		EnablePerformanceMonitoring: true,
		SlowRequestThreshold:        500 * time.Millisecond,
		SkipPaths:                   []string{"/health", "/ping", "/metrics"},
	}
}

// ComprehensiveMonitoringMiddleware creates a comprehensive monitoring middleware
func ComprehensiveMonitoringMiddleware(config ...MonitoringConfig) fiber.Handler {
	cfg := DefaultMonitoringConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	logger := logging.GetLogger().WithComponent("monitoring")
	metricsCollector := monitoring.GetMetrics()
	performanceMonitor := monitoring.GetPerformanceMonitor()
	auditLogger := logging.GetAuditLogger()

	return func(c *fiber.Ctx) error {
		// Check if we should skip monitoring for this path
		if shouldSkipPath(c.Path(), cfg.SkipPaths) {
			return c.Next()
		}

		// Record start time
		start := time.Now()

		// Increment active connections if metrics are enabled
		if cfg.EnableMetrics {
			metricsCollector.IncrementActiveConnections()
			defer metricsCollector.DecrementActiveConnections()
		}

		// Create request logger
		requestLogger := logger.NewRequestLogger(c)

		// Process the request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)
		statusCode := c.Response().StatusCode()
		endpoint := c.Method() + " " + c.Route().Path

		// Get user ID if available
		var userID *uint
		if uid := c.Locals("user_id"); uid != nil {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Record metrics if enabled
		if cfg.EnableMetrics {
			metricsCollector.RecordRequest(endpoint, duration, statusCode)
			
			if err != nil {
				metricsCollector.RecordError("request_error", endpoint)
			}

			// Record security events for specific status codes
			recordSecurityEvents(metricsCollector, c, statusCode, userID)
		}

		// Record performance metrics if enabled
		if cfg.EnablePerformanceMonitoring {
			success := err == nil && statusCode < 400
			performanceMonitor.RecordOperation(endpoint, duration, success)

			// Record detailed performance metrics for slow requests
			if duration > cfg.SlowRequestThreshold {
				metadata := map[string]interface{}{
					"status_code":   statusCode,
					"response_size": len(c.Response().Body()),
					"ip":           c.IP(),
					"user_agent":   c.Get("User-Agent"),
					"endpoint":     endpoint,
				}

				performanceMonitor.RecordOperation("slow_request", duration, success)
				
				// Log slow request
				requestLogger.LogPerformanceMetric("slow_request", duration, metadata)
			}
		}

		// Record audit events if enabled
		if cfg.EnableAuditLogging && auditLogger != nil {
			recordAuditEvent(auditLogger, c, userID, duration, statusCode, err)
		}

		// Log errors with context
		if err != nil {
			logRequestError(requestLogger, c, err, duration, userID)
		}

		// Log successful requests (debug level)
		if err == nil && statusCode < 400 {
			requestLogger.Debug("Request completed successfully", map[string]interface{}{
				"endpoint":     endpoint,
				"status_code":  statusCode,
				"duration_ms":  duration.Milliseconds(),
				"user_id":      userID,
			})
		}

		return err
	}
}

// shouldSkipPath checks if a path should be skipped for monitoring
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// recordSecurityEvents records security-related events in metrics
func recordSecurityEvents(metricsCollector *monitoring.MetricsCollector, c *fiber.Ctx, statusCode int, userID *uint) {
	var event monitoring.SecurityEvent
	
	switch statusCode {
	case 401:
		event = monitoring.SecurityEvent{
			Timestamp: time.Now(),
			Event:     "UNAUTHORIZED_ACCESS",
			Severity:  "medium",
			UserID:    userID,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			Details:   "Unauthorized access attempt",
			RequestID: c.Get("X-Request-ID", ""),
		}
	case 403:
		event = monitoring.SecurityEvent{
			Timestamp: time.Now(),
			Event:     "FORBIDDEN_ACCESS",
			Severity:  "high",
			UserID:    userID,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			Details:   "Forbidden access attempt",
			RequestID: c.Get("X-Request-ID", ""),
		}
	case 429:
		event = monitoring.SecurityEvent{
			Timestamp: time.Now(),
			Event:     "RATE_LIMIT_EXCEEDED",
			Severity:  "medium",
			UserID:    userID,
			IP:        c.IP(),
			UserAgent: c.Get("User-Agent"),
			Details:   "Rate limit exceeded",
			RequestID: c.Get("X-Request-ID", ""),
		}
	default:
		return // No security event to record
	}

	metricsCollector.RecordSecurityEvent(event)
}

// recordAuditEvent records an audit event
func recordAuditEvent(auditLogger *logging.AuditLogger, c *fiber.Ctx, userID *uint, duration time.Duration, statusCode int, err error) {
	// Determine event type based on HTTP method
	var eventType logging.AuditEventType
	switch c.Method() {
	case "GET":
		eventType = logging.AuditDataRead
	case "POST":
		eventType = logging.AuditDataCreate
	case "PUT", "PATCH":
		eventType = logging.AuditDataUpdate
	case "DELETE":
		eventType = logging.AuditDataDelete
	default:
		eventType = logging.AuditDataRead
	}

	// Determine result
	result := "SUCCESS"
	if err != nil || statusCode >= 400 {
		result = "FAILURE"
	}

	// Determine severity
	var severity logging.AuditSeverity
	switch {
	case statusCode >= 500:
		severity = logging.AuditSeverityHigh
	case statusCode >= 400:
		severity = logging.AuditSeverityMedium
	default:
		severity = logging.AuditSeverityLow
	}

	// Create audit event
	durationMs := duration.Milliseconds()
	event := logging.AuditEvent{
		Timestamp:  time.Now(),
		EventType:  eventType,
		Severity:   severity,
		UserID:     userID,
		RequestID:  c.Get("X-Request-ID", ""),
		IP:         c.IP(),
		UserAgent:  c.Get("User-Agent"),
		Resource:   c.Route().Path,
		Action:     c.Method(),
		Result:     result,
		Method:     c.Method(),
		Path:       c.Path(),
		StatusCode: &statusCode,
		Duration:   &durationMs,
	}

	if err != nil {
		event.ErrorMessage = err.Error()
	}

	auditLogger.LogEvent(event)
}

// logRequestError logs request errors with comprehensive context
func logRequestError(requestLogger *logging.RequestLogger, c *fiber.Ctx, err error, duration time.Duration, userID *uint) {
	fields := map[string]interface{}{
		"method":       c.Method(),
		"path":         c.Path(),
		"status_code":  c.Response().StatusCode(),
		"duration_ms":  duration.Milliseconds(),
		"ip":           c.IP(),
		"user_agent":   c.Get("User-Agent"),
		"request_id":   c.Get("X-Request-ID", ""),
		"user_id":      userID,
		"query_params": c.Queries(),
		"body_size":    len(c.Body()),
	}

	// Add request headers (filtered)
	headers := make(map[string]string)
	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := string(key)
		// Skip sensitive headers
		if !isSensitiveHeaderMonitoring(keyStr) {
			headers[keyStr] = string(value)
		}
	})
	fields["headers"] = headers

	requestLogger.Error("Request failed", err, fields)
}

// isSensitiveHeaderMonitoring checks if a header contains sensitive information
func isSensitiveHeaderMonitoring(header string) bool {
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

// HealthCheckMiddleware creates a simple health check endpoint
func HealthCheckMiddleware() fiber.Handler {
	performanceMonitor := monitoring.GetPerformanceMonitor()
	metricsCollector := monitoring.GetMetrics()

	return func(c *fiber.Ctx) error {
		if c.Path() != "/health" {
			return c.Next()
		}

		// Get system metrics
		systemMetrics := performanceMonitor.GetSystemMetrics()
		metricsSummary := metricsCollector.GetSummary()

		health := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
			"uptime":    metricsSummary.Uptime.String(),
			"system": map[string]interface{}{
				"memory_usage_percent": systemMetrics.MemoryUsagePercent,
				"goroutine_count":      systemMetrics.GoroutineCount,
				"heap_size_mb":         float64(systemMetrics.HeapSize) / 1024 / 1024,
			},
			"metrics": map[string]interface{}{
				"total_requests":       metricsSummary.TotalRequests,
				"error_rate":           metricsSummary.ErrorRate,
				"avg_response_time_ms": metricsSummary.AverageResponseTime,
				"active_connections":   metricsSummary.ActiveConnections,
			},
		}

		// Determine overall health status
		if systemMetrics.MemoryUsagePercent > 90 || metricsSummary.ErrorRate > 20 {
			health["status"] = "degraded"
		}

		if systemMetrics.MemoryUsagePercent > 95 || metricsSummary.ErrorRate > 50 {
			health["status"] = "unhealthy"
		}

		return c.JSON(health)
	}
}

// MetricsEndpointMiddleware creates a metrics endpoint
func MetricsEndpointMiddleware() fiber.Handler {
	performanceMonitor := monitoring.GetPerformanceMonitor()
	metricsCollector := monitoring.GetMetrics()

	return func(c *fiber.Ctx) error {
		if c.Path() != "/metrics" {
			return c.Next()
		}

		// Get comprehensive metrics
		performanceSummary := performanceMonitor.GetPerformanceSummary()
		metricsSummary := metricsCollector.GetSummary()

		metrics := map[string]interface{}{
			"timestamp":           time.Now(),
			"performance_metrics": performanceSummary,
			"request_metrics":     metricsSummary,
		}

		return c.JSON(metrics)
	}
}