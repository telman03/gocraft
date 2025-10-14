package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/telman03/gocraft-backend/internal/utils"
)

// TemplateMapping defines where each template should be rendered
type TemplateMapping struct {
	SourcePath      string
	DestinationPath string
	IsRootFile      bool // Files that go in project root
	IsInternalFile  bool // Files that go in internal/ directory
	IsCmdFile       bool // Files that go in cmd/ directory
	IsDocsFile      bool // Files that go in docs/ directory
}

// Feature mappings - maps feature names to their template destinations
var featureMappings = map[string]TemplateMapping{
	// Root level files
	"dockerfile": {
		SourcePath:      "dockerfile.tmpl",
		DestinationPath: "Dockerfile",
		IsRootFile:      true,
	},
	"docker": {
		SourcePath:      "dockerfile.tmpl",
		DestinationPath: "Dockerfile",
		IsRootFile:      true,
	},
	"docker-compose": {
		SourcePath:      "docker-compose.tmpl",
		DestinationPath: "docker-compose.yml",
		IsRootFile:      true,
	},
	"makefile": {
		SourcePath:      "makefile.tmpl",
		DestinationPath: "Makefile",
		IsRootFile:      true,
	},
	"gitignore": {
		SourcePath:      "gitignore.tmpl",
		DestinationPath: ".gitignore",
		IsRootFile:      true,
	},
	"env": {
		SourcePath:      "env.tmpl",
		DestinationPath: ".env.example",
		IsRootFile:      true,
	},
	"env-config": {
		SourcePath:      "env.tmpl",
		DestinationPath: ".env.example",
		IsRootFile:      true,
	},
	"environment": {
		SourcePath:      "env.tmpl",
		DestinationPath: ".env.example",
		IsRootFile:      true,
	},
	"readme": {
		SourcePath:      "readme.tmpl",
		DestinationPath: "README.md",
		IsRootFile:      true,
	},
	"github-actions": {
		SourcePath:      "github-actions.tmpl",
		DestinationPath: ".github/workflows/ci.yml",
		IsRootFile:      true,
	},

	// Main application files
	"main": {
		SourcePath:      "main.tmpl",
		DestinationPath: "cmd/{{.ProjectName}}/main.go",
		IsRootFile:      false,
	},
	"router": {
		SourcePath:      "router.tmpl",
		DestinationPath: "internal/router/router.go",
		IsInternalFile:  true,
	},
	"config": {
		SourcePath:      "config.tmpl",
		DestinationPath: "internal/config/config.go",
		IsInternalFile:  true,
	},
	"middleware": {
		SourcePath:      "middleware.tmpl",
		DestinationPath: "internal/middleware/middleware.go",
		IsInternalFile:  true,
	},
	"logger": {
		SourcePath:      "logger.tmpl",
		DestinationPath: "internal/logger/logger.go",
		IsInternalFile:  true,
	},
	"logging": {
		SourcePath:      "logger.tmpl",
		DestinationPath: "internal/logger/logger.go",
		IsInternalFile:  true,
	},

	// Database related
	"db": {
		SourcePath:      "db.tmpl",
		DestinationPath: "internal/db/db.go",
		IsInternalFile:  true,
	},
	"postgresql": {
		SourcePath:      "db.tmpl",
		DestinationPath: "internal/db/db.go",
		IsInternalFile:  true,
	},
	"mysql": {
		SourcePath:      "mysql.tmpl",
		DestinationPath: "internal/db/mysql.go",
		IsInternalFile:  true,
	},
	"sqlite": {
		SourcePath:      "sqlite.tmpl",
		DestinationPath: "internal/db/sqlite.go",
		IsInternalFile:  true,
	},
	"mongodb": {
		SourcePath:      "mongodb.tmpl",
		DestinationPath: "internal/db/mongodb.go",
		IsInternalFile:  true,
	},
	"redis": {
		SourcePath:      "redis.tmpl",
		DestinationPath: "internal/db/redis.go",
		IsInternalFile:  true,
	},
	"badger": {
		SourcePath:      "badger.tmpl",
		DestinationPath: "internal/db/badger.go",
		IsInternalFile:  true,
	},
	"migrations": {
		SourcePath:      "migrations.tmpl",
		DestinationPath: "internal/db/migrations.go",
		IsInternalFile:  true,
	},

	// ORM related
	"gorm": {
		SourcePath:      "gorm.tmpl",
		DestinationPath: "internal/db/gorm.go",
		IsInternalFile:  true,
	},
	"sqlc": {
		SourcePath:      "sqlc.tmpl",
		DestinationPath: "internal/db/sqlc.go",
		IsInternalFile:  true,
	},

	// Authentication
	"auth": {
		SourcePath:      "auth.tmpl",
		DestinationPath: "internal/auth/auth.go",
		IsInternalFile:  true,
	},
	"oauth2": {
		SourcePath:      "oauth2.tmpl",
		DestinationPath: "internal/auth/oauth2.go",
		IsInternalFile:  true,
	},

	// AI Integration
	"openai": {
		SourcePath:      "openai.tmpl",
		DestinationPath: "internal/ai/openai.go",
		IsInternalFile:  true,
	},
	"openrouter": {
		SourcePath:      "openrouter.tmpl",
		DestinationPath: "internal/ai/openrouter.go",
		IsInternalFile:  true,
	},
	"claude": {
		SourcePath:      "claude.tmpl",
		DestinationPath: "internal/ai/claude.go",
		IsInternalFile:  true,
	},

	// Web frameworks
	"gin": {
		SourcePath:      "gin.tmpl",
		DestinationPath: "internal/framework/gin.go",
		IsInternalFile:  true,
	},
	"echo": {
		SourcePath:      "echo.tmpl",
		DestinationPath: "internal/framework/echo.go",
		IsInternalFile:  true,
	},
	"fiber": {
		SourcePath:      "fiber.tmpl",
		DestinationPath: "internal/framework/fiber.go",
		IsInternalFile:  true,
	},

	// API Documentation
	"swagger": {
		SourcePath:      "swagger.tmpl",
		DestinationPath: "internal/docs/swagger.go",
		IsInternalFile:  true,
	},
	"redoc": {
		SourcePath:      "redoc.tmpl",
		DestinationPath: "internal/docs/redoc.go",
		IsInternalFile:  true,
	},
	"postman": {
		SourcePath:      "postman.tmpl",
		DestinationPath: "docs/postman_collection.json",
		IsDocsFile:      true,
	},

	// gRPC
	"grpc": {
		SourcePath:      "grpc.tmpl",
		DestinationPath: "internal/grpc/grpc.go",
		IsInternalFile:  true,
	},
	"proto": {
		SourcePath:      "proto.tmpl",
		DestinationPath: "proto/service.proto",
		IsRootFile:      true,
	},

	// WebSocket
	"websocket": {
		SourcePath:      "websocket.tmpl",
		DestinationPath: "internal/websocket/websocket.go",
		IsInternalFile:  true,
	},
	"websockets": {
		SourcePath:      "websocket.tmpl",
		DestinationPath: "internal/websocket/websocket.go",
		IsInternalFile:  true,
	},

	// CLI tools
	"cobra": {
		SourcePath:      "cobra.tmpl",
		DestinationPath: "cmd/cli/main.go",
		IsCmdFile:       true,
	},
	"urfave-cli": {
		SourcePath:      "urfave-cli.tmpl",
		DestinationPath: "cmd/cli/main.go",
		IsCmdFile:       true,
	},

	// Testing
	"testify": {
		SourcePath:      "testify.tmpl",
		DestinationPath: "internal/testing/testify.go",
		IsInternalFile:  true,
	},
	"gomock": {
		SourcePath:      "gomock.tmpl",
		DestinationPath: "internal/testing/gomock.go",
		IsInternalFile:  true,
	},

	// Observability
	"observability": {
		SourcePath:      "observability.tmpl",
		DestinationPath: "internal/observability/observability.go",
		IsInternalFile:  true,
	},
	"prometheus": {
		SourcePath:      "observability.tmpl",
		DestinationPath: "internal/observability/observability.go",
		IsInternalFile:  true,
	},

	// Generator config
	"generator-config": {
		SourcePath:      "generator-config.tmpl",
		DestinationPath: "internal/config/generator.go",
		IsInternalFile:  true,
	},
}

func GenerateProject(projectName string, features []string) (string, error) {
	// Fallback to a UUID if no project name is provided
	id := projectName
	if id == "" {
		id = uuid.New().String()
	}

	// Sanitize project name for file system (replace spaces and special characters)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ToLower(id)

	projectPath := filepath.Join("output", id)

	// Create project directory
	err := os.MkdirAll(projectPath, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("failed to create project directory: %v", err)
	}

	// Track which modules are selected for main.go
	flags := map[string]bool{
		"Auth":          false,
		"DB":            false,
		"Router":        false,
		"OpenAI":        false,
		"gRPC":          false,
		"WebSocket":     false,
		"CLI":           false,
		"Testing":       false,
		"Observability": false,
		"Logging":       false,
		"Middleware":    false,
		"Config":        false,
	}

	// Pre-process features to set flags before template generation
	for _, feature := range features {
		updateFlags(strings.ToLower(feature), flags)
	}

	// Always generate .env.example file first
	envTemplatePath := filepath.Join("internal", "templates", "env.tmpl")
	envDestPath := filepath.Join(projectPath, ".env.example")
	
	templateData := map[string]interface{}{
		"ProjectName": projectName,
		"Features":    features,
		"Flags":       flags,
	}

	if _, err := os.Stat(envTemplatePath); err == nil {
		err = utils.ApplyTemplate(envTemplatePath, envDestPath, templateData)
		if err != nil {
			return "", fmt.Errorf("failed to render .env.example: %v", err)
		}
	}

	// Always generate .gitignore file
	gitignoreTemplatePath := filepath.Join("internal", "templates", "gitignore.tmpl")
	gitignoreDestPath := filepath.Join(projectPath, ".gitignore")
	
	if _, err := os.Stat(gitignoreTemplatePath); err == nil {
		err = utils.ApplyTemplate(gitignoreTemplatePath, gitignoreDestPath, templateData)
		if err != nil {
			return "", fmt.Errorf("failed to render .gitignore: %v", err)
		}
	}

	// Process each feature
	for _, feature := range features {
		feature = strings.ToLower(feature)

		// Update flags for main.go
		updateFlags(feature, flags)

		// Handle special cases
		if feature == "postgresql" {
			feature = "db" // Map postgresql to db template
		}

		// Skip env and gitignore as they're already handled above
		if feature == "env" || feature == "env-config" || feature == "gitignore" {
			continue
		}

		// Get template mapping
		mapping, exists := featureMappings[feature]
		if !exists {
			// Skip silently if template doesn't exist
			continue
		}

		// Build full paths
		sourcePath := filepath.Join("internal", "templates", mapping.SourcePath)
		destPath := filepath.Join(projectPath, mapping.DestinationPath)

		// Create destination directory if needed
		destDir := filepath.Dir(destPath)
		err := os.MkdirAll(destDir, os.ModePerm)
		if err != nil {
			return "", fmt.Errorf("failed to create directory %s: %v", destDir, err)
		}

		// Check if template file exists
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			// Skip silently if template file doesn't exist
			continue
		}

		// Update template data for this specific feature
		templateData := map[string]interface{}{
			"ProjectName": projectName,
			"Features":    features,
			"Flags":       flags,
		}

		// Render template
		err = utils.ApplyTemplate(sourcePath, destPath, templateData)
		if err != nil {
			return "", fmt.Errorf("failed to render template %s: %v", sourcePath, err)
		}
	}

	// Render main.go with feature flags if main template exists
	mainTemplatePath := filepath.Join("internal", "templates", "main.tmpl")
	if _, err := os.Stat(mainTemplatePath); err == nil {
		// Create cmd directory structure
		cmdDir := filepath.Join(projectPath, "cmd", projectName)
		err = os.MkdirAll(cmdDir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create cmd directory: %v", err)
		}
		
		mainDestPath := filepath.Join(cmdDir, "main.go")
		templateData := map[string]interface{}{
			"ProjectName": projectName,
			"Features":    features,
			"Flags":       flags,
		}

		err = utils.ApplyTemplate(mainTemplatePath, mainDestPath, templateData)
		if err != nil {
			return "", fmt.Errorf("failed to render main.go: %v", err)
		}
	}

	// Create go.mod file
	err = createGoMod(projectPath, features)
	if err != nil {
		return "", fmt.Errorf("failed to create go.mod: %v", err)
	}

	// Zip the folder and return path
	zipPath := filepath.Join("output", id+".zip")
	err = utils.ZipFolder(projectPath, zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip: %v", err)
	}

	// Schedule cleanup of old files (older than 1 hour)
	go func() {
		if err := utils.CleanupOldFiles("output", 1*time.Hour); err != nil {
			fmt.Printf("Cleanup error: %v\n", err)
		}
	}()

	// Schedule cleanup of current files after 10 minutes (enough time for download)
	utils.CleanupAfterDownload(zipPath, projectPath, 10*time.Minute)

	return zipPath, nil
}

func updateFlags(feature string, flags map[string]bool) {
	switch feature {
	case "auth", "oauth2", "jwt", "authentication":
		flags["Auth"] = true
	case "db", "database", "postgresql", "mysql", "sqlite", "mongodb", "redis", "badger", "gorm", "sqlc", "migrations":
		flags["DB"] = true
	case "router", "gin", "echo", "fiber", "mux":
		flags["Router"] = true
	case "openai", "openrouter", "claude", "ai", "llm":
		flags["OpenAI"] = true
	case "grpc", "proto", "protobuf":
		flags["gRPC"] = true
	case "websocket", "websockets", "ws":
		flags["WebSocket"] = true
	case "cobra", "urfave-cli", "cli":
		flags["CLI"] = true
	case "testify", "gomock", "testing", "test":
		flags["Testing"] = true
	case "observability", "prometheus", "monitoring", "metrics", "tracing":
		flags["Observability"] = true
	case "logging", "logger", "log":
		flags["Logging"] = true
	case "middleware":
		flags["Middleware"] = true
	case "config", "configuration":
		flags["Config"] = true
	}
}

func createGoMod(projectPath string, features []string) error {
	// Extract project name from path
	projectName := filepath.Base(projectPath)
	
	// Create a proper module name (avoid single words that might conflict)
	moduleName := projectName
	if projectName == "go" || len(projectName) < 3 {
		moduleName = fmt.Sprintf("github.com/user/%s", projectName)
	}
	
	// Create a basic go.mod file
	goModContent := fmt.Sprintf("module %s\n\n", moduleName)
	goModContent += "go 1.21\n\n"
	goModContent += "require (\n"
	goModContent += "\tgithub.com/gofiber/fiber/v2 v2.52.0\n"

	// Add dependencies based on features
	dependencies := getDependencies(features)
	for _, dep := range dependencies {
		goModContent += dep + "\n"
	}

	goModContent += ")\n"

	goModPath := filepath.Join(projectPath, "go.mod")
	return os.WriteFile(goModPath, []byte(goModContent), 0644)
}

func getDependencies(features []string) []string {
	var deps []string

	for _, feature := range features {
		feature = strings.ToLower(feature)

		switch feature {
		case "swagger":
			deps = append(deps, "\tgithub.com/gofiber/swagger v0.1.14")
		case "grpc":
			deps = append(deps, "\tgoogle.golang.org/grpc v1.60.1")
			deps = append(deps, "\tgoogle.golang.org/protobuf v1.32.0")
		case "gin":
			deps = append(deps, "\tgithub.com/gin-gonic/gin v1.9.1")
		case "echo":
			deps = append(deps, "\tgithub.com/labstack/echo/v4 v4.11.4")
		case "fiber":
			deps = append(deps, "\tgithub.com/gofiber/fiber/v2 v2.52.0")
		case "postgresql":
			deps = append(deps, "\tgithub.com/lib/pq v1.10.9")
		case "mysql":
			deps = append(deps, "\tgithub.com/go-sql-driver/mysql v1.7.1")
		case "gorm":
			deps = append(deps, "\tgorm.io/gorm v1.25.5")
		case "redis":
			deps = append(deps, "\tgithub.com/redis/go-redis/v9 v9.3.1")
		case "mongodb":
			deps = append(deps, "\tgo.mongodb.org/mongo-driver v1.13.1")
		case "badger":
			deps = append(deps, "\tgithub.com/dgraph-io/badger/v4 v4.2.0")
		case "cobra":
			deps = append(deps, "\tgithub.com/spf13/cobra v1.8.0")
		case "urfave-cli":
			deps = append(deps, "\tgithub.com/urfave/cli/v2 v2.27.1")
		case "testify":
			deps = append(deps, "\tgithub.com/stretchr/testify v1.8.4")
		case "gomock":
			deps = append(deps, "\tgo.uber.org/mock v0.4.0")
		case "observability":
			deps = append(deps, "\tgo.uber.org/zap v1.26.0")
			deps = append(deps, "\tgithub.com/prometheus/client_golang v1.17.0")
		}
	}

	return deps
}
