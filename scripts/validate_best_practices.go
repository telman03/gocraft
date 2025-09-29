//go:build ignore
// +build ignore

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// BestPracticeIssue represents a best practice violation
type BestPracticeIssue struct {
	File        string
	Line        int
	Category    string
	Severity    string
	Description string
	Suggestion  string
}

// BestPracticeValidator validates templates against Go and security best practices
type BestPracticeValidator struct {
	issues []BestPracticeIssue
}

func main() {
	fmt.Println("📋 Starting Best Practices Validation...")

	validator := &BestPracticeValidator{}
	templateDir := "internal/templates"

	// Find all template files
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tmpl") {
			if err := validator.validateFile(path); err != nil {
				log.Printf("Error validating %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk template directory: %v", err)
	}

	// Print results
	validator.printResults()

	// Exit with error code if critical issues found
	if validator.hasCriticalIssues() {
		os.Exit(1)
	}
}

func (v *BestPracticeValidator) validateFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		v.checkLine(filePath, lineNum, line)
	}

	return scanner.Err()
}

func (v *BestPracticeValidator) checkLine(filePath string, lineNum int, line string) {
	// Check Go best practices
	v.checkGoConventions(filePath, lineNum, line)

	// Check error handling
	v.checkErrorHandling(filePath, lineNum, line)

	// Check logging practices
	v.checkLoggingPractices(filePath, lineNum, line)

	// Check configuration management
	v.checkConfigurationManagement(filePath, lineNum, line)

	// Check database practices
	v.checkDatabasePractices(filePath, lineNum, line)

	// Check HTTP/API practices
	v.checkHTTPPractices(filePath, lineNum, line)

	// Check Docker practices
	v.checkDockerPractices(filePath, lineNum, line)

	// Check testing practices
	v.checkTestingPractices(filePath, lineNum, line)
}

func (v *BestPracticeValidator) checkGoConventions(filePath string, lineNum int, line string) {
	// Check for proper package naming
	if strings.HasPrefix(line, "package ") {
		packageName := strings.TrimSpace(strings.TrimPrefix(line, "package"))
		if strings.Contains(packageName, "_") || strings.Contains(packageName, "-") {
			v.addIssue(filePath, lineNum, "Go Conventions", "MEDIUM",
				"Package names should not contain underscores or hyphens",
				"Use lowercase letters only for package names")
		}
	}

	// Check for proper error variable naming
	if matched, _ := regexp.MatchString(`\s+err\s*:=`, line); matched {
		// Good practice
	} else if matched, _ := regexp.MatchString(`\s+(error|e)\s*:=`, line); matched {
		v.addIssue(filePath, lineNum, "Go Conventions", "LOW",
			"Use 'err' as the standard error variable name",
			"Rename error variables to 'err' for consistency")
	}

	// Check for proper context usage
	if strings.Contains(line, "context.Background()") && !strings.Contains(line, "timeout") {
		v.addIssue(filePath, lineNum, "Go Conventions", "MEDIUM",
			"Consider using context with timeout instead of Background()",
			"Use context.WithTimeout() or context.WithCancel() for better control")
	}

	// Check for proper struct field tags
	if matched, _ := regexp.MatchString(`\s+\w+\s+\w+\s+`+"`json:\"\\w+\"`", line); matched {
		// Good practice - has JSON tags
	} else if matched, _ := regexp.MatchString(`\s+\w+\s+string\s*$`, line); matched && strings.Contains(filePath, "models") {
		v.addIssue(filePath, lineNum, "Go Conventions", "LOW",
			"Consider adding struct tags for JSON serialization",
			"Add `json:\"field_name\"` tags to struct fields")
	}
}

func (v *BestPracticeValidator) checkErrorHandling(filePath string, lineNum int, line string) {
	// Check for ignored errors
	if matched, _ := regexp.MatchString(`\w+\(\)$`, line); matched && !strings.Contains(line, "defer") {
		v.addIssue(filePath, lineNum, "Error Handling", "MEDIUM",
			"Function call result ignored - potential error not handled",
			"Check and handle the error return value")
	}

	// Check for panic usage
	if strings.Contains(line, "panic(") && !strings.Contains(line, "// panic is acceptable here") {
		v.addIssue(filePath, lineNum, "Error Handling", "HIGH",
			"Using panic() - consider returning an error instead",
			"Return errors instead of panicking, except for unrecoverable situations")
	}

	// Check for proper error wrapping
	if strings.Contains(line, "fmt.Errorf") && !strings.Contains(line, "%w") {
		v.addIssue(filePath, lineNum, "Error Handling", "MEDIUM",
			"Consider using error wrapping with %w verb",
			"Use fmt.Errorf(\"message: %w\", err) to wrap errors")
	}

	// Check for empty error handling
	if strings.Contains(line, "if err != nil {") {
		// This is good, but we should check the next lines for empty handling
	}
}

func (v *BestPracticeValidator) checkLoggingPractices(filePath string, lineNum int, line string) {
	// Check for fmt.Print usage instead of proper logging
	if matched, _ := regexp.MatchString(`fmt\.Print`, line); matched && !strings.Contains(line, "debug") {
		v.addIssue(filePath, lineNum, "Logging", "MEDIUM",
			"Using fmt.Print* for logging - consider using a proper logger",
			"Use structured logging with zap, logrus, or slog")
	}

	// Check for log.Fatal usage
	if strings.Contains(line, "log.Fatal") {
		v.addIssue(filePath, lineNum, "Logging", "MEDIUM",
			"log.Fatal calls os.Exit() - consider returning error instead",
			"Return errors to allow graceful shutdown and testing")
	}

	// Check for sensitive data in logs
	sensitivePatterns := []string{"password", "secret", "token", "key"}
	for _, pattern := range sensitivePatterns {
		if strings.Contains(strings.ToLower(line), pattern) &&
			(strings.Contains(line, "log.") || strings.Contains(line, "fmt.Print")) {
			v.addIssue(filePath, lineNum, "Logging", "HIGH",
				"Potential sensitive data in logs",
				"Avoid logging sensitive information like passwords, tokens, or keys")
		}
	}
}

func (v *BestPracticeValidator) checkConfigurationManagement(filePath string, lineNum int, line string) {
	// Check for hardcoded values
	if matched, _ := regexp.MatchString(`:\d{4,5}`, line); matched && !strings.Contains(line, "getEnv") {
		v.addIssue(filePath, lineNum, "Configuration", "MEDIUM",
			"Hardcoded port number - consider making it configurable",
			"Use environment variables for configuration values")
	}

	// Check for proper environment variable usage
	if strings.Contains(line, "os.Getenv") && !strings.Contains(line, "getEnv") {
		v.addIssue(filePath, lineNum, "Configuration", "LOW",
			"Direct os.Getenv usage - consider using a helper function",
			"Use a helper function that provides default values")
	}

	// Check for missing default values
	if matched, _ := regexp.MatchString(`os\.Getenv\("[^"]+"\)$`, line); matched {
		v.addIssue(filePath, lineNum, "Configuration", "MEDIUM",
			"Environment variable without default value",
			"Provide sensible default values for configuration")
	}
}

func (v *BestPracticeValidator) checkDatabasePractices(filePath string, lineNum int, line string) {
	// Check for SQL injection vulnerabilities
	if matched, _ := regexp.MatchString(`fmt\.Sprintf.*SELECT`, line); matched {
		v.addIssue(filePath, lineNum, "Database", "HIGH",
			"Potential SQL injection - using string formatting for SQL",
			"Use parameterized queries or prepared statements")
	}

	// Check for missing connection pooling configuration
	if strings.Contains(line, "sql.Open") && !strings.Contains(filePath, "pool") {
		v.addIssue(filePath, lineNum, "Database", "MEDIUM",
			"Database connection without pool configuration",
			"Configure connection pool settings (MaxOpenConns, MaxIdleConns, ConnMaxLifetime)")
	}

	// Check for missing transaction handling
	if strings.Contains(line, "db.Exec") || strings.Contains(line, "db.Query") {
		v.addIssue(filePath, lineNum, "Database", "LOW",
			"Consider using transactions for data consistency",
			"Use database transactions for operations that modify multiple records")
	}

	// Check for missing context in database operations
	if (strings.Contains(line, ".Exec(") || strings.Contains(line, ".Query(")) &&
		!strings.Contains(line, "Context") {
		v.addIssue(filePath, lineNum, "Database", "MEDIUM",
			"Database operation without context",
			"Use ExecContext() and QueryContext() for better cancellation support")
	}
}

func (v *BestPracticeValidator) checkHTTPPractices(filePath string, lineNum int, line string) {
	// Check for missing request timeouts
	if strings.Contains(line, "http.Client") && !strings.Contains(line, "Timeout") {
		v.addIssue(filePath, lineNum, "HTTP", "MEDIUM",
			"HTTP client without timeout",
			"Set appropriate timeouts for HTTP clients")
	}

	// Check for insecure HTTP usage
	if strings.Contains(line, "http://") && !strings.Contains(line, "localhost") {
		v.addIssue(filePath, lineNum, "HTTP", "MEDIUM",
			"Using HTTP instead of HTTPS",
			"Use HTTPS for secure communication")
	}

	// Check for missing CORS configuration
	if strings.Contains(line, "AllowOrigins") && strings.Contains(line, "*") {
		v.addIssue(filePath, lineNum, "HTTP", "MEDIUM",
			"CORS allows all origins",
			"Restrict CORS to specific trusted origins")
	}

	// Check for missing rate limiting
	if strings.Contains(line, "app.Listen") || strings.Contains(line, "r.Run") {
		// Could suggest rate limiting, but this might be too noisy
	}
}

func (v *BestPracticeValidator) checkDockerPractices(filePath string, lineNum int, line string) {
	if !strings.Contains(filePath, "dockerfile") {
		return
	}

	// Check for running as root
	if strings.Contains(line, "USER root") {
		v.addIssue(filePath, lineNum, "Docker", "HIGH",
			"Running container as root user",
			"Create and use a non-root user for security")
	}

	// Check for missing health check
	if strings.Contains(line, "FROM") && !strings.Contains(line, "scratch") {
		// We'll check for HEALTHCHECK in the full file validation
	}

	// Check for using latest tag
	if matched, _ := regexp.MatchString(`FROM.*:latest`, line); matched {
		v.addIssue(filePath, lineNum, "Docker", "MEDIUM",
			"Using 'latest' tag is not recommended",
			"Pin to specific version tags for reproducible builds")
	}

	// Check for missing .dockerignore reference
	if strings.Contains(line, "COPY . .") {
		v.addIssue(filePath, lineNum, "Docker", "LOW",
			"Copying all files - ensure .dockerignore is properly configured",
			"Use .dockerignore to exclude unnecessary files")
	}
}

func (v *BestPracticeValidator) checkTestingPractices(filePath string, lineNum int, line string) {
	if !strings.Contains(filePath, "test") {
		return
	}

	// Check for proper test function naming
	if matched, _ := regexp.MatchString(`func Test\w+`, line); matched {
		// Good practice
	} else if matched, _ := regexp.MatchString(`func test\w+`, line); matched {
		v.addIssue(filePath, lineNum, "Testing", "MEDIUM",
			"Test function should start with 'Test' (capital T)",
			"Use proper test function naming: func TestFunctionName(t *testing.T)")
	}

	// Check for table-driven tests
	if strings.Contains(line, "for _, tt := range tests") {
		// Good practice - table-driven test
	}

	// Check for proper error handling in tests
	if strings.Contains(line, "if err != nil") && strings.Contains(line, "t.Fatal") {
		// Good practice
	} else if strings.Contains(line, "if err != nil") && !strings.Contains(line, "t.") {
		v.addIssue(filePath, lineNum, "Testing", "MEDIUM",
			"Error in test not properly handled",
			"Use t.Fatal() or t.Error() to handle test errors")
	}
}

func (v *BestPracticeValidator) addIssue(filePath string, lineNum int, category, severity, description, suggestion string) {
	v.issues = append(v.issues, BestPracticeIssue{
		File:        filePath,
		Line:        lineNum,
		Category:    category,
		Severity:    severity,
		Description: description,
		Suggestion:  suggestion,
	})
}

func (v *BestPracticeValidator) printResults() {
	if len(v.issues) == 0 {
		fmt.Println("✅ No best practice violations found!")
		return
	}

	// Group issues by category and severity
	categories := make(map[string]map[string][]BestPracticeIssue)

	for _, issue := range v.issues {
		if categories[issue.Category] == nil {
			categories[issue.Category] = make(map[string][]BestPracticeIssue)
		}
		categories[issue.Category][issue.Severity] = append(categories[issue.Category][issue.Severity], issue)
	}

	// Print summary
	fmt.Printf("\n📊 Best Practices Validation Summary:\n")
	fmt.Printf("  Total Issues: %d\n", len(v.issues))

	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, severityMap := range categories {
		highCount += len(severityMap["HIGH"])
		mediumCount += len(severityMap["MEDIUM"])
		lowCount += len(severityMap["LOW"])
	}

	fmt.Printf("  High Priority: %d\n", highCount)
	fmt.Printf("  Medium Priority: %d\n", mediumCount)
	fmt.Printf("  Low Priority: %d\n", lowCount)

	// Print issues by category
	for category, severityMap := range categories {
		fmt.Printf("\n📂 %s Issues:\n", category)

		for _, severity := range []string{"HIGH", "MEDIUM", "LOW"} {
			issues := severityMap[severity]
			if len(issues) == 0 {
				continue
			}

			var emoji string
			switch severity {
			case "HIGH":
				emoji = "🚨"
			case "MEDIUM":
				emoji = "⚠️"
			case "LOW":
				emoji = "ℹ️"
			}

			fmt.Printf("  %s %s Priority:\n", emoji, severity)
			for _, issue := range issues {
				fmt.Printf("    📁 %s:%d\n", issue.File, issue.Line)
				fmt.Printf("       Issue: %s\n", issue.Description)
				fmt.Printf("       Fix: %s\n", issue.Suggestion)
				fmt.Println()
			}
		}
	}
}

func (v *BestPracticeValidator) hasCriticalIssues() bool {
	for _, issue := range v.issues {
		if issue.Severity == "HIGH" {
			return true
		}
	}
	return false
}
