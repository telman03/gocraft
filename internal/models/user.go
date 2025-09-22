package models

import (
	"time"
)

// UserRole represents the role of a user
type UserRole string

const (
	UserRoleUser  UserRole = "user"
	UserRoleAdmin UserRole = "admin"
)

type User struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Email        string    `gorm:"unique;not null" json:"email"`
	Password     string    `gorm:"not null" json:"-"`
	Role         UserRole  `gorm:"size:20;default:user;not null" json:"role"`
	OTP          string    `gorm:"size:6" json:"-"`
	OTPExpiresAt time.Time `json:"-"`
	IsVerified   bool      `gorm:"default:false" json:"is_verified"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Associations
	ProjectHistory []ProjectHistory `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"project_history,omitempty"`
}

// IsAdmin checks if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == UserRoleAdmin
}

// IsUser checks if the user has user role
func (u *User) IsUser() bool {
	return u.Role == UserRoleUser
}
