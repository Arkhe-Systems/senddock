# Updating

SendDock checks for new releases on its own and tells you when one is available. The dashboard displays the current version next to the logout button. When a newer version is published on GitHub, the version label turns into a yellow "Update available" badge — click it to see what changed and copy the update command.

Behind the scenes the dashboard polls `GET /api/v1/version` once on load. The backend caches the GitHub releases response in Redis for 1 hour, so check traffic stays well inside GitHub's anonymous rate limit even with hundreds of self-hosters.

The update path depends on which install you started from.

| Install path | Update command |
|---|---|
| Prebuilt image (Option 1) | `docker compose pull && docker compose up -d` |
| Dokploy (Option 2) | One-click "Redeploy" in the Dokploy UI |
| Build from source (Option 3) | `git pull && ./setup.sh` |

In every case **your data is preserved**. Postgres lives in a named Docker volume that is not touched by image updates. Migrations are run by `goose`, which deduplicates against the `goose_db_version` table and only applies what hasn't been applied before.

---

## Updating a prebuilt image install

Use this when you installed via the canonical `ghcr.io/arkhe-systems/senddock` image (Option 1 in [Installation](./installation)).

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

Use this when you installed via `git clone` + `./setup.sh`.

```bash
cd senddock
git pull origin main
./setup.sh        # Linux / macOS
.\setup.ps1       # Windows
```

What happens:

1. Detects the existing `.env` and reuses it.
2. `docker compose -f docker-compose.prod.yml build --pull` rebuilds the image with the new code and pulls fresh base images.
3. `docker compose up -d` restarts the services.
4. The container's entrypoint runs any new migrations on the existing database.
5. A health check polls `GET /health` for up to 60 seconds. The script only reports success once SendDock actually responds.

Same data preservation guarantees as the image flow.

If the build fails or the app never becomes healthy, the script exits non-zero and points at `docker compose logs senddock`.

---

## One-click updates from the dashboard (Watchtower)

If you want the "Update now" button in the dashboard's update modal to actually pull the new image and recreate the container, run [Watchtower](https://containrrr.dev/watchtower/) alongside SendDock and point SendDock at its HTTP API.

Add Watchtower to your `docker-compose.yml`:

```yaml
services:
  watchtower:
    image: containrrr/watchtower
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      WATCHTOWER_HTTP_API_UPDATE: "true"
      WATCHTOWER_HTTP_API_TOKEN: "change-this-to-a-long-random-string"
      WATCHTOWER_CLEANUP: "true"
      WATCHTOWER_LABEL_ENABLE: "true"
    command: --interval 86400

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
- `WATCHTOWER_HTTP_API_TOKEN` is a bearer token required for that endpoint — use a long random string, the same value goes into `SENDDOCK_WATCHTOWER_TOKEN`.
- `WATCHTOWER_LABEL_ENABLE=true` + the `com.centurylinklabs.watchtower.enable` label on SendDock scopes Watchtower to only update opted-in containers. Without this it would update everything on the host.
- `WATCHTOWER_CLEANUP=true` removes the old image after a successful update so disk doesn't bloat.
- `--interval 86400` makes Watchtower poll every 24 hours on its own. Set it higher (or pass `--no-pull` and skip the interval) if you only want updates triggered from the SendDock UI.
- `SENDDOCK_WATCHTOWER_URL` / `SENDDOCK_WATCHTOWER_TOKEN` tell SendDock where to call. When set, the update modal shows a one-click "Update now" button. When unset, it shows the manual command as before.

When you click "Update now":

1. SendDock POSTs to `http://watchtower:8080/v1/update` with the bearer token.
2. Watchtower pulls the latest `ghcr.io/arkhe-systems/senddock` image and recreates the container.
3. The dashboard polls `/api/v1/version` every 2 seconds; it briefly fails while the container restarts, then sees the new version and shows a success toast.

**Security notes:**

- Watchtower needs `/var/run/docker.sock` mounted — anything in that container effectively controls Docker on the host. Keep Watchtower's image up to date and don't expose its port outside the Docker network.
- The HTTP API token should be a long random string. Treat it like a credential.
- The label-scoped mode means Watchtower only touches containers with the `watchtower.enable` label, which limits blast radius if something goes wrong.

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

## Backing up before an update

While the update path itself is safe, taking a snapshot before any update is good hygiene.

```bash
docker compose exec -T postgres pg_dump -U senddock senddock \
    > senddock-backup-$(date +%Y%m%d).sql
```

To restore later into a fresh install:

```bash
cat senddock-backup-YYYYMMDD.sql | \
    docker compose exec -T postgres psql -U senddock senddock
```

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
./setup.sh
```

The script picks up the older code and rebuilds.

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
./setup.sh --reset        # Linux / macOS
.\setup.ps1 -Reset        # Windows
```

What gets deleted in either flow:

- Both Docker volumes (Postgres + Redis) — all your data.
- For source builds, `.env` is regenerated with new secrets.

After reset, the next start behaves like a fresh install: setup screen appears, create the admin account again.

---

## Checking the current version

The latest release and changelog live on [GitHub releases](https://github.com/arkhe-systems/senddock/releases). The image tag matching that release is published to [GHCR](https://github.com/Arkhe-Systems/senddock/pkgs/container/senddock) within a few minutes of the GitHub release going live.
