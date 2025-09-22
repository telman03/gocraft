package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/utils"
)

// RequireAdmin middleware ensures the user has admin role
func RequireAdmin(c *fiber.Ctx) error {
	// Get user ID from context (set by RequireAuth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Authentication required", map[string]string{
			"code": "AUTH_REQUIRED",
		})
	}

	// Convert user ID to uint
	var userIDUint uint
	switch v := userID.(type) {
	case uint:
		userIDUint = v
	case int:
		userIDUint = uint(v)
	case float64:
		userIDUint = uint(v)
	default:
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid user ID format", map[string]string{
			"code": "INVALID_USER_ID",
		})
	}

	// Get user from database to check role
	var user models.User
	if err := database.DB.First(&user, userIDUint).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not found", map[string]string{
			"code": "USER_NOT_FOUND",
		})
	}

	// Check if user has admin role
	if !user.IsAdmin() {
		LogUserAction(c, userIDUint, "admin_access_denied", "admin", false, "Insufficient permissions")
		return utils.SendErrorResponse(c, fiber.StatusForbidden, "Admin access required", map[string]string{
			"code": "ADMIN_REQUIRED",
			"message": "You need admin privileges to access this resource",
		})
	}

	// Log successful admin access
	LogUserAction(c, userIDUint, "admin_access_granted", "admin", true, "")

	return c.Next()
}