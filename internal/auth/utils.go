package auth

import (
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	// "errors"
	"golang.org/x/crypto/bcrypt"
)

// GetJWTSecret returns the JWT secret, removing any surrounding quotes
func GetJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	// Remove surrounding quotes if present
	secret = strings.Trim(secret, "\"'")
	return []byte(secret)
}

// jwtKey is initialized when needed to ensure env vars are loaded
var jwtKey []byte

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

//func CheckPassword(hash, password string) error {
//	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
//}

func GenerateJWT(userID uint) (string, error) {
	// Get the JWT secret, removing quotes if present
	jwtKey = GetJWTSecret()

	// Using standard claims with "sub" for user ID
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // Token valid for 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
