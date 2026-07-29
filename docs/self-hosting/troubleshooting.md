# Troubleshooting & FAQ

Common problems you might run into when self-hosting SendDock, and how to fix them.

## Outgoing emails

### Test Connection times out / "could not reach SMTP server within 10s"

**Symptom:** SMTP Settings → Test Connection sits for 10 seconds, then errors with a message about ISPs blocking outbound SMTP. The same SMTP credentials work fine from another machine (cloud, office, phone tether).

**Cause:** Your network — almost certainly a residential ISP (Comcast, Claro, Movistar, BT, and most others worldwide) — blocks outbound TCP on ports 25, 465 and 587 to stop spam botnets. The block is at the ISP's edge router, so the TCP packet never leaves your house. Every email tool hits the same wall — it is not specific to SendDock.

**Fix (in order of preference):**

1. **Run SendDock on a public cloud server** (DigitalOcean, Hetzner, AWS, etc.). Cloud providers don't apply the residential SMTP block. This is the supported production deployment.
2. **Ask your SMTP provider for port `2525`.** SendGrid, Mailgun, Postmark, Resend and Brevo all listen on `2525` specifically as an escape hatch. If your provider does, change Port to `2525` and Test Connection again.
3. **Use a VPN** that doesn't apply ISP filters. Paid VPNs (NordVPN, Mullvad, ProtonVPN) work; free ones often also block SMTP.
4. **For local dev**, use [Mailpit](https://mailpit.axllent.org/) on `localhost:1025` — see the [SMTP guide](/guide/smtp#testing) for the workflow. Mailpit captures sends without ever leaving the host.

In other words: SendDock cannot deliver mail from a network that blocks outbound mail ports. There is no software workaround for ISP egress filtering.

### 429 "rate limit exceeded for this project on this endpoint"

**Cause:** You hit the per-project rate limit on a sending endpoint. The limits are intentional spam protection. See [Rate limits](/guide/sending#rate-limits-and-abuse-prevention) for the full table.

**Fix:**

- Check the `Retry-After` header for how long to wait before retrying.
- If you legitimately need higher throughput, switch to **subscribers + broadcast** (a single broadcast call to a 50k list is one request, not 50k).
- If you are looping over `/send` to reach all your subscribers, that is exactly the pattern these limits are meant to block — store the recipients as subscribers and call `/broadcast` once.
- If your tests are tripping the limit, scope them to use a unique project per test run, or unset `REDIS_URL` in your test environment to disable rate limiting (do not do this in production).

### "Newsletters are disabled" banner / cannot send broadcasts

**Symptom:** The Newsletters page shows a yellow banner saying broadcasts are disabled, the "+ New Campaign" button is greyed out, or the API returns 400 with an error mentioning that your public URL is not publicly reachable.

**Cause:** SendDock refuses to send broadcasts and schedule campaigns when your public URL is unset or resolves to `localhost` / `127.0.0.1` / `::1`. Without a public URL, the unsubscribe links inside outgoing emails would not work for recipients, which is the canonical spam pattern.

**Fix:**

1. Put SendDock behind a real domain (reverse proxy + DNS — see [Reverse Proxy](/self-hosting/installation#reverse-proxy-https)).
2. Set your public URL to `https://your-domain.com` under **Instance** in the dashboard (no trailing slash). It applies immediately — no restart. See [Instance settings](/guide/instance-settings#public-url).

Single-recipient sends (`/send`) and `/send/batch` continue to work without a public URL, so transactional flows (password resets, contact-form notifications) do not require a public domain.

### Unsubscribe links don't work / land on a 404

**Symptom:** Recipients click the unsubscribe link inside an email and get a 404 or a "Link expired or invalid" page.

**Cause:** Your public URL is wrong (often still `http://localhost:8080`), or the request never reaches the backend because your reverse proxy is routing only `/api/*`.

**Fix:**

1. Open the dashboard → any project → **Settings** → check the "Instance URL" panel. The URL shown there is what every email will use.
2. Set your public URL under **Instance** in the dashboard to the domain where SendDock is reachable, e.g. `https://email.mycompany.com`.
3. Make sure your reverse proxy forwards `/unsubscribe/*` and `/t/*` (open-tracking pixel) to the backend, not just `/api/*`. With the single-binary deploy, both routes are served by the Go process on the same port as the API.
4. Restart the backend so the new value takes effect.

If you migrated from an older version, links generated **before** you set your public URL are signed against whatever URL/secret was active then — they will fail validation. New emails will work.

### Tracking pixel never registers opens

Same root cause as above. The pixel is `GET /t/{logId}` on the backend (returns a 1×1 transparent GIF; no file extension on the path). If your reverse proxy only forwards `/api/*`, the pixel returns 404 and opens never get marked.

### "Sender address rejected" / SMTP authentication failed

**Cause:** The SMTP `From` address you configured doesn't match the authenticated user, or your provider requires SPF/DKIM alignment.

**Fix:** In **Project → SMTP**, set the **From Email** field to an address on a domain you own and have authorized. Most providers (SES, Mailgun, Postmark) reject sends from unverified domains.

### TLS certificate errors (`x509: certificate has expired`)

**Cause:** Your SMTP server is presenting an expired or self-signed certificate. This is common with self-hosted Poste.io, Mailcow, or Mail-in-a-Box instances when the Let's Encrypt cert hasn't auto-renewed.

**Fix options (in order of preference):**

1. **Renew the certificate** on your mail server. This is the right fix.
2. **Switch port:** if you're on `465` (implicit TLS) and getting handshake errors, try `587` (STARTTLS) or vice-versa. SendDock auto-detects from the port.
3. **Last resort:** SendDock currently skips strict certificate validation for outgoing SMTP connections so a flaky cert doesn't block your sends. Mail still goes out over TLS, you just don't get man-in-the-middle protection. Renew the cert ASAP.

### `connection refused` when sending

**Cause:** Wrong host/port, or the mail server is firewalled.

**Fix:** From the SendDock host, confirm you can reach the SMTP server: `nc -vz smtp.example.com 587`. If that fails, the problem is your network/firewall, not SendDock. Cloud providers often block port 25 — use 587 or 465 with auth.

## Authentication & cookies

### "missing access token" on every request after deploy

**Cause:** Your frontend is on HTTPS but cookies are being set without `Secure: true`, so the browser drops them.

**Fix:** Ensure your `FRONTEND_URL` env starts with `https://` in production. SendDock detects this and sets `Secure: true` on auth cookies automatically. If you reverse-proxy TLS, also make sure the proxy forwards the `X-Forwarded-Proto` header.

### Login works but the next request fails / "401 unauthorized"

**Cause:** Backend and frontend are on different origins and CORS isn't configured to send credentials.

**Fix:** Set `FRONTEND_URL` to the exact origin of your dashboard, e.g. `https://email.mycompany.com` (no trailing slash, scheme included). The CORS middleware uses this to set `Access-Control-Allow-Origin` and `Access-Control-Allow-Credentials: true`.

### `JWT_SECRET must be at least 32 characters` panic at startup

**Cause:** The placeholder secret was kept, or a weak secret was set.

**Fix:** Generate a strong random secret:

```bash
openssl rand -base64 48
```

Paste it into `JWT_SECRET` in `.env` and restart.

## Networking & deploys

### Container starts but I can't reach it on the public domain

**Cause:** No reverse proxy is forwarding traffic to the SendDock container, or the proxy only forwards `/api/*`.

**Fix:** SendDock is a single binary that serves both the API and the SPA on one port. Forward **everything** for your domain to the backend port (default `8080`). Example Caddy:

```caddy
email.mycompany.com {
    reverse_proxy senddock:8080
}
```

### Frontend hits `localhost:8080` in production

**Cause:** The frontend was built without `VITE_API_URL` set, so it baked in the dev default.

**Fix:** Either rebuild with `VITE_API_URL=/api/v1` (recommended for single-binary deploys), or set it to your full backend URL like `VITE_API_URL=https://email.mycompany.com/api/v1`.

### Browser shows CORS error on the public waitlist endpoint

**Cause:** You're embedding the waitlist form on a different domain than your SendDock instance.

**Fix:** This is intentional for the public waitlist endpoint — it explicitly sets `Access-Control-Allow-Origin: *` so you can embed the form anywhere. If you're getting a CORS error on this endpoint, double-check that you're hitting `POST /api/v1/projects/{id}/waitlist` (not the protected `/subscribers` endpoint).

## Database & migrations

### "relation does not exist" errors after upgrade

**Cause:** Pending migrations weren't applied.

**Fix in production (Docker):** migrations run automatically via `goose` on container startup — the entrypoint script blocks until they finish before exec'ing the binary. If you see this error after pulling a new image, the container almost certainly hit the error during boot. Check the logs:

```bash
docker compose logs app --tail 100 | grep -i 'goose\|migration'
```

If goose printed an error (e.g. couldn't connect to Postgres, or hit a conflicting schema), fix the underlying cause and restart with `docker compose up -d`. The next boot retries from where the last successful version stopped.

**Fix in dev (running from source):**

```bash
cd backend && make migrate
```

### Container crashes on startup with `pq: SSL is not enabled on the server`

**Cause:** Your `DATABASE_URL` has `sslmode=require` but the Postgres instance isn't configured for TLS.

**Fix:** For local Postgres, use `?sslmode=disable`. For managed Postgres (Supabase, RDS, Neon), use `?sslmode=require`.

## Source build (`docker-compose.prod.yml`)

### Postgres "password authentication failed" after rebuilding from source

**Cause:** A previous source build created a Postgres volume with one password, then `.env` was deleted and regenerated with a new password. Postgres does not re-initialize when its data directory already has data, so it keeps the old password — the app then fails to authenticate against it.

**Fix:** Tear down the volumes and `.env` together, then start fresh:

```bash
docker compose -f docker-compose.prod.yml down -v
rm .env
```

This wipes the Postgres volume. Recreate `.env` from `.env.production.example` and bring the stack back up with `docker compose -f docker-compose.prod.yml up -d --build`, which gives you matching credentials.

### App container never becomes healthy after a source build

**Cause:** `GET /health` on the app container doesn't respond. The container is either still building, restarting, or failing on startup.

**Fix:** Inspect what the app container is doing:

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs app --tail 100
```

Common causes:

- The build was very slow (large npm install, slow disk). Wait another minute and check `docker compose ps` — if `app` is `Up`, you are fine; just refresh the browser.
- Migration failure — usually a Postgres password mismatch, see the section above on credential mismatches.
- `JWT_SECRET must be at least 32 characters` — fix the value in `.env` and bring the stack back up.

## Docker Swarm / Dokploy

### All services suddenly stop talking to each other ("the whole stack went down at once")

**Symptom:** Multiple services on the same Swarm host (your SendDock backend, Postgres, Redis, often also unrelated apps) become unreachable at the same time. The backend logs show `dial tcp: lookup <service-name> on 127.0.0.11:53: no such host` for hostnames that *do* exist. Containers exit cleanly (`Exited (0)`) every 2-3 minutes in a crash-replace loop, even though individual processes were healthy. Restarting the affected services manually brings everything back temporarily, only for the same outage to return hours or days later.

**Cause:** Docker Swarm corruption from accumulated dead containers and orphaned services on the same host. The embedded DNS resolver `127.0.0.11:53` (which is how containers find each other by service name on the `dokploy-network` overlay or any overlay network) keeps stale references for every service that was created and then deleted, plus every dead-but-not-pruned container record. Past a certain threshold the resolver intermittently fails to resolve names that actually exist — and because all services on the same overlay use the same resolver, **they all fail simultaneously**, which masquerades as a "the database died" issue when it is really a swarm-state issue.

This is **not** specific to SendDock — any Dokploy / Docker Swarm host accumulating long-lived deployments is vulnerable, especially if you have created and destroyed services repeatedly during development.

**Diagnose:** SSH to the host and check for accumulated dead containers and orphan services:

```bash
# count Dead containers (zero is healthy; tens or hundreds means the swarm is full of corpses)
docker ps -aq --filter 'status=dead' | wc -l

# any service that consumes CPU continuously even when idle is a crash-loop suspect
docker stats --no-stream --format 'table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}'

# if you find services here that you do not recognize or whose database has been deleted,
# they are orphans
docker service ls --format 'table {{.Name}}\t{{.Replicas}}\t{{.Image}}'
```

A single container chewing >10% CPU continuously while doing nothing useful is almost always a crash-loop pinned to a missing dependency (deleted database, deleted Redis, etc.).

**Fix:**

1. **Remove orphan services first** — anything in `docker service ls` that no longer has its database / dependencies. From the Dokploy UI if it still appears there, otherwise:
   ```bash
   docker service rm <orphan-service-name>
   ```
2. **Force-kill any container that survives** (sometimes an old container outlives its service if it stops responding to SIGTERM):
   ```bash
   docker rm -f <container-id>
   ```
3. **Prune dead containers and unused networks** (safe — does not touch volumes, images, or running services):
   ```bash
   docker container prune -f
   docker network prune -f
   ```
4. **Do not** run `docker volume prune` or `docker system prune -a` — those will delete Postgres data and the SendDock image you just built.

After cleanup, the embedded DNS resolver returns to a stable state and the recurring "the whole stack went down" pattern stops.

### Repeated Dokploy "Image is up to date" deploys do not pick up a new build

**Symptom:** You triggered a new image build, but Dokploy logs say `Status: Image is up to date for ghcr.io/...:dev` and the running container still serves the old code. The Dokploy webhook returns success.

**Cause:** Docker is caching the image manifest by tag, not by digest. When the registry tag (`:dev`, `:latest`) is reused, `docker pull` looks at the local cache first, sees a tag with that name, and skips the pull. The new manifest never reaches the host.

**Fix:** In the Dokploy UI for the application, toggle on **Clean Cache** in **General** before clicking **Deploy** or **Rebuild**. That forces an unconditional pull. From the host directly, the same effect:

```bash
docker service update --force --image ghcr.io/your-org/your-image:dev <service-name>
```

The `--force` flag rotates the task even when the image reference appears unchanged, and combined with `--image` triggers a manifest re-pull.

## Redis

### Do I need Redis?

**The binary boots without Redis, but you almost certainly do need it.** `REDIS_URL` is technically optional, and when unset:

- The **global per-IP rate limiter** (default 600 req/min, every endpoint except `/health`) becomes a no-op.
- The **per-project sending limits** on `/send`, `/send/batch`, `/broadcast` become no-ops.
- The **GitHub releases cache** is skipped — every dashboard load hits the GitHub API directly (you'll burn through the anonymous rate limit fast on busy instances).
- Email stats fall back to direct database queries — slightly slower, but functionally identical.

For any production instance reachable from the internet, run Redis. The bundled production composes already include it; there's no good reason to disable it outside of throwaway test environments where you've explicitly accepted there are no rate limits.

### "redis: connection refused"

If `REDIS_URL` is set but Redis is unreachable, SendDock falls back to direct database queries for stats and **disables rate limiting entirely** until Redis is back. The error in the logs is recoverable but should not be ignored — check the connection string and that the container is running.

## Asking for help

If your issue isn't here, open an issue on [GitHub](https://github.com/arkhe-systems/senddock/issues) with:

- SendDock version (`docker logs senddock | grep version` or check the bottom of the dashboard)
- Deployment mode (`self-hosted` / `cloud`)
- Relevant logs (redact secrets)
- Steps to reproduce
