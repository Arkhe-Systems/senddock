# Environment Variables

All configuration is done via environment variables. Copy `.env.example` to `.env` in the `backend/` directory.

## Required

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret key for JWT signing (use a random string, min 32 chars) |

## Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `REDIS_URL` | Redis connection string. Required for rate-limit enforcement on `/send`, `/send/batch` and `/broadcast`. | — |
| `FRONTEND_URL` | Frontend URL for CORS headers | `http://localhost:5173` |
| `PUBLIC_URL` | Public URL of this instance, used to build unsubscribe and tracking links inside outgoing emails. Leave blank in single-binary deploys to fall back to `FRONTEND_URL`. | _falls back to `FRONTEND_URL`_ |
| `DEPLOYMENT_MODE` | `self-hosted` or `cloud` | `self-hosted` |
| `SENDDOCK_LICENSE_KEY` | Pro license key. Validated against Lemon Squeezy. Empty in self-hosted mode unlocks Pro locally for development; empty in cloud mode keeps the deployment on the free tier. See [Pro license](/self-hosting/configuration#pro-license). | — |

::: tip Why PUBLIC_URL?
Outgoing emails contain links like the unsubscribe URL and the open-tracking pixel. SendDock cannot guess what URL recipients will see — it has to be told. In most single-binary deploys (Go binary serves both the API and the SPA), `PUBLIC_URL` and `FRONTEND_URL` are the same value, so you can leave `PUBLIC_URL` blank and only set `FRONTEND_URL`.

Set `PUBLIC_URL` explicitly when:
- Your backend is on a different domain/subdomain than the frontend SPA.
- You're behind a reverse proxy that terminates TLS.
- You want unsubscribe links to point at a different domain than the dashboard.
:::

## Deployment Modes

### self-hosted (default)

- Public registration is disabled
- First user created via setup screen becomes admin
- Single user only (team members require Pro)

### cloud

- Public registration enabled at `/api/v1/auth/register`
- Multiple users, each with their own account
- Plan-based limits (for senddock.dev managed hosting)

## Example .env

```bash
DATABASE_URL=postgres://senddock:senddock_dev@localhost:5434/senddock?sslmode=disable
JWT_SECRET=change-this-to-a-random-secret
PORT=8080
REDIS_URL=redis://localhost:6380
FRONTEND_URL=https://email.mycompany.com
PUBLIC_URL=https://email.mycompany.com
DEPLOYMENT_MODE=self-hosted
SENDDOCK_LICENSE_KEY=
```
