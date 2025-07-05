package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func ZipFolder(source, target string) error {
	os.MkdirAll(filepath.Dir(target), 0755)
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	// Walk through the source
	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories (we only zip files)
		if info.IsDir() {
			return nil
		}

		// Calculate the relative path inside the ZIP
		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Create file entry in ZIP
		f, err := archive.Create(relPath)
		if err != nil {
			return err
		}

		// Copy contents into ZIP
		_, err = io.Copy(f, file)
		return err
	})

	return err
}
