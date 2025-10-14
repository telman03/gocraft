package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/database"
	"github.com/telman03/gocraft-backend/internal/middleware"
	"github.com/telman03/gocraft-backend/internal/models"
	"github.com/telman03/gocraft-backend/internal/utils"
)

// GetAllUsers godoc
// @Summary Get all users (Admin only)
// @Description Retrieves a list of all users with their roles and basic information
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Items per page (default: 20, max: 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/users [get]
func GetAllUsers(c *fiber.Ctx) error {
	// Parse pagination parameters
	page := 1
	pageSize := 20

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := c.Query("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	var total int64
	if err := database.DB.Model(&models.User{}).Count(&total).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to count users", map[string]string{
			"details": err.Error(),
		})
	}

	// Get users with pagination
	var users []models.User
	if err := database.DB.Select("id, email, role, is_verified, created_at, updated_at").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch users", map[string]string{
			"details": err.Error(),
		})
	}

	// Calculate total pages
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	// Get admin user ID for logging
	adminID, _ := middleware.GetValidatedUserID(c)
	middleware.LogUserAction(c, adminID, "get_all_users", "admin", true, "")

	return c.JSON(map[string]interface{}{
		"users":       users,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// UpdateUserRole godoc
// @Summary Update user role (Admin only)
// @Description Updates a user's role between 'user' and 'admin'
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param data body UpdateUserRoleRequest true "Role update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/users/{id}/role [put]
func UpdateUserRole(c *fiber.Ctx) error {
	// Get user ID from path
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid user ID", map[string]string{
			"code": "INVALID_USER_ID",
		})
	}

	// Parse request body
	var req UpdateUserRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format", map[string]string{
			"details": err.Error(),
		})
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Get the user to update
	var user models.User
	if err := database.DB.First(&user, uint(userID)).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found", map[string]string{
			"code": "USER_NOT_FOUND",
		})
	}

	// Get admin user ID for logging
	adminID, _ := middleware.GetValidatedUserID(c)

	// Prevent admin from demoting themselves
	if adminID == uint(userID) && req.Role != string(models.UserRoleAdmin) {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Cannot change your own admin role", map[string]string{
			"code": "CANNOT_DEMOTE_SELF",
		})
	}

	// Update user role
	oldRole := user.Role
	user.Role = models.UserRole(req.Role)

	if err := database.DB.Save(&user).Error; err != nil {
		middleware.LogUserAction(c, adminID, "update_user_role", "admin", false, err.Error())
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to update user role", map[string]string{
			"details": err.Error(),
		})
	}

	// Log successful role update
	middleware.LogUserAction(c, adminID, "update_user_role", "admin", true, 
		"Updated user "+user.Email+" role from "+string(oldRole)+" to "+string(user.Role))

	return c.JSON(map[string]interface{}{
		"message":  "User role updated successfully",
		"user_id":  user.ID,
		"email":    user.Email,
		"old_role": oldRole,
		"new_role": user.Role,
	})
}

// GetUserStats godoc
// @Summary Get user statistics (Admin only)
// @Description Retrieves comprehensive statistics about users and system usage
// @Tags Admin
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/admin/stats [get]
func GetUserStats(c *fiber.Ctx) error {
	// Get total users count
	var totalUsers int64
	if err := database.DB.Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to count users", map[string]string{
			"details": err.Error(),
		})
	}

	// Get verified users count
	var verifiedUsers int64
	if err := database.DB.Model(&models.User{}).Where("is_verified = ?", true).Count(&verifiedUsers).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to count verified users", map[string]string{
			"details": err.Error(),
		})
	}

	// Get admin users count
	var adminUsers int64
	if err := database.DB.Model(&models.User{}).Where("role = ?", models.UserRoleAdmin).Count(&adminUsers).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to count admin users", map[string]string{
			"details": err.Error(),
		})
	}

	// Get total projects count
	var totalProjects int64
	if err := database.DB.Model(&models.ProjectHistory{}).Count(&totalProjects).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to count projects", map[string]string{
			"details": err.Error(),
		})
	}

	// Get recent users (last 30 days)
	var recentUsers int64
	if err := database.DB.Model(&models.User{}).Where("created_at > NOW() - INTERVAL '30 days'").Count(&recentUsers).Error; err != nil {
		// Fallback for databases that don't support INTERVAL
		if err := database.DB.Model(&models.User{}).Where("created_at > datetime('now', '-30 days')").Count(&recentUsers).Error; err != nil {
			recentUsers = 0 // Set to 0 if query fails
		}
	}

	// Get admin user ID for logging
	adminID, _ := middleware.GetValidatedUserID(c)
	middleware.LogUserAction(c, adminID, "get_user_stats", "admin", true, "")

	return c.JSON(map[string]interface{}{
		"total_users":     totalUsers,
		"verified_users":  verifiedUsers,
		"admin_users":     adminUsers,
		"regular_users":   totalUsers - adminUsers,
		"recent_users":    recentUsers,
		"total_projects":  totalProjects,
		"verification_rate": func() float64 {
			if totalUsers == 0 {
				return 0
			}
			return float64(verifiedUsers) / float64(totalUsers) * 100
		}(),
	})
}

// UpdateUserRoleRequest represents the request to update a user's role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=user admin"`
}