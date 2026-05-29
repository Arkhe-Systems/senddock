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
| `SENDDOCK_LICENSE_KEY` | Pro / Team license key. Validated against Lemon Squeezy. Empty leaves the deployment on the free tier (Core only) regardless of `DEPLOYMENT_MODE`. See [Pro license](/self-hosting/configuration#plans-and-licensing). | — |
| `RATE_LIMIT_PER_MINUTE` | Per-IP request cap for the **global** rate limiter (rolling 60s fixed window, applied to every HTTP endpoint except `/health`). Independent from the hard-coded per-project sending limits on `/send`, `/send/batch`, `/broadcast` (those are not configurable). Lower this on small deployments behind a single egress IP; raise it for high-traffic apps. Only enforced when `REDIS_URL` is set. | `600` |
| `SENDDOCK_WATCHTOWER_URL` | URL of the [Watchtower](https://containrrr.dev/watchtower/) HTTP API (typically `http://watchtower:8080` on the Docker network). When set and reachable, the dashboard's update modal shows an "Update now" button that triggers a Watchtower scan + image refresh. When unset, the modal falls back to the manual `docker compose pull && up -d` command. See [Updating → One-click updates](/self-hosting/updating#one-click-updates-from-the-dashboard-watchtower). | — |
| `SENDDOCK_WATCHTOWER_TOKEN` | Bearer token expected by Watchtower's HTTP API (set as `WATCHTOWER_HTTP_API_TOKEN` on the Watchtower container). Required when `SENDDOCK_WATCHTOWER_URL` is set. | — |

## Advanced overrides

You almost never need these — they're escape hatches for non-default deployments.

| Variable | Description | Default |
|----------|-------------|---------|
| `SENDDOCK_LICENSE_ENDPOINT` | Override the Lemon Squeezy API root the validator hits. Useful when running against Lemon Squeezy's test mode (`https://api.lemonsqueezy.com/v1` is the production default; test mode uses the same URL but a different key). | `https://api.lemonsqueezy.com/v1` |
| `FRONTEND_DIST_PATH` | Filesystem path to the built frontend SPA (`index.html` and assets). The official Docker image sets this to `/app/frontend/dist`. Override only if you serve the SPA from a custom location. | `./frontend/dist` |
| `DISPOSABLE_DOMAINS_FILE` | Path to a newline-separated list of disposable email domains used by the import validator. Built-in list ships with the binary; setting this replaces it (does not extend). | _built-in_ |

::: tip Why PUBLIC_URL?
Outgoing emails contain links like the unsubscribe URL and the open-tracking pixel. SendDock cannot guess what URL recipients will see — it has to be told. In most single-binary deploys (Go binary serves both the API and the SPA), `PUBLIC_URL` and `FRONTEND_URL` are the same value, so you can leave `PUBLIC_URL` blank and only set `FRONTEND_URL`.

Set `PUBLIC_URL` explicitly when:
- Your backend is on a different domain/subdomain than the frontend SPA.
- You're behind a reverse proxy that terminates TLS.
- You want unsubscribe links to point at a different domain than the dashboard.
:::

## Deployment Modes

### self-hosted (default)

- Public sign-up endpoints are not mounted; the only path to the first user is the **Setup** screen on first boot.
- Single-user by default. Adding more users requires the **Team** license, which unlocks the admin "Create user" flow on workspaces (see [Workspaces](./workspaces)).

### cloud

- Public sign-up endpoints are mounted — used by the senddock.dev managed product, not by self-host installs.
- Multi-user accounts with plan-based limits.

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
