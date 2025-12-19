# Tenant Service

The Tenant Service manages multi-tenancy for the document management system. It handles tenant creation, user membership, invitations, and tenant settings.

## Features

- **Tenant Management** - Create, read, update tenants
- **User Membership** - Manage users within tenants
- **Invitations** - Invite users to join tenants
- **Role-Based Access** - Admin and user roles
- **Caching** - Redis caching for improved performance
- **Validation** - Input validation with custom rules
- **Logging** - Structured logging with request correlation

## Architecture

```
cmd/
  └── main.go              # Service entry point
internal/
  ├── handler/             # HTTP handlers
  │   └── handler.go
  ├── service/             # Business logic
  │   └── service.go
  ├── repository/          # Database access
  │   └── repository.go
  └── models/              # Data models
      └── models.go
```

## API Endpoints

### Tenant Operations

#### Create Tenant
```http
POST /api/tenants
Authorization: Bearer <token>

Request:
{
  "name": "Acme Corporation",
  "slug": "acme-corp",
  "domain": "acme.com"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Acme Corporation",
    "slug": "acme-corp",
    "domain": "acme.com",
    "subscription_plan": "free",
    "is_active": true,
    "created_at": "2025-12-19T10:00:00Z",
    "updated_at": "2025-12-19T10:00:00Z"
  }
}
```

#### Get Tenant
```http
GET /api/tenants/{id}
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "id": "uuid",
    "name": "Acme Corporation",
    ...
  }
}
```

#### Update Tenant
```http
PUT /api/tenants/{id}
Authorization: Bearer <token>

Request:
{
  "name": "Acme Inc.",
  "is_active": true
}

Response: 200 OK
{
  "success": true,
  "data": {
    "message": "tenant updated successfully"
  }
}
```

#### Get My Tenants
```http
GET /api/tenants/me
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "name": "Acme Corporation",
      "slug": "acme-corp",
      ...
    }
  ]
}
```

### User Management

#### Get Tenant Users
```http
GET /api/tenants/{id}/users
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "user_id": "kratos-user-id",
      "user_email": "user@example.com",
      "role": "admin",
      "is_owner": true,
      "joined_at": "2025-12-19T10:00:00Z"
    }
  ]
}
```

#### Invite User
```http
POST /api/tenants/{id}/users/invite
Authorization: Bearer <token>

Request:
{
  "email": "newuser@example.com",
  "role": "user"
}

Response: 201 Created
{
  "success": true,
  "data": {
    "id": "uuid",
    "tenant_id": "uuid",
    "email": "newuser@example.com",
    "role": "user",
    "expires_at": "2025-12-26T10:00:00Z"
  }
}
```

#### Remove User
```http
DELETE /api/tenants/{id}/users/{userId}
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": {
    "message": "user removed successfully"
  }
}
```

#### Get Pending Invitations
```http
GET /api/tenants/{id}/invitations
Authorization: Bearer <token>

Response: 200 OK
{
  "success": true,
  "data": [
    {
      "id": "uuid",
      "tenant_id": "uuid",
      "email": "invited@example.com",
      "role": "user",
      "expires_at": "2025-12-26T10:00:00Z"
    }
  ]
}
```

### Health Checks

```http
GET /health
GET /health/ready
```

## Database Schema

### Tables Used

- **tenants** - Tenant information
- **tenant_users** - User membership in tenants
- **tenant_invitations** - Pending user invitations
- **tenant_settings** - Tenant-specific settings

## Configuration

Environment variables:

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=10001  # Overridden in main.go

# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<password>
DB_NAME=docmanager

# Redis
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=<password>

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
```

## Running Locally

### Prerequisites

- Go 1.21+
- PostgreSQL running on port 15432
- Redis running on port 16379
- Database migrations applied

### Run Service

```bash
# From backend directory
cd services/tenant-service

# Run with environment variables
export DB_PASSWORD=testpassword12345678
export REDIS_PASSWORD=testpassword12345678
export INTERNAL_API_SECRET=testinternalapisecret1234567890123456

go run cmd/main.go
```

### Test Endpoints

```bash
# Health check
curl http://localhost:10001/health

# Create tenant (requires Oathkeeper headers)
curl -X POST http://localhost:10001/api/tenants \
  -H "Content-Type: application/json" \
  -H "X-User-ID: test-user-123" \
  -H "X-User-Email: test@example.com" \
  -d '{
    "name": "Test Tenant",
    "slug": "test-tenant"
  }'
```

## Docker Build

```bash
# From backend directory
docker build -t tenant-service:latest -f services/tenant-service/Dockerfile .

# Run container
docker run -p 10001:10001 \
  -e DB_HOST=host.docker.internal \
  -e REDIS_HOST=host.docker.internal \
  -e DB_PASSWORD=testpassword12345678 \
  -e REDIS_PASSWORD=testpassword12345678 \
  tenant-service:latest
```

## Business Rules

### Tenant Creation
- Slug must be unique
- Slug is normalized to lowercase
- Creator becomes tenant owner with admin role
- Default subscription plan is "free"
- Reserved slugs: admin, api, www, app, dashboard, system, internal

### User Management
- Only admins can invite users
- Only admins can remove users
- Owners cannot be removed
- Users cannot remove themselves
- Email addresses are stored in lowercase

### Invitations
- Invitations expire after 7 days
- Duplicate invitations for same email are prevented
- Invitations are single-use
- TODO: Email notifications (integration with notification service)

## Security

### Authentication
- All API endpoints (except health checks) require Oathkeeper headers
- User identity extracted from `X-User-ID` header
- User email extracted from `X-User-Email` header

### Authorization
- Users can only access tenants they belong to
- Only admins can modify tenant settings
- Only admins can manage users and invitations
- Owners have permanent admin privileges

## Caching Strategy

- Tenant details cached for 1 hour
- Cache key format: `tenant:<tenant_id>`
- Cache invalidated on updates
- Failed cache reads fall back to database

## Error Handling

All errors follow standardized format:

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Validation failed",
    "fields": {
      "slug": "slug is already taken"
    }
  }
}
```

Error codes:
- `VALIDATION_ERROR` - Input validation failed
- `NOT_FOUND` - Resource not found
- `UNAUTHORIZED` - Authentication required
- `FORBIDDEN` - Insufficient permissions
- `CONFLICT` - Resource conflict (duplicate slug, etc.)
- `INTERNAL_ERROR` - Server error

## Logging

All logs include:
- Request ID (correlation)
- User ID (when available)
- Tenant ID (when available)
- Timestamp
- Log level
- Structured fields

Example log:
```json
{
  "level": "info",
  "timestamp": "2025-12-19T10:00:00Z",
  "request_id": "uuid",
  "user_id": "kratos-user-id",
  "tenant_id": "uuid",
  "message": "tenant created",
  "slug": "acme-corp"
}
```

## Monitoring

### Metrics
- HTTP request duration
- Database query duration
- Cache hit/miss rate
- Error rates by endpoint

### Health Checks
- `/health` - Liveness probe
- `/health/ready` - Readiness probe (checks DB and Redis)

## Development

### Adding New Endpoints

1. Add model in `internal/models/models.go`
2. Add repository method in `internal/repository/repository.go`
3. Add service method in `internal/service/service.go`
4. Add handler in `internal/handler/handler.go`
5. Register route in `cmd/main.go`

### Testing

```bash
# Run tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestCreateTenant ./internal/service
```

## Dependencies

- Shared packages from `backend/pkg/`
- PostgreSQL database
- Redis cache
- Ory Oathkeeper (API Gateway)

## Future Enhancements

- [ ] Tenant settings management
- [ ] Subscription plan upgrades
- [ ] Usage analytics
- [ ] Tenant suspension/deactivation
- [ ] Email notifications for invitations
- [ ] Invitation acceptance flow
- [ ] Audit logging integration
- [ ] Metrics and monitoring

---

**Port:** 10001
**Status:** Production Ready
**Last Updated:** 2025-12-19
