# Configuration

## Environment Variables

See [Environment Variables](/guide/environment) for the full list.

## Database

SendDock uses PostgreSQL 17. The Docker Compose file included in the repo sets up a PostgreSQL instance with:

- User: `senddock`
- Password: `senddock_dev` (change in production)
- Database: `senddock`
- Port: `5434` (to avoid conflicts)

### Using an external database

Set `DATABASE_URL` to your PostgreSQL connection string:

```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

Then run migrations:

```bash
make migrate
```

## Redis

Redis is used for job queues and caching (planned features). Port `6380` by default.

## Security Checklist

Before exposing to the internet:

- [ ] Change `JWT_SECRET` to a random string (min 32 characters): `openssl rand -base64 48`
- [ ] Change PostgreSQL password from default
- [ ] Set `FRONTEND_URL` to your actual domain (used for CORS)
- [ ] Set `PUBLIC_URL` to the same domain (used for unsubscribe + tracking links inside emails)
- [ ] Use HTTPS via reverse proxy — `Secure` cookies are set automatically when `FRONTEND_URL` starts with `https://`
- [ ] Run with Redis enabled — without it the per-project rate limits on `/send`, `/send/batch` and `/broadcast` are bypassed
- [ ] Keep SendDock updated

If anything misbehaves after going live, see [Troubleshooting](/self-hosting/troubleshooting).

## Pro license

Pro features (the [Analytics dashboard](/guide/analytics) and [Webhooks management](/guide/webhooks)) are gated by a single environment variable validated against [Lemon Squeezy](https://lemonsqueezy.com).

```bash
SENDDOCK_LICENSE_KEY=
```

Behavior depends on the deployment mode:

| `DEPLOYMENT_MODE` | `SENDDOCK_LICENSE_KEY` | Pro features |
|---|---|---|
| `self-hosted` (default) | empty | **Unlocked locally** — for development and evaluation. |
| `self-hosted` | valid | Unlocked. License is checked on startup and re-validated periodically. |
| `cloud` | empty | Locked (free tier). |
| `cloud` | valid | Unlocked. |

The validator only needs the license key — there is no API key, store ID or webhook secret to configure on the self-hosted side. Those are provisioned on the senddock.dev managed service that issues licenses.

If the license check fails (network outage, key revoked, etc.) SendDock keeps running with the **last successful** validation result for a grace period, then falls back to free-tier behavior. You'll see the cause on stdout: `license: …`.

::: warning Pro requires the prebuilt image
Pro code lives in a private repository and is compiled into the official `ghcr.io/arkhe-systems/senddock` image. Building from source (Option 3 in [Installation](/self-hosting/installation)) gives you Core only — setting `SENDDOCK_LICENSE_KEY` on a source build does nothing because the gated routes are not in the binary.
:::

## Ports

| Service | Default Port |
|---------|-------------|
| SendDock API | 8080 |
| Frontend (dev) | 5173 |
| PostgreSQL | 5434 |
| Redis | 6380 |
