package models

import (
	"time"

	"gorm.io/datatypes"
)

// ZipFileStatus represents the status of a project's ZIP file
type ZipFileStatus string

const (
	ZipFileStatusAvailable ZipFileStatus = "available"
	ZipFileStatusExpired   ZipFileStatus = "expired"
	ZipFileStatusDeleted   ZipFileStatus = "deleted"
	ZipFileStatusError     ZipFileStatus = "error"
)

// ProjectHistory represents a user's project generation history
type ProjectHistory struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	UserID               uint           `gorm:"not null;index" json:"user_id"`
	ProjectName          string         `gorm:"size:100;not null" json:"project_name" validate:"required,min=1,max=100"`
	Framework            string         `gorm:"size:20;not null;index" json:"framework" validate:"required,oneof=gin echo fiber"`
	Features             datatypes.JSON `gorm:"type:jsonb" json:"features" validate:"required"`
	AdjustedFeatures     datatypes.JSON `gorm:"type:jsonb" json:"adjusted_features" validate:"required"`
	ZipFilePath          *string        `gorm:"size:500" json:"zip_file_path,omitempty"`
	ZipFileSize          *int64         `json:"zip_file_size,omitempty"`
	ZipFileStatus        ZipFileStatus  `gorm:"size:20;default:available;index" json:"zip_file_status"`
	GenerationDurationMs *int           `json:"generation_duration_ms,omitempty"`
	CreatedAt            time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`

	// Associations
	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

// TableName specifies the table name for ProjectHistory
func (ProjectHistory) TableName() string {
	return "project_history"
}