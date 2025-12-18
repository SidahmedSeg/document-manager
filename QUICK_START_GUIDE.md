# Quick Start Guide - Document Manager
**Get Up and Running in 30 Minutes**

---

## Prerequisites

### Required Software
```bash
# Check versions
go version          # Go 1.22+
node --version      # Node 20+
docker --version    # Docker 24+
docker-compose --version  # Docker Compose 2.x

# Install if missing
brew install go node docker docker-compose  # macOS
apt install golang nodejs docker.io         # Ubuntu
```

### Required Knowledge
- Go basics (or willingness to learn)
- React/Next.js fundamentals
- Docker Compose basics
- REST API concepts
- SQL basics

---

## Phase 1: Project Setup (10 minutes)

### Step 1: Create Project Structure

```bash
# Create root directory
mkdir -p document-manager && cd document-manager

# Create backend structure
mkdir -p backend/{pkg,services,migrations}
mkdir -p backend/pkg/{config,database,cache,logger,middleware,validator,response,errors,metrics}
mkdir -p backend/services/{tenant-service,document-service,storage-service}
mkdir -p backend/services/{share-service,rbac-service,quota-service}
mkdir -p backend/services/{ocr-service,categorization-service,search-service}
mkdir -p backend/services/{notification-service,audit-service}

# Create each service structure (example for tenant-service)
for service in tenant-service document-service storage-service share-service rbac-service quota-service ocr-service categorization-service search-service notification-service audit-service; do
    mkdir -p backend/services/$service/{cmd,internal/{models,repository,service,handler}}
done

# Create frontend structure
mkdir -p frontend/{user-app,admin-app}

# Create config directories
mkdir -p config/{oathkeeper,prometheus,grafana/{datasources,dashboards}}

# Create other directories
mkdir -p {docs,scripts}

# Initialize Git
git init
echo "# Document Manager - Multi-Tenant Document Management System" > README.md
```

### Step 2: Create Core Configuration Files

**docker-compose.yml** (Copy from Specs.txt lines 283-508)

**Makefile**:
```makefile
.PHONY: help up down restart logs db-migrate db-rollback test

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

restart: ## Restart all services
	docker-compose restart

logs: ## View logs (usage: make logs service=tenant)
	@if [ -z "$(service)" ]; then \
		docker-compose logs -f; \
	else \
		docker-compose logs -f $(service); \
	fi

ps: ## Show running services
	docker-compose ps

db-migrate: ## Run database migrations
	@echo "Running migrations..."
	docker exec -it docmanager-postgres psql -U postgres -d docmanager -f /migrations/all.sql

db-rollback: ## Rollback last migration
	@echo "Rolling back last migration..."

db-reset: ## Reset database (WARNING: deletes all data)
	docker-compose down -v
	docker-compose up -d postgres
	sleep 5
	$(MAKE) db-migrate

test-shared-auth: ## Test connectivity to shared Kratos/Hydra
	@echo "Testing Kratos..."
	curl -f http://127.0.0.1:4433/health/ready || echo "Kratos not available"
	@echo "Testing Hydra..."
	curl -f http://127.0.0.1:4444/.well-known/jwks.json || echo "Hydra not available"

test: ## Run all tests
	@echo "Running backend tests..."
	cd backend && go test ./...
	@echo "Running frontend tests..."
	cd frontend/user-app && npm test

build-services: ## Build all backend services
	@echo "Building services..."
	@for service in tenant document storage share rbac quota ocr categorization search notification audit; do \
		echo "Building $$service-service..."; \
		cd backend/services/$$service-service && go build -o bin/$$service-service cmd/main.go; \
	done

run-tenant: ## Run tenant service
	cd backend/services/tenant-service && go run cmd/main.go

clean: ## Clean build artifacts
	find backend/services -name "bin" -type d -exec rm -rf {} +
	docker-compose down -v
```

**.env.example**:
```bash
# Database
DB_HOST=localhost
DB_PORT=15432
DB_USER=postgres
DB_PASSWORD=your_secure_password_here
DB_NAME=docmanager
DB_SSL_MODE=disable

# Redis
REDIS_HOST=localhost
REDIS_PORT=16379
REDIS_PASSWORD=your_redis_password_here
REDIS_DB=0

# MinIO
MINIO_ENDPOINT=localhost:19000
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=your_minio_password_here
MINIO_USE_SSL=false

# Meilisearch
MEILI_HOST=localhost:17700
MEILI_MASTER_KEY=your_32_character_master_key_here

# NATS
NATS_URL=nats://localhost:14222

# ClickHouse
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=18123
CLICKHOUSE_DB=docmanager

# Shared Authentication (Obtain from Infrastructure Team)
SHARED_KRATOS_PUBLIC_URL=http://shared-kratos:14433
SHARED_KRATOS_ADMIN_URL=http://shared-kratos:14434
SHARED_HYDRA_PUBLIC_URL=http://shared-hydra:14444
SHARED_HYDRA_ADMIN_URL=http://shared-hydra:14445

# OAuth2 Client (Register with Shared Hydra)
OAUTH2_CLIENT_ID=document-manager-client
OAUTH2_CLIENT_SECRET=your_client_secret_from_infra_team
OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback
OAUTH2_SCOPES=openid email profile offline_access

# Oathkeeper
OATHKEEPER_PROXY_URL=http://localhost:14455
OATHKEEPER_API_URL=http://localhost:14456

# Services
TENANT_SERVICE_URL=http://localhost:10001
DOCUMENT_SERVICE_URL=http://localhost:10002
STORAGE_SERVICE_URL=http://localhost:10003
SHARE_SERVICE_URL=http://localhost:10004
RBAC_SERVICE_URL=http://localhost:10005
QUOTA_SERVICE_URL=http://localhost:10006
OCR_SERVICE_URL=http://localhost:10007
CATEGORIZATION_SERVICE_URL=http://localhost:10008
SEARCH_SERVICE_URL=http://localhost:10009
NOTIFICATION_SERVICE_URL=http://localhost:10010
AUDIT_SERVICE_URL=http://localhost:10011

# Monitoring
PROMETHEUS_URL=http://localhost:19090
GRAFANA_URL=http://localhost:13002
GRAFANA_ADMIN_PASSWORD=your_grafana_password

# Application
LOG_LEVEL=debug
LOG_FORMAT=json
ENVIRONMENT=development
```

**.gitignore**:
```
# Environment
.env
.env.local

# Build artifacts
bin/
dist/
build/
*.exe

# Dependencies
node_modules/
vendor/

# IDE
.vscode/
.idea/
*.swp
*.swo

# Logs
*.log

# OS
.DS_Store
Thumbs.db

# Database
*.db
*.sqlite

# Docker
docker-compose.override.yml

# Certificates
*.pem
*.key
*.crt

# Temporary
tmp/
temp/
```

---

## Phase 2: Backend Setup (10 minutes)

### Step 1: Initialize Go Modules

```bash
cd backend

# Create go.mod
cat > go.mod << 'EOF'
module github.com/yourusername/document-manager

go 1.22

require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/jackc/pgx/v5 v5.5.1
    github.com/redis/go-redis/v9 v9.4.0
    github.com/minio/minio-go/v7 v7.0.66
    github.com/nats-io/nats.go v1.31.0
    github.com/meilisearch/meilisearch-go v0.26.2
    github.com/spf13/viper v1.18.2
    go.uber.org/zap v1.26.0
    github.com/go-playground/validator/v10 v10.16.0
    github.com/google/uuid v1.5.0
    github.com/prometheus/client_golang v1.18.0
)
EOF

# Download dependencies
go mod download
go mod tidy
```

### Step 2: Create Shared Config Package

**backend/pkg/config/config.go**:
```go
package config

import (
    "github.com/spf13/viper"
    "log"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    MinIO    MinIOConfig
    Auth     AuthConfig
}

type ServerConfig struct {
    Host string
    Port string
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type MinIOConfig struct {
    Endpoint  string
    AccessKey string
    SecretKey string
    UseSSL    bool
}

type AuthConfig struct {
    KratosPublicURL string
    KratosAdminURL  string
    HydraPublicURL  string
    OAuthClientID   string
    OAuthClientSecret string
}

func Load() *Config {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    if err := viper.ReadInConfig(); err != nil {
        log.Printf("Warning: .env file not found: %v", err)
    }

    return &Config{
        Server: ServerConfig{
            Host: viper.GetString("SERVER_HOST"),
            Port: viper.GetString("SERVER_PORT"),
        },
        Database: DatabaseConfig{
            Host:     viper.GetString("DB_HOST"),
            Port:     viper.GetString("DB_PORT"),
            User:     viper.GetString("DB_USER"),
            Password: viper.GetString("DB_PASSWORD"),
            DBName:   viper.GetString("DB_NAME"),
            SSLMode:  viper.GetString("DB_SSL_MODE"),
        },
        Redis: RedisConfig{
            Host:     viper.GetString("REDIS_HOST"),
            Port:     viper.GetString("REDIS_PORT"),
            Password: viper.GetString("REDIS_PASSWORD"),
            DB:       viper.GetInt("REDIS_DB"),
        },
        MinIO: MinIOConfig{
            Endpoint:  viper.GetString("MINIO_ENDPOINT"),
            AccessKey: viper.GetString("MINIO_ROOT_USER"),
            SecretKey: viper.GetString("MINIO_ROOT_PASSWORD"),
            UseSSL:    viper.GetBool("MINIO_USE_SSL"),
        },
        Auth: AuthConfig{
            KratosPublicURL:   viper.GetString("SHARED_KRATOS_PUBLIC_URL"),
            KratosAdminURL:    viper.GetString("SHARED_KRATOS_ADMIN_URL"),
            HydraPublicURL:    viper.GetString("SHARED_HYDRA_PUBLIC_URL"),
            OAuthClientID:     viper.GetString("OAUTH2_CLIENT_ID"),
            OAuthClientSecret: viper.GetString("OAUTH2_CLIENT_SECRET"),
        },
    }
}
```

### Step 3: Create Logger Package

**backend/pkg/logger/logger.go**:
```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New(level, format string) *zap.Logger {
    var config zap.Config

    if format == "json" {
        config = zap.NewProductionConfig()
    } else {
        config = zap.NewDevelopmentConfig()
    }

    // Set log level
    switch level {
    case "debug":
        config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    case "info":
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    case "warn":
        config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
    case "error":
        config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
    default:
        config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
    }

    logger, err := config.Build()
    if err != nil {
        panic(err)
    }

    return logger
}
```

### Step 4: Create First Service (Tenant Service)

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
    "go.uber.org/zap"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize logger
    log := logger.New("debug", "text")
    defer log.Sync()

    log.Info("Starting Tenant Service", zap.String("port", "10001"))

    // Create Fiber app
    app := fiber.New(fiber.Config{
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
        ErrorHandler: func(c *fiber.Ctx, err error) error {
            code := fiber.StatusInternalServerError
            if e, ok := err.(*fiber.Error); ok {
                code = e.Code
            }
            return c.Status(code).JSON(fiber.Map{
                "success": false,
                "error":   err.Error(),
            })
        },
    })

    // Middleware
    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins:     "http://localhost:13000,http://localhost:13001",
        AllowMethods:     "GET,POST,PUT,DELETE,PATCH",
        AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
        AllowCredentials: true,
    }))

    // Routes
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "status":  "healthy",
            "service": "tenant-service",
            "version": "1.0.0",
        })
    })

    api := app.Group("/api/v1")

    api.Get("/tenants/:id", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "success": true,
            "data": fiber.Map{
                "id":   c.Params("id"),
                "name": "Example Tenant",
            },
        })
    })

    // Start server
    go func() {
        if err := app.Listen(":10001"); err != nil {
            log.Fatal("Server failed", zap.Error(err))
        }
    }()

    log.Info("Tenant Service started successfully on :10001")

    // Graceful shutdown
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
    <-quit

    log.Info("Shutting down server...")
    if err := app.ShutdownWithTimeout(30 * time.Second); err != nil {
        log.Error("Server forced to shutdown", zap.Error(err))
    }

    log.Info("Server stopped")
}
```

---

## Phase 3: Frontend Setup (10 minutes)

### Step 1: Initialize Next.js User App

```bash
cd frontend

# Create Next.js app
npx create-next-app@latest user-app \
    --typescript \
    --tailwind \
    --app \
    --no-src-dir \
    --import-alias "@/*"

cd user-app

# Install dependencies
npm install @ory/client axios @tanstack/react-query zustand zod react-hook-form
npm install @radix-ui/react-dialog @radix-ui/react-dropdown-menu
npm install @radix-ui/react-select @radix-ui/react-toast
npm install class-variance-authority clsx tailwind-merge lucide-react

# Install shadcn/ui
npx shadcn-ui@latest init

# Add shadcn components
npx shadcn-ui@latest add button
npx shadcn-ui@latest add dialog
npx shadcn-ui@latest add dropdown-menu
npx shadcn-ui@latest add input
npx shadcn-ui@latest add select
npx shadcn-ui@latest add toast
```

### Step 2: Create Environment File

**frontend/user-app/.env.local**:
```bash
NEXT_PUBLIC_API_URL=http://localhost:14455/api/v1
NEXT_PUBLIC_HYDRA_PUBLIC_URL=http://127.0.0.1:4444
NEXT_PUBLIC_OAUTH2_CLIENT_ID=document-manager-client
NEXT_PUBLIC_OAUTH2_CLIENT_SECRET=your_client_secret
NEXT_PUBLIC_OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback
```

### Step 3: Create API Client

**frontend/user-app/lib/api-client.ts**:
```typescript
import axios from 'axios'

export const apiClient = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:14455/api/v1',
    withCredentials: true,
    headers: {
        'Content-Type': 'application/json',
    },
})

// Add request interceptor for JWT
apiClient.interceptors.request.use((config) => {
    const token = getCookie('access_token')
    if (token) {
        config.headers.Authorization = `Bearer ${token}`
    }
    return config
})

// Add response interceptor for error handling
apiClient.interceptors.response.use(
    (response) => response,
    (error) => {
        if (error.response?.status === 401) {
            // Redirect to login
            window.location.href = '/'
        }
        return Promise.reject(error)
    }
)

function getCookie(name: string): string | undefined {
    const value = `; ${document.cookie}`
    const parts = value.split(`; ${name}=`)
    if (parts.length === 2) return parts.pop()?.split(';').shift()
}
```

### Step 4: Create OAuth2 Client

**frontend/user-app/lib/oauth-client.ts**:
```typescript
export async function redirectToLogin() {
    const authorizeUrl = new URL(`${process.env.NEXT_PUBLIC_HYDRA_PUBLIC_URL}/oauth2/auth`)
    authorizeUrl.searchParams.set('client_id', process.env.NEXT_PUBLIC_OAUTH2_CLIENT_ID!)
    authorizeUrl.searchParams.set('response_type', 'code')
    authorizeUrl.searchParams.set('redirect_uri', process.env.NEXT_PUBLIC_OAUTH2_REDIRECT_URI!)
    authorizeUrl.searchParams.set('scope', 'openid email profile offline_access')

    window.location.href = authorizeUrl.toString()
}

export async function exchangeCodeForToken(code: string) {
    const response = await fetch(`${process.env.NEXT_PUBLIC_HYDRA_PUBLIC_URL}/oauth2/token`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
            grant_type: 'authorization_code',
            code,
            redirect_uri: process.env.NEXT_PUBLIC_OAUTH2_REDIRECT_URI!,
            client_id: process.env.NEXT_PUBLIC_OAUTH2_CLIENT_ID!,
            client_secret: process.env.NEXT_PUBLIC_OAUTH2_CLIENT_SECRET!,
        }),
    })

    if (!response.ok) {
        throw new Error('Token exchange failed')
    }

    return response.json()
}
```

---

## Testing Your Setup

### 1. Start Infrastructure

```bash
# Copy .env.example to .env
cp .env.example .env

# Edit .env with your values
nano .env

# Start all services
make up

# Wait for services to be healthy (30-60 seconds)
make ps

# Check logs
make logs
```

### 2. Run Database Migrations

```bash
# Create migrations file (TODO: copy from Specs.txt)
make db-migrate

# Verify tables created
docker exec -it docmanager-postgres psql -U postgres -d docmanager -c "\dt"
```

### 3. Start Tenant Service

```bash
# In terminal 1
cd backend
cp .env.example .env
make run-tenant

# Should see: "Tenant Service started successfully on :10001"
```

### 4. Test Tenant Service

```bash
# In terminal 2
curl http://localhost:10001/health

# Should return:
# {"status":"healthy","service":"tenant-service","version":"1.0.0"}
```

### 5. Start Frontend

```bash
# In terminal 3
cd frontend/user-app
npm run dev

# Should start on http://localhost:13000
```

### 6. Access Application

Open browser to: http://localhost:13000

You should see the Next.js welcome page. If authentication is configured, it will redirect to login.

---

## Next Steps

### Immediate Tasks (Today)
1. ✅ Review IMPLEMENTATION_PLAN.md
2. ✅ Review ARCHITECTURE_DECISIONS.md
3. ✅ Complete project setup above
4. ✅ Start infrastructure with `make up`
5. ✅ Test tenant service health endpoint

### This Week
1. Complete all shared backend packages (pkg/*)
2. Implement tenant service fully
3. Create database migrations
4. Setup Oathkeeper configuration
5. Coordinate with infrastructure team for OAuth2 client registration

### Next Week
1. Implement document service
2. Implement storage service
3. Create frontend OAuth2 integration
4. Build basic document list UI

---

## Common Issues and Solutions

### Issue: Docker services won't start
```bash
# Check Docker daemon
docker info

# Check port conflicts
lsof -i :15432  # PostgreSQL
lsof -i :16379  # Redis

# Clean up and restart
make clean
make up
```

### Issue: Go dependencies not downloading
```bash
# Clear module cache
go clean -modcache

# Verify GOPROXY
go env GOPROXY

# Re-download
go mod download
```

### Issue: Cannot connect to database
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
docker logs docmanager-postgres

# Test connection
docker exec -it docmanager-postgres psql -U postgres -c "SELECT version();"
```

### Issue: Port already in use
```bash
# Find process using port
lsof -i :10001

# Kill process
kill -9 <PID>

# Or change port in .env
```

---

## Development Workflow

### Daily Development

```bash
# Morning: Start infrastructure
make up
make logs  # Check all services healthy

# Start backend service you're working on
cd backend/services/tenant-service
go run cmd/main.go

# Start frontend (separate terminal)
cd frontend/user-app
npm run dev

# Make changes and test

# Evening: Stop services
make down
```

### Testing Changes

```bash
# Run backend tests
cd backend
go test ./...

# Run specific service tests
cd backend/services/tenant-service
go test ./... -v

# Run frontend tests
cd frontend/user-app
npm test

# Run E2E tests
npm run test:e2e
```

### Debugging

```bash
# View service logs
make logs service=postgres
make logs service=redis
make logs service=tenant

# Check service health
curl http://localhost:10001/health
curl http://localhost:10002/health

# PostgreSQL shell
docker exec -it docmanager-postgres psql -U postgres -d docmanager

# Redis shell
docker exec -it docmanager-redis redis-cli -a your_password
```

---

## Resources

**Project Documentation**:
- [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) - Complete implementation guide
- [ARCHITECTURE_DECISIONS.md](./ARCHITECTURE_DECISIONS.md) - Architectural choices
- [Specs.txt](./Specs.txt) - Original specifications

**Ory Authentication**:
- [ORY_API_SPEC.md](./ORY_API_SPEC.md) - API specifications
- [ORY_INTEGRATION_GUIDE.md](./ORY_INTEGRATION_GUIDE.md) - Integration guide
- [ORY_QUICK_REFERENCE.md](./ORY_QUICK_REFERENCE.md) - Quick reference

**External Links**:
- Go Fiber: https://docs.gofiber.io
- Next.js: https://nextjs.org/docs
- PostgreSQL: https://www.postgresql.org/docs
- Ory Docs: https://www.ory.sh/docs

---

## Get Help

**If you're stuck:**
1. Check logs: `make logs`
2. Review documentation above
3. Search GitHub issues
4. Ask the team

**Contact**:
- Backend questions: Backend team lead
- Frontend questions: Frontend team lead
- Infrastructure: DevOps team
- Authentication: Infrastructure team (for OAuth2 client)

---

**You're now ready to start building! 🚀**

Focus on completing Phase 1 (Infrastructure) this week, then move to Phase 2 (Shared Packages) next week.

Good luck!