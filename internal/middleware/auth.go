package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/telman03/ai-backend-generator/internal/auth"
	"github.com/telman03/ai-backend-generator/internal/utils"
)

func RequireAuth(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid Authorization header format")
	}
	tokenStr := parts[1]

	// Use the auth package's GetJWTSecret function
	jwtSecret := auth.GetJWTSecret()

	if len(jwtSecret) == 0 {
		fmt.Println("[DEBUG] JWT_SECRET is not set")
		return utils.SendErrorResponse(c, fiber.StatusInternalServerError, "Server misconfiguration")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		fmt.Printf("[DEBUG] JWT parse error: %v\n", err)
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("[DEBUG] Invalid token claims")
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token claims")
	}

	userID, ok := claims["sub"]
	if !ok {
		fmt.Println("[DEBUG] sub claim missing")
		return utils.SendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token: sub claim missing")
	}

	c.Locals("user_id", userID)
	return c.Next()
}
