# GoCraft Backend Generator

A powerful Go backend project generator that creates production-ready scaffolds with configurable features and automatic environment setup.

## Features

### 🚀 Core Features
- **Automatic Project Scaffolding**: Generate complete Go backend projects with selected features
- **Environment Configuration**: Auto-generated `.env.example` files with feature-specific variables
- **Security Best Practices**: Secure OTP generation, JWT authentication, input validation
- **Database Support**: PostgreSQL, MySQL, SQLite, MongoDB, Redis, BadgerDB
- **Web Frameworks**: Fiber, Gin, Echo support
- **Authentication**: JWT, OAuth2 integration
- **AI Integration**: OpenAI, OpenRouter, Claude API support
- **Real-time Communication**: WebSocket and gRPC support
- **Observability**: Prometheus metrics, logging, monitoring
- **Development Tools**: Docker, Docker Compose, Makefile, GitHub Actions

### 🔧 Environment Configuration

Every generated project includes a comprehensive `.env.example` file with:

#### Core Configuration
```env
# Server Configuration
APP_NAME=your-project-name
APP_ENV=development
PORT=8080
HOST=0.0.0.0
```

#### Feature-Specific Variables
The generator automatically includes environment variables based on selected features:

- **Database Features** (`postgresql`, `mysql`, `sqlite`, `mongodb`, `redis`):
  ```env
  # PostgreSQL Configuration
  DATABASE_URL=postgres://postgres:password@localhost:5432/your-project?sslmode=disable
  DB_HOST=localhost
  DB_PORT=5432
  DB_NAME=your-project
  DB_USER=postgres
  DB_PASSWORD=password
  ```

- **Authentication** (`auth`, `jwt`):
  ```env
  # Authentication & Security
  JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
  JWT_EXPIRATION=24h
  BCRYPT_COST=12
  ```

- **AI Integration** (`openai`, `claude`, `openrouter`):
  ```env
  # AI/LLM Configuration
  OPENAI_API_KEY=your-openai-api-key
  OPENAI_MODEL=gpt-3.5-turbo
  ```

- **Real-time Features** (`websocket`, `grpc`):
  ```env
  # WebSocket Configuration
  WS_READ_BUFFER_SIZE=1024
  WS_WRITE_BUFFER_SIZE=1024
  
  # gRPC Configuration
  GRPC_PORT=9090
  GRPC_HOST=0.0.0.0
  ```

- **Observability** (`observability`, `prometheus`):
  ```env
  # Monitoring & Observability
  PROMETHEUS_ENABLED=true
  SENTRY_DSN=your-sentry-dsn
  ```

## API Endpoints

### Authentication
- `POST /auth/register` - User registration with OTP verification
- `POST /auth/verify-otp` - Verify OTP code
- `POST /auth/resend-otp` - Resend OTP code
- `POST /auth/login` - User login
- `GET /auth/me` - Get current user (requires authentication)

### Project Generation
- `POST /generate` - Generate and download project scaffold (requires authentication)

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL (for user management)
- Environment variables configured

### Installation

1. Clone the repository:
```bash
git clone https://github.com/telman03/ai-backend-generator.git
cd ai-backend-generator
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

4. Run database migrations:
```bash
go run main.go
```

5. Start the server:
```bash
go run main.go
```

The server will start on `http://localhost:8081`

### Usage

1. **Register a new account**:
```bash
curl -X POST http://localhost:8081/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

2. **Verify your email** with the OTP received:
```bash
curl -X POST http://localhost:8081/auth/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"123456"}'
```

3. **Generate a project**:
```bash
curl -X POST http://localhost:8081/generate \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "projectName": "my-awesome-api",
    "features": ["postgresql", "auth", "redis", "openai", "websocket"]
  }' \
  --output my-awesome-api.zip
```

## Available Features

| Feature | Description | Environment Variables Added |
|---------|-------------|----------------------------|
| `postgresql` | PostgreSQL database support | `DATABASE_URL`, `DB_HOST`, `DB_PORT`, etc. |
| `mysql` | MySQL database support | `MYSQL_HOST`, `MYSQL_PORT`, etc. |
| `sqlite` | SQLite database support | `SQLITE_DATABASE`, `SQLITE_CACHE` |
| `mongodb` | MongoDB support | `MONGODB_URL`, `MONGODB_HOST`, etc. |
| `redis` | Redis cache/session store | `REDIS_URL`, `REDIS_HOST`, etc. |
| `auth` | JWT authentication | `JWT_SECRET`, `JWT_EXPIRATION`, etc. |
| `oauth2` | OAuth2 integration | OAuth provider configurations |
| `openai` | OpenAI API integration | `OPENAI_API_KEY`, `OPENAI_MODEL` |
| `claude` | Claude API integration | `CLAUDE_API_KEY`, `CLAUDE_MODEL` |
| `openrouter` | OpenRouter API integration | `OPENROUTER_API_KEY` |
| `websocket` | WebSocket support | `WS_READ_BUFFER_SIZE`, etc. |
| `grpc` | gRPC server support | `GRPC_PORT`, `GRPC_HOST`, etc. |
| `gin` | Gin web framework | Framework-specific configs |
| `echo` | Echo web framework | Framework-specific configs |
| `observability` | Monitoring & metrics | `PROMETHEUS_ENABLED`, `SENTRY_DSN` |
| `docker` | Docker configuration | Dockerfile and docker-compose.yml |
| `swagger` | API documentation | Swagger/OpenAPI setup |

## Project Structure

Generated projects follow this structure:

```
your-project/
├── .env.example          # Environment configuration template
├── .gitignore           # Git ignore rules
├── main.go              # Application entry point
├── go.mod               # Go module definition
├── Dockerfile           # Docker configuration (if selected)
├── docker-compose.yml   # Docker Compose (if selected)
├── Makefile            # Build automation (if selected)
├── README.md           # Project documentation
├── internal/
│   ├── auth/           # Authentication logic
│   ├── config/         # Configuration management
│   ├── db/             # Database connections
│   ├── handlers/       # HTTP handlers
│   ├── middleware/     # HTTP middleware
│   ├── models/         # Data models
│   └── router/         # Route definitions
└── docs/               # API documentation
```

## Security Features

- **Cryptographically Secure OTP**: Uses `crypto/rand` for secure random generation
- **Input Validation**: Comprehensive validation with detailed error messages
- **JWT Authentication**: Secure token-based authentication
- **Password Hashing**: bcrypt with configurable cost
- **Rate Limiting**: Built-in rate limiting support
- **CORS Configuration**: Configurable CORS policies

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o bin/gocraft main.go
```

### Docker
```bash
docker build -t gocraft .
docker run -p 8081:8081 gocraft
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support, email support@gocraft.dev or create an issue on GitHub.