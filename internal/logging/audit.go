package logging

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// AuditEventType represents the type of audit event
type AuditEventType string

const (
	// Authentication Events
	AuditLoginSuccess    AuditEventType = "LOGIN_SUCCESS"
	AuditLoginFailure    AuditEventType = "LOGIN_FAILURE"
	AuditLogout          AuditEventType = "LOGOUT"
	AuditTokenRefresh    AuditEventType = "TOKEN_REFRESH"
	AuditPasswordChange  AuditEventType = "PASSWORD_CHANGE"

	// Authorization Events
	AuditAccessGranted   AuditEventType = "ACCESS_GRANTED"
	AuditAccessDenied    AuditEventType = "ACCESS_DENIED"
	AuditPermissionCheck AuditEventType = "PERMISSION_CHECK"

	// Data Access Events
	AuditDataRead        AuditEventType = "DATA_READ"
	AuditDataCreate      AuditEventType = "DATA_CREATE"
	AuditDataUpdate      AuditEventType = "DATA_UPDATE"
	AuditDataDelete      AuditEventType = "DATA_DELETE"

	// Project History Events
	AuditProjectView     AuditEventType = "PROJECT_VIEW"
	AuditProjectDownload AuditEventType = "PROJECT_DOWNLOAD"
	AuditProjectDelete   AuditEventType = "PROJECT_DELETE"
	AuditProjectGenerate AuditEventType = "PROJECT_GENERATE"
	AuditProjectDuplicate AuditEventType = "PROJECT_DUPLICATE"

	// Security Events
	AuditSecurityViolation AuditEventType = "SECURITY_VIOLATION"
	AuditSuspiciousActivity AuditEventType = "SUSPICIOUS_ACTIVITY"
	AuditRateLimitExceeded AuditEventType = "RATE_LIMIT_EXCEEDED"
	AuditInvalidInput      AuditEventType = "INVALID_INPUT"

	// System Events
	AuditSystemStart     AuditEventType = "SYSTEM_START"
	AuditSystemShutdown  AuditEventType = "SYSTEM_SHUTDOWN"
	AuditConfigChange    AuditEventType = "CONFIG_CHANGE"
	AuditMaintenanceMode AuditEventType = "MAINTENANCE_MODE"
)

// AuditSeverity represents the severity level of an audit event
type AuditSeverity string

const (
	AuditSeverityLow      AuditSeverity = "LOW"
	AuditSeverityMedium   AuditSeverity = "MEDIUM"
	AuditSeverityHigh     AuditSeverity = "HIGH"
	AuditSeverityCritical AuditSeverity = "CRITICAL"
)

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp    time.Time              `json:"timestamp"`
	EventType    AuditEventType         `json:"event_type"`
	Severity     AuditSeverity          `json:"severity"`
	UserID       *uint                  `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	RequestID    string                 `json:"request_id,omitempty"`
	IP           string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Resource     string                 `json:"resource,omitempty"`
	Action       string                 `json:"action,omitempty"`
	Result       string                 `json:"result"`
	Details      string                 `json:"details,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	Method       string                 `json:"http_method,omitempty"`
	Path         string                 `json:"http_path,omitempty"`
	StatusCode   *int                   `json:"status_code,omitempty"`
	Duration     *int64                 `json:"duration_ms,omitempty"`
	ErrorCode    string                 `json:"error_code,omitempty"`
	ErrorMessage string                 `json:"error_message,omitempty"`
}

// AuditLogger provides audit logging capabilities
type AuditLogger struct {
	logger   *Logger
	filePath string
	file     *os.File
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(filePath string) (*AuditLogger, error) {
	// Create audit log file if it doesn't exist
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log file: %w", err)
	}

	return &AuditLogger{
		logger:   GetLogger().WithComponent("audit"),
		filePath: filePath,
		file:     file,
	}, nil
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event AuditEvent) {
	// Ensure timestamp is set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(event)
	if err != nil {
		al.logger.Error("Failed to marshal audit event", err, map[string]interface{}{
			"event_type": event.EventType,
			"user_id":    event.UserID,
		})
		return
	}

	// Write to audit log file
	if al.file != nil {
		if _, err := al.file.WriteString(string(jsonData) + "\n"); err != nil {
			al.logger.Error("Failed to write audit event to file", err)
		}
		al.file.Sync() // Ensure data is written to disk
	}

	// Also log to structured logger based on severity
	switch event.Severity {
	case AuditSeverityCritical:
		al.logger.Error(fmt.Sprintf("AUDIT: %s", event.EventType), nil, map[string]interface{}{
			"audit_event": event,
		})
	case AuditSeverityHigh:
		al.logger.Warn(fmt.Sprintf("AUDIT: %s", event.EventType), map[string]interface{}{
			"audit_event": event,
		})
	default:
		al.logger.Info(fmt.Sprintf("AUDIT: %s", event.EventType), map[string]interface{}{
			"audit_event": event,
		})
	}
}

// LogAuthenticationEvent logs authentication-related events
func (al *AuditLogger) LogAuthenticationEvent(eventType AuditEventType, c *fiber.Ctx, userID *uint, result string, details string) {
	event := AuditEvent{
		EventType: eventType,
		Severity:  al.getSeverityForAuthEvent(eventType, result),
		UserID:    userID,
		RequestID: c.Get("X-Request-ID", ""),
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Action:    string(eventType),
		Result:    result,
		Details:   details,
		Method:    c.Method(),
		Path:      c.Path(),
	}

	if result != "SUCCESS" {
		event.ErrorMessage = details
	}

	al.LogEvent(event)
}

// LogDataAccessEvent logs data access events
func (al *AuditLogger) LogDataAccessEvent(eventType AuditEventType, c *fiber.Ctx, userID uint, resource string, resourceID *uint, result string, details string) {
	context := map[string]interface{}{
		"resource": resource,
	}
	if resourceID != nil {
		context["resource_id"] = *resourceID
	}

	event := AuditEvent{
		EventType: eventType,
		Severity:  al.getSeverityForDataEvent(eventType, result),
		UserID:    &userID,
		RequestID: c.Get("X-Request-ID", ""),
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Resource:  resource,
		Action:    string(eventType),
		Result:    result,
		Details:   details,
		Context:   context,
		Method:    c.Method(),
		Path:      c.Path(),
	}

	al.LogEvent(event)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(eventType AuditEventType, c *fiber.Ctx, userID *uint, severity AuditSeverity, details string, errorCode string) {
	event := AuditEvent{
		EventType:    eventType,
		Severity:     severity,
		UserID:       userID,
		RequestID:    c.Get("X-Request-ID", ""),
		IP:           c.IP(),
		UserAgent:    c.Get("User-Agent"),
		Action:       string(eventType),
		Result:       "SECURITY_EVENT",
		Details:      details,
		Method:       c.Method(),
		Path:         c.Path(),
		ErrorCode:    errorCode,
		ErrorMessage: details,
	}

	al.LogEvent(event)
}

// LogSystemEvent logs system-related events
func (al *AuditLogger) LogSystemEvent(eventType AuditEventType, details string, context map[string]interface{}) {
	event := AuditEvent{
		EventType: eventType,
		Severity:  AuditSeverityMedium,
		Action:    string(eventType),
		Result:    "SYSTEM_EVENT",
		Details:   details,
		Context:   context,
	}

	al.LogEvent(event)
}

// LogProjectHistoryEvent logs project history specific events
func (al *AuditLogger) LogProjectHistoryEvent(eventType AuditEventType, c *fiber.Ctx, userID uint, projectID uint, result string, details string, duration *time.Duration) {
	context := map[string]interface{}{
		"project_id": projectID,
	}

	event := AuditEvent{
		EventType: eventType,
		Severity:  al.getSeverityForProjectEvent(eventType, result),
		UserID:    &userID,
		RequestID: c.Get("X-Request-ID", ""),
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Resource:  "project_history",
		Action:    string(eventType),
		Result:    result,
		Details:   details,
		Context:   context,
		Method:    c.Method(),
		Path:      c.Path(),
	}

	if duration != nil {
		durationMs := duration.Milliseconds()
		event.Duration = &durationMs
	}

	statusCode := c.Response().StatusCode()
	event.StatusCode = &statusCode

	al.LogEvent(event)
}

// getSeverityForAuthEvent determines severity for authentication events
func (al *AuditLogger) getSeverityForAuthEvent(eventType AuditEventType, result string) AuditSeverity {
	if result != "SUCCESS" {
		switch eventType {
		case AuditLoginFailure:
			return AuditSeverityHigh
		case AuditPasswordChange:
			return AuditSeverityMedium
		default:
			return AuditSeverityMedium
		}
	}

	switch eventType {
	case AuditPasswordChange:
		return AuditSeverityMedium
	default:
		return AuditSeverityLow
	}
}

// getSeverityForDataEvent determines severity for data access events
func (al *AuditLogger) getSeverityForDataEvent(eventType AuditEventType, result string) AuditSeverity {
	if result != "SUCCESS" {
		return AuditSeverityMedium
	}

	switch eventType {
	case AuditDataDelete:
		return AuditSeverityMedium
	case AuditDataUpdate:
		return AuditSeverityLow
	default:
		return AuditSeverityLow
	}
}

// getSeverityForProjectEvent determines severity for project events
func (al *AuditLogger) getSeverityForProjectEvent(eventType AuditEventType, result string) AuditSeverity {
	if result != "SUCCESS" {
		return AuditSeverityMedium
	}

	switch eventType {
	case AuditProjectDelete:
		return AuditSeverityMedium
	case AuditProjectGenerate:
		return AuditSeverityLow
	default:
		return AuditSeverityLow
	}
}

// Close closes the audit logger and its file handle
func (al *AuditLogger) Close() error {
	if al.file != nil {
		return al.file.Close()
	}
	return nil
}

// Global audit logger instance
var globalAuditLogger *AuditLogger

// InitAuditLogger initializes the global audit logger
func InitAuditLogger(filePath string) error {
	var err error
	globalAuditLogger, err = NewAuditLogger(filePath)
	return err
}

// GetAuditLogger returns the global audit logger
func GetAuditLogger() *AuditLogger {
	if globalAuditLogger == nil {
		// Fallback to a default audit logger if not initialized
		logger, err := NewAuditLogger("./logs/audit.log")
		if err != nil {
			// If we can't create the audit logger, log the error and return nil
			GetLogger().Error("Failed to create fallback audit logger", err)
			return nil
		}
		globalAuditLogger = logger
	}
	return globalAuditLogger
}

// AuditMiddleware creates a middleware that logs audit events for all requests
func AuditMiddleware() fiber.Handler {
	auditLogger := GetAuditLogger()
	if auditLogger == nil {
		// If audit logger is not available, return a no-op middleware
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		// Get user ID if available
		var userID *uint
		if uid := c.Locals("user_id"); uid != nil {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// Determine event type based on method and status
		var eventType AuditEventType
		var result string

		if statusCode >= 200 && statusCode < 300 {
			result = "SUCCESS"
		} else {
			result = "FAILURE"
		}

		switch c.Method() {
		case "GET":
			eventType = AuditDataRead
		case "POST":
			eventType = AuditDataCreate
		case "PUT", "PATCH":
			eventType = AuditDataUpdate
		case "DELETE":
			eventType = AuditDataDelete
		default:
			eventType = AuditDataRead
		}

		// Log the audit event
		event := AuditEvent{
			EventType:  eventType,
			Severity:   AuditSeverityLow,
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
			Duration:   &[]int64{duration.Milliseconds()}[0],
		}

		if err != nil {
			event.ErrorMessage = err.Error()
			event.Severity = AuditSeverityMedium
		}

		auditLogger.LogEvent(event)

		return err
	}
}