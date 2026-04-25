#!/bin/sh
set -e

COMPOSE_FILE="docker-compose.prod.yml"
RESET=0
LOGS=0

usage() {
    cat <<EOF
Usage: $0 [--reset] [--logs] [--help]

  --reset   Wipe containers, volumes and .env before reinstalling.
  --logs    Tail the app logs after starting.
  --help    Show this help.

Without flags the script auto-detects the right action:
  - No .env present  -> fresh install (generates secrets, builds, starts).
  - .env present     -> update (rebuilds image with current code, restarts).
EOF
}

for arg in "$@"; do
    case "$arg" in
        --reset) RESET=1 ;;
        --logs) LOGS=1 ;;
        --help|-h) usage; exit 0 ;;
        *)
            echo "Unknown argument: $arg"
            usage
            exit 1
            ;;
    esac
done

if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running or not reachable."
    echo "Start the Docker daemon and re-run this script."
    exit 1
fi

project="$(basename "$(pwd)")"
volumes="$(docker volume ls --format '{{.Name}}' 2>/dev/null | grep -E "^${project}_" || true)"

if [ "$RESET" = "1" ]; then
    echo "Reset requested. Tearing down containers and volumes..."
    docker compose -f "$COMPOSE_FILE" down -v > /dev/null 2>&1 || true
    rm -f .env
    echo "Reset complete. Continuing with a fresh install."
    echo ""
fi

env_exists=0
[ -f .env ] && env_exists=1

if [ "$env_exists" = "1" ] && [ -z "$volumes" ]; then
    echo "Found a .env file but no Postgres volume."
    echo "This usually means a previous install was wiped at the Docker level but the .env was left behind."
    echo "Continuing will reuse the existing .env, including its Postgres password."
    echo "If that password no longer matches anything, the database will fail to start."
    echo ""
    printf "Continue with existing .env? [y/N] "
    read confirm
    case "$confirm" in
        [yY]*) ;;
        *)
            echo "Aborted. Re-run with --reset to wipe everything and start clean."
            exit 1
            ;;
    esac
fi

if [ "$env_exists" = "0" ]; then
    echo "Fresh install detected. Generating secrets..."
    JWT_SECRET=$(openssl rand -hex 32)
    POSTGRES_PASSWORD=$(openssl rand -hex 16)
    cat > .env <<EOF
JWT_SECRET=${JWT_SECRET}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
FRONTEND_URL=http://localhost:8080
PUBLIC_URL=http://localhost:8080
DEPLOYMENT_MODE=self-hosted
EOF
    echo ".env created."
    mode="install"
else
    echo "Existing install detected. Keeping current .env."
    echo "Run with --reset to wipe data and start fresh."
    mode="update"
fi

echo ""
echo "Building image (this picks up any code changes)..."
if ! docker compose -f "$COMPOSE_FILE" build --pull; then
    echo "Build failed. See output above for details."
    exit 1
fi

echo ""
echo "Starting services..."
if ! docker compose -f "$COMPOSE_FILE" up -d; then
    echo "docker compose up failed. Check 'docker compose logs' for details."
    exit 1
fi

echo ""
printf "Waiting for SendDock to be ready"
ready=0
i=0
while [ $i -lt 30 ]; do
    sleep 2
    printf "."
    if curl -fsS --max-time 2 http://localhost:8080/health > /dev/null 2>&1; then
        ready=1
        break
    fi
    i=$((i + 1))
done
echo ""

if [ "$ready" = "0" ]; then
    echo ""
    echo "SendDock did not become ready within 60 seconds."
    echo "Check what the app container is doing:"
    echo "  docker compose -f $COMPOSE_FILE ps"
    echo "  docker compose -f $COMPOSE_FILE logs app --tail 100"
    echo ""
    echo "Common causes are documented at https://docs.senddock.dev/self-hosting/troubleshooting"
    exit 1
fi

echo ""
if [ "$mode" = "install" ]; then
    echo "SendDock is running at http://localhost:8080"
    echo "Open it in your browser to create your admin account."
else
    echo "SendDock updated and running at http://localhost:8080"
fi

if [ "$LOGS" = "1" ]; then
    echo ""
    echo "Following app logs (Ctrl+C to stop)..."
    docker compose -f "$COMPOSE_FILE" logs -f app
fi
