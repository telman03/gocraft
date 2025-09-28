package unit

import (
	"testing"

	"github.com/telman03/ai-backend-generator/internal/models"
	"gorm.io/gorm"
)

// stringPtr returns a pointer to the given string
func stringPtr(s string) *string {
	return &s
}

// int64Ptr returns a pointer to the given int64
func int64Ptr(i int64) *int64 {
	return &i
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		Email:      "test@example.com",
		Password:   "hashedpassword",
		Role:       models.UserRoleUser,
		IsVerified: true,
	}
	
	err := db.Create(user).Error
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	
	return user
}

