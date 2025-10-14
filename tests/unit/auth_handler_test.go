package unit

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/telman03/gocraft-backend/internal/handlers"
	"github.com/telman03/gocraft-backend/internal/models"
)

func TestGetCurrentUser_Success(t *testing.T) {
	app := fiber.New()
	
	// Mock user data
	mockUserID := float64(123) // JWT claims parse numbers as float64
	
	app.Get("/auth/me", func(c *fiber.Ctx) error {
		// Simulate middleware setting user_id
		c.Locals("user_id", mockUserID)
		return handlers.GetCurrentUser(c)
	})

	req := httptest.NewRequest("GET", "/auth/me", nil)
	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	
	// Check that response contains expected fields
	assert.Contains(t, response, "id")
	assert.Contains(t, response, "email")
	assert.Contains(t, response, "projects_count")
	assert.Contains(t, response, "joined_date")
	assert.Contains(t, response, "created_at")
}

func TestGetCurrentUser_InvalidUserID(t *testing.T) {
	app := fiber.New()
	
	app.Get("/auth/me", func(c *fiber.Ctx) error {
		// Simulate invalid user_id type
		c.Locals("user_id", "invalid")
		return handlers.GetCurrentUser(c)
	})

	req := httptest.NewRequest("GET", "/auth/me", nil)
	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "error")
	assert.Equal(t, "Invalid user ID format", response["error"])
}

func TestGetCurrentUser_MissingUserID(t *testing.T) {
	app := fiber.New()
	
	app.Get("/auth/me", handlers.GetCurrentUser)

	req := httptest.NewRequest("GET", "/auth/me", nil)
	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	
	assert.Contains(t, response, "error")
	assert.Equal(t, "User not authenticated", response["error"])
}