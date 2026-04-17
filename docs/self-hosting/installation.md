# Installation

## Requirements

- A Linux server (Ubuntu, Debian, etc.)
- Docker and Docker Compose
- Go 1.22+ (for building from source)
- Node.js 20+ (for building the frontend)

## Quick Start with Docker (recommended)

```bash
git clone https://github.com/arkhe-systems/senddock.git
cd senddock/backend
cp .env.example .env
```

Edit `.env` with your settings:

```bash
DATABASE_URL=postgres://senddock:your_secure_password@localhost:5434/senddock?sslmode=disable
JWT_SECRET=generate-a-random-64-char-string-here
PORT=8080
REDIS_URL=redis://localhost:6380
FRONTEND_URL=https://your-domain.com
DEPLOYMENT_MODE=self-hosted
```

Start the services:

```bash
make dev
```

## Building from Source

### Backend

```bash
cd backend
make build
```

This creates a binary at `bin/senddock`. Run it with:

```bash
./bin/senddock
```

### Frontend

```bash
cd frontend
npm install
npm run build
```

The built files will be in `frontend/dist/`. In production, Go will serve these static files (planned).

## First-time Setup

1. Open your instance in a browser
2. The setup screen will appear (since no users exist)
3. Create your admin account
4. Start using SendDock

## Reverse Proxy

For production, put SendDock behind a reverse proxy (Nginx, Caddy, Traefik) with HTTPS.

### Caddy example

```
your-domain.com {
    reverse_proxy localhost:8080
}
```

### Nginx example

```nginx
server {
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Remember to update `FRONTEND_URL` in `.env` to match your domain.
