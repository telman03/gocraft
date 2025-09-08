package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// GetJWTSecret returns the JWT secret from environment variables
func GetJWTSecret() ([]byte, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET environment variable not set")
	}
	return []byte(secret), nil
}

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
	// Get the JWT secret
	jwtSecret, err := GetJWTSecret()
	if err != nil {
		return "", err
	}

	// Using MapClaims with "sub" for the user ID to match the Login function
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(), // Token valid for 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
