# Environment Variables

All configuration is done via environment variables. Copy `.env.example` to `.env` in the `backend/` directory.

## Required

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_SECRET` | Secret key for JWT signing (use a random string, min 32 chars) |

## Optional

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | HTTP server port | `8080` |
| `REDIS_URL` | Redis connection string | — |
| `FRONTEND_URL` | Frontend URL for CORS headers | `http://localhost:5173` |
| `DEPLOYMENT_MODE` | `self-hosted` or `cloud` | `self-hosted` |

## Deployment Modes

### self-hosted (default)

- Public registration is disabled
- First user created via setup screen becomes admin
- Single user only (team members require Pro)

### cloud

- Public registration enabled at `/api/v1/auth/register`
- Multiple users, each with their own account
- Plan-based limits (for senddock.dev managed hosting)

## Waitlist (optional)

| Variable | Description |
|----------|-------------|
| `WAITLIST_PROJECT_ID` | Project ID for the public waitlist endpoint |
| `WAITLIST_TEMPLATE_ID` | Template ID for the waitlist confirmation email |

When both are set, `POST /api/v1/waitlist` becomes available as a public endpoint. It creates a subscriber with `pending` status and sends the confirmation template. Used for landing page waitlist forms without exposing API keys.

## Example .env

```bash
DATABASE_URL=postgres://senddock:senddock_dev@localhost:5434/senddock?sslmode=disable
JWT_SECRET=change-this-to-a-random-secret
PORT=8080
REDIS_URL=redis://localhost:6380
FRONTEND_URL=http://localhost:5173
DEPLOYMENT_MODE=self-hosted
WAITLIST_PROJECT_ID=
WAITLIST_TEMPLATE_ID=
```
