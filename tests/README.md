# Testing Guidelines

This directory contains all test files for the AI Backend Generator project, organized into a clear structure for better maintainability and execution.

## Directory Structure

```
tests/
├── unit/           # Unit tests for individual components
├── integration/    # Integration tests for component interactions
├── fixtures/       # Test data and fixtures
└── README.md       # This file
```

## Test Organization

### Unit Tests (`tests/unit/`)

Unit tests focus on testing individual components in isolation. All test files use the `unit` package for consistency. Each test file corresponds to a specific module or service:

- `history_handler_test.go` - Tests for HTTP handlers
- `project_history_test.go` - Tests for ProjectHistory model
- `project_history_requests_test.go` - Tests for request/response models
- `project_history_service_test.go` - Tests for ProjectHistoryService
- `database_maintenance_service_test.go` - Tests for DatabaseMaintenanceService
- `file_service_test.go` - Tests for FileService
- `project_history_performance_test.go` - Performance benchmarks
- `test_utils.go` - Shared test utilities and helper functions

### Integration Tests (`tests/integration/`)

Integration tests verify that different components work together correctly. These tests may involve:
- Database interactions
- API endpoint testing
- Service layer integration
- File system operations

### Fixtures (`tests/fixtures/`)

Test fixtures contain reusable test data, mock objects, and helper utilities that can be shared across different test files.

## Running Tests

### Run All Tests
```bash
go test ./tests/...
```

### Run Unit Tests Only
```bash
go test ./tests/unit/...
```

### Run Integration Tests Only
```bash
go test ./tests/integration/...
```

### Run Specific Test File
```bash
go test ./tests/unit/project_history_test.go ./tests/unit/test_utils.go
```

### Run with Coverage
```bash
go test -cover ./tests/...
```

### Run Performance Benchmarks
```bash
go test -bench=. ./tests/unit/project_history_performance_test.go ./tests/unit/test_utils.go
```

### Run Tests in Verbose Mode
```bash
go test -v ./tests/...
```

## Test Naming Conventions

### Test Functions
- Use descriptive names that explain what is being tested
- Format: `TestComponentName_MethodName_Scenario`
- Example: `TestProjectHistoryService_GetUserHistory_WithFilters`

### Benchmark Functions
- Format: `BenchmarkComponentName_MethodName`
- Example: `BenchmarkProjectHistoryService_GetUserHistory`

### Test Files
- Format: `component_name_test.go`
- Example: `project_history_service_test.go`

## Test Structure Best Practices

### Setup and Teardown
- Use `setupTestDB()` functions for database initialization
- Clean up resources in `defer` statements or test teardown
- Use in-memory databases (SQLite) for unit tests

### Test Data
- Create helper functions for generating test data
- Use meaningful test data that reflects real-world scenarios
- Avoid hardcoded values where possible

### Assertions
- Use testify/assert for readable assertions
- Group related assertions together
- Provide meaningful error messages

### Test Cases
- Use table-driven tests for multiple scenarios
- Test both success and error cases
- Include edge cases and boundary conditions

## Example Test Structure

```go
func TestProjectHistoryService_GetUserHistory(t *testing.T) {
    // Setup
    db, service := setupTestDB(t)
    user := createTestUser(t, db)
    
    // Create test data
    createTestProjects(t, db, user.ID, 3)
    
    // Test cases
    tests := []struct {
        name     string
        filters  models.HistoryFilters
        expected int
    }{
        {
            name:     "get all projects",
            filters:  models.HistoryFilters{Page: 1, PageSize: 10},
            expected: 3,
        },
        // More test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Execute
            result, err := service.GetUserHistory(user.ID, tt.filters)
            
            // Assert
            assert.NoError(t, err)
            assert.Len(t, result.Projects, tt.expected)
        })
    }
}
```

## Mock and Stub Guidelines

- Use interfaces to enable easy mocking
- Create mocks for external dependencies
- Keep mocks simple and focused
- Verify mock interactions when necessary

## Performance Testing

- Use benchmarks for performance-critical code
- Test with realistic data volumes
- Monitor memory allocations
- Set reasonable performance thresholds

## Continuous Integration

Tests are automatically run on:
- Pull requests
- Main branch commits
- Release builds

Ensure all tests pass before submitting code changes.

## Troubleshooting

### Common Issues

1. **Database Connection Errors**: Ensure test database is properly initialized
2. **File Path Issues**: Use `t.TempDir()` for temporary file operations
3. **Race Conditions**: Use proper synchronization in concurrent tests
4. **Memory Leaks**: Clean up resources in test teardown

### Debugging Tests

```bash
# Run specific test with verbose output
go test -v -run TestSpecificFunction ./tests/unit/

# Run with race detection
go test -race ./tests/...

# Generate test coverage report
go test -coverprofile=coverage.out ./tests/...
go tool cover -html=coverage.out
```

## Contributing

When adding new tests:

1. Follow the established naming conventions
2. Place tests in the appropriate directory
3. Include both positive and negative test cases
4. Add performance benchmarks for critical paths
5. Update this README if adding new test categories

For questions about testing practices, refer to the project's contribution guidelines or reach out to the development team.