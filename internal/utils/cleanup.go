package utils

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// CleanupOldFiles removes files older than the specified duration from the output directory
func CleanupOldFiles(outputDir string, maxAge time.Duration) error {
	return filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if file is older than maxAge
		if time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				log.Printf("Failed to remove old file %s: %v", path, err)
				return nil // Continue with other files
			}
			log.Printf("Cleaned up old file: %s", path)
		}

		return nil
	})
}

// CleanupAfterDownload removes the generated files after a delay to allow download completion
func CleanupAfterDownload(zipPath, projectPath string, delay time.Duration) {
	go func() {
		time.Sleep(delay)
		
		// Remove zip file
		if err := os.Remove(zipPath); err != nil {
			log.Printf("Failed to cleanup zip file %s: %v", zipPath, err)
		}
		
		// Remove project directory
		if err := os.RemoveAll(projectPath); err != nil {
			log.Printf("Failed to cleanup project directory %s: %v", projectPath, err)
		}
		
		log.Printf("Cleaned up generated files: %s, %s", zipPath, projectPath)
	}()
}