# SendDock

> **Warning:** SendDock is in early development. Not recommended for production use yet.

Open-source email marketing platform built with Go and Vue. Self-hostable, API-first, and designed for developers and businesses that need reliable email delivery at scale.

Part of [Arkhe Systems](https://arkhe.systems).

## Tech Stack

### Backend
- **Go** (stdlib `net/http`) — HTTP server, no framework
- **PostgreSQL** — primary database
- **Redis** — job queues and caching
- **sqlc** — type-safe SQL query generation
- **goose** — database migrations
- **bcrypt + JWT** — authentication

### Frontend
- **Vue 3** — reactive UI framework
- **TypeScript** — type safety
- **Pinia** — state management
- **Vue Router** — SPA routing with auth guards
- **Tailwind CSS 4** — utility-first styling
- **Vite** — build tool and dev server

## Project Structure

```
senddock/
├── backend/
│   ├── cmd/
│   │   ├── server/          # HTTP server entrypoint
│   │   └── worker/          # Background workers entrypoint
│   ├── internal/
│   │   ├── config/          # Environment variable loading
│   │   ├── db/              # sqlc generated code (DO NOT EDIT)
│   │   ├── handler/         # HTTP handlers (request/response)
│   │   ├── middleware/       # Auth, CORS, rate limiting
│   │   ├── service/         # Business logic
│   │   └── worker/          # Async job definitions
│   ├── migrations/          # SQL migration files (goose)
│   ├── sqlc/queries/        # SQL queries for sqlc
│   ├── pkg/sdk/             # Public Go SDK
│   ├── Makefile             # Dev commands (make dev, make run, etc.)
│   ├── sqlc.yaml
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/             # API client (fetch wrapper)
│   │   ├── components/ui/   # Reusable UI components
│   │   ├── layouts/         # Page layouts
│   │   ├── router/          # Vue Router with auth guards
│   │   ├── stores/          # Pinia stores (auth, projects)
│   │   └── views/           # Page components
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml       # PostgreSQL + Redis for development
├── docs/                    # Documentation
├── LICENSE                  # AGPL-3.0
└── ROADMAP.md               # Development roadmap
```

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker and Docker Compose
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- [sqlc](https://github.com/sqlc-dev/sqlc) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Getting Started

### 1. Clone and setup

```bash
git clone https://github.com/arkhe-systems/senddock.git
cd senddock
```

### 2. Backend

```bash
cd backend
cp .env.example .env          # Configure your environment
make dev                      # Starts DB + runs migrations + starts server
```

The backend will be available at `http://localhost:8080`.

#### Available make commands

| Command | Description |
|---------|-------------|
| `make dev` | Start everything (DB + migrations + server) |
| `make run` | Start the server only |
| `make db-up` | Start PostgreSQL and Redis |
| `make db-down` | Stop PostgreSQL and Redis |
| `make db-restart` | Restart Docker + containers (fixes port conflicts) |
| `make migrate` | Run database migrations |
| `make build` | Build production binary |

### 3. Frontend

```bash
cd frontend
npm install
npm run dev                   # Starts dev server at http://localhost:5173
```

### 4. Test it

```bash
# Health check
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"yourpassword","name":"Your Name"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"you@example.com","password":"yourpassword"}'
```

Or simply open `http://localhost:5173` and register through the UI.

## API Endpoints

### Authentication
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register a new user |
| POST | `/api/v1/auth/login` | Login |
| POST | `/api/v1/auth/refresh` | Refresh access token |
| POST | `/api/v1/auth/logout` | Logout and invalidate tokens |
| GET | `/api/v1/me` | Get current user (requires auth) |

### Projects
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/projects` | Create a project |
| GET | `/api/v1/projects` | List user's projects |
| GET | `/api/v1/projects/{id}` | Get a specific project |
| DELETE | `/api/v1/projects/{id}` | Delete a project |

All project endpoints require authentication. Tokens are managed via HttpOnly cookies (set automatically on login/register).

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | — |
| `REDIS_URL` | Redis connection string | — |
| `JWT_SECRET` | Secret key for JWT signing | — |
| `FRONTEND_URL` | Frontend URL for CORS | `http://localhost:5173` |

## Open Core Model

SendDock follows an open-core model:

- **Community Edition** (this repo) — free, open-source, full core functionality
- **Pro Edition** (private repo) — advanced features for teams and enterprises

### Community (free, self-hosted)

- Unlimited email sending
- Unlimited subscribers
- Project management
- Subscriber management
- Email templates
- REST API
- Basic analytics (sent/failed)
- Single SMTP per project

### Pro (paid license)

- Multi-SMTP with automatic failover
- Advanced analytics (opens, clicks, bounces)
- Webhooks with retry logic
- Team members with roles
- White-label (remove SendDock branding)
- SSO / LDAP
- Priority support

### Cloud (senddock.dev)

Managed hosting with usage-based pricing. No setup required.

| Plan | Emails/month | Subscribers | Price |
|------|-------------|-------------|-------|
| Free | 2,000 | 500 | $0 |
| Starter | 10,000 | 2,500 | $15/mo |
| Pro | 50,000 | 10,000 | $39/mo |
| Business | 200,000 | 50,000 | $79/mo |
| Enterprise | 500,000 | Unlimited | $199/mo |

## License

AGPL-3.0 — See [LICENSE](LICENSE) file.

The community edition is free and open-source under AGPL-3.0. If you modify SendDock and offer it as a service (SaaS), you must open-source your modifications. For commercial licensing (Pro/Enterprise), contact hello@arkhe.systems.
