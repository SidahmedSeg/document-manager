# Ory Kratos/Hydra Integration Guide
## For New Applications Using Unified Authentication

This guide provides everything you need to integrate your Document Library App (or any new application) with the existing Ory Kratos/Hydra unified authentication system.

---

## Table of Contents

1. [Architecture Overview](#architecture-overview)
2. [Docker Services Setup](#docker-services-setup)
3. [Kratos Configuration](#kratos-configuration)
4. [Backend Integration](#backend-integration)
5. [Frontend Integration](#frontend-integration)
6. [API Endpoints](#api-endpoints)
7. [Environment Variables](#environment-variables)
8. [Testing](#testing)
9. [Production Deployment](#production-deployment)

---

## Architecture Overview

### Components

- **Ory Kratos** (Port 4433/4434) - Identity and user management
- **Ory Hydra** (Port 4444/4445) - OAuth 2.0 and OpenID Connect provider
- **PostgreSQL** - Stores user identities, sessions, and OAuth clients
- **MailSlurper** (Dev only) - Email testing server
- **Your Backend** - Proxies Kratos flows and validates sessions
- **Your Frontend** - Initiates authentication flows

### Authentication Flow

```
┌─────────────┐          ┌─────────────┐          ┌─────────────┐
│             │          │             │          │             │
│  Frontend   │─────────▶│   Backend   │─────────▶│   Kratos    │
│             │  1. Init │   (Proxy)   │  2. Get  │             │
│             │    Flow  │             │    Flow  │             │
└─────────────┘          └─────────────┘          └─────────────┘
       │                        │                        │
       │                        │                        │
       │ 3. Submit              │ 4. Submit              │ 5. Create
       │    Form                │    to Kratos           │    Session
       │                        │                        │
       └───────────────────────▶│───────────────────────▶│
                                │                        │
                                │◀───────────────────────┘
                                │ 6. Return session cookie
                                │
                                │ 7. Set cookie in browser
                                ▼
```

**Key Points:**
- Frontend never talks to Kratos directly (security best practice)
- Backend proxies all Kratos flows
- Sessions are cookie-based (httpOnly, secure in production)
- Session cookies are automatically validated on protected endpoints

---

## Docker Services Setup

### 1. Add to your `docker-compose.yml`

```yaml
services:
  # PostgreSQL - Shared database for Kratos and your app
  postgres:
    image: postgres:15-alpine
    container_name: doclib_postgres
    ports:
      - "5434:5432"
    environment:
      - POSTGRES_DB=doclib_app
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped
    networks:
      - doclib_network
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Ory Kratos Migration
  kratos-migrate:
    image: oryd/kratos:v1.1.0
    container_name: doclib_kratos_migrate
    environment:
      - DSN=postgres://postgres:postgres@postgres:5432/kratos_db?sslmode=disable
    volumes:
      - ./ory/kratos:/etc/config/kratos
    command: -c /etc/config/kratos/kratos.yml migrate sql -e --yes
    networks:
      - doclib_network
    depends_on:
      postgres:
        condition: service_healthy

  # Ory Kratos - Identity Management
  kratos:
    image: oryd/kratos:v1.1.0
    container_name: doclib_kratos
    ports:
      - "4433:4433" # Public API
      - "4434:4434" # Admin API
    environment:
      - DSN=postgres://postgres:postgres@postgres:5432/kratos_db?sslmode=disable
      - LOG_LEVEL=debug
      - SERVE_PUBLIC_BASE_URL=http://127.0.0.1:4433/
      - SERVE_ADMIN_BASE_URL=http://127.0.0.1:4434/
    volumes:
      - ./ory/kratos:/etc/config/kratos
    command: serve -c /etc/config/kratos/kratos.yml --dev --watch-courier
    restart: unless-stopped
    networks:
      - doclib_network
    depends_on:
      - kratos-migrate

  # Ory Hydra Migration
  hydra-migrate:
    image: oryd/hydra:v2.2.0
    container_name: doclib_hydra_migrate
    environment:
      - DSN=postgres://postgres:postgres@postgres:5432/doclib_app?sslmode=disable
    command: migrate sql -e --yes
    networks:
      - doclib_network
    depends_on:
      postgres:
        condition: service_healthy

  # Ory Hydra - OAuth2/OIDC Provider
  hydra:
    image: oryd/hydra:v2.2.0
    container_name: doclib_hydra
    ports:
      - "4444:4444" # Public API
      - "4445:4445" # Admin API
    environment:
      - DSN=postgres://postgres:postgres@postgres:5432/doclib_app?sslmode=disable
      - URLS_SELF_ISSUER=http://127.0.0.1:4444/
      - URLS_LOGIN=http://127.0.0.1:3000/api/hydra/login
      - URLS_CONSENT=http://127.0.0.1:3000/api/hydra/consent
      - SECRETS_SYSTEM=CHANGE-THIS-IN-PRODUCTION-MINIMUM-32-CHARACTERS-LONG
      - OIDC_SUBJECT_IDENTIFIERS_SUPPORTED_TYPES=public
      - OAUTH2_EXPOSE_INTERNAL_ERRORS=true
    command: serve all --dev
    restart: unless-stopped
    networks:
      - doclib_network
    depends_on:
      - hydra-migrate
      - kratos

  # MailSlurper - Email Testing (Development only)
  mailslurper:
    image: oryd/mailslurper:latest-smtps
    container_name: doclib_mailslurper
    ports:
      - "4436:4436" # Web UI
      - "4437:4437" # SMTP
    networks:
      - doclib_network

volumes:
  postgres_data:
    driver: local

networks:
  doclib_network:
    driver: bridge
```

---

## Kratos Configuration

### 2. Create `ory/kratos/kratos.yml`

```yaml
version: v1.1.0

dsn: postgres://postgres:postgres@postgres:5432/kratos_db?sslmode=disable

serve:
  public:
    base_url: http://127.0.0.1:4433/
    cors:
      enabled: true
      allowed_origins:
        - http://127.0.0.1:3001  # Your Document Library frontend
        - http://localhost:3001
      allowed_methods:
        - POST
        - GET
        - PUT
        - PATCH
        - DELETE
      allowed_headers:
        - Authorization
        - Content-Type
        - Cookie
        - X-Session-Token
      exposed_headers:
        - Content-Type
        - Set-Cookie
      allow_credentials: true
  admin:
    base_url: http://127.0.0.1:4434/

cookies:
  domain: 127.0.0.1
  path: /
  same_site: Lax

selfservice:
  default_browser_return_url: http://127.0.0.1:3001/
  allowed_return_urls:
    - http://127.0.0.1:3001
    - http://localhost:3001

  methods:
    password:
      enabled: true
      config:
        min_password_length: 8
        max_breaches: 100  # Allow breached passwords in dev
        ignore_network_errors: true
    oidc:
      enabled: false # Enable later for social login

  flows:
    error:
      ui_url: http://127.0.0.1:3001/auth/error

    settings:
      ui_url: http://127.0.0.1:3001/auth/settings
      privileged_session_max_age: 15m

    recovery:
      enabled: true
      ui_url: http://127.0.0.1:3001/auth/recovery
      use: code

    verification:
      enabled: true
      ui_url: http://127.0.0.1:3001/auth/verification
      use: code
      after:
        default_browser_return_url: http://127.0.0.1:3001/auth/verified

    logout:
      after:
        default_browser_return_url: http://127.0.0.1:3001/

    login:
      ui_url: http://127.0.0.1:3001/auth/login
      lifespan: 10m
      after:
        password:
          default_browser_return_url: http://127.0.0.1:3001/auth/callback

    registration:
      lifespan: 10m
      ui_url: http://127.0.0.1:3001/auth/signup
      after:
        password:
          hooks:
            - hook: session
            - hook: show_verification_ui
            - hook: web_hook
              config:
                url: http://doclib-backend:3000/internal/auth/user-created
                method: POST
                body: file:///etc/config/kratos/webhooks/user-created.jsonnet
                response:
                  ignore: false
                  parse: false
          default_browser_return_url: http://127.0.0.1:3001/auth/verify-email

log:
  level: debug
  format: text
  leak_sensitive_values: true

secrets:
  cookie:
    - PLEASE-CHANGE-ME-I-AM-VERY-INSECURE-FOR-DEVELOPMENT-ONLY
  cipher:
    - DEVXXXXXXXXXXXXXXXXXXXXXXXXXXXXX

ciphers:
  algorithm: xchacha20-poly1305

hashers:
  algorithm: argon2
  argon2:
    memory: "128MB"
    iterations: 3
    parallelism: 4
    salt_length: 16
    key_length: 32

identity:
  default_schema_id: default
  schemas:
    - id: default
      url: file:///etc/config/kratos/identity.schema.json

courier:
  smtp:
    connection_uri: smtps://test:test@mailslurper:1025/?skip_ssl_verify=true
  templates:
    verification:
      valid:
        email:
          body:
            html: file:///etc/config/kratos/email.templates/verification/valid/email.body.gotmpl
    recovery:
      valid:
        email:
          body:
            html: file:///etc/config/kratos/email.templates/recovery/valid/email.body.gotmpl
```

### 3. Create `ory/kratos/identity.schema.json`

**IMPORTANT:** Customize this schema for your Document Library App's user fields.

```json
{
  "$id": "https://your-domain.com/schemas/identity.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Document Library User",
  "type": "object",
  "properties": {
    "traits": {
      "type": "object",
      "properties": {
        "email": {
          "type": "string",
          "format": "email",
          "title": "Email Address",
          "minLength": 3,
          "maxLength": 320,
          "ory.sh/kratos": {
            "credentials": {
              "password": {
                "identifier": true
              }
            },
            "verification": {
              "via": "email"
            },
            "recovery": {
              "via": "email"
            }
          }
        },
        "username": {
          "type": "string",
          "title": "Username",
          "minLength": 3,
          "maxLength": 30,
          "pattern": "^[a-z0-9._]+$"
        },
        "first_name": {
          "type": "string",
          "title": "First Name",
          "minLength": 1,
          "maxLength": 100
        },
        "last_name": {
          "type": "string",
          "title": "Last Name",
          "minLength": 1,
          "maxLength": 100
        }
      },
      "required": [
        "email",
        "username",
        "first_name",
        "last_name"
      ],
      "additionalProperties": false
    }
  }
}
```

### 4. Create `ory/kratos/webhooks/user-created.jsonnet`

This webhook payload is sent to your backend when a user registers:

```jsonnet
function(ctx) {
  identity: {
    id: ctx.identity.id,
    traits: {
      email: ctx.identity.traits.email,
      first_name: if std.objectHas(ctx.identity.traits, 'first_name') then ctx.identity.traits.first_name else '',
      last_name: if std.objectHas(ctx.identity.traits, 'last_name') then ctx.identity.traits.last_name else '',
      username: if std.objectHas(ctx.identity.traits, 'username') then ctx.identity.traits.username else '',
    },
    created_at: ctx.identity.created_at,
    updated_at: ctx.identity.updated_at,
  },
}
```

---

## Backend Integration

### 5. Backend API Structure (Any Language)

Your backend needs to:
1. **Proxy Kratos flows** (registration, login, logout)
2. **Validate sessions** using Kratos whoami endpoint
3. **Protect routes** with authentication middleware

#### Example: Rust with Axum

**Dependencies (Cargo.toml):**
```toml
[dependencies]
axum = "0.7"
reqwest = { version = "0.11", features = ["json", "cookies"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1.0", features = ["full"] }
uuid = { version = "1.0", features = ["serde", "v4"] }
anyhow = "1.0"
tracing = "0.1"
```

**Kratos Client (`src/kratos_client.rs`):**

```rust
use anyhow::Result;
use reqwest::Client;
use serde::{Deserialize, Serialize};
use uuid::Uuid;

#[derive(Clone)]
pub struct KratosClient {
    client: Client,
    public_url: String,
    admin_url: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KratosSession {
    pub id: String,
    pub active: bool,
    pub identity: KratosIdentity,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct KratosIdentity {
    pub id: Uuid,
    pub traits: IdentityTraits,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct IdentityTraits {
    pub email: String,
    pub first_name: String,
    pub last_name: String,
    pub username: String,
}

impl KratosClient {
    pub fn new(public_url: String, admin_url: String) -> Self {
        Self {
            client: Client::new(),
            public_url,
            admin_url,
        }
    }

    /// Validate session using whoami endpoint
    pub async fn whoami(&self, cookie_header: &str) -> Result<KratosSession> {
        let session_token = cookie_header
            .split(';')
            .find_map(|cookie| {
                let cookie = cookie.trim();
                if cookie.starts_with("ory_kratos_session=") {
                    Some(cookie.trim_start_matches("ory_kratos_session="))
                } else {
                    None
                }
            })
            .ok_or_else(|| anyhow::anyhow!("No session cookie found"))?;

        let response = self
            .client
            .get(format!("{}/sessions/whoami", self.public_url))
            .header("X-Session-Token", session_token)
            .send()
            .await?;

        if !response.status().is_success() {
            anyhow::bail!("Session validation failed: {}", response.status());
        }

        let session: KratosSession = response.json().await?;
        Ok(session)
    }

    /// Initialize registration flow
    pub async fn init_registration_flow(&self) -> Result<serde_json::Value> {
        let url = format!("{}/self-service/registration/api", self.public_url);
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            anyhow::bail!("Registration flow init failed: {}", response.status());
        }

        Ok(response.json().await?)
    }

    /// Submit registration data
    pub async fn submit_registration(
        &self,
        flow_id: &str,
        email: &str,
        password: &str,
        first_name: &str,
        last_name: &str,
        username: &str,
    ) -> Result<(serde_json::Value, Vec<String>)> {
        let url = format!("{}/self-service/registration?flow={}", self.public_url, flow_id);

        let response = self.client
            .post(&url)
            .json(&serde_json::json!({
                "method": "password",
                "traits": {
                    "email": email,
                    "first_name": first_name,
                    "last_name": last_name,
                    "username": username
                },
                "password": password
            }))
            .send()
            .await?;

        // Extract Set-Cookie headers
        let cookies: Vec<String> = response.headers()
            .get_all("set-cookie")
            .iter()
            .filter_map(|h| h.to_str().ok())
            .map(|s| s.to_string())
            .collect();

        if !response.status().is_success() {
            let error_body = response.text().await?;
            anyhow::bail!("Registration failed: {}", error_body);
        }

        let body = response.json().await?;
        Ok((body, cookies))
    }

    /// Initialize login flow
    pub async fn init_login_flow(&self) -> Result<serde_json::Value> {
        let url = format!("{}/self-service/login/api", self.public_url);
        let response = self.client.get(&url).send().await?;

        if !response.status().is_success() {
            anyhow::bail!("Login flow init failed: {}", response.status());
        }

        Ok(response.json().await?)
    }

    /// Submit login credentials
    pub async fn submit_login(
        &self,
        flow_id: &str,
        identifier: &str,
        password: &str,
    ) -> Result<(serde_json::Value, Vec<String>)> {
        let url = format!("{}/self-service/login?flow={}", self.public_url, flow_id);

        let response = self.client
            .post(&url)
            .json(&serde_json::json!({
                "method": "password",
                "identifier": identifier,
                "password": password
            }))
            .send()
            .await?;

        let cookies: Vec<String> = response.headers()
            .get_all("set-cookie")
            .iter()
            .filter_map(|h| h.to_str().ok())
            .map(|s| s.to_string())
            .collect();

        if !response.status().is_success() {
            let error_body = response.text().await?;
            anyhow::bail!("Login failed: {}", error_body);
        }

        let body = response.json().await?;
        Ok((body, cookies))
    }

    /// Initialize logout flow
    pub async fn init_logout_flow(&self, cookie: Option<&str>) -> Result<String> {
        let url = format!("{}/self-service/logout/browser", self.public_url);
        let mut request = self.client.get(&url);

        if let Some(cookie_header) = cookie {
            request = request.header("Cookie", cookie_header);
        }

        let response = request.send().await?;

        if !response.status().is_success() {
            anyhow::bail!("Logout flow init failed: {}", response.status());
        }

        let logout_data: serde_json::Value = response.json().await?;
        let logout_url = logout_data["logout_url"]
            .as_str()
            .ok_or_else(|| anyhow::anyhow!("No logout_url in response"))?
            .to_string();

        Ok(logout_url)
    }
}
```

**Authentication Middleware (`src/auth_middleware.rs`):**

```rust
use axum::{
    extract::{Request, State},
    http::{StatusCode, HeaderMap},
    middleware::Next,
    response::{IntoResponse, Response},
    Json,
};
use std::sync::Arc;

use crate::kratos_client::{KratosClient, KratosSession};

/// Extension for storing Kratos session in request
#[derive(Clone)]
pub struct AuthSession(pub KratosSession);

/// Middleware to require authentication
pub async fn require_auth(
    State(kratos): State<Arc<KratosClient>>,
    headers: HeaderMap,
    mut request: Request,
    next: Next,
) -> Response {
    let cookie_header = match headers.get("cookie") {
        Some(h) => h.to_str().unwrap_or(""),
        None => {
            return (StatusCode::UNAUTHORIZED, Json(serde_json::json!({
                "error": "Authentication required"
            }))).into_response();
        }
    };

    match kratos.whoami(cookie_header).await {
        Ok(session) if session.active => {
            request.extensions_mut().insert(AuthSession(session));
            next.run(request).await
        }
        _ => {
            (StatusCode::UNAUTHORIZED, Json(serde_json::json!({
                "error": "Invalid session"
            }))).into_response()
        }
    }
}
```

**API Endpoints (`src/api.rs`):**

```rust
use axum::{
    extract::{Extension, State},
    http::{StatusCode, HeaderMap},
    routing::{get, post},
    Json, Router,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;

use crate::kratos_client::KratosClient;
use crate::auth_middleware::{AuthSession, require_auth};

#[derive(Deserialize)]
pub struct RegistrationRequest {
    flow_id: String,
    email: String,
    password: String,
    first_name: String,
    last_name: String,
    username: String,
}

#[derive(Deserialize)]
pub struct LoginRequest {
    flow_id: String,
    identifier: String,
    password: String,
}

pub fn create_router(kratos: Arc<KratosClient>) -> Router {
    Router::new()
        // Public auth endpoints
        .route("/api/auth/flows/registration", get(init_registration).post(submit_registration))
        .route("/api/auth/flows/login", get(init_login).post(submit_login))
        .route("/api/auth/flows/logout", get(init_logout))
        .route("/api/auth/whoami", get(whoami))
        // Protected endpoints
        .route("/api/documents", get(get_documents).layer(axum::middleware::from_fn_with_state(kratos.clone(), require_auth)))
        .with_state(kratos)
}

// Registration flow init
async fn init_registration(State(kratos): State<Arc<KratosClient>>) -> Result<Json<serde_json::Value>, StatusCode> {
    kratos.init_registration_flow().await
        .map(|flow| Json(serde_json::json!({"success": true, "data": flow})))
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

// Registration flow submit
async fn submit_registration(
    State(kratos): State<Arc<KratosClient>>,
    Json(req): Json<RegistrationRequest>,
) -> Result<(StatusCode, HeaderMap, Json<serde_json::Value>), StatusCode> {
    match kratos.submit_registration(
        &req.flow_id,
        &req.email,
        &req.password,
        &req.first_name,
        &req.last_name,
        &req.username,
    ).await {
        Ok((body, cookies)) => {
            let mut headers = HeaderMap::new();
            for cookie in cookies {
                if let Ok(value) = cookie.parse() {
                    headers.append("set-cookie", value);
                }
            }
            Ok((StatusCode::OK, headers, Json(serde_json::json!({
                "success": true,
                "data": body
            }))))
        }
        Err(e) => {
            eprintln!("Registration error: {}", e);
            Err(StatusCode::BAD_REQUEST)
        }
    }
}

// Login flow init
async fn init_login(State(kratos): State<Arc<KratosClient>>) -> Result<Json<serde_json::Value>, StatusCode> {
    kratos.init_login_flow().await
        .map(|flow| Json(serde_json::json!({"success": true, "data": flow})))
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

// Login flow submit
async fn submit_login(
    State(kratos): State<Arc<KratosClient>>,
    Json(req): Json<LoginRequest>,
) -> Result<(StatusCode, HeaderMap, Json<serde_json::Value>), StatusCode> {
    match kratos.submit_login(&req.flow_id, &req.identifier, &req.password).await {
        Ok((body, cookies)) => {
            let mut headers = HeaderMap::new();
            for cookie in cookies {
                if let Ok(value) = cookie.parse() {
                    headers.append("set-cookie", value);
                }
            }
            Ok((StatusCode::OK, headers, Json(serde_json::json!({
                "success": true,
                "data": body
            }))))
        }
        Err(e) => {
            eprintln!("Login error: {}", e);
            Err(StatusCode::UNAUTHORIZED)
        }
    }
}

// Logout flow init
async fn init_logout(
    State(kratos): State<Arc<KratosClient>>,
    headers: HeaderMap,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let cookie = headers.get("cookie").and_then(|h| h.to_str().ok());
    kratos.init_logout_flow(cookie).await
        .map(|logout_url| Json(serde_json::json!({
            "success": true,
            "data": {"logout_url": logout_url}
        })))
        .map_err(|_| StatusCode::INTERNAL_SERVER_ERROR)
}

// Whoami endpoint
async fn whoami(
    State(kratos): State<Arc<KratosClient>>,
    headers: HeaderMap,
) -> Result<Json<serde_json::Value>, StatusCode> {
    let cookie_header = headers.get("cookie")
        .and_then(|h| h.to_str().ok())
        .ok_or(StatusCode::UNAUTHORIZED)?;

    kratos.whoami(cookie_header).await
        .map(|session| Json(serde_json::json!({
            "success": true,
            "data": {
                "id": session.identity.id,
                "email": session.identity.traits.email,
                "first_name": session.identity.traits.first_name,
                "last_name": session.identity.traits.last_name,
                "authenticated": session.active
            }
        })))
        .map_err(|_| StatusCode::UNAUTHORIZED)
}

// Example protected endpoint
async fn get_documents(Extension(session): Extension<AuthSession>) -> Json<serde_json::Value> {
    Json(serde_json::json!({
        "success": true,
        "data": {
            "user_id": session.0.identity.id,
            "documents": []
        }
    }))
}
```

---

## Frontend Integration

### 6. TypeScript/JavaScript Kratos API Client

Create `src/lib/kratos.ts`:

```typescript
import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:3000';

export interface KratosSession {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  authenticated: boolean;
}

export interface RegistrationData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  username: string;
}

export interface LoginData {
  identifier: string;
  password: string;
}

// Initialize registration flow
export async function initRegistrationFlow() {
  const response = await axios.get(`${API_BASE_URL}/api/auth/flows/registration`);
  return response.data.data;
}

// Submit registration
export async function submitRegistration(flowId: string, data: RegistrationData) {
  await axios.post(
    `${API_BASE_URL}/api/auth/flows/registration`,
    { flow_id: flowId, ...data },
    { withCredentials: true }
  );
}

// Initialize login flow
export async function initLoginFlow() {
  const response = await axios.get(`${API_BASE_URL}/api/auth/flows/login`);
  return response.data.data;
}

// Submit login
export async function submitLogin(flowId: string, data: LoginData) {
  await axios.post(
    `${API_BASE_URL}/api/auth/flows/login`,
    { flow_id: flowId, ...data },
    { withCredentials: true }
  );
}

// Get current session
export async function whoami(): Promise<KratosSession> {
  const response = await axios.get(`${API_BASE_URL}/api/auth/whoami`, {
    withCredentials: true
  });
  return response.data.data;
}

// Logout
export async function logout() {
  const response = await axios.get(`${API_BASE_URL}/api/auth/flows/logout`, {
    withCredentials: true
  });
  const logoutUrl = response.data.data.logout_url;
  await axios.get(logoutUrl, { withCredentials: true });
}
```

### 7. Example React/Svelte Components

**Registration Form (React):**

```typescript
import { useState } from 'react';
import { initRegistrationFlow, submitRegistration } from './lib/kratos';

export function RegisterForm() {
  const [flowId, setFlowId] = useState('');
  const [formData, setFormData] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    username: '',
  });

  useEffect(() => {
    initRegistrationFlow().then(flow => setFlowId(flow.id));
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await submitRegistration(flowId, formData);
      window.location.href = '/dashboard';
    } catch (error) {
      console.error('Registration failed:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        placeholder="Email"
        value={formData.email}
        onChange={(e) => setFormData({ ...formData, email: e.target.value })}
      />
      <input
        type="password"
        placeholder="Password"
        value={formData.password}
        onChange={(e) => setFormData({ ...formData, password: e.target.value })}
      />
      <input
        type="text"
        placeholder="First Name"
        value={formData.first_name}
        onChange={(e) => setFormData({ ...formData, first_name: e.target.value })}
      />
      <input
        type="text"
        placeholder="Last Name"
        value={formData.last_name}
        onChange={(e) => setFormData({ ...formData, last_name: e.target.value })}
      />
      <input
        type="text"
        placeholder="Username"
        value={formData.username}
        onChange={(e) => setFormData({ ...formData, username: e.target.value })}
      />
      <button type="submit">Register</button>
    </form>
  );
}
```

**Login Form (React):**

```typescript
import { useState, useEffect } from 'react';
import { initLoginFlow, submitLogin } from './lib/kratos';

export function LoginForm() {
  const [flowId, setFlowId] = useState('');
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');

  useEffect(() => {
    initLoginFlow().then(flow => setFlowId(flow.id));
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      await submitLogin(flowId, { identifier, password });
      window.location.href = '/dashboard';
    } catch (error) {
      console.error('Login failed:', error);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        placeholder="Email or Username"
        value={identifier}
        onChange={(e) => setIdentifier(e.target.value)}
      />
      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
      />
      <button type="submit">Login</button>
    </form>
  );
}
```

---

## API Endpoints

Your backend should expose these endpoints:

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/auth/flows/registration` | Initialize registration flow | No |
| POST | `/api/auth/flows/registration` | Submit registration | No |
| GET | `/api/auth/flows/login` | Initialize login flow | No |
| POST | `/api/auth/flows/login` | Submit login | No |
| GET | `/api/auth/flows/logout` | Initialize logout | Yes |
| GET | `/api/auth/whoami` | Get current user | Yes |
| POST | `/internal/auth/user-created` | Webhook for new users | No (internal) |

---

## Environment Variables

### Backend `.env`

```bash
# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=3000

# Kratos
KRATOS_PUBLIC_URL=http://kratos:4433
KRATOS_ADMIN_URL=http://kratos:4434

# Database
DATABASE_URL=postgresql://postgres:postgres@postgres:5432/doclib_app

# Hydra (if using OAuth)
HYDRA_PUBLIC_URL=http://hydra:4444
HYDRA_ADMIN_URL=http://hydra:4445
```

### Frontend `.env`

```bash
VITE_API_URL=http://localhost:3000
```

---

## Testing

### 1. Start Services

```bash
docker-compose up -d postgres kratos hydra mailslurper
```

### 2. Verify Kratos is Running

```bash
curl http://127.0.0.1:4433/health/ready
# Should return: {"status":"ok"}
```

### 3. Test Registration Flow

```bash
# Get registration flow
curl http://localhost:3000/api/auth/flows/registration

# Submit registration (use flow ID from above)
curl -X POST http://localhost:3000/api/auth/flows/registration \
  -H "Content-Type: application/json" \
  -d '{
    "flow_id": "FLOW_ID_HERE",
    "email": "test@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe"
  }'
```

### 4. Check Email Verification

Open MailSlurper UI: http://127.0.0.1:4436

### 5. Test Login

```bash
# Get login flow
curl http://localhost:3000/api/auth/flows/login

# Submit login (save cookies)
curl -X POST http://localhost:3000/api/auth/flows/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d '{
    "flow_id": "FLOW_ID_HERE",
    "identifier": "test@example.com",
    "password": "password123"
  }'
```

### 6. Test Protected Endpoint

```bash
curl http://localhost:3000/api/documents \
  -b cookies.txt
```

---

## Production Deployment

### Security Checklist

- [ ] Change `SECRETS_SYSTEM` in Hydra to strong random value (min 32 chars)
- [ ] Change `secrets.cookie` in Kratos to strong random value
- [ ] Change `secrets.cipher` in Kratos to strong random value
- [ ] Set `max_breaches: 0` in Kratos password config
- [ ] Use real SMTP server (not MailSlurper)
- [ ] Enable HTTPS/TLS for all services
- [ ] Set `cookies.secure: true` in Kratos
- [ ] Update CORS allowed origins to production domains
- [ ] Set `LOG_LEVEL=info` (not debug)
- [ ] Set `leak_sensitive_values: false` in Kratos
- [ ] Use strong PostgreSQL passwords
- [ ] Enable database backups
- [ ] Set up monitoring and alerting

### Nginx Configuration

```nginx
# Frontend
server {
    listen 443 ssl http2;
    server_name docs.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://doclib-frontend:3001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Backend API
server {
    listen 443 ssl http2;
    server_name api.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://doclib-backend:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Cookie $http_cookie;
    }
}

# Kratos (optional - for direct access)
server {
    listen 443 ssl http2;
    server_name auth.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    location / {
        proxy_pass http://kratos:4433;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Cookie $http_cookie;
    }
}
```

---

## Additional Resources

- **Ory Kratos Documentation:** https://www.ory.sh/docs/kratos/
- **Ory Hydra Documentation:** https://www.ory.sh/docs/hydra/
- **Identity Schema Guide:** https://www.ory.sh/docs/kratos/manage-identities/identity-schema
- **Self-Service Flows:** https://www.ory.sh/docs/kratos/self-service

---

## Summary

You now have everything needed to integrate your Document Library App with the unified Ory Kratos/Hydra authentication system:

1. ✅ Docker services configuration
2. ✅ Kratos configuration files
3. ✅ Backend integration code (Rust example)
4. ✅ Frontend integration code (TypeScript)
5. ✅ API endpoints structure
6. ✅ Environment variables
7. ✅ Testing procedures
8. ✅ Production deployment guide

**Next Steps:**
1. Copy the configuration files to your Document Library App
2. Implement the backend API endpoints
3. Integrate the frontend Kratos client
4. Test registration and login flows
5. Deploy to production with security hardening

If you need help with any specific language/framework integration (Node.js, Python, Go, etc.), let me know!
