package utils

import (
	"os"
	"text/template"
)

func ApplyTemplate(srcPath, destPath string, data any) error {
	tmpl, err := template.ParseFiles(srcPath)
	if err != nil {
		return err
	}

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return tmpl.Execute(f, data)
}