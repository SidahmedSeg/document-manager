# Ory Integration Quick Reference

## Essential URLs

| Service | Port | URL (Development) |
|---------|------|-------------------|
| Kratos Public | 4433 | http://127.0.0.1:4433 |
| Kratos Admin | 4434 | http://127.0.0.1:4434 |
| Hydra Public | 4444 | http://127.0.0.1:4444 |
| Hydra Admin | 4445 | http://127.0.0.1:4445 |
| MailSlurper UI | 4436 | http://127.0.0.1:4436 |

## Required Environment Variables

```bash
# Backend
KRATOS_PUBLIC_URL=http://kratos:4433
KRATOS_ADMIN_URL=http://kratos:4434
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/your_db

# Frontend
VITE_API_URL=http://localhost:3000
```

## Required Configuration Files

```
your-project/
├── docker-compose.yml          # Add Kratos/Hydra services
└── ory/
    └── kratos/
        ├── kratos.yml          # Main Kratos config
        ├── identity.schema.json # User fields schema
        └── webhooks/
            └── user-created.jsonnet # Webhook payload
```

## API Endpoints to Implement

```
GET  /api/auth/flows/registration    # Init registration
POST /api/auth/flows/registration    # Submit registration
GET  /api/auth/flows/login           # Init login
POST /api/auth/flows/login           # Submit login
GET  /api/auth/flows/logout          # Init logout
GET  /api/auth/whoami                # Get current user
POST /internal/auth/user-created     # User creation webhook
```

## Frontend Integration (5 functions)

```typescript
import {
  initRegistrationFlow,
  submitRegistration,
  initLoginFlow,
  submitLogin,
  whoami,
  logout
} from './lib/kratos';

// Registration
const flow = await initRegistrationFlow();
await submitRegistration(flow.id, {
  email: 'user@example.com',
  password: 'password',
  first_name: 'John',
  last_name: 'Doe',
  username: 'johndoe'
});

// Login
const loginFlow = await initLoginFlow();
await submitLogin(loginFlow.id, {
  identifier: 'user@example.com',
  password: 'password'
});

// Check session
const user = await whoami();

// Logout
await logout();
```

## Backend Session Validation

**Rust Example:**
```rust
// Extract session from cookie
let cookie_header = headers.get("cookie").unwrap();
let session = kratos.whoami(cookie_header).await?;

// Use session data
let user_id = session.identity.id;
let email = session.identity.traits.email;
```

**Node.js Example:**
```javascript
const session = await kratos.whoami(req.headers.cookie);
const userId = session.identity.id;
const email = session.identity.traits.email;
```

## Database Setup

Kratos requires a separate PostgreSQL database:

```sql
-- Kratos creates its own tables automatically
CREATE DATABASE kratos_db;
```

Your app uses its own database:

```sql
CREATE DATABASE your_app_db;
```

## Testing Checklist

- [ ] Start services: `docker-compose up -d`
- [ ] Check Kratos health: `curl http://127.0.0.1:4433/health/ready`
- [ ] Test registration flow
- [ ] Check email in MailSlurper (http://127.0.0.1:4436)
- [ ] Test login flow
- [ ] Test protected endpoint with session cookie
- [ ] Test logout flow

## Common Issues

### Issue: "No session cookie found"
**Solution:** Ensure `withCredentials: true` in axios config

### Issue: CORS errors
**Solution:** Add your frontend URL to Kratos `kratos.yml` CORS config

### Issue: Session not persisting
**Solution:** Check cookie domain matches your setup (127.0.0.1 vs localhost)

### Issue: Webhook not firing
**Solution:** Ensure webhook URL is reachable from Kratos container

## Production Security Checklist

- [ ] Change Kratos cookie secrets
- [ ] Change Hydra system secrets
- [ ] Set password `max_breaches: 0`
- [ ] Use real SMTP (not MailSlurper)
- [ ] Enable HTTPS/TLS
- [ ] Set `cookies.secure: true`
- [ ] Update CORS to production domains
- [ ] Set `LOG_LEVEL=info`
- [ ] Disable `leak_sensitive_values`

## Identity Schema Customization

Edit `ory/kratos/identity.schema.json` to add custom fields:

```json
{
  "traits": {
    "properties": {
      "email": { ... },
      "custom_field": {
        "type": "string",
        "title": "Custom Field"
      }
    },
    "required": ["email", "custom_field"]
  }
}
```

## Session Cookie Name

Kratos uses the cookie: `ory_kratos_session`

To validate, extract it and send as `X-Session-Token` header to Kratos.

## Useful Commands

```bash
# View Kratos logs
docker logs search_engine_kratos

# Check Kratos database
docker exec -it search_engine_postgres psql -U postgres -d kratos_db

# List identities (admin API)
curl http://127.0.0.1:4434/admin/identities

# Delete all sessions (testing)
docker exec -it search_engine_postgres psql -U postgres -d kratos_db \
  -c "TRUNCATE TABLE sessions CASCADE;"
```

## Support

For detailed implementation, see `ORY_INTEGRATION_GUIDE.md`
