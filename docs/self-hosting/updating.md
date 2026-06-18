# Updating

SendDock checks for new releases on its own and tells you when one is available. The dashboard displays the current version next to the logout button. When a newer version is published on GitHub, the version label turns into a yellow "Update available" badge — click it to see what changed and copy the update command.

Behind the scenes the dashboard polls `GET /api/v1/version` once on load. The backend caches the GitHub releases response in Redis for 1 hour, so check traffic stays well inside GitHub's anonymous rate limit even with hundreds of self-hosters.

The update path depends on which install you started from.

| Install path | Update command |
|---|---|
| install.sh (Option 1) | **"Update now"** in the dashboard nav (Watchtower auto-wired) — or `docker compose pull && docker compose up -d` as a manual fallback |
| Manual Docker Compose (Option 2) | `docker compose pull && docker compose up -d` |
| Dokploy (Option 3) | One-click **"Redeploy"** in the Dokploy UI |
| Build from source (Option 4) | `git pull && docker compose -f docker-compose.prod.yml up -d --build` |

In every case **your data is preserved**. Postgres lives in a named Docker volume that is not touched by image updates. Migrations are run by `goose`, which deduplicates against the `goose_db_version` table and only applies what hasn't been applied before.

---

## Updating a prebuilt image install

Use this when you installed via the canonical `ghcr.io/arkhe-systems/senddock` image — either through `install.sh` (Option 1) or Manual Docker Compose (Option 2) in [Installation](./installation). If `install.sh` set up Watchtower for you, **click "Update now"** in the dashboard nav instead of running the command below; the result is identical.

```bash
cd senddock
docker compose pull
docker compose up -d
docker compose logs -f senddock
```

What happens:

1. `docker compose pull` fetches the latest `ghcr.io/arkhe-systems/senddock` image from GitHub Container Registry. Postgres and Redis images update too unless you've pinned them.
2. `docker compose up -d` recreates the `senddock` container with the new image. Postgres and Redis containers stay running unless their image changed.
3. The new container's entrypoint runs `goose -dir /app/migrations postgres "$DATABASE_URL" up`, which applies any new migrations. Existing rows in `goose_db_version` are skipped.
4. The healthcheck in `docker-compose.yml` flips to "healthy" once `/health` returns 200 — usually within 10 seconds.

What is preserved:

- `pgdata` volume (subscribers, templates, projects, logs, campaigns, API keys, audit log)
- `redisdata` volume (caches; not critical, will rebuild)
- `.env` (secrets, license key, configuration)

Pin the version in `docker-compose.yml` to control update timing:

```yaml
services:
  senddock:
    image: ghcr.io/arkhe-systems/senddock:0.6.1
```

When you decide to upgrade, edit the tag, then `docker compose pull && docker compose up -d`.

### License key on update

The license validator caches its last-good response for 1 hour. After an update the new container runs through validation again on first startup. If your subscription is active, Pro features unlock immediately. If the license has expired or been revoked, Pro routes start returning `402 Payment Required` on the next validation tick — Core features remain fully available.

---

## Updating a Dokploy install

Open the SendDock app in Dokploy → click "Redeploy". Dokploy runs `docker compose pull && docker compose up -d` against the same volumes.

If you switch to a pinned version, edit the env var in Dokploy and Redeploy. The volume is not touched.

---

## Updating a source build

Use this when you installed via `git clone` + `docker compose -f docker-compose.prod.yml`.

```bash
cd senddock
git pull origin main
docker compose -f docker-compose.prod.yml up -d --build
```

Windows users run the same command from PowerShell.

What happens:

1. Your existing `.env` is reused.
2. `docker compose -f docker-compose.prod.yml up -d --build` rebuilds the image with the new code and recreates the services.
3. The container's entrypoint runs any new migrations on the existing database.
4. The healthcheck flips to "healthy" once `/health` returns 200.

Same data preservation guarantees as the image flow.

If the build fails or the app never becomes healthy, check `docker compose -f docker-compose.prod.yml logs senddock`.

---

## One-click updates from the dashboard (Watchtower)

The dashboard ships with an **"Update available"** badge in the top nav. Click it → modal opens with release notes → **"Update now"** button → SendDock asks [Watchtower](https://containrrr.dev/watchtower/) to pull the new image and recreate the container. The dashboard polls `/api/v1/version` every 2 seconds and shows a success toast when the new version is live. Postgres + Redis volumes are preserved.

::: tip Already wired up if you used setup.sh
The `scripts/setup.sh` installer drops a `docker-compose.override.yml` next to the main compose file with Watchtower fully wired (label-scoped to SendDock, HTTP-API-only, no auto-polling). The "Update now" button works out of the box — no extra steps. Skip the manual block below.
:::

### Manual setup (non-setup.sh installs)

Add Watchtower to your `docker-compose.yml` (or as a separate `docker-compose.override.yml`):

```yaml
services:
  watchtower:
    # Maintained fork — the original containrrr/watchtower was abandoned in
    # 2023 and its :latest image ships a Docker client too old for modern
    # daemons (API 1.25 vs required 1.40+). Use the nickfedor fork instead.
    image: nickfedor/watchtower:latest
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      WATCHTOWER_HTTP_API_UPDATE: "true"
      WATCHTOWER_HTTP_API_TOKEN: "change-this-to-a-long-random-string"
      WATCHTOWER_HTTP_API_PERIODIC_POLLS: "false"
      WATCHTOWER_CLEANUP: "true"
      WATCHTOWER_LABEL_ENABLE: "true"

  senddock:
    # ... your existing config ...
    labels:
      com.centurylinklabs.watchtower.enable: "true"
    environment:
      # ... your existing env vars ...
      SENDDOCK_WATCHTOWER_URL: http://watchtower:8080
      SENDDOCK_WATCHTOWER_TOKEN: change-this-to-a-long-random-string
```

What each piece does:

- `WATCHTOWER_HTTP_API_UPDATE=true` exposes `POST /v1/update` on port 8080 inside the Docker network.
- `WATCHTOWER_HTTP_API_TOKEN` — bearer token required for that endpoint. Use a long random string (`openssl rand -hex 32`); the same value goes into `SENDDOCK_WATCHTOWER_TOKEN`.
- `WATCHTOWER_HTTP_API_PERIODIC_POLLS=false` disables Watchtower's background polling. Updates only fire when you click the button — no surprise auto-updates.
- `WATCHTOWER_LABEL_ENABLE=true` + the `com.centurylinklabs.watchtower.enable` label on SendDock scopes Watchtower to only the SendDock container. Without this it would manage every container on the host.
- `WATCHTOWER_CLEANUP=true` removes the old image after a successful update so disk doesn't bloat.
- `SENDDOCK_WATCHTOWER_URL` / `SENDDOCK_WATCHTOWER_TOKEN` tell SendDock where to call. **Both must be set together** — `URL` alone leaves the dashboard with no credential and the update fails silently with a 401. When both are set, the modal shows the "Update now" button. When unset, the modal still surfaces the update — it just falls back to a platform-agnostic prompt (rebuild from your hosting panel, or run the docker compose command).

When you click "Update now":

1. SendDock POSTs to `http://watchtower:8080/v1/update` with the bearer token.
2. Watchtower pulls the latest `ghcr.io/arkhe-systems/senddock` image and recreates the container.
3. The dashboard polls `/api/v1/version` every 2 seconds; it tolerates the brief network errors while the container restarts, then sees the new version and shows a success toast.

**Security notes:**

- Watchtower mounts `/var/run/docker.sock` — anything in that container effectively controls Docker on the host. Keep Watchtower's image up to date and don't expose its port outside the Docker network.
- The HTTP API token should be a long random string. Treat it like a credential.
- The label-scoped mode means Watchtower only touches containers carrying the `watchtower.enable` label, which limits blast radius if something goes wrong.

---

## Update API endpoints

The dashboard's "Update available" badge and the Watchtower one-click button are built on three small admin endpoints. They're documented here so you can drive them from your own tooling (status pages, custom dashboards, CI checks). All three are cookie-auth — they're not callable with a project-scoped API key.

### `GET /api/v1/version`

```json
{
  "current": "0.6.5.1",
  "latest": "0.6.5.1",
  "update_available": false,
  "release_notes": "...",
  "release_url": "https://github.com/arkhe-systems/senddock/releases/tag/v0.6.5.1"
}
```

`current` is the version baked into the running binary; `latest` is the most recent published release fetched from GitHub. The backend caches the GitHub response in Redis for 1 hour, so polling this endpoint is cheap and stays well inside GitHub's anonymous rate limit even with hundreds of self-hosters.

### `GET /api/v1/update/auto-status`

```json
{
  "configured": true,
  "healthy": true,
  "url": "http://watchtower:8080",
  "last_check": "2026-06-16T10:00:00Z",
  "last_error": null
}
```

Tells the dashboard whether the Watchtower integration is wired up and reachable. `configured` is `true` when both `SENDDOCK_WATCHTOWER_URL` and `SENDDOCK_WATCHTOWER_TOKEN` are set; `healthy` is `true` when the most recent ping to Watchtower's `/v1/update` endpoint succeeded. Cached for 30 seconds.

### `POST /api/v1/update/trigger`

Fires the Watchtower call in a goroutine and returns immediately with `202 Accepted`. The dashboard polls `/api/v1/version` every two seconds afterward to detect the new version once the container restarts.

Returns `400` if Watchtower isn't configured (`SENDDOCK_WATCHTOWER_URL`/`TOKEN` unset), `502` if Watchtower is configured but unreachable.

---

## Manual update without scripts or Compose

For CI/CD pipelines, restricted environments, or step-by-step debugging:

```bash
docker pull ghcr.io/arkhe-systems/senddock:latest
docker stop senddock-old
docker rm senddock-old
docker run -d \
    --name senddock \
    --env-file /etc/senddock/.env \
    --network senddock-net \
    -p 8080:8080 \
    ghcr.io/arkhe-systems/senddock:latest
```

Whatever orchestrator you use (Kubernetes, Nomad, plain systemd-with-podman), the steps boil down to:

1. Pull the new image.
2. Stop the old container.
3. Start a new container against the same `DATABASE_URL` and `JWT_SECRET`.
4. The entrypoint runs `goose up`. New migrations apply, old ones are skipped.

---

## Backup & restore

While the update path itself is safe, a regular backup is the only thing standing between you and a corrupted volume, a fat-fingered `down -v`, or a host disk failure.

### What to back up

| What | Where it lives | Need it? |
|---|---|---|
| Postgres data | `pgdata` volume | **Yes** — subscribers, templates, projects, logs, campaigns, API keys, audit log, encrypted SMTP credentials. |
| `.env` | Repo root on the host | **Yes** — `JWT_SECRET` rotates click-tracking + session tokens, `POSTGRES_PASSWORD` is needed to read the dump back. Losing `JWT_SECRET` invalidates every active session and every unsubscribe / tracking URL signed against it. |
| `docker-compose.yml` | Repo root on the host | Helpful — captures any tag pin, port override, Watchtower wiring you've added. |
| Redis data | `redisdata` volume | No — only caches and counters; rebuilds from the database on next boot. |

### Take a snapshot

```bash
# Inside the compose dir
docker compose exec -T postgres pg_dump -U senddock senddock \
    > senddock-backup-$(date +%F).sql
```

For a smaller, faster restore use the custom format:

```bash
docker compose exec -T postgres pg_dump -U senddock -Fc senddock \
    > senddock-backup-$(date +%F).dump
```

Whichever format you pick, also copy `.env` somewhere off-host the same day:

```bash
cp .env env-backup-$(date +%F).env
```

### Restore into a fresh install

`.sql` (plain SQL):

```bash
cat senddock-backup-YYYY-MM-DD.sql | \
    docker compose exec -T postgres psql -U senddock senddock
```

`.dump` (custom format) — `pg_restore` lets you parallelise and skip objects:

```bash
docker compose cp senddock-backup-YYYY-MM-DD.dump postgres:/tmp/in.dump
docker compose exec postgres pg_restore -U senddock -d senddock --clean --if-exists -j 4 /tmp/in.dump
```

The restored database keeps every project, subscriber, template, suppression entry, and log row. Restart SendDock so any in-memory state reseeds.

### Recommended cadence

- **Daily** Postgres dump, retained 7 days locally + 30 days off-site (S3, Backblaze, or whatever bucket you trust).
- **Weekly** full backup that also captures `.env` and `docker-compose.yml`.
- **Before every upgrade** — a one-off dump labeled with the source and target version (`senddock-pre-0.6.5.sql`). This is the lowest-friction rollback path if a migration goes sideways.

A minimal cron job on the host:

```bash
0 3 * * * cd /opt/senddock && docker compose exec -T postgres pg_dump -U senddock -Fc senddock > /var/backups/senddock/$(date +\%F).dump && find /var/backups/senddock -mtime +7 -delete
```

### Verify a backup actually restores

The only thing worse than no backup is a backup you've never restored. Once a month, spin up a throwaway Postgres container, `pg_restore` the latest dump into it, and confirm the row counts match production. Five minutes that turn an untested artifact into a tested recovery path.

### Off-site & encryption

Local backups protect against application-level corruption — they don't protect against the whole host going away. Sync the daily dump to object storage with server-side encryption enabled (S3 SSE-S3 / SSE-KMS, Backblaze B2 default-at-rest, GCS CMEK). If your dataset is small enough and you don't want to manage object storage, `restic` or `rclone crypt` to any remote with client-side encryption works just as well.

Encrypt before upload if regulation or sensitivity demands it — `gpg -c senddock-backup-$(date +%F).dump` or `age -p` against a passphrase you store outside the same provider.

---

## Rolling back

### Image install

Edit `docker-compose.yml` to point at the previous tag, then:

```bash
docker compose pull
docker compose up -d
```

Data is preserved as long as the older version's schema is compatible with what your current database has. Major releases may include irreversible migrations — if rolling back across one, restore from the backup you took before the upgrade.

### Source build

```bash
git checkout v0.x.x
docker compose -f docker-compose.prod.yml up -d --build
```

This picks up the older code and rebuilds.

### Manual migration rollback

If a single migration is the problem:

```bash
docker compose exec senddock goose -dir /app/migrations postgres "$DATABASE_URL" down
```

Run it once per migration you want to undo. Verify the application still starts before doing more.

---

## Wiping and reinstalling

Use this only when:

- Your install is broken in a way that an update can't fix (corrupted volume, password mismatch you can't recover from).
- You're testing on a throwaway instance.
- You took a backup and want a fresh start.

### Image install

```bash
docker compose down -v
docker compose up -d
```

`down -v` removes the named volumes — **all data is gone**.

### Source build

```bash
docker compose -f docker-compose.prod.yml down -v
rm .env
```

Then recreate `.env` from `.env.production.example` and bring the stack back up with `docker compose -f docker-compose.prod.yml up -d --build`.

What gets deleted in either flow:

- Both Docker volumes (Postgres + Redis) — all your data.
- For source builds, `.env` is removed so you regenerate it with new secrets.

After reset, the next start behaves like a fresh install: setup screen appears, create the admin account again.

---

## Checking the current version

The latest release and changelog live on [GitHub releases](https://github.com/arkhe-systems/senddock/releases). The image tag matching that release is published to [GHCR](https://github.com/Arkhe-Systems/senddock/pkgs/container/senddock) within a few minutes of the GitHub release going live.
