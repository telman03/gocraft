# 🐛 GitHub Issues for Contributors

Copy and paste these issues into your GitHub repository to help contributors get started.

---

## Issue #1: Add MongoDB Template Support

**Labels:** `enhancement`, `good first issue`, `templates`

**Title:** Add MongoDB template support for NoSQL projects

**Description:**
Currently, GoCraft supports PostgreSQL, MySQL, and SQLite, but lacks MongoDB support for NoSQL applications. This feature would enable developers to generate projects with MongoDB integration.

**Acceptance Criteria:**
- [ ] Create MongoDB connection template in `internal/templates/databases/mongodb/`
- [ ] Add MongoDB driver configuration
- [ ] Include environment variables for MongoDB connection
- [ ] Add MongoDB to feature validation system
- [ ] Create example models with MongoDB tags
- [ ] Add conflict rules (MongoDB cannot be combined with SQL databases)
- [ ] Update documentation with MongoDB examples

**Technical Requirements:**
- Use `go.mongodb.org/mongo-driver/mongo` as the official driver
- Support connection string and individual parameter configuration
- Include connection pooling and timeout settings
- Add health check for MongoDB connection

**Files to Modify:**
- `internal/templates/databases/mongodb/` (new directory)
- `internal/validation/conflicts.go`
- `internal/builder/features.go`
- `docs/FEATURES.md`

**Example Environment Variables:**
```env
MONGODB_URL=mongodb://localhost:27017
MONGODB_DATABASE=myapp
MONGODB_USERNAME=
MONGODB_PASSWORD=
```

---

## Issue #2: Improve Error Messages in Feature Validation

**Labels:** `enhancement`, `user-experience`, `validation`

**Title:** Make feature validation error messages more user-friendly

**Description:**
Current error messages for feature conflicts are technical and not very helpful for new users. We need to improve them with clear explanations and suggestions.

**Current Behavior:**
```json
{
  "error": "Multiple relational databases selected: mysql, postgresql"
}
```

**Desired Behavior:**
```json
{
  "error": "Cannot use multiple primary databases in one project",
  "conflicts": ["mysql", "postgresql"],
  "suggestions": [
    "Choose one primary database: MySQL, PostgreSQL, or SQLite",
    "You can add Redis as a cache alongside any primary database",
    "Consider MongoDB if you need a NoSQL solution"
  ],
  "learnMore": "https://docs.gocraft.dev/features/databases"
}
```

**Acceptance Criteria:**
- [ ] Update error response format to include suggestions
- [ ] Add helpful explanations for each conflict type
- [ ] Include links to documentation
- [ ] Add examples of valid combinations
- [ ] Update API documentation with new error format
- [ ] Add unit tests for new error messages

**Files to Modify:**
- `internal/validation/validator.go`
- `internal/validation/conflicts.go`
- `internal/handlers/validate.go`
- `docs/api/swagger.yaml`

---

## Issue #3: Add CLI Tool for Offline Project Generation

**Labels:** `enhancement`, `cli`, `major-feature`

**Title:** Create command-line interface for GoCraft

**Description:**
Many developers prefer CLI tools for project generation. A CLI would allow offline usage and integration with development workflows.

**Proposed Usage:**
```bash
# Install CLI
go install github.com/telman03/gocraft/cmd/gocraft-cli

# Generate project
gocraft-cli generate \
  --name my-api \
  --framework gin \
  --features postgresql,auth,redis \
  --output ./my-api

# List available features
gocraft-cli features

# Validate feature combination
gocraft-cli validate --features gin,postgresql,mongodb
```

**Acceptance Criteria:**
- [ ] Create CLI application in `cmd/gocraft-cli/`
- [ ] Implement `generate` command with feature selection
- [ ] Add `features` command to list available options
- [ ] Add `validate` command for feature validation
- [ ] Include offline template bundling
- [ ] Add progress indicators for generation
- [ ] Support configuration files (`.gocraft.yaml`)
- [ ] Add shell completion (bash, zsh, fish)

**Technical Requirements:**
- Use `github.com/spf13/cobra` for CLI framework
- Bundle templates in the binary for offline use
- Support both flags and interactive prompts
- Include proper error handling and user feedback

---

## Issue #4: Create Integration Tests for Project Generation

**Labels:** `testing`, `good first issue`, `quality`

**Title:** Add comprehensive integration tests for generated projects

**Description:**
We need integration tests that verify generated projects actually compile and run correctly with different feature combinations.

**Test Scenarios:**
- [ ] Generate project with Gin + PostgreSQL + Auth
- [ ] Generate project with Fiber + MongoDB + Redis
- [ ] Generate project with Echo + SQLite + Swagger
- [ ] Verify all generated projects compile successfully
- [ ] Test that Docker builds work for generated projects
- [ ] Validate environment variable loading
- [ ] Test database connections in generated projects

**Acceptance Criteria:**
- [ ] Create `tests/integration/` directory structure
- [ ] Add test cases for popular feature combinations
- [ ] Implement project compilation verification
- [ ] Add Docker build testing
- [ ] Create test database setup/teardown
- [ ] Add CI/CD integration for integration tests
- [ ] Document test execution process

**Files to Create:**
- `tests/integration/generation_test.go`
- `tests/integration/docker_test.go`
- `tests/integration/helpers.go`
- `tests/fixtures/` (test data)

---

## Issue #5: Add Swagger Documentation Generation

**Labels:** `enhancement`, `documentation`, `good first issue`

**Title:** Auto-generate Swagger/OpenAPI documentation for generated projects

**Description:**
Generated projects should include automatic API documentation using Swagger/OpenAPI specifications.

**Features to Implement:**
- [ ] Add Swagger annotations to generated handlers
- [ ] Include Swagger UI in generated projects
- [ ] Generate OpenAPI 3.0 specification
- [ ] Add documentation middleware
- [ ] Include example requests/responses
- [ ] Support authentication documentation

**Acceptance Criteria:**
- [ ] Create Swagger template files
- [ ] Add `swagger` feature to feature list
- [ ] Generate handler annotations automatically
- [ ] Include Swagger UI route (`/swagger/`)
- [ ] Add environment configuration for Swagger
- [ ] Update project README with documentation link
- [ ] Add examples for different HTTP methods

**Dependencies:**
- `github.com/swaggo/swag`
- Framework-specific Swagger middleware (gin-swagger, echo-swagger, etc.)

**Example Generated Code:**
```go
// GetUser godoc
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 404 {object} ErrorResponse
// @Router /users/{id} [get]
func GetUser(c *gin.Context) {
    // Implementation
}
```

---

## Issue #6: Implement Template Caching for Better Performance

**Labels:** `performance`, `enhancement`, `caching`

**Title:** Add template caching to improve project generation speed

**Description:**
Template parsing is currently done on every request, which impacts performance. Implementing caching would significantly improve generation speed.

**Current Performance:**
- Template parsing: ~200ms per request
- Total generation time: ~2-3 seconds

**Target Performance:**
- Template parsing: ~5ms per request (cached)
- Total generation time: ~500ms-1s

**Implementation Plan:**
- [ ] Add in-memory template cache
- [ ] Implement cache invalidation on template updates
- [ ] Add cache warming on application startup
- [ ] Include cache statistics in admin dashboard
- [ ] Add cache configuration options
- [ ] Implement cache size limits and LRU eviction

**Acceptance Criteria:**
- [ ] Reduce template parsing time by 90%+
- [ ] Add cache hit/miss metrics
- [ ] Implement graceful cache invalidation
- [ ] Add cache configuration to environment variables
- [ ] Include cache status in health checks
- [ ] Add unit tests for caching logic

**Files to Modify:**
- `internal/builder/template_cache.go` (new)
- `internal/builder/builder.go`
- `internal/config/config.go`
- `internal/handlers/admin_handler.go`

---

## Issue #7: Add Rate Limiting Documentation and Examples

**Labels:** `documentation`, `good first issue`, `security`

**Title:** Document rate limiting implementation and provide usage examples

**Description:**
The current rate limiting implementation lacks proper documentation and examples for developers to understand and customize it.

**Missing Documentation:**
- How rate limiting works in generated projects
- Configuration options and environment variables
- Examples of different rate limiting strategies
- Integration with Redis for distributed rate limiting
- Custom rate limiting rules

**Acceptance Criteria:**
- [ ] Create rate limiting documentation in `docs/`
- [ ] Add configuration examples
- [ ] Document Redis integration for distributed systems
- [ ] Include code examples for custom rate limiters
- [ ] Add troubleshooting guide
- [ ] Update API documentation with rate limit headers

**Documentation Sections:**
1. **Overview** - What is rate limiting and why use it
2. **Configuration** - Environment variables and options
3. **Strategies** - Different rate limiting approaches
4. **Redis Integration** - Distributed rate limiting
5. **Custom Rules** - Per-endpoint rate limits
6. **Monitoring** - Metrics and alerting
7. **Troubleshooting** - Common issues and solutions

---

## Issue #8: Create Docker Compose Templates for Development

**Labels:** `enhancement`, `docker`, `development`

**Title:** Generate Docker Compose files for local development environments

**Description:**
Generated projects should include Docker Compose configurations for easy local development setup with all required services.

**Services to Include:**
- [ ] Application container
- [ ] Database (PostgreSQL, MySQL, MongoDB)
- [ ] Redis (if selected)
- [ ] Nginx (reverse proxy)
- [ ] Adminer/pgAdmin (database management)

**Features:**
- [ ] Environment-specific configurations (dev, staging, prod)
- [ ] Volume mounting for development
- [ ] Health checks for all services
- [ ] Network configuration
- [ ] Service dependencies and startup order

**Acceptance Criteria:**
- [ ] Generate `docker-compose.yml` for development
- [ ] Include `docker-compose.prod.yml` for production
- [ ] Add database initialization scripts
- [ ] Include environment variable templates
- [ ] Add service health checks
- [ ] Document usage in project README

**Example Structure:**
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - postgres
      - redis
    environment:
      - DATABASE_URL=postgres://user:pass@postgres:5432/db
  
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: myapp
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
```

---

## Issue #9: Add WebSocket Support Template

**Labels:** `enhancement`, `websocket`, `real-time`

**Title:** Create WebSocket template for real-time applications

**Description:**
Add WebSocket support template for applications requiring real-time communication (chat apps, live updates, notifications).

**Features to Implement:**
- [ ] WebSocket connection handling
- [ ] Message broadcasting
- [ ] Room/channel management
- [ ] Connection authentication
- [ ] Heartbeat/ping-pong mechanism
- [ ] Graceful connection cleanup

**Acceptance Criteria:**
- [ ] Create WebSocket handler templates
- [ ] Add connection management utilities
- [ ] Include message types and routing
- [ ] Add authentication middleware for WebSocket
- [ ] Create example chat implementation
- [ ] Add WebSocket client examples (JavaScript)
- [ ] Include load testing examples

**Dependencies:**
- `github.com/gorilla/websocket`

**Example Usage:**
```go
// WebSocket endpoint
func HandleWebSocket(c *gin.Context) {
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    defer conn.Close()
    
    // Handle messages
    for {
        var msg Message
        err := conn.ReadJSON(&msg)
        if err != nil {
            break
        }
        
        // Process and broadcast message
        hub.Broadcast(msg)
    }
}
```

---

## Issue #10: Implement Project Template Versioning

**Labels:** `enhancement`, `templates`, `versioning`

**Title:** Add versioning system for project templates

**Description:**
Implement a versioning system for templates to ensure compatibility and allow users to choose specific template versions.

**Requirements:**
- [ ] Template version metadata
- [ ] Backward compatibility checking
- [ ] Version selection in API
- [ ] Migration guides between versions
- [ ] Deprecation warnings for old templates

**Acceptance Criteria:**
- [ ] Add version field to template metadata
- [ ] Implement version validation
- [ ] Add version selection to generation API
- [ ] Create template migration system
- [ ] Add version information to generated projects
- [ ] Include changelog for template versions

**API Changes:**
```json
{
  "projectName": "my-api",
  "framework": "gin",
  "features": ["postgresql", "auth"],
  "templateVersion": "1.2.0"
}
```

**Template Metadata:**
```yaml
# template.yaml
version: "1.2.0"
compatibility:
  go: ">=1.21"
  features:
    postgresql: ">=1.0.0"
    auth: ">=2.1.0"
changelog:
  - "Added JWT refresh token support"
  - "Improved error handling"
```

---

These issues provide a good mix of difficulty levels and different areas of the codebase, making them perfect for attracting contributors with various skill levels and interests.