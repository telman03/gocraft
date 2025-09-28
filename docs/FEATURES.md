# ✨ GoCraft Features Documentation

## Overview

GoCraft supports a wide range of features that can be combined to create production-ready Go microservices. This document provides comprehensive information about all available features, their configurations, and compatibility.

## 🏗️ Web Frameworks

### Gin Framework
**Feature ID:** `gin`

**Description:** Fast and minimalist web framework with a martini-like API.

**Generated Components:**
- HTTP router setup
- Middleware integration
- JSON binding and validation
- Error handling
- CORS configuration

**Template Files:**
- `cmd/[project]/main.go` - Server setup with Gin
- `internal/router/router.go` - Route definitions
- `internal/middleware/` - Gin-specific middleware

**Environment Variables:**
```env
GIN_MODE=release
PORT=8080
```

**Example Usage:**
```go
r := gin.Default()
r.Use(middleware.CORS())
r.GET("/health", handlers.Health)
```

---

### Echo Framework
**Feature ID:** `echo`

**Description:** High performance, extensible, minimalist web framework.

**Generated Components:**
- Echo server configuration
- Middleware stack
- Request/response handling
- Validation integration
- Error handling

**Template Files:**
- `cmd/[project]/main.go` - Echo server setup
- `internal/router/routes.go` - Route registration
- `internal/middleware/` - Echo middleware

**Environment Variables:**
```env
ECHO_DEBUG=false
PORT=8080
```

**Example Usage:**
```go
e := echo.New()
e.Use(middleware.Logger())
e.GET("/health", handlers.Health)
```

---

### Fiber Framework
**Feature ID:** `fiber`

**Description:** Express-inspired web framework built on Fasthttp.

**Generated Components:**
- Fiber app configuration
- High-performance routing
- Middleware integration
- JSON handling
- Static file serving

**Template Files:**
- `cmd/[project]/main.go` - Fiber app setup
- `internal/routes/` - Route definitions
- `internal/middleware/` - Fiber middleware

**Environment Variables:**
```env
FIBER_PREFORK=false
PORT=8080
```

**Example Usage:**
```go
app := fiber.New()
app.Use(cors.New())
app.Get("/health", handlers.Health)
```

## 🗄️ Database Support

### PostgreSQL
**Feature ID:** `postgresql`

**Description:** Advanced open-source relational database.

**Generated Components:**
- Database connection setup
- GORM integration
- Migration system
- Connection pooling
- Health checks

**Template Files:**
- `internal/database/postgres.go` - Connection setup
- `internal/models/` - GORM models
- `migrations/` - Database migrations

**Environment Variables:**
```env
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
```

**Dependencies:**
- `gorm.io/gorm`
- `gorm.io/driver/postgres`

---

### MySQL
**Feature ID:** `mysql`

**Description:** Popular open-source relational database.

**Generated Components:**
- MySQL connection setup
- GORM integration
- Migration support
- Connection configuration

**Environment Variables:**
```env
MYSQL_HOST=localhost
MYSQL_PORT=3306
MYSQL_DATABASE=myapp
MYSQL_USER=root
MYSQL_PASSWORD=password
MYSQL_CHARSET=utf8mb4
MYSQL_PARSE_TIME=true
MYSQL_LOC=Local
```

**Dependencies:**
- `gorm.io/driver/mysql`

---

### SQLite
**Feature ID:** `sqlite`

**Description:** Lightweight, embedded SQL database.

**Generated Components:**
- SQLite file database setup
- GORM integration
- Local development configuration

**Environment Variables:**
```env
SQLITE_DATABASE=./data/app.db
SQLITE_CACHE=shared
SQLITE_MODE=rwc
```

**Dependencies:**
- `gorm.io/driver/sqlite`

---

### MongoDB
**Feature ID:** `mongodb`

**Description:** NoSQL document database.

**Generated Components:**
- MongoDB connection setup
- Document models
- Collection operations
- Aggregation pipelines

**Environment Variables:**
```env
MONGODB_URL=mongodb://localhost:27017
MONGODB_HOST=localhost
MONGODB_PORT=27017
MONGODB_DATABASE=myapp
MONGODB_USERNAME=
MONGODB_PASSWORD=
MONGODB_AUTH_SOURCE=admin
```

**Dependencies:**
- `go.mongodb.org/mongo-driver/mongo`

---

### Redis
**Feature ID:** `redis`

**Description:** In-memory data structure store for caching and sessions.

**Generated Components:**
- Redis client setup
- Caching utilities
- Session storage
- Rate limiting support

**Environment Variables:**
```env
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
```

**Dependencies:**
- `github.com/redis/go-redis/v9`

## 🔐 Authentication & Security

### JWT Authentication
**Feature ID:** `auth`

**Description:** JSON Web Token authentication system.

**Generated Components:**
- JWT token generation/validation
- Authentication middleware
- User registration/login
- Password hashing
- Protected routes

**Template Files:**
- `internal/auth/jwt.go` - JWT utilities
- `internal/auth/handlers.go` - Auth handlers
- `internal/middleware/auth.go` - Auth middleware
- `internal/models/user.go` - User model

**Environment Variables:**
```env
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
JWT_EXPIRATION=24h
JWT_ISSUER=myapp
BCRYPT_COST=12
```

**Dependencies:**
- `github.com/golang-jwt/jwt/v5`
- `golang.org/x/crypto/bcrypt`

---

### OAuth2 Integration
**Feature ID:** `oauth2`

**Description:** OAuth2 authentication with popular providers.

**Generated Components:**
- OAuth2 configuration
- Provider integrations (Google, GitHub, etc.)
- Token exchange
- User profile fetching

**Environment Variables:**
```env
OAUTH2_GOOGLE_CLIENT_ID=your-google-client-id
OAUTH2_GOOGLE_CLIENT_SECRET=your-google-client-secret
OAUTH2_GITHUB_CLIENT_ID=your-github-client-id
OAUTH2_GITHUB_CLIENT_SECRET=your-github-client-secret
OAUTH2_REDIRECT_URL=http://localhost:8080/auth/callback
```

**Dependencies:**
- `golang.org/x/oauth2`

## 🤖 AI Integration

### OpenAI Integration
**Feature ID:** `openai`

**Description:** OpenAI GPT models integration.

**Generated Components:**
- OpenAI client setup
- Chat completion handlers
- Streaming responses
- Error handling

**Environment Variables:**
```env
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_MAX_TOKENS=1000
OPENAI_TEMPERATURE=0.7
```

**Dependencies:**
- `github.com/sashabaranov/go-openai`

---

### Claude Integration
**Feature ID:** `claude`

**Description:** Anthropic Claude AI integration.

**Environment Variables:**
```env
CLAUDE_API_KEY=your-claude-api-key
CLAUDE_MODEL=claude-3-sonnet-20240229
CLAUDE_MAX_TOKENS=1000
```

---

### OpenRouter Integration
**Feature ID:** `openrouter`

**Description:** Multi-model AI API access through OpenRouter.

**Environment Variables:**
```env
OPENROUTER_API_KEY=your-openrouter-api-key
OPENROUTER_MODEL=openai/gpt-3.5-turbo
OPENROUTER_SITE_URL=https://yoursite.com
OPENROUTER_APP_NAME=YourApp
```

## 📡 Communication Protocols

### gRPC Support
**Feature ID:** `grpc`

**Description:** High-performance RPC framework.

**Generated Components:**
- gRPC server setup
- Protocol buffer definitions
- Service implementations
- Client generation

**Environment Variables:**
```env
GRPC_PORT=9090
GRPC_HOST=0.0.0.0
GRPC_MAX_RECV_MSG_SIZE=4194304
GRPC_MAX_SEND_MSG_SIZE=4194304
```

**Dependencies:**
- `google.golang.org/grpc`
- `google.golang.org/protobuf`

---

### WebSocket Support
**Feature ID:** `websocket`

**Description:** Real-time bidirectional communication.

**Generated Components:**
- WebSocket server setup
- Connection management
- Message broadcasting
- Room/channel support

**Environment Variables:**
```env
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_CHECK_ORIGIN=false
WS_ENABLE_COMPRESSION=true
```

**Dependencies:**
- `github.com/gorilla/websocket`

## 🐳 DevOps & Deployment

### Docker Support
**Feature ID:** `docker`

**Description:** Container configuration for deployment.

**Generated Files:**
- `Dockerfile` - Multi-stage build
- `docker-compose.yml` - Development environment
- `.dockerignore` - Build optimization

**Configuration:**
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

---

### Makefile
**Feature ID:** `makefile`

**Description:** Build automation and development tasks.

**Generated Targets:**
- `make build` - Build the application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make docker` - Build Docker image
- `make dev` - Start development server

---

### Swagger Documentation
**Feature ID:** `swagger`

**Description:** Automatic API documentation generation.

**Generated Components:**
- Swagger annotations
- API documentation server
- Interactive UI
- JSON/YAML export

**Environment Variables:**
```env
SWAGGER_ENABLED=true
SWAGGER_HOST=localhost:8080
SWAGGER_BASE_PATH=/api/v1
```

**Dependencies:**
- `github.com/swaggo/swag`
- `github.com/swaggo/gin-swagger` (or echo/fiber variants)

## 🔧 Utility Features

### Environment Configuration
**Feature ID:** `env` (auto-included)

**Description:** Environment variable management.

**Generated Components:**
- `.env.example` file
- Configuration loading
- Environment validation

**Dependencies:**
- `github.com/joho/godotenv`

---

### Git Integration
**Feature ID:** `gitignore` (auto-included)

**Description:** Git repository setup.

**Generated Files:**
- `.gitignore` - Ignore patterns
- `README.md` - Project documentation

---

### Middleware Stack
**Feature ID:** `middleware` (auto-included)

**Description:** Common HTTP middleware.

**Components:**
- CORS handling
- Request logging
- Rate limiting
- Recovery middleware
- Security headers

## 🚫 Feature Conflicts

### Database Conflicts
- **Cannot combine:** Multiple primary databases
- **Examples:** `postgresql` + `mysql`, `mongodb` + `sqlite`
- **Allowed:** Primary database + cache (e.g., `postgresql` + `redis`)

### Framework Conflicts
- **Cannot combine:** Multiple web frameworks
- **Examples:** `gin` + `echo`, `fiber` + `gin`
- **Solution:** Choose one framework per project

### ORM Conflicts
- **Cannot combine:** Multiple ORMs
- **Examples:** `gorm` + `sqlc`
- **Solution:** Choose one ORM approach

## 🔗 Feature Dependencies

### Automatic Dependencies
When you select certain features, others are automatically included:

- `auth` → `jwt`, `bcrypt`
- `grpc` → `protobuf`
- `websocket` → `gorilla/websocket`
- `docker` → `dockerfile`, `dockerignore`
- `swagger` → `swag`, framework-specific swagger

### Recommended Combinations

**REST API with Database:**
```json
{
  "framework": "gin",
  "features": ["postgresql", "auth", "swagger", "docker"]
}
```

**Microservice with gRPC:**
```json
{
  "framework": "gin",
  "features": ["postgresql", "grpc", "redis", "docker"]
}
```

**AI-Powered API:**
```json
{
  "framework": "fiber",
  "features": ["postgresql", "auth", "openai", "redis", "swagger"]
}
```

## 📊 Feature Statistics

Based on user generation patterns:

| Feature | Usage Rate | Category |
|---------|------------|----------|
| `postgresql` | 80% | Database |
| `auth` | 75% | Security |
| `docker` | 70% | DevOps |
| `redis` | 45% | Cache |
| `swagger` | 60% | Documentation |
| `gin` | 65% | Framework |
| `fiber` | 25% | Framework |
| `echo` | 10% | Framework |

## 🔮 Upcoming Features

### Planned Additions
- **TimescaleDB** - Time-series database
- **CockroachDB** - Distributed SQL database
- **NATS** - Message streaming
- **Prometheus** - Metrics collection
- **Jaeger** - Distributed tracing
- **GraphQL** - Query language
- **Chi Router** - Lightweight HTTP router

### Community Requests
- **Kafka** integration
- **Elasticsearch** support
- **Kubernetes** manifests
- **Terraform** configurations

---

For the most up-to-date feature list, check the [live API documentation](http://localhost:8080/features) or visit our [GitHub repository](https://github.com/telman03/gocraft).