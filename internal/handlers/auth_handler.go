package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/gocraft-backend/internal/auth"
	"github.com/telman03/gocraft-backend/internal/database"
	"github.com/telman03/gocraft-backend/internal/models"
	"github.com/telman03/gocraft-backend/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// Register RegisterHandler godoc
// @Summary Register user and send OTP verification
// @Description Accepts credentials, creates user, and sends OTP verification email
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.RegisterInput true "User Registration Request"
// @Router /auth/register [post]
func Register(c *fiber.Ctx) error {
	var body models.RegisterInput
	if err := c.BodyParser(&body); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&body); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	// Start database transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if user already exists
	var existingUser models.User
	result := tx.Where("email = ?", body.Email).First(&existingUser)
	if result.Error == nil {
		// If user exists but is not verified, allow re-registration
		if !existingUser.IsVerified {
			// Delete the unverified user and create a new one
			if err := tx.Delete(&existingUser).Error; err != nil {
				tx.Rollback()
				return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to process registration")
			}
		} else {
			tx.Rollback()
			return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Email already exists")
		}
	}

	hashedPwd, err := auth.HashPassword(body.Password)
	if err != nil {
		tx.Rollback()
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Password processing failed")
	}

	// Generate OTP
	otp := utils.GenerateOTP()
	otpExpiresAt := time.Now().Add(10 * time.Minute)

	// Create user with OTP, unverified
	user := models.User{
		Email:        body.Email,
		Password:     hashedPwd,
		OTP:          otp,
		OTPExpiresAt: otpExpiresAt,
		IsVerified:   false,
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to create user")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to complete registration")
	}

	// Send OTP email
	go func() {
		if err := utils.SendOTPEmail(body.Email, otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
		}
	}()

	return c.JSON(fiber.Map{
		"message": "Registration initiated. Please check your email for verification code.",
		"email":   body.Email,
	})
}

// VerifyOTP godoc
// @Summary Verify OTP code
// @Description Verify the OTP code sent during registration
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.VerifyOTPInput true "OTP Verification Request"
// @Router /auth/verify-otp [post]
func VerifyOTP(c *fiber.Ctx) error {
	var body models.VerifyOTPInput
	if err := c.BodyParser(&body); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&body); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "User not found")
	}

	// Check if OTP is expired
	if time.Now().After(user.OTPExpiresAt) {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "OTP expired. Please request a new one.")
	}

	// Check if OTP matches
	if user.OTP != body.OTP {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid OTP code.")
	}

	// Mark user as verified
	user.IsVerified = true
	user.OTP = "" // Clear the OTP
	if err := database.DB.Save(&user).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to verify user")
	}

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate token")
	}

	// Send welcome email
	go func() {
		if err := utils.SendWelcomeEmail(user.Email); err != nil {
			log.Printf("Failed to send welcome email: %v", err)
		}
	}()

	return c.JSON(fiber.Map{
		"message": "Account verified successfully",
		"token":   token,
	})
}

// ResendOTP godoc
// @Summary Resend OTP code
// @Description Resend the OTP code for verification
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.ResendOTPInput true "Resend OTP Request"
// @Router /auth/resend-otp [post]
func ResendOTP(c *fiber.Ctx) error {
	var body models.ResendOTPInput
	if err := c.BodyParser(&body); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&body); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "User not found")
	}

	// Don't allow resending OTP for already verified users
	if user.IsVerified {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "User is already verified")
	}

	// Generate new OTP
	otp := utils.GenerateOTP()
	otpExpiresAt := time.Now().Add(10 * time.Minute)

	// Update user with new OTP
	user.OTP = otp
	user.OTPExpiresAt = otpExpiresAt
	database.DB.Save(&user)

	// Send OTP email
	go func() {
		if err := utils.SendOTPEmail(body.Email, otp); err != nil {
			log.Printf("Failed to send OTP email: %v", err)
		}
	}()

	return c.JSON(fiber.Map{
		"message": "OTP code resent. Please check your email.",
		"email":   body.Email,
	})
}

// Login LoginHandler godoc
// @Summary Authenticate user and return JWT
// @Description Accepts credentials and returns a JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body models.LoginInput true "User Login Request"
// @Router /auth/login [post]
func Login(c *fiber.Ctx) error {
	var req models.LoginInput
	if err := c.BodyParser(&req); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusBadRequest, "Invalid request format")
	}

	// Validate input
	if validationErr := utils.ValidateStruct(&req); validationErr != nil {
		return utils.SendValidationError(c, validationErr)
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Check if the user is verified
	if !user.IsVerified {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Email not verified. Please verify your email first.", map[string]string{
			"email": user.Email,
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Use the centralized token generation function
	signedToken, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Println("Token sign error:", err)
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Failed to generate authentication token")
	}

	return c.JSON(fiber.Map{"token": signedToken})
}

// GetCurrentUser godoc
// @Summary Get current authenticated user profile
// @Description Returns user profile information including email, projects count, and joined date
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "User profile information"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/me [get]
func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	if userID == nil {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	// Convert userID to uint (JWT claims parse numbers as float64)
	var uid uint
	switch v := userID.(type) {
	case float64:
		uid = uint(v)
	case uint:
		uid = v
	case int:
		uid = uint(v)
	default:
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Invalid user ID format")
	}

	// Fetch user from database
	var user models.User
	if err := database.DB.Where("id = ?", uid).First(&user).Error; err != nil {
		return utils.SendErrorResponse(c, fiber.StatusNotFound, "User not found")
	}

	// Count user's projects
	var projectCount int64
	if err := database.DB.Model(&models.ProjectHistory{}).Where("user_id = ?", uid).Count(&projectCount).Error; err != nil {
		log.Printf("Failed to count user projects: %v", err)
		projectCount = 0 // Default to 0 if count fails
	}

	return c.JSON(fiber.Map{
		"id":              user.ID,
		"email":           user.Email,
		"role":            user.Role,
		"is_verified":     user.IsVerified,
		"projects_count":  projectCount,
		"joined_date":     user.CreatedAt,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
	})
}
