# Document Manager

A production-ready, multi-tenant document management system with OCR, intelligent categorization, full-text search, and enterprise-grade security.

## Features

- **Multi-Tenancy** - Complete data isolation with tenant-based access control
- **Document Management** - Upload, organize, version, and manage documents
- **OCR Processing** - Automatic text extraction from scanned documents
- **Full-Text Search** - Powered by Meilisearch for lightning-fast searches
- **Intelligent Categorization** - AI-powered document classification
- **Secure Sharing** - Share documents with users or via public links
- **Role-Based Access Control** - Granular permissions system
- **Quota Management** - Plan-based usage limits and tracking
- **Audit Logging** - Comprehensive activity tracking for compliance
- **Real-time Notifications** - Email and in-app notifications
- **Analytics Dashboard** - Usage insights and reporting

## Architecture

### Microservices Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Frontend Applications                        │
│  ┌──────────────────────┐      ┌──────────────────────┐        │
│  │   User App (Next.js) │      │  Admin App (Next.js) │        │
│  │   Port: 13000        │      │   Port: 13001        │        │
│  └──────────────────────┘      └──────────────────────┘        │
└───────────────────┬─────────────────────┬───────────────────────┘
                    │                     │
                    ▼                     ▼
        ┌───────────────────────────────────────────┐
        │     Ory Oathkeeper (API Gateway)          │
        │     Ports: 14455 (Proxy), 14456 (API)     │
        │  - JWT Validation                         │
        │  - Request Routing                        │
        │  - CORS Handling                          │
        └───────────────────┬───────────────────────┘
                            │
        ┌───────────────────┴────────────────────────────────┐
        │               Backend Microservices                 │
        │  ┌─────────────────────────────────────────────┐   │
        │  │ Tenant Service (10001)                      │   │
        │  │ Document Service (10002)                    │   │
        │  │ Storage Service (10003)                     │   │
        │  │ Share Service (10004)                       │   │
        │  │ RBAC Service (10005)                        │   │
        │  │ Quota Service (10006)                       │   │
        │  │ OCR Service (10007)                         │   │
        │  │ Categorization Service (10008)              │   │
        │  │ Search Service (10009)                      │   │
        │  │ Notification Service (10010)                │   │
        │  │ Audit Service (10011)                       │   │
        │  └─────────────────────────────────────────────┘   │
        └────────────────────┬───────────────────────────────┘
                             │
        ┌────────────────────┴────────────────────────┐
        │         Infrastructure Services              │
        │  - PostgreSQL 16 (Primary Database)         │
        │  - Redis 7 (Cache & Sessions)               │
        │  - MinIO (S3-Compatible Storage)            │
        │  - Meilisearch (Full-Text Search)           │
        │  - NATS JetStream (Message Queue)           │
        │  - ClickHouse (Analytics Database)          │
        │  - PaddleOCR (OCR Engine)                   │
        └─────────────────────────────────────────────┘
```

### Technology Stack

**Backend:**
- Go 1.21+ (11 microservices)
- PostgreSQL 16 (primary database)
- Redis 7 (cache & sessions)
- NATS JetStream (message queue)

**Frontend:**
- Next.js 14+ (App Router)
- TypeScript
- shadcn/ui (Radix UI + Tailwind CSS)
- React Query (data fetching)

**Infrastructure:**
- Ory Oathkeeper (API Gateway)
- Ory Kratos (Identity Management - external)
- Ory Hydra (OAuth2/OIDC - external)
- MinIO (object storage)
- Meilisearch (full-text search)
- ClickHouse (analytics)
- PaddleOCR (OCR processing)

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Go 1.21+ (for backend development)
- Node.js 18+ (for frontend development)
- golang-migrate (for database migrations)

### 1. Clone Repository

```bash
git clone https://github.com/YOUR_USERNAME/document-manager.git
cd document-manager
```

### 2. Environment Setup

```bash
# Copy environment template
cp .env.example .env

# Edit .env and configure your settings
# For local development, the default test values work out of the box
```

### 3. Start Infrastructure

```bash
# Create external network for shared auth services
docker network create shared-auth-network

# Start all infrastructure services
make up

# Wait for services to be healthy (30-60 seconds)
make health
```

### 4. Run Database Migrations

```bash
# Install golang-migrate if not already installed
brew install golang-migrate

# Run all migrations
make db-migrate

# Verify migration status
migrate -path backend/migrations \
  -database "postgresql://postgres:testpassword12345678@localhost:15432/docmanager?sslmode=disable" \
  version
```

### 5. Verify Installation

```bash
# Check all services are healthy
make health

# Expected output:
# ✓ Oathkeeper: Healthy
# ✓ PostgreSQL: Healthy
# ✓ Redis: Healthy
# ✓ MinIO: Healthy
# ✓ Meilisearch: Healthy
# ✓ NATS: Healthy
# ✓ ClickHouse: Healthy
```

## 📦 Services

### Infrastructure (9 services)

1. **Oathkeeper** - API Gateway with JWT validation
2. **PostgreSQL** - Primary database
3. **Redis** - Cache and session store
4. **MinIO** - S3-compatible object storage
5. **Meilisearch** - Search engine
6. **PaddleOCR** - OCR processing engine
7. **NATS JetStream** - Message queue
8. **ClickHouse** - Analytics database
9. **MailSlurper** - Email testing (dev only)

### Backend Services (11 microservices)

| Service | Port | Description |
|---------|------|-------------|
| Tenant Service | 10001 | User & tenant management |
| Document Service | 10002 | Document CRUD operations |
| Storage Service | 10003 | File storage (MinIO) |
| Share Service | 10004 | Sharing & permissions |
| RBAC Service | 10005 | Role-based access control |
| Quota Service | 10006 | Usage tracking |
| OCR Service | 10007 | Text extraction |
| Categorization Service | 10008 | ML classification |
| Search Service | 10009 | Federated search |
| Notification Service | 10010 | Emails & notifications |
| Audit Service | 10011 | Activity logging |

### Frontend Applications

| App | Port | Description |
|-----|------|-------------|
| User App | 13000 | Document management interface |
| Admin App | 13001 | Admin dashboard |

## 🔧 Development

### Backend Development

```bash
# Start specific service in dev mode
make dev-backend service=tenant

# Run tests
make test

# Build service
make build-service service=tenant

# Format code
make fmt

# Run linter
make lint
```

### Frontend Development

```bash
# Start user app
make dev-frontend

# Start admin app
make dev-admin
```

### Database Operations

```bash
# Run migrations
make db-migrate

# Rollback migration
make db-rollback

# Reset database (WARNING: deletes all data)
make db-reset

# Backup database
make db-backup

# Restore from backup
make db-restore file=backups/docmanager_20240101_120000.sql.gz

# Open PostgreSQL shell
make db-shell
```

### Redis Operations

```bash
# Open Redis CLI
make redis-cli

# Clear all cached data
make redis-flush
```

## 🔐 Authentication Flow

### 1. User Login

```
User → Frontend → Hydra (Shared) → Kratos (Shared) → Login Page
                     ↓
              Issues JWT token
                     ↓
              Redirect to frontend
```

### 2. API Request

```
Frontend → Oathkeeper (validates JWT) → Backend Service
              ↓                              ↓
    Injects X-User-ID header      Trusts Oathkeeper headers
    Injects X-User-Email
    Injects X-User-Name
```

### 3. Tenant Auto-Creation

```
First API call → Tenant Service checks cache
                     ↓
               Tenant not found
                     ↓
          Fetch user from Kratos Admin API
                     ↓
          Create tenant with Free plan
                     ↓
          Initialize quotas (5GB, 50 OCR pages)
                     ↓
               Cache tenant
                     ↓
             Return tenant context
```

## 📊 Monitoring

```bash
# Check service health
make health

# View resource usage
make stats

# View running processes
make top

# Open all monitoring dashboards
make monitor
```

## 🧪 Testing

```bash
# Test shared authentication connectivity
make test-auth

# Run unit tests
make test

# Run integration tests
make test-integration

# Run E2E tests
make test-e2e
```

## 🚢 Deployment

### Staging

```bash
make deploy-staging
```

### Production

```bash
make deploy-prod
```

## 📝 Configuration

### Required Environment Variables

```bash
# Database
DB_PASSWORD=<strong-password>

# Redis
REDIS_PASSWORD=<strong-password>

# MinIO
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=<strong-password>

# Meilisearch (must be exactly 32 characters)
MEILI_MASTER_KEY=<32-character-key>

# Shared Authentication (from infrastructure team)
SHARED_KRATOS_PUBLIC_URL=http://shared-kratos:14433
SHARED_KRATOS_ADMIN_URL=http://shared-kratos:14434
SHARED_HYDRA_PUBLIC_URL=http://shared-hydra:14444
SHARED_HYDRA_ADMIN_URL=http://shared-hydra:14445

# OAuth2 Client (register with infrastructure team)
OAUTH2_CLIENT_ID=document-manager-client
OAUTH2_CLIENT_SECRET=<client-secret>
OAUTH2_REDIRECT_URI=http://localhost:13000/auth/callback

# Internal API
INTERNAL_API_SECRET=<strong-secret>
```

## 🎯 Subscription Plans

| Plan | Storage | OCR Pages/mo | Users | Price |
|------|---------|--------------|-------|-------|
| Free | 5 GB | 50 | 1 | $0 |
| Pro | 100 GB | 500 | 10 | $9.99 |
| Enterprise | 1 TB | 5000 | Unlimited | $49.99 |

## 🔍 Useful Commands

```bash
# Show all available commands
make help

# Start services
make up

# Stop services
make down

# Restart services
make restart

# View logs (all services)
make logs

# View logs (specific service)
make logs service=postgres

# Check health of all services
make health

# Show service status
make ps

# Show resource usage
make stats

# Clean up everything
make clean
```

## 📚 Documentation

- [PRODUCTION_IMPLEMENTATION_PLAN.md](./PRODUCTION_IMPLEMENTATION_PLAN.md) - Complete implementation guide
- [ARCHITECTURE_DECISIONS.md](./ARCHITECTURE_DECISIONS.md) - Architectural choices explained
- [QUICK_START_GUIDE.md](./QUICK_START_GUIDE.md) - Developer quick start
- [ORY_INTEGRATION_GUIDE.md](./ORY_INTEGRATION_GUIDE.md) - Authentication integration
- [ORY_API_SPEC.md](./ORY_API_SPEC.md) - Ory API specifications
- [ORY_QUICK_REFERENCE.md](./ORY_QUICK_REFERENCE.md) - Quick reference

## 🏗️ Project Structure

```
document-manager/
├── backend/
│   ├── pkg/                    # Shared packages
│   ├── services/               # 11 microservices
│   └── migrations/             # Database migrations
├── frontend/
│   ├── user-app/               # Next.js user interface
│   └── admin-app/              # Next.js admin dashboard
├── config/
│   ├── oathkeeper/             # API gateway configuration
│   ├── clickhouse/             # ClickHouse configuration
│   └── prometheus/             # Monitoring (future)
├── scripts/                    # Utility scripts
├── docs/                       # Documentation
├── docker-compose.yml          # Infrastructure services
├── .env.example                # Environment template
├── Makefile                    # Development commands
└── README.md                   # This file
```

## 🔥 Common Issues

### Port Already in Use

```bash
# Find process using port
lsof -i :14455

# Kill process
kill -9 <PID>
```

### Cannot Connect to Database

```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check logs
make logs service=postgres

# Restart service
docker-compose restart postgres
```

### Oathkeeper Not Validating JWT

```bash
# Check if Hydra is reachable
make test-auth

# Check Oathkeeper configuration
cat config/oathkeeper/oathkeeper.yml

# View Oathkeeper logs
make logs service=oathkeeper
```

### Services Not Starting

```bash
# Check Docker resources
docker system df

# Clean up
make clean

# Restart
make up
```

## 🤝 Contributing

1. Create feature branch
2. Make changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Run linter: `make lint`
6. Submit pull request

## 📄 License

Proprietary - All rights reserved

## 📞 Support

**For issues with:**
- **Shared Authentication** - Contact infrastructure team
- **Application Features** - Create GitHub issue
- **Deployment** - Refer to deployment documentation

## Current Status

**Phase 1: Infrastructure Setup** ✅ **COMPLETE**

All infrastructure services are running and healthy:
- 9 infrastructure services configured and tested
- 33 database tables created via 10 migrations
- Seed data loaded (roles, plans)
- Automation scripts ready (Makefile)
- Documentation complete

**Next:** Phase 2 - Implement shared Go packages for all microservices.

## Roadmap

- [x] **Phase 1:** Infrastructure Setup (Complete)
- [ ] **Phase 2:** Shared Backend Packages (Next)
- [ ] **Phase 3:** Core Backend Services
- [ ] **Phase 4:** Advanced Backend Services
- [ ] **Phase 5:** Frontend - User App
- [ ] **Phase 6:** Frontend - Admin App
- [ ] **Phase 7:** Testing & Quality Assurance
- [ ] **Phase 8:** Production Deployment

See [PLAN.md](./PLAN.md) for detailed phase breakdown.

## License

[Add your license here]

## Support

For issues and questions:
- GitHub Issues: [Create an issue](https://github.com/YOUR_USERNAME/document-manager/issues)

## Acknowledgments

- Built with [Ory](https://www.ory.sh/) for authentication and authorization
- UI components from [shadcn/ui](https://ui.shadcn.com/)
- Powered by open-source infrastructure

---

**Last Updated:** 2025-12-19
**Version:** 0.1.0 (Phase 1 Complete)
**Status:** Active Development
