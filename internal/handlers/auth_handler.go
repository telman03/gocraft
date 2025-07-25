package handlers

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/telman03/ai-backend-generator/internal/auth"
	"github.com/telman03/ai-backend-generator/internal/database"
	"github.com/telman03/ai-backend-generator/internal/models"
	"golang.org/x/crypto/bcrypt"
)



// RegisterHandler godoc
// @Summary Reguster user and return JWT
// @Description Accepts credentials and returns a JWT token
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

	hashedPwd, err := auth.HashPassword(body.Password)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Password hash failed")
	}

	user := models.User{Email: body.Email, Password: hashedPwd}
	if err := database.DB.Create(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Email already exists")
	}

	token, err := auth.GenerateJWT(user.ID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(fiber.Map{"token": token})
}

// LoginHandler godoc
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Incorrect password"})
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Println("Missing JWT_SECRET")
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // Token valid for 7 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Println("Token sign error:", err)
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": signedToken})
}
