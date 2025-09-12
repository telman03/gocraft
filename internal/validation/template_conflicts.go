package validation

import (
	"fmt"
	"strings"
)

// ConflictError represents a template conflict with suggestions
type ConflictError struct {
	Message     string   `json:"message"`
	Conflicts   []string `json:"conflicts"`
	Suggestions []string `json:"suggestions"`
}

func (e *ConflictError) Error() string {
	return e.Message
}

// ValidationResult contains the validation outcome
type ValidationResult struct {
	IsValid          bool            `json:"is_valid"`
	AdjustedFeatures []string        `json:"adjusted_features"`
	Errors           []*ConflictError `json:"errors,omitempty"`
	Warnings         []string        `json:"warnings,omitempty"`
	AddedDependencies []string       `json:"added_dependencies,omitempty"`
}

// TemplateValidator validates feature combinations and resolves conflicts
type TemplateValidator struct {
	// Database categories
	relationalDBs []string
	nosqlDBs      []string
	cacheDBs      []string
	
	// Framework categories
	webFrameworks []string
	
	// ORM categories
	orms []string
	
	// Auth categories
	authMethods []string
	
	// AI integrations
	aiProviders []string
}

// NewTemplateValidator creates a new validator with predefined rules
func NewTemplateValidator() *TemplateValidator {
	return &TemplateValidator{
		relationalDBs: []string{"mysql", "postgresql", "postgres", "sqlite"},
		nosqlDBs:      []string{"mongodb", "mongo"},
		cacheDBs:      []string{"redis", "badger"},
		webFrameworks: []string{"gin", "echo", "fiber"},
		orms:          []string{"gorm", "sqlc"},
		authMethods:   []string{"auth", "oauth2"},
		aiProviders:   []string{"openai", "openrouter", "claude"},
	}
}

// ValidateFeatures validates a list of features against conflict rules
func (v *TemplateValidator) ValidateFeatures(features []string) *ValidationResult {
	result := &ValidationResult{
		IsValid:          true,
		AdjustedFeatures: make([]string, 0),
		Errors:           make([]*ConflictError, 0),
		Warnings:         make([]string, 0),
		AddedDependencies: make([]string, 0),
	}
	
	// Normalize features (lowercase, handle aliases)
	normalizedFeatures := v.normalizeFeatures(features)
	
	// Check database conflicts
	v.validateDatabases(normalizedFeatures, result)
	
	// Check ORM conflicts
	v.validateORMs(normalizedFeatures, result)
	
	// Check framework conflicts
	v.validateFrameworks(normalizedFeatures, result)
	
	// Check auth conflicts
	v.validateAuth(normalizedFeatures, result)
	
	// Check gRPC dependencies
	v.validateGRPC(normalizedFeatures, result)
	
	// Check Swagger dependencies
	v.validateSwagger(normalizedFeatures, result)
	
	// Add required dependencies
	v.addRequiredDependencies(normalizedFeatures, result)
	
	// Set final feature list if valid
	if len(result.Errors) == 0 {
		result.AdjustedFeatures = v.removeDuplicates(append(normalizedFeatures, result.AddedDependencies...))
	} else {
		result.IsValid = false
	}
	
	return result
}

// normalizeFeatures handles aliases and case normalization
func (v *TemplateValidator) normalizeFeatures(features []string) []string {
	normalized := make([]string, 0)
	aliases := map[string]string{
		"postgres":     "postgresql",
		"mongo":        "mongodb",
		"jwt":          "auth",
		"authentication": "auth",
		"websockets":   "websocket",
		"ws":           "websocket",
		"docker":       "dockerfile",
		"env-config":   "env",
		"environment":  "env",
		"logging":      "logger",
		"log":          "logger",
	}
	
	for _, feature := range features {
		feature = strings.ToLower(strings.TrimSpace(feature))
		if feature == "" {
			continue
		}
		
		// Handle aliases
		if alias, exists := aliases[feature]; exists {
			feature = alias
		}
		
		normalized = append(normalized, feature)
	}
	
	return v.removeDuplicates(normalized)
}

// validateDatabases checks for database conflicts
func (v *TemplateValidator) validateDatabases(features []string, result *ValidationResult) {
	selectedRelational := v.filterFeatures(features, v.relationalDBs)
	selectedNoSQL := v.filterFeatures(features, v.nosqlDBs)
	_ = v.filterFeatures(features, v.cacheDBs) // Cache DBs don't have conflicts
	
	// Rule: Only one relational database allowed
	if len(selectedRelational) > 1 {
		result.Errors = append(result.Errors, &ConflictError{
			Message:   fmt.Sprintf("Multiple relational databases selected: %s", strings.Join(selectedRelational, ", ")),
			Conflicts: selectedRelational,
			Suggestions: []string{
				"Choose only one relational database (MySQL, PostgreSQL, or SQLite)",
				"MongoDB and Redis can be used alongside a relational database",
			},
		})
	}
	
	// Warning: No database selected
	if len(selectedRelational) == 0 && len(selectedNoSQL) == 0 {
		result.Warnings = append(result.Warnings, "No database selected. Consider adding a database for data persistence.")
	}
	
	// Info: Multiple database types (allowed)
	if len(selectedRelational) > 0 && len(selectedNoSQL) > 0 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Using multiple database types: %s. Ensure your application handles multiple connections properly.", 
			strings.Join(append(selectedRelational, selectedNoSQL...), ", ")))
	}
}

// validateORMs checks for ORM conflicts
func (v *TemplateValidator) validateORMs(features []string, result *ValidationResult) {
	selectedORMs := v.filterFeatures(features, v.orms)
	
	// Rule: Only one ORM allowed
	if len(selectedORMs) > 1 {
		result.Errors = append(result.Errors, &ConflictError{
			Message:   fmt.Sprintf("Multiple ORMs selected: %s", strings.Join(selectedORMs, ", ")),
			Conflicts: selectedORMs,
			Suggestions: []string{
				"Choose either GORM or SQLC, not both",
				"GORM: Full-featured ORM with migrations and relationships",
				"SQLC: Type-safe SQL code generation from raw SQL queries",
			},
		})
	}
	
	// Check ORM compatibility with databases
	if v.containsFeature(features, "gorm") {
		relationalDBs := v.filterFeatures(features, v.relationalDBs)
		if len(relationalDBs) == 0 {
			result.Warnings = append(result.Warnings, "GORM selected but no relational database found. GORM works best with PostgreSQL, MySQL, or SQLite.")
		}
		
		if v.containsFeature(features, "mongodb") {
			result.Warnings = append(result.Warnings, "GORM is not compatible with MongoDB. Consider using the official MongoDB Go driver instead.")
		}
	}
	
	if v.containsFeature(features, "sqlc") {
		if !v.containsFeature(features, "postgresql") && !v.containsFeature(features, "mysql") {
			result.Warnings = append(result.Warnings, "SQLC works best with PostgreSQL or MySQL for optimal type safety.")
		}
	}
}

// validateFrameworks checks for web framework conflicts
func (v *TemplateValidator) validateFrameworks(features []string, result *ValidationResult) {
	selectedFrameworks := v.filterFeatures(features, v.webFrameworks)
	
	// Rule: Only one web framework allowed
	if len(selectedFrameworks) > 1 {
		result.Errors = append(result.Errors, &ConflictError{
			Message:   fmt.Sprintf("Multiple web frameworks selected: %s", strings.Join(selectedFrameworks, ", ")),
			Conflicts: selectedFrameworks,
			Suggestions: []string{
				"Choose only one web framework",
				"Gin: Fast, minimalist framework with good performance",
				"Echo: High performance, extensible, minimalist framework",
				"Fiber: Express-inspired framework built on Fasthttp",
			},
		})
	}
	
	// Default framework if none selected
	if len(selectedFrameworks) == 0 {
		result.AddedDependencies = append(result.AddedDependencies, "gin")
		result.Warnings = append(result.Warnings, "No web framework selected. Adding Gin as default framework.")
	}
}

// validateAuth checks for authentication conflicts
func (v *TemplateValidator) validateAuth(features []string, result *ValidationResult) {
	hasAuth := v.containsFeature(features, "auth")
	hasOAuth2 := v.containsFeature(features, "oauth2")
	
	// Rule: Both auth methods can coexist but warn about complexity
	if hasAuth && hasOAuth2 {
		result.Warnings = append(result.Warnings, "Both JWT and OAuth2 authentication selected. Ensure they are configured as complementary (e.g., JWT for API, OAuth2 for external providers).")
	}
}

// validateGRPC checks gRPC dependencies
func (v *TemplateValidator) validateGRPC(features []string, result *ValidationResult) {
	hasGRPC := v.containsFeature(features, "grpc")
	hasProto := v.containsFeature(features, "proto")
	
	// Rule: gRPC requires proto
	if hasGRPC && !hasProto {
		result.AddedDependencies = append(result.AddedDependencies, "proto")
		result.Warnings = append(result.Warnings, "gRPC selected. Adding Protocol Buffers (proto) as required dependency.")
	}
	
	// Rule: proto without gRPC is unusual but allowed
	if hasProto && !hasGRPC {
		result.Warnings = append(result.Warnings, "Protocol Buffers selected without gRPC. This is valid if you're using protobuf for serialization only.")
	}
}

// validateSwagger checks Swagger dependencies
func (v *TemplateValidator) validateSwagger(features []string, result *ValidationResult) {
	hasSwagger := v.containsFeature(features, "swagger")
	selectedFrameworks := v.filterFeatures(features, v.webFrameworks)
	
	// Rule: Swagger requires a web framework
	if hasSwagger && len(selectedFrameworks) == 0 {
		result.Warnings = append(result.Warnings, "Swagger documentation selected but no web framework found. Swagger requires a REST API framework.")
	}
}

// addRequiredDependencies adds always-required templates
func (v *TemplateValidator) addRequiredDependencies(features []string, result *ValidationResult) {
	// Always include core templates
	coreTemplates := []string{"env", "gitignore", "main"}
	
	for _, template := range coreTemplates {
		if !v.containsFeature(features, template) && !v.containsFeature(result.AddedDependencies, template) {
			result.AddedDependencies = append(result.AddedDependencies, template)
		}
	}
	
	// Add migrations if using GORM with relational DB
	if v.containsFeature(features, "gorm") {
		relationalDBs := v.filterFeatures(features, v.relationalDBs)
		if len(relationalDBs) > 0 && !v.containsFeature(features, "migrations") {
			result.AddedDependencies = append(result.AddedDependencies, "migrations")
		}
	}
	
	// Add middleware if using auth
	if (v.containsFeature(features, "auth") || v.containsFeature(features, "oauth2")) && !v.containsFeature(features, "middleware") {
		result.AddedDependencies = append(result.AddedDependencies, "middleware")
	}
	
	// Add config if using multiple services
	hasMultipleServices := len(v.filterFeatures(features, append(v.relationalDBs, v.nosqlDBs...))) > 1 ||
		len(v.filterFeatures(features, v.aiProviders)) > 0 ||
		v.containsFeature(features, "grpc") ||
		v.containsFeature(features, "websocket")
	
	if hasMultipleServices && !v.containsFeature(features, "config") {
		result.AddedDependencies = append(result.AddedDependencies, "config")
	}
}

// Helper functions
func (v *TemplateValidator) filterFeatures(features []string, category []string) []string {
	var result []string
	for _, feature := range features {
		for _, cat := range category {
			if feature == cat {
				result = append(result, feature)
				break
			}
		}
	}
	return result
}

func (v *TemplateValidator) containsFeature(features []string, target string) bool {
	for _, feature := range features {
		if feature == target {
			return true
		}
	}
	return false
}

func (v *TemplateValidator) removeDuplicates(features []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, feature := range features {
		if !seen[feature] {
			seen[feature] = true
			result = append(result, feature)
		}
	}
	
	return result
}

// GetSupportedFeatures returns all supported features organized by category
func (v *TemplateValidator) GetSupportedFeatures() map[string][]string {
	return map[string][]string{
		"databases": append(append(v.relationalDBs, v.nosqlDBs...), v.cacheDBs...),
		"frameworks": v.webFrameworks,
		"orms": v.orms,
		"auth": v.authMethods,
		"ai": v.aiProviders,
		"communication": []string{"grpc", "websocket", "proto"},
		"devops": []string{"dockerfile", "docker-compose", "makefile"},
		"documentation": []string{"swagger", "postman", "readme"},
		"utilities": []string{"logger", "middleware", "config", "migrations"},
		"core": []string{"env", "gitignore", "main"},
	}
}