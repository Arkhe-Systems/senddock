#!/bin/sh
set -e

echo "[senddock] Running migrations..."
attempt=1
until goose -dir ./migrations postgres "$DATABASE_URL" up; do
    if [ "$attempt" -ge 5 ]; then
        echo "[senddock] Migrations failed after $attempt attempts, giving up"
        exit 1
    fi
    echo "[senddock] Migration attempt $attempt failed (another replica may be migrating), retrying in 3s..."
    attempt=$((attempt + 1))
    sleep 3
done

echo "[senddock] Starting server..."
exec ./senddock
