# Shared Backend Packages

This directory contains shared Go packages used across all microservices in the document management system.

## Overview

These packages provide common functionality to ensure consistency, reduce code duplication, and enforce best practices across all services.

## Packages

### 1. errors - Custom Error Handling

**Location:** `pkg/errors/`

**Purpose:** Provides structured error types with HTTP status code mapping for consistent error handling across all services.

**Features:**
- Machine-readable error codes for frontend handling
- HTTP status code mapping
- Field-level validation errors
- Error metadata support
- Error wrapping and unwrapping
- Predefined common errors

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/errors"

// Using predefined errors
if user == nil {
    return errors.ErrNotFound
}

// Creating custom errors
if exists {
    return errors.Conflictf("user with email %s already exists", email)
}

// Wrapping internal errors
if err := db.Query(...); err != nil {
    return errors.Internalf(err, "failed to query users")
}

// Validation errors with field details
err := errors.ErrValidation.
    WithField("email", "must be a valid email").
    WithField("password", "must be at least 8 characters")
```

### 2. config - Configuration Management

**Location:** `pkg/config/`

**Purpose:** Loads and validates application configuration from environment variables using Viper.

**Features:**
- Environment variable support
- Configuration validation
- Type-safe configuration structs
- Default values
- Development/production modes

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/config"

// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Access configuration
dbDSN := cfg.Database.GetDSN()
redisAddr := cfg.Redis.GetRedisAddr()

// Check environment
if cfg.IsDevelopment() {
    // Development-specific logic
}
```

### 3. logger - Structured Logging

**Location:** `pkg/logger/`

**Purpose:** Provides structured logging with Zap, supporting context-aware logging with request/tenant/user IDs.

**Features:**
- Structured JSON logging (production)
- Colorized console logging (development)
- Context-aware logging
- Request ID, Tenant ID, User ID correlation
- Multiple log levels (debug, info, warn, error, fatal)
- Global and instance-based loggers

**Usage:**
```go
import (
    "github.com/SidahmedSeg/document-manager/backend/pkg/logger"
    "go.uber.org/zap"
)

// Create logger
log, err := logger.New("production", "info", "json")

// Simple logging
logger.Info("service started", zap.Int("port", 8080))

// Context-aware logging
ctx = logger.WithRequestID(ctx, requestID)
ctx = logger.WithTenantID(ctx, tenantID)
logger.InfoContext(ctx, "processing request")
```

### 4. database - PostgreSQL Client

**Location:** `pkg/database/`

**Purpose:** Provides PostgreSQL connection pooling, transaction management, and helper functions.

**Features:**
- Connection pool management
- Health checks
- Transaction helpers
- Tenant context support
- Query error wrapping
- Common database operations

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/database"

// Connect to database
db, err := database.NewPostgresDB(cfg.Database, logger)
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Health check
if err := db.HealthCheck(ctx); err != nil {
    log.Error("database unhealthy", zap.Error(err))
}

// Execute query
result, err := db.ExecContext(ctx, "INSERT INTO users ...")

// Use transaction
err = db.WithTransaction(ctx, func(tx *sql.Tx) error {
    _, err := tx.Exec("INSERT ...")
    return err
})
```

### 5. cache - Redis Client

**Location:** `pkg/cache/`

**Purpose:** Provides Redis client wrapper with JSON serialization and common caching patterns.

**Features:**
- Connection pooling
- Automatic JSON serialization
- TTL management
- Hash, Set, String operations
- Health checks
- Key namespace helpers

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/cache"

// Connect to Redis
cache, err := cache.NewRedisCache(cfg.Redis, logger)
if err != nil {
    log.Fatal(err)
}
defer cache.Close()

// Set value with TTL
err = cache.Set(ctx, "user:123", user, 1*time.Hour)

// Get value
var user User
err = cache.Get(ctx, "user:123", &user)

// Tenant-scoped keys
key := cache.TenantKey(tenantID, "documents", docID)
```

### 6. middleware - HTTP Middleware

**Location:** `pkg/middleware/`

**Purpose:** Provides HTTP middleware for authentication, logging, recovery, CORS, and more.

**Features:**
- Oathkeeper header extraction
- Request ID generation
- Structured request logging
- Panic recovery
- CORS support
- Request timeout
- Tenant context enforcement

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/middleware"

// Build middleware chain
handler := middleware.RequestID()(handler)
handler = middleware.ExtractAuthHeaders(logger)(handler)
handler = middleware.Logging(logger)(handler)
handler = middleware.Recovery(logger)(handler)
handler = middleware.CORS(allowedOrigins)(handler)

// Get auth context in handler
userID := middleware.GetUserID(r.Context())
tenantID := middleware.GetTenantID(r.Context())
```

### 7. validator - Input Validation

**Location:** `pkg/validator/`

**Purpose:** Provides input validation using go-playground/validator with custom rules.

**Features:**
- Struct validation
- Custom validation rules (UUID, file types, etc.)
- User-friendly error messages
- Field-level error mapping
- Helper validation functions

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/validator"

// Define struct with validation tags
type CreateUserRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Password string `json:"password" validate:"required,min=8"`
}

// Validate struct
if err := validator.Validate(req); err != nil {
    response.ValidationError(w, err)
    return
}

// Validate single values
if err := validator.ValidateUUID(userID); err != nil {
    // Handle error
}
```

### 8. response - Standardized JSON Responses

**Location:** `pkg/response/`

**Purpose:** Provides standardized JSON response formatting for all API endpoints.

**Features:**
- Consistent response structure
- Error response formatting
- Pagination support
- Success/error helpers
- HTTP status code helpers

**Usage:**
```go
import "github.com/SidahmedSeg/document-manager/backend/pkg/response"

// Success response
response.Success(w, users)

// Created response
response.Created(w, newUser)

// Error response
response.Error(w, err)

// Paginated response
response.Paginated(w, users, page, limit, total)

// Validation error
response.ValidationError(w, err)

// Helper responses
response.NotFound(w, "User not found")
response.Forbidden(w, "Access denied")
```

## Response Format

All API responses follow this structure:

### Success Response
```json
{
  "success": true,
  "data": { ... },
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "total_pages": 5
  }
}
```

### Error Response
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "email": "email is required",
      "password": "password must be at least 8 characters"
    }
  }
}
```

## Dependencies

All packages use these core dependencies:

- **Viper** - Configuration management
- **Zap** - Structured logging
- **PostgreSQL** (lib/pq) - Database driver
- **Redis** (go-redis/v9) - Cache client
- **Validator** (go-playground/validator/v10) - Input validation
- **UUID** (google/uuid) - UUID generation

## Usage in Services

All microservices should use these shared packages:

```go
package main

import (
    "github.com/SidahmedSeg/document-manager/backend/pkg/config"
    "github.com/SidahmedSeg/document-manager/backend/pkg/logger"
    "github.com/SidahmedSeg/document-manager/backend/pkg/database"
    "github.com/SidahmedSeg/document-manager/backend/pkg/cache"
    "github.com/SidahmedSeg/document-manager/backend/pkg/middleware"
    "github.com/SidahmedSeg/document-manager/backend/pkg/validator"
    "github.com/SidahmedSeg/document-manager/backend/pkg/response"
    "github.com/SidahmedSeg/document-manager/backend/pkg/errors"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }

    // Initialize logger
    log, err := logger.New(cfg.Environment, cfg.Logger.Level, cfg.Logger.Format)
    if err != nil {
        panic(err)
    }
    logger.SetGlobal(log)

    // Connect to database
    db, err := database.NewPostgresDB(cfg.Database, log)
    if err != nil {
        log.Fatal("failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Connect to cache
    cache, err := cache.NewRedisCache(cfg.Redis, log)
    if err != nil {
        log.Fatal("failed to connect to cache", zap.Error(err))
    }
    defer cache.Close()

    // Setup HTTP server with middleware
    mux := http.NewServeMux()

    // Apply middleware chain
    handler := middleware.RequestID()(mux)
    handler = middleware.ExtractAuthHeaders(log)(handler)
    handler = middleware.Logging(log)(handler)
    handler = middleware.Recovery(log)(handler)

    // Start server
    log.Info("starting server", zap.String("addr", cfg.Server.GetServerAddr()))
    http.ListenAndServe(cfg.Server.GetServerAddr(), handler)
}
```

## Testing

Each package includes unit tests. Run tests with:

```bash
go test ./pkg/...
```

## Best Practices

1. **Always use structured logging** - Never use `fmt.Println` or `log.Println`
2. **Wrap errors properly** - Use `errors.Wrap()` to add context to errors
3. **Validate all input** - Use validator package for all user input
4. **Use standardized responses** - Always use response package for HTTP responses
5. **Context propagation** - Always pass context through the call stack
6. **Connection reuse** - Initialize database and cache connections once, reuse across handlers

## Contributing

When adding new shared functionality:

1. Create new package in `pkg/`
2. Add comprehensive documentation
3. Include unit tests
4. Update this README
5. Update all services to use the new package

---

**Last Updated:** 2025-12-19
**Phase:** Phase 2 Complete
