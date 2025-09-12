package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
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

// SendErrorResponse sends a standardized error response
func SendErrorResponse(c *fiber.Ctx, status int, message string, details ...map[string]string) error {
	response := ErrorResponse{
		Error:   message,
		Message: "Please check your request and try again",
	}
	
	if len(details) > 0 {
		response.Details = details[0]
	}
	
	return c.Status(status).JSON(response)
}

// SendValidationError sends a validation error response
func SendValidationError(c *fiber.Ctx, validationErr *ErrorResponse) error {
	return c.Status(fiber.StatusBadRequest).JSON(validationErr)
}