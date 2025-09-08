package models

import "time"

// UserProfileResponse defines the response for the user profile endpoint
type UserProfileResponse struct {
	ID         uint      `json:"id"`
	Email      string    `json:"email"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
