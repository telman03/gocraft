# Error Handling and Logging Integration Guide

This guide demonstrates how to integrate the comprehensive error handling and logging system into your application.

## Overview

The new error handling and logging system provides:

1. **Structured Error Responses** - Consistent, standardized error responses
2. **Comprehensive Logging** - Structured logging with multiple levels and contexts
3. **Audit Logging** - Security-focused audit trails
4. **Performance Monitoring** - Real-time performance metrics and alerts
5. **Metrics Collection** - Request/response metrics and analytics

## Quick Start

### 1. Initialize the Systems

```go
package main

import (
    "log"
    "os"
    
    "github.com/gofiber/fiber/v2"
    "github.com/telman03/ai-backend-generator/internal/errors"
    "github.com/telman03/ai-backend-generator/internal/logging"
    "github.com/telman03/ai-backend-generator/internal/middleware"
    "github.com/telman03/ai-backend-generator/internal/monitoring"
)

func main() {
    // Initialize logging
    logging.InitLogger("gocraft")
    
    // Initialize audit logging
    if err := logging.InitAuditLogger("./logs/audit.log"); err != nil {
        log.Fatal("Failed to initialize audit logger:", err)
    }
    
    // Initialize monitoring
    monitoring.InitMetrics()
    monitoring.InitPerformanceMonitor()
    
    // Create Fiber app
    app := fiber.New(fiber.Config{
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            // This will be handled by our error middleware
            return err
        },
    })
    
    // Add middleware in the correct order
    app.Use(middleware.RequestIDMiddleware())
    app.Use(middleware.PanicRecoveryMiddleware())
    app.Use(middleware.ErrorHandlerMiddleware())
    app.Use(middleware.ComprehensiveMonitoringMiddleware())
    app.Use(logging.LoggingMiddleware())
    app.Use(logging.AuditMiddleware())
    
    // Add health check and metrics endpoints
    app.Use(middleware.HealthCheckMiddleware())
    app.Use(middleware.MetricsEndpointMiddleware())
    
    // Your routes here
    setupRoutes(app)
    
    log.Fatal(app.Listen(":8080"))
}
```

### 2. Using Structured Errors in Handlers

```go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/telman03/ai-backend-generator/internal/errors"
    "github.com/telman03/ai-backend-generator/internal/logging"
)

func GetProjectHandler(c *fiber.Ctx) error {
    logger := logging.GetLogger().WithComponent("project_handler")
    requestLogger := logger.NewRequestLogger(c)
    
    // Get user ID from context
    userID, err := getUserIDFromContext(c)
    if err != nil {
        // Return structured authentication error
        appErr := errors.NewAuthError("User authentication required").
            WithRequestID(c.Get("X-Request-ID", "")).
            WithDetails(err.Error())
        
        return errors.SendError(c, appErr)
    }
    
    // Get project ID from params
    projectID, err := getProjectIDFromParams(c)
    if err != nil {
        // Return structured validation error
        appErr := errors.NewInvalidInputError("project_id").
            WithRequestID(c.Get("X-Request-ID", "")).
            WithUserID(userID).
            WithDetails(err.Error())
        
        return errors.SendError(c, appErr)
    }
    
    // Log the access attempt
    requestLogger.LogProjectAccess(projectID, "view_project", true, "")
    
    // Your business logic here...
    project, err := getProject(userID, projectID)
    if err != nil {
        // Return structured not found error
        appErr := errors.NewProjectNotFoundError().
            WithRequestID(c.Get("X-Request-ID", "")).
            WithUserID(userID).
            WithContext("project_id", projectID).
            WithInternalError(err)
        
        return errors.SendError(c, appErr)
    }
    
    // Log successful operation
    requestLogger.LogUserAction("view_project", "project", true, "")
    
    return c.JSON(project)
}
```

### 3. Custom Error Handling

```go
// Create custom errors for your domain
func NewProjectGenerationError(reason string) *errors.AppError {
    return errors.NewAppError(
        errors.ErrCodeGenerationFailed,
        "Project generation failed",
        500,
    ).WithDetails(reason)
}

// Use in handlers
func GenerateProjectHandler(c *fiber.Ctx) error {
    // ... validation logic ...
    
    err := generateProject(config)
    if err != nil {
        appErr := NewProjectGenerationError(err.Error()).
            WithRequestID(c.Get("X-Request-ID", "")).
            WithUserID(userID).
            WithContext("config", config)
        
        return errors.SendError(c, appErr)
    }
    
    return c.JSON(fiber.Map{"success": true})
}
```

### 4. Validation Error Handling

```go
func CreateProjectHandler(c *fiber.Ctx) error {
    var req CreateProjectRequest
    
    if err := c.BodyParser(&req); err != nil {
        // Return structured validation error
        fieldErrors := map[string][]string{
            "body": {"Invalid JSON format"},
        }
        return errors.SendValidationError(c, "Request validation failed", fieldErrors)
    }
    
    // Validate the request
    if validationErr := validateCreateProjectRequest(&req); validationErr != nil {
        return errors.SendValidationError(c, "Validation failed", validationErr)
    }
    
    // Continue with business logic...
}

func validateCreateProjectRequest(req *CreateProjectRequest) map[string][]string {
    fieldErrors := make(map[string][]string)
    
    if req.Name == "" {
        fieldErrors["name"] = append(fieldErrors["name"], "Project name is required")
    }
    
    if len(req.Name) > 50 {
        fieldErrors["name"] = append(fieldErrors["name"], "Project name must be 50 characters or less")
    }
    
    if req.Framework == "" {
        fieldErrors["framework"] = append(fieldErrors["framework"], "Framework is required")
    }
    
    if len(fieldErrors) > 0 {
        return fieldErrors
    }
    
    return nil
}
```

### 5. Performance Monitoring

```go
func SlowOperationHandler(c *fiber.Ctx) error {
    performanceMonitor := monitoring.GetPerformanceMonitor()
    
    start := time.Now()
    
    // Your slow operation here
    result, err := performSlowOperation()
    
    duration := time.Since(start)
    success := err == nil
    
    // Record the operation performance
    performanceMonitor.RecordOperation("slow_operation", duration, success)
    
    if err != nil {
        return errors.NewOperationFailedError("slow_operation").
            WithInternalError(err)
    }
    
    return c.JSON(result)
}
```

### 6. Security Event Logging

```go
func LoginHandler(c *fiber.Ctx) error {
    auditLogger := logging.GetAuditLogger()
    
    var req LoginRequest
    if err := c.BodyParser(&req); err != nil {
        // Log failed login attempt
        auditLogger.LogAuthenticationEvent(
            logging.AuditLoginFailure,
            c,
            nil,
            "FAILURE",
            "Invalid request format",
        )
        
        return errors.NewValidationError("Invalid request format")
    }
    
    user, err := authenticateUser(req.Email, req.Password)
    if err != nil {
        // Log failed authentication
        auditLogger.LogAuthenticationEvent(
            logging.AuditLoginFailure,
            c,
            nil,
            "FAILURE",
            "Invalid credentials",
        )
        
        return errors.NewAuthError("Invalid credentials")
    }
    
    // Log successful authentication
    auditLogger.LogAuthenticationEvent(
        logging.AuditLoginSuccess,
        c,
        &user.ID,
        "SUCCESS",
        "User logged in successfully",
    )
    
    return c.JSON(fiber.Map{
        "token": generateJWT(user),
        "user":  user,
    })
}
```

## Configuration

### Environment Variables

```bash
# Logging configuration
LOG_LEVEL=INFO                    # DEBUG, INFO, WARN, ERROR, FATAL
AUDIT_LOG_PATH=./logs/audit.log   # Path to audit log file

# Monitoring configuration
METRICS_ENABLED=true              # Enable metrics collection
PERFORMANCE_MONITORING=true       # Enable performance monitoring
SLOW_REQUEST_THRESHOLD=500ms      # Threshold for slow request logging

# Error handling configuration
ENABLE_STACK_TRACE=false          # Enable stack traces in errors (dev only)
ENABLE_DETAILED_ERRORS=false      # Enable detailed error messages (dev only)
```

### Custom Configuration

```go
// Custom error handler configuration
errorConfig := middleware.ErrorHandlerConfig{
    EnableStackTrace:     os.Getenv("ENV") == "development",
    EnableDetailedErrors: os.Getenv("ENV") == "development",
    Logger:              logging.GetLogger().WithComponent("error_handler"),
}
app.Use(middleware.ErrorHandlerMiddleware(errorConfig))

// Custom monitoring configuration
monitoringConfig := middleware.MonitoringConfig{
    EnableMetrics:               true,
    EnableAuditLogging:          true,
    EnablePerformanceMonitoring: true,
    SlowRequestThreshold:        1 * time.Second,
    SkipPaths:                   []string{"/health", "/ping", "/metrics", "/favicon.ico"},
}
app.Use(middleware.ComprehensiveMonitoringMiddleware(monitoringConfig))
```

## API Response Examples

### Successful Response
```json
{
    "success": true,
    "data": {
        "id": 123,
        "name": "My Project",
        "framework": "gin"
    },
    "timestamp": "2024-01-15T10:30:00Z"
}
```

### Error Response
```json
{
    "success": false,
    "error": {
        "code": "PROJECT_NOT_FOUND",
        "message": "Project not found or access denied",
        "details": "No project found with ID 123 for user 456"
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_1705312200000",
    "context": {
        "project_id": 123,
        "user_id": 456
    }
}
```

### Validation Error Response
```json
{
    "success": false,
    "error": {
        "code": "VALIDATION_FAILED",
        "message": "Request validation failed",
        "field_count": 2
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "request_id": "req_1705312200000",
    "fields": {
        "name": ["Project name is required"],
        "framework": ["Framework must be one of: gin, echo, fiber"]
    }
}
```

## Monitoring Endpoints

### Health Check
```bash
GET /health
```

Response:
```json
{
    "status": "healthy",
    "timestamp": "2024-01-15T10:30:00Z",
    "uptime": "2h30m15s",
    "system": {
        "memory_usage_percent": 45.2,
        "goroutine_count": 25,
        "heap_size_mb": 128.5
    },
    "metrics": {
        "total_requests": 1250,
        "error_rate": 2.4,
        "avg_response_time_ms": 125.8,
        "active_connections": 5
    }
}
```

### Metrics Endpoint
```bash
GET /metrics
```

Response includes comprehensive performance and request metrics.

## Best Practices

1. **Always use structured errors** instead of plain fiber errors
2. **Include request IDs** in all error responses for tracing
3. **Log security events** for authentication and authorization failures
4. **Monitor performance** of critical operations
5. **Use appropriate log levels** (DEBUG for development, INFO for production)
6. **Include context** in errors and logs for better debugging
7. **Sanitize sensitive data** before logging
8. **Set up alerts** based on error rates and performance metrics

This system provides comprehensive observability and error handling for your application while maintaining security and performance.