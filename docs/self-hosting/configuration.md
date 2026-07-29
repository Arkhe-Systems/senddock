# Configuration

## Environment Variables

See [Environment Variables](/guide/environment) for the full list.

## Database

SendDock uses PostgreSQL 17+. The development compose (`docker-compose.yml`, used by `make dev`) exposes Postgres on host port `5434` so you can connect from `make dev` running on the host without colliding with a local Postgres. **Production composes** (`docker-compose.image.yml`, `docker-compose.prod.yml`) keep Postgres on the internal Docker network only — no host port is exposed, and the container talks to it on the standard `5432`.

Defaults bundled by the production compose:

- User: `senddock`
- Database: `senddock`
- Password: `$POSTGRES_PASSWORD` from your `.env`
- Internal port: `5432`

### Using an external database

Set `DATABASE_URL` to your PostgreSQL connection string and drop the bundled `postgres` service from the compose file:

```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

Migrations run automatically on container start (the entrypoint calls `goose up` against `DATABASE_URL`). For source-from-scratch runs without Docker:

```bash
make migrate
```

::: warning Managed Postgres (Neon, Supabase, RDS, Aiven) needs `pg_trgm`
The email-logs search uses GIN indexes on `pg_trgm`. The migration `20260528193111_add_email_logs_search_index.sql` runs `CREATE EXTENSION IF NOT EXISTS pg_trgm` for you — this works out of the box on the bundled Postgres image, but some managed Postgres providers require the extension to be on an allowlist before `CREATE EXTENSION` succeeds:

- **Neon, Supabase, Aiven** — `pg_trgm` is available by default, no action needed.
- **AWS RDS** — add `pg_trgm` to the `shared_preload_libraries` parameter group, or enable via `rds_superuser`.
- **Google Cloud SQL** — enable from the **Database flags** UI.

If migration fails with `permission denied to create extension "pg_trgm"`, your DB user lacks privileges. Either run `CREATE EXTENSION pg_trgm;` once as superuser, or grant `CREATE` on the schema to the SendDock user.
:::

## Redis

Redis powers three things:

1. **Global per-IP rate limiter.** `600` requests per rolling 60s window by default (every HTTP endpoint except `/health`). Tune with [`RATE_LIMIT_PER_MINUTE`](/guide/environment#optional).
2. **Per-project rate limits on the sending endpoints** — `/send` (60 req/min), `/send/batch` (10 req/min), `/broadcast` (5 req/hour). These are hard-coded.
3. **Cache for the GitHub releases response** that drives the "update available" badge in the dashboard.

The development compose exposes Redis on host port `6380`. The production composes keep it on the internal Docker network only — the binary talks to it on the standard `6379`.

::: warning Redis is effectively required in production
The binary boots without Redis — `REDIS_URL` is technically optional — but **every limiter silently turns into a no-op**, including the global per-IP cap and the per-project sending caps. For any instance exposed to the internet, that means open abuse vectors: bot traffic, credential stuffing on `/auth/login`, broadcast spam.

The bundled production composes already ship Redis. Only unset `REDIS_URL` for ephemeral test environments where you've accepted there is no rate limit.
:::

## Security Checklist

Before exposing to the internet:

- [ ] Generate `JWT_SECRET` from `openssl rand -hex 32` (or `base64 48`) — used for JWT signing **and** the HMAC on click-tracking URLs. Min 32 chars.
- [ ] Generate `POSTGRES_PASSWORD` from `openssl rand -base64 32` instead of using a guessable value
- [ ] Set your public HTTPS domain under **Instance** in the dashboard (drives unsubscribe + tracking links inside outgoing emails, and the CORS origin)
- [ ] Put SendDock behind HTTPS — `Secure: true` is set on auth cookies when the resolved URL starts with `https://`
- [ ] Keep Redis enabled (it ships in the production composes); without it every rate limit is a no-op
- [ ] Enable **two-factor authentication** on every account from **Settings → Account** — see [Your account & security](/guide/account#two-factor-authentication)
- [ ] On Team workspaces, give each human a least-privilege role (`viewer` / `developer` / `admin`) instead of `owner`
- [ ] Set up Postgres backups before going live — see [Backup & restore](/self-hosting/updating#backup-restore)
- [ ] Keep SendDock updated — the dashboard surfaces a yellow badge when a new release is available

If anything misbehaves after going live, see [Troubleshooting](/self-hosting/troubleshooting).

## Plans and licensing

SendDock is open-core: the AGPL-3.0 **Community** edition does everything most one-person operations need, and two paid tiers (**Pro** and **Team**) add deliverability, reporting and team collaboration. Paid tiers are unlocked with a license key validated against [Lemon Squeezy](https://lemonsqueezy.com).

On a **self-hosted** instance you activate that key from the dashboard, under **Instance → License** — it's stored encrypted in the database and applies immediately, no restart. See [Instance settings → Pro license](/guide/instance-settings#pro-license). The legacy `SENDDOCK_LICENSE_KEY` environment variable still works but is **deprecated**: if set, it's imported into the database once on boot and support for it is removed in v0.9. On the **cloud** deployment the environment owns the key.

::: info Self-hosted prices below are flat per instance
The `$9` and `$29` prices on this page are the **self-hosted** rates — one license per SendDock instance, unlimited subscribers, unlimited sends. The [managed cloud at senddock.dev](https://senddock.dev/#pricing) is live and priced separately on a per-subscriber tier basis (Free up to 2,000, then $19 / $49 / $129 per month) because we operate the platform for you. BYO SMTP applies to both — neither path charges per send.
:::

### What each tier unlocks

| Capability | Community (Free) | Pro ($9 / mo, $90 / yr) | Team ($29 / mo, $290 / yr) |
|---|---|---|---|
| Single user, multiple workspaces | ✓ | ✓ | ✓ |
| Subscribers, templates, broadcasts | ✓ | ✓ | ✓ |
| BYO SMTP, unlimited sends | ✓ | ✓ | ✓ |
| Bounce ingestion, suppression list | ✓ | ✓ | ✓ |
| Click & open tracking | ✓ | ✓ | ✓ |
| API keys + webhook dispatcher | ✓ | ✓ | ✓ |
| Analytics dashboard (Overview, Campaigns, Audience, Engagement) | ✓ | ✓ | ✓ |
| Webhooks management UI | — | ✓ | ✓ |
| Deliverability — domain health + per-provider rates + spam rate | — | ✓ | ✓ |
| Report builder — pivots, segment filters, saved reports, CSV | — | ✓ | ✓ |
| Audit log | — | ✓ | ✓ |
| Multi-user workspaces (members + invites) | — | — | ✓ |
| Roles: owner / admin / developer / viewer | — | — | ✓ |
| Admin user creation (no public registration needed) | — | — | ✓ |
| A/B testing, segments & tags, approval workflow | — | — | roadmap |

A future **Enterprise** tier will add SSO/SCIM, per-project ACLs, white-label tracking domain, per-seat billing and SLA support — pricing on request, starting around $149 / mo.

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 320" role="img" aria-label="License key decision flow" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="lm-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <g transform="translate(290,20)"><rect x="0" y="0" width="180" height="50" rx="25" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="90" y="24" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">SENDDOCK_LICENSE_KEY</text><text x="90" y="40" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">checked at startup</text></g>
    <path d="M 380 70 L 380 110 L 150 110 L 150 218" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.5" stroke-dasharray="5 4" fill="none" marker-end="url(#lm-a)"/>
    <text x="262" y="100" text-anchor="middle" font-size="11" font-weight="600" fill="currentColor" fill-opacity="0.7">empty</text>
    <path d="M 380 70 L 380 218" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="394" y="146" font-size="11" font-weight="600" fill="currentColor" fill-opacity="0.75">valid Pro key</text>
    <text x="394" y="162" font-size="10" fill="currentColor" fill-opacity="0.55">variant: pro_monthly · pro_annual</text>
    <path d="M 380 70 L 380 110 L 610 110 L 610 218" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="498" y="100" text-anchor="middle" font-size="11" font-weight="600" fill="currentColor" fill-opacity="0.75">valid Team key</text>
    <g transform="translate(60,224)"><rect x="0" y="0" width="180" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.4" stroke-dasharray="5 4"/><text x="90" y="26" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor" fill-opacity="0.85">Community</text><text x="90" y="44" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">core only</text><text x="90" y="58" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.5">Pro · Team locked</text></g>
    <g transform="translate(290,224)"><rect x="0" y="0" width="180" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="90" y="26" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Pro tier</text><text x="90" y="44" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">deliverability · reports · audit</text><text x="90" y="58" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.55">Team locked</text></g>
    <g transform="translate(520,224)"><rect x="0" y="0" width="180" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="90" y="26" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Team tier</text><text x="90" y="44" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">members · roles · admin</text><text x="90" y="58" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.55">Pro + Team unlocked</text></g>
    <text x="380" y="308" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.55">Tier classification by Lemon Squeezy variant_id · re-validated periodically · the CLOUD flag does not change gating</text>
  </g>
</svg>

| `SENDDOCK_LICENSE_KEY` | What unlocks |
|---|---|
| empty | **Community** — Pro features locked, Team features locked. |
| valid Pro key | Pro features unlocked (Deliverability, Reports, Audit log). Team features stay locked. |
| valid Team key | Pro and Team features unlocked. |
| invalid / revoked | Locked after the next validation tick (24h grace from the last successful check). |

The validator distinguishes Pro from Team by the Lemon Squeezy `variant_id` of the license — there is no separate Team key file or env var. Buying Team gives you a single key that the validator recognizes as Team-tier; buying Pro gives a key that validates only the Pro features.

The `CLOUD` flag (which replaced the old `DEPLOYMENT_MODE` string — see [Environment Variables](/guide/environment)) does not change Pro gating — the license requirement applies the same way to self-hosted and cloud. What it does control is who owns configuration and registration: `CLOUD=true` opens the public sign-up endpoints used by the senddock.dev managed product and lets the environment own settings like the public URL and license key. A self-hosted instance (the default) keeps registration closed — relying on the first-boot setup screen plus the admin **Create user** flow on Team workspaces — and owns those settings from the dashboard instead. There is no public sign-up endpoint to call on a self-hosted deploy.

The validator only needs the license key — there is no API key, store ID or webhook secret to configure on the self-hosted side. Those are provisioned on the senddock.dev managed service that issues licenses.

If the license check fails (network outage, key revoked, etc.) SendDock keeps running with the **last successful** validation result for a grace period, then falls back to free-tier behavior. You'll see the cause on stdout: `license: …`.

::: warning Pro requires the prebuilt image
Pro features ship only in the official prebuilt `ghcr.io/arkhe-systems/senddock` image. Building from source (Option 4 in [Installation](/self-hosting/installation)) gives you Core only — setting `SENDDOCK_LICENSE_KEY` on a source build does nothing because the Pro features aren't included.
:::

## Ports

The numbers depend on which compose file you run.

### Production (`docker-compose.image.yml`, `docker-compose.prod.yml`)

| Service | Where it listens | Exposed to host? |
|---------|------------------|------------------|
| SendDock | `:8080` inside the container | yes — on `$SENDDOCK_PORT` (default `8080`) |
| PostgreSQL | `:5432` inside the container | no — Docker network only |
| Redis | `:6379` inside the container | no — Docker network only |

The bundled Postgres and Redis are reachable to the SendDock container by their service names (`postgres`, `redis`) on the standard internal ports. They are **not** published to the host — change that only if you also want to connect from outside Docker, and add an explicit `ports:` block.

### Development (`docker-compose.yml`, used by `make dev`)

| Service | Host port |
|---------|-----------|
| SendDock API (`make run`) | 8080 |
| Frontend dev server (Vite) | 5173 |
| PostgreSQL | 5434 |
| Redis | 6380 |

The 5434 / 6380 offsets exist so the dev stack does not collide with a local Postgres / Redis you may have running on the host.
