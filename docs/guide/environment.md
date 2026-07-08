# Environment Variables

All configuration is done via environment variables. For self-hosting deployments see the [installation guide](/self-hosting/installation) — `install.sh` autogenerates `.env` for you, and Option 2 (manual compose) ships a copy-paste reference. If you're running the backend from source, copy `.env.example` to `.env` in the `backend/` directory.

## Required

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret used for JWT session signing **and** the HMAC on click-tracking URLs. Use a random string, min 32 chars (`openssl rand -hex 32`). Rotating it invalidates every active session and every unsubscribe / tracking link signed against it. |

## Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP port the binary listens on inside the container. Not the same as `SENDDOCK_PORT`, which the compose file uses to map the host port. Leave at `8080` unless you know why you're changing it. | `8080` |
| `REDIS_URL` | Redis connection string. Rate limiting is backed by Redis, so **always configure it in production / any internet-facing deployment**. Only omit it in ephemeral local/test environments. | — |
| `FRONTEND_URL` | Frontend origin for CORS headers. When it starts with `https://`, auth cookies are issued with `Secure: true`. | `http://localhost:5173` |
| `PUBLIC_URL` | **Deprecated — set this from the dashboard instead** (Instance settings). If present it is imported into the database on first boot and support for it is removed in v0.9. | _stored in the database_ |
| `DEPLOYMENT_MODE` | `self-hosted` or `cloud` | `self-hosted` |
| `SENDDOCK_LICENSE_KEY` | Pro / Team license key. Validated against Lemon Squeezy. Empty leaves the deployment on the free tier (Core only) regardless of `DEPLOYMENT_MODE`. See [Pro license](/self-hosting/configuration#plans-and-licensing). | — |
| `RATE_LIMIT_PER_MINUTE` | Per-IP request cap for the **global** rate limiter (rolling 60s fixed window, applied to every HTTP endpoint except `/health`). Independent from the hard-coded per-project sending limits on `/send`, `/send/batch`, `/broadcast` (those are not configurable). Lower this on small deployments behind a single egress IP; raise it for high-traffic apps. Only enforced when `REDIS_URL` is set. | `600` |
| `SENDDOCK_WATCHTOWER_URL` | URL of the [Watchtower](https://containrrr.dev/watchtower/) HTTP API (typically `http://watchtower:8080` on the Docker network). When set and reachable, the dashboard's update modal shows an "Update now" button that triggers a Watchtower scan + image refresh. When unset, the modal falls back to the manual `docker compose pull && up -d` command. See [Updating → One-click updates](/self-hosting/updating#one-click-updates-from-the-dashboard-watchtower). | — |
| `SENDDOCK_WATCHTOWER_TOKEN` | Bearer token expected by Watchtower's HTTP API (set as `WATCHTOWER_HTTP_API_TOKEN` on the Watchtower container). Required when `SENDDOCK_WATCHTOWER_URL` is set. | — |
| `TEMPLATE_LIBRARY_URL` | Manifest URL for the starter library shown by the **★ Browse library** modal on the Templates page. Override to point at a private fork or a curated internal gallery — the schema is documented in the [public repo](https://github.com/Arkhe-Systems/senddock-templates). The backend caches the manifest in Redis for one hour, so changes take up to an hour to propagate. Empty disables the library button. | `https://raw.githubusercontent.com/Arkhe-Systems/senddock-templates/main/index.json` |

## Compose-only (not read by the binary)

These are consumed by the bundled `docker-compose.image.yml` / `docker-compose.prod.yml` to wire the stack together. The Go binary never reads them.

| Variable | Description | Default |
|----------|-------------|---------|
| `POSTGRES_PASSWORD` | Password for the bundled Postgres service. Required — the compose fails fast if it's unset. Generate with `openssl rand -base64 32`. | — |
| `SENDDOCK_PORT` | Host port the SendDock container is exposed on. Mapped to `8080` inside the container (which is what `PORT` controls). Set to e.g. `9090` to reach SendDock at `http://host:9090`. | `8080` |

## Advanced overrides

You almost never need these — they're escape hatches for non-default deployments.

| Variable | Description | Default |
|----------|-------------|---------|
| `SENDDOCK_LICENSE_ENDPOINT` | Override the Lemon Squeezy API root the validator hits. Useful when running against Lemon Squeezy's test mode (`https://api.lemonsqueezy.com/v1` is the production default; test mode uses the same URL but a different key). | `https://api.lemonsqueezy.com/v1` |
| `FRONTEND_DIST_PATH` | Filesystem path to the built frontend SPA (`index.html` and assets). The official Docker image sets this to `/app/frontend/dist`. Override only if you serve the SPA from a custom location. | `./frontend/dist` |
| `DISPOSABLE_DOMAINS_FILE` | Path to a newline-separated list of disposable email domains used by the import validator. Built-in list ships with the binary; setting this replaces it (does not extend). | _built-in_ |

::: tip Where is the public URL now?
Outgoing emails contain links like the unsubscribe URL and the open-tracking pixel. SendDock cannot guess what URL recipients will see — it has to be told.

Since v0.8 this lives in the database, not in the environment. You set it on the Setup screen the first time you open SendDock, and change it any time under **Instance** in the dashboard. Changes apply immediately, with no restart.

Existing installations that still have `PUBLIC_URL` in their environment keep working: the value is imported into the database on first boot and a deprecation warning is logged. The variable is removed in v0.9.
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
DEPLOYMENT_MODE=self-hosted
SENDDOCK_LICENSE_KEY=
```
