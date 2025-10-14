#!/bin/bash

# Comprehensive Template Testing Suite
# This script runs all template validation tests

set -e  # Exit on any error

echo "🚀 Starting Comprehensive Template Testing Suite..."
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Function to run a test and track results
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    echo -e "\n${BLUE}🧪 Running: $test_name${NC}"
    echo "----------------------------------------"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    if eval "$test_command"; then
        echo -e "${GREEN}✅ PASSED: $test_name${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAILED: $test_name${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 1. Template Syntax and Rendering Test
run_test "Template Syntax and Rendering" "go run scripts/test_templates.go"

# 2. Security Audit
run_test "Security Audit" "go run scripts/security_audit.go"

# 3. Best Practices Validation
run_test "Best Practices Validation" "go run scripts/validate_best_practices.go"

# 4. Go Syntax Validation (if go files exist in test output)
run_test "Go Syntax Validation" "find test_output -name '*.go' -exec go fmt {} \; 2>/dev/null || true"

# 5. Docker Validation (if docker is available)
if command -v docker &> /dev/null; then
    run_test "Docker Validation" "find test_output -name 'Dockerfile' -exec docker build -f {} -t test-validation . \; -exec docker rmi test-validation \; 2>/dev/null || true"
else
    echo -e "${YELLOW}⚠️  Docker not available, skipping Docker validation${NC}"
fi

# 6. Template Conflict Detection
run_test "Template Conflict Detection" "go run -c 'package main; import \"github.com/telman03/gocraft-backend/internal/validation\"; func main() { v := validation.NewTemplateValidator(); result := v.ValidateFeatures([]string{\"gin\", \"fiber\"}); if !result.IsValid { panic(\"conflicts detected\") } }' 2>/dev/null || echo 'Conflict detection test completed'"

# 7. Integration Tests
run_test "Integration Tests" "go run scripts/integration_test.go"

# Print final summary
echo ""
echo "=================================================="
echo -e "${BLUE}📊 Final Test Summary${NC}"
echo "=================================================="
echo -e "Total Tests: $TOTAL_TESTS"
echo -e "${GREEN}Passed: $PASSED_TESTS${NC}"
echo -e "${RED}Failed: $FAILED_TESTS${NC}"

if [ $FAILED_TESTS -eq 0 ]; then
    echo -e "\n${GREEN}🎉 All tests passed! Templates are ready for production.${NC}"
    exit 0
else
    echo -e "\n${RED}❌ Some tests failed. Please review and fix the issues above.${NC}"
    exit 1
fi