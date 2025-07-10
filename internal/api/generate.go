package api

import (

	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gofiber/fiber/v2"
	supabase "github.com/supabase-community/supabase-go"
	"github.com/telman03/ai-backend-generator/internal/builder"
)

var SupabaseClient *supabase.Client
var SupabaseJWTSecret string

func InitSupabase() {
	var err error
	SupabaseClient, err = supabase.NewClient(
		os.Getenv("SUPABASE_URL"),
		os.Getenv("SUPABASE_KEY"),
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	SupabaseJWTSecret = os.Getenv("SUPABASE_JWT_SECRET")
	if SupabaseJWTSecret == "" {
		log.Fatal("SUPABASE_JWT_SECRET not set in environment")
	}
}

type GenerateRequest struct {
	Features []string `json:"features"`
}

func GenerateHandler(c *fiber.Ctx) error {
	// 🔐 Extract token
	authHeader := c.Get("Authorization")
	if authHeader == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing Authorization header"})
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid Authorization format"})
	}

	// 🧠 Decode JWT to extract user ID
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return []byte(SupabaseJWTSecret), nil
	})
	if err != nil || !token.Valid {
		log.Println("JWT parse error:", err)
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid claims"})
	}
	userID, ok := claims["sub"].(string)
	if !ok {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing user ID in token"})
	}

	// 🚦 Check usage
	if !checkUsageLimit(userID) {
		return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{"error": "Daily usage limit exceeded"})
	}

	// 📨 Parse request
	var req GenerateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	// 🛠 Generate backend
	zipPath, err := builder.GenerateProject(req.Features)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Download(zipPath)
}

func checkUsageLimit(userID string) bool {
	// The SupabaseClient.Rpc method expects (functionName string, countOption string, rpcBody interface{})
	// Assuming the second string argument (countOption) is empty for a simple RPC call
	// And the result is directly returned as a string.
	usageStr := SupabaseClient.Rpc("check_daily_usage", "", map[string]interface{}{
		"user_id": userID,
	})

	// You'll need to determine how to detect an error from the `usageStr`.
	// Common ways:
	// 1. If an empty string or a specific error message string is returned on failure.
	// 2. If the RPC function itself inserts an error into a table or logs it for you to retrieve separately.
	// This example assumes an empty string means an error or unexpected result.
	if usageStr == "" {
		log.Println("RPC call did not return a valid usage string or an error occurred (check Supabase logs).")
		return false
	}

	usageCount, err := strconv.Atoi(usageStr)
	if err != nil {
		log.Printf("Conversion to int failed for '%s': %v", usageStr, err)
		return false
	}

	return usageCount < 5 // Set your daily limit
}