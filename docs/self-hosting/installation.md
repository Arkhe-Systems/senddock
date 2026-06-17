# Installation

SendDock ships as a single Docker image: `ghcr.io/arkhe-systems/senddock`. Three install paths are supported, in order of preference.

## What gets deployed

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 380" role="img" aria-label="SendDock self-hosted architecture" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="ar-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <g transform="translate(20,50)"><rect x="0" y="0" width="160" height="80" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.35" stroke-dasharray="5 4"/><text x="80" y="34" text-anchor="middle" font-size="13" font-weight="700" fill="currentColor">Recipients</text><text x="80" y="54" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">opens · clicks</text><text x="80" y="68" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">unsubscribes</text></g>
    <line x1="180" y1="90" x2="228" y2="90" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ar-a)"/>
    <g transform="translate(230,50)"><rect x="0" y="0" width="170" height="80" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="85" y="30" text-anchor="middle" font-size="13" font-weight="700" fill="currentColor">Reverse proxy</text><text x="85" y="48" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">HTTPS · your domain</text><text x="85" y="66" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="currentColor" fill-opacity="0.5">caddy · nginx · traefik</text></g>
    <line x1="400" y1="90" x2="448" y2="90" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ar-a)"/>
    <g transform="translate(450,40)"><rect x="0" y="0" width="290" height="100" rx="12" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="1.6"/><text x="145" y="28" text-anchor="middle" font-size="13" font-weight="700" fill="currentColor">senddock</text><text x="145" y="46" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">single Docker container · :8080</text><text x="145" y="78" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.05em" text-transform="uppercase" fill="currentColor" fill-opacity="0.5">go api + vue spa + dispatcher</text></g>
    <line x1="520" y1="140" x2="520" y2="198" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ar-a)"/>
    <line x1="670" y1="140" x2="670" y2="198" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ar-a)"/>
    <g transform="translate(458,200)"><rect x="0" y="0" width="124" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="62" y="28" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">PostgreSQL 17</text><text x="62" y="46" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">pgdata volume</text></g>
    <g transform="translate(608,200)"><rect x="0" y="0" width="124" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.5"/><text x="62" y="28" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Redis 7</text><text x="62" y="46" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">redisdata volume</text></g>
    <line x1="595" y1="140" x2="595" y2="288" stroke="currentColor" stroke-opacity="0.45" stroke-width="1.5" stroke-dasharray="4 3" marker-end="url(#ar-a)"/>
    <text x="605" y="282" font-size="11" fill="currentColor" fill-opacity="0.55">SMTP submit</text>
    <g transform="translate(510,300)"><rect x="0" y="0" width="170" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.35" stroke-dasharray="5 4"/><text x="85" y="26" text-anchor="middle" font-size="13" font-weight="700" fill="currentColor">Your SMTP relay</text><text x="85" y="44" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">SES · Mailgun · Postmark</text></g>
    <line x1="450" y1="140" x2="260" y2="298" stroke="currentColor" stroke-opacity="0.45" stroke-width="1.5" stroke-dasharray="4 3" marker-end="url(#ar-a)"/>
    <text x="430" y="135" text-anchor="end" font-size="11" fill="currentColor" fill-opacity="0.55">webhook POSTs</text>
    <g transform="translate(80,300)"><rect x="0" y="0" width="200" height="60" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.35" stroke-dasharray="5 4"/><text x="100" y="26" text-anchor="middle" font-size="13" font-weight="700" fill="currentColor">Your webhook URLs</text><text x="100" y="44" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">CRM · Slack · Zapier</text></g>
  </g>
</svg>

A single Docker container running the Go binary serves both the API and the Vue SPA. PostgreSQL holds all your data; Redis caches the GitHub releases response and powers per-project rate limits. Your own SMTP relay handles outbound mail — SendDock never proxies your sends.

## Requirements

- A Linux server (Ubuntu, Debian, etc.)
- Docker and Docker Compose
- A **publicly reachable domain** pointed at the server (required for unsubscribe links, broadcasts, and SMTP)
- The server's network must allow **outbound TCP** on the SMTP port your provider uses (usually 465 or 587)

That's it. Everything else is handled by Docker.

::: tip Multi-arch image
The official `ghcr.io/arkhe-systems/senddock` image is built for both **`linux/amd64`** and **`linux/arm64`** in the same release tag. So x86 cloud servers (most VPS, AWS x86, DigitalOcean droplets), AWS Graviton / Hetzner ARM nodes, Raspberry Pi 4+ and Apple Silicon dev machines all pull the right binary automatically — no `--platform` flag, no separate tag.
:::

::: warning Don't expect to send real emails from a laptop
Most residential ISPs block outbound SMTP ports (25, 465, 587) at the network edge to prevent spam botnets. A SendDock instance on your home network cannot deliver mail through any external SMTP provider — including Gmail, SES, Mailgun, or your own Mailcow if it lives elsewhere. **SendDock in production must run on a cloud server (DigitalOcean, Hetzner, AWS, etc.) where outbound SMTP is unblocked.**

For local development and evaluation, ship `make mail` (or any other [Mailpit](https://mailpit.axllent.org/)-like catcher) and configure SendDock to talk to it on `localhost:1025`. See the [SMTP guide](/guide/smtp) for the workaround details.
:::

---

## Option 1: Prebuilt image (recommended)

The fastest path. Pull the canonical image, point a `.env` at it, run.

```bash
mkdir senddock && cd senddock
curl -fsSL https://raw.githubusercontent.com/arkhe-systems/senddock/main/docker-compose.image.yml -o docker-compose.yml
```

Create `.env` next to it:

```bash
cat > .env <<EOF
POSTGRES_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -hex 32)
PUBLIC_URL=https://your-domain.com
SENDDOCK_PORT=8080
SENDDOCK_LICENSE_KEY=
EOF
```

Replace `your-domain.com` with your real public hostname before running. `PUBLIC_URL` is required for unsubscribe and broadcast links to work.

Then:

```bash
docker compose up -d
```

### Full `.env` reference (copy-paste)

The snippet above is the minimum to boot. Below is the complete reference — paste it, replace the `change-me-…` placeholders, and uncomment whatever optional knobs you actually want.

```bash
# === REQUIRED ===

# Public URL where this instance is reachable from the internet.
# Used to build unsubscribe + tracking links inside outgoing emails.
PUBLIC_URL=https://your-domain.com

# Postgres password — generate with: openssl rand -base64 32
POSTGRES_PASSWORD=change-me-openssl-rand-base64-32

# JWT signing secret (min 32 chars) — generate with: openssl rand -hex 32
JWT_SECRET=change-me-openssl-rand-hex-32

# === OPTIONAL ===

# Host port to expose. The container always listens on 8080 internally.
# SENDDOCK_PORT=8080

# Pro / Team license key from senddock.dev. Empty = Community tier (free).
SENDDOCK_LICENSE_KEY=

# Per-IP request cap, rolling 60s window. Default 600. Only enforced with Redis.
# RATE_LIMIT_PER_MINUTE=600

# Watchtower integration for one-click updates from the dashboard.
# Both must be set together. See ./updating.md#one-click-updates-from-the-dashboard-watchtower
# SENDDOCK_WATCHTOWER_URL=http://watchtower:8080
# SENDDOCK_WATCHTOWER_TOKEN=change-me-long-random-string

# === ADVANCED (escape hatches — rarely needed) ===

# Override the Lemon Squeezy license endpoint. Production default works for everyone.
# SENDDOCK_LICENSE_ENDPOINT=https://api.lemonsqueezy.com/v1

# Override the path the binary serves the SPA from. The official image already
# sets this correctly inside the container — only touch it for custom builds.
# FRONTEND_DIST_PATH=/app/frontend/dist

# Path to a newline-separated list of disposable email domains. Replaces the
# built-in list (does not extend it).
# DISPOSABLE_DOMAINS_FILE=/etc/senddock/disposable.txt
```

The container runs `goose up` against Postgres on first start, then serves on `:8080`. Healthcheck tells you when it's ready (~10 seconds cold).

Open `http://your-domain.com` (or `http://localhost:8080` for a local test) and create your admin account on the setup screen.

### What you get

| `SENDDOCK_LICENSE_KEY` | Behavior |
|---|---|
| empty | Free tier — Core features only (projects, subscribers, templates, transactional sends, broadcasts, campaigns, BYO SMTP, click & open tracking, suppression list, webhook dispatcher). |
| valid Pro key | Pro tier — adds the [Analytics dashboard](/guide/analytics), [Webhooks management UI](/guide/webhooks), and the audit log. |
| valid Team key | Team tier — Pro features plus [multi-user workspaces with roles](/guide/workspaces) and admin user creation. |

The image is the same in all cases. The license key, validated against a hosted endpoint, toggles the gated routes. Read the full tier matrix on the [pricing page](https://senddock.dev/#pricing) or in [Configuration → Plans and licensing](/self-hosting/configuration#plans-and-licensing).

### Pinning a version in production

`:latest` resolves to the most recently tagged release. For production, pin to a specific version so you control when updates land:

```yaml
services:
  senddock:
    image: ghcr.io/arkhe-systems/senddock:0.6.1
```

See available tags on [GHCR](https://github.com/Arkhe-Systems/senddock/pkgs/container/senddock). Only versioned tags (`X.Y.Z`, `X.Y`, `X`) and `:latest` are public — pre-release builds (`:dev`) live in a separate, private package and are not intended for end users.

---

## Option 2: Dokploy (one-click)

If you run [Dokploy](https://dokploy.com) on your server, SendDock is available in the templates marketplace. Open Dokploy → Templates → search for "SendDock" → Install. Dokploy renders the same `docker-compose.image.yml` shown above, prompts you for the env vars, and starts the stack.

This is the recommended path if you already use Dokploy to manage other apps.

---

## Option 3: Build from source

For self-hosters who want to compile Core from source — useful for audits, custom patches, or running an unreleased branch.

```bash
git clone https://github.com/arkhe-systems/senddock.git
cd senddock
cp .env.production.example .env
# edit .env: set JWT_SECRET and POSTGRES_PASSWORD (e.g. openssl rand -hex 32)
docker compose -f docker-compose.prod.yml up -d --build
```

Windows users run the same `docker compose` command from PowerShell — no separate script.

This path runs `docker-compose.prod.yml`, which builds the image locally instead of pulling the prebuilt one. To update later, `git pull && docker compose -f docker-compose.prod.yml up -d --build`. To reset, run `docker compose -f docker-compose.prod.yml down -v && rm .env` (**this deletes all data**, so use it only on test instances) and repeat the fresh-install steps above.

The result is the **Core only** — Pro features (Analytics dashboard, Webhooks management) ship only in the official prebuilt image, not in source builds. To run Pro, use the prebuilt image (Option 1 or 2) and a license key.

### What gets started

- Go backend compiled from source
- Vue frontend built and served by Go
- PostgreSQL 17 database
- Redis for caching
- Database migrations run automatically

---

## What's running

| Service | Description |
|---------|-------------|
| `senddock` | The SendDock binary (Go API + Vue frontend) on port 8080 |
| `postgres` | PostgreSQL 17+ database (named volume `pgdata`) |
| `redis` | Redis 7+ for caching and rate limits (named volume `redisdata`) |

Migrations run inside the `senddock` container's entrypoint on every start, against the existing database. Goose deduplicates already-applied migrations, so restarts and updates never replay them.

---

## Custom port or domain

The `.env` you created controls both:

```bash
PUBLIC_URL=https://your-domain.com
SENDDOCK_PORT=8080
```

`PUBLIC_URL` is required for unsubscribe links and broadcasts to work — without it the broadcast endpoint refuses to send.

::: tip `PORT` vs `SENDDOCK_PORT`
Two separate things. The Go binary always listens on **`PORT`** (default `8080`) **inside** the container. **`SENDDOCK_PORT`** is the **host** port the compose maps to it (`"${SENDDOCK_PORT:-8080}:8080"`). For a stock install you only ever set `SENDDOCK_PORT` — leave `PORT` alone. Set `SENDDOCK_PORT=9090` to expose SendDock on `http://host:9090` while keeping the container's internal port at 8080.
:::

---

## Reverse proxy (HTTPS)

For production, put SendDock behind a reverse proxy with HTTPS.

### Caddy

```
your-domain.com {
    reverse_proxy localhost:8080
}
```

### Nginx

```nginx
server {
    listen 443 ssl;
    server_name your-domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Set `PUBLIC_URL` in `.env` to `https://your-domain.com` and restart.

---

## First-time setup

1. Open your instance in a browser.
2. The setup screen appears automatically (no users exist yet).
3. Create your admin account (name, email, password).
4. You're logged in and ready to create your first project.

---

## Stopping

```bash
docker compose down
```

To stop and remove all data (including database):

```bash
docker compose down -v
```

`down -v` is destructive — it removes the named volumes. Take a backup first if the data matters.

---

## Building from source without Docker

If you prefer not to use Docker at all:

### Requirements

- Go 1.25+
- Node.js 20+
- PostgreSQL 17+ (with the `pg_trgm` extension — auto-enabled by the migration on first start)
- Redis 7+
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`

### Build

```bash
cd frontend && npm ci && npm run build && cd ..
cd backend && make build
```

### Run

```bash
cd backend
cp .env.example .env
# fill DATABASE_URL, JWT_SECRET, PUBLIC_URL
make migrate
./bin/senddock
```

Go serves the Vue frontend automatically from `frontend/dist/`.
