# Authentication

SendDock uses two authentication methods:

- **Cookie auth** — for the web UI (HttpOnly cookies with JWT)
- **API key auth** — for external applications (`Authorization: Bearer sk_...`)

## Cookie Auth (Web UI)

### Register (cloud mode only)

```
POST /api/v1/auth/register
```

```json
{
  "email": "user@example.com",
  "password": "yourpassword",
  "name": "Your Name"
}
```

Sets `access_token` and `refresh_token` cookies.

### Login

```
POST /api/v1/auth/login
```

```json
{
  "email": "user@example.com",
  "password": "yourpassword"
}
```

Sets `access_token` (15 min) and `refresh_token` (7 days) cookies.

### Refresh

```
POST /api/v1/auth/refresh
```

No body required. The refresh token cookie is sent automatically. Returns new access and refresh token cookies.

### Logout

```
POST /api/v1/auth/logout
```

Invalidates the refresh token and clears cookies.

### Get Current User

```
GET /api/v1/me
```

Returns the authenticated user's ID.

```json
{"user_id": "uuid"}
```

## API Key Auth

For endpoints that support API key auth, pass the key in the Authorization header:

```
Authorization: Bearer sk_your_api_key
```

API keys are project-scoped. They authenticate the request and bind it to the key's project.

### Endpoints supporting API key auth

- `POST /projects/{id}/send`
- `POST /projects/{id}/broadcast`
- `POST /projects/{id}/send/direct`
- `GET /projects/{id}/stats`

All other endpoints require cookie auth.

## Setup (first-time)

### Check setup status

```
GET /api/v1/setup/status
```

```json
{"setup_required": true, "deployment_mode": "self-hosted"}
```

### Complete setup

```
POST /api/v1/setup
```

```json
{
  "name": "Admin",
  "email": "admin@example.com",
  "password": "yourpassword"
}
```

Only works when no users exist in the database.
