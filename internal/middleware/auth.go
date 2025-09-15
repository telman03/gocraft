package middleware

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/telman03/ai-backend-generator/internal/auth"
)

// AuthError represents authentication-specific error codes
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Common authentication error responses
var (
	ErrMissingAuthHeader = AuthError{
		Code:    "MISSING_AUTH_HEADER",
		Message: "Authentication required",
		Details: "Authorization header is missing",
	}
	
	ErrInvalidAuthFormat = AuthError{
		Code:    "INVALID_AUTH_FORMAT",
		Message: "Invalid authentication format",
		Details: "Authorization header must be in format: Bearer <token>",
	}
	
	ErrInvalidToken = AuthError{
		Code:    "INVALID_TOKEN",
		Message: "Invalid or expired token",
		Details: "Please login again to get a valid token",
	}
	
	ErrInvalidClaims = AuthError{
		Code:    "INVALID_TOKEN_CLAIMS",
		Message: "Invalid token claims",
		Details: "Token does not contain required user information",
	}
	
	ErrServerMisconfiguration = AuthError{
		Code:    "SERVER_ERROR",
		Message: "Authentication service unavailable",
		Details: "Please try again later",
	}
)

// RequireAuth validates JWT tokens and extracts user context
func RequireAuth(c *fiber.Ctx) error {
	// Get Authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		logAuthFailure(c, "missing_auth_header", "No Authorization header provided")
		return sendAuthError(c, fiber.StatusUnauthorized, ErrMissingAuthHeader)
	}

	// Parse Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		logAuthFailure(c, "invalid_auth_format", fmt.Sprintf("Invalid format: %s", authHeader))
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidAuthFormat)
	}
	tokenStr := parts[1]

	// Validate token is not empty
	if strings.TrimSpace(tokenStr) == "" {
		logAuthFailure(c, "empty_token", "Empty token provided")
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidToken)
	}

	// Get JWT secret
	jwtSecret := auth.GetJWTSecret()
	if len(jwtSecret) == 0 {
		log.Printf("[ERROR] JWT_SECRET is not configured")
		return sendAuthError(c, fiber.StatusInternalServerError, ErrServerMisconfiguration)
	}

	// Parse and validate JWT token
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		logAuthFailure(c, "token_parse_error", err.Error())
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidToken)
	}

	if !token.Valid {
		logAuthFailure(c, "invalid_token", "Token validation failed")
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidToken)
	}

	// Extract and validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logAuthFailure(c, "invalid_claims_type", "Claims are not in expected format")
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidClaims)
	}

	// Validate required claims
	userID, ok := claims["sub"]
	if !ok {
		logAuthFailure(c, "missing_sub_claim", "Token missing user ID claim")
		return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidClaims)
	}

	// Validate expiration if present
	if exp, ok := claims["exp"]; ok {
		if expFloat, ok := exp.(float64); ok {
			if time.Now().Unix() > int64(expFloat) {
				logAuthFailure(c, "token_expired", "Token has expired")
				return sendAuthError(c, fiber.StatusUnauthorized, ErrInvalidToken)
			}
		}
	}

	// Store user context for downstream handlers
	c.Locals("user_id", userID)
	c.Locals("jwt_claims", claims)
	
	// Log successful authentication for audit purposes
	logAuthSuccess(c, userID)
	
	return c.Next()
}

// sendAuthError sends a standardized authentication error response
func sendAuthError(c *fiber.Ctx, status int, authErr AuthError) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   authErr.Message,
		"code":    authErr.Code,
		"details": authErr.Details,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// logAuthFailure logs authentication failures for security monitoring
func logAuthFailure(c *fiber.Ctx, reason, details string) {
	log.Printf("[AUTH_FAILURE] IP: %s, Path: %s, Method: %s, Reason: %s, Details: %s", 
		c.IP(), c.Path(), c.Method(), reason, details)
}

// logAuthSuccess logs successful authentication for audit purposes
func logAuthSuccess(c *fiber.Ctx, userID interface{}) {
	log.Printf("[AUTH_SUCCESS] IP: %s, Path: %s, Method: %s, UserID: %v", 
		c.IP(), c.Path(), c.Method(), userID)
}
