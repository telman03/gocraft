# 🚀 GoCraft - Go Backend Generator

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![API Documentation](https://img.shields.io/badge/API-Swagger-85EA2D?style=flat&logo=swagger)](http://localhost:8080/swagger/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)

**GoCraft** is a powerful, production-ready Go backend generator that helps developers quickly scaffold modern microservices with their preferred frameworks, databases, and features. Generate complete, well-structured Go applications in seconds with built-in best practices, security features, and comprehensive documentation.

---

## ✨ Why Choose GoCraft?

- 🚀 **Lightning Fast**: Generate production-ready Go backends in seconds
- 🏗️ **Modular Architecture**: Choose exactly the features you need
- 🔒 **Security First**: Built-in authentication, validation, and security best practices
- 📚 **Well Documented**: Comprehensive documentation and examples for every feature
- 🎯 **Production Ready**: Generated code follows Go best practices and industry standards
- 🔧 **Highly Configurable**: Extensive customization options for every component

---

## 🎯 Quick Start

### Prerequisites

- **Go 1.23+** - [Install Go](https://golang.org/doc/install)
- **PostgreSQL** - For user management and project history
- **Git** - For version control

### 🚀 Get Started in 3 Steps

1. **Clone and Setup**
   ```bash
   git clone https://github.com/telman03/gocraft-backend.git
   cd gocraft-backend
   go mod tidy
   ```

2. **Configure Environment**
   ```bash
   cp .env.example .env
   # Edit .env with your database credentials
   ```

3. **Start the Server**
   ```bash
   go run cmd/gocraft/main.go
   ```

Your GoCraft server is now running at `http://localhost:8080` 🎉

---

## 🏗️ Architecture Overview

GoCraft follows a clean, modular architecture designed for scalability and maintainability:

```
gocraft/
├── cmd/gocraft/           # Application entry point
├── internal/              # Private application code
│   ├── api/              # HTTP routing and middleware
│   ├── auth/             # Authentication system
│   ├── builder/          # Project generation engine
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware stack
│   ├── models/           # Data models
│   ├── services/         # Business logic services
│   ├── templates/        # Code generation templates
│   └── validation/       # Input validation
├── docs/                 # Documentation
├── examples/             # Usage examples
└── tests/               # Test suites
```

---

## 🛠️ Supported Technologies

### 🌐 Web Frameworks
- **Gin** - Fast and minimalist web framework
- **Echo** - High performance, extensible framework  
- **Fiber** - Express-inspired framework built on Fasthttp

### 🗄️ Databases
- **PostgreSQL** - Advanced relational database
- **MySQL** - Popular relational database
- **SQLite** - Lightweight embedded database
- **MongoDB** - NoSQL document database
- **Redis** - In-memory cache and data store

### 🔐 Authentication & Security
- **JWT Authentication** - Secure token-based auth
- **OAuth2** - Social login integration
- **OTP Verification** - Email-based verification
- **Rate Limiting** - Built-in request throttling
- **Input Validation** - Comprehensive request validation

### 🤖 AI Integration
- **OpenAI** - GPT models integration
- **Claude** - Anthropic's AI assistant
- **OpenRouter** - Multi-model AI access

### 📡 Communication
- **gRPC** - High-performance RPC framework
- **WebSocket** - Real-time bidirectional communication
- **REST API** - Standard HTTP API structure

### 🐳 DevOps & Deployment
- **Docker** - Container configuration
- **Docker Compose** - Multi-service orchestration
- **Makefile** - Build automation
- **Swagger** - Interactive API documentation

---

## 📖 Usage Examples

### 🔐 User Registration & Authentication

```bash
# Register a new user
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"developer@example.com","password":"SecurePass123!"}'

# Verify email with OTP
curl -X POST http://localhost:8080/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"developer@example.com","otp":"123456"}'

# Login and get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"developer@example.com","password":"SecurePass123!"}'
```

### 🏗️ Project Generation

```bash
# Validate feature combination
curl -X POST http://localhost:8080/generate/validate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "projectName": "my-awesome-api",
    "features": ["gin", "postgresql", "auth", "redis", "swagger", "docker"]
  }'

# Generate and download project
curl -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "projectName": "my-awesome-api",
    "features": ["gin", "postgresql", "auth", "redis", "swagger", "docker"]
  }' \
  --output my-awesome-api.zip
```

### 📊 Project Management

```bash
# View project history
curl -X GET http://localhost:8080/api/history \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Get project statistics
curl -X GET http://localhost:8080/api/history/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# Download previous project
curl -X GET http://localhost:8080/api/history/123/download \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  --output project-backup.zip
```

---

## 🎨 Generated Project Structure

Every generated project follows Go best practices and includes:

```
your-awesome-api/
├── .env.example          # Environment configuration template
├── .gitignore           # Git ignore patterns
├── .dockerignore        # Docker ignore patterns (if Docker selected)
├── Dockerfile           # Multi-stage Docker build (if Docker selected)
├── docker-compose.yml   # Development environment (if Docker selected)
├── Makefile            # Build automation (if Makefile selected)
├── README.md           # Comprehensive project documentation
├── go.mod              # Go module definition
├── cmd/                # Application entry points
│   └── gocraft/        # Main application
│       └── main.go     # Application entry point
├── internal/           # Private application code
│   ├── auth/          # Authentication logic (if auth selected)
│   ├── config/        # Configuration management
│   ├── database/      # Database connections
│   ├── handlers/      # HTTP request handlers
│   ├── middleware/    # HTTP middleware stack
│   ├── models/        # Data models and schemas
│   ├── routes/        # Route definitions
│   └── services/      # Business logic services
├── docs/              # API documentation (if Swagger selected)
├── migrations/        # Database migrations (if database selected)
└── tests/            # Test files and fixtures
```

---

## 🔧 Configuration

### Environment Variables

GoCraft automatically generates comprehensive `.env.example` files based on selected features:

```env
# Core Application
APP_NAME=my-awesome-api
APP_ENV=development
PORT=8080
HOST=0.0.0.0

# Database (PostgreSQL example)
DATABASE_URL=postgres://user:password@localhost:5432/myapp?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=myapp
DB_USER=postgres
DB_PASSWORD=password

# Authentication & Security
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
JWT_EXPIRATION=24h
BCRYPT_COST=12

# Redis Cache
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# AI Integration (if selected)
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_MAX_TOKENS=1000
```

### Feature Compatibility

GoCraft includes intelligent conflict detection:

✅ **Compatible Combinations:**
- `gin` + `postgresql` + `auth` + `redis` + `swagger`
- `fiber` + `mongodb` + `websocket` + `docker`
- `echo` + `mysql` + `grpc` + `openai`

❌ **Conflicting Features:**
- Multiple frameworks: `gin` + `echo` + `fiber`
- Multiple primary databases: `postgresql` + `mysql` + `mongodb`

---

## 🔒 Security Features

GoCraft prioritizes security with built-in features:

- 🔐 **JWT Authentication** - Secure token-based authentication
- 🛡️ **Input Validation** - Comprehensive request validation and sanitization
- 🔒 **Password Security** - bcrypt hashing with configurable cost
- 📧 **OTP Verification** - Cryptographically secure email verification
- 🚦 **Rate Limiting** - Configurable request throttling (100 req/min default)
- 🌐 **CORS Protection** - Configurable cross-origin resource sharing
- 🔍 **SQL Injection Prevention** - Parameterized queries and ORM protection
- 📝 **Audit Logging** - Comprehensive security event logging

---

## 🧪 Development & Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test package
go test ./internal/handlers/...
```

### Building for Production

```bash
# Build binary
go build -o bin/gocraft cmd/gocraft/main.go

# Build with optimizations
go build -ldflags="-s -w" -o bin/gocraft cmd/gocraft/main.go

# Cross-compile for Linux
GOOS=linux GOARCH=amd64 go build -o bin/gocraft-linux cmd/gocraft/main.go
```

### Docker Development

```bash
# Build Docker image
docker build -t gocraft:latest .

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f gocraft
```

---

## 📊 API Documentation

### Interactive Documentation

Visit `http://localhost:8080/swagger/` for interactive API documentation with:

- 📖 Complete endpoint documentation
- 🧪 Interactive request testing
- 📝 Request/response examples
- 🔐 Authentication testing
- 📋 Schema definitions

### Key Endpoints

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| `GET` | `/health` | Health check | ❌ |
| `GET` | `/features` | List supported features | ❌ |
| `POST` | `/auth/register` | User registration | ❌ |
| `POST` | `/auth/login` | User authentication | ❌ |
| `POST` | `/generate` | Generate project | ✅ |
| `GET` | `/api/history` | Project history | ✅ |
| `GET` | `/api/admin/users` | Admin: List users | 👑 |

---

## 🤝 Contributing

We welcome contributions from the community! Here's how to get started:

### 🚀 Quick Contribution Guide

1. **Fork the Repository**
   ```bash
   git clone https://github.com/YOUR_USERNAME/gocraft-backend.git
   cd ai-backend-generator
   ```

2. **Create a Feature Branch**
   ```bash
   git checkout -b feature/amazing-new-feature
   ```

3. **Make Your Changes**
   - Follow Go best practices and existing code style
   - Add tests for new functionality
   - Update documentation as needed

4. **Test Your Changes**
   ```bash
   go test ./...
   go fmt ./...
   go vet ./...
   ```

5. **Submit a Pull Request**
   - Provide a clear description of your changes
   - Reference any related issues
   - Ensure all tests pass

### 🎯 Areas for Contribution

- 🆕 **New Features**: Add support for new frameworks, databases, or integrations
- 🐛 **Bug Fixes**: Help us squash bugs and improve reliability
- 📚 **Documentation**: Improve docs, add examples, or create tutorials
- 🧪 **Testing**: Expand test coverage and add integration tests
- 🎨 **Templates**: Create new project templates or improve existing ones

### 📋 Development Guidelines

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write clear, descriptive commit messages
- Add tests for new functionality
- Update documentation for user-facing changes
- Use `gofmt` and `go vet` before submitting

---

## 📄 License

This project is licensed under the **MIT License** - see the [LICENSE](LICENSE) file for details.

The MIT License allows you to:
- ✅ Use the software for any purpose
- ✅ Modify and distribute the software
- ✅ Include in proprietary software
- ✅ Sell copies of the software

---

## 🆘 Support & Community

### 📞 Getting Help

- 📖 **Documentation**: Check our comprehensive [docs](./docs/)
- 🐛 **Bug Reports**: [Create an issue](https://github.com/telman03/gocraft-backend/issues/new)
- 💡 **Feature Requests**: [Request a feature](https://github.com/telman03/gocraft-backend/issues/new)
- 💬 **Discussions**: [Join the conversation](https://github.com/telman03/gocraft-backend/discussions)

### 🌟 Show Your Support

If GoCraft helps you build better Go applications, please consider:

- ⭐ **Star the repository** on GitHub
- 🐦 **Share on social media** with `#GoCraft`
- 📝 **Write a blog post** about your experience
- 🤝 **Contribute** to the project

---

## 🚀 What's Next?

### 🔮 Upcoming Features

- 🌐 **GraphQL Support** - GraphQL server generation
- 📊 **Prometheus Metrics** - Built-in observability
- 🔄 **Message Queues** - NATS, RabbitMQ, Kafka integration
- 🏗️ **Microservices** - Multi-service project generation
- 🎨 **Custom Templates** - User-defined project templates
- 📱 **CLI Tool** - Command-line interface for project generation

### 📈 Roadmap

- **Q1 2025**: GraphQL support, Prometheus metrics
- **Q2 2025**: Message queue integrations, CLI tool
- **Q3 2025**: Microservices architecture, custom templates
- **Q4 2025**: Advanced monitoring, performance optimizations

---

<div align="center">

**Built with ❤️ by the GoCraft team**

[Website](https://gocraft.dev) • [Documentation](./docs/) • [API Reference](http://localhost:8080/swagger/) • [GitHub](https://github.com/telman03/gocraft-backend)

</div>