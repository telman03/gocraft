package builder

import (

	"os"
	"path/filepath"


	"github.com/google/uuid"
	"github.com/telman03/ai-backend-generator/internal/utils"
)

func GenerateProject(features []string) (string, error) {
	id := uuid.New().String()
	projectPath := filepath.Join("output", id)
	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Track which modules are selected for main.go
	flags := map[string]bool{
		"Auth":   false,
		"DB":     false,
		"Router": false,
		"OpenAI": false,
	}

	// Output subfolders like internal/auth, internal/db, etc.
	for _, feature := range features {
		flags[toFlagName(feature)] = true

		tmplPath := filepath.Join("internal", "templates", feature+".tmpl")
		targetDir := filepath.Join(projectPath, "internal", feature)
		targetFile := filepath.Join(targetDir, feature+".go")

		err := os.MkdirAll(targetDir, os.ModePerm)
		if err != nil {
			return "", err
		}

		err = utils.ApplyTemplate(tmplPath, targetFile, nil)
		if err != nil {
			return "", err
		}
	}

	// Render main.go from main.tmpl with feature flags
	err = utils.ApplyTemplate(
		filepath.Join("internal", "templates", "main.tmpl"),
		filepath.Join(projectPath, "main.go"),
		flags,
	)
	if err != nil {
		return "", err
	}

	// Zip the folder and return path
	zipPath := filepath.Join("output", id+".zip")
	err = utils.ZipFolder(projectPath, zipPath)
	if err != nil {
		return "", err
	}

	return zipPath, nil
}

func toFlagName(feature string) string {
	switch feature {
	case "auth":
		return "Auth"
	case "db":
		return "DB"
	case "router":
		return "Router"
	case "openai":
		return "OpenAI"
	default:
		return ""
	}
}