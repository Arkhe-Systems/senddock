# Updating

SendDock checks for new releases on its own and tells you when one is available. The dashboard displays the current version next to the logout button. When a newer version is published on GitHub, the version label turns into a yellow "Update available" badge — click it to see what changed and copy the update command.

Behind the scenes the dashboard polls `GET /api/v1/version` once on load. The backend caches the GitHub releases response in Redis for 1 hour, so check traffic stays well inside GitHub's anonymous rate limit even with hundreds of self-hosters.

Two paths exist depending on what you want to do.

| Goal | Command | Loses data? |
|------|---------|-------------|
| Pick up a new release, keep all my data | `git pull && ./setup.sh` | No |
| My install is broken, wipe and start over | `./setup.sh --reset` | **Yes, everything** |

The `setup.sh` / `setup.ps1` script is the single entry point. It detects what state you're in and acts accordingly — there is no separate `update` command to remember.

## Update without losing data

Use this when a new version is available and you want to keep your subscribers, templates, email history, and configuration.

```bash
cd senddock
git pull origin main
./setup.sh        # Linux / macOS
.\setup.ps1       # Windows
```

What is preserved:

- `.env` (secrets, database password, JWT secret, configuration)
- Postgres volume (subscribers, templates, projects, logs, campaigns, API keys)
- Redis volume (caches; not critical, will rebuild)

What happens during the update:

1. Detects the existing `.env` and reuses it.
2. `docker compose build --pull` rebuilds the image with the new code.
3. `docker compose up -d` restarts the services.
4. The `entrypoint.sh` runs any new migrations on the existing database.
5. A health check polls `GET /health` for up to 60 seconds. The script only reports success once SendDock actually responds.

If the build fails or the app never becomes healthy, the script exits non-zero and points at `docker compose logs app`.

## Clean reinstall (wipe everything)

Use this only when:

- Your install is broken in a way that an update can't fix (corrupted volume, password mismatch you can't recover from, etc.).
- You're testing on a throwaway instance.
- You took a backup and want a fresh start.

```bash
./setup.sh --reset        # Linux / macOS
.\setup.ps1 -Reset        # Windows
```

What gets deleted:

- Both Docker volumes (Postgres + Redis) — **all your data**.
- `.env` (regenerated with new secrets).
- Existing containers (recreated from scratch).

After reset, the script behaves like a fresh install: generates a new `.env`, builds the image, starts services, waits for the health check, and prompts you to create the admin account at `http://localhost:8080`.

## Manual update without the setup script

If you want to control each step yourself (CI/CD pipelines, restricted environments, debugging a release):

```bash
cd senddock
git fetch origin
git status                             # confirm there are no local changes you want to keep
git reset --hard origin/main           # or origin/v0.x.x for a specific tag
docker compose -f docker-compose.prod.yml build --pull
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml logs -f app
```

Same idea on Windows:

```powershell
cd senddock
git fetch origin
git reset --hard origin/main
docker compose -f docker-compose.prod.yml build --pull
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml logs -f app
```

What each step does:

- `git fetch origin` — downloads remote refs without touching your working tree.
- `git reset --hard origin/main` — replaces your working tree with the remote contents. **Drops any local edits** to repo files. Your `.env` stays because it is gitignored.
- `docker compose build --pull` — rebuilds the image with the new code and pulls fresh base images.
- `docker compose up -d` — restarts the containers; migrations run automatically inside the entrypoint.
- `docker compose logs -f app` — tail the log so you can see if startup succeeds.

This is exactly what `setup.sh` does in update mode, just spelled out so you can wedge in extra steps (smoke tests, backups, monitoring) between any two lines.

## Backing up before a reset

Before running `--reset` on a production instance, dump the database:

```bash
docker compose -f docker-compose.prod.yml exec -T postgres pg_dump -U senddock senddock > senddock-backup-$(date +%Y%m%d).sql
```

To restore later into a fresh install:

```bash
cat senddock-backup-YYYYMMDD.sql | docker compose -f docker-compose.prod.yml exec -T postgres psql -U senddock senddock
```

## Updating without Docker

If you build SendDock from source instead of using Docker, the update flow is manual:

```bash
cd senddock
git pull origin main
cd frontend && npm ci && npm run build && cd ..
cd backend
make migrate
make build
```

Restart the `senddock` binary.

## Rolling back to a previous version

```bash
git checkout v0.x.x
./setup.sh
```

The script picks up the older code and rebuilds. Data is preserved as long as the older version's schema is compatible with what your current database has — major releases may include irreversible migrations, in which case you need to restore from a backup taken before the upgrade.

To roll back a single migration manually (without Docker):

```bash
goose -dir backend/migrations postgres "$DATABASE_URL" down
```

## Checking the current version

The latest release and changelog live on [GitHub releases](https://github.com/arkhe-systems/senddock/releases).
