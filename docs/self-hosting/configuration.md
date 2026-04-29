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

Redis is used for per-project rate limiting on `/send`, `/send/batch` and `/broadcast`, and for caching the GitHub releases response that powers the "update available" badge in the dashboard. Port `6380` by default. Self-hosted instances exposed to the internet should run Redis even if you don't think you need it for caching, specifically so the rate limits stay enforced.

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

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 320" role="img" aria-label="Pro license decision matrix" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="lm-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.6"/></marker>
    <marker id="lm-ag" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#10b981"/></marker>
    <marker id="lm-ar" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#ef4444"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <g transform="translate(310,16)"><rect x="0" y="0" width="160" height="50" rx="25" fill="none" stroke="currentColor" stroke-opacity="0.85" stroke-width="1.6"/><text x="80" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">DEPLOYMENT_MODE</text><text x="80" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">self-hosted · cloud</text></g>
    <path d="M 360 66 L 200 110" stroke="currentColor" stroke-opacity="0.6" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="260" y="86" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.75">self-hosted</text>
    <path d="M 420 66 L 580 110" stroke="currentColor" stroke-opacity="0.6" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="520" y="86" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.75">cloud</text>
    <g transform="translate(110,116)"><rect x="0" y="0" width="180" height="50" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="90" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">SENDDOCK_LICENSE_KEY?</text><text x="90" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">empty · valid</text></g>
    <g transform="translate(490,116)"><rect x="0" y="0" width="180" height="50" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="90" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">SENDDOCK_LICENSE_KEY?</text><text x="90" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">empty · valid</text></g>
    <path d="M 150 166 L 80 220" stroke="#10b981" stroke-width="1.5" fill="none" marker-end="url(#lm-ag)"/>
    <text x="92" y="190" text-anchor="middle" font-size="11" fill="#10b981">empty</text>
    <path d="M 250 166 L 320 220" stroke="#10b981" stroke-width="1.5" fill="none" marker-end="url(#lm-ag)"/>
    <text x="305" y="190" text-anchor="middle" font-size="11" fill="#10b981">valid</text>
    <path d="M 530 166 L 460 220" stroke="#ef4444" stroke-width="1.5" fill="none" marker-end="url(#lm-ar)"/>
    <text x="472" y="190" text-anchor="middle" font-size="11" fill="#ef4444">empty</text>
    <path d="M 630 166 L 700 220" stroke="#10b981" stroke-width="1.5" fill="none" marker-end="url(#lm-ag)"/>
    <text x="685" y="190" text-anchor="middle" font-size="11" fill="#10b981">valid</text>
    <g transform="translate(20,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="#10b981" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="#10b981">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="#10b981">local dev</text></g>
    <g transform="translate(260,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="#10b981" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="#10b981">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="#10b981">licensed</text></g>
    <g transform="translate(400,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="#ef4444" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="#ef4444">Locked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="#ef4444">free tier</text></g>
    <g transform="translate(640,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="#10b981" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="#10b981">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="#10b981">licensed</text></g>
    <text x="380" y="306" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">Lemon Squeezy validates the key on startup and re-validates periodically</text>
  </g>
</svg>

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
