package errors

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Authentication & Authorization Errors
	ErrCodeAuthFailed           ErrorCode = "AUTH_FAILED"
	ErrCodeInvalidToken         ErrorCode = "INVALID_TOKEN"
	ErrCodeTokenExpired         ErrorCode = "TOKEN_EXPIRED"
	ErrCodeInsufficientPerms    ErrorCode = "INSUFFICIENT_PERMISSIONS"
	ErrCodeAccessDenied         ErrorCode = "ACCESS_DENIED"

	// Validation Errors
	ErrCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
	ErrCodeInvalidInput         ErrorCode = "INVALID_INPUT"
	ErrCodeInvalidFormat        ErrorCode = "INVALID_FORMAT"
	ErrCodeMissingRequired      ErrorCode = "MISSING_REQUIRED_FIELD"
	ErrCodeInvalidRange         ErrorCode = "INVALID_RANGE"

	// Resource Errors
	ErrCodeResourceNotFound     ErrorCode = "RESOURCE_NOT_FOUND"
	ErrCodeProjectNotFound      ErrorCode = "PROJECT_NOT_FOUND"
	ErrCodeUserNotFound         ErrorCode = "USER_NOT_FOUND"
	ErrCodeResourceExists       ErrorCode = "RESOURCE_ALREADY_EXISTS"
	ErrCodeResourceConflict     ErrorCode = "RESOURCE_CONFLICT"

	// File Operation Errors
	ErrCodeFileNotFound         ErrorCode = "FILE_NOT_FOUND"
	ErrCodeFileExpired          ErrorCode = "FILE_EXPIRED"
	ErrCodeFileNotAvailable     ErrorCode = "FILE_NOT_AVAILABLE"
	ErrCodeInvalidFilePath      ErrorCode = "INVALID_FILE_PATH"
	ErrCodeFileOperationFailed  ErrorCode = "FILE_OPERATION_FAILED"
	ErrCodeFileTooLarge         ErrorCode = "FILE_TOO_LARGE"

	// Database Errors
	ErrCodeDatabaseError        ErrorCode = "DATABASE_ERROR"
	ErrCodeQueryFailed          ErrorCode = "QUERY_FAILED"
	ErrCodeTransactionFailed    ErrorCode = "TRANSACTION_FAILED"
	ErrCodeConnectionFailed     ErrorCode = "CONNECTION_FAILED"

	// Business Logic Errors
	ErrCodeGenerationFailed     ErrorCode = "GENERATION_FAILED"
	ErrCodeInsufficientData     ErrorCode = "INSUFFICIENT_DATA"
	ErrCodeOperationFailed      ErrorCode = "OPERATION_FAILED"
	ErrCodeServiceUnavailable   ErrorCode = "SERVICE_UNAVAILABLE"

	// Rate Limiting & Quota Errors
	ErrCodeRateLimitExceeded    ErrorCode = "RATE_LIMIT_EXCEEDED"
	ErrCodeQuotaExceeded        ErrorCode = "QUOTA_EXCEEDED"
	ErrCodeTooManyRequests      ErrorCode = "TOO_MANY_REQUESTS"

	// System Errors
	ErrCodeInternalError        ErrorCode = "INTERNAL_ERROR"
	ErrCodeServiceError         ErrorCode = "SERVICE_ERROR"
	ErrCodeTimeoutError         ErrorCode = "TIMEOUT_ERROR"
	ErrCodeConfigurationError   ErrorCode = "CONFIGURATION_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code        ErrorCode              `json:"code"`
	Message     string                 `json:"message"`
	Details     string                 `json:"details,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      *uint                  `json:"user_id,omitempty"`
	StatusCode  int                    `json:"-"`
	InternalErr error                  `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s - %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithDetails adds additional details to the error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithUserID adds user ID to the error for tracking
func (e *AppError) WithUserID(userID uint) *AppError {
	e.UserID = &userID
	return e
}

// WithRequestID adds request ID for tracing
func (e *AppError) WithRequestID(requestID string) *AppError {
	e.RequestID = requestID
	return e
}

// WithInternalError adds the underlying error for debugging
func (e *AppError) WithInternalError(err error) *AppError {
	e.InternalErr = err
	if e.Details == "" && err != nil {
		e.Details = err.Error()
	}
	return e
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string, statusCode int) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Timestamp:  time.Now(),
	}
}

// Predefined error constructors for common scenarios

// Authentication Errors
func NewAuthError(message string) *AppError {
	return NewAppError(ErrCodeAuthFailed, message, http.StatusUnauthorized)
}

func NewInvalidTokenError() *AppError {
	return NewAppError(ErrCodeInvalidToken, "Invalid or expired authentication token", http.StatusUnauthorized)
}

func NewAccessDeniedError(resource string) *AppError {
	return NewAppError(ErrCodeAccessDenied, fmt.Sprintf("Access denied to %s", resource), http.StatusForbidden)
}

// Validation Errors
func NewValidationError(message string) *AppError {
	return NewAppError(ErrCodeValidationFailed, message, http.StatusBadRequest)
}

func NewInvalidInputError(field string) *AppError {
	return NewAppError(ErrCodeInvalidInput, fmt.Sprintf("Invalid input for field: %s", field), http.StatusBadRequest)
}

func NewMissingRequiredFieldError(field string) *AppError {
	return NewAppError(ErrCodeMissingRequired, fmt.Sprintf("Missing required field: %s", field), http.StatusBadRequest)
}

// Resource Errors
func NewResourceNotFoundError(resource string) *AppError {
	return NewAppError(ErrCodeResourceNotFound, fmt.Sprintf("%s not found", resource), http.StatusNotFound)
}

func NewProjectNotFoundError() *AppError {
	return NewAppError(ErrCodeProjectNotFound, "Project not found or access denied", http.StatusNotFound)
}

func NewResourceExistsError(resource string) *AppError {
	return NewAppError(ErrCodeResourceExists, fmt.Sprintf("%s already exists", resource), http.StatusConflict)
}

// File Operation Errors
func NewFileNotFoundError(filename string) *AppError {
	return NewAppError(ErrCodeFileNotFound, fmt.Sprintf("File not found: %s", filename), http.StatusNotFound)
}

func NewFileExpiredError() *AppError {
	return NewAppError(ErrCodeFileExpired, "File has expired and is no longer available", http.StatusGone).
		WithContext("can_regenerate", true)
}

func NewFileNotAvailableError(reason string) *AppError {
	return NewAppError(ErrCodeFileNotAvailable, "File is not available", http.StatusNotFound).
		WithDetails(reason)
}

func NewInvalidFilePathError() *AppError {
	return NewAppError(ErrCodeInvalidFilePath, "Invalid file path provided", http.StatusBadRequest)
}

// Database Errors
func NewDatabaseError(operation string) *AppError {
	return NewAppError(ErrCodeDatabaseError, fmt.Sprintf("Database error during %s", operation), http.StatusInternalServerError)
}

func NewQueryFailedError(query string) *AppError {
	return NewAppError(ErrCodeQueryFailed, "Database query failed", http.StatusInternalServerError).
		WithContext("query_type", query)
}

// Business Logic Errors
func NewGenerationFailedError(reason string) *AppError {
	return NewAppError(ErrCodeGenerationFailed, "Project generation failed", http.StatusInternalServerError).
		WithDetails(reason)
}

func NewInsufficientDataError(missing string) *AppError {
	return NewAppError(ErrCodeInsufficientData, "Insufficient data to complete operation", http.StatusBadRequest).
		WithDetails(fmt.Sprintf("Missing: %s", missing))
}

func NewOperationFailedError(operation string) *AppError {
	return NewAppError(ErrCodeOperationFailed, fmt.Sprintf("Operation failed: %s", operation), http.StatusInternalServerError)
}

// Rate Limiting Errors
func NewRateLimitExceededError(retryAfter string) *AppError {
	return NewAppError(ErrCodeRateLimitExceeded, "Rate limit exceeded", http.StatusTooManyRequests).
		WithContext("retry_after", retryAfter)
}

// System Errors
func NewInternalError(message string) *AppError {
	return NewAppError(ErrCodeInternalError, message, http.StatusInternalServerError)
}

func NewServiceUnavailableError(service string) *AppError {
	return NewAppError(ErrCodeServiceUnavailable, fmt.Sprintf("Service temporarily unavailable: %s", service), http.StatusServiceUnavailable)
}

func NewTimeoutError(operation string) *AppError {
	return NewAppError(ErrCodeTimeoutError, fmt.Sprintf("Operation timed out: %s", operation), http.StatusRequestTimeout)
}

// ErrorResponse represents the standardized API error response format
type ErrorResponse struct {
	Success   bool                   `json:"success"`
	Error     ErrorInfo              `json:"error"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// ErrorInfo contains the core error information
type ErrorInfo struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

// ValidationErrorResponse represents validation-specific error response
type ValidationErrorResponse struct {
	Success   bool                     `json:"success"`
	Error     ValidationErrorInfo      `json:"error"`
	Timestamp time.Time                `json:"timestamp"`
	RequestID string                   `json:"request_id,omitempty"`
	Fields    map[string][]string      `json:"fields,omitempty"`
}

// ValidationErrorInfo contains validation error information
type ValidationErrorInfo struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Count   int       `json:"field_count"`
}

// SendError sends a standardized error response
func SendError(c *fiber.Ctx, appErr *AppError) error {
	// Extract request ID from context if available
	requestID := c.Get("X-Request-ID", "")
	if requestID == "" {
		requestID = c.Get("X-Trace-Id", "")
	}

	response := ErrorResponse{
		Success:   false,
		Error: ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		Timestamp: time.Now(),
		RequestID: requestID,
		Context:   appErr.Context,
	}

	return c.Status(appErr.StatusCode).JSON(response)
}

// SendValidationError sends a validation-specific error response
func SendValidationError(c *fiber.Ctx, message string, fieldErrors map[string][]string) error {
	requestID := c.Get("X-Request-ID", "")
	if requestID == "" {
		requestID = c.Get("X-Trace-Id", "")
	}

	response := ValidationErrorResponse{
		Success: false,
		Error: ValidationErrorInfo{
			Code:    ErrCodeValidationFailed,
			Message: message,
			Count:   len(fieldErrors),
		},
		Timestamp: time.Now(),
		RequestID: requestID,
		Fields:    fieldErrors,
	}

	return c.Status(http.StatusBadRequest).JSON(response)
}

// IsAppError checks if an error is an AppError
func IsAppError(err error) (*AppError, bool) {
	if appErr, ok := err.(*AppError); ok {
		return appErr, true
	}
	return nil, false
}

// WrapError wraps a regular error into an AppError
func WrapError(err error, code ErrorCode, message string, statusCode int) *AppError {
	return NewAppError(code, message, statusCode).WithInternalError(err)
}

// FromFiberError converts a Fiber error to an AppError
func FromFiberError(err error) *AppError {
	if fiberErr, ok := err.(*fiber.Error); ok {
		var code ErrorCode
		switch fiberErr.Code {
		case 400:
			code = ErrCodeInvalidInput
		case 401:
			code = ErrCodeAuthFailed
		case 403:
			code = ErrCodeAccessDenied
		case 404:
			code = ErrCodeResourceNotFound
		case 409:
			code = ErrCodeResourceConflict
		case 429:
			code = ErrCodeRateLimitExceeded
		case 500:
			code = ErrCodeInternalError
		case 503:
			code = ErrCodeServiceUnavailable
		default:
			code = ErrCodeInternalError
		}
		
		return NewAppError(code, fiberErr.Message, fiberErr.Code).WithInternalError(err)
	}
	
	return NewInternalError("Unknown error occurred").WithInternalError(err)
}