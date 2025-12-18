# Simplified Implementation Plan - Document Manager
**Multi-Tenant Document Management System (Without Prometheus & Oathkeeper)**

---

## Simplified Architecture

### System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                  Shared Authentication Layer                     │
│  ┌──────────────┐                    ┌──────────────┐           │
│  │ Ory Kratos   │                    │  Ory Hydra   │           │
│  │ (Identity)   │◄──────────────────►│ (OAuth2/OIDC)│           │
│  │ :14433/:14434│                    │ :14444/:14445│           │
│  └──────────────┘                    └──────────────┘           │
└─────────────────────────────────────────────────────────────────┘
                           │
                           │ OAuth2 Flow
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Document Manager Application                     │
│                                                                   │
│  ┌─────────────────────────────────────────────────┐             │
│  │         Backend Microservices (Go + Fiber)      │             │
│  │  ┌──────────────────────────────────────────┐   │             │
│  │  │ Each service validates JWT directly      │   │             │
│  │  │ No API Gateway needed                    │   │             │
│  │  └──────────────────────────────────────────┘   │             │
│  │                                                  │             │
│  │  Tenant (10001)        Document (10002)         │             │
│  │  Storage (10003)       Share (10004)            │             │
│  │  RBAC (10005)          Quota (10006)            │             │
│  │  OCR (10007)           Categorization (10008)   │             │
│  │  Search (10009)        Notification (10010)     │             │
│  │  Audit (10011)                                  │             │
│  └─────────────────────────────────────────────────┘             │
│                                                                   │
│  ┌─────────────────────────────────────────────────┐             │
│  │           Infrastructure Services                │             │
│  │  ┌──────────────────────────────────────────┐   │             │
│  │  │ PostgreSQL │ Redis │ MinIO             │   │             │
│  │  │ Meilisearch │ NATS │ ClickHouse       │   │             │
│  │  │ PaddleOCR │ MailSlurper (dev only)    │   │             │
│  │  └──────────────────────────────────────────┘   │             │
│  └─────────────────────────────────────────────────┘             │
│                                                                   │
│  ┌─────────────────────────────────────────────────┐             │
│  │              Frontend Applications               │             │
│  │  ┌──────────────────────────────────────────┐   │             │
│  │  │ User App (13000)  │ Admin App (13001)   │   │             │
│  │  │ Next.js 14+       │ Next.js 14+         │   │             │
│  │  └──────────────────────────────────────────┘   │             │
│  └─────────────────────────────────────────────────┘             │
└─────────────────────────────────────────────────────────────────┘
```

### Key Changes

**Removed**:
- ❌ Ory Oathkeeper (API Gateway)
- ❌ Prometheus (Metrics collection)
- ❌ Grafana (Monitoring dashboards)

**Impact**:
- ✅ Each backend service validates JWTs directly using Hydra's public key
- ✅ Simpler deployment (fewer services to manage)
- ✅ Direct frontend → backend communication
- ✅ Can add monitoring later when needed

---

## Updated Port Allocation

```
Backend Services:
- Tenant Service:          10001
- Document Service:        10002
- Storage Service:         10003
- Share Service:           10004
- RBAC Service:            10005
- Quota Service:           10006
- OCR Service:             10007
- Categorization Service:  10008
- Search Service:          10009
- Notification Service:    10010
- Audit Service:           10011

Frontend:
- User App:                13000
- Admin App:               13001

Infrastructure:
- PostgreSQL:              15432
- Redis:                   16379
- Meilisearch:             17700
- MinIO (API):             19000
- MinIO (Console):         19001
- NATS:                    14222
- NATS (Monitor):          18222
- ClickHouse (HTTP):       18123
- ClickHouse (Native):     19000
- MailSlurper (SMTP):      14436
- MailSlurper (Web):       14437
```

---

## Updated Phase 1: Infrastructure Setup (Week 1-2)

**Simplified Deliverables**:
1. ✅ Docker Compose with core services (8 services instead of 15+)
2. ✅ Environment configuration (.env.example)
3. ✅ Database migrations (all 10 migrations)
4. ✅ Makefile with automation commands
5. ✅ JWT validation middleware (replaces Oathkeeper)

### Simplified docker-compose.yml

```yaml
version: '3.9'

networks:
  app-network:
    driver: bridge

services:
  # PostgreSQL Database
  postgres:
    image: postgres:16-alpine
    container_name: docmanager-postgres
    ports:
      - "15432:5432"
    environment:
      POSTGRES_DB: docmanager
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
      - ./backend/migrations:/migrations
    networks:
      - app-network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: docmanager-redis
    ports:
      - "16379:6379"
    command: redis-server --requirepass ${REDIS_PASSWORD} --maxmemory 2gb --maxmemory-policy allkeys-lru
    volumes:
      - redis-data:/data
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "redis-cli", "--raw", "incr", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # MinIO Object Storage
  minio:
    image: minio/minio:latest
    container_name: docmanager-minio
    ports:
      - "19000:9000"
      - "19001:9001"
    environment:
      MINIO_ROOT_USER: ${MINIO_ROOT_USER}
      MINIO_ROOT_PASSWORD: ${MINIO_ROOT_PASSWORD}
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3
    restart: unless-stopped

  # Meilisearch
  meilisearch:
    image: getmeili/meilisearch:v1.5
    container_name: docmanager-meilisearch
    ports:
      - "17700:7700"
    environment:
      MEILI_MASTER_KEY: ${MEILI_MASTER_KEY}
      MEILI_ENV: development
    volumes:
      - meilisearch-data:/meili_data
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:7700/health"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # PaddleOCR Service
  paddleocr:
    image: paddlepaddle/paddle:2.5.1
    container_name: docmanager-paddleocr
    ports:
      - "18080:8080"
    volumes:
      - ./backend/services/ocr-service/paddle-models:/models
    networks:
      - app-network
    restart: unless-stopped

  # NATS JetStream
  nats:
    image: nats:2.10-alpine
    container_name: docmanager-nats
    ports:
      - "14222:4222"
      - "18222:8222"
    command: ["-js", "-m", "8222"]
    volumes:
      - nats-data:/data
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8222/healthz"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # ClickHouse
  clickhouse:
    image: clickhouse/clickhouse-server:23-alpine
    container_name: docmanager-clickhouse
    ports:
      - "18123:8123"
      - "19000:9000"
    volumes:
      - clickhouse-data:/var/lib/clickhouse
    networks:
      - app-network
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8123/ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # MailSlurper (Development only)
  mailslurper:
    image: marcopas/docker-mailslurper:latest
    container_name: docmanager-mailslurper
    ports:
      - "14436:4436"  # SMTP
      - "14437:4437"  # Web UI
    networks:
      - app-network
    restart: unless-stopped

volumes:
  postgres-data:
  redis-data:
  minio-data:
  meilisearch-data:
  nats-data:
  clickhouse-data:
```

### Updated .env.example

```bash
# Database
DB_PASSWORD=your_secure_password_here

# Redis
REDIS_PASSWORD=your_redis_password_here

# MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_password_here

# Meilisearch
MEILI_MASTER_KEY=your_32_character_master_key_here

# Shared Authentication (Obtain from Infrastructure Team)
SHARED_KRATOS_PUBLIC_URL=http://shared-kratos:14433
SHARED_KRATOS_ADMIN_URL=http://shared-kratos:14434
SHARED_HYDRA_PUBLIC_URL=http://shared-hydra:14444
SHARED_HYDRA_ADMIN_URL=http://shared-hydra:14445

# OAuth2 Client (Register with Shared Hydra)
OAUTH2_CLIENT_ID=document-manager-client
OAUTH2_CLIENT_SECRET=your_client_secret_from_infra_team
OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback

# JWT Validation (for backend services)
HYDRA_JWKS_URL=http://shared-hydra:14444/.well-known/jwks.json
JWT_ISSUER=http://shared-hydra:14444
JWT_AUDIENCE=document-manager-client
```

---

## Updated Phase 2: JWT Validation Middleware

Since we removed Oathkeeper, each backend service needs to validate JWTs directly.

### JWT Middleware Package

**backend/pkg/middleware/jwt.go**:

```go
package middleware

import (
    "context"
    "crypto/rsa"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/golang-jwt/jwt/v5"
    "go.uber.org/zap"
)

type JWTMiddleware struct {
    jwksURL      string
    issuer       string
    audience     string
    publicKeys   map[string]*rsa.PublicKey
    mu           sync.RWMutex
    logger       *zap.Logger
    lastRefresh  time.Time
    refreshTTL   time.Duration
}

type JWKS struct {
    Keys []JWK `json:"keys"`
}

type JWK struct {
    Kid string   `json:"kid"`
    Kty string   `json:"kty"`
    Use string   `json:"use"`
    N   string   `json:"n"`
    E   string   `json:"e"`
}

type Claims struct {
    Subject      string   `json:"sub"`
    Email        string   `json:"email"`
    EmailVerified bool    `json:"email_verified"`
    Name         string   `json:"name"`
    Scope        string   `json:"scope"`
    jwt.RegisteredClaims
}

func NewJWTMiddleware(jwksURL, issuer, audience string, logger *zap.Logger) *JWTMiddleware {
    return &JWTMiddleware{
        jwksURL:     jwksURL,
        issuer:      issuer,
        audience:    audience,
        publicKeys:  make(map[string]*rsa.PublicKey),
        logger:      logger,
        refreshTTL:  1 * time.Hour,
    }
}

// Initialize fetches JWKS on startup
func (m *JWTMiddleware) Initialize() error {
    return m.refreshKeys()
}

// RequireAuth validates JWT token
func (m *JWTMiddleware) RequireAuth() fiber.Handler {
    return func(c *fiber.Ctx) error {
        // Extract token from Authorization header
        authHeader := c.Get("Authorization")
        if authHeader == "" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "error":   "Missing authorization header",
            })
        }

        // Extract token (format: "Bearer <token>")
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "error":   "Invalid authorization header format",
            })
        }

        tokenString := parts[1]

        // Refresh keys if needed
        if time.Since(m.lastRefresh) > m.refreshTTL {
            if err := m.refreshKeys(); err != nil {
                m.logger.Error("Failed to refresh JWKS", zap.Error(err))
            }
        }

        // Parse and validate token
        claims, err := m.validateToken(tokenString)
        if err != nil {
            m.logger.Error("Token validation failed", zap.Error(err))
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "success": false,
                "error":   "Invalid token",
            })
        }

        // Store claims in context
        c.Locals("user_id", claims.Subject)
        c.Locals("email", claims.Email)
        c.Locals("name", claims.Name)
        c.Locals("scopes", strings.Split(claims.Scope, " "))

        return c.Next()
    }
}

func (m *JWTMiddleware) validateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        // Verify signing algorithm
        if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }

        // Get key ID from token header
        kid, ok := token.Header["kid"].(string)
        if !ok {
            return nil, fmt.Errorf("missing kid in token header")
        }

        // Get public key
        m.mu.RLock()
        publicKey, exists := m.publicKeys[kid]
        m.mu.RUnlock()

        if !exists {
            // Try refreshing keys
            if err := m.refreshKeys(); err != nil {
                return nil, fmt.Errorf("failed to refresh keys: %w", err)
            }

            m.mu.RLock()
            publicKey, exists = m.publicKeys[kid]
            m.mu.RUnlock()

            if !exists {
                return nil, fmt.Errorf("public key not found for kid: %s", kid)
            }
        }

        return publicKey, nil
    })

    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    // Validate issuer
    if claims.Issuer != m.issuer {
        return nil, fmt.Errorf("invalid issuer: %s", claims.Issuer)
    }

    // Validate audience
    validAudience := false
    for _, aud := range claims.Audience {
        if aud == m.audience {
            validAudience = true
            break
        }
    }
    if !validAudience {
        return nil, fmt.Errorf("invalid audience")
    }

    return claims, nil
}

func (m *JWTMiddleware) refreshKeys() error {
    resp, err := http.Get(m.jwksURL)
    if err != nil {
        return fmt.Errorf("failed to fetch JWKS: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("JWKS endpoint returned status: %d", resp.StatusCode)
    }

    var jwks JWKS
    if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
        return fmt.Errorf("failed to decode JWKS: %w", err)
    }

    newKeys := make(map[string]*rsa.PublicKey)
    for _, key := range jwks.Keys {
        if key.Kty != "RSA" || key.Use != "sig" {
            continue
        }

        publicKey, err := parseRSAPublicKey(key.N, key.E)
        if err != nil {
            m.logger.Warn("Failed to parse public key", zap.String("kid", key.Kid), zap.Error(err))
            continue
        }

        newKeys[key.Kid] = publicKey
    }

    if len(newKeys) == 0 {
        return fmt.Errorf("no valid keys found in JWKS")
    }

    m.mu.Lock()
    m.publicKeys = newKeys
    m.lastRefresh = time.Now()
    m.mu.Unlock()

    m.logger.Info("JWKS refreshed", zap.Int("key_count", len(newKeys)))
    return nil
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
    // Implementation to parse base64url encoded N and E
    // This is simplified - you'll need proper base64url decoding
    // and big.Int conversion
    // Use a JWT library helper or implement according to JWT spec
    return nil, fmt.Errorf("not implemented - use jwt library helper")
}

// Helper to get user ID from context
func GetUserID(c *fiber.Ctx) (string, error) {
    userID, ok := c.Locals("user_id").(string)
    if !ok {
        return "", fmt.Errorf("user_id not found in context")
    }
    return userID, nil
}

// Helper to get email from context
func GetEmail(c *fiber.Ctx) (string, error) {
    email, ok := c.Locals("email").(string)
    if !ok {
        return "", fmt.Errorf("email not found in context")
    }
    return email, nil
}

// Helper to check if user has specific scope
func HasScope(c *fiber.Ctx, scope string) bool {
    scopes, ok := c.Locals("scopes").([]string)
    if !ok {
        return false
    }
    for _, s := range scopes {
        if s == scope {
            return true
        }
    }
    return false
}
```

### Using JWT Middleware in Services

**backend/services/tenant-service/cmd/main.go**:

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/recover"
    "github.com/yourusername/document-manager/pkg/config"
    "github.com/yourusername/document-manager/pkg/logger"
    "github.com/yourusername/document-manager/pkg/middleware"
    "go.uber.org/zap"
)

func main() {
    cfg := config.Load()
    log := logger.New("debug", "text")
    defer log.Sync()

    // Initialize JWT middleware
    jwtMiddleware := middleware.NewJWTMiddleware(
        cfg.Auth.HydraJWKSURL,
        cfg.Auth.JWTIssuer,
        cfg.Auth.JWTAudience,
        log,
    )

    // Fetch JWKS on startup
    if err := jwtMiddleware.Initialize(); err != nil {
        log.Fatal("Failed to initialize JWT middleware", zap.Error(err))
    }

    app := fiber.New()

    // Global middleware
    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "http://localhost:13000,http://localhost:13001",
        AllowMethods:     "GET,POST,PUT,DELETE,PATCH",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
    }))

    // Public routes
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "healthy"})
    })

    // Protected routes
    api := app.Group("/api/v1", jwtMiddleware.RequireAuth())

    api.Get("/tenants/:id", func(c *fiber.Ctx) error {
        userID, _ := middleware.GetUserID(c)
        email, _ := middleware.GetEmail(c)

        return c.JSON(fiber.Map{
            "success": true,
            "data": fiber.Map{
                "id":      c.Params("id"),
                "user_id": userID,
                "email":   email,
            },
        })
    })

    go func() {
        if err := app.Listen(":10001"); err != nil {
            log.Fatal("Server failed", zap.Error(err))
        }
    }()

    log.Info("Tenant Service started on :10001")

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down...")
    app.ShutdownWithTimeout(30 * time.Second)
}
```

---

## Updated Implementation Phases Summary

### Phase 1: Infrastructure Setup (Week 1-2)
- **8 services** in Docker Compose (no Prometheus/Oathkeeper/Grafana)
- Database migrations
- Makefile
- **NEW**: JWT validation middleware package

### Phase 2: Shared Backend Packages (Week 2-3)
- All packages as before
- **NEW**: JWT middleware with JWKS refresh
- Helper functions for extracting user context

### Phase 3-8: Same as Original Plan
All other phases remain the same, but:
- Each service validates JWTs directly
- No Oathkeeper configuration needed
- No Prometheus metrics (can add later)
- Frontend calls backend services directly

---

## Benefits of Simplified Approach

✅ **Faster Setup**: 8 services instead of 15+
✅ **Simpler Architecture**: No API gateway layer
✅ **Direct Communication**: Frontend → Backend services
✅ **Easier Debugging**: Fewer moving parts
✅ **Lower Resource Usage**: Fewer containers running

## What We Can Add Later

📊 **Monitoring**: Add Prometheus + Grafana when needed
🛡️ **API Gateway**: Add Oathkeeper/Kong if centralized policies needed
📈 **Distributed Tracing**: Add Jaeger/Zipkin
🔍 **Service Mesh**: Add Istio/Linkerd for advanced scenarios

---

Would you like me to create the complete simplified docker-compose.yml file and updated Makefile?