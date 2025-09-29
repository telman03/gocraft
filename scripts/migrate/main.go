package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/telman03/gocraft-backend/internal/database"
	"github.com/telman03/gocraft-backend/internal/models"
)

func main() {
	// Load env file
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	// Initialize database connection
	database.InitDB()

	// Auto migrate models
	fmt.Println("Migrating User model...")
	database.DB.AutoMigrate(&models.User{})
	
	fmt.Println("Migrating ProjectHistory model...")
	database.DB.AutoMigrate(&models.ProjectHistory{})
	
	fmt.Println("Migration complete!")
}
