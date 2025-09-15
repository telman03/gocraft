package services

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileService handles file operations for project history ZIP files
type FileService struct {
	basePath string
	maxAge   time.Duration
}

// NewFileService creates a new instance of FileService
func NewFileService(basePath string, maxAge time.Duration) *FileService {
	return &FileService{
		basePath: basePath,
		maxAge:   maxAge,
	}
}

// GetFilePath generates a secure file path for a project ZIP file
func (s *FileService) GetFilePath(projectID uint, filename string) string {
	// Sanitize filename to prevent directory traversal
	sanitizedFilename := s.sanitizeFilename(filename)
	
	// Create a subdirectory based on project ID to distribute files
	subDir := fmt.Sprintf("project_%d", projectID)
	
	// Generate the full path
	fullPath := filepath.Join(s.basePath, subDir, sanitizedFilename)
	
	return filepath.Clean(fullPath)
}

// GenerateSecureFilePath creates a secure file path with hash-based naming
func (s *FileService) GenerateSecureFilePath(userID uint, projectID uint, originalName string) string {
	// Create a hash of user ID, project ID, and timestamp for uniqueness
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%d_%d_%s_%d", userID, projectID, originalName, time.Now().UnixNano())))
	hashStr := fmt.Sprintf("%x", hash.Sum(nil))[:16] // Use first 16 characters
	
	// Extract file extension
	ext := filepath.Ext(originalName)
	if ext == "" {
		ext = ".zip" // Default to .zip if no extension
	}
	
	// Create filename with hash
	filename := fmt.Sprintf("%s_%s%s", s.sanitizeFilename(strings.TrimSuffix(originalName, ext)), hashStr, ext)
	
	// Create subdirectory structure: user_id/year/month
	now := time.Now()
	subDir := filepath.Join(fmt.Sprintf("user_%d", userID), fmt.Sprintf("%d", now.Year()), fmt.Sprintf("%02d", now.Month()))
	
	// Generate the full path
	fullPath := filepath.Join(s.basePath, subDir, filename)
	
	return filepath.Clean(fullPath)
}

// FileExists checks if a file exists at the given path
func (s *FileService) FileExists(path string) bool {
	if path == "" {
		return false
	}
	
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return false
	}
	
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// GetFileSize returns the size of the file in bytes
func (s *FileService) GetFileSize(path string) (int64, error) {
	if path == "" {
		return 0, errors.New("empty file path")
	}
	
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return 0, errors.New("invalid file path")
	}
	
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, errors.New("file not found")
		}
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	
	return fileInfo.Size(), nil
}

// DeleteFile safely deletes a file with error handling
func (s *FileService) DeleteFile(path string) error {
	if path == "" {
		return nil // Nothing to delete
	}
	
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return errors.New("invalid file path")
	}
	
	// Check if file exists before attempting to delete
	if !s.FileExists(path) {
		return nil // File doesn't exist, consider it already deleted
	}
	
	err := os.Remove(path)
	if err != nil {
		return fmt.Errorf("failed to delete file %s: %w", path, err)
	}
	
	// Try to remove empty parent directories (optional cleanup)
	s.cleanupEmptyDirectories(filepath.Dir(path))
	
	return nil
}

// IsFileExpired checks if a file has exceeded the maximum age
func (s *FileService) IsFileExpired(path string) bool {
	if path == "" {
		return true
	}
	
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return true
	}
	
	fileInfo, err := os.Stat(path)
	if err != nil {
		return true // If we can't stat the file, consider it expired
	}
	
	// Check if file is older than maxAge
	return time.Since(fileInfo.ModTime()) > s.maxAge
}

// GetFileModTime returns the modification time of the file
func (s *FileService) GetFileModTime(path string) (time.Time, error) {
	if path == "" {
		return time.Time{}, errors.New("empty file path")
	}
	
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return time.Time{}, errors.New("invalid file path")
	}
	
	fileInfo, err := os.Stat(path)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get file info: %w", err)
	}
	
	return fileInfo.ModTime(), nil
}

// EnsureDirectoryExists creates the directory structure if it doesn't exist
func (s *FileService) EnsureDirectoryExists(path string) error {
	// Validate path is within base directory
	if !s.isPathSafe(path) {
		return errors.New("invalid directory path")
	}
	
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	return nil
}

// ValidateFilePath validates that a file path is safe and within allowed boundaries
func (s *FileService) ValidateFilePath(path string) error {
	if path == "" {
		return errors.New("empty file path")
	}
	
	// Clean the path to resolve any .. or . components
	cleanPath := filepath.Clean(path)
	
	// Check if the path is within the base directory
	if !s.isPathSafe(cleanPath) {
		return errors.New("file path is outside allowed directory")
	}
	
	// Check for suspicious patterns
	if strings.Contains(path, "..") {
		return errors.New("path contains directory traversal")
	}
	
	// Check path length
	if len(path) > 500 {
		return errors.New("file path too long")
	}
	
	return nil
}

// GetFileInfo returns comprehensive file information
func (s *FileService) GetFileInfo(path string) (*FileInfo, error) {
	if err := s.ValidateFilePath(path); err != nil {
		return nil, err
	}
	
	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("file not found")
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	
	return &FileInfo{
		Path:     path,
		Size:     fileInfo.Size(),
		ModTime:  fileInfo.ModTime(),
		IsDir:    fileInfo.IsDir(),
		Exists:   true,
		Expired:  s.IsFileExpired(path),
	}, nil
}

// FileInfo represents comprehensive file information
type FileInfo struct {
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"mod_time"`
	IsDir    bool      `json:"is_dir"`
	Exists   bool      `json:"exists"`
	Expired  bool      `json:"expired"`
}

// sanitizeFilename removes or replaces dangerous characters in filenames
func (s *FileService) sanitizeFilename(filename string) string {
	// Remove or replace dangerous characters
	dangerous := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	sanitized := filename
	
	for _, char := range dangerous {
		sanitized = strings.ReplaceAll(sanitized, char, "_")
	}
	
	// Remove leading/trailing spaces and dots
	sanitized = strings.Trim(sanitized, " .")
	
	// Ensure filename is not empty
	if sanitized == "" {
		sanitized = "file"
	}
	
	// Limit filename length
	if len(sanitized) > 100 {
		ext := filepath.Ext(sanitized)
		name := strings.TrimSuffix(sanitized, ext)
		if len(name) > 96 {
			name = name[:96]
		}
		sanitized = name + ext
	}
	
	return sanitized
}

// isPathSafe checks if a path is within the allowed base directory
func (s *FileService) isPathSafe(path string) bool {
	// Clean both paths to resolve any .. or . components
	cleanPath := filepath.Clean(path)
	cleanBasePath := filepath.Clean(s.basePath)
	
	// Convert to absolute paths for comparison
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return false
	}
	
	absBasePath, err := filepath.Abs(cleanBasePath)
	if err != nil {
		return false
	}
	
	// Check if the path starts with the base path
	return strings.HasPrefix(absPath, absBasePath)
}

// cleanupEmptyDirectories removes empty parent directories up to the base path
func (s *FileService) cleanupEmptyDirectories(dir string) {
	// Don't remove the base directory itself
	if dir == s.basePath || !s.isPathSafe(dir) {
		return
	}
	
	// Check if directory is empty
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) > 0 {
		return // Directory is not empty or can't be read
	}
	
	// Remove the empty directory
	if err := os.Remove(dir); err == nil {
		// Recursively try to remove parent directories
		s.cleanupEmptyDirectories(filepath.Dir(dir))
	}
}

// GetBasePath returns the base path for file storage
func (s *FileService) GetBasePath() string {
	return s.basePath
}

// SetMaxAge updates the maximum age for file expiration
func (s *FileService) SetMaxAge(maxAge time.Duration) {
	s.maxAge = maxAge
}

// GetMaxAge returns the current maximum age setting
func (s *FileService) GetMaxAge() time.Duration {
	return s.maxAge
}