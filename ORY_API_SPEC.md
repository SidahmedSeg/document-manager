# Ory Kratos API Specification
## Language-Agnostic REST API Reference

This document describes the exact HTTP requests and responses for integrating with Ory Kratos authentication, regardless of your backend language/framework.

---

## Base URLs

```
Backend API: http://localhost:3000
Kratos Public: http://127.0.0.1:4433
Kratos Admin: http://127.0.0.1:4434
```

---

## 1. Initialize Registration Flow

**Your Backend Endpoint:**
```
GET /api/auth/flows/registration
```

**What Your Backend Does:**
```http
GET http://127.0.0.1:4433/self-service/registration/api
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "a8f1d3c2-4b7e-4d8a-9c3f-5e6b7a8d9c2f",
    "type": "api",
    "expires_at": "2024-01-15T12:30:00Z",
    "issued_at": "2024-01-15T12:20:00Z",
    "request_url": "http://127.0.0.1:4433/self-service/registration/api",
    "ui": {
      "action": "http://127.0.0.1:4433/self-service/registration?flow=a8f1d3c2-4b7e-4d8a-9c3f-5e6b7a8d9c2f",
      "method": "POST",
      "nodes": [
        {
          "type": "input",
          "group": "default",
          "attributes": {
            "name": "csrf_token",
            "type": "hidden",
            "value": "csrf_token_value_here",
            "required": true
          }
        },
        {
          "type": "input",
          "group": "password",
          "attributes": {
            "name": "traits.email",
            "type": "email",
            "required": true
          }
        },
        {
          "type": "input",
          "group": "password",
          "attributes": {
            "name": "password",
            "type": "password",
            "required": true
          }
        }
      ]
    }
  }
}
```

**Important:** Save `data.id` as the `flow_id` for the next request.

---

## 2. Submit Registration

**Your Backend Endpoint:**
```
POST /api/auth/flows/registration
Content-Type: application/json
```

**Request Body:**
```json
{
  "flow_id": "a8f1d3c2-4b7e-4d8a-9c3f-5e6b7a8d9c2f",
  "email": "user@example.com",
  "password": "SecurePass123!",
  "first_name": "John",
  "last_name": "Doe",
  "username": "johndoe"
}
```

**What Your Backend Does:**
```http
POST http://127.0.0.1:4433/self-service/registration?flow=a8f1d3c2-4b7e-4d8a-9c3f-5e6b7a8d9c2f
Content-Type: application/json

{
  "method": "password",
  "traits": {
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe"
  },
  "password": "SecurePass123!"
}
```

**Kratos Response (200 OK):**
```
Headers:
  Set-Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx; Path=/; HttpOnly; SameSite=Lax
  Set-Cookie: csrf_token_xxxxxxxxxxx=...; Path=/; HttpOnly; SameSite=Lax

Body:
{
  "session": {
    "id": "session-uuid",
    "active": true,
    "expires_at": "2024-01-22T12:20:00Z",
    "authenticated_at": "2024-01-15T12:20:00Z",
    "identity": {
      "id": "identity-uuid",
      "schema_id": "default",
      "traits": {
        "email": "user@example.com",
        "first_name": "John",
        "last_name": "Doe",
        "username": "johndoe"
      }
    }
  }
}
```

**Your Backend Response:**
```http
HTTP/1.1 200 OK
Set-Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx; Path=/; HttpOnly; SameSite=Lax
Set-Cookie: csrf_token_xxxxxxxxxxx=...; Path=/; HttpOnly; SameSite=Lax
Content-Type: application/json

{
  "success": true,
  "data": {
    "session": { ... }
  }
}
```

**CRITICAL:** Forward ALL `Set-Cookie` headers from Kratos to the client.

---

## 3. Initialize Login Flow

**Your Backend Endpoint:**
```
GET /api/auth/flows/login
```

**What Your Backend Does:**
```http
GET http://127.0.0.1:4433/self-service/login/api
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "login-flow-uuid",
    "type": "api",
    "expires_at": "2024-01-15T12:30:00Z",
    "issued_at": "2024-01-15T12:20:00Z",
    "ui": {
      "action": "http://127.0.0.1:4433/self-service/login?flow=login-flow-uuid",
      "method": "POST",
      "nodes": [...]
    }
  }
}
```

---

## 4. Submit Login

**Your Backend Endpoint:**
```
POST /api/auth/flows/login
Content-Type: application/json
```

**Request Body:**
```json
{
  "flow_id": "login-flow-uuid",
  "identifier": "user@example.com",
  "password": "SecurePass123!"
}
```

**What Your Backend Does:**
```http
POST http://127.0.0.1:4433/self-service/login?flow=login-flow-uuid
Content-Type: application/json

{
  "method": "password",
  "identifier": "user@example.com",
  "password": "SecurePass123!"
}
```

**Kratos Response (200 OK):**
```
Headers:
  Set-Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx; Path=/; HttpOnly; SameSite=Lax
  Set-Cookie: csrf_token_xxxxxxxxxxx=...; Path=/; HttpOnly; SameSite=Lax

Body:
{
  "session": {
    "id": "session-uuid",
    "active": true,
    "identity": { ... }
  }
}
```

**Your Backend Response:**
```http
HTTP/1.1 200 OK
Set-Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx; Path=/; HttpOnly; SameSite=Lax
Set-Cookie: csrf_token_xxxxxxxxxxx=...; Path=/; HttpOnly; SameSite=Lax
Content-Type: application/json

{
  "success": true,
  "data": {
    "session": { ... }
  }
}
```

---

## 5. Validate Session (Whoami)

**Your Backend Endpoint:**
```
GET /api/auth/whoami
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx
```

**What Your Backend Does:**

1. Extract session token from cookie header:
```
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx; other=value
```

2. Call Kratos with `X-Session-Token` header:
```http
GET http://127.0.0.1:4433/sessions/whoami
X-Session-Token: ory_st_xxxxxxxxxxx
```

**Kratos Response (200 OK):**
```json
{
  "id": "session-uuid",
  "active": true,
  "expires_at": "2024-01-22T12:20:00Z",
  "authenticated_at": "2024-01-15T12:20:00Z",
  "identity": {
    "id": "identity-uuid",
    "schema_id": "default",
    "traits": {
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "username": "johndoe"
    },
    "verifiable_addresses": [
      {
        "id": "address-uuid",
        "value": "user@example.com",
        "verified": false,
        "via": "email",
        "status": "sent"
      }
    ]
  }
}
```

**Your Backend Response:**
```json
{
  "success": true,
  "data": {
    "id": "identity-uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "username": "johndoe",
    "authenticated": true
  }
}
```

---

## 6. Initialize Logout Flow

**Your Backend Endpoint:**
```
GET /api/auth/flows/logout
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx
```

**What Your Backend Does:**
```http
GET http://127.0.0.1:4433/self-service/logout/browser
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx
```

**Kratos Response (200 OK):**
```json
{
  "logout_url": "http://127.0.0.1:4433/self-service/logout?token=logout-token-here"
}
```

**Your Backend Response:**
```json
{
  "success": true,
  "data": {
    "logout_url": "http://127.0.0.1:4433/self-service/logout?token=logout-token-here"
  }
}
```

---

## 7. Complete Logout

**Your Frontend Does:**
```http
GET http://127.0.0.1:4433/self-service/logout?token=logout-token-here
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx
```

**Kratos Response (302 Found):**
```
Location: http://127.0.0.1:3001/
Set-Cookie: ory_kratos_session=; Path=/; Max-Age=0
```

---

## 8. Protected Endpoint (Example)

**Your Backend Endpoint:**
```
GET /api/documents
Cookie: ory_kratos_session=ory_st_xxxxxxxxxxx
```

**Validation Logic:**

```
1. Extract cookie header from request
2. Call Kratos whoami (see #5 above)
3. If session.active == true:
     Continue to handler
   Else:
     Return 401 Unauthorized
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "user_id": "identity-uuid",
    "documents": [...]
  }
}
```

**Response (401 Unauthorized):**
```json
{
  "success": false,
  "error": "Invalid session"
}
```

---

## 9. User Created Webhook (Internal)

**Kratos Calls Your Backend:**
```http
POST http://your-backend:3000/internal/auth/user-created
Content-Type: application/json

{
  "identity": {
    "id": "identity-uuid",
    "traits": {
      "email": "user@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "username": "johndoe"
    },
    "created_at": "2024-01-15T12:20:00Z",
    "updated_at": "2024-01-15T12:20:00Z"
  }
}
```

**Your Backend Response:**
```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "success": true
}
```

**Use Case:** Create user record in your database, provision resources, send welcome email, etc.

---

## HTTP Client Configuration

### Required Headers

**For all requests TO Kratos:**
- `Content-Type: application/json`
- `X-Session-Token: <token>` (for whoami endpoint)
- `Cookie: <cookies>` (for browser flows)

**For all responses FROM your backend:**
- `Content-Type: application/json`
- `Set-Cookie: <forward-from-kratos>` (for registration/login)
- `Access-Control-Allow-Credentials: true` (CORS)
- `Access-Control-Allow-Origin: <frontend-origin>` (CORS)

### Cookie Handling

**CRITICAL:** Your HTTP client MUST:
1. Accept cookies from Kratos responses
2. Forward cookies to frontend
3. Accept cookies from frontend requests
4. Send cookies to Kratos

**Python (requests):**
```python
import requests

session = requests.Session()
response = session.post('http://127.0.0.1:4433/...')
cookies = response.cookies
```

**Node.js (axios):**
```javascript
const axios = require('axios');

const response = await axios.post('http://127.0.0.1:4433/...', {}, {
  withCredentials: true
});
const cookies = response.headers['set-cookie'];
```

**Go (net/http):**
```go
jar, _ := cookiejar.New(nil)
client := &http.Client{Jar: jar}
resp, _ := client.Post("http://127.0.0.1:4433/...", ...)
cookies := resp.Cookies()
```

---

## Error Responses

### Registration/Login Errors (400 Bad Request)

```json
{
  "id": "flow-uuid",
  "ui": {
    "messages": [
      {
        "id": 4000006,
        "text": "The provided credentials are invalid, check for spelling mistakes in your password or username, email address, or phone number.",
        "type": "error"
      }
    ]
  }
}
```

**Common Error IDs:**
- `4000001` - Field required
- `4000002` - Field invalid format
- `4000006` - Invalid credentials
- `4000007` - Account already exists
- `4000010` - Password too weak

### Session Validation Errors (401 Unauthorized)

```json
{
  "error": {
    "code": 401,
    "status": "Unauthorized",
    "reason": "The request could not be authorized",
    "message": "No valid session cookie found"
  }
}
```

---

## Testing with cURL

### 1. Registration Flow

```bash
# Init registration
FLOW=$(curl -s http://localhost:3000/api/auth/flows/registration | jq -r '.data.id')

# Submit registration
curl -X POST http://localhost:3000/api/auth/flows/registration \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d "{
    \"flow_id\": \"$FLOW\",
    \"email\": \"test@example.com\",
    \"password\": \"password123\",
    \"first_name\": \"John\",
    \"last_name\": \"Doe\",
    \"username\": \"johndoe\"
  }"
```

### 2. Login Flow

```bash
# Init login
FLOW=$(curl -s http://localhost:3000/api/auth/flows/login | jq -r '.data.id')

# Submit login
curl -X POST http://localhost:3000/api/auth/flows/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d "{
    \"flow_id\": \"$FLOW\",
    \"identifier\": \"test@example.com\",
    \"password\": \"password123\"
  }"
```

### 3. Whoami

```bash
curl http://localhost:3000/api/auth/whoami \
  -b cookies.txt
```

### 4. Protected Endpoint

```bash
curl http://localhost:3000/api/documents \
  -b cookies.txt
```

### 5. Logout

```bash
# Get logout URL
LOGOUT_URL=$(curl -s http://localhost:3000/api/auth/flows/logout \
  -b cookies.txt | jq -r '.data.logout_url')

# Complete logout
curl -L "$LOGOUT_URL" \
  -b cookies.txt \
  -c cookies.txt
```

---

## Language-Specific Examples

### Python (FastAPI)

```python
from fastapi import FastAPI, Request, Response
import httpx

app = FastAPI()
kratos_public = "http://127.0.0.1:4433"

@app.get("/api/auth/flows/registration")
async def init_registration():
    async with httpx.AsyncClient() as client:
        resp = await client.get(f"{kratos_public}/self-service/registration/api")
        return {"success": True, "data": resp.json()}

@app.post("/api/auth/flows/registration")
async def submit_registration(request: Request, response: Response):
    body = await request.json()
    flow_id = body["flow_id"]

    async with httpx.AsyncClient() as client:
        resp = await client.post(
            f"{kratos_public}/self-service/registration?flow={flow_id}",
            json={
                "method": "password",
                "traits": {
                    "email": body["email"],
                    "first_name": body["first_name"],
                    "last_name": body["last_name"],
                    "username": body["username"]
                },
                "password": body["password"]
            }
        )

        # Forward cookies
        for cookie in resp.cookies:
            response.set_cookie(cookie.name, cookie.value, httponly=True, samesite="lax")

        return {"success": True, "data": resp.json()}
```

### Node.js (Express)

```javascript
const express = require('express');
const axios = require('axios');

const app = express();
const kratosPublic = 'http://127.0.0.1:4433';

app.get('/api/auth/flows/registration', async (req, res) => {
  const { data } = await axios.get(`${kratosPublic}/self-service/registration/api`);
  res.json({ success: true, data });
});

app.post('/api/auth/flows/registration', async (req, res) => {
  const { flow_id, email, password, first_name, last_name, username } = req.body;

  const { data, headers } = await axios.post(
    `${kratosPublic}/self-service/registration?flow=${flow_id}`,
    {
      method: 'password',
      traits: { email, first_name, last_name, username },
      password
    }
  );

  // Forward cookies
  const cookies = headers['set-cookie'];
  if (cookies) {
    cookies.forEach(cookie => res.append('Set-Cookie', cookie));
  }

  res.json({ success: true, data });
});
```

### Go (net/http)

```go
package main

import (
    "encoding/json"
    "net/http"
    "io"
)

const kratosPublic = "http://127.0.0.1:4433"

func initRegistration(w http.ResponseWriter, r *http.Request) {
    resp, _ := http.Get(kratosPublic + "/self-service/registration/api")
    defer resp.Body.Close()

    body, _ := io.ReadAll(resp.Body)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "data":    json.RawMessage(body),
    })
}

func submitRegistration(w http.ResponseWriter, r *http.Request) {
    var req struct {
        FlowID    string `json:"flow_id"`
        Email     string `json:"email"`
        Password  string `json:"password"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
        Username  string `json:"username"`
    }
    json.NewDecoder(r.Body).Decode(&req)

    // Call Kratos...
    // Forward cookies...
}
```

---

## Summary

**Key Points:**
1. Frontend calls YOUR backend, not Kratos directly
2. YOUR backend proxies requests to Kratos
3. Forward ALL Set-Cookie headers from Kratos to frontend
4. Extract session cookie and validate with whoami for protected routes
5. Session cookie name: `ory_kratos_session`
6. Use `X-Session-Token` header for Kratos whoami endpoint

This API spec works with any language/framework!
