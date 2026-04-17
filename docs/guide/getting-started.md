# Getting Started

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker and Docker Compose
- [goose](https://github.com/pressly/goose) — `go install github.com/pressly/goose/v3/cmd/goose@latest`
- [sqlc](https://github.com/sqlc-dev/sqlc) — `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`

## Installation

```bash
git clone https://github.com/arkhe-systems/senddock.git
cd senddock
```

## Backend

```bash
cd backend
cp .env.example .env
make dev
```

This starts PostgreSQL + Redis via Docker, runs migrations, and starts the API server at `http://localhost:8080`.

## Frontend

In a separate terminal:

```bash
cd frontend
npm install
npm run dev
```

The dashboard runs at `http://localhost:5173`.

## First-time Setup

Open `http://localhost:5173` in your browser. Since no users exist yet, you'll see the setup screen. Create your admin account and you're ready to go.

## Next Steps

1. [Create a project](/guide/projects)
2. [Configure SMTP](/guide/smtp)
3. [Add subscribers](/guide/subscribers)
4. [Build a template](/guide/templates)
5. [Send your first email](/guide/sending)

## Make Commands

| Command | Description |
|---------|-------------|
| `make dev` | Start DB + migrations + server |
| `make run` | Start server only |
| `make test` | Run unit tests |
| `make sqlc` | Regenerate sqlc code |
| `make migrate` | Run database migrations |
| `make build` | Build production binary |
| `make db-up` | Start PostgreSQL and Redis |
| `make db-down` | Stop PostgreSQL and Redis |
