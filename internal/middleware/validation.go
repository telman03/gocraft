package middleware

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ValidationError represents validation-specific error codes
type ValidationError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Common validation error responses
var (
	ErrInvalidInput = ValidationError{
		Code:    "INVALID_INPUT",
		Message: "Invalid input provided",
	}
	
	ErrSQLInjectionDetected = ValidationError{
		Code:    "SECURITY_VIOLATION",
		Message: "Invalid characters detected",
		Details: map[string]string{
			"reason": "Input contains potentially dangerous characters",
		},
	}
	
	ErrPathTraversalDetected = ValidationError{
		Code:    "SECURITY_VIOLATION",
		Message: "Invalid path detected",
		Details: map[string]string{
			"reason": "Path contains directory traversal patterns",
		},
	}
)

// SQL injection patterns to detect
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|update\s+.*\s+set|delete\s+from)`),
	regexp.MustCompile(`(?i)(drop\s+table|create\s+table|alter\s+table|truncate\s+table)`),
	regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\(|sp_executesql)`),
	regexp.MustCompile(`(?i)(script\s*>|javascript:|vbscript:|onload\s*=|onerror\s*=)`),
	regexp.MustCompile(`(?i)(\'\s*or\s+\d+\s*=\s*\d+|\'\s*or\s+\'.*\'=\'.*\')`),
	regexp.MustCompile(`(?i)(--\s|\/\*|\*\/|;\s*--)`),
}

// Path traversal patterns to detect
var pathTraversalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.\/|\.\.\\`),
	regexp.MustCompile(`\/\.\.\/|\\\.\.\\`),
	regexp.MustCompile(`%2e%2e%2f|%2e%2e%5c`),
	regexp.MustCompile(`%252e%252e%252f|%252e%252e%255c`),
}

// InputSanitizer provides comprehensive input validation and sanitization
type InputSanitizer struct{}

// NewInputSanitizer creates a new input sanitizer instance
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

// ValidateAndSanitizeQueryParams validates and sanitizes query parameters
func (s *InputSanitizer) ValidateAndSanitizeQueryParams() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get all query parameters
		queries := c.Queries()
		
		for key, value := range queries {
			// Skip empty values
			if value == "" {
				continue
			}
			
			// Validate against SQL injection
			if s.containsSQLInjection(value) {
				logSecurityViolation(c, "sql_injection_attempt", fmt.Sprintf("Parameter: %s, Value: %s", key, value))
				return sendValidationError(c, fiber.StatusBadRequest, ErrSQLInjectionDetected)
			}
			
			// Validate against path traversal
			if s.containsPathTraversal(value) {
				logSecurityViolation(c, "path_traversal_attempt", fmt.Sprintf("Parameter: %s, Value: %s", key, value))
				return sendValidationError(c, fiber.StatusBadRequest, ErrPathTraversalDetected)
			}
			
			// Sanitize the value
			sanitized := s.sanitizeString(value)
			
			// Update the query parameter with sanitized value
			c.Request().URI().QueryArgs().Set(key, sanitized)
		}
		
		return c.Next()
	}
}

// ValidateHistoryFilters validates specific history filter parameters
func (s *InputSanitizer) ValidateHistoryFilters() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Validate page parameter
		if pageStr := c.Query("page"); pageStr != "" {
			if page, err := strconv.Atoi(pageStr); err != nil || page < 1 || page > 10000 {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_PAGE",
					Message: "Invalid page parameter",
					Details: map[string]string{
						"page": "Page must be a number between 1 and 10000",
					},
				})
			}
		}
		
		// Validate page_size parameter
		if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
			if pageSize, err := strconv.Atoi(pageSizeStr); err != nil || pageSize < 1 || pageSize > 100 {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_PAGE_SIZE",
					Message: "Invalid page size parameter",
					Details: map[string]string{
						"page_size": "Page size must be a number between 1 and 100",
					},
				})
			}
		}
		
		// Validate search parameter length
		if search := c.Query("search"); search != "" {
			if len(search) > 100 {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_SEARCH",
					Message: "Search term too long",
					Details: map[string]string{
						"search": "Search term must be 100 characters or less",
					},
				})
			}
		}
		
		// Validate framework parameter
		if framework := c.Query("framework"); framework != "" {
			validFrameworks := []string{"gin", "echo", "fiber"}
			if !s.isValidFramework(framework, validFrameworks) {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_FRAMEWORK",
					Message: "Invalid framework parameter",
					Details: map[string]string{
						"framework": "Framework must be one of: gin, echo, fiber",
					},
				})
			}
		}
		
		// Validate frameworks parameter (comma-separated)
		if frameworks := c.Query("frameworks"); frameworks != "" {
			frameworkList := strings.Split(frameworks, ",")
			validFrameworks := []string{"gin", "echo", "fiber"}
			for _, fw := range frameworkList {
				fw = strings.TrimSpace(fw)
				if fw != "" && !s.isValidFramework(fw, validFrameworks) {
					return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
						Code:    "INVALID_FRAMEWORKS",
						Message: "Invalid frameworks parameter",
						Details: map[string]string{
							"frameworks": "All frameworks must be one of: gin, echo, fiber",
						},
					})
				}
			}
		}
		
		// Validate status parameter
		if status := c.Query("status"); status != "" {
			validStatuses := []string{"available", "expired", "deleted", "error"}
			if !s.isValidStatus(status, validStatuses) {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_STATUS",
					Message: "Invalid status parameter",
					Details: map[string]string{
						"status": "Status must be one of: available, expired, deleted, error",
					},
				})
			}
		}
		
		// Validate date parameters
		if dateFrom := c.Query("date_from"); dateFrom != "" {
			if !s.isValidDate(dateFrom) {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_DATE_FROM",
					Message: "Invalid date_from parameter",
					Details: map[string]string{
						"date_from": "Date must be in YYYY-MM-DD format",
					},
				})
			}
		}
		
		if dateTo := c.Query("date_to"); dateTo != "" {
			if !s.isValidDate(dateTo) {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_DATE_TO",
					Message: "Invalid date_to parameter",
					Details: map[string]string{
						"date_to": "Date must be in YYYY-MM-DD format",
					},
				})
			}
		}
		
		// Validate sort_by parameter
		if sortBy := c.Query("sort_by"); sortBy != "" {
			validSortFields := []string{"created_at", "project_name", "framework", "zip_file_size", "generation_duration_ms"}
			if !s.isValidSortField(sortBy, validSortFields) {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_SORT_BY",
					Message: "Invalid sort_by parameter",
					Details: map[string]string{
						"sort_by": "Sort field must be one of: created_at, project_name, framework, zip_file_size, generation_duration_ms",
					},
				})
			}
		}
		
		// Validate sort_order parameter
		if sortOrder := c.Query("sort_order"); sortOrder != "" {
			if sortOrder != "asc" && sortOrder != "desc" {
				return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
					Code:    "INVALID_SORT_ORDER",
					Message: "Invalid sort_order parameter",
					Details: map[string]string{
						"sort_order": "Sort order must be 'asc' or 'desc'",
					},
				})
			}
		}
		
		return c.Next()
	}
}

// ValidateProjectID validates project ID parameter
func (s *InputSanitizer) ValidateProjectID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		projectIDStr := c.Params("id")
		if projectIDStr == "" {
			return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
				Code:    "MISSING_PROJECT_ID",
				Message: "Project ID is required",
				Details: map[string]string{
					"id": "Project ID must be provided in the URL path",
				},
			})
		}
		
		// Validate project ID format
		projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil || projectID == 0 {
			return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
				Code:    "INVALID_PROJECT_ID",
				Message: "Invalid project ID format",
				Details: map[string]string{
					"id": "Project ID must be a positive number",
				},
			})
		}
		
		// Check for reasonable upper limit
		if projectID > 4294967295 { // uint32 max
			return sendValidationError(c, fiber.StatusBadRequest, ValidationError{
				Code:    "INVALID_PROJECT_ID",
				Message: "Project ID out of range",
				Details: map[string]string{
					"id": "Project ID is too large",
				},
			})
		}
		
		return c.Next()
	}
}

// ValidateJSONBody validates and sanitizes JSON request body
func (s *InputSanitizer) ValidateJSONBody() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only validate JSON bodies for POST/PUT/PATCH requests
		if c.Method() != "POST" && c.Method() != "PUT" && c.Method() != "PATCH" {
			return c.Next()
		}
		
		// Check Content-Type
		contentType := c.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return c.Next() // Skip validation for non-JSON bodies
		}
		
		// Get body
		body := c.Body()
		if len(body) == 0 {
			return c.Next() // Skip validation for empty bodies
		}
		
		// Check body size limit (1MB)
		if len(body) > 1024*1024 {
			return sendValidationError(c, fiber.StatusRequestEntityTooLarge, ValidationError{
				Code:    "BODY_TOO_LARGE",
				Message: "Request body too large",
				Details: map[string]string{
					"limit": "Request body must be less than 1MB",
				},
			})
		}
		
		// Validate against path traversal only — SQL injection patterns produce too many
		// false positives on legitimate project/feature names (e.g. "select", "delete")
		bodyStr := string(body)
		if s.containsPathTraversal(bodyStr) {
			logSecurityViolation(c, "path_traversal_attempt", "Request body contains path traversal patterns")
			return sendValidationError(c, fiber.StatusBadRequest, ErrPathTraversalDetected)
		}
		
		return c.Next()
	}
}

// containsSQLInjection checks if input contains SQL injection patterns
func (s *InputSanitizer) containsSQLInjection(input string) bool {
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// containsPathTraversal checks if input contains path traversal patterns
func (s *InputSanitizer) containsPathTraversal(input string) bool {
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// sanitizeString removes potentially dangerous characters from input
func (s *InputSanitizer) sanitizeString(input string) string {
	// Remove null bytes
	sanitized := strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters except tab, newline, and carriage return
	var result strings.Builder
	for _, r := range sanitized {
		if r >= 32 || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		}
	}
	
	return result.String()
}

// isValidFramework checks if framework is in the list of valid frameworks
func (s *InputSanitizer) isValidFramework(framework string, validFrameworks []string) bool {
	framework = strings.ToLower(strings.TrimSpace(framework))
	for _, valid := range validFrameworks {
		if framework == valid {
			return true
		}
	}
	return false
}

// isValidStatus checks if status is in the list of valid statuses
func (s *InputSanitizer) isValidStatus(status string, validStatuses []string) bool {
	status = strings.ToLower(strings.TrimSpace(status))
	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// isValidSortField checks if sort field is in the list of valid sort fields
func (s *InputSanitizer) isValidSortField(sortBy string, validSortFields []string) bool {
	sortBy = strings.ToLower(strings.TrimSpace(sortBy))
	for _, valid := range validSortFields {
		if sortBy == valid {
			return true
		}
	}
	return false
}

// isValidDate checks if date string is in YYYY-MM-DD format
func (s *InputSanitizer) isValidDate(dateStr string) bool {
	_, err := time.Parse("2006-01-02", dateStr)
	return err == nil
}

// sendValidationError sends a standardized validation error response
func sendValidationError(c *fiber.Ctx, status int, validationErr ValidationError) error {
	return c.Status(status).JSON(fiber.Map{
		"error":     validationErr.Message,
		"code":      validationErr.Code,
		"details":   validationErr.Details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// logSecurityViolation logs security violations for monitoring
func logSecurityViolation(c *fiber.Ctx, violationType, details string) {
	log.Printf("[SECURITY_VIOLATION] IP: %s, Path: %s, Method: %s, Type: %s, Details: %s", 
		c.IP(), c.Path(), c.Method(), violationType, details)
}

// RateLimiter provides basic rate limiting functionality
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit middleware for basic rate limiting
func (rl *RateLimiter) RateLimit() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		now := time.Now()

		rl.mu.Lock()

		// Evict expired timestamps
		if requests, exists := rl.requests[ip]; exists {
			var valid []time.Time
			for _, t := range requests {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			rl.requests[ip] = valid
		}

		// Check rate limit
		if len(rl.requests[ip]) >= rl.limit {
			rl.mu.Unlock()
			logSecurityViolation(c, "rate_limit_exceeded", fmt.Sprintf("IP: %s exceeded rate limit", ip))
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"code":        "RATE_LIMIT_EXCEEDED",
				"details":     fmt.Sprintf("Maximum %d requests per %v allowed", rl.limit, rl.window),
				"retry_after": rl.window.Seconds(),
			})
		}

		// Record current request
		rl.requests[ip] = append(rl.requests[ip], now)
		rl.mu.Unlock()
		
		return c.Next()
	}
}