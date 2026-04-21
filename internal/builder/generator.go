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
	IsFrameworkMain bool // Framework templates that replace main.go
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

	// Web frameworks — these replace main.go with a richer setup
	"gin": {
		SourcePath:      "gin.tmpl",
		IsFrameworkMain: true,
	},
	"echo": {
		SourcePath:      "echo.tmpl",
		IsFrameworkMain: true,
	},
	"fiber": {
		SourcePath:      "fiber.tmpl",
		IsFrameworkMain: true,
	},

	// Internal packages
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
	// migrations generates a SQL file, not a .go file
	"migrations": {
		SourcePath:      "migrations.tmpl",
		DestinationPath: "migrations/001_init.sql",
		IsRootFile:      true,
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
		DestinationPath: "internal/testing/helpers_test.go",
		IsInternalFile:  true,
	},
	"gomock": {
		SourcePath:      "gomock.tmpl",
		DestinationPath: "internal/testing/mocks_test.go",
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

func outputDir() string {
	if dir := os.Getenv("OUTPUT_DIR"); dir != "" {
		return dir
	}
	return "output"
}

func GenerateProject(projectName string, features []string) (string, error) {
	id := projectName
	if id == "" {
		id = uuid.New().String()
	}

	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ToLower(id)

	base := outputDir()
	projectPath := filepath.Join(base, id)

	if err := os.MkdirAll(projectPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create project directory: %v", err)
	}

	// Build feature flags once before any rendering
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
	for _, feature := range features {
		updateFlags(strings.ToLower(feature), flags)
	}

	// Build template data once — include both Flags.X and flat .X for template compatibility
	templateData := map[string]interface{}{
		"ProjectName":   projectName,
		"Features":      features,
		"Flags":         flags,
		"Auth":          flags["Auth"],
		"DB":            flags["DB"],
		"Router":        flags["Router"],
		"OpenAI":        flags["OpenAI"],
		"GRPC":          flags["gRPC"],
		"WebSocket":     flags["WebSocket"],
		"CLI":           flags["CLI"],
		"Testing":       flags["Testing"],
		"Observability": flags["Observability"],
		"Logger":        flags["Logging"],
		"Logging":       flags["Logging"],
		"Middleware":    flags["Middleware"],
		"Config":        flags["Config"],
	}

	// Always generate .env.example and .gitignore
	for _, pair := range []struct{ tmpl, dest string }{
		{"env.tmpl", ".env.example"},
		{"gitignore.tmpl", ".gitignore"},
	} {
		src := filepath.Join("internal", "templates", pair.tmpl)
		if _, err := os.Stat(src); err == nil {
			if err := utils.ApplyTemplate(src, filepath.Join(projectPath, pair.dest), templateData); err != nil {
				return "", fmt.Errorf("failed to render %s: %v", pair.dest, err)
			}
		}
	}

	// Process each feature
	var frameworkMainRendered bool
	for _, feature := range features {
		feature = strings.ToLower(feature)

		// postgresql is an alias for the db template
		if feature == "postgresql" {
			feature = "db"
		}

		// .env.example and .gitignore already written above
		if feature == "env" || feature == "env-config" || feature == "gitignore" {
			continue
		}

		mapping, exists := featureMappings[feature]
		if !exists {
			continue
		}

		sourcePath := filepath.Join("internal", "templates", mapping.SourcePath)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			continue
		}

		var destPath string
		if mapping.IsFrameworkMain {
			// Framework templates generate the main entry point
			cmdDir := filepath.Join(projectPath, "cmd", projectName)
			if err := os.MkdirAll(cmdDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create cmd directory: %v", err)
			}
			destPath = filepath.Join(cmdDir, "main.go")
			frameworkMainRendered = true
		} else {
			destPath = filepath.Join(projectPath, mapping.DestinationPath)
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return "", fmt.Errorf("failed to create directory for %s: %v", destPath, err)
			}
		}

		if err := utils.ApplyTemplate(sourcePath, destPath, templateData); err != nil {
			return "", fmt.Errorf("failed to render template %s: %v", sourcePath, err)
		}
	}

	// Fall back to main.tmpl if no framework template was used
	if !frameworkMainRendered {
		mainTemplatePath := filepath.Join("internal", "templates", "main.tmpl")
		if _, err := os.Stat(mainTemplatePath); err == nil {
			cmdDir := filepath.Join(projectPath, "cmd", projectName)
			if err := os.MkdirAll(cmdDir, 0755); err != nil {
				return "", fmt.Errorf("failed to create cmd directory: %v", err)
			}
			mainDestPath := filepath.Join(cmdDir, "main.go")
			if err := utils.ApplyTemplate(mainTemplatePath, mainDestPath, templateData); err != nil {
				return "", fmt.Errorf("failed to render main.go: %v", err)
			}
		}
	}

	if err := createGoMod(projectPath, projectName, features); err != nil {
		return "", fmt.Errorf("failed to create go.mod: %v", err)
	}

	zipPath := filepath.Join(base, id+".zip")
	if err := utils.ZipFolder(projectPath, zipPath); err != nil {
		return "", fmt.Errorf("failed to create zip: %v", err)
	}

	go func() {
		if err := utils.CleanupOldFiles(base, 1*time.Hour); err != nil {
			fmt.Printf("Cleanup error: %v\n", err)
		}
	}()

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

func createGoMod(projectPath, projectName string, features []string) error {
	moduleName := projectName
	// Ensure the module name is valid (not a bare reserved word, not too short)
	if len(moduleName) < 2 || moduleName == "go" {
		moduleName = fmt.Sprintf("github.com/user/%s", projectName)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("module %s\n\ngo 1.21\n\nrequire (\n", moduleName))

	seen := map[string]bool{}
	addDep := func(dep string) {
		if !seen[dep] {
			seen[dep] = true
			sb.WriteString("\t" + dep + "\n")
		}
	}

	// Always include fiber as the default HTTP framework
	addDep("github.com/gofiber/fiber/v2 v2.52.0")
	addDep("github.com/joho/godotenv v1.5.1")

	for _, feature := range features {
		switch strings.ToLower(feature) {
		case "gin":
			addDep("github.com/gin-gonic/gin v1.9.1")
			addDep("github.com/gin-contrib/cors v1.4.0")
		case "echo":
			addDep("github.com/labstack/echo/v4 v4.11.4")
		case "fiber":
			// already included above
		case "auth", "jwt", "authentication":
			addDep("github.com/golang-jwt/jwt/v5 v5.2.0")
			addDep("golang.org/x/crypto v0.17.0")
		case "oauth2":
			addDep("golang.org/x/oauth2 v0.15.0")
			addDep("github.com/golang-jwt/jwt/v5 v5.2.0")
		case "gorm":
			addDep("gorm.io/gorm v1.25.5")
			addDep("gorm.io/driver/postgres v1.5.4")
			addDep("gorm.io/driver/mysql v1.5.2")
			addDep("gorm.io/driver/sqlite v1.5.4")
		case "postgresql", "db", "database":
			addDep("github.com/lib/pq v1.10.9")
		case "mysql":
			addDep("github.com/go-sql-driver/mysql v1.7.1")
		case "sqlite":
			addDep("github.com/mattn/go-sqlite3 v1.14.19")
		case "mongodb":
			addDep("go.mongodb.org/mongo-driver v1.13.1")
		case "redis":
			addDep("github.com/go-redis/redis/v8 v8.11.5")
		case "badger":
			addDep("github.com/dgraph-io/badger/v4 v4.2.0")
		case "swagger":
			addDep("github.com/gofiber/swagger v0.1.14")
			addDep("github.com/swaggo/swag v1.16.3")
		case "grpc", "protobuf":
			addDep("google.golang.org/grpc v1.60.1")
			addDep("google.golang.org/protobuf v1.32.0")
		case "websocket", "websockets":
			addDep("github.com/gofiber/websocket/v2 v2.2.1")
			addDep("github.com/gorilla/websocket v1.5.1")
		case "openai":
			addDep("github.com/sashabaranov/go-openai v1.17.9")
		case "openrouter":
			addDep("github.com/sashabaranov/go-openai v1.17.9")
		case "claude":
			addDep("github.com/anthropics/anthropic-sdk-go v0.1.0")
		case "logger", "logging":
			addDep("go.uber.org/zap v1.26.0")
			addDep("gopkg.in/natefinch/lumberjack.v2 v2.2.1")
		case "observability", "prometheus":
			addDep("go.uber.org/zap v1.26.0")
			addDep("github.com/prometheus/client_golang v1.17.0")
		case "cobra":
			addDep("github.com/spf13/cobra v1.8.0")
		case "urfave-cli":
			addDep("github.com/urfave/cli/v2 v2.27.1")
		case "testify", "testing":
			addDep("github.com/stretchr/testify v1.8.4")
		case "gomock":
			addDep("go.uber.org/mock v0.4.0")
		}
	}

	sb.WriteString(")\n")

	return os.WriteFile(filepath.Join(projectPath, "go.mod"), []byte(sb.String()), 0644)
}
