package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/telman03/gocraft-backend/internal/database"
	"github.com/telman03/gocraft-backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database
	database.InitDB()

	fmt.Println("Adding user roles and creating admin user...")

	// Add role column to existing users table
	if err := database.DB.Exec("ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'user' NOT NULL").Error; err != nil {
		log.Printf("Warning: Failed to add role column (may already exist): %v", err)
	}

	// Auto-migrate to ensure the model is up to date
	if err := database.DB.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate User model: %v", err)
	}

	// Create admin user
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if adminEmail == "" {
		log.Fatal("ADMIN_EMAIL environment variable is required")
	}
	if adminPassword == "" {
		log.Fatal("ADMIN_PASSWORD environment variable is required")
	}

	// Check if admin user already exists
	var existingAdmin models.User
	if err := database.DB.Where("email = ? OR role = ?", adminEmail, models.UserRoleAdmin).First(&existingAdmin).Error; err == nil {
		fmt.Printf("Admin user already exists: %s (ID: %d)\n", existingAdmin.Email, existingAdmin.ID)
		return
	}

	// Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash admin password: %v", err)
	}

	// Create admin user
	adminUser := models.User{
		Email:      adminEmail,
		Password:   string(hashedPassword),
		Role:       models.UserRoleAdmin,
		IsVerified: true, // Admin is pre-verified
	}

	if err := database.DB.Create(&adminUser).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	fmt.Printf("✅ Admin user created successfully!\n")
	fmt.Printf("   Email: %s\n", adminEmail)
	fmt.Printf("   Password: %s\n", adminPassword)
	fmt.Printf("   Role: %s\n", adminUser.Role)
	fmt.Printf("   ID: %d\n", adminUser.ID)
	fmt.Println()
	fmt.Println("⚠️  IMPORTANT: Please change the admin password after first login!")
	fmt.Println("   You can set custom admin credentials using environment variables:")
	fmt.Println("   ADMIN_EMAIL=your-admin@email.com")
	fmt.Println("   ADMIN_PASSWORD=your-secure-password")
	fmt.Println()
	fmt.Println("🔧 Admin can now access maintenance endpoints at /api/maintenance/*")
}