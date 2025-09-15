# GoCraft Template Validation System

## Overview

The GoCraft validation system prevents template conflicts and ensures generated projects are coherent and functional. It validates feature combinations, resolves conflicts, and automatically adds required dependencies.

## 🔍 Validation Rules

### 🗄️ Database Conflicts

#### Rule: Only One Relational Database
- **Allowed**: `mysql` OR `postgresql` OR `sqlite`
- **Forbidden**: Multiple relational databases together
- **Valid**: `postgresql` + `mongodb` + `redis` (different types)

```json
// ❌ Invalid - Multiple relational DBs
{
  "features": ["mysql", "postgresql", "gin"]
}

// ✅ Valid - Mixed database types
{
  "features": ["postgresql", "mongodb", "redis", "gin"]
}
```

### 🔧 ORM Conflicts

#### Rule: Only One ORM Framework
- **Allowed**: `gorm` OR `sqlc`
- **Forbidden**: Both GORM and SQLC together
- **Auto-dependency**: GORM + relational DB → adds `migrations`

```json
// ❌ Invalid - Multiple ORMs
{
  "features": ["gorm", "sqlc", "postgresql"]
}

// ✅ Valid - Single ORM
{
  "features": ["gorm", "postgresql"]
  // Auto-adds: migrations
}
```

### 🌐 Web Framework Conflicts

#### Rule: Only One Web Framework
- **Allowed**: `gin` OR `echo` OR `fiber`
- **Auto-default**: If none selected → adds `gin`
- **Forbidden**: Multiple frameworks together

```json
// ❌ Invalid - Multiple frameworks
{
  "features": ["gin", "echo", "postgresql"]
}

// ✅ Valid - Auto-adds framework
{
  "features": ["postgresql"]
  // Auto-adds: gin
}
```

### 🔐 Authentication Rules

#### Rule: Multiple Auth Methods Allowed (with Warning)
- **Allowed**: `auth` + `oauth2` (complementary use)
- **Warning**: Complexity warning when both selected
- **Auto-dependency**: Any auth → adds `middleware`

```json
// ✅ Valid but warns about complexity
{
  "features": ["auth", "oauth2", "gin"]
  // Auto-adds: middleware
}
```

### 🔄 Communication Protocol Rules

#### Rule: gRPC Requires Protocol Buffers
- **Dependency**: `grpc` → automatically adds `proto`
- **Valid**: `proto` without `grpc` (serialization only)

```json
// ✅ Valid - Auto-adds proto
{
  "features": ["grpc", "gin"]
  // Auto-adds: proto
}
```

### 📚 Documentation Rules

#### Rule: Swagger Requires Web Framework
- **Warning**: Swagger without framework
- **Valid**: Swagger works with any web framework

```json
// ⚠️ Warning - Swagger needs framework
{
  "features": ["swagger", "postgresql"]
  // Warning: No framework for Swagger
}
```

## 🚀 API Endpoints

### 1. Validate Features
```http
POST /generate/validate
Authorization: Bearer <jwt_token>
Content-Type: application/json

{
  "projectName": "my-api",
  "features": ["gin", "postgresql", "gorm", "auth"]
}
```

**Response:**
```json
{
  "project_name": "my-api",
  "original_features": ["gin", "postgresql", "gorm", "auth"],
  "validation_result": {
    "is_valid": true,
    "adjusted_features": ["gin", "postgresql", "gorm", "auth", "env", "gitignore", "main", "migrations", "middleware"],
    "warnings": [],
    "added_dependencies": ["env", "gitignore", "main", "migrations", "middleware"]
  },
  "supported_features": { ... }
}
```

### 2. Get Supported Features
```http
GET /features
```

**Response:**
```json
{
  "categories": {
    "databases": ["mysql", "postgresql", "sqlite", "mongodb", "redis"],
    "frameworks": ["gin", "echo", "fiber"],
    "orms": ["gorm", "sqlc"],
    "auth": ["auth", "oauth2"],
    "ai": ["openai", "claude", "openrouter"],
    "communication": ["grpc", "websocket", "proto"],
    "devops": ["dockerfile", "docker-compose", "makefile"],
    "documentation": ["swagger", "postman", "readme"],
    "utilities": ["logger", "middleware", "config", "migrations"],
    "core": ["env", "gitignore", "main"]
  },
  "descriptions": { ... },
  "conflict_rules": { ... }
}
```

## 🔧 Integration with Generation

The validation system is automatically integrated into the generation process:

1. **Pre-validation**: Features are validated before generation
2. **Conflict Resolution**: Conflicts are reported with suggestions
3. **Dependency Addition**: Required dependencies are automatically added
4. **Warning System**: Non-blocking warnings for complex configurations

```go
// In generate handler
validator := validation.NewTemplateValidator()
result := validator.ValidateFeatures(req.Features)

if !result.IsValid {
    return c.Status(400).JSON(result)
}

// Use adjusted features for generation
zipPath, err := builder.GenerateProject(req.ProjectName, result.AdjustedFeatures)
```

## 📋 Feature Categories

### Core Templates (Always Included)
- `env` - Environment configuration
- `gitignore` - Git ignore rules  
- `main` - Application entry point

### Database Templates
- **Relational**: `mysql`, `postgresql`, `sqlite` (mutually exclusive)
- **NoSQL**: `mongodb` (can coexist with relational)
- **Cache**: `redis`, `badger` (can coexist with others)

### Framework Templates
- **Web**: `gin`, `echo`, `fiber` (mutually exclusive)
- **Default**: `gin` (auto-added if none selected)

### Integration Templates
- **Auth**: `auth`, `oauth2` (can coexist with warnings)
- **AI**: `openai`, `claude`, `openrouter` (multiple allowed)
- **Communication**: `grpc`, `websocket`, `proto`

### Development Templates
- **DevOps**: `dockerfile`, `docker-compose`, `makefile`
- **Documentation**: `swagger`, `postman`, `readme`
- **Utilities**: `logger`, `middleware`, `config`, `migrations`

## 🎯 Automatic Dependencies

The system automatically adds required dependencies:

| Feature | Auto-Adds | Reason |
|---------|-----------|---------|
| `grpc` | `proto` | gRPC requires Protocol Buffers |
| `gorm` + relational DB | `migrations` | Database migrations needed |
| `auth` or `oauth2` | `middleware` | Authentication middleware required |
| Multiple services | `config` | Configuration management needed |
| No framework | `gin` | Default web framework |

## 🚨 Error Handling

### Conflict Errors (HTTP 400)
```json
{
  "error": "Feature validation failed",
  "validation_result": {
    "is_valid": false,
    "errors": [
      {
        "message": "Multiple relational databases selected: mysql, postgresql",
        "conflicts": ["mysql", "postgresql"],
        "suggestions": [
          "Choose only one relational database (MySQL, PostgreSQL, or SQLite)",
          "MongoDB and Redis can be used alongside a relational database"
        ]
      }
    ]
  }
}
```

### Warning System
Non-blocking warnings for:
- Complex configurations (multiple auth methods)
- Missing recommended features (no database)
- Compatibility notes (multiple database types)

## 🧪 Testing Examples

### Valid Configurations
```json
// Basic API
["gin", "postgresql", "gorm", "auth"]

// Microservice with gRPC
["grpc", "postgresql", "gorm", "redis"]
// Auto-adds: proto, gin, env, gitignore, main, migrations, middleware, config

// AI-powered API
["gin", "postgresql", "openai", "claude", "auth"]
// Auto-adds: env, gitignore, main, middleware

// Full-stack setup
["gin", "postgresql", "gorm", "auth", "redis", "websocket", "swagger", "dockerfile"]
// Auto-adds: env, gitignore, main, migrations, middleware, config
```

### Invalid Configurations
```json
// Multiple relational DBs
["mysql", "postgresql", "gin"] // ❌ Error

// Multiple ORMs  
["gorm", "sqlc", "postgresql"] // ❌ Error

// Multiple frameworks
["gin", "echo", "postgresql"] // ❌ Error
```

## 🔄 Alias Handling

The system handles common aliases:
- `postgres` → `postgresql`
- `mongo` → `mongodb`
- `jwt` → `auth`
- `authentication` → `auth`
- `websockets` → `websocket`
- `docker` → `dockerfile`
- `env-config` → `env`

This ensures consistent feature naming regardless of input variations.