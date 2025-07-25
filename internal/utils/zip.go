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


	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}


		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}


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
