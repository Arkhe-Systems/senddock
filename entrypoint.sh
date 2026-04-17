#!/bin/sh
set -e

echo "[senddock] Running migrations..."
goose -dir ./migrations postgres "$DATABASE_URL" up

echo "[senddock] Starting server..."
exec ./senddock
