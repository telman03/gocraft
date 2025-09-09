package handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/telman03/ai-backend-generator/internal/auth"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/models"
	"github.com/telman03/ai-backend-generator/internal/utils"
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
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	// Check if user already exists
	var existingUser models.User
	result := database.DB.Where("email = ?", body.Email).First(&existingUser)
	if result.Error == nil {
		// If user exists but is not verified, allow re-registration
		if !existingUser.IsVerified {
			// Delete the unverified user and create a new one
			database.DB.Delete(&existingUser)
		} else {
			return fiber.NewError(fiber.StatusBadRequest, "Email already exists")
		}
	}

	hashedPwd, err := auth.HashPassword(body.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Password hash failed")
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

	if err := database.DB.Create(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
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
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Check if OTP is expired
	if time.Now().After(user.OTPExpiresAt) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "OTP expired. Please request a new one.",
		})
	}

	// Check if OTP matches
	if user.OTP != body.OTP {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid OTP code.",
		})
	}

	// Mark user as verified
	user.IsVerified = true
	user.OTP = "" // Clear the OTP
	database.DB.Save(&user)

	// Generate JWT token
	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate token")
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
		return fiber.NewError(fiber.StatusBadRequest, "Invalid input")
	}

	var user models.User
	if err := database.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// Don't allow resending OTP for already verified users
	if user.IsVerified {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User is already verified",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	var user models.User
	if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	// Check if the user is verified
	if !user.IsVerified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Email not verified. Please verify your email first.",
			"email": user.Email,
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect password"})
	}

	// Use the centralized token generation function
	signedToken, err := auth.GenerateJWT(user.ID)
	if err != nil {
		log.Println("Token sign error:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": signedToken})
}

// GetCurrentUser godoc
// @Summary Get current authenticated user
// @Description Returns user ID and email from JWT
// @Tags Auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Router /auth/me [get]
func GetCurrentUser(c *fiber.Ctx) error {
	userID := c.Locals("user_id")

	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	return c.JSON(fiber.Map{
		"id": userID,
	})
}
