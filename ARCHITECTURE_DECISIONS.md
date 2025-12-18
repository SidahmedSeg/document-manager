# Architecture Decision Records (ADR)
**Document Manager - Multi-Tenant System**

---

## ADR-001: Unified Authentication via Ory Infrastructure

**Status**: Accepted
**Date**: 2025-12-18
**Deciders**: Architecture Team

### Context
The Document Manager is part of an ecosystem with existing applications (Search Engine, Email App) that already use Ory Kratos + Hydra for authentication. We need to decide whether to:
1. Implement local authentication
2. Integrate with existing shared authentication infrastructure

### Decision
**Use shared Ory Kratos + Hydra infrastructure via OAuth2/OIDC flow.**

### Rationale
- **Single Sign-On (SSO)**: Users can access all applications with one login
- **Reduced Development Time**: No need to build authentication flows
- **Security**: Leverage battle-tested identity management
- **Consistency**: Same user experience across all applications
- **Compliance**: Centralized audit trail for authentication events

### Implementation
- Frontend redirects to Hydra authorization endpoint
- Exchange authorization code for JWT
- Store JWT in httpOnly cookie
- Validate JWT via Oathkeeper API Gateway
- Auto-create tenant on first user access

### Consequences
**Positive**:
- Faster time to market
- Lower maintenance burden
- Better security posture
- Unified user experience

**Negative**:
- Dependency on shared infrastructure availability
- Need coordination with infrastructure team for OAuth2 client registration
- Cannot customize authentication flow without affecting other apps

---

## ADR-002: Microservices Architecture

**Status**: Accepted
**Date**: 2025-12-18

### Context
We need to decide on the architectural pattern for the backend:
1. Monolithic application
2. Microservices architecture
3. Modular monolith

### Decision
**Implement microservices architecture with 11 independent services.**

### Rationale
- **Scalability**: Each service can scale independently based on load
- **Technology Flexibility**: Can use different tech stacks if needed
- **Team Autonomy**: Different teams can work on different services
- **Fault Isolation**: Failure in one service doesn't bring down entire system
- **Deployment Independence**: Services can be deployed separately

### Service Boundaries
```
Tenant Service       → User & tenant management
Document Service     → Document CRUD operations
Storage Service      → File storage (MinIO)
Share Service        → Sharing & permissions
RBAC Service         → Role-based access control
Quota Service        → Usage tracking & enforcement
OCR Service          → Text extraction
Categorization Service → ML classification
Search Service       → Federated search
Notification Service → Emails & notifications
Audit Service        → Activity logging
```

### Communication Patterns
- **Synchronous**: HTTP/REST for real-time operations
- **Asynchronous**: NATS JetStream for background jobs
- **Internal**: gRPC for service-to-service communication

### Consequences
**Positive**:
- Highly scalable
- Technology agnostic
- Easy to maintain and extend
- Clear separation of concerns

**Negative**:
- Increased operational complexity
- Network latency between services
- Distributed tracing required
- More complex deployment

---

## ADR-003: Automatic Tenant Creation

**Status**: Accepted
**Date**: 2025-12-18

### Context
In a multi-tenant system, we need to decide when to create tenants:
1. During user registration (in Kratos)
2. On first access to the application
3. Through manual admin action

### Decision
**Automatically create tenant on first user access to the Document Manager.**

### Rationale
- **Seamless UX**: Users don't need separate onboarding
- **Lazy Initialization**: Only create resources when needed
- **Shared Auth**: Users register through Search Engine, not our app
- **Just-in-Time Provisioning**: Tenant created exactly when needed

### Implementation Flow
```
1. User logs in via OAuth2
2. JWT extracted from callback
3. First API call to Document Manager
4. Tenant Service checks if tenant exists for user
5. If not:
   a. Fetch user data from Kratos Admin API
   b. Create tenant with default Free plan
   c. Initialize quotas (5GB storage, 50 OCR pages)
   d. Create user record in tenant
   e. Cache tenant info in Redis
6. Return tenant context with response
```

### Consequences
**Positive**:
- Zero-friction user experience
- No manual provisioning needed
- Aligns with shared auth model
- Reduces abandoned sign-ups

**Negative**:
- First API call slower (tenant creation overhead)
- Need robust error handling for creation failures
- Requires Kratos Admin API access

---

## ADR-004: PostgreSQL for Primary Data Storage

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose primary database for relational data:
1. PostgreSQL
2. MySQL
3. MongoDB

### Decision
**Use PostgreSQL 16 as primary database.**

### Rationale
- **JSONB Support**: Store flexible metadata without schema changes
- **Full-Text Search**: Native trgm extension for text search
- **Transaction Support**: ACID guarantees for data integrity
- **Advanced Features**: CTEs, window functions, custom types
- **Scalability**: Proven at enterprise scale
- **Ecosystem**: Rich tooling and community support

### Schema Design Patterns
- **Closure Table**: For unlimited document hierarchy
- **Soft Deletes**: For trash functionality
- **Row-Level Security**: For multi-tenancy (future consideration)
- **Partitioning**: By tenant_id for large datasets

### Consequences
**Positive**:
- Excellent data integrity
- Powerful query capabilities
- Battle-tested reliability
- Strong ecosystem

**Negative**:
- Vertical scaling limits
- Requires careful index management
- Need connection pooling (pgx)

---

## ADR-005: MinIO for Object Storage

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose object storage solution for files:
1. AWS S3
2. MinIO (S3-compatible)
3. Local filesystem
4. Azure Blob Storage

### Decision
**Use MinIO as S3-compatible object storage.**

### Rationale
- **S3 Compatibility**: Can migrate to AWS S3 later without code changes
- **Self-Hosted**: Full control over data and costs
- **High Performance**: Optimized for large files
- **Kubernetes Native**: Easy to deploy in K8s
- **Multi-Tenancy**: Built-in bucket policies
- **Deduplication**: Content-addressable storage

### Bucket Structure
```
tenant-{tenant-id}/
  ├── documents/
  │   ├── {document-id}/
  │   │   └── {filename}
  └── thumbnails/
      └── {document-id}.jpg
```

### Consequences
**Positive**:
- Cost-effective for self-hosting
- S3 API compatibility
- Easy migration path to cloud
- Excellent performance

**Negative**:
- Need to manage storage infrastructure
- Backup strategy required
- Monitoring setup needed

---

## ADR-006: Meilisearch for Search Engine

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose search engine for document search:
1. Elasticsearch
2. Meilisearch
3. Algolia
4. PostgreSQL full-text search

### Decision
**Use Meilisearch v1.5+ as search engine.**

### Rationale
- **Instant Results**: <200ms search latency
- **Typo Tolerance**: Handles spelling mistakes automatically
- **Easy to Use**: Simple API, minimal configuration
- **Lightweight**: Lower resource usage than Elasticsearch
- **Multi-Tenant**: Index per tenant for isolation
- **Faceted Search**: Advanced filtering out of the box

### Index Strategy
```
Index per tenant: documents_tenant_{tenant_id}

Document Schema:
{
  "id": "uuid",
  "name": "string",
  "content": "string",     // OCR extracted text
  "type": "string",
  "category": "string",
  "tags": ["string"],
  "owner_name": "string",
  "created_at": "timestamp"
}
```

### Consequences
**Positive**:
- Fast search performance
- Great developer experience
- Lower operational overhead
- Built-in typo tolerance

**Negative**:
- Less mature than Elasticsearch
- Fewer advanced features
- Smaller community

---

## ADR-007: NATS JetStream for Message Queue

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose message queue for async processing:
1. RabbitMQ
2. Apache Kafka
3. NATS JetStream
4. AWS SQS

### Decision
**Use NATS JetStream 2.10+ as message queue.**

### Rationale
- **Lightweight**: Minimal resource footprint
- **High Performance**: Millions of messages per second
- **Simple**: Easy to setup and operate
- **At-Least-Once Delivery**: Message guarantees
- **Kubernetes Native**: Excellent K8s integration
- **Persistence**: JetStream adds durability to NATS

### Event Streams
```
documents.uploaded     → Triggers OCR processing
documents.ocr.completed → Triggers categorization
documents.categorized  → Updates search index
documents.shared       → Sends notifications
tenants.created        → Initializes resources
```

### Consequences
**Positive**:
- Simple to operate
- Excellent performance
- Low latency
- Good monitoring

**Negative**:
- Fewer features than Kafka
- Smaller ecosystem
- Less mature tooling

---

## ADR-008: ClickHouse for Analytics

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose analytics database for audit logs and usage tracking:
1. PostgreSQL
2. ClickHouse
3. TimescaleDB
4. BigQuery

### Decision
**Use ClickHouse 23+ for analytics and audit logs.**

### Rationale
- **Columnar Storage**: Optimized for analytical queries
- **High Performance**: Billions of rows, sub-second queries
- **Real-Time**: Insert and query simultaneously
- **Compression**: 10x better than row-based databases
- **Scalability**: Petabyte-scale data
- **Time Series**: Perfect for audit logs and metrics

### Schema Design
```sql
CREATE TABLE audit_logs (
    timestamp DateTime64(3),
    tenant_id UUID,
    user_id UUID,
    action String,
    resource_type String,
    resource_id UUID,
    metadata String,
    ip_address String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, timestamp);
```

### Consequences
**Positive**:
- Blazing fast analytics
- Efficient storage
- Real-time insights
- Excellent compression

**Negative**:
- Not suitable for transactional data
- Limited UPDATE/DELETE support
- Different query syntax (SQL-like)

---

## ADR-009: Next.js 14 with App Router

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose frontend framework:
1. React with Create React App
2. Next.js (Pages Router)
3. Next.js 14 (App Router)
4. Vue.js / Nuxt
5. SvelteKit

### Decision
**Use Next.js 14+ with App Router for both user and admin apps.**

### Rationale
- **Server Components**: Better performance, reduced bundle size
- **Built-in Routing**: File-based routing with App Router
- **API Routes**: Backend for frontend pattern
- **SEO Friendly**: Server-side rendering
- **Developer Experience**: Fast refresh, TypeScript support
- **Production Ready**: Used by Fortune 500 companies

### Project Structure
```
src/
├── app/                    # App Router
│   ├── auth/
│   │   └── callback/       # OAuth2 callback
│   ├── documents/
│   │   └── [id]/
│   ├── shared/
│   └── settings/
├── components/             # React components
│   ├── ui/                 # shadcn/ui components
│   └── documents/
├── lib/                    # Utilities
│   ├── api/                # API client
│   └── auth/               # OAuth2 client
└── hooks/                  # Custom hooks
```

### Consequences
**Positive**:
- Modern React features
- Better performance
- Excellent DX
- Large community

**Negative**:
- Learning curve for App Router
- Server Components paradigm shift
- Need Node.js runtime

---

## ADR-010: shadcn/ui for Component Library

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose UI component library:
1. Material-UI
2. Ant Design
3. Chakra UI
4. shadcn/ui + Radix UI

### Decision
**Use shadcn/ui with Radix UI primitives.**

### Rationale
- **Customizable**: Copy components to your codebase
- **Accessible**: Built on Radix UI (ARIA compliant)
- **Modern Design**: Beautiful default styling
- **No Bundle Size**: Only use what you need
- **TypeScript**: Full type safety
- **Tailwind**: Styled with TailwindCSS

### Component Usage
```typescript
import { Button } from "@/components/ui/button"
import { Dialog } from "@/components/ui/dialog"
import { Select } from "@/components/ui/select"

<Button variant="outline" size="lg">
  Upload Document
</Button>
```

### Consequences
**Positive**:
- Full control over components
- Excellent accessibility
- Beautiful design
- Small bundle size

**Negative**:
- Need to copy components
- Manual updates required
- Less comprehensive than MUI

---

## ADR-011: PaddleOCR for Text Extraction

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose OCR engine:
1. Tesseract OCR
2. Google Cloud Vision
3. AWS Textract
4. PaddleOCR

### Decision
**Use PaddleOCR v2.7+ for text extraction.**

### Rationale
- **High Accuracy**: 95%+ accuracy on printed text
- **Multi-Language**: Supports 80+ languages
- **Self-Hosted**: No cloud dependencies
- **Fast**: GPU acceleration support
- **Free**: Open source, no API costs
- **Multi-Page**: Handles PDFs efficiently

### Processing Pipeline
```
1. Receive OCR job from NATS
2. Download file from MinIO
3. Convert PDF to images (if needed)
4. Process each page with PaddleOCR
5. Extract text + confidence scores
6. Store in PostgreSQL
7. Update Meilisearch index
8. Trigger categorization
```

### Consequences
**Positive**:
- No API costs
- Data privacy (on-premise)
- High accuracy
- Fast processing

**Negative**:
- Requires GPU for best performance
- Need to manage OCR service
- Resource intensive

---

## ADR-012: Prometheus + Grafana for Monitoring

**Status**: Accepted
**Date**: 2025-12-18

### Context
Choose monitoring solution:
1. Prometheus + Grafana
2. Datadog
3. New Relic
4. ELK Stack

### Decision
**Use Prometheus for metrics collection and Grafana for visualization.**

### Rationale
- **Industry Standard**: De facto standard in Kubernetes
- **Pull-Based**: Services expose metrics, Prometheus scrapes
- **Powerful Queries**: PromQL for complex queries
- **Alerting**: Built-in alert manager
- **Free**: Open source, no license costs
- **Ecosystem**: Large collection of exporters

### Metrics to Track
```
# HTTP Metrics
http_requests_total
http_request_duration_seconds
http_request_size_bytes

# Business Metrics
documents_uploaded_total
storage_bytes_used
ocr_pages_processed_total
active_users_gauge

# System Metrics
go_goroutines
go_memstats_alloc_bytes
process_cpu_seconds_total
```

### Consequences
**Positive**:
- Comprehensive monitoring
- Excellent alerting
- Beautiful dashboards
- No vendor lock-in

**Negative**:
- Requires separate setup
- Retention management needed
- Alert fatigue possible

---

## Summary of Key Decisions

| Decision | Choice | Primary Reason |
|----------|--------|----------------|
| Authentication | Shared Ory (Kratos + Hydra) | SSO across applications |
| Architecture | Microservices | Independent scaling & deployment |
| Tenant Creation | Automatic on first access | Seamless user experience |
| Primary Database | PostgreSQL 16 | ACID compliance & JSONB |
| Object Storage | MinIO | S3-compatible, self-hosted |
| Search Engine | Meilisearch | Fast, simple, typo-tolerant |
| Message Queue | NATS JetStream | Lightweight, high-performance |
| Analytics | ClickHouse | Columnar storage for logs |
| Frontend | Next.js 14 App Router | Modern React, SSR |
| UI Library | shadcn/ui + Radix | Accessible, customizable |
| OCR Engine | PaddleOCR | High accuracy, self-hosted |
| Monitoring | Prometheus + Grafana | Industry standard |

---

## Risk Mitigation

### Risk: Shared Auth Infrastructure Downtime
**Mitigation**:
- Monitor Kratos/Hydra health proactively
- Implement graceful degradation (read-only mode)
- Cache user sessions (15 min TTL)
- SLA agreement with infrastructure team

### Risk: Service Communication Failures
**Mitigation**:
- Circuit breaker pattern
- Retry with exponential backoff
- Fallback responses
- Health checks on all services

### Risk: OCR Processing Bottleneck
**Mitigation**:
- Horizontal scaling of OCR service
- GPU acceleration
- Queue depth monitoring
- Priority queue for premium users

### Risk: Storage Costs
**Mitigation**:
- Deduplication via SHA256
- Compression for text files
- Lifecycle policies (archive old files)
- Usage monitoring and alerts

### Risk: Database Performance
**Mitigation**:
- Proper indexing strategy
- Connection pooling
- Read replicas for heavy read workloads
- Partitioning by tenant_id

---

**Document Version**: 1.0
**Last Updated**: 2025-12-18
**Status**: Approved
