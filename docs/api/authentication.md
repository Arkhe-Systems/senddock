# Authentication

SendDock has two auth schemes for two audiences:

- **API key (`Authorization: Bearer sk_...`)** — the primary way to integrate. Project-scoped. Use for backend code, transactional sends, CI scripts.
- **Cookie session (HttpOnly JWT)** — used by the bundled dashboard SPA. Documented here in case you build your own UI on top of the same API.

Two more public endpoints have their own credentials (no Bearer header, no cookie):

- **Webhook delivery signatures** — outbound HTTP POSTs to your URL carry `X-SendDock-Signature: t=...,v1=...`. Verify with HMAC-SHA256. See [Webhooks guide → Verifying the signature](../guide/webhooks#verifying-the-signature).
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
{
  "user_id": "01H...",
  "email": "you@example.com",
  "name": "Your Name",
  "plan": "pro",
  "created_at": "2026-01-01T00:00:00Z"
}
```

`plan` is `"free"`, `"pro"` or `"team"` for self-host (from the license activated under **Instance → License**); on cloud it's `"free"`, `"starter"`, `"growth"` or `"scale"`.

### Change password

```
POST /api/v1/me/password
```

```json
{
  "current_password": "...",
  "new_password": "..."
}
```

Verifies `current_password` with bcrypt, then validates `new_password` against the registration rules: min 8 chars, at least one uppercase, one digit, one special character. Returns `200` on success, `400` on weak new password, `401` on bad current password.

Changing the password does **not** invalidate existing sessions on other devices — sign out from those manually if needed.

## Two-factor authentication

TOTP with single-use recovery codes. See the dashboard-side flow in [Your account & security → 2FA](../guide/account#two-factor-authentication).

### Begin setup

```
POST /api/v1/me/2fa/setup
```

No body. Response:

```json
{
  "otpauth_url": "otpauth://totp/SendDock:you@example.com?secret=...&issuer=SendDock",
  "secret": "BASE32SECRET",
  "recovery_codes": ["...", "...", "..."]
}
```

Recovery codes are returned **only here** and only this once. Persist them client-side immediately or you lose them.

### Confirm setup

```
POST /api/v1/me/2fa/verify
```

```json
{ "code": "123456" }
```

Validates the 6-digit TOTP code against the secret from `/setup`. On success, 2FA is **enabled** on the account; subsequent logins require the second step. Returns `400` on wrong code.

### Disable 2FA

```
POST /api/v1/me/2fa/disable
```

```json
{ "code": "123456" }
```

`code` is either a valid TOTP code or a single-use recovery code. No code, no disable — there is no admin override.

### Regenerate recovery codes

```
POST /api/v1/me/2fa/recovery-codes
```

```json
{ "code": "123456" }
```

Requires a valid TOTP code. Invalidates the existing set and returns ten fresh ones:

```json
{ "recovery_codes": ["...", "...", "..."] }
```

### Login second step

When `/api/v1/auth/login` succeeds against an account with 2FA enabled, it returns `200` with `needs_2fa: true` and an intermediate token instead of setting session cookies:

```json
{ "needs_2fa": true, "intermediate_token": "..." }
```

Complete the login by posting to:

```
POST /api/v1/auth/2fa
```

```json
{ "intermediate_token": "...", "code": "123456" }
```

`code` is a TOTP code or a recovery code. On success, session cookies are set the same way as a regular login. The intermediate token is short-lived (a few minutes) and only valid for this exchange.

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
