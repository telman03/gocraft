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

// SecurityIssue represents a security issue found in templates
type SecurityIssue struct {
	File        string
	Line        int
	Type        string
	Severity    string
	Description string
	Suggestion  string
}

// SecurityAuditor performs security audits on templates
type SecurityAuditor struct {
	issues []SecurityIssue
}

func main() {
	fmt.Println("🔒 Starting Security Audit of Templates...")

	auditor := &SecurityAuditor{}
	templateDir := "internal/templates"

	// Find all template files
	err := filepath.Walk(templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".tmpl") {
			if err := auditor.auditFile(path); err != nil {
				log.Printf("Error auditing %s: %v", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Failed to walk template directory: %v", err)
	}

	// Print results
	auditor.printResults()

	// Exit with error code if high severity issues found
	if auditor.hasHighSeverityIssues() {
		os.Exit(1)
	}
}

func (sa *SecurityAuditor) auditFile(filePath string) error {
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
		sa.checkLine(filePath, lineNum, line)
	}

	return scanner.Err()
}

func (sa *SecurityAuditor) checkLine(filePath string, lineNum int, line string) {
	// Check for hardcoded secrets
	sa.checkHardcodedSecrets(filePath, lineNum, line)
	
	// Check for SQL injection vulnerabilities
	sa.checkSQLInjection(filePath, lineNum, line)
	
	// Check for XSS vulnerabilities
	sa.checkXSS(filePath, lineNum, line)
	
	// Check for insecure defaults
	sa.checkInsecureDefaults(filePath, lineNum, line)
	
	// Check for weak cryptography
	sa.checkWeakCryptography(filePath, lineNum, line)
	
	// Check for insecure network configurations
	sa.checkInsecureNetwork(filePath, lineNum, line)
	
	// Check for information disclosure
	sa.checkInformationDisclosure(filePath, lineNum, line)
	
	// Check for authentication issues
	sa.checkAuthenticationIssues(filePath, lineNum, line)
}

func (sa *SecurityAuditor) checkHardcodedSecrets(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`(?i)(password|pwd|secret|key|token)\s*[:=]\s*["'][^"']{8,}["']`: "Potential hardcoded secret",
		`(?i)api[_-]?key\s*[:=]\s*["'][^"']+["']`:                        "Potential hardcoded API key",
		`(?i)(jwt|bearer)[_-]?secret\s*[:=]\s*["'][^"']+["']`:            "Potential hardcoded JWT secret",
		`sk-[a-zA-Z0-9]{48}`:                                             "OpenAI API key pattern",
		`xoxb-[0-9]{11}-[0-9]{11}-[a-zA-Z0-9]{24}`:                      "Slack bot token pattern",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			// Skip if it's a template variable or environment variable reference
			if strings.Contains(line, "{{") || strings.Contains(line, "getEnv") || strings.Contains(line, "os.Getenv") {
				continue
			}
			
			sa.addIssue(filePath, lineNum, "Hardcoded Secret", "HIGH", description, 
				"Use environment variables or secure configuration management")
		}
	}
}

func (sa *SecurityAuditor) checkSQLInjection(filePath string, lineNum int, line string) {
	patterns := []string{
		`fmt\.Sprintf.*SELECT.*%s`,
		`fmt\.Sprintf.*INSERT.*%s`,
		`fmt\.Sprintf.*UPDATE.*%s`,
		`fmt\.Sprintf.*DELETE.*%s`,
		`"SELECT.*" \+ `,
		`"INSERT.*" \+ `,
		`"UPDATE.*" \+ `,
		`"DELETE.*" \+ `,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			sa.addIssue(filePath, lineNum, "SQL Injection", "HIGH", 
				"Potential SQL injection vulnerability", 
				"Use parameterized queries or prepared statements")
		}
	}
}

func (sa *SecurityAuditor) checkXSS(filePath string, lineNum int, line string) {
	patterns := []string{
		`innerHTML\s*=`,
		`document\.write\s*\(`,
		`eval\s*\(`,
		`\.html\(\s*[^)]*\+`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			sa.addIssue(filePath, lineNum, "XSS", "MEDIUM", 
				"Potential XSS vulnerability", 
				"Sanitize user input and use safe DOM manipulation methods")
		}
	}
}

func (sa *SecurityAuditor) checkInsecureDefaults(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`sslmode\s*[:=]\s*["']disable["']`:                    "SSL disabled by default",
		`TLSConfig.*InsecureSkipVerify.*true`:                 "TLS verification disabled",
		`AllowOrigins.*\*`:                                    "CORS allows all origins",
		`DEBUG.*true`:                                         "Debug mode enabled by default",
		`password.*[:=].*["']password["']`:                    "Default password used",
		`secret.*[:=].*["']secret["']`:                        "Default secret used",
		`bcrypt.*cost.*[1-9](?![0-9])`:                       "Weak bcrypt cost",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			severity := "MEDIUM"
			if strings.Contains(description, "password") || strings.Contains(description, "secret") {
				severity = "HIGH"
			}
			
			sa.addIssue(filePath, lineNum, "Insecure Default", severity, description, 
				"Use secure defaults and require explicit configuration for insecure options")
		}
	}
}

func (sa *SecurityAuditor) checkWeakCryptography(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`md5\.`:                    "MD5 is cryptographically broken",
		`sha1\.`:                   "SHA1 is cryptographically weak",
		`des\.`:                    "DES encryption is weak",
		`rc4\.`:                    "RC4 encryption is weak",
		`rand\.Read`:               "Use crypto/rand instead of math/rand for cryptographic purposes",
		`bcrypt.*cost.*[1-9](?![0-9])`: "BCrypt cost too low (should be >= 10)",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			// Skip if it's crypto/rand
			if strings.Contains(line, "crypto/rand") {
				continue
			}
			
			sa.addIssue(filePath, lineNum, "Weak Cryptography", "HIGH", description, 
				"Use strong cryptographic algorithms and sufficient key lengths")
		}
	}
}

func (sa *SecurityAuditor) checkInsecureNetwork(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`http://`:                           "Unencrypted HTTP connection",
		`ftp://`:                            "Unencrypted FTP connection",
		`telnet://`:                         "Unencrypted Telnet connection",
		`0\.0\.0\.0`:                        "Binding to all interfaces",
		`AllowInsecureConnections.*true`:    "Insecure connections allowed",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			// Skip if it's in comments or localhost
			if strings.Contains(line, "//") || strings.Contains(line, "localhost") || strings.Contains(line, "127.0.0.1") {
				continue
			}
			
			severity := "MEDIUM"
			if strings.Contains(description, "HTTP") {
				severity = "LOW" // HTTP might be acceptable in development
			}
			
			sa.addIssue(filePath, lineNum, "Insecure Network", severity, description, 
				"Use encrypted connections (HTTPS, SFTP, SSH) and bind to specific interfaces")
		}
	}
}

func (sa *SecurityAuditor) checkInformationDisclosure(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`fmt\.Printf.*password`:     "Password might be logged",
		`fmt\.Printf.*secret`:       "Secret might be logged",
		`fmt\.Printf.*token`:        "Token might be logged",
		`log\..*password`:           "Password might be logged",
		`log\..*secret`:             "Secret might be logged",
		`log\..*token`:              "Token might be logged",
		`EnableStackTrace.*true`:    "Stack traces enabled (information disclosure)",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			sa.addIssue(filePath, lineNum, "Information Disclosure", "MEDIUM", description, 
				"Avoid logging sensitive information and disable stack traces in production")
		}
	}
}

func (sa *SecurityAuditor) checkAuthenticationIssues(filePath string, lineNum int, line string) {
	patterns := map[string]string{
		`jwt\.SigningMethodNone`:           "JWT with no signature verification",
		`jwt\.UnsafeAllowNoneSignatureType`: "Unsafe JWT signature type allowed",
		`session.*secure.*false`:           "Session cookies not marked as secure",
		`session.*httponly.*false`:         "Session cookies not marked as HttpOnly",
		`ExpiresAt.*time\.Now\(\)\.Add\(.*365`: "JWT token expires in more than a year",
	}

	for pattern, description := range patterns {
		if matched, _ := regexp.MatchString(pattern, line); matched {
			sa.addIssue(filePath, lineNum, "Authentication Issue", "HIGH", description, 
				"Use secure authentication practices and proper session management")
		}
	}
}

func (sa *SecurityAuditor) addIssue(filePath string, lineNum int, issueType, severity, description, suggestion string) {
	sa.issues = append(sa.issues, SecurityIssue{
		File:        filePath,
		Line:        lineNum,
		Type:        issueType,
		Severity:    severity,
		Description: description,
		Suggestion:  suggestion,
	})
}

func (sa *SecurityAuditor) printResults() {
	if len(sa.issues) == 0 {
		fmt.Println("✅ No security issues found!")
		return
	}

	// Group issues by severity
	severityGroups := map[string][]SecurityIssue{
		"HIGH":   {},
		"MEDIUM": {},
		"LOW":    {},
	}

	for _, issue := range sa.issues {
		severityGroups[issue.Severity] = append(severityGroups[issue.Severity], issue)
	}

	// Print summary
	fmt.Printf("\n📊 Security Audit Summary:\n")
	fmt.Printf("  Total Issues: %d\n", len(sa.issues))
	fmt.Printf("  High Severity: %d\n", len(severityGroups["HIGH"]))
	fmt.Printf("  Medium Severity: %d\n", len(severityGroups["MEDIUM"]))
	fmt.Printf("  Low Severity: %d\n", len(severityGroups["LOW"]))

	// Print issues by severity
	for _, severity := range []string{"HIGH", "MEDIUM", "LOW"} {
		issues := severityGroups[severity]
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

		fmt.Printf("\n%s %s Severity Issues:\n", emoji, severity)
		for _, issue := range issues {
			fmt.Printf("  📁 %s:%d\n", issue.File, issue.Line)
			fmt.Printf("     Type: %s\n", issue.Type)
			fmt.Printf("     Issue: %s\n", issue.Description)
			fmt.Printf("     Fix: %s\n", issue.Suggestion)
			fmt.Println()
		}
	}
}

func (sa *SecurityAuditor) hasHighSeverityIssues() bool {
	for _, issue := range sa.issues {
		if issue.Severity == "HIGH" {
			return true
		}
	}
	return false
}