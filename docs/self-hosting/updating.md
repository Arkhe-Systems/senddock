# Updating

## From Git

```bash
cd senddock
git pull origin main
```

### Backend

```bash
cd backend
make migrate    # Apply new migrations
make build      # Rebuild binary
```

Restart the server.

### Frontend

```bash
cd frontend
npm install     # Install new dependencies
npm run build   # Rebuild
```

## Migrations

New versions may include database migrations. Always run `make migrate` after pulling updates. Migrations are incremental and safe to run multiple times (goose tracks which ones have been applied).

To check migration status:

```bash
goose -dir migrations postgres "$DATABASE_URL" status
```

## Breaking Changes

Check the [GitHub releases](https://github.com/arkhe-systems/senddock/releases) page for release notes. Breaking changes will be documented with migration instructions.

## Rollback

If something goes wrong, rollback the last migration:

```bash
goose -dir migrations postgres "$DATABASE_URL" down
```

And checkout the previous version:

```bash
git checkout v0.x.x
make build
```
