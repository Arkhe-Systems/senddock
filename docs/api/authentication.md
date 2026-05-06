# Authentication

SendDock has two auth schemes for two audiences:

- **API key (`Authorization: Bearer sk_...`)** — the primary way to integrate. Project-scoped. Use for backend code, transactional sends, CI scripts.
- **Cookie session (HttpOnly JWT)** — used by the bundled dashboard SPA. Documented here in case you build your own UI on top of the same API.

Two more public endpoints have their own credentials (no Bearer header, no cookie):

- **Webhook delivery signatures** — outbound HTTP POSTs to your URL carry `X-SendDock-Signature: t=...,v1=...`. Verify with HMAC-SHA256. See [Webhooks API → Verifying signatures](./webhooks#verifying-the-signature).
- **Bounce webhook ingest** — `POST /webhooks/bounces/{projectId}?token=...` uses a per-project URL token. See [Bounces API → Public ingest endpoint](./bounces#public-ingest-endpoint).

## API key auth (recommended)

Pass the key in the `Authorization` header:

```
Authorization: Bearer sk_your_api_key
```

Keys are project-scoped — the key identifies the project, so the URL `{id}` segment is ignored when API-key auth is used. Create and revoke keys in **Project → Settings → API Keys** or via the [API Keys API](./api-keys).

### Endpoints that accept API keys

| Endpoint | What it does |
|---|---|
| `POST /api/v1/projects/{id}/send` | Single transactional send (template or raw HTML). |
| `POST /api/v1/projects/{id}/send/batch` | Same template, many recipients with per-recipient `data`. |
| `POST /api/v1/projects/{id}/broadcast` | Send to every active subscriber. |
| `POST /api/v1/projects/{id}/subscribers/import` | CSV / JSON import with email validation. |
| `GET /api/v1/projects/{id}/stats` | Read-only summary counts (`sent`, `failed`, `bounced`, `suppressed`, opens, clicks). |

Every other endpoint requires cookie auth — see the next section. As a rule of thumb: **anything that mutates project, workspace, template, subscriber, webhook or audit-log state is cookie-only**, because role-based capabilities key off the user identity that an API key doesn't carry.

### Errors

| Status | Body | Meaning |
|---|---|---|
| `401` | `{"error":"missing or invalid api key"}` | Header missing, malformed, or key revoked. |
| `403` | `{"error":"forbidden"}` | Key valid but capability denied. (API keys carry full project access today, so 403 only happens on cookie auth.) |

## Cookie session (dashboard SPA)

The bundled web dashboard uses HttpOnly JWT cookies. If you're building your own frontend against the same backend, you can use the same flow.

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

Sets an `access_token` (15 min lifetime) and `refresh_token` (7 days) cookie, both `HttpOnly`, `Secure` and `SameSite=Lax`.

### Refresh

```
POST /api/v1/auth/refresh
```

No body. The browser sends `refresh_token` automatically. Returns rotated cookies. The dashboard SPA calls this transparently when a 401 fires from an expired access token.

### Logout

```
POST /api/v1/auth/logout
```

Invalidates the refresh token server-side and clears both cookies.

### Get current user

```
GET /api/v1/me
```

```json
{ "user_id": "01H..." }
```

Returns just the id. Resolve to display name / email through `GET /workspaces/{id}/members`.

## First-boot setup

These two endpoints exist only until the first user is created — after that they return errors. Self-hosted instances run them via the **Setup** screen the first time you open the dashboard; you almost never need to call them by hand.

### Status

```
GET /api/v1/setup/status
```

```json
{ "setup_required": true, "deployment_mode": "self-hosted" }
```

`setup_required` flips to `false` once at least one user exists.

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

Returns `200 OK` with the new user's `user_id` and sets login cookies in the same response. Calling this when users already exist returns `409 Conflict`.
