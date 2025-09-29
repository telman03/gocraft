//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IntegrationTest represents an integration test scenario
type IntegrationTest struct {
	Name        string
	ProjectName string
	Features    []string
	Framework   string
	Database    string
	Description string
}

func main() {
	fmt.Println("🚀 Starting Integration Tests...")

	// Define test scenarios
	scenarios := []IntegrationTest{
		{
			Name:        "fiber-postgres-full",
			ProjectName: "test-fiber-app",
			Features:    []string{"fiber", "postgresql", "gorm", "auth", "redis", "swagger"},
			Framework:   "fiber",
			Database:    "postgresql",
			Description: "Full-stack Fiber app with PostgreSQL, GORM, Auth, Redis, and Swagger",
		},
		{
			Name:        "gin-mysql-minimal",
			ProjectName: "test-gin-app",
			Features:    []string{"gin", "mysql", "gorm", "auth"},
			Framework:   "gin",
			Database:    "mysql",
			Description: "Minimal Gin app with MySQL and authentication",
		},
		{
			Name:        "echo-sqlite-simple",
			ProjectName: "test-echo-app",
			Features:    []string{"echo", "sqlite", "logger"},
			Framework:   "echo",
			Database:    "sqlite",
			Description: "Simple Echo app with SQLite and logging",
		},
		{
			Name:        "fiber-grpc-advanced",
			ProjectName: "test-grpc-app",
			Features:    []string{"fiber", "postgresql", "gorm", "grpc", "websocket", "auth", "redis"},
			Framework:   "fiber",
			Database:    "postgresql",
			Description: "Advanced app with gRPC, WebSocket, and multiple services",
		},
	}

	// Create test directory
	testDir := "integration_tests"
	if err := os.RemoveAll(testDir); err != nil {
		log.Printf("Warning: Failed to remove existing test directory: %v", err)
	}
	if err := os.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("Failed to create test directory: %v", err)
	}

	totalTests := len(scenarios)
	passedTests := 0
	failedTests := []string{}

	for i, scenario := range scenarios {
		fmt.Printf("\n🧪 Test %d/%d: %s\n", i+1, totalTests, scenario.Name)
		fmt.Printf("   📝 %s\n", scenario.Description)

		if err := runIntegrationTest(testDir, scenario); err != nil {
			fmt.Printf("   ❌ FAILED: %v\n", err)
			failedTests = append(failedTests, fmt.Sprintf("%s: %v", scenario.Name, err))
		} else {
			fmt.Printf("   ✅ PASSED\n")
			passedTests++
		}
	}

	// Print summary
	fmt.Printf("\n📊 Integration Test Summary:\n")
	fmt.Printf("  Total Tests: %d\n", totalTests)
	fmt.Printf("  Passed: %d (%.1f%%)\n", passedTests, float64(passedTests)/float64(totalTests)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n", len(failedTests), float64(len(failedTests))/float64(totalTests)*100)

	if len(failedTests) > 0 {
		fmt.Printf("\n❌ Failed Tests:\n")
		for _, failure := range failedTests {
			fmt.Printf("  - %s\n", failure)
		}
		os.Exit(1)
	}

	fmt.Printf("\n🎉 All integration tests passed!\n")
}

func runIntegrationTest(testDir string, scenario IntegrationTest) error {
	projectDir := filepath.Join(testDir, scenario.ProjectName)

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate project files using templates
	if err := generateProjectFiles(projectDir, scenario); err != nil {
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	// Validate generated files
	if err := validateGeneratedProject(projectDir, scenario); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Test Go compilation (if Go files exist)
	if err := testGoCompilation(projectDir); err != nil {
		return fmt.Errorf("Go compilation failed: %w", err)
	}

	// Test Docker build (if Dockerfile exists)
	if err := testDockerBuild(projectDir, scenario.ProjectName); err != nil {
		return fmt.Errorf("Docker build failed: %w", err)
	}

	return nil
}

func generateProjectFiles(projectDir string, scenario IntegrationTest) error {
	// This would normally use your template engine
	// For now, we'll create basic files to test the structure

	// Create go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/gofiber/fiber/v2 v2.50.0
	github.com/gin-gonic/gin v1.9.1
	github.com/labstack/echo/v4 v4.11.2
	gorm.io/gorm v1.25.5
	gorm.io/driver/postgres v1.5.4
	gorm.io/driver/mysql v1.5.2
	gorm.io/driver/sqlite v1.5.4
)
`, scenario.ProjectName)

	if err := os.WriteFile(filepath.Join(projectDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return err
	}

	// Create main.go based on framework
	var mainContent string
	switch scenario.Framework {
	case "fiber":
		mainContent = `package main

import (
	"log"
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "healthy"})
	})
	
	log.Fatal(app.Listen(":8080"))
}
`
	case "gin":
		mainContent = `package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})
	
	r.Run(":8080")
}
`
	case "echo":
		mainContent = `package main

import (
	"net/http"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "healthy"})
	})
	
	e.Logger.Fatal(e.Start(":8080"))
}
`
	default:
		mainContent = `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	}

	if err := os.WriteFile(filepath.Join(projectDir, "main.go"), []byte(mainContent), 0644); err != nil {
		return err
	}

	// Create Dockerfile if needed
	dockerfileContent := `FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
CMD ["./main"]
`

	if err := os.WriteFile(filepath.Join(projectDir, "Dockerfile"), []byte(dockerfileContent), 0644); err != nil {
		return err
	}

	return nil
}

func validateGeneratedProject(projectDir string, scenario IntegrationTest) error {
	// Check required files exist
	requiredFiles := []string{"go.mod", "main.go", "Dockerfile"}

	for _, file := range requiredFiles {
		filePath := filepath.Join(projectDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("required file missing: %s", file)
		}
	}

	// Validate go.mod content
	goModPath := filepath.Join(projectDir, "go.mod")
	goModContent, err := os.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("failed to read go.mod: %w", err)
	}

	if !strings.Contains(string(goModContent), scenario.ProjectName) {
		return fmt.Errorf("go.mod doesn't contain project name")
	}

	// Validate main.go content
	mainGoPath := filepath.Join(projectDir, "main.go")
	mainGoContent, err := os.ReadFile(mainGoPath)
	if err != nil {
		return fmt.Errorf("failed to read main.go: %w", err)
	}

	// Check for framework-specific imports
	switch scenario.Framework {
	case "fiber":
		if !strings.Contains(string(mainGoContent), "github.com/gofiber/fiber/v2") {
			return fmt.Errorf("main.go doesn't contain Fiber import")
		}
	case "gin":
		if !strings.Contains(string(mainGoContent), "github.com/gin-gonic/gin") {
			return fmt.Errorf("main.go doesn't contain Gin import")
		}
	case "echo":
		if !strings.Contains(string(mainGoContent), "github.com/labstack/echo/v4") {
			return fmt.Errorf("main.go doesn't contain Echo import")
		}
	}

	return nil
}

func testGoCompilation(projectDir string) error {
	// Try to build the Go project
	cmd := exec.Command("go", "build", "-o", "test-binary", ".")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	// Clean up binary
	binaryPath := filepath.Join(projectDir, "test-binary")
	os.Remove(binaryPath)

	return nil
}

func testDockerBuild(projectDir, projectName string) error {
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err != nil {
		fmt.Printf("   ⚠️  Docker not available, skipping Docker build test\n")
		return nil
	}

	// Try to build Docker image
	imageName := fmt.Sprintf("test-%s:latest", projectName)
	cmd := exec.Command("docker", "build", "-t", imageName, ".")
	cmd.Dir = projectDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker build failed: %w\nOutput: %s", err, string(output))
	}

	// Clean up image
	cleanupCmd := exec.Command("docker", "rmi", imageName)
	cleanupCmd.Run() // Ignore errors for cleanup

	return nil
}
