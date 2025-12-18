# Document Manager - Implementation Plan
**Multi-Tenant Document Management System with Unified Authentication**

---

## Executive Summary

This document outlines the complete implementation plan for a production-ready multi-tenant document management system with intelligent OCR, auto-categorization, and federated search. The system integrates with existing Ory Kratos + Hydra authentication infrastructure shared across multiple applications.

**Project Duration**: 12-16 weeks (depending on team size)
**Team Size**: Recommended 3-5 developers
**Architecture**: Microservices with 11 backend services, 2 frontend apps

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Technology Stack](#technology-stack)
3. [Project Structure](#project-structure)
4. [Implementation Phases](#implementation-phases)
5. [Authentication Strategy](#authentication-strategy)
6. [Service Breakdown](#service-breakdown)
7. [Database Design](#database-design)
8. [Security Considerations](#security-considerations)
9. [Deployment Strategy](#deployment-strategy)
10. [Success Criteria](#success-criteria)

---

## Architecture Overview

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
│                      Document Manager Application                 │
│                                                                   │
│  ┌──────────────────┐                                            │
│  │ Ory Oathkeeper   │ ◄────── API Gateway (JWT Validation)       │
│  │ :14455/:14456    │                                            │
│  └─────────┬────────┘                                            │
│            │                                                      │
│  ┌─────────▼──────────────────────────────────────┐             │
│  │            Backend Microservices                │             │
│  │  ┌──────────────────────────────────────────┐  │             │
│  │  │ Tenant (10001) │ Document (10002)        │  │             │
│  │  │ Storage (10003) │ Share (10004)          │  │             │
│  │  │ RBAC (10005)   │ Quota (10006)          │  │             │
│  │  │ OCR (10007)    │ Categorization (10008) │  │             │
│  │  │ Search (10009) │ Notification (10010)   │  │             │
│  │  │ Audit (10011)                            │  │             │
│  │  └──────────────────────────────────────────┘  │             │
│  └─────────────────────────────────────────────────┘             │
│                                                                   │
│  ┌─────────────────────────────────────────────────┐             │
│  │           Infrastructure Services                │             │
│  │  ┌──────────────────────────────────────────┐   │             │
│  │  │ PostgreSQL │ Redis │ MinIO             │   │             │
│  │  │ Meilisearch │ NATS │ ClickHouse       │   │             │
│  │  │ PaddleOCR │ Prometheus │ Grafana       │   │             │
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

### Key Architectural Decisions

1. **Microservices Architecture**: Each service has a single responsibility and can scale independently
2. **Unified Authentication**: NO local auth - all authentication via shared Ory infrastructure
3. **Automatic Tenant Creation**: Tenants are created on first user access (not during registration)
4. **Event-Driven Architecture**: NATS JetStream for asynchronous processing (OCR, notifications)
5. **Multi-Tenancy**: Complete data isolation per tenant at the database level
6. **API Gateway Pattern**: Oathkeeper validates JWTs before requests reach services

---

## Technology Stack

### Backend
- **Framework**: Go 1.22+ with Fiber v2
- **Communication**: gRPC for inter-service, REST for external
- **Database**: PostgreSQL 16 (primary data)
- **Cache**: Redis 7 (sessions, caching)
- **Object Storage**: MinIO (S3-compatible)
- **Search**: Meilisearch v1.5+
- **OCR**: PaddleOCR v2.7+
- **Message Queue**: NATS JetStream 2.10+
- **Analytics**: ClickHouse 23+

### Frontend
- **Framework**: Next.js 14+ (App Router)
- **Language**: TypeScript 5.3+
- **UI Components**: shadcn/ui + Radix UI
- **Styling**: TailwindCSS 3.4+
- **State Management**:
  - Server: TanStack Query v5
  - Client: Zustand 4.5+
- **Forms**: React Hook Form + Zod

### Authentication (Shared)
- **Identity**: Ory Kratos v1.2 (SHARED)
- **OAuth2/OIDC**: Ory Hydra v2.2 (SHARED)
- **API Gateway**: Ory Oathkeeper v0.40 (OUR INSTANCE)

### Infrastructure
- **Development**: Docker Compose
- **Production**: Kubernetes 1.28+
- **Monitoring**: Prometheus + Grafana
- **Package Management**: Helm 3+

### Port Allocation (All 5-digit)
```
Backend Services:
- Tenant:          10001
- Document:        10002
- Storage:         10003
- Share:           10004
- RBAC:            10005
- Quota:           10006
- OCR:             10007
- Categorization:  10008
- Search:          10009
- Notification:    10010
- Audit:           10011

Frontend:
- User App:        13000
- Admin App:       13001
- Grafana:         13002

Infrastructure:
- PostgreSQL:      15432
- Redis:           16379
- Meilisearch:     17700
- MinIO:           19000/19001
- NATS:            14222/18222
- ClickHouse:      18123/19000
- Oathkeeper:      14455/14456
- Prometheus:      19090
```

---

## Project Structure

```
document-manager/
├── backend/
│   ├── pkg/                              # Shared packages
│   │   ├── config/                       # Configuration (Viper)
│   │   ├── database/                     # PostgreSQL pool
│   │   ├── cache/                        # Redis client
│   │   ├── logger/                       # Zap logging
│   │   ├── middleware/                   # Fiber middleware
│   │   ├── validator/                    # Input validation
│   │   ├── response/                     # Standardized responses
│   │   ├── errors/                       # Custom errors
│   │   └── metrics/                      # Prometheus metrics
│   │
│   ├── services/
│   │   ├── tenant-service/               # User & tenant management
│   │   ├── document-service/             # Document CRUD
│   │   ├── storage-service/              # File storage (MinIO)
│   │   ├── share-service/                # Sharing & permissions
│   │   ├── rbac-service/                 # Role-based access
│   │   ├── quota-service/                # Usage tracking
│   │   ├── ocr-service/                  # OCR processing
│   │   ├── categorization-service/       # ML classification
│   │   ├── search-service/               # Federated search
│   │   ├── notification-service/         # Emails & notifications
│   │   └── audit-service/                # Activity logging
│   │
│   └── migrations/                       # Database migrations
│
├── frontend/
│   ├── user-app/                         # User-facing app
│   │   ├── src/
│   │   │   ├── app/                      # Next.js App Router
│   │   │   ├── components/               # React components
│   │   │   ├── lib/                      # Utilities
│   │   │   └── hooks/                    # Custom hooks
│   │   └── package.json
│   │
│   └── admin-app/                        # Admin dashboard
│       └── ...
│
├── config/
│   ├── oathkeeper/                       # API gateway config
│   ├── prometheus/                       # Monitoring config
│   └── grafana/                          # Dashboards
│
├── docs/
│   ├── ARCHITECTURE.md
│   ├── AUTH_INTEGRATION.md
│   ├── API_DOCUMENTATION.md
│   └── DEPLOYMENT.md
│
├── scripts/
│   ├── init-db.sql
│   ├── test-shared-auth.sh
│   └── setup-oauth-client.sh
│
├── docker-compose.yml
├── .env.example
├── Makefile
└── README.md
```

---

## Implementation Phases

### Phase 1: Infrastructure Setup (Week 1-2)

**Objective**: Set up complete development environment

**Deliverables**:
1. ✅ Docker Compose with all infrastructure services
2. ✅ Environment configuration (.env.example)
3. ✅ Oathkeeper configuration (API gateway)
4. ✅ Database migrations (all 10 migrations)
5. ✅ Makefile with automation commands

**Critical Files**:
- `docker-compose.yml` - 15+ services configured
- `.env.example` - All required environment variables
- `config/oathkeeper/oathkeeper.yml` - JWT validation rules
- `config/oathkeeper/access-rules.yml` - Route protection
- `backend/migrations/` - Complete database schema

**Success Criteria**:
- [ ] All Docker services start successfully
- [ ] Database migrations run without errors
- [ ] Oathkeeper validates JWT tokens correctly
- [ ] Can connect to shared Kratos/Hydra (test script)

**Key Commands**:
```bash
make up              # Start all services
make db-migrate      # Run migrations
make test-shared-auth # Test auth connectivity
make logs service=tenant # View service logs
```

---

### Phase 2: Shared Backend Packages (Week 2-3)

**Objective**: Build reusable Go packages used by all services

**Deliverables**:
1. ✅ Config management (Viper)
2. ✅ Database connection pool (pgx)
3. ✅ Redis client wrapper
4. ✅ Structured logging (Zap)
5. ✅ Fiber middleware (auth, logging, metrics)
6. ✅ Input validation
7. ✅ Standardized JSON responses
8. ✅ Custom error types
9. ✅ Prometheus metrics

**Package Structure**:
```go
// pkg/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Auth     AuthConfig
    // ...
}

// pkg/database/database.go
func New(dsn string) (*pgxpool.Pool, error)

// pkg/cache/cache.go
func New(addr, password string, db int) *redis.Client

// pkg/logger/logger.go
func New(level, format string) *zap.Logger

// pkg/middleware/auth.go
func RequireAuth(kratos KratosClient) fiber.Handler

// pkg/response/response.go
func Success(data interface{}) map[string]interface{}
func Error(code int, message string) map[string]interface{}
```

**Success Criteria**:
- [ ] All packages have unit tests (>80% coverage)
- [ ] Documentation for each exported function
- [ ] Example usage in README
- [ ] No external dependencies leak into packages

---

### Phase 3: Core Backend Services (Week 3-6)

**Priority Order**: Tenant → Document → Storage → RBAC

#### 3.1 Tenant Service (Port 10001)

**Objective**: Manage tenants, users, and auto-creation logic

**Key Features**:
- Auto-create tenant on first user access
- Sync user data from Kratos Admin API
- Plan management (Free/Pro/Enterprise)
- Quota initialization
- Webhook handler for user creation

**Critical Functions**:
```go
// GetOrCreateTenantForUser - CORE FEATURE
func (s *TenantService) GetOrCreateTenantForUser(ctx context.Context, userID string) (*Tenant, error) {
    // 1. Check cache for existing tenant
    // 2. Query database
    // 3. If not found:
    //    a. Fetch user from Kratos Admin API
    //    b. Create tenant with default plan (Free)
    //    c. Initialize quotas (5GB storage, 50 OCR pages)
    //    d. Create user record in tenant
    // 4. Cache and return
}
```

**Endpoints**:
```
POST   /api/v1/tenants                      # Create tenant (internal)
GET    /api/v1/tenants/:id                  # Get tenant
PUT    /api/v1/tenants/:id                  # Update tenant
GET    /api/v1/tenants/:id/users            # List users
POST   /api/v1/tenants/:id/users            # Add user
PUT    /api/v1/tenants/:id/plan             # Upgrade/downgrade plan
POST   /internal/auth/user-created          # Kratos webhook
```

#### 3.2 Document Service (Port 10002)

**Objective**: Manage documents and folders with unlimited hierarchy

**Key Features**:
- Document/folder CRUD
- Nested hierarchy (unlimited depth)
- Move/copy operations
- Soft delete (trash)
- Metadata management
- Link objects (URLs)

**Database Schema Considerations**:
```sql
-- Using Closure Table pattern for hierarchy
CREATE TABLE documents (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    type document_type NOT NULL,
    parent_id UUID,
    -- ... other fields
);

CREATE TABLE document_closure (
    ancestor_id UUID NOT NULL,
    descendant_id UUID NOT NULL,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);
```

**Endpoints**:
```
POST   /api/v1/documents                    # Create document/folder
GET    /api/v1/documents/:id                # Get document
PUT    /api/v1/documents/:id                # Update document
DELETE /api/v1/documents/:id                # Move to trash
GET    /api/v1/documents/:id/children       # List children
POST   /api/v1/documents/:id/move           # Move document
POST   /api/v1/documents/:id/copy           # Copy document
GET    /api/v1/documents/trash              # List trash
POST   /api/v1/documents/:id/restore        # Restore from trash
```

#### 3.3 Storage Service (Port 10003)

**Objective**: Handle file uploads/downloads with MinIO

**Key Features**:
- Multipart upload
- Resumable upload for large files (>100MB)
- Deduplication via SHA256
- Pre-signed URLs
- Virus scanning integration point

**Upload Flow**:
```
1. Client requests upload → Storage Service
2. Generate pre-signed URL (MinIO)
3. Client uploads directly to MinIO
4. Client notifies completion → Storage Service
5. Calculate SHA256 checksum
6. Check for duplicate (deduplication)
7. Update Document Service with file metadata
8. Trigger OCR job (NATS message)
```

**Endpoints**:
```
POST   /api/v1/storage/upload/init          # Initialize upload
POST   /api/v1/storage/upload/complete      # Complete upload
GET    /api/v1/storage/download/:id         # Download file
GET    /api/v1/storage/preview/:id          # Preview URL
DELETE /api/v1/storage/:id                  # Delete file
```

#### 3.4 RBAC Service (Port 10005)

**Objective**: Permission checking for documents

**Key Features**:
- Role-based access control
- Document-level permissions
- Permission inheritance
- Custom roles (Enterprise plan)

**Permission Levels**:
- `view`: Read-only access
- `download`: Can download files
- `edit`: Can modify documents
- `manage`: Full control (delete, share)
- `owner`: All permissions + transfer ownership

**Endpoints**:
```
POST   /api/v1/permissions/check            # Check permission
GET    /api/v1/documents/:id/permissions    # List permissions
POST   /api/v1/documents/:id/permissions    # Grant permission
DELETE /api/v1/documents/:id/permissions/:userId # Revoke permission
```

---

### Phase 4: Advanced Backend Services (Week 7-9)

#### 4.1 OCR Service (Port 10007)

**Objective**: Extract text from PDFs and images

**Key Features**:
- PaddleOCR integration
- Multi-page PDF processing
- Quota checking before processing
- 95%+ accuracy target
- Async processing via NATS

**Processing Flow**:
```
1. Receive OCR job from NATS
2. Check tenant quota (OCR pages remaining)
3. Download file from MinIO
4. Process with PaddleOCR
5. Extract text and confidence scores
6. Store text in database
7. Update quota usage
8. Trigger categorization job
9. Update search index
```

#### 4.2 Search Service (Port 10009)

**Objective**: Federated search across all documents

**Key Features**:
- Meilisearch integration
- Tenant-scoped indexing
- Full-text search (OCR content)
- Typo tolerance
- Advanced filters
- Highlighted results

**Search Index Schema**:
```json
{
  "document_id": "uuid",
  "tenant_id": "uuid",
  "name": "string",
  "content": "string",
  "type": "string",
  "category": "string",
  "tags": ["string"],
  "owner_name": "string",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

#### 4.3 Quota Service (Port 10006)

**Objective**: Track and enforce usage limits

**Key Features**:
- Real-time tracking
- Storage quota enforcement
- OCR pages quota
- Monthly reset for OCR quota
- Visual indicators
- Block operations at 100%

**Plan Limits**:
```
Free Plan:
- Storage: 5GB
- OCR Pages: 50/month
- Users: 1

Pro Plan:
- Storage: 100GB
- OCR Pages: 500/month
- Users: 10

Enterprise Plan:
- Storage: 1TB
- OCR Pages: 5000/month
- Users: Unlimited
```

#### 4.4 Share Service (Port 10004)

**Objective**: Document sharing management

**Key Features**:
- Internal sharing (same tenant)
- External sharing (email invites)
- Link sharing with expiration
- Permission levels
- Access tracking

**Share Types**:
```go
type ShareType string

const (
    ShareInternal ShareType = "internal"  // Share with tenant users
    ShareExternal ShareType = "external"  // Email invite
    ShareLink     ShareType = "link"      // Public link
)
```

---

### Phase 5: Frontend - User App (Week 10-12)

**Objective**: Build user-facing document management interface

#### 5.1 Authentication Integration

**OAuth2 Flow Implementation**:
```typescript
// lib/auth/oauth-client.ts
export async function redirectToLogin() {
  const authorizeUrl = new URL(`${HYDRA_PUBLIC_URL}/oauth2/auth`)
  authorizeUrl.searchParams.set('client_id', OAUTH2_CLIENT_ID)
  authorizeUrl.searchParams.set('response_type', 'code')
  authorizeUrl.searchParams.set('redirect_uri', `${window.location.origin}/auth/callback`)
  authorizeUrl.searchParams.set('scope', 'openid email profile offline_access')
  window.location.href = authorizeUrl.toString()
}
```

**Callback Handler**:
```typescript
// app/auth/callback/page.tsx
export default function AuthCallbackPage() {
  // 1. Extract authorization code from URL
  // 2. Exchange code for JWT token
  // 3. Store JWT in httpOnly cookie
  // 4. Redirect to /documents
}
```

#### 5.2 Core UI Components

**Document List**:
- Breadcrumb navigation
- Grid/list view toggle
- Drag-drop upload zone
- Bulk selection
- Context menu (right-click)
- Upload progress indicator

**Document Preview**:
- PDF viewer
- Image viewer
- Text preview
- Video player
- Download button

**Search Interface**:
- Search bar with autocomplete
- Advanced filters (type, date, owner, tags)
- Highlighted results
- Result snippets

**Sharing Dialog**:
- User search (internal)
- Email input (external)
- Link generation
- Expiration settings
- Permission level selector

#### 5.3 State Management

**React Query for Server State**:
```typescript
// hooks/useDocuments.ts
export function useDocuments(folderId?: string) {
  return useQuery({
    queryKey: ['documents', folderId],
    queryFn: () => apiClient.get(`/api/v1/documents?parent_id=${folderId}`)
  })
}
```

**Zustand for Client State**:
```typescript
// stores/uploadStore.ts
interface UploadStore {
  uploads: Upload[]
  addUpload: (file: File) => void
  updateProgress: (id: string, progress: number) => void
}
```

---

### Phase 6: Frontend - Admin App (Week 13-14)

**Objective**: Admin dashboard for tenant management

**Key Features**:
1. Tenant overview dashboard
2. User management (add, remove, roles)
3. Usage analytics (storage, OCR, API calls)
4. Audit log viewer
5. System configuration
6. Plan management

**Dashboard Metrics**:
- Total documents
- Total storage used
- OCR pages used (current month)
- Active users
- Recent activity
- Popular categories

---

### Phase 7: Testing (Week 15)

**Testing Strategy**:

1. **Unit Tests** (All services, >80% coverage)
   ```bash
   cd backend/services/tenant-service
   go test -v -cover ./...
   ```

2. **Integration Tests** (Testcontainers)
   ```go
   func TestTenantCreation(t *testing.T) {
       // Start PostgreSQL container
       // Run migrations
       // Test tenant creation flow
   }
   ```

3. **E2E Tests** (Playwright)
   ```typescript
   test('user can upload document', async ({ page }) => {
       await page.goto('/documents')
       await page.setInputFiles('input[type="file"]', 'test.pdf')
       await expect(page.locator('.upload-success')).toBeVisible()
   })
   ```

4. **Load Testing** (k6)
   ```javascript
   export default function () {
       http.get('http://localhost:14455/api/v1/documents')
   }
   ```

---

### Phase 8: Deployment (Week 16)

**Deployment Steps**:

1. **Kubernetes Manifests**
   - Deployments for all 11 services
   - Services for networking
   - ConfigMaps for configuration
   - Secrets for sensitive data
   - PersistentVolumeClaims for storage
   - HorizontalPodAutoscaler for scaling

2. **Helm Charts**
   ```
   helm install document-manager ./charts/document-manager \
     --set image.tag=v1.0.0 \
     --set oauth2.clientId=xxx \
     --set oauth2.clientSecret=xxx
   ```

3. **CI/CD Pipeline** (GitHub Actions)
   - Build Docker images
   - Run tests
   - Push to registry
   - Deploy to staging
   - Run smoke tests
   - Deploy to production

4. **Monitoring Setup**
   - Prometheus metrics collection
   - Grafana dashboards
   - Alerting rules
   - Log aggregation

---

## Authentication Strategy

### CRITICAL: Unified Authentication

**⚠️ DO NOT IMPLEMENT LOCAL AUTHENTICATION**

This application MUST use the shared Ory Kratos + Hydra infrastructure. Users register and log in through the Search Engine application.

### OAuth2/OIDC Flow

```
1. User navigates to Document Manager (http://localhost:13000)
   ↓
2. No session detected → Redirect to Hydra authorization endpoint
   ↓
3. Hydra checks for Kratos session → If none, redirect to Search Engine login
   ↓
4. User logs in via Search Engine (or already logged in)
   ↓
5. Hydra prompts for consent (first time only)
   ↓
6. Authorization code issued → Redirect to Document Manager callback
   ↓
7. Document Manager exchanges code for JWT (via Hydra token endpoint)
   ↓
8. First API call → Tenant Service auto-creates tenant for user
   ↓
9. User has access to their tenant
```

### JWT Token Structure

```json
{
  "sub": "kratos-identity-uuid",
  "email": "user@example.com",
  "email_verified": true,
  "name": "John Doe",
  "iss": "http://shared-hydra:14444",
  "aud": ["document-manager-client"],
  "exp": 1735689600,
  "iat": 1735686000,
  "scope": "openid email profile offline_access"
}
```

### Required Configuration

**Environment Variables**:
```bash
# Shared Authentication URLs
SHARED_KRATOS_PUBLIC_URL=http://shared-kratos:14433
SHARED_KRATOS_ADMIN_URL=http://shared-kratos:14434
SHARED_HYDRA_PUBLIC_URL=http://shared-hydra:14444
SHARED_HYDRA_ADMIN_URL=http://shared-hydra:14445

# OAuth2 Client (Register with infrastructure team)
OAUTH2_CLIENT_ID=document-manager-client
OAUTH2_CLIENT_SECRET=<obtain_from_infra_team>
OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback
```

---

## Service Breakdown

### Service Dependencies

```
Tenant Service (10001)
  ├─ Depends on: Kratos Admin API, PostgreSQL, Redis
  └─ Used by: All other services

Document Service (10002)
  ├─ Depends on: Tenant Service, PostgreSQL, Redis
  └─ Used by: Storage, Search, Share, Audit

Storage Service (10003)
  ├─ Depends on: Document Service, MinIO
  └─ Publishes: OCR jobs (NATS)

OCR Service (10007)
  ├─ Depends on: Storage Service, Quota Service, PaddleOCR
  └─ Publishes: Categorization jobs (NATS)

Categorization Service (10008)
  ├─ Depends on: Document Service, OCR Service
  └─ Publishes: Search index updates (NATS)

Search Service (10009)
  ├─ Depends on: Document Service, Meilisearch
  └─ Subscribes: Index update events (NATS)

RBAC Service (10005)
  ├─ Depends on: Tenant Service, PostgreSQL, Redis
  └─ Used by: All services (permission checks)

Share Service (10004)
  ├─ Depends on: Document Service, RBAC Service
  └─ Publishes: Notification events (NATS)

Quota Service (10006)
  ├─ Depends on: Tenant Service, PostgreSQL, Redis
  └─ Used by: Storage, OCR services

Notification Service (10010)
  ├─ Depends on: Tenant Service
  └─ Subscribes: Notification events (NATS)

Audit Service (10011)
  ├─ Depends on: ClickHouse
  └─ Subscribes: Audit events (NATS)
```

---

## Database Design

### Core Tables (10 Migrations)

#### Migration 001: Extensions and Types
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TYPE document_type AS ENUM ('file', 'folder', 'link');
CREATE TYPE share_type AS ENUM ('internal', 'external', 'link');
CREATE TYPE plan_tier AS ENUM ('free', 'pro', 'enterprise');
```

#### Migration 002: Tenants and Users
```sql
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    plan plan_tier NOT NULL DEFAULT 'free',
    storage_quota_gb INTEGER NOT NULL,
    ocr_quota_pages INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id UUID PRIMARY KEY,  -- Same as Kratos identity ID
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    email VARCHAR(320) UNIQUE NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### Migration 003: RBAC
```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    permissions JSONB NOT NULL,
    is_custom BOOLEAN DEFAULT FALSE
);

CREATE TABLE user_roles (
    user_id UUID REFERENCES users(id),
    role_id UUID REFERENCES roles(id),
    PRIMARY KEY (user_id, role_id)
);
```

#### Migration 004: Documents
```sql
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    type document_type NOT NULL,
    parent_id UUID REFERENCES documents(id),
    owner_id UUID NOT NULL REFERENCES users(id),
    size_bytes BIGINT,
    mime_type VARCHAR(127),
    storage_key VARCHAR(500),
    checksum VARCHAR(64),
    is_deleted BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Closure table for hierarchy
CREATE TABLE document_closure (
    ancestor_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    descendant_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);
```

#### Migration 005: Sharing
```sql
CREATE TABLE shares (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id),
    shared_by UUID NOT NULL REFERENCES users(id),
    shared_with_user_id UUID REFERENCES users(id),
    shared_with_email VARCHAR(320),
    type share_type NOT NULL,
    permission_level VARCHAR(20) NOT NULL,
    link_token VARCHAR(64) UNIQUE,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

#### Migration 006: OCR and Quota
```sql
CREATE TABLE ocr_results (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id),
    extracted_text TEXT,
    confidence_score DECIMAL(5,2),
    page_count INTEGER,
    processed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE quota_usage (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    usage_type VARCHAR(50) NOT NULL,
    amount BIGINT NOT NULL,
    period VARCHAR(20) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

---

## Security Considerations

### 1. Authentication Security
- ✅ JWT signature validation via Oathkeeper
- ✅ Token expiration checking
- ✅ httpOnly cookies (no XSS)
- ✅ Secure cookies in production (HTTPS only)

### 2. Authorization Security
- ✅ Tenant isolation (all queries filtered by tenant_id)
- ✅ Document-level permissions (RBAC)
- ✅ Role validation on every operation
- ✅ Permission caching (5 min TTL)

### 3. Input Validation
- ✅ All input validated with go-playground/validator
- ✅ File type validation (whitelist)
- ✅ File size limits enforced
- ✅ Parameterized SQL queries (no SQL injection)

### 4. Data Protection
- ✅ TLS for all connections in production
- ✅ Encrypted sensitive data at rest
- ✅ No sensitive data in logs
- ✅ Regular database backups

### 5. Rate Limiting
- ✅ Per-tenant rate limits
- ✅ Per-user rate limits
- ✅ API key rotation
- ✅ Abuse monitoring

---

## Deployment Strategy

### Development Environment
```bash
# Start all services
docker-compose up -d

# Run migrations
make db-migrate

# Start backend services
make run-services

# Start frontend
cd frontend/user-app && npm run dev
```

### Staging Environment
```bash
# Deploy to Kubernetes staging
kubectl apply -f k8s/staging/

# Run smoke tests
make test-staging
```

### Production Environment
```bash
# Deploy with Helm
helm upgrade --install document-manager ./charts/document-manager \
  --namespace production \
  --values values.production.yaml

# Verify deployment
kubectl rollout status deployment/tenant-service -n production
```

---

## Success Criteria

### Functional Requirements
- ✅ Users authenticate via shared OAuth2 system
- ✅ Tenants created automatically on first access
- ✅ Documents uploaded, organized, and shared
- ✅ OCR extracts text with 95%+ accuracy
- ✅ Search returns results in <200ms
- ✅ Quotas enforced per plan

### Performance Requirements
- ✅ API response time <100ms (p95)
- ✅ Upload throughput: 100+ concurrent
- ✅ Search latency: <200ms
- ✅ OCR processing: <2s per page

### Security Requirements
- ✅ All requests authenticated via JWT
- ✅ Complete tenant data isolation
- ✅ No SQL injection vulnerabilities
- ✅ No XSS vulnerabilities
- ✅ Rate limiting active

### Reliability Requirements
- ✅ 99.9% uptime SLA
- ✅ Graceful error handling
- ✅ Automated database backups
- ✅ Disaster recovery tested

---

## Timeline and Milestones

| Phase | Duration | Milestone | Deliverables |
|-------|----------|-----------|--------------|
| Phase 1 | Week 1-2 | Infrastructure Ready | Docker Compose, Migrations, Oathkeeper |
| Phase 2 | Week 2-3 | Shared Packages | 9 reusable Go packages |
| Phase 3 | Week 3-6 | Core Services | Tenant, Document, Storage, RBAC |
| Phase 4 | Week 7-9 | Advanced Services | OCR, Search, Quota, Share |
| Phase 5 | Week 10-12 | User Frontend | Next.js user app |
| Phase 6 | Week 13-14 | Admin Frontend | Next.js admin app |
| Phase 7 | Week 15 | Testing | Unit, Integration, E2E, Load tests |
| Phase 8 | Week 16 | Deployment | K8s, Helm, CI/CD, Monitoring |

---

## Next Steps

1. **Review and Approve Plan**
   - Stakeholder review
   - Technical review
   - Budget approval

2. **Team Formation**
   - Backend developers (2-3)
   - Frontend developers (1-2)
   - DevOps engineer (1)

3. **Environment Setup**
   - Obtain OAuth2 credentials from infrastructure team
   - Setup development environments
   - Configure CI/CD pipelines

4. **Kickoff Phase 1**
   - Create repository
   - Setup project structure
   - Begin infrastructure configuration

---

## Resources

**Reference Documentation**:
- `/Users/intelifoxdz/Document Library/ORY_API_SPEC.md`
- `/Users/intelifoxdz/Document Library/ORY_INTEGRATION_GUIDE.md`
- `/Users/intelifoxdz/Document Library/ORY_QUICK_REFERENCE.md`

**External Documentation**:
- Ory Docs: https://www.ory.sh/docs
- Next.js: https://nextjs.org/docs
- Fiber: https://docs.gofiber.io
- Meilisearch: https://docs.meilisearch.com

---

## Support and Contact

For questions regarding:
- **Shared Authentication**: Contact infrastructure team
- **Application Features**: Create GitHub issue
- **Deployment**: Refer to DEPLOYMENT.md

---

**Document Version**: 1.0
**Last Updated**: 2025-12-18
**Status**: Ready for Implementation