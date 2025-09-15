package logging

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
	LevelFatal LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   time.Time              `json:"timestamp"`
	Level       LogLevel               `json:"level"`
	Message     string                 `json:"message"`
	Service     string                 `json:"service"`
	Component   string                 `json:"component,omitempty"`
	Operation   string                 `json:"operation,omitempty"`
	UserID      *uint                  `json:"user_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Duration    *int64                 `json:"duration_ms,omitempty"`
	StatusCode  *int                   `json:"status_code,omitempty"`
	Method      string                 `json:"method,omitempty"`
	Path        string                 `json:"path,omitempty"`
	IP          string                 `json:"ip,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Error       *ErrorDetails          `json:"error,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Caller      *CallerInfo            `json:"caller,omitempty"`
}

// ErrorDetails contains error-specific information
type ErrorDetails struct {
	Type       string `json:"type"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`
}

// CallerInfo contains information about the code location
type CallerInfo struct {
	File     string `json:"file"`
	Function string `json:"function"`
	Line     int    `json:"line"`
}

// Logger provides structured logging capabilities
type Logger struct {
	serviceName string
	component   string
	minLevel    LogLevel
	output      *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(serviceName, component string) *Logger {
	return &Logger{
		serviceName: serviceName,
		component:   component,
		minLevel:    getLogLevelFromEnv(),
		output:      log.New(os.Stdout, "", 0),
	}
}

// getLogLevelFromEnv gets the minimum log level from environment
func getLogLevelFromEnv() LogLevel {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch level {
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// shouldLog determines if a log entry should be written based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
		LevelFatal: 4,
	}
	return levels[level] >= levels[l.minLevel]
}

// getCaller gets information about the calling function
func getCaller(skip int) *CallerInfo {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return nil
	}

	function := runtime.FuncForPC(pc).Name()
	// Extract just the function name from the full path
	if idx := strings.LastIndex(function, "/"); idx >= 0 {
		function = function[idx+1:]
	}
	if idx := strings.LastIndex(function, "."); idx >= 0 {
		function = function[idx+1:]
	}

	// Extract just the filename from the full path
	if idx := strings.LastIndex(file, "/"); idx >= 0 {
		file = file[idx+1:]
	}

	return &CallerInfo{
		File:     file,
		Function: function,
		Line:     line,
	}
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Service:   l.serviceName,
		Component: l.component,
		Context:   fields,
		Caller:    getCaller(3), // Skip log, Debug/Info/Warn/Error, and the calling function
	}

	// Marshal to JSON
	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.output.Printf("[%s] %s: %s - %v", level, l.serviceName, message, err)
		return
	}

	l.output.Println(string(jsonData))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var context map[string]interface{}
	if len(fields) > 0 {
		context = fields[0]
	}
	l.log(LevelDebug, message, context)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var context map[string]interface{}
	if len(fields) > 0 {
		context = fields[0]
	}
	l.log(LevelInfo, message, context)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var context map[string]interface{}
	if len(fields) > 0 {
		context = fields[0]
	}
	l.log(LevelWarn, message, context)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	var context map[string]interface{}
	if len(fields) > 0 {
		context = fields[0]
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LevelError,
		Message:   message,
		Service:   l.serviceName,
		Component: l.component,
		Context:   context,
		Caller:    getCaller(2),
	}

	if err != nil {
		entry.Error = &ErrorDetails{
			Type:    fmt.Sprintf("%T", err),
			Message: err.Error(),
		}
	}

	// Marshal to JSON
	jsonData, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		// Fallback to simple logging if JSON marshaling fails
		l.output.Printf("[ERROR] %s: %s - %v", l.serviceName, message, err)
		return
	}

	l.output.Println(string(jsonData))
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, fields ...map[string]interface{}) {
	l.Error(message, err, fields...)
	os.Exit(1)
}

// WithComponent creates a new logger with a specific component
func (l *Logger) WithComponent(component string) *Logger {
	return &Logger{
		serviceName: l.serviceName,
		component:   component,
		minLevel:    l.minLevel,
		output:      l.output,
	}
}

// RequestLogger provides HTTP request-specific logging
type RequestLogger struct {
	*Logger
	requestID string
	userID    *uint
	startTime time.Time
}

// NewRequestLogger creates a request-specific logger
func (l *Logger) NewRequestLogger(c *fiber.Ctx) *RequestLogger {
	requestID := c.Get("X-Request-ID", "")
	if requestID == "" {
		requestID = c.Get("X-Trace-Id", "")
	}

	var userID *uint
	if uid := c.Locals("user_id"); uid != nil {
		if id, ok := uid.(uint); ok {
			userID = &id
		}
	}

	return &RequestLogger{
		Logger:    l,
		requestID: requestID,
		userID:    userID,
		startTime: time.Now(),
	}
}

// LogRequest logs HTTP request details
func (rl *RequestLogger) LogRequest(c *fiber.Ctx) {
	if !rl.shouldLog(LevelInfo) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     LevelInfo,
		Message:   "HTTP Request",
		Service:   rl.serviceName,
		Component: "http",
		Operation: "request",
		UserID:    rl.userID,
		RequestID: rl.requestID,
		Method:    c.Method(),
		Path:      c.Path(),
		IP:        c.IP(),
		UserAgent: c.Get("User-Agent"),
		Context: map[string]interface{}{
			"query_params": c.Queries(),
			"headers":      getFilteredHeaders(c),
		},
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		rl.output.Printf("[INFO] HTTP Request: %s %s", c.Method(), c.Path())
		return
	}

	rl.output.Println(string(jsonData))
}

// LogResponse logs HTTP response details
func (rl *RequestLogger) LogResponse(c *fiber.Ctx, statusCode int) {
	if !rl.shouldLog(LevelInfo) {
		return
	}

	duration := time.Since(rl.startTime).Milliseconds()

	entry := LogEntry{
		Timestamp:  time.Now(),
		Level:      LevelInfo,
		Message:    "HTTP Response",
		Service:    rl.serviceName,
		Component:  "http",
		Operation:  "response",
		UserID:     rl.userID,
		RequestID:  rl.requestID,
		Duration:   &duration,
		StatusCode: &statusCode,
		Method:     c.Method(),
		Path:       c.Path(),
		IP:         c.IP(),
		Context: map[string]interface{}{
			"response_size": len(c.Response().Body()),
		},
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		rl.output.Printf("[INFO] HTTP Response: %d %dms", statusCode, duration)
		return
	}

	rl.output.Println(string(jsonData))
}

// LogUserAction logs user-specific actions
func (rl *RequestLogger) LogUserAction(action, resource string, success bool, details string) {
	level := LevelInfo
	if !success {
		level = LevelWarn
	}

	context := map[string]interface{}{
		"action":   action,
		"resource": resource,
		"success":  success,
	}

	if details != "" {
		context["details"] = details
	}

	message := fmt.Sprintf("User action: %s on %s", action, resource)
	if !success {
		message += " (failed)"
	}

	rl.log(level, message, context)
}

// LogProjectAccess logs project-specific access
func (rl *RequestLogger) LogProjectAccess(projectID uint, action string, success bool, reason string) {
	level := LevelInfo
	if !success {
		level = LevelWarn
	}

	context := map[string]interface{}{
		"project_id": projectID,
		"action":     action,
		"success":    success,
	}

	if reason != "" {
		context["reason"] = reason
	}

	message := fmt.Sprintf("Project access: %s on project %d", action, projectID)
	if !success {
		message += " (failed)"
	}

	rl.log(level, message, context)
}

// LogSecurityEvent logs security-related events
func (rl *RequestLogger) LogSecurityEvent(event, severity, details string) {
	level := LevelWarn
	if severity == "high" || severity == "critical" {
		level = LevelError
	}

	context := map[string]interface{}{
		"event":    event,
		"severity": severity,
		"details":  details,
	}

	message := fmt.Sprintf("Security event: %s (%s)", event, severity)
	rl.log(level, message, context)
}

// LogPerformanceMetric logs performance-related metrics
func (rl *RequestLogger) LogPerformanceMetric(operation string, duration time.Duration, metadata map[string]interface{}) {
	if !rl.shouldLog(LevelInfo) {
		return
	}

	durationMs := duration.Milliseconds()
	context := map[string]interface{}{
		"operation":   operation,
		"duration_ms": durationMs,
	}

	// Add metadata if provided
	for k, v := range metadata {
		context[k] = v
	}

	message := fmt.Sprintf("Performance metric: %s completed in %dms", operation, durationMs)
	rl.log(LevelInfo, message, context)
}

// getFilteredHeaders returns filtered HTTP headers (excluding sensitive ones)
func getFilteredHeaders(c *fiber.Ctx) map[string]string {
	headers := make(map[string]string)
	sensitiveHeaders := map[string]bool{
		"authorization": true,
		"cookie":        true,
		"x-api-key":     true,
		"x-auth-token":  true,
	}

	c.Request().Header.VisitAll(func(key, value []byte) {
		keyStr := strings.ToLower(string(key))
		if !sensitiveHeaders[keyStr] {
			headers[keyStr] = string(value)
		}
	})

	return headers
}

// Global logger instance
var defaultLogger *Logger

// InitLogger initializes the global logger
func InitLogger(serviceName string) {
	defaultLogger = NewLogger(serviceName, "")
}

// GetLogger returns the global logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		defaultLogger = NewLogger("gocraft", "")
	}
	return defaultLogger
}

// Convenience functions for global logging
func Debug(message string, fields ...map[string]interface{}) {
	GetLogger().Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	GetLogger().Info(message, fields...)
}

func Warn(message string, fields ...map[string]interface{}) {
	GetLogger().Warn(message, fields...)
}

func Error(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Error(message, err, fields...)
}

func Fatal(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Fatal(message, err, fields...)
}

// LoggingMiddleware creates a Fiber middleware for request/response logging
func LoggingMiddleware() fiber.Handler {
	logger := GetLogger().WithComponent("middleware")

	return func(c *fiber.Ctx) error {
		// Skip logging for health check endpoints
		if strings.HasPrefix(c.Path(), "/health") || strings.HasPrefix(c.Path(), "/ping") {
			return c.Next()
		}

		requestLogger := logger.NewRequestLogger(c)
		
		// Log the incoming request
		requestLogger.LogRequest(c)

		// Process the request
		err := c.Next()

		// Log the response
		requestLogger.LogResponse(c, c.Response().StatusCode())

		return err
	}
}