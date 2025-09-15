package models

import (
	"time"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	OTP          string    `gorm:"size:6" json:"-"`
	OTPExpiresAt time.Time `json:"-"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Associations
	ProjectHistory []ProjectHistory `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"project_history,omitempty"`
}
