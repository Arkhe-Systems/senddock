# SendDock

**Official launch: May 1st, 2026. Star this repo to get notified.**

Open-source email marketing platform. Self-hostable, API-first, BYOSMTP. Built with Go and Vue.

Bring your own SMTP. Zero cost per email. Full control over your data.

Part of [Arkhe Systems](https://arkhe.systems).

## Quick Start

```bash
git clone https://github.com/arkhe-systems/senddock.git
cd senddock
chmod +x setup.sh && ./setup.sh    # Windows: .\setup.ps1
```

Open `http://localhost:8080`, create your admin account, and start sending.

## Development Setup

### Prerequisites

- Go 1.22+
- Node.js 20+
- Docker and Docker Compose
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- [sqlc](https://github.com/sqlc-dev/sqlc) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

### Backend

```bash
cd backend
cp .env.example .env
make dev
```

Runs at `http://localhost:8080`.

### Frontend

```bash
cd frontend
npm install
npm run dev
```

Runs at `http://localhost:5173`.

### Make commands

| Command | Description |
|---------|-------------|
| `make dev` | Start DB + migrations + server |
| `make run` | Start server only |
| `make test` | Run unit tests |
| `make sqlc` | Regenerate sqlc code |
| `make migrate` | Run database migrations |
| `make build` | Build production binary |
| `make db-up` | Start PostgreSQL and Redis |
| `make db-down` | Stop PostgreSQL and Redis |

## API

Authentication is managed via HttpOnly cookies, set automatically on login/register.

### Auth

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout |
| GET | `/api/v1/me` | Current user |

### Projects

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects` | Create project |
| GET | `/api/v1/projects` | List projects |
| GET | `/api/v1/projects/{id}` | Get project |
| PUT | `/api/v1/projects/{id}` | Update project |
| DELETE | `/api/v1/projects/{id}` | Delete project |
| PUT | `/api/v1/projects/{id}/smtp` | Update SMTP settings |

### Subscribers

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects/{id}/subscribers` | Add subscriber |
| GET | `/api/v1/projects/{id}/subscribers` | List subscribers |
| POST | `/api/v1/projects/{id}/subscribers/import` | Bulk import subscribers |
| POST | `/api/v1/projects/{id}/subscribers/bulk` | Bulk action (update status / delete) |
| PATCH | `/api/v1/projects/{id}/subscribers/{subscriberId}` | Update status |
| DELETE | `/api/v1/projects/{id}/subscribers/{subscriberId}` | Remove subscriber |

### Templates

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects/{id}/templates` | Create template |
| GET | `/api/v1/projects/{id}/templates` | List templates |
| GET | `/api/v1/projects/{id}/templates/{templateId}` | Get template |
| PUT | `/api/v1/projects/{id}/templates/{templateId}` | Update template |
| DELETE | `/api/v1/projects/{id}/templates/{templateId}` | Delete template |

### API Keys

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects/{id}/keys` | Create API key |
| GET | `/api/v1/projects/{id}/keys` | List API keys |
| DELETE | `/api/v1/projects/{id}/keys/{keyId}` | Revoke API key |

### Email Sending

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/projects/{id}/send` | Cookie or API key | Send email (template or direct) |
| POST | `/api/v1/projects/{id}/send/batch` | Cookie or API key | Send template to multiple recipients |
| POST | `/api/v1/projects/{id}/broadcast` | Cookie or API key | Send template to all active subscribers |
| POST | `/api/v1/projects/{id}/smtp/test` | Cookie | Test SMTP connection |
| GET | `/api/v1/projects/{id}/logs` | Cookie | List email logs |
| GET | `/api/v1/projects/{id}/stats` | Cookie or API key | Get email stats |
| GET | `/unsubscribe/{id}/{subscriberId}` | Public | Unsubscribe confirmation page |
| POST | `/unsubscribe/{id}/{subscriberId}` | Public | One-click unsubscribe (RFC 8058) |
| GET | `/t/{logId}.gif` | Public | Open tracking pixel |
| GET | `/c/{logId}/{payload}` | Public | Click tracking redirect |

### Campaigns

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects/{id}/campaigns` | Create campaign |
| GET | `/api/v1/projects/{id}/campaigns` | List campaigns |
| DELETE | `/api/v1/projects/{id}/campaigns/{campaignId}` | Delete/cancel campaign (scheduled only) |

### Waitlist

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/v1/projects/{id}/waitlist` | Public | Join waitlist (creates subscriber + sends confirmation email) |

### Setup

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/setup/status` | Check if setup is required |
| POST | `/api/v1/setup` | Create admin account (first-time only) |

### Pro endpoints

Gated by a valid `SENDDOCK_LICENSE_KEY` (or self-hosted with empty key for local development). Compiled into the official `ghcr.io/arkhe-systems/senddock` image; not present in source builds.

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/projects/{id}/analytics/overview` | Aggregated metrics for the Analytics dashboard |
| POST | `/api/v1/projects/{id}/webhooks` | Create webhook |
| GET | `/api/v1/projects/{id}/webhooks` | List webhooks |
| GET | `/api/v1/projects/{id}/webhooks/{webhookId}` | Get webhook |
| PATCH | `/api/v1/projects/{id}/webhooks/{webhookId}` | Pause/resume webhook |
| DELETE | `/api/v1/projects/{id}/webhooks/{webhookId}` | Delete webhook |
| GET | `/api/v1/projects/{id}/webhooks/{webhookId}/deliveries` | List recent delivery attempts |

API key auth uses `Authorization: Bearer sk_...` header.

## Core vs Pro

SendDock is open-core: the open-source binary you can self-host today is fully usable on its own. A license key unlocks an extra dashboard and management surface for teams that want it.

**Core (AGPL, free, in this repo)**
- Project, subscriber, template, API key, campaign and SMTP management
- Transactional sends, broadcasts, batch sends, scheduled campaigns
- Open tracking, click tracking, one-click unsubscribe (RFC 8058)
- Webhook **dispatcher** with HMAC signing and retries — webhooks created on a Pro instance keep firing here
- Per-project rate limits (Redis-backed), encrypted SMTP credentials, JWT auth
- Update-available notice in the dashboard polled from GitHub releases

**Pro (private, license-gated)**
- Analytics dashboard with funnel, opens-over-time, top templates, top clicked links, insights and trend pills against the previous period
- Webhooks management UI and REST API (CRUD, pause/resume, deliveries history)
- Future: team members + roles, SMTP failover, SSO/LDAP, white-label

An empty `SENDDOCK_LICENSE_KEY` keeps Pro locked in any deployment mode — Core stays fully usable for free, Pro requires a license. See [docs/self-hosting/configuration.md](docs/self-hosting/configuration.md) for details.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | — |
| `REDIS_URL` | Redis connection string (required for rate limits) | — |
| `JWT_SECRET` | Secret key for JWT signing and click-tracking HMAC | — |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:5173` |
| `PUBLIC_URL` | Public URL of the instance, used in unsubscribe and tracking links inside emails | falls back to `FRONTEND_URL` |
| `DEPLOYMENT_MODE` | `self-hosted` or `cloud` | `self-hosted` |
| `SENDDOCK_LICENSE_KEY` | Pro license key (Lemon Squeezy). Empty leaves the deployment on the free tier (Core only) | — |

## License

AGPL-3.0 — See [LICENSE](LICENSE).

This software is free to use and self-host. If you modify SendDock and offer it as a hosted service, you must open-source your modifications under the same license.

For commercial licensing, contact hello@senddock.dev.
