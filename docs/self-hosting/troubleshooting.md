# Troubleshooting & FAQ

Common problems you might run into when self-hosting SendDock, and how to fix them.

## Outgoing emails

### Unsubscribe links don't work / land on a 404

**Symptom:** Recipients click the unsubscribe link inside an email and get a 404 or a "Link expired or invalid" page.

**Cause:** Your `PUBLIC_URL` is wrong (often still `http://localhost:8080`), or the request never reaches the backend because your reverse proxy is routing only `/api/*`.

**Fix:**

1. Open the dashboard → any project → **Settings** → check the "Instance URL" panel. The URL shown there is what every email will use.
2. Set `PUBLIC_URL` in your `.env` to the public domain where SendDock is reachable, e.g. `PUBLIC_URL=https://email.mycompany.com`.
3. Make sure your reverse proxy forwards `/unsubscribe/*` and `/t/*` (open-tracking pixel) to the backend, not just `/api/*`. With the single-binary deploy, both routes are served by the Go process on the same port as the API.
4. Restart the backend so the new value takes effect.

If you migrated from an older version, links generated **before** you set `PUBLIC_URL` are signed against whatever URL/secret was active then — they will fail validation. New emails will work.

### Tracking pixel never registers opens

Same root cause as above. The pixel is `GET /t/{logId}.gif` on the backend. If your reverse proxy only forwards `/api/*`, the pixel returns 404 and opens never get marked.

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

**Fix:**

```bash
cd backend && make migrate
```

Or, if running via Docker, exec into the backend container and run the same command.

### Container crashes on startup with `pq: SSL is not enabled on the server`

**Cause:** Your `DATABASE_URL` has `sslmode=require` but the Postgres instance isn't configured for TLS.

**Fix:** For local Postgres, use `?sslmode=disable`. For managed Postgres (Supabase, RDS, Neon), use `?sslmode=require`.

## Setup script (`setup.sh` / `setup.ps1`)

### `setup.ps1` blocked on Windows: "is not digitally signed"

**Symptom:** Running `.\setup.ps1` produces:

```
.\setup.ps1 cannot be loaded. The file is not digitally signed.
```

**Cause:** Default Windows execution policy blocks unsigned PowerShell scripts. SendDock's `setup.ps1` is plain text, not signed by a certificate.

**Fix:** Allow scripts **for the current session only** — no permanent security change, the bypass reverts when you close the window:

```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
.\setup.ps1
```

### "`.env` already exists. Delete it first if you want to re-run setup"

**Cause:** A previous run of `setup.ps1` / `setup.sh` got far enough to create the `.env` file but failed before finishing (usually because the Docker daemon wasn't ready yet, or an image pull errored). The script refuses to overwrite your `.env` so it doesn't clobber a working config.

**Fix:** Clean the partial state and re-run:

```bash
# Linux / macOS
rm .env
docker compose -f docker-compose.prod.yml down -v
./setup.sh
```

```powershell
# Windows
Remove-Item .env
docker compose -f docker-compose.prod.yml down -v
.\setup.ps1
```

The `down -v` flag also removes the Postgres and Redis volumes — important so the next run doesn't try to run migrations against a half-initialized database.

### `setup.ps1` prints "SendDock is running" but `localhost:8080` refuses the connection

**Cause:** The script reports success based on the file generation steps, not on whether `docker compose up` actually brought the containers online. If an image pull fails partway through, you'll see the green "running" message but no container is actually serving on 8080.

**Fix:** Check what's actually up:

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs app
```

If `app` isn't listed or is restarting, scroll up in the `setup.ps1` / `setup.sh` output for the real error (commonly a Docker pull failure or a migration error). Once you've fixed the underlying issue, follow the cleanup steps from the previous section and re-run setup.

## Redis

### Do I need Redis?

No — Redis is **optional**. It's used for caching email stats and rate-limiting. SendDock works without it; you'll just see slightly slower stats endpoints under heavy load. Leave `REDIS_URL` blank to disable.

### "redis: connection refused"

If you set `REDIS_URL` but Redis is unreachable, SendDock falls back to direct database queries. The error in the logs is harmless — but if you intended Redis to be enabled, check the connection string and that the container is running.

## Asking for help

If your issue isn't here, open an issue on [GitHub](https://github.com/arkhe-systems/senddock/issues) with:

- SendDock version (`docker logs senddock | grep version` or check the bottom of the dashboard)
- Deployment mode (`self-hosted` / `cloud`)
- Relevant logs (redact secrets)
- Steps to reproduce
