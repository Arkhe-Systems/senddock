#!/usr/bin/env bash
# setup.sh — SendDock self-host installer for Ubuntu.
#
# Installs Docker (if missing), downloads the official compose file,
# generates a .env with random secrets, and starts the stack
# (senddock + postgres + redis).
#
# Quick install (interactive):
#   curl -fsSL https://raw.githubusercontent.com/Arkhe-Systems/senddock/main/scripts/setup.sh -o setup.sh
#   sudo bash setup.sh
#
# Non-interactive (CI / one-shot):
#   curl -fsSL https://raw.githubusercontent.com/Arkhe-Systems/senddock/main/scripts/setup.sh \
#     | sudo SENDDOCK_PUBLIC_URL=https://email.example.com bash
#
# Environment overrides:
#   INSTALL_DIR            install directory (default: /opt/senddock)
#   SENDDOCK_PUBLIC_URL    public URL (skips prompt when set)
#   SENDDOCK_PORT          host port to expose (default: 8080)

set -euo pipefail

INSTALL_DIR="${INSTALL_DIR:-/opt/senddock}"
SENDDOCK_PORT="${SENDDOCK_PORT:-8080}"
COMPOSE_URL="https://raw.githubusercontent.com/Arkhe-Systems/senddock/main/docker-compose.image.yml"

if [[ -t 1 ]]; then
  b=$(printf '\033[1m'); g=$(printf '\033[32m')
  y=$(printf '\033[33m'); r=$(printf '\033[31m'); n=$(printf '\033[0m')
else
  b=""; g=""; y=""; r=""; n=""
fi

say()  { echo "${b}==>${n} $*"; }
ok()   { echo "${g}✓${n} $*"; }
warn() { echo "${y}!${n} $*" >&2; }
die()  { echo "${r}✗ $*${n}" >&2; exit 1; }

# --- preflight ---------------------------------------------------------------

[[ "$(uname -s)" == "Linux" ]] || die "This installer only supports Linux."

[[ -r /etc/os-release ]] || die "Can't read /etc/os-release — unsupported distribution."
# shellcheck disable=SC1091
. /etc/os-release
[[ "${ID:-}" == "ubuntu" ]] || die "This installer currently only supports Ubuntu (detected: ${PRETTY_NAME:-unknown}). For other distros, see https://docs.senddock.dev/self-hosting/installation"

[[ $EUID -eq 0 ]] || die "Run with sudo: sudo bash setup.sh"

for cmd in curl openssl; do
  command -v "$cmd" >/dev/null 2>&1 || {
    say "Installing missing dependency: $cmd"
    apt-get update -qq && apt-get install -y -qq "$cmd"
  }
done

# --- Docker ------------------------------------------------------------------

install_docker() {
  say "Installing Docker (official repository)…"
  apt-get update -qq
  apt-get install -y -qq ca-certificates curl gnupg
  install -m 0755 -d /etc/apt/keyrings
  curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  chmod a+r /etc/apt/keyrings/docker.asc
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" \
    > /etc/apt/sources.list.d/docker.list
  apt-get update -qq
  apt-get install -y -qq docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
  systemctl enable --now docker
  ok "Docker installed."
}

docker_ready() {
  command -v docker >/dev/null 2>&1 \
    && docker compose version >/dev/null 2>&1 \
    && docker info >/dev/null 2>&1
}

if docker_ready; then
  ok "Docker already installed and running ($(docker --version))."
elif command -v docker >/dev/null 2>&1; then
  warn "Docker CLI present but daemon not reachable — trying to start it…"
  systemctl enable --now docker 2>/dev/null || true
  sleep 3
  if docker_ready; then
    ok "Docker daemon started ($(docker --version))."
  else
    warn "Daemon still not reachable — reinstalling Docker…"
    install_docker
  fi
else
  install_docker
fi

# --- install dir + compose ---------------------------------------------------

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"
say "Install dir: $INSTALL_DIR"

if [[ -f docker-compose.yml ]]; then
  warn "docker-compose.yml already exists — keeping the existing file."
else
  say "Downloading docker-compose.yml…"
  curl -fsSL "$COMPOSE_URL" -o docker-compose.yml
  ok "docker-compose.yml ready."
fi

# --- .env --------------------------------------------------------------------

PUBLIC_URL_FINAL=""

if [[ -f .env ]]; then
  warn ".env already exists — not overwriting. Edit it manually if needed, then re-run:"
  warn "    cd $INSTALL_DIR && docker compose up -d"
  PUBLIC_URL_FINAL="$(grep -E '^PUBLIC_URL=' .env | cut -d= -f2- || true)"
else
  PUBLIC_URL_INPUT="${SENDDOCK_PUBLIC_URL:-}"
  if [[ -z "$PUBLIC_URL_INPUT" && -t 0 ]]; then
    echo ""
    echo "${b}Public URL${n} — the address users (and your emails' tracking links) will reach this instance at."
    echo "  Example: https://email.example.com"
    echo "  Leave empty for a local-only test (http://localhost:${SENDDOCK_PORT})."
    read -rp "  > " PUBLIC_URL_INPUT
  fi
  if [[ -z "$PUBLIC_URL_INPUT" ]]; then
    PUBLIC_URL_INPUT="http://localhost:${SENDDOCK_PORT}"
    warn "No public URL provided — defaulting to ${PUBLIC_URL_INPUT}. Edit PUBLIC_URL in $INSTALL_DIR/.env before sending real emails."
  fi
  PUBLIC_URL_FINAL="$PUBLIC_URL_INPUT"

  say "Generating .env with random secrets…"
  JWT_SECRET="$(openssl rand -hex 32)"
  POSTGRES_PASSWORD="$(openssl rand -hex 24)"
  SENDDOCK_WATCHTOWER_TOKEN="$(openssl rand -hex 32)"

  umask 077
  cat > .env <<EOF
# Generated by setup.sh on $(date -u +"%Y-%m-%dT%H:%M:%SZ")

# === REQUIRED ===

# Public URL where this instance is reachable from the internet.
# Used to build unsubscribe + tracking links inside outgoing emails.
PUBLIC_URL=${PUBLIC_URL_INPUT}

# Postgres password (regenerate with: openssl rand -hex 24).
POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

# JWT signing secret (regenerate with: openssl rand -hex 32).
JWT_SECRET=${JWT_SECRET}

# === ONE-CLICK UPDATES (Watchtower) ===
# Wired up by setup.sh so the dashboard's "Update now" button works.
# Watchtower runs in HTTP-API-only mode (no auto-polling); updates only
# fire when you explicitly click the button.
SENDDOCK_WATCHTOWER_URL=http://watchtower:8080
SENDDOCK_WATCHTOWER_TOKEN=${SENDDOCK_WATCHTOWER_TOKEN}

# === OPTIONAL ===

# Host port to expose. The container always listens on 8080 internally.
SENDDOCK_PORT=${SENDDOCK_PORT}

# Pro / Team license key from senddock.dev. Empty = Community tier (free).
SENDDOCK_LICENSE_KEY=

# Per-IP request cap, rolling 60s window. Default 600. Only enforced with Redis.
# RATE_LIMIT_PER_MINUTE=600
EOF
  chmod 600 .env
  ok ".env written (mode 600)."
fi

# --- docker-compose.override.yml (Watchtower) -------------------------------

if [[ -f docker-compose.override.yml ]]; then
  warn "docker-compose.override.yml already exists — keeping the existing file."
else
  say "Writing docker-compose.override.yml (Watchtower for one-click updates)…"
  cat > docker-compose.override.yml <<'EOF'
# Generated by setup.sh — adds Watchtower so the dashboard's "Update now"
# button can pull the new image and recreate the SendDock container.
#
# Watchtower runs in HTTP-API-only mode (no auto-polling). Updates only
# fire when you click the button in the dashboard.
#
# Scope is limited to containers with the watchtower.enable label, so
# Watchtower only ever touches the SendDock container — not your other
# services on the same Docker host.
#
# To opt out: delete this file, remove the two SENDDOCK_WATCHTOWER_* lines
# from .env, then run `docker compose up -d`. Updates will revert to the
# manual `docker compose pull && up -d` flow.

services:
  watchtower:
    image: containrrr/watchtower
    restart: unless-stopped
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      WATCHTOWER_HTTP_API_UPDATE: "true"
      WATCHTOWER_HTTP_API_TOKEN: ${SENDDOCK_WATCHTOWER_TOKEN:?SENDDOCK_WATCHTOWER_TOKEN is required in .env}
      WATCHTOWER_HTTP_API_PERIODIC_POLLS: "false"
      WATCHTOWER_CLEANUP: "true"
      WATCHTOWER_LABEL_ENABLE: "true"

  senddock:
    labels:
      com.centurylinklabs.watchtower.enable: "true"
    environment:
      SENDDOCK_WATCHTOWER_URL: ${SENDDOCK_WATCHTOWER_URL:-http://watchtower:8080}
      SENDDOCK_WATCHTOWER_TOKEN: ${SENDDOCK_WATCHTOWER_TOKEN:?SENDDOCK_WATCHTOWER_TOKEN is required in .env}
EOF
  ok "docker-compose.override.yml ready."
fi

# --- bring up ----------------------------------------------------------------

say "Pulling images…"
docker compose pull

say "Starting the stack…"
docker compose up -d

# --- final ------------------------------------------------------------------

cat <<EOF

${g}${b}SendDock is starting.${n}

  Install dir: ${INSTALL_DIR}
  Compose:     ${INSTALL_DIR}/docker-compose.yml + docker-compose.override.yml
  Env:         ${INSTALL_DIR}/.env  (mode 600 — contains secrets)

${b}Next steps${n}

  1. Wait ~10s for the healthcheck:
       cd ${INSTALL_DIR} && docker compose ps

  2. Open SendDock and create your admin account on the Setup screen:
       ${PUBLIC_URL_FINAL:-http://<your-host>:${SENDDOCK_PORT}}

  3. For production, put a reverse proxy with HTTPS in front (Caddy / Nginx / Traefik).
     Guide: https://docs.senddock.dev/self-hosting/installation#reverse-proxy-https

  4. Configure your SMTP relay and license key from the dashboard, or by editing
     ${INSTALL_DIR}/.env and running:
       cd ${INSTALL_DIR} && docker compose up -d

${b}Updates${n}

  When a new SendDock release lands, an "Update available" badge appears in
  the dashboard nav. Click it → "Update now" → Watchtower pulls the new
  image and recreates the container. Postgres + Redis volumes are preserved
  (your data stays). Takes ~30–60s.

  Note: Watchtower mounts /var/run/docker.sock and is therefore root-equivalent
  on the host. It's label-scoped to only ever touch the SendDock container —
  not your other services. To opt out, see docker-compose.override.yml.

EOF
