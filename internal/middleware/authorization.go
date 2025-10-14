package middleware

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/database"
	"github.com/telman03/gocraft-backend/internal/models"
)

// AuthorizationError represents authorization-specific error codes
type AuthorizationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Common authorization error responses
var (
	ErrProjectNotFound = AuthorizationError{
		Code:    "PROJECT_NOT_FOUND",
		Message: "Project not found or access denied",
		Details: "The requested project does not exist or you don't have permission to access it",
	}
	
	ErrInsufficientPermissions = AuthorizationError{
		Code:    "INSUFFICIENT_PERMISSIONS",
		Message: "Insufficient permissions",
		Details: "You don't have permission to perform this action",
	}
	
	ErrInvalidUserContext = AuthorizationError{
		Code:    "INVALID_USER_CONTEXT",
		Message: "Invalid user context",
		Details: "User authentication context is invalid",
	}
	
	ErrResourceAccessDenied = AuthorizationError{
		Code:    "RESOURCE_ACCESS_DENIED",
		Message: "Access denied",
		Details: "You don't have permission to access this resource",
	}
)

// ProjectOwnershipValidator validates that a user owns a specific project
func ProjectOwnershipValidator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			logAuthorizationFailure(c, "missing_user_context", "No user context found")
			return sendAuthorizationError(c, fiber.StatusUnauthorized, ErrInvalidUserContext)
		}

		// Convert userID to uint
		userIDFloat, ok := userID.(float64)
		if !ok {
			logAuthorizationFailure(c, "invalid_user_id_format", fmt.Sprintf("Invalid user ID type: %T", userID))
			return sendAuthorizationError(c, fiber.StatusUnauthorized, ErrInvalidUserContext)
		}
		userIDUint := uint(userIDFloat)

		// Extract project ID from URL parameter
		projectIDStr := c.Params("id")
		if projectIDStr == "" {
			logAuthorizationFailure(c, "missing_project_id", "No project ID in URL parameters")
			return sendAuthorizationError(c, fiber.StatusBadRequest, AuthorizationError{
				Code:    "MISSING_PROJECT_ID",
				Message: "Project ID is required",
				Details: "Project ID must be provided in the URL path",
			})
		}

		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			logAuthorizationFailure(c, "invalid_project_id_format", fmt.Sprintf("Invalid project ID: %s", projectIDStr))
			return sendAuthorizationError(c, fiber.StatusBadRequest, AuthorizationError{
				Code:    "INVALID_PROJECT_ID",
				Message: "Invalid project ID format",
				Details: "Project ID must be a valid number",
			})
		}

		// Validate project ownership
		if err := validateProjectOwnership(userIDUint, uint(projectID)); err != nil {
			logAuthorizationFailure(c, "ownership_validation_failed", err.Error())
			return sendAuthorizationError(c, fiber.StatusNotFound, ErrProjectNotFound)
		}

		// Store validated project ID in context for downstream handlers
		c.Locals("validated_project_id", uint(projectID))
		c.Locals("validated_user_id", userIDUint)
		
		// Log successful authorization
		logAuthorizationSuccess(c, userIDUint, uint(projectID))
		
		return c.Next()
	}
}

// UserContextValidator ensures user context is properly set and valid
func UserContextValidator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract user ID from context (set by auth middleware)
		userID := c.Locals("user_id")
		if userID == nil {
			logAuthorizationFailure(c, "missing_user_context", "No user context found")
			return sendAuthorizationError(c, fiber.StatusUnauthorized, ErrInvalidUserContext)
		}

		// Convert and validate userID format
		userIDFloat, ok := userID.(float64)
		if !ok {
			logAuthorizationFailure(c, "invalid_user_id_format", fmt.Sprintf("Invalid user ID type: %T", userID))
			return sendAuthorizationError(c, fiber.StatusUnauthorized, ErrInvalidUserContext)
		}
		userIDUint := uint(userIDFloat)

		// Validate user ID is positive
		if userIDUint == 0 {
			logAuthorizationFailure(c, "invalid_user_id_value", "User ID cannot be zero")
			return sendAuthorizationError(c, fiber.StatusUnauthorized, ErrInvalidUserContext)
		}

		// Store validated user ID in context
		c.Locals("validated_user_id", userIDUint)
		
		return c.Next()
	}
}

// validateProjectOwnership checks if a user owns a specific project
func validateProjectOwnership(userID, projectID uint) error {
	var project models.ProjectHistory
	
	result := database.DB.Where("id = ? AND user_id = ?", projectID, userID).First(&project)
	if result.Error != nil {
		return fmt.Errorf("project not found or access denied")
	}
	
	return nil
}

// ValidateProjectOwnership is a utility function for manual ownership validation
func ValidateProjectOwnership(userID, projectID uint) error {
	return validateProjectOwnership(userID, projectID)
}

// GetValidatedUserID extracts the validated user ID from context
func GetValidatedUserID(c *fiber.Ctx) (uint, error) {
	userID := c.Locals("validated_user_id")
	if userID == nil {
		// Fallback to extracting from user_id if validated_user_id is not set
		rawUserID := c.Locals("user_id")
		if rawUserID == nil {
			return 0, fmt.Errorf("no user context found")
		}
		
		userIDFloat, ok := rawUserID.(float64)
		if !ok {
			return 0, fmt.Errorf("invalid user ID format")
		}
		
		return uint(userIDFloat), nil
	}
	
	userIDUint, ok := userID.(uint)
	if !ok {
		return 0, fmt.Errorf("invalid validated user ID format")
	}
	
	return userIDUint, nil
}

// GetValidatedProjectID extracts the validated project ID from context
func GetValidatedProjectID(c *fiber.Ctx) (uint, error) {
	projectID := c.Locals("validated_project_id")
	if projectID == nil {
		return 0, fmt.Errorf("no validated project ID found")
	}
	
	projectIDUint, ok := projectID.(uint)
	if !ok {
		return 0, fmt.Errorf("invalid validated project ID format")
	}
	
	return projectIDUint, nil
}

// sendAuthorizationError sends a standardized authorization error response
func sendAuthorizationError(c *fiber.Ctx, status int, authErr AuthorizationError) error {
	return c.Status(status).JSON(fiber.Map{
		"error":     authErr.Message,
		"code":      authErr.Code,
		"details":   authErr.Details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// logAuthorizationFailure logs authorization failures for security monitoring
func logAuthorizationFailure(c *fiber.Ctx, reason, details string) {
	log.Printf("[AUTHORIZATION_FAILURE] IP: %s, Path: %s, Method: %s, Reason: %s, Details: %s", 
		c.IP(), c.Path(), c.Method(), reason, details)
}

// logAuthorizationSuccess logs successful authorization for audit purposes
func logAuthorizationSuccess(c *fiber.Ctx, userID, projectID uint) {
	log.Printf("[AUTHORIZATION_SUCCESS] IP: %s, Path: %s, Method: %s, UserID: %d, ProjectID: %d", 
		c.IP(), c.Path(), c.Method(), userID, projectID)
}

// AuditLogger logs security events for monitoring and compliance
type AuditEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	UserID      uint      `json:"user_id,omitempty"`
	ProjectID   uint      `json:"project_id,omitempty"`
	Action      string    `json:"action"`
	Resource    string    `json:"resource"`
	IP          string    `json:"ip"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Success     bool      `json:"success"`
	ErrorCode   string    `json:"error_code,omitempty"`
	Details     string    `json:"details,omitempty"`
}

// LogSecurityEvent logs security-related events for audit purposes
func LogSecurityEvent(c *fiber.Ctx, event AuditEvent) {
	event.Timestamp = time.Now()
	event.IP = c.IP()
	event.UserAgent = c.Get("User-Agent")
	
	// In a production environment, you might want to send this to a dedicated audit log system
	log.Printf("[SECURITY_AUDIT] %+v", event)
}

// LogProjectAccess logs project access events
func LogProjectAccess(c *fiber.Ctx, userID, projectID uint, action string, success bool, errorCode string) {
	LogSecurityEvent(c, AuditEvent{
		UserID:    userID,
		ProjectID: projectID,
		Action:    action,
		Resource:  "project_history",
		Success:   success,
		ErrorCode: errorCode,
	})
}

// LogUserAction logs general user actions
func LogUserAction(c *fiber.Ctx, userID uint, action, resource string, success bool, details string) {
	LogSecurityEvent(c, AuditEvent{
		UserID:   userID,
		Action:   action,
		Resource: resource,
		Success:  success,
		Details:  details,
	})
}