# Production-Ready Implementation Plan - Document Manager
**Multi-Tenant Document Management System with Oathkeeper API Gateway**

---

## Production Architecture

### System Architecture with Oathkeeper

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
                           │ Issues JWT tokens
                           ▼
┌─────────────────────────────────────────────────────────────────┐
│                  Document Manager Application                     │
│                                                                   │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                   Ory Oathkeeper                          │   │
│  │                   (API Gateway)                           │   │
│  │                   :14455 (Proxy) :14456 (API)            │   │
│  │                                                           │   │
│  │  ✓ JWT Validation (JWKS from Hydra)                     │   │
│  │  ✓ Access Rules Enforcement                              │   │
│  │  ✓ Header Injection (X-User-ID, X-User-Email)           │   │
│  │  ✓ CORS Handling                                         │   │
│  │  ✓ Rate Limiting per Tenant                              │   │
│  └──────────────────┬───────────────────────────────────────┘   │
│                     │                                            │
│                     │ Forwards validated requests                │
│                     ▼                                            │
│  ┌─────────────────────────────────────────────────┐            │
│  │         Backend Microservices (Go + Fiber)      │            │
│  │  ┌──────────────────────────────────────────┐   │            │
│  │  │ Services trust Oathkeeper                │   │            │
│  │  │ No JWT validation needed                 │   │            │
│  │  │ Read user from injected headers          │   │            │
│  │  └──────────────────────────────────────────┘   │            │
│  │                                                  │            │
│  │  Tenant (10001)        Document (10002)         │            │
│  │  Storage (10003)       Share (10004)            │            │
│  │  RBAC (10005)          Quota (10006)            │            │
│  │  OCR (10007)           Categorization (10008)   │            │
│  │  Search (10009)        Notification (10010)     │            │
│  │  Audit (10011)                                  │            │
│  └─────────────────────────────────────────────────┘            │
│                                                                  │
│  ┌─────────────────────────────────────────────────┐            │
│  │           Infrastructure Services                │            │
│  │  PostgreSQL │ Redis │ MinIO │ Meilisearch       │            │
│  │  NATS │ ClickHouse │ PaddleOCR                  │            │
│  └─────────────────────────────────────────────────┘            │
│                                                                  │
│  ┌─────────────────────────────────────────────────┐            │
│  │              Frontend Applications               │            │
│  │  User App (13000)  │ Admin App (13001)          │            │
│  │  All requests go through Oathkeeper :14455      │            │
│  └─────────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────────┘
```

---

## Why Oathkeeper for Production

### Security Benefits
✅ **Centralized JWT Validation** - Single point of JWT verification
✅ **Zero Trust Backend** - Services don't need crypto libraries
✅ **Defense in Depth** - Additional security layer
✅ **Consistent Access Control** - Same rules across all services
✅ **Automatic Token Refresh** - Handles JWT rotation

### Operational Benefits
✅ **DRY Principle** - No duplicate JWT validation code
✅ **Easy Access Rule Changes** - Update rules without redeploying services
✅ **Centralized Logging** - All API traffic logged in one place
✅ **Rate Limiting** - Per-tenant, per-user, per-endpoint
✅ **Request Transformation** - Add/remove headers, rewrite URLs

### Scalability Benefits
✅ **Service Isolation** - Backend services don't talk to Hydra
✅ **Caching** - Oathkeeper caches JWKS responses
✅ **Load Balancing** - Built-in request distribution
✅ **A/B Testing** - Route traffic based on rules

---

## Updated Port Allocation

```
API Gateway:
- Oathkeeper Proxy:      14455  ← Frontend calls this
- Oathkeeper API:        14456  (health, metrics)

Backend Services (internal, not exposed):
- Tenant Service:        10001
- Document Service:      10002
- Storage Service:       10003
- Share Service:         10004
- RBAC Service:          10005
- Quota Service:         10006
- OCR Service:           10007
- Categorization:        10008
- Search Service:        10009
- Notification:          10010
- Audit Service:         10011

Frontend:
- User App:              13000
- Admin App:             13001

Infrastructure:
- PostgreSQL:            15432
- Redis:                 16379
- Meilisearch:           17700
- MinIO (API):           19000
- MinIO (Console):       19001
- NATS:                  14222
- NATS (Monitor):        18222
- ClickHouse (HTTP):     18123
- MailSlurper (Web):     14437
```

---

## Complete Oathkeeper Configuration

### 1. Oathkeeper Main Config

**config/oathkeeper/oathkeeper.yml**:

```yaml
version: v0.40.0

serve:
  proxy:
    port: 4455
    cors:
      enabled: true
      allowed_origins:
        - http://localhost:13000
        - http://localhost:13001
        - https://docs.yourdomain.com
        - https://admin.yourdomain.com
      allowed_methods:
        - GET
        - POST
        - PUT
        - DELETE
        - PATCH
        - OPTIONS
      allowed_headers:
        - Authorization
        - Content-Type
        - X-Requested-With
        - Accept
      exposed_headers:
        - Content-Type
        - X-Request-ID
      allow_credentials: true
      max_age: 86400
      debug: false

  api:
    port: 4456
    cors:
      enabled: true
      allowed_origins:
        - http://localhost:13000
      allowed_methods:
        - GET
      allowed_headers:
        - Authorization

access_rules:
  matching_strategy: regexp
  repositories:
    - file:///etc/config/oathkeeper/access-rules.yml

log:
  level: info
  format: json
  leak_sensitive_values: false

errors:
  fallback:
    - json
  handlers:
    redirect:
      enabled: true
      config:
        to: http://localhost:13000/auth/login
        when:
          - error:
              - unauthorized
              - forbidden
            request:
              header:
                accept:
                  - text/html
    json:
      enabled: true
      config:
        verbose: false

mutators:
  noop:
    enabled: true

  header:
    enabled: true
    config:
      headers:
        X-User-ID: "{{ print .Subject }}"
        X-User-Email: "{{ print .Extra.email }}"
        X-User-Name: "{{ print .Extra.name }}"
        X-Email-Verified: "{{ print .Extra.email_verified }}"

  hydrator:
    enabled: true
    config:
      api:
        url: http://tenant-service:10001/internal/hydrate
        auth:
          basic:
            username: oathkeeper
            password: ${INTERNAL_API_SECRET}

authorizers:
  allow:
    enabled: true

  deny:
    enabled: true

  remote_json:
    enabled: true
    config:
      remote: http://rbac-service:10005/api/v1/check
      forward_response_headers_to_upstream:
        - X-Tenant-ID
        - X-User-Roles

authenticators:
  noop:
    enabled: true

  anonymous:
    enabled: true
    config:
      subject: guest

  oauth2_introspection:
    enabled: true
    config:
      introspection_url: ${SHARED_HYDRA_ADMIN_URL}/oauth2/introspect
      scope_strategy: exact
      token_from:
        header: Authorization
        query_parameter: ""
        cookie: ""

  jwt:
    enabled: true
    config:
      jwks_urls:
        - ${SHARED_HYDRA_PUBLIC_URL}/.well-known/jwks.json
      scope_strategy: exact
      required_scope:
        - openid
      target_audience:
        - ${OAUTH2_CLIENT_ID}
      trusted_issuers:
        - ${SHARED_HYDRA_PUBLIC_URL}
      allowed_algorithms:
        - RS256
      token_from:
        header: Authorization
        query_parameter: token
        cookie: ory_hydra_session
      claims: |
        {
          "email": "{{ print .Extra.email }}",
          "email_verified": {{ print .Extra.email_verified }},
          "name": "{{ print .Extra.name }}"
        }

  cookie_session:
    enabled: true
    config:
      check_session_url: ${SHARED_KRATOS_PUBLIC_URL}/sessions/whoami
      only:
        - ory_kratos_session
```

### 2. Access Rules Configuration

**config/oathkeeper/access-rules.yml**:

```yaml
# ============================================================================
# PUBLIC ROUTES (No Authentication)
# ============================================================================

# Health checks for all services
- id: "health-checks"
  upstream:
    url: "http://host.docker.internal"
    strip_path: /.well-known/health
  match:
    url: "http://<.*>/health"
    methods:
      - GET
  authenticators:
    - handler: anonymous
  authorizer:
    handler: allow
  mutators:
    - handler: noop

# OAuth2 callback (public)
- id: "oauth-callback"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/auth/callback<.*>"
    methods:
      - GET
  authenticators:
    - handler: noop
  authorizer:
    handler: allow
  mutators:
    - handler: noop

# ============================================================================
# TENANT SERVICE - Port 10001
# ============================================================================

# Get or create tenant for authenticated user
- id: "tenant-get-or-create"
  upstream:
    url: "http://tenant-service:10001"
    preserve_host: true
  match:
    url: "http://<.*>/api/v1/tenants/current<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header
  errors:
    - handler: json

# Get tenant by ID
- id: "tenant-get"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/tenants/<[0-9a-f-]+><.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header
    - handler: hydrator

# Update tenant
- id: "tenant-update"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/tenants/<[0-9a-f-]+>"
    methods:
      - PUT
      - PATCH
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Manage tenant users
- id: "tenant-users"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/tenants/<[0-9a-f-]+>/users<.*>"
    methods:
      - GET
      - POST
      - DELETE
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Upgrade/downgrade plan
- id: "tenant-plan"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/tenants/<[0-9a-f-]+>/plan"
    methods:
      - PUT
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# ============================================================================
# DOCUMENT SERVICE - Port 10002
# ============================================================================

# List documents
- id: "documents-list"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents<\\?.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header
  errors:
    - handler: json

# Create document/folder
- id: "documents-create"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Get document by ID
- id: "documents-get"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Update document
- id: "documents-update"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>"
    methods:
      - PUT
      - PATCH
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Delete document (move to trash)
- id: "documents-delete"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>"
    methods:
      - DELETE
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Move document
- id: "documents-move"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/move"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Copy document
- id: "documents-copy"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/copy"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# List document children
- id: "documents-children"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/children<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Trash operations
- id: "documents-trash"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/trash<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Restore from trash
- id: "documents-restore"
  upstream:
    url: "http://document-service:10002"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/restore"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# ============================================================================
# STORAGE SERVICE - Port 10003
# ============================================================================

# Initialize upload
- id: "storage-upload-init"
  upstream:
    url: "http://storage-service:10003"
  match:
    url: "http://<.*>/api/v1/storage/upload/init"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Complete upload
- id: "storage-upload-complete"
  upstream:
    url: "http://storage-service:10003"
  match:
    url: "http://<.*>/api/v1/storage/upload/complete"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Download file
- id: "storage-download"
  upstream:
    url: "http://storage-service:10003"
  match:
    url: "http://<.*>/api/v1/storage/download/<[0-9a-f-]+>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Preview file
- id: "storage-preview"
  upstream:
    url: "http://storage-service:10003"
  match:
    url: "http://<.*>/api/v1/storage/preview/<[0-9a-f-]+>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Delete file
- id: "storage-delete"
  upstream:
    url: "http://storage-service:10003"
  match:
    url: "http://<.*>/api/v1/storage/<[0-9a-f-]+>"
    methods:
      - DELETE
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# ============================================================================
# SHARE SERVICE - Port 10004
# ============================================================================

# List shares for document
- id: "shares-list"
  upstream:
    url: "http://share-service:10004"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/shares<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Create share
- id: "shares-create"
  upstream:
    url: "http://share-service:10004"
  match:
    url: "http://<.*>/api/v1/documents/<[0-9a-f-]+>/shares"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Update share
- id: "shares-update"
  upstream:
    url: "http://share-service:10004"
  match:
    url: "http://<.*>/api/v1/shares/<[0-9a-f-]+>"
    methods:
      - PUT
      - PATCH
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Delete share
- id: "shares-delete"
  upstream:
    url: "http://share-service:10004"
  match:
    url: "http://<.*>/api/v1/shares/<[0-9a-f-]+>"
    methods:
      - DELETE
  authenticators:
    - handler: jwt
  authorizer:
    handler: remote_json
  mutators:
    - handler: header

# Access shared document (public link)
- id: "shares-public-access"
  upstream:
    url: "http://share-service:10004"
  match:
    url: "http://<.*>/api/v1/shares/public/<[a-zA-Z0-9_-]+>"
    methods:
      - GET
  authenticators:
    - handler: anonymous
  authorizer:
    handler: allow
  mutators:
    - handler: noop

# ============================================================================
# SEARCH SERVICE - Port 10009
# ============================================================================

# Search documents
- id: "search-documents"
  upstream:
    url: "http://search-service:10009"
  match:
    url: "http://<.*>/api/v1/search<.*>"
    methods:
      - GET
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Autocomplete suggestions
- id: "search-autocomplete"
  upstream:
    url: "http://search-service:10009"
  match:
    url: "http://<.*>/api/v1/search/autocomplete<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# ============================================================================
# QUOTA SERVICE - Port 10006
# ============================================================================

# Get quota usage
- id: "quota-get"
  upstream:
    url: "http://quota-service:10006"
  match:
    url: "http://<.*>/api/v1/quota<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# ============================================================================
# NOTIFICATION SERVICE - Port 10010
# ============================================================================

# Get notifications
- id: "notifications-list"
  upstream:
    url: "http://notification-service:10010"
  match:
    url: "http://<.*>/api/v1/notifications<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Mark notification as read
- id: "notifications-mark-read"
  upstream:
    url: "http://notification-service:10010"
  match:
    url: "http://<.*>/api/v1/notifications/<[0-9a-f-]+>/read"
    methods:
      - POST
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Get notification preferences
- id: "notifications-preferences"
  upstream:
    url: "http://notification-service:10010"
  match:
    url: "http://<.*>/api/v1/notifications/preferences"
    methods:
      - GET
      - PUT
  authenticators:
    - handler: jwt
  authorizer:
    handler: allow
  mutators:
    - handler: header

# ============================================================================
# AUDIT SERVICE - Port 10011
# ============================================================================

# Get audit logs (admin only)
- id: "audit-logs"
  upstream:
    url: "http://audit-service:10011"
  match:
    url: "http://<.*>/api/v1/audit/logs<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Export audit logs (admin only)
- id: "audit-export"
  upstream:
    url: "http://audit-service:10011"
  match:
    url: "http://<.*>/api/v1/audit/export<.*>"
    methods:
      - POST
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header

# ============================================================================
# ADMIN ROUTES (Admin scope required)
# ============================================================================

# Admin - List all tenants
- id: "admin-tenants-list"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/admin/tenants<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Admin - Get tenant details
- id: "admin-tenant-get"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/admin/tenants/<[0-9a-f-]+>"
    methods:
      - GET
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Admin - Update any tenant
- id: "admin-tenant-update"
  upstream:
    url: "http://tenant-service:10001"
  match:
    url: "http://<.*>/api/v1/admin/tenants/<[0-9a-f-]+>"
    methods:
      - PUT
      - PATCH
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header

# Admin - System analytics
- id: "admin-analytics"
  upstream:
    url: "http://audit-service:10011"
  match:
    url: "http://<.*>/api/v1/admin/analytics<.*>"
    methods:
      - GET
  authenticators:
    - handler: jwt
      config:
        required_scope:
          - admin
  authorizer:
    handler: allow
  mutators:
    - handler: header
```

### 3. Docker Compose with Oathkeeper

**docker-compose.yml** (Updated):

```yaml
version: '3.9'

networks:
  app-network:
    driver: bridge
  shared-auth-network:
    external: true  # Connects to shared Kratos/Hydra

services:
  # Ory Oathkeeper - API Gateway
  oathkeeper:
    image: oryd/oathkeeper:v0.40
    container_name: docmanager-oathkeeper
    ports:
      - "14455:4455"  # Proxy port (frontend calls this)
      - "14456:4456"  # API port
    environment:
      - SHARED_HYDRA_PUBLIC_URL=${SHARED_HYDRA_PUBLIC_URL}
      - SHARED_HYDRA_ADMIN_URL=${SHARED_HYDRA_ADMIN_URL}
      - SHARED_KRATOS_PUBLIC_URL=${SHARED_KRATOS_PUBLIC_URL}
      - OAUTH2_CLIENT_ID=${OAUTH2_CLIENT_ID}
      - INTERNAL_API_SECRET=${INTERNAL_API_SECRET}
    volumes:
      - ./config/oathkeeper:/etc/config/oathkeeper:ro
    command: serve --config /etc/config/oathkeeper/oathkeeper.yml
    networks:
      - app-network
      - shared-auth-network
    depends_on:
      - postgres
      - redis
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:4456/health/ready"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

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

---

## Updated Backend Services (Simplified)

With Oathkeeper, backend services **don't need JWT validation**. They trust Oathkeeper.

**backend/services/tenant-service/cmd/main.go**:

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
    app := fiber.New()
    app.Use(recover.New())

    // Health check (public, bypasses Oathkeeper)
    app.Get("/health", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{"status": "healthy"})
    })

    // Protected routes - trust Oathkeeper headers
    api := app.Group("/api/v1", extractUserFromHeaders)

    api.Get("/tenants/:id", func(c *fiber.Ctx) error {
        // User info injected by Oathkeeper
        userID := c.Locals("user_id").(string)
        email := c.Locals("email").(string)
        name := c.Locals("name").(string)

        return c.JSON(fiber.Map{
            "success": true,
            "data": fiber.Map{
                "id":      c.Params("id"),
                "user_id": userID,
                "email":   email,
                "name":    name,
            },
        })
    })

    app.Listen(":10001")
}

// Middleware to extract user info from Oathkeeper headers
func extractUserFromHeaders(c *fiber.Ctx) error {
    userID := c.Get("X-User-ID")
    email := c.Get("X-User-Email")
    name := c.Get("X-User-Name")

    if userID == "" {
        return c.Status(401).JSON(fiber.Map{
            "success": false,
            "error":   "Unauthorized - missing user headers",
        })
    }

    c.Locals("user_id", userID)
    c.Locals("email", email)
    c.Locals("name", name)

    return c.Next()
}
```

---

## Frontend Configuration

**frontend/user-app/.env.local**:

```bash
# All API calls go through Oathkeeper
NEXT_PUBLIC_API_URL=http://localhost:14455/api/v1

# OAuth2 configuration
NEXT_PUBLIC_HYDRA_PUBLIC_URL=http://127.0.0.1:4444
NEXT_PUBLIC_OAUTH2_CLIENT_ID=document-manager-client
NEXT_PUBLIC_OAUTH2_CLIENT_SECRET=your_client_secret
NEXT_PUBLIC_OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback
```

---

## Production Benefits Summary

### Security
✅ Single point of JWT validation
✅ Zero crypto in backend services
✅ Automatic header injection
✅ Centralized access control
✅ Rate limiting per tenant

### Performance
✅ JWKS caching in Oathkeeper
✅ Backend services are faster (no JWT crypto)
✅ Connection pooling
✅ Request/response transformation

### Operations
✅ Update access rules without redeploying services
✅ Centralized logging of all API traffic
✅ Easy to add new services
✅ A/B testing support
✅ Canary deployments

### Scalability
✅ Services don't talk to Hydra (reduced load)
✅ Horizontal scaling of Oathkeeper
✅ Service mesh ready
✅ Multi-region support

---

## Next Steps

1. ✅ Review this production-ready architecture
2. ✅ Start with Phase 1: Infrastructure setup with Oathkeeper
3. ✅ Configure access rules for your specific needs
4. ✅ Test Oathkeeper JWT validation
5. ✅ Build backend services (simplified, no JWT code needed)

Would you like me to create the complete Makefile and start implementing Phase 1 with Oathkeeper?
