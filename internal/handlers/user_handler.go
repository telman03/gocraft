package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/models"
)

// GetUserProfile godoc
// @Summary Get user profile
// @Description Returns the profile information for the authenticated user
// @Tags User
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} models.UserProfileResponse
// @Failure 401 {object} map[string]string "User not authenticated"
// @Failure 404 {object} map[string]string "User not found"
// @Router /user/profile [get]
func GetUserProfile(c *fiber.Ctx) error {
	// Get user_id from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Query the database for the user
	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Create profile response
	profile := models.UserProfileResponse{
		ID:         user.ID,
		Email:      user.Email,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}

	// Return the user profile information
	return c.JSON(profile)
}
