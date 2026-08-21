# Backups & Recovery

SendDock keeps everything that matters in one Postgres database — subscribers, templates, projects, campaigns, logs, API keys, the audit log, encrypted SMTP credentials, and (on self-hosted) your public URL, session timeout and activated license key. A regular, tested backup is the only thing standing between you and a corrupted volume, a fat-fingered `down -v`, or a host disk failure.

::: warning Why this matters more than it looks
Migrations are **forward-only** and run automatically on container start. They're additive today — destructive DDL lives only in goose `Down` blocks, which an update never runs — so the everyday risk isn't data loss, it's **no rollback path**: once a migration runs forward, the previous image can't read the new schema. If a maintenance window goes sideways, restoring a dump you took **before** the update is your way back. Take one every time.
:::

## What to back up

| What | Where it lives | Need it? |
|---|---|---|
| Postgres data | `pgdata` volume | **Yes** — subscribers, templates, projects, logs, campaigns, API keys, audit log, encrypted SMTP credentials. |
| `.env` | Repo root on the host | **Yes** — `JWT_SECRET` signs click-tracking, unsubscribe and session tokens; `POSTGRES_PASSWORD` is needed to read the dump back. Losing `JWT_SECRET` invalidates every active session and every tracking / unsubscribe URL signed against it. |
| `docker-compose.yml` | Repo root on the host | Helpful — captures any tag pin, port override or Watchtower wiring you've added. |
| Redis data | `redisdata` volume | No — only caches and counters; rebuilds from the database on next boot. |

## Take a snapshot

Plain SQL — portable and human-readable:

```bash
docker compose exec -T postgres pg_dump -U senddock senddock \
    > senddock-backup-$(date +%F).sql
```

Custom format — smaller, and `pg_restore` can parallelise it:

```bash
docker compose exec -T postgres pg_dump -U senddock -Fc senddock \
    > senddock-backup-$(date +%F).dump
```

Whichever format you pick, copy `.env` off-host the same day — the dump is useless for a full recovery without the secrets that signed the data:

```bash
cp .env env-backup-$(date +%F).env
```

## Schedule it

A minimal host cron job: nightly custom-format dump, keep 7 days locally.

```bash
0 3 * * * cd /opt/senddock && docker compose exec -T postgres pg_dump -U senddock -Fc senddock > /var/backups/senddock/$(date +\%F).dump && find /var/backups/senddock -mtime +7 -delete
```

Recommended cadence:

- **Daily** Postgres dump, retained 7 days locally + 30 days off-site.
- **Weekly** full backup that also captures `.env` and `docker-compose.yml`.
- **Before every upgrade** — a one-off dump labeled with the source and target version (`senddock-pre-0.8.1.dump`). This is the lowest-friction rollback path if a migration goes sideways.

### Off-site & encryption

Local backups protect against application-level corruption — they don't protect against the host going away. Sync the daily dump to object storage with server-side encryption (S3 SSE-S3 / SSE-KMS, Backblaze B2 at-rest, GCS CMEK), or use `restic` / `rclone crypt` for client-side encryption to any remote. If regulation or sensitivity demands it, encrypt before upload: `gpg -c senddock-backup-$(date +%F).dump` or `age -p` against a passphrase you store with a different provider.

## Restore into a clean instance

This is the full recovery path — a brand-new host, or the same host after a `down -v`. Restore into an **empty** database so nothing collides.

1. Bring up a fresh stack with the **same `JWT_SECRET` and `POSTGRES_PASSWORD`** as the backup (restore your `.env` first). Let it start once so migrations create the schema, or restore into the empty database directly — both work; the `--clean` flags below make the restore authoritative either way.

2. Restore the dump.

   Plain SQL:

   ```bash
   cat senddock-backup-YYYY-MM-DD.sql | \
       docker compose exec -T postgres psql -U senddock senddock
   ```

   Custom format — parallel, and `--clean --if-exists` drops existing objects first so a re-run is idempotent:

   ```bash
   docker compose cp senddock-backup-YYYY-MM-DD.dump postgres:/tmp/in.dump
   docker compose exec postgres pg_restore -U senddock -d senddock --clean --if-exists -j 4 /tmp/in.dump
   ```

3. Restart SendDock so any in-memory state (schedulers, caches) reseeds from the restored data:

   ```bash
   docker compose restart senddock
   ```

The restored database keeps every project, subscriber, template, suppression entry and log row. Because SMTP credentials are encrypted with `JWT_SECRET`, they only decrypt if you restored the matching `.env` — this is why the secret is part of the backup, not an afterthought.

## Rehearse the restore

The only thing worse than no backup is a backup you've never restored. Once a month, spin up a throwaway Postgres container, `pg_restore` the latest dump into it, and confirm the row counts match production:

```bash
docker run --rm -d --name senddock-restore-test -e POSTGRES_PASSWORD=test -e POSTGRES_USER=senddock -e POSTGRES_DB=senddock postgres:16
docker cp senddock-backup-YYYY-MM-DD.dump senddock-restore-test:/tmp/in.dump
docker exec senddock-restore-test pg_restore -U senddock -d senddock -j 4 /tmp/in.dump
docker exec senddock-restore-test psql -U senddock -d senddock -c "SELECT count(*) FROM subscribers;"
docker rm -f senddock-restore-test
```

Five minutes that turn an untested artifact into a tested recovery path.

## Before you update

Updates preserve your data — the `pgdata` volume is never touched by an image swap — but the safe habit is unchanged:

1. **Pin your image tag** so you control update timing (see [Updating](./updating#updating-a-prebuilt-image-install)).
2. **Take a labeled dump first** (`senddock-pre-<version>.dump`).
3. Update. If a migration misbehaves and rolling back the tag isn't enough because the schema already moved forward, restore that pre-update dump into a clean instance using the steps above.

See [Updating → Rolling back](./updating#rolling-back) for tag and single-migration rollbacks that don't need a full restore.
