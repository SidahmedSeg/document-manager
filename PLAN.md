# Document Manager - Implementation Plan

## Project Overview

A production-ready, multi-tenant document management system with OCR, intelligent categorization, full-text search, and enterprise-grade security.

**Architecture:** Microservices (11 services) + API Gateway (Ory Oathkeeper) + Shared Auth (Kratos/Hydra)

**Tech Stack:**
- **Backend:** Go 1.21+ (11 microservices)
- **Frontend:** Next.js 14+ with TypeScript & shadcn/ui
- **API Gateway:** Ory Oathkeeper
- **Auth:** Ory Kratos (Identity) + Ory Hydra (OAuth2/OIDC)
- **Database:** PostgreSQL 16
- **Cache:** Redis 7
- **Storage:** MinIO (S3-compatible)
- **Search:** Meilisearch
- **Message Queue:** NATS JetStream
- **Analytics:** ClickHouse
- **OCR:** PaddleOCR

---

## Phase 1: Infrastructure Setup ✅ COMPLETED

### 1.1 Docker Compose Infrastructure ✅
- [x] 9 infrastructure services configured
- [x] Ory Oathkeeper API Gateway (ports 14455, 14456)
- [x] PostgreSQL 16 with optimized settings (port 15432)
- [x] Redis 7 with persistence (port 16379)
- [x] MinIO S3-compatible storage (ports 19000, 19001)
- [x] Meilisearch full-text search (port 17700)
- [x] NATS JetStream message queue (ports 14222, 18222)
- [x] ClickHouse analytics database (ports 18123, 19009)
- [x] PaddleOCR placeholder (port 18080)
- [x] MailSlurper email testing (ports 14436-14439)

### 1.2 Oathkeeper Configuration ✅
- [x] Production config with JWT validation (oathkeeper.yml)
- [x] Test config for infrastructure testing (oathkeeper-test.yml)
- [x] Access rules for all 11 microservices
- [x] CORS configuration for frontend apps
- [x] Health check endpoints

### 1.3 Database Schema ✅
- [x] 10 database migrations (20 files total)
- [x] 33 tables created with full multi-tenant schema
- [x] Extensions: uuid-ossp, pgcrypto, pg_trgm
- [x] Custom types: user_status, document_status, share_permission, etc.
- [x] Comprehensive indexes for performance
- [x] Full-text search vectors
- [x] Triggers for auto-updates
- [x] Seed data: 3 roles, 3 subscription plans

**Migration Files:**
1. `000001_create_extensions_and_types.up.sql` - PostgreSQL extensions & custom types
2. `000002_create_tenants_and_users.up.sql` - Multi-tenancy & user management
3. `000003_create_rbac.up.sql` - Role-Based Access Control
4. `000004_create_documents.up.sql` - Document management core
5. `000005_create_sharing.up.sql` - Document sharing & collaboration
6. `000006_create_ocr_and_quota.up.sql` - OCR processing & quota tracking
7. `000007_create_activity_logs.up.sql` - Audit logs & analytics
8. `000008_create_plans_and_config.up.sql` - Subscription plans & settings
9. `000009_seed_data.up.sql` - Default roles & plans
10. `000010_create_triggers.up.sql` - Database triggers & functions

### 1.4 Environment Configuration ✅
- [x] Comprehensive .env file with 350+ variables
- [x] Service connection strings
- [x] Authentication URLs (Kratos/Hydra)
- [x] Security settings (rate limits, CORS, CSP)
- [x] Feature flags
- [x] Plan configurations

### 1.5 Automation & Tools ✅
- [x] Makefile with 20+ commands
- [x] Service lifecycle management (up, down, restart)
- [x] Health checks
- [x] Database migrations with golang-migrate
- [x] Log viewing shortcuts
- [x] Cleanup commands

### Infrastructure Status
```
✅ All services healthy
✅ Database schema v10 applied
✅ Seed data loaded
✅ Ready for backend development
```

---

## Phase 2: Shared Backend Packages (Go) 🔄 NEXT

### Goals
Create reusable packages for all microservices to ensure consistency and reduce duplication.

### 2.1 Core Configuration (pkg/config)
**File:** `backend/pkg/config/config.go`

**Features:**
- Viper-based configuration loading
- Environment variable support
- Configuration validation
- Hot reload capability
- Structured config types

**Config Structure:**
```go
type Config struct {
    Environment string
    Server      ServerConfig
    Database    DatabaseConfig
    Redis       RedisConfig
    Auth        AuthConfig
    Services    map[string]string
}
```

### 2.2 Database Package (pkg/database)
**File:** `backend/pkg/database/postgres.go`

**Features:**
- PostgreSQL connection pool management
- Auto-reconnect with exponential backoff
- Query timeout handling
- Health check function
- Transaction helpers
- Multi-tenancy helpers (SetTenantID)

**API:**
```go
func NewPostgresDB(config DatabaseConfig) (*sql.DB, error)
func HealthCheck(db *sql.DB) error
func WithTenant(ctx context.Context, tenantID string) context.Context
```

### 2.3 Cache Package (pkg/cache)
**File:** `backend/pkg/cache/redis.go`

**Features:**
- Redis client wrapper
- Connection pooling
- Automatic JSON serialization
- TTL management
- Pub/Sub helpers
- Session store integration

**API:**
```go
func NewRedisClient(config RedisConfig) (*redis.Client, error)
func Get(key string, dest interface{}) error
func Set(key string, value interface{}, ttl time.Duration) error
func Delete(key string) error
```

### 2.4 Logging Package (pkg/logger)
**File:** `backend/pkg/logger/logger.go`

**Features:**
- Zap structured logging
- Log levels (debug, info, warn, error)
- Request ID correlation
- Tenant ID in logs
- JSON format for production
- Pretty format for development

**API:**
```go
func NewLogger(env string) *zap.Logger
func WithRequestID(ctx context.Context, requestID string) context.Context
func WithTenant(ctx context.Context, tenantID string) context.Context
```

### 2.5 Middleware Package (pkg/middleware)
**File:** `backend/pkg/middleware/headers.go`

**Features:**
- Extract Oathkeeper headers (X-User-ID, X-User-Email)
- Request ID generation
- Tenant context injection
- Logging middleware
- Recovery middleware
- CORS middleware

**Headers Extracted:**
- `X-User-ID` → user identity from JWT
- `X-User-Email` → user email
- `X-Tenant-ID` → tenant context
- `X-Request-ID` → request correlation

### 2.6 Validation Package (pkg/validator)
**File:** `backend/pkg/validator/validator.go`

**Features:**
- go-playground/validator integration
- Custom validation rules
- UUID validation
- Email validation
- File type validation
- Size limit validation

### 2.7 Response Package (pkg/response)
**File:** `backend/pkg/response/response.go`

**Features:**
- Standardized JSON responses
- Error response formatting
- Pagination helpers
- Success/error wrappers

**Response Format:**
```json
{
  "success": true,
  "data": {},
  "error": null,
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

### 2.8 Error Package (pkg/errors)
**File:** `backend/pkg/errors/errors.go`

**Features:**
- Custom error types
- HTTP status code mapping
- Error codes for frontend
- Stack trace support

**Error Types:**
- `ErrNotFound` → 404
- `ErrUnauthorized` → 401
- `ErrForbidden` → 403
- `ErrValidation` → 400
- `ErrInternal` → 500

---

## Phase 3: Core Backend Services 📋 PLANNED

### Service Architecture
All services follow the same structure:
```
backend/services/{service-name}/
├── cmd/
│   └── main.go              # Entry point
├── internal/
│   ├── handler/             # HTTP handlers
│   ├── service/             # Business logic
│   ├── repository/          # Database access
│   └── models/              # Data models
├── Dockerfile
└── README.md
```

### 3.1 Tenant Service (Port 10001)
**Responsibility:** Multi-tenant management

**Endpoints:**
- `POST /api/tenants` - Create tenant
- `GET /api/tenants/:id` - Get tenant details
- `PUT /api/tenants/:id` - Update tenant
- `GET /api/tenants/:id/users` - List tenant users
- `POST /api/tenants/:id/users/invite` - Invite user to tenant
- `DELETE /api/tenants/:id/users/:userId` - Remove user from tenant

**Tables:** tenants, tenant_users, tenant_invitations, tenant_settings

### 3.2 Document Service (Port 10002)
**Responsibility:** Core document management

**Endpoints:**
- `POST /api/documents` - Upload document
- `GET /api/documents` - List documents (with filters)
- `GET /api/documents/:id` - Get document details
- `PUT /api/documents/:id` - Update document metadata
- `DELETE /api/documents/:id` - Delete document
- `POST /api/documents/:id/versions` - Create version
- `GET /api/documents/:id/versions` - List versions
- `POST /api/folders` - Create folder
- `GET /api/folders` - List folders
- `PUT /api/folders/:id` - Update folder
- `DELETE /api/folders/:id` - Delete folder

**Tables:** documents, document_versions, folders, tags, document_tags, categories

### 3.3 Storage Service (Port 10003)
**Responsibility:** File storage in MinIO

**Endpoints:**
- `POST /api/storage/upload` - Upload file to MinIO
- `GET /api/storage/download/:objectId` - Download file
- `DELETE /api/storage/:objectId` - Delete file
- `POST /api/storage/presigned-url` - Generate presigned URL
- `GET /api/storage/thumbnail/:objectId` - Get thumbnail

**Integration:** MinIO S3 API, NATS for async processing

### 3.4 RBAC Service (Port 10005)
**Responsibility:** Role-Based Access Control

**Endpoints:**
- `GET /api/rbac/roles` - List roles
- `POST /api/rbac/roles` - Create custom role
- `GET /api/rbac/permissions` - List all permissions
- `POST /api/rbac/users/:userId/roles` - Assign role to user
- `GET /api/rbac/users/:userId/permissions` - Get user permissions
- `POST /api/rbac/check` - Check permission

**Tables:** roles, permissions, role_permissions, user_roles, resource_permissions

---

## Phase 4: Advanced Backend Services 📋 PLANNED

### 4.1 Share Service (Port 10004)
**Endpoints:**
- `POST /api/shares/documents/:id` - Share document with users
- `POST /api/shares/folders/:id` - Share folder
- `POST /api/shares/public-link` - Create public share link
- `GET /api/shares/with-me` - List documents shared with me
- `DELETE /api/shares/:shareId` - Revoke share

**Tables:** shared_documents, shared_folders, public_share_links, share_link_access_log

### 4.2 Quota Service (Port 10006)
**Endpoints:**
- `GET /api/quota/usage` - Get current usage
- `POST /api/quota/check` - Check if action allowed
- `GET /api/quota/history` - Get usage history

**Tables:** quota_usage, quota_history

### 4.3 OCR Service (Port 10007)
**Endpoints:**
- `POST /api/ocr/jobs` - Submit OCR job
- `GET /api/ocr/jobs/:id` - Get job status
- `GET /api/ocr/jobs/:id/result` - Get OCR text

**Integration:** PaddleOCR, NATS for async processing

**Tables:** ocr_jobs

### 4.4 Categorization Service (Port 10008)
**Endpoints:**
- `POST /api/categorize/auto` - Auto-categorize document
- `POST /api/categorize/manual` - Manually categorize
- `GET /api/categories` - List categories

**Tables:** categories

### 4.5 Search Service (Port 10009)
**Endpoints:**
- `POST /api/search` - Full-text search
- `POST /api/search/index` - Index document
- `DELETE /api/search/index/:id` - Remove from index

**Integration:** Meilisearch

### 4.6 Notification Service (Port 10010)
**Endpoints:**
- `GET /api/notifications` - List notifications
- `PUT /api/notifications/:id/read` - Mark as read
- `GET /api/notifications/preferences` - Get preferences
- `PUT /api/notifications/preferences` - Update preferences

**Tables:** notifications, notification_preferences

### 4.7 Audit Service (Port 10011)
**Endpoints:**
- `POST /api/audit/log` - Create audit log
- `GET /api/audit/logs` - Query audit logs
- `GET /api/analytics/dashboard` - Get analytics

**Integration:** ClickHouse for fast analytics

**Tables:** activity_logs (PostgreSQL), usage_events (ClickHouse)

---

## Phase 5: Frontend - User App 📋 PLANNED

### Technology Stack
- Next.js 14+ (App Router)
- TypeScript
- shadcn/ui (Radix UI + Tailwind CSS)
- React Query (TanStack Query)
- Zustand (State Management)
- Zod (Validation)

### 5.1 Project Setup
- [x] Initialize Next.js with TypeScript
- [x] Configure shadcn/ui
- [x] Setup Tailwind CSS
- [x] Configure path aliases (@/components, @/lib)

### 5.2 Authentication (OAuth2 + Hydra)
**Flow:**
1. User clicks "Login"
2. Redirect to Hydra OAuth2 flow
3. User authenticates with Kratos
4. Hydra issues OAuth2 token
5. Frontend stores token
6. All API calls include token
7. Oathkeeper validates token and extracts user info

**Implementation:**
- OAuth2 client library (NextAuth.js or custom)
- Token refresh logic
- Protected route middleware

### 5.3 API Client
**File:** `frontend/lib/api-client.ts`

**Features:**
- Axios-based HTTP client
- All requests through Oathkeeper (http://localhost:14455)
- Automatic token injection
- Response/error interceptors
- Type-safe with TypeScript

### 5.4 Document Management UI
**Pages:**
- `/dashboard` - Document list with grid/list view
- `/documents/:id` - Document viewer with preview
- `/folders/:id` - Folder view
- `/upload` - File upload with drag-drop

**Components:**
- DocumentList
- DocumentCard
- DocumentViewer
- FolderTree
- FileUpload

### 5.5 File Upload
**Features:**
- Drag-and-drop zone
- Multiple file upload
- Progress indicators
- File type validation
- Size limit checks
- Thumbnail generation

### 5.6 Search UI
**Features:**
- Global search bar
- Filters (date, type, category, tags)
- Real-time suggestions
- Highlighted results
- Advanced search modal

### 5.7 Sharing Dialog
**Features:**
- Share with specific users
- Permission levels (view/edit)
- Public link generation
- Expiration date
- Password protection

---

## Phase 6: Frontend - Admin App 📋 PLANNED

### 6.1 Admin Dashboard
**Pages:**
- `/admin/dashboard` - Overview with key metrics
- `/admin/tenants` - Tenant management
- `/admin/users` - User management across tenants
- `/admin/analytics` - Usage analytics
- `/admin/audit-logs` - Audit log viewer
- `/admin/settings` - System settings

### 6.2 Tenant Management
**Features:**
- Create/edit/delete tenants
- Tenant settings configuration
- Plan assignment
- Usage monitoring
- Quota management

### 6.3 Analytics & Audit Logs
**Features:**
- Document upload trends
- User activity heatmap
- Storage usage charts
- OCR processing stats
- Audit log filtering and export

---

## Phase 7: Testing 📋 PLANNED

### 7.1 Unit Tests
**Target:** 80%+ code coverage per service

**Tools:**
- Go: `testing`, `testify`, `mockery`
- Frontend: `vitest`, `@testing-library/react`

**What to Test:**
- Business logic in service layer
- Repository functions
- Validators
- Error handling

### 7.2 Integration Tests
**Scenarios:**
- Service-to-service communication
- Database operations
- Redis caching
- MinIO storage operations
- NATS message publishing/consuming

**Tools:** Go `testing` with docker-compose for dependencies

### 7.3 E2E Tests
**Tool:** Playwright

**Scenarios:**
- User login flow
- Document upload
- Folder creation
- Document sharing
- Search functionality
- OCR processing

### 7.4 Load Testing
**Tool:** k6

**Scenarios:**
- 1000 concurrent users
- Document upload stress test
- Search query load
- API response times

**Metrics:**
- 95th percentile < 200ms
- Error rate < 0.1%
- No memory leaks

---

## Phase 8: Production Deployment 📋 PLANNED

### 8.1 Kubernetes Manifests
**Structure:**
```
k8s/
├── base/
│   ├── namespace.yaml
│   ├── configmap.yaml
│   ├── secrets.yaml
│   └── services/
│       ├── tenant-service.yaml
│       ├── document-service.yaml
│       └── ... (all 11 services)
├── overlays/
│   ├── development/
│   ├── staging/
│   └── production/
└── infrastructure/
    ├── postgres.yaml
    ├── redis.yaml
    ├── minio.yaml
    └── ...
```

### 8.2 Helm Charts
**Features:**
- Parameterized deployments
- Environment-specific values
- Dependency management
- Rollback support

### 8.3 CI/CD Pipeline
**Platform:** GitHub Actions

**Stages:**
1. **Build:**
   - Lint code
   - Run unit tests
   - Build Docker images
   - Tag with git SHA

2. **Test:**
   - Integration tests
   - E2E tests
   - Security scanning (Trivy)

3. **Deploy:**
   - Push images to registry
   - Update Kubernetes manifests
   - Rolling deployment
   - Health checks

**Environments:**
- Development (auto-deploy on push to `develop`)
- Staging (auto-deploy on push to `main`)
- Production (manual approval + tag)

---

## Infrastructure Services Overview

| Service | Port(s) | Purpose | Health Check |
|---------|---------|---------|--------------|
| Oathkeeper | 14455, 14456 | API Gateway | ✅ Healthy |
| PostgreSQL | 15432 | Primary Database | ✅ Healthy |
| Redis | 16379 | Cache & Sessions | ✅ Healthy |
| MinIO | 19000, 19001 | Object Storage | ✅ Healthy |
| Meilisearch | 17700 | Full-text Search | ✅ Healthy |
| NATS | 14222, 18222 | Message Queue | ✅ Healthy |
| ClickHouse | 18123, 19009 | Analytics DB | ✅ Healthy |
| PaddleOCR | 18080 | OCR Engine | ⏸️ Placeholder |
| MailSlurper | 14436-14439 | Email Testing | 🔧 Dev Only |

---

## Backend Microservices Overview

| Service | Port | Status | Dependencies |
|---------|------|--------|--------------|
| Tenant Service | 10001 | 📋 Planned | PostgreSQL, Redis |
| Document Service | 10002 | 📋 Planned | PostgreSQL, Redis, NATS |
| Storage Service | 10003 | 📋 Planned | MinIO, NATS |
| Share Service | 10004 | 📋 Planned | PostgreSQL, Redis |
| RBAC Service | 10005 | 📋 Planned | PostgreSQL, Redis |
| Quota Service | 10006 | 📋 Planned | PostgreSQL, Redis |
| OCR Service | 10007 | 📋 Planned | PaddleOCR, NATS, PostgreSQL |
| Categorization Service | 10008 | 📋 Planned | PostgreSQL, NATS |
| Search Service | 10009 | 📋 Planned | Meilisearch, NATS |
| Notification Service | 10010 | 📋 Planned | PostgreSQL, NATS, SMTP |
| Audit Service | 10011 | 📋 Planned | ClickHouse, NATS |

---

## Current Progress

### ✅ Completed (Phase 1)
- Full infrastructure stack running
- Database schema with 33 tables
- Migration system with golang-migrate
- Oathkeeper API Gateway configured
- Environment configuration
- Automation scripts (Makefile)
- All services healthy and tested

### 🔄 Next Up (Phase 2)
- Implement shared Go packages
- Setup common utilities
- Create service templates

### 📊 Overall Progress
**Phase 1:** ████████████████████ 100%
**Phase 2:** ░░░░░░░░░░░░░░░░░░░░ 0%
**Total:**   ████░░░░░░░░░░░░░░░░ 12.5%

---

## Quick Start Commands

```bash
# Start all infrastructure services
make up

# Check service health
make health

# View logs
make logs-oathkeeper
make logs-postgres

# Run database migrations
make db-migrate

# Stop all services
make down

# Full cleanup
make clean
```

---

## Repository Structure

```
document-manager/
├── backend/
│   ├── pkg/                 # Shared packages
│   ├── services/            # 11 microservices
│   └── migrations/          # Database migrations (10 files)
├── frontend/
│   ├── user-app/            # Next.js user application
│   └── admin-app/           # Next.js admin application
├── config/
│   └── oathkeeper/          # Oathkeeper configuration
├── scripts/                 # Utility scripts
├── k8s/                     # Kubernetes manifests
├── docker-compose.yml       # Infrastructure orchestration
├── Makefile                 # Automation commands
├── .env                     # Environment variables (not in git)
├── .env.example             # Environment template
└── PLAN.md                  # This file
```

---

## Next Steps

1. **Implement Phase 2** - Build shared Go packages
2. **Start with Tenant Service** - First microservice as template
3. **Build remaining services** - Follow established patterns
4. **Frontend development** - User and admin apps
5. **Testing** - Comprehensive test coverage
6. **Deployment** - Production-ready infrastructure

---

## Notes

- All passwords in `.env` are test credentials for local development
- Production deployment will use Kubernetes secrets
- Shared Kratos/Hydra instances must be configured separately
- OAuth2 client registration required before frontend auth works
- Database migrations are idempotent and support rollback
- All services support graceful shutdown
- Health checks ensure proper service orchestration

---

**Last Updated:** 2025-12-19
**Current Phase:** Phase 1 Complete, Starting Phase 2
**Team:** Solo Developer
