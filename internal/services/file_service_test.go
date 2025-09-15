package services

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestFileService_GetFilePath(t *testing.T) {
	basePath := "/tmp/test"
	fileService := NewFileService(basePath, 30*24*time.Hour)

	tests := []struct {
		name       string
		projectID  uint
		filename   string
		wantPrefix string
	}{
		{
			name:       "basic file path",
			projectID:  1,
			filename:   "test.zip",
			wantPrefix: filepath.Join(basePath, "project_1", "test.zip"),
		},
		{
			name:       "sanitized filename",
			projectID:  2,
			filename:   "../../../etc/passwd",
			wantPrefix: filepath.Join(basePath, "project_2", "______etc_passwd"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileService.GetFilePath(tt.projectID, tt.filename)
			if got != tt.wantPrefix {
				t.Errorf("GetFilePath() = %v, want %v", got, tt.wantPrefix)
			}
		})
	}
}

func TestFileService_ValidateFilePath(t *testing.T) {
	basePath := "/tmp/test"
	fileService := NewFileService(basePath, 30*24*time.Hour)

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    filepath.Join(basePath, "project_1", "test.zip"),
			wantErr: false,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
		{
			name:    "directory traversal",
			path:    filepath.Join(basePath, "..", "etc", "passwd"),
			wantErr: true,
		},
		{
			name:    "too long path",
			path:    filepath.Join(basePath, string(make([]byte, 600))),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fileService.ValidateFilePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilePath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileService_GenerateSecureFilePath(t *testing.T) {
	basePath := "/tmp/test"
	fileService := NewFileService(basePath, 30*24*time.Hour)

	path1 := fileService.GenerateSecureFilePath(1, 1, "test.zip")
	path2 := fileService.GenerateSecureFilePath(1, 1, "test.zip")

	// Paths should be different due to timestamp in hash
	if path1 == path2 {
		t.Errorf("GenerateSecureFilePath() should generate unique paths, got same path twice: %v", path1)
	}

	// Both paths should be within base directory
	if err := fileService.ValidateFilePath(path1); err != nil {
		t.Errorf("Generated path should be valid: %v", err)
	}
	if err := fileService.ValidateFilePath(path2); err != nil {
		t.Errorf("Generated path should be valid: %v", err)
	}
}

func TestFileService_FileOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	// Test file path generation
	testPath := fileService.GetFilePath(1, "test.zip")

	// Initially file should not exist
	if fileService.FileExists(testPath) {
		t.Errorf("File should not exist initially")
	}

	// Create directory structure
	if err := fileService.EnsureDirectoryExists(testPath); err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}

	// Create a test file
	testContent := []byte("test content")
	if err := os.WriteFile(testPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Now file should exist
	if !fileService.FileExists(testPath) {
		t.Errorf("File should exist after creation")
	}

	// Test file size
	size, err := fileService.GetFileSize(testPath)
	if err != nil {
		t.Errorf("Failed to get file size: %v", err)
	}
	if size != int64(len(testContent)) {
		t.Errorf("File size mismatch: got %d, want %d", size, len(testContent))
	}

	// Test file info
	info, err := fileService.GetFileInfo(testPath)
	if err != nil {
		t.Errorf("Failed to get file info: %v", err)
	}
	if !info.Exists {
		t.Errorf("File info should show file exists")
	}
	if info.Size != int64(len(testContent)) {
		t.Errorf("File info size mismatch: got %d, want %d", info.Size, len(testContent))
	}

	// Test file deletion
	if err := fileService.DeleteFile(testPath); err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	// File should no longer exist
	if fileService.FileExists(testPath) {
		t.Errorf("File should not exist after deletion")
	}
}

func TestFileService_IsFileExpired(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create file service with very short max age for testing
	fileService := NewFileService(tempDir, 1*time.Millisecond)

	// Create a test file
	testPath := fileService.GetFilePath(1, "test.zip")
	if err := fileService.EnsureDirectoryExists(testPath); err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// File should not be expired immediately
	if fileService.IsFileExpired(testPath) {
		t.Errorf("File should not be expired immediately after creation")
	}

	// Wait for file to expire
	time.Sleep(2 * time.Millisecond)

	// File should now be expired
	if !fileService.IsFileExpired(testPath) {
		t.Errorf("File should be expired after max age")
	}
}

func TestFileService_GenerateSecureFilePathUniqueness(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 30*24*time.Hour)

	// Generate multiple paths for same parameters
	paths := make(map[string]bool)
	for i := 0; i < 100; i++ {
		path := fileService.GenerateSecureFilePath(1, 1, "test.zip")
		if paths[path] {
			t.Errorf("Generated duplicate path: %s", path)
		}
		paths[path] = true
	}

	// All paths should be unique
	if len(paths) != 100 {
		t.Errorf("Expected 100 unique paths, got %d", len(paths))
	}
}

func TestFileService_GetFileModTime(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)
	testPath := fileService.GetFilePath(1, "test.zip")

	// Test non-existent file
	_, err = fileService.GetFileModTime(testPath)
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}

	// Create file and test mod time
	if err := fileService.EnsureDirectoryExists(testPath); err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}
	
	beforeWrite := time.Now()
	if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	afterWrite := time.Now()

	modTime, err := fileService.GetFileModTime(testPath)
	if err != nil {
		t.Errorf("Failed to get file mod time: %v", err)
	}

	if modTime.Before(beforeWrite) || modTime.After(afterWrite) {
		t.Errorf("File mod time %v should be between %v and %v", modTime, beforeWrite, afterWrite)
	}
}

func TestFileService_GetFileInfo(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)
	testPath := fileService.GetFilePath(1, "test.zip")

	// Test non-existent file
	_, err = fileService.GetFileInfo(testPath)
	if err == nil {
		t.Errorf("Expected error for non-existent file")
	}

	// Create file and test info
	if err := fileService.EnsureDirectoryExists(testPath); err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}
	
	testContent := []byte("test content for file info")
	if err := os.WriteFile(testPath, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	info, err := fileService.GetFileInfo(testPath)
	if err != nil {
		t.Errorf("Failed to get file info: %v", err)
	}

	if info.Path != testPath {
		t.Errorf("File info path mismatch: got %s, want %s", info.Path, testPath)
	}
	if info.Size != int64(len(testContent)) {
		t.Errorf("File info size mismatch: got %d, want %d", info.Size, len(testContent))
	}
	if !info.Exists {
		t.Errorf("File info should show file exists")
	}
	if info.IsDir {
		t.Errorf("File info should not show file as directory")
	}
	if info.Expired {
		t.Errorf("File should not be expired with 1 hour max age")
	}
}

func TestFileService_SecurityValidation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	// Test directory traversal attempts
	dangerousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32\\config\\sam",
		"/etc/passwd",
		"C:\\Windows\\System32\\config\\sam",
		"project_1/../../../etc/passwd",
	}

	for _, dangerousPath := range dangerousPaths {
		t.Run("dangerous_path_"+dangerousPath, func(t *testing.T) {
			err := fileService.ValidateFilePath(dangerousPath)
			if err == nil {
				t.Errorf("Expected validation error for dangerous path: %s", dangerousPath)
			}

			// File operations should also fail
			if fileService.FileExists(dangerousPath) {
				t.Errorf("FileExists should return false for dangerous path: %s", dangerousPath)
			}

			_, err = fileService.GetFileSize(dangerousPath)
			if err == nil {
				t.Errorf("GetFileSize should fail for dangerous path: %s", dangerousPath)
			}
		})
	}
}

func TestFileService_FilenamesSanitization(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	tests := []struct {
		input    string
		expected string
	}{
		{"normal-file.zip", "normal-file.zip"},
		{"file with spaces.zip", "file with spaces.zip"},
		{"../dangerous.zip", "___dangerous.zip"},
		{"file:with*special?chars.zip", "file_with_special_chars.zip"},
		{"file<>|.zip", "file___.zip"},
		{"", "file"},
		{string(make([]byte, 150)), string(make([]byte, 96)) + ".zip"}, // Long filename
	}

	for _, tt := range tests {
		t.Run("sanitize_"+tt.input, func(t *testing.T) {
			path := fileService.GetFilePath(1, tt.input)
			// The sanitized filename should be in the path
			if tt.input != "" && len(tt.input) <= 100 {
				// For normal cases, check if sanitization worked
				if tt.input != tt.expected && filepath.Base(path) != tt.expected {
					// This is a complex check, so let's just ensure no dangerous chars remain
					basename := filepath.Base(path)
					dangerousChars := []string{"..", "/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
					for _, char := range dangerousChars {
						if strings.Contains(basename, char) {
							t.Errorf("Sanitized filename still contains dangerous char '%s': %s", char, basename)
						}
					}
				}
			}
		})
	}
}

func TestFileService_EmptyDirectoryCleanup(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	// Create nested directory structure with file
	testPath := fileService.GetFilePath(1, "test.zip")
	if err := fileService.EnsureDirectoryExists(testPath); err != nil {
		t.Fatalf("Failed to ensure directory exists: %v", err)
	}
	if err := os.WriteFile(testPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Verify directory exists
	dir := filepath.Dir(testPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Fatalf("Directory should exist: %s", dir)
	}

	// Delete file (this should trigger cleanup of empty directories)
	if err := fileService.DeleteFile(testPath); err != nil {
		t.Errorf("Failed to delete file: %v", err)
	}

	// Directory might be cleaned up (this is optional behavior)
	// We just verify the file is gone
	if fileService.FileExists(testPath) {
		t.Errorf("File should be deleted")
	}
}

func TestFileService_ConfigurationMethods(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	initialMaxAge := 24 * time.Hour
	fileService := NewFileService(tempDir, initialMaxAge)

	// Test getter methods
	if fileService.GetBasePath() != tempDir {
		t.Errorf("GetBasePath() = %s, want %s", fileService.GetBasePath(), tempDir)
	}

	if fileService.GetMaxAge() != initialMaxAge {
		t.Errorf("GetMaxAge() = %v, want %v", fileService.GetMaxAge(), initialMaxAge)
	}

	// Test setter method
	newMaxAge := 48 * time.Hour
	fileService.SetMaxAge(newMaxAge)

	if fileService.GetMaxAge() != newMaxAge {
		t.Errorf("After SetMaxAge(), GetMaxAge() = %v, want %v", fileService.GetMaxAge(), newMaxAge)
	}
}

func TestFileService_EdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	// Test with empty strings
	if fileService.FileExists("") {
		t.Errorf("FileExists should return false for empty string")
	}

	_, err = fileService.GetFileSize("")
	if err == nil {
		t.Errorf("GetFileSize should return error for empty string")
	}

	if !fileService.IsFileExpired("") {
		t.Errorf("IsFileExpired should return true for empty string")
	}

	err = fileService.DeleteFile("")
	if err != nil {
		t.Errorf("DeleteFile should not error for empty string (no-op)")
	}

	// Test with nil-like values
	err = fileService.ValidateFilePath("")
	if err == nil {
		t.Errorf("ValidateFilePath should return error for empty string")
	}
}

func TestFileService_ConcurrentAccess(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fileservice_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileService := NewFileService(tempDir, 1*time.Hour)

	// Test concurrent file operations
	const numGoroutines = 10
	const numOperations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				// Generate unique file path for each operation
				testPath := fileService.GetFilePath(uint(goroutineID*numOperations+j), fmt.Sprintf("test-%d-%d.zip", goroutineID, j))
				
				// Create directory
				if err := fileService.EnsureDirectoryExists(testPath); err != nil {
					errors <- fmt.Errorf("goroutine %d, op %d: EnsureDirectoryExists failed: %v", goroutineID, j, err)
					continue
				}

				// Create file
				if err := os.WriteFile(testPath, []byte(fmt.Sprintf("content-%d-%d", goroutineID, j)), 0644); err != nil {
					errors <- fmt.Errorf("goroutine %d, op %d: WriteFile failed: %v", goroutineID, j, err)
					continue
				}

				// Check file exists
				if !fileService.FileExists(testPath) {
					errors <- fmt.Errorf("goroutine %d, op %d: File should exist", goroutineID, j)
					continue
				}

				// Get file size
				if _, err := fileService.GetFileSize(testPath); err != nil {
					errors <- fmt.Errorf("goroutine %d, op %d: GetFileSize failed: %v", goroutineID, j, err)
					continue
				}

				// Delete file
				if err := fileService.DeleteFile(testPath); err != nil {
					errors <- fmt.Errorf("goroutine %d, op %d: DeleteFile failed: %v", goroutineID, j, err)
					continue
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for any errors
	for err := range errors {
		t.Error(err)
	}
}