# Configuration

## Environment Variables

See [Environment Variables](/guide/environment) for the full list.

## Database

SendDock uses PostgreSQL 17. The Docker Compose file included in the repo sets up a PostgreSQL instance with:

- User: `senddock`
- Password: `senddock_dev` (change in production)
- Database: `senddock`
- Port: `5434` (to avoid conflicts)

### Using an external database

Set `DATABASE_URL` to your PostgreSQL connection string:

```
DATABASE_URL=postgres://user:password@host:5432/dbname?sslmode=require
```

Then run migrations:

```bash
make migrate
```

## Redis

Redis is used for job queues and caching (planned features). Port `6380` by default.

## Security Checklist

Before exposing to the internet:

- [ ] Change `JWT_SECRET` to a random string (min 32 characters): `openssl rand -base64 48`
- [ ] Change PostgreSQL password from default
- [ ] Set `FRONTEND_URL` to your actual domain (used for CORS)
- [ ] Set `PUBLIC_URL` to the same domain (used for unsubscribe + tracking links inside emails)
- [ ] Use HTTPS via reverse proxy — `Secure` cookies are set automatically when `FRONTEND_URL` starts with `https://`
- [ ] Keep SendDock updated

If anything misbehaves after going live, see [Troubleshooting](/self-hosting/troubleshooting).

## Ports

| Service | Default Port |
|---------|-------------|
| SendDock API | 8080 |
| Frontend (dev) | 5173 |
| PostgreSQL | 5434 |
| Redis | 6380 |
