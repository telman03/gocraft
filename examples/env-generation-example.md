# Environment Configuration Example

This example demonstrates how the GoCraft generator creates feature-specific environment configurations.

## Example Request

```json
{
  "projectName": "ecommerce-api",
  "features": [
    "postgresql",
    "auth", 
    "redis",
    "openai",
    "websocket",
    "observability",
    "docker"
  ]
}
```

## Generated .env.example

The generator will create a comprehensive `.env.example` file with sections for each selected feature:

```env
# =============================================================================
# ecommerce-api Environment Configuration
# =============================================================================
# Copy this file to .env and update the values for your environment
# Never commit .env files to version control!

# =============================================================================
# Server Configuration
# =============================================================================
APP_NAME=ecommerce-api
APP_ENV=development
PORT=8080
HOST=0.0.0.0
APP_VERSION=1.0.0

# =============================================================================
# Authentication & Security
# =============================================================================
JWT_SECRET=your-super-secret-jwt-key-change-in-production-min-32-chars
JWT_EXPIRATION=24h
JWT_REFRESH_EXPIRATION=168h
BCRYPT_COST=12
SESSION_SECRET=your-session-secret-key-change-in-production

# =============================================================================
# Database Configuration
# =============================================================================
# PostgreSQL Configuration
DATABASE_URL=postgres://postgres:password@localhost:5432/ecommerce-api?sslmode=disable
DB_HOST=localhost
DB_PORT=5432
DB_NAME=ecommerce-api
DB_USER=postgres
DB_PASSWORD=password
DB_SSLMODE=disable
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=25
DB_CONN_MAX_LIFETIME=300s

# =============================================================================
# Redis Configuration
# =============================================================================
REDIS_URL=redis://localhost:6379
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF=8ms
REDIS_MAX_RETRY_BACKOFF=512ms

# =============================================================================
# AI/LLM Configuration
# =============================================================================
# OpenAI Configuration
OPENAI_API_KEY=your-openai-api-key
OPENAI_MODEL=gpt-3.5-turbo
OPENAI_MAX_TOKENS=1000
OPENAI_TEMPERATURE=0.7

# =============================================================================
# WebSocket Configuration
# =============================================================================
WS_READ_BUFFER_SIZE=1024
WS_WRITE_BUFFER_SIZE=1024
WS_HANDSHAKE_TIMEOUT=10s
WS_CHECK_ORIGIN=false

# =============================================================================
# Monitoring & Observability
# =============================================================================
# Prometheus Configuration
PROMETHEUS_ENABLED=true
PROMETHEUS_PATH=/metrics
PROMETHEUS_PORT=2112

# Sentry Configuration
SENTRY_DSN=your-sentry-dsn
SENTRY_ENVIRONMENT=development
SENTRY_RELEASE=ecommerce-api@1.0.0
SENTRY_SAMPLE_RATE=1.0

# ... (additional configuration sections)
```

## Key Benefits

1. **Feature-Specific**: Only includes environment variables for selected features
2. **Production-Ready**: Includes security considerations and production overrides
3. **Well-Documented**: Each section is clearly labeled and documented
4. **Extensible**: Easy to add new variables as features are added
5. **Secure Defaults**: Includes security best practices and warnings

## Usage in Generated Project

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Update values for your environment:
   ```bash
   # Update database credentials
   DB_PASSWORD=your_secure_password
   
   # Add your API keys
   OPENAI_API_KEY=sk-your-actual-openai-key
   
   # Set production JWT secret
   JWT_SECRET=your-production-jwt-secret-min-32-characters
   ```

3. The generated application will automatically load these variables using a configuration package.

## Security Notes

- The `.env` file is automatically added to `.gitignore`
- All sensitive values use placeholder text with security warnings
- Production-specific overrides are clearly marked
- Minimum security requirements are documented (e.g., JWT secret length)