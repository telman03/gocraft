package utils

import (
	"fmt"
	"html"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/errors"
)

var validate *validator.Validate

// Regex patterns for additional validation
var (
	// Project name validation - alphanumeric, hyphens, underscores
	projectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	
	// Email validation (more strict than validator package)
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	
	// Dangerous characters that should be escaped
	dangerousCharsRegex = regexp.MustCompile(`[<>\"'&]`)
)

func init() {
	validate = validator.New()
	
	// Register custom validators
	validate.RegisterValidation("project_name", validateProjectName)
	validate.RegisterValidation("safe_string", validateSafeString)
	validate.RegisterValidation("no_html", validateNoHTML)
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidateStruct validates a struct and returns standardized error response
func ValidateStruct(s interface{}) *ErrorResponse {
	if err := validate.Struct(s); err != nil {
		details := make(map[string]string)
		
		for _, err := range err.(validator.ValidationErrors) {
			field := strings.ToLower(err.Field())
			switch err.Tag() {
			case "required":
				details[field] = fmt.Sprintf("%s is required", field)
			case "email":
				details[field] = "Invalid email format"
			case "min":
				details[field] = fmt.Sprintf("%s must be at least %s characters", field, err.Param())
			case "max":
				details[field] = fmt.Sprintf("%s must be at most %s characters", field, err.Param())
			case "len":
				details[field] = fmt.Sprintf("%s must be exactly %s characters", field, err.Param())
			case "numeric":
				details[field] = fmt.Sprintf("%s must contain only numbers", field)
			case "alphanum":
				details[field] = fmt.Sprintf("%s must contain only letters and numbers", field)
			default:
				details[field] = fmt.Sprintf("%s is invalid", field)
			}
		}
		
		return &ErrorResponse{
			Error:   "Validation failed",
			Message: "Please check the provided data",
			Details: details,
		}
	}
	return nil
}

// SendErrorResponse sends a standardized error response using the new error system
func SendErrorResponse(c *fiber.Ctx, status int, message string, details ...map[string]string) error {
	// Map HTTP status codes to appropriate error codes
	var errorCode errors.ErrorCode
	switch status {
	case 400:
		errorCode = errors.ErrCodeInvalidInput
	case 401:
		errorCode = errors.ErrCodeAuthFailed
	case 403:
		errorCode = errors.ErrCodeAccessDenied
	case 404:
		errorCode = errors.ErrCodeResourceNotFound
	case 409:
		errorCode = errors.ErrCodeResourceConflict
	case 429:
		errorCode = errors.ErrCodeRateLimitExceeded
	case 500:
		errorCode = errors.ErrCodeInternalError
	default:
		errorCode = errors.ErrCodeInternalError
	}

	appErr := errors.NewAppError(errorCode, message, status)
	
	// Add details as context if provided
	if len(details) > 0 {
		for key, value := range details[0] {
			appErr = appErr.WithContext(key, value)
		}
	}

	return errors.SendError(c, appErr)
}

// SendValidationError sends a validation error response using the new error system
func SendValidationError(c *fiber.Ctx, validationErr *ErrorResponse) error {
	fieldErrors := make(map[string][]string)
	
	// Convert details to field errors format
	if validationErr.Details != nil {
		for field, message := range validationErr.Details {
			fieldErrors[field] = []string{message}
		}
	}

	return errors.SendValidationError(c, validationErr.Error, fieldErrors)
}

// Custom validation functions
func validateProjectName(fl validator.FieldLevel) bool {
	projectName := fl.Field().String()
	return projectNameRegex.MatchString(projectName) && len(projectName) >= 1 && len(projectName) <= 50
}

func validateSafeString(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	return !dangerousCharsRegex.MatchString(str)
}

func validateNoHTML(fl validator.FieldLevel) bool {
	str := fl.Field().String()
	return !strings.Contains(str, "<") && !strings.Contains(str, ">")
}

// SanitizeInput provides comprehensive input sanitization
func SanitizeInput(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")
	
	// Trim whitespace
	sanitized = strings.TrimSpace(sanitized)
	
	// HTML escape dangerous characters
	sanitized = html.EscapeString(sanitized)
	
	// Remove control characters except tab, newline, and carriage return
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// SanitizeProjectName sanitizes project names specifically
func SanitizeProjectName(projectName string) string {
	// Remove leading/trailing spaces
	sanitized := strings.TrimSpace(projectName)
	
	// Replace spaces with hyphens
	sanitized = strings.ReplaceAll(sanitized, " ", "-")
	
	// Remove or replace dangerous characters
	dangerous := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t", "@", "#", "$", "%", "^", "&", "(", ")", "+", "=", "[", "]", "{", "}", ";", "'", ","}
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "")
	}
	
	// Convert to lowercase
	sanitized = strings.ToLower(sanitized)
	
	// Remove multiple consecutive hyphens
	for strings.Contains(sanitized, "--") {
		sanitized = strings.ReplaceAll(sanitized, "--", "-")
	}
	
	// Remove leading/trailing hyphens
	sanitized = strings.Trim(sanitized, "-")
	
	// Ensure name is not empty and within length limits
	if sanitized == "" {
		sanitized = "project"
	}
	
	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
		sanitized = strings.Trim(sanitized, "-")
	}
	
	return sanitized
}

// ValidateEmail validates email format more strictly
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// SanitizeFilePath sanitizes file paths to prevent directory traversal
func SanitizeFilePath(filePath string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(filePath, "\x00", "")
	
	// Remove path traversal patterns
	sanitized = strings.ReplaceAll(sanitized, "../", "")
	sanitized = strings.ReplaceAll(sanitized, "..\\", "")
	sanitized = strings.ReplaceAll(sanitized, "..", "")
	
	// Remove leading slashes and backslashes
	sanitized = strings.TrimLeft(sanitized, "/\\")
	
	// Replace backslashes with forward slashes
	sanitized = strings.ReplaceAll(sanitized, "\\", "/")
	
	// Remove multiple consecutive slashes
	for strings.Contains(sanitized, "//") {
		sanitized = strings.ReplaceAll(sanitized, "//", "/")
	}
	
	return sanitized
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// ContainsSQLKeywords checks if input contains common SQL keywords
func ContainsSQLKeywords(input string) bool {
	sqlKeywords := []string{
		"select", "insert", "update", "delete", "drop", "create", "alter", "truncate",
		"union", "exec", "execute", "script", "javascript", "vbscript",
	}
	
	lowerInput := strings.ToLower(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(lowerInput, keyword) {
			return true
		}
	}
	return false
}

// ValidateAndSanitizeStruct validates and sanitizes a struct
func ValidateAndSanitizeStruct(s interface{}) *ErrorResponse {
	// First validate the struct
	if validationErr := ValidateStruct(s); validationErr != nil {
		return validationErr
	}
	
	// Additional security checks could be added here
	// For now, we rely on the field-level validation tags
	
	return nil
}