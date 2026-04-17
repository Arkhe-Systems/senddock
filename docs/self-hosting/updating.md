# Updating

## With Docker

```bash
cd senddock
git pull origin main
docker compose -f docker-compose.prod.yml up -d --build
```

The `--build` flag rebuilds the image with the latest code. Migrations run automatically on startup.

## Without Docker

```bash
cd senddock
git pull origin main
cd frontend && npm ci && npm run build && cd ..
cd backend
make migrate
make build
```

Restart the server with the new binary.

## Checking Current Version

Check the [GitHub releases](https://github.com/arkhe-systems/senddock/releases) page for the latest version and release notes.

## Rollback

### With Docker

```bash
git checkout v0.x.x
docker compose -f docker-compose.prod.yml up -d --build
```

### Without Docker

Rollback the last migration:

```bash
goose -dir migrations postgres "$DATABASE_URL" down
```

Checkout the previous version and rebuild:

```bash
git checkout v0.x.x
cd frontend && npm ci && npm run build && cd ..
cd backend && make build
```
