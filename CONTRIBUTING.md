# Contributing to GoCraft

Thank you for your interest in contributing to GoCraft! We welcome contributions from the community and are excited to work with you.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Docker (optional, for containerized development)

### Development Setup

1. **Fork the repository**
   ```bash
   # Fork the repo on GitHub, then clone your fork
   git clone https://github.com/YOUR_USERNAME/gocraft.git
   cd gocraft
   ```

2. **Set up the development environment**
   ```bash
   # Install dependencies
   go mod download
   
   # Copy environment configuration
   cp .env.example .env
   
   # Edit .env with your local configuration
   ```

3. **Build and run the project**
   ```bash
   # Build the project
   go build -o bin/gocraft cmd/gocraft/main.go
   
   # Run the application
   ./bin/gocraft
   
   # Or use the Makefile
   make build
   make run
   ```

4. **Run tests**
   ```bash
   # Run all tests
   go test ./...
   
   # Run tests with coverage
   go test -cover ./...
   
   # Or use the Makefile
   make test
   ```

## How to Contribute

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When creating a bug report, include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Go version and operating system
- Any relevant logs or error messages

### Suggesting Features

Feature requests are welcome! Please:

- Check existing issues and discussions first
- Provide a clear description of the feature
- Explain the use case and benefits
- Consider the scope and complexity

### Contributing Code

1. **Choose an issue** - Look for issues labeled `good first issue` or `help wanted`
2. **Create a branch** - Use a descriptive name like `feature/add-oauth-support` or `fix/template-validation`
3. **Make your changes** - Follow our coding standards
4. **Write tests** - Ensure your changes are well-tested
5. **Update documentation** - Update relevant docs and comments
6. **Submit a pull request** - Follow our PR template

## Pull Request Process

1. **Before submitting:**
   - Ensure your code follows our style guidelines
   - Run all tests and ensure they pass
   - Update documentation as needed
   - Rebase your branch on the latest main branch

2. **PR Requirements:**
   - Fill out the PR template completely
   - Reference any related issues
   - Include screenshots for UI changes
   - Ensure CI checks pass

3. **Review process:**
   - At least one maintainer review is required
   - Address any feedback promptly
   - Keep your branch up to date with main

4. **After approval:**
   - Maintainers will merge your PR
   - Your branch will be deleted automatically

## Coding Standards

### Go Style Guide

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Use `golint` and `go vet` to check for issues
- Follow Go naming conventions

### Code Organization

- Keep functions small and focused
- Use meaningful variable and function names
- Add comments for complex logic
- Organize imports: standard library, third-party, local packages

### Error Handling

- Always handle errors appropriately
- Use custom error types when beneficial
- Provide meaningful error messages
- Log errors at appropriate levels

### Example Code Style

```go
// Package comment
package handlers

import (
    "context"
    "fmt"
    "log"
    
    "github.com/gin-gonic/gin"
    
    "github.com/gocraft/internal/models"
    "github.com/gocraft/internal/services"
)

// HandlerFunc represents a handler function with proper error handling
func (h *Handler) CreateProject(c *gin.Context) {
    var req models.CreateProjectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("failed to bind request", "error", err)
        c.JSON(400, gin.H{"error": "invalid request format"})
        return
    }
    
    project, err := h.projectService.Create(c.Request.Context(), &req)
    if err != nil {
        h.logger.Error("failed to create project", "error", err)
        c.JSON(500, gin.H{"error": "internal server error"})
        return
    }
    
    c.JSON(201, project)
}
```

## Testing Guidelines

### Test Structure

- Use table-driven tests when appropriate
- Test both success and error cases
- Use meaningful test names that describe the scenario
- Keep tests focused and independent

### Test Categories

1. **Unit Tests** - Test individual functions and methods
2. **Integration Tests** - Test component interactions
3. **End-to-End Tests** - Test complete workflows

### Example Test

```go
func TestProjectService_Create(t *testing.T) {
    tests := []struct {
        name    string
        request *models.CreateProjectRequest
        want    *models.Project
        wantErr bool
    }{
        {
            name: "valid project creation",
            request: &models.CreateProjectRequest{
                Name:      "test-project",
                Framework: "gin",
            },
            want: &models.Project{
                Name:      "test-project",
                Framework: "gin",
            },
            wantErr: false,
        },
        // Add more test cases...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Documentation

### Code Documentation

- Add package comments for all packages
- Document exported functions and types
- Use examples in documentation when helpful
- Keep comments up to date with code changes

### Project Documentation

- Update README.md for user-facing changes
- Update API documentation for endpoint changes
- Add examples for new features
- Update architecture docs for significant changes

## Community

### Getting Help

- Check the [documentation](docs/)
- Search existing [issues](https://github.com/gocraft/gocraft/issues)
- Join our discussions
- Ask questions in issues with the `question` label

### Communication Guidelines

- Be respectful and inclusive
- Provide context and details
- Be patient with responses
- Help others when you can

## Development Workflow

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
type(scope): description

feat(auth): add OAuth2 authentication support
fix(templates): resolve template parsing error
docs(api): update endpoint documentation
test(handlers): add unit tests for project creation
```

### Release Process

- Releases follow semantic versioning (SemVer)
- Changelog is maintained automatically
- Release notes highlight major changes
- Breaking changes are clearly documented

## Recognition

Contributors are recognized in:
- Release notes for significant contributions
- README.md contributors section
- Special recognition for first-time contributors

Thank you for contributing to GoCraft! 🚀