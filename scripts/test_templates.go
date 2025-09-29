package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TestData represents test data for template rendering
type TestData struct {
	ProjectName string
	Features    []string
	Flags       map[string]bool
	Auth        bool
	DB          bool
	Router      bool
	Config      bool
	Logger      bool
}

// TemplateTest represents a template test case
type TemplateTest struct {
	Name        string
	TemplatePath string
	OutputPath   string
	Data         TestData
	ShouldPass   bool
	Description  string
}

func main() {
	fmt.Println("🧪 Starting Template Testing Suite...")
	
	// Create test data scenarios
	testScenarios := []TestData{
		{
			ProjectName: "test-basic",
			Features:    []string{"fiber", "postgresql", "gorm"},
			Flags: map[string]bool{
				"Auth":   true,
				"DB":     true,
				"Router": true,
				"Config": true,
				"Logger": true,
			},
			Auth:   true,
			DB:     true,
			Router: true,
			Config: true,
			Logger: true,
		},
		{
			ProjectName: "test-minimal",
			Features:    []string{"gin", "sqlite"},
			Flags: map[string]bool{
				"Auth":   false,
				"DB":     true,
				"Router": true,
				"Config": false,
				"Logger": false,
			},
			Auth:   false,
			DB:     true,
			Router: true,
			Config: false,
			Logger: false,
		},
		{
			ProjectName: "test-full-stack",
			Features:    []string{"fiber", "postgresql", "gorm", "redis", "mongodb", "websocket", "grpc"},
			Flags: map[string]bool{
				"Auth":   true,
				"DB":     true,
				"Router": true,
				"Config": true,
				"Logger": true,
			},
			Auth:   true,
			DB:     true,
			Router: true,
			Config: true,
			Logger: true,
		},
	}

	// Find all template files
	templateDir := "internal/templates"
	templates, err := findTemplateFiles(templateDir)
	if err != nil {
		log.Fatalf("Failed to find template files: %v", err)
	}

	fmt.Printf("📁 Found %d template files\n", len(templates))

	// Create output directory
	outputDir := "test_output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Test each template with each scenario
	totalTests := 0
	passedTests := 0
	failedTests := []string{}

	for _, scenario := range testScenarios {
		scenarioDir := filepath.Join(outputDir, scenario.ProjectName)
		if err := os.MkdirAll(scenarioDir, 0755); err != nil {
			log.Printf("Failed to create scenario directory: %v", err)
			continue
		}

		fmt.Printf("\n🎯 Testing scenario: %s\n", scenario.ProjectName)

		for _, templatePath := range templates {
			totalTests++
			templateName := filepath.Base(templatePath)
			outputPath := filepath.Join(scenarioDir, strings.TrimSuffix(templateName, ".tmpl"))

			fmt.Printf("  📝 Testing %s... ", templateName)

			if err := testTemplate(templatePath, outputPath, scenario); err != nil {
				fmt.Printf("❌ FAILED: %v\n", err)
				failedTests = append(failedTests, fmt.Sprintf("%s/%s: %v", scenario.ProjectName, templateName, err))
			} else {
				fmt.Printf("✅ PASSED\n")
				passedTests++
			}
		}
	}

	// Print summary
	fmt.Printf("\n📊 Test Summary:\n")
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

	fmt.Printf("\n🎉 All templates passed validation!\n")

	// Run additional validation tests
	fmt.Printf("\n🔍 Running additional validation tests...\n")
	runValidationTests(outputDir, testScenarios)
}

func findTemplateFiles(dir string) ([]string, error) {
	var templates []string
	
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && strings.HasSuffix(path, ".tmpl") {
			templates = append(templates, path)
		}
		
		return nil
	})
	
	return templates, err
}

func testTemplate(templatePath, outputPath string, data TestData) error {
	// Template helper functions
	funcMap := template.FuncMap{
		"contains": func(slice []string, item string) bool {
			for _, s := range slice {
				if strings.EqualFold(s, item) {
					return true
				}
			}
			return false
		},
		"hasPrefix": func(s, prefix string) bool {
			return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
		},
		"hasSuffix": func(s, suffix string) bool {
			return strings.HasSuffix(strings.ToLower(s), strings.ToLower(suffix))
		},
		"toLower": strings.ToLower,
		"toUpper": strings.ToUpper,
		"replace": strings.ReplaceAll,
	}

	// Parse template
	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(funcMap).ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Execute template
	templateName := filepath.Base(templatePath)
	if err := tmpl.ExecuteTemplate(outFile, templateName, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

func runValidationTests(outputDir string, scenarios []TestData) {
	fmt.Printf("  🔍 Checking for syntax errors in generated Go files...\n")
	
	for _, scenario := range scenarios {
		scenarioDir := filepath.Join(outputDir, scenario.ProjectName)
		
		err := filepath.WalkDir(scenarioDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			
			if !d.IsDir() && strings.HasSuffix(path, ".go") {
				if err := validateGoSyntax(path); err != nil {
					fmt.Printf("    ❌ Syntax error in %s: %v\n", path, err)
				} else {
					fmt.Printf("    ✅ %s\n", filepath.Base(path))
				}
			}
			
			return nil
		})
		
		if err != nil {
			fmt.Printf("    ❌ Error walking directory %s: %v\n", scenarioDir, err)
		}
	}

	fmt.Printf("  🔍 Checking Docker files...\n")
	validateDockerFiles(outputDir, scenarios)

	fmt.Printf("  🔍 Checking configuration files...\n")
	validateConfigFiles(outputDir, scenarios)
}

func validateGoSyntax(filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Basic syntax checks
	contentStr := string(content)
	
	// Check for common syntax issues
	if strings.Contains(contentStr, "{{") && strings.Contains(contentStr, "}}") {
		return fmt.Errorf("unprocessed template variables found")
	}
	
	// Check for balanced braces (basic check)
	openBraces := strings.Count(contentStr, "{")
	closeBraces := strings.Count(contentStr, "}")
	if openBraces != closeBraces {
		return fmt.Errorf("unbalanced braces: %d open, %d close", openBraces, closeBraces)
	}
	
	// Check for balanced parentheses
	openParens := strings.Count(contentStr, "(")
	closeParens := strings.Count(contentStr, ")")
	if openParens != closeParens {
		return fmt.Errorf("unbalanced parentheses: %d open, %d close", openParens, closeParens)
	}

	return nil
}

func validateDockerFiles(outputDir string, scenarios []TestData) {
	for _, scenario := range scenarios {
		scenarioDir := filepath.Join(outputDir, scenario.ProjectName)
		
		// Check Dockerfile
		dockerfilePath := filepath.Join(scenarioDir, "dockerfile")
		if _, err := os.Stat(dockerfilePath); err == nil {
			if err := validateDockerfile(dockerfilePath); err != nil {
				fmt.Printf("    ❌ Dockerfile error: %v\n", err)
			} else {
				fmt.Printf("    ✅ Dockerfile\n")
			}
		}
		
		// Check docker-compose.yml
		composePath := filepath.Join(scenarioDir, "docker-compose")
		if _, err := os.Stat(composePath); err == nil {
			fmt.Printf("    ✅ docker-compose.yml\n")
		}
	}
}

func validateDockerfile(filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	contentStr := string(content)
	
	// Check for required Dockerfile instructions
	requiredInstructions := []string{"FROM", "WORKDIR", "COPY", "RUN", "EXPOSE", "CMD"}
	for _, instruction := range requiredInstructions {
		if !strings.Contains(contentStr, instruction) {
			return fmt.Errorf("missing required instruction: %s", instruction)
		}
	}

	return nil
}

func validateConfigFiles(outputDir string, scenarios []TestData) {
	for _, scenario := range scenarios {
		scenarioDir := filepath.Join(outputDir, scenario.ProjectName)
		
		// Check various config files
		configFiles := []string{"env", "makefile", "gitignore"}
		
		for _, configFile := range configFiles {
			configPath := filepath.Join(scenarioDir, configFile)
			if _, err := os.Stat(configPath); err == nil {
				fmt.Printf("    ✅ %s\n", configFile)
			}
		}
	}
}