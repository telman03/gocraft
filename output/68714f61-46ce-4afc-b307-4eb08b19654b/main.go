package main

import (
	"github.com/gofiber/fiber/v2"
	"your_project/auth"
	"your_project/db"
	
)

func main() {
	app := fiber.New()

	db.Connect()
	

	app.Listen(":8080")
}