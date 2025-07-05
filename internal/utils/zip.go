package utils

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func ZipFolder(source, target string) error {
	zipfile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		relPath, _ := filepath.Rel(source, path)
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := archive.Create(relPath)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		return err
	})

	return nil
}