package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/utils"
	"github.com/telman03/ai-backend-generator/internal/validation"
)

// ValidateFeatures godoc
// @Summary Validate feature combinations for conflicts
// @Description Validates selected features against conflict rules and returns adjusted feature set
// @Tags Validation
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param data body models.GenerateRequest true "Features to validate"
// @Router /validate [post]
func ValidateFeatures(c *fiber.Ctx) error {
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	var req models.GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Basic input validation
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Merge framework into features if provided separately
	allFeatures := req.Features
	if req.Framework != "" {
		// Check if framework is already in features
		frameworkExists := false
		for _, feature := range req.Features {
			if strings.EqualFold(feature, req.Framework) {
				frameworkExists = true
				break
			}
		}
		// Add framework to features if not already present
		if !frameworkExists {
			allFeatures = append([]string{req.Framework}, req.Features...)
		}
	}

	// Validate template conflicts
	validator := validation.NewTemplateValidator()
	result := validator.ValidateFeatures(allFeatures)

	// Return validation result
	response := map[string]interface{}{
		"project_name":        req.ProjectName,
		"original_features":   req.Features,
		"validation_result":   result,
		"supported_features":  validator.GetSupportedFeatures(),
	}

	if !result.IsValid {
		return c.Status(fiber.StatusBadRequest).JSON(response)
	}

	return c.JSON(response)
}

// GetSupportedFeatures godoc
// @Summary Get all supported features organized by category
// @Description Returns all available features organized by category with descriptions
// @Tags Validation
// @Accept json
// @Produce json
// @Router /features [get]
func GetSupportedFeatures(c *fiber.Ctx) error {
	validator := validation.NewTemplateValidator()
	
	response := map[string]interface{}{
		"categories": validator.GetSupportedFeatures(),
		"descriptions": map[string]string{
			// Databases
			"mysql":      "MySQL relational database",
			"postgresql": "PostgreSQL relational database", 
			"sqlite":     "SQLite embedded database",
			"mongodb":    "MongoDB NoSQL database",
			"redis":      "Redis in-memory cache/store",
			"badger":     "BadgerDB embedded key-value store",
			
			// Frameworks
			"gin":   "Gin web framework - fast and minimalist",
			"echo":  "Echo web framework - high performance",
			"fiber": "Fiber web framework - Express-inspired",
			
			// ORMs
			"gorm": "GORM - full-featured ORM with migrations",
			"sqlc": "SQLC - type-safe SQL code generation",
			
			// Auth
			"auth":   "JWT authentication system",
			"oauth2": "OAuth2 integration for external providers",
			
			// AI
			"openai":     "OpenAI API integration",
			"openrouter": "OpenRouter API integration", 
			"claude":     "Claude API integration",
			
			// Communication
			"grpc":      "gRPC server implementation",
			"websocket": "WebSocket server for real-time communication",
			"proto":     "Protocol Buffers definitions",
			
			// DevOps
			"dockerfile":      "Docker containerization",
			"docker-compose": "Multi-service Docker setup",
			"makefile":       "Build automation scripts",
			
			// Documentation
			"swagger": "Swagger/OpenAPI documentation",
			"postman": "Postman collection generation",
			"readme":  "Project README documentation",
			
			// Utilities
			"logger":      "Logging configuration",
			"middleware":  "HTTP middleware (CORS, auth, etc.)",
			"config":      "Application configuration management",
			"migrations":  "Database migration utilities",
			
			// Core (always included)
			"env":       "Environment variables configuration",
			"gitignore": "Git ignore rules",
			"main":      "Application entry point",
		},
		"conflict_rules": map[string]interface{}{
			"databases": map[string]string{
				"rule": "Only one relational database allowed",
				"allowed": "MySQL OR PostgreSQL OR SQLite (+ optional MongoDB/Redis)",
				"forbidden": "Multiple relational databases together",
			},
			"frameworks": map[string]string{
				"rule": "Only one web framework allowed",
				"allowed": "Gin OR Echo OR Fiber",
				"forbidden": "Multiple web frameworks together",
			},
			"orms": map[string]string{
				"rule": "Only one ORM allowed",
				"allowed": "GORM OR SQLC",
				"forbidden": "GORM and SQLC together",
			},
			"auth": map[string]string{
				"rule": "Multiple auth methods allowed but discouraged",
				"allowed": "JWT and OAuth2 can coexist as complementary",
				"warning": "Ensure proper configuration when using both",
			},
			"dependencies": map[string]string{
				"grpc_requires_proto": "gRPC automatically includes Protocol Buffers",
				"gorm_includes_migrations": "GORM automatically includes migration utilities",
				"auth_includes_middleware": "Authentication automatically includes middleware",
			},
		},
	}

	return c.JSON(response)
}