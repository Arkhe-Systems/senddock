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

## Plans & licensing

SendDock is open-core: the AGPL-3.0 **Community** edition does everything most one-person operations need, and two paid tiers (**Pro** and **Team**) add features for analytics and team collaboration. All paid tiers are unlocked through a single environment variable validated against [Lemon Squeezy](https://lemonsqueezy.com).

```bash
SENDDOCK_LICENSE_KEY=
```

::: info Self-hosted prices below are flat per instance
The `$9` and `$29` prices on this page are the **self-hosted** rates — one license per SendDock instance, unlimited subscribers, unlimited sends. The managed cloud at senddock.dev (when it ships) is priced separately on a volume basis because we operate the SMTP relays and deliverability for you.
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
| Pro Analytics dashboard | — | ✓ | ✓ |
| Webhooks management UI | — | ✓ | ✓ |
| Audit log | — | ✓ | ✓ |
| Multi-user workspaces (members + invites) | — | — | ✓ |
| Roles: owner / admin / developer / viewer | — | — | ✓ |
| Admin user creation (no public registration needed) | — | — | ✓ |
| A/B testing, segments & tags, approval workflow | — | — | roadmap |

A future **Enterprise** tier will add SSO/SCIM, per-project ACLs, white-label tracking domain, per-seat billing and SLA support — pricing on request, starting around $149 / mo.

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 320" role="img" aria-label="Pro license decision matrix" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="lm-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <g transform="translate(310,16)"><rect x="0" y="0" width="160" height="50" rx="25" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="80" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">DEPLOYMENT_MODE</text><text x="80" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">self-hosted · cloud</text></g>
    <path d="M 360 66 L 200 110" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="252" y="78" text-anchor="middle" font-size="11" font-weight="600" fill="currentColor" fill-opacity="0.75">self-hosted</text>
    <path d="M 420 66 L 580 110" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="528" y="78" text-anchor="middle" font-size="11" font-weight="600" fill="currentColor" fill-opacity="0.75">cloud</text>
    <g transform="translate(110,116)"><rect x="0" y="0" width="180" height="50" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="90" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">SENDDOCK_LICENSE_KEY?</text><text x="90" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">empty · valid</text></g>
    <g transform="translate(490,116)"><rect x="0" y="0" width="180" height="50" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="90" y="22" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">SENDDOCK_LICENSE_KEY?</text><text x="90" y="38" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">empty · valid</text></g>
    <path d="M 200 166 L 200 200 L 80 200 L 80 218" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="140" y="194" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.75">empty</text>
    <path d="M 200 200 L 320 200 L 320 218" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="260" y="194" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.75">valid</text>
    <path d="M 580 166 L 580 200 L 460 200 L 460 218" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.5" stroke-dasharray="5 4" fill="none" marker-end="url(#lm-a)"/>
    <text x="520" y="194" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">empty</text>
    <path d="M 580 200 L 700 200 L 700 218" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#lm-a)"/>
    <text x="640" y="194" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.75">valid</text>
    <g transform="translate(20,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">local dev</text></g>
    <g transform="translate(260,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">licensed</text></g>
    <g transform="translate(400,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.4" stroke-dasharray="5 4"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor" fill-opacity="0.7">Locked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.5">free tier</text></g>
    <g transform="translate(640,224)"><rect x="0" y="0" width="120" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="60" y="26" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">Unlocked</text><text x="60" y="46" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">licensed</text></g>
    <text x="380" y="306" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">Lemon Squeezy validates the key on startup and re-validates periodically</text>
  </g>
</svg>

| `SENDDOCK_LICENSE_KEY` | What unlocks |
|---|---|
| empty | **Community** — Pro features locked, Team features locked. |
| valid Pro key | Pro features unlocked (Analytics, Webhooks, Audit log). Team features stay locked. |
| valid Team key | Pro and Team features unlocked. |
| invalid / revoked | Locked after the next validation tick (24h grace from the last successful check). |

The validator distinguishes Pro from Team by the Lemon Squeezy `variant_id` of the license — there is no separate Team key file or env var. Buying Team gives you a single key that the validator recognizes as Team-tier; buying Pro gives a key that validates only the Pro features.

`DEPLOYMENT_MODE` no longer changes Pro gating — the license requirement applies the same way to self-hosted and cloud. The mode still controls registration (`cloud` enables `POST /api/v1/auth/register`; `self-hosted` keeps registration disabled and relies on the setup screen).

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
