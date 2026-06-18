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
#   INSTALL_DIR                install directory (default: /opt/senddock)
#   SENDDOCK_PUBLIC_URL        public URL (skips prompt when set; wins over
#                              existing .env value, so this is how you fix a
#                              .env that was generated with the wrong URL —
#                              just re-run with SENDDOCK_PUBLIC_URL=... set)
#   SENDDOCK_PORT              host port to expose (default: 8080)
#   SENDDOCK_ALLOW_LOCALHOST   set to "yes" to allow PUBLIC_URL=localhost
#                              (non-interactive escape hatch for local tests —
#                              newsletters / unsubscribe won't work in this mode)

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

# Resolve PUBLIC_URL in this order of preference:
#   1. SENDDOCK_PUBLIC_URL env var (highest — always wins, lets the user fix
#      a broken .env by re-running setup.sh with the right env)
#   2. PUBLIC_URL line already in .env (if it's a valid public URL — keeps
#      re-runs idempotent for installs that are already correct)
#   3. Interactive prompt with retries until non-empty
#   4. Hard error if non-interactive and nothing supplied

PUBLIC_URL_INPUT="${SENDDOCK_PUBLIC_URL:-}"

if [[ -z "$PUBLIC_URL_INPUT" && -f .env ]]; then
  existing_url="$(grep -E '^PUBLIC_URL=' .env | head -1 | cut -d= -f2- || true)"
  if [[ -n "$existing_url" && "$existing_url" != *localhost* && "$existing_url" != *127.0.0.1* ]]; then
    PUBLIC_URL_INPUT="$existing_url"
    ok "Reusing PUBLIC_URL from existing .env: ${PUBLIC_URL_INPUT}"
  fi
fi

if [[ -z "$PUBLIC_URL_INPUT" ]]; then
  if [[ -t 0 ]]; then
    while true; do
      echo ""
      echo "${b}Public URL${n} — the address users (and your emails' tracking links) will reach this instance at."
      echo "  Examples: ${b}https://mail.tudominio.com${n}, ${b}https://email.example.com${n}"
      echo "  ${y}This must be a public domain.${n} Newsletters and unsubscribe links won't work with localhost."
      read -rp "  > " PUBLIC_URL_INPUT
      if [[ -n "$PUBLIC_URL_INPUT" ]]; then
        break
      fi
      warn "Public URL is required. Try again, or Ctrl+C to abort."
    done
  else
    die "Non-interactive mode requires SENDDOCK_PUBLIC_URL=https://your-domain.com. See the script header for examples."
  fi
fi

# Auto-prefix https:// if no scheme — accepting bare domains is way friendlier
# than silently writing them and letting SendDock parse them weirdly later.
if [[ "$PUBLIC_URL_INPUT" != http://* && "$PUBLIC_URL_INPUT" != https://* ]]; then
  PUBLIC_URL_INPUT="https://${PUBLIC_URL_INPUT}"
  ok "No URL scheme provided — using ${PUBLIC_URL_INPUT}"
fi

# Refuse localhost unless explicitly opted in. This was the silent footgun:
# previously an empty prompt → default to localhost → install completes →
# user opens dashboard and discovers "Newsletters are disabled" without
# knowing why.
if [[ "$PUBLIC_URL_INPUT" == *localhost* || "$PUBLIC_URL_INPUT" == *127.0.0.1* ]]; then
  warn "PUBLIC_URL points to localhost. Newsletters, unsubscribe links and tracking pixels in outgoing emails won't work."
  if [[ "${SENDDOCK_ALLOW_LOCALHOST:-}" != "yes" ]]; then
    if [[ -t 0 ]]; then
      read -rp "  Continue anyway for a local test? Type ${b}yes${n} to confirm: " confirm
      [[ "$confirm" == "yes" ]] || die "Aborted. Re-run with a public URL like https://mail.tudominio.com, or pass SENDDOCK_ALLOW_LOCALHOST=yes for an intentional local test."
    else
      die "Refusing to set localhost as PUBLIC_URL in non-interactive mode. Pass SENDDOCK_ALLOW_LOCALHOST=yes to allow it."
    fi
  fi
fi

PUBLIC_URL_FINAL="$PUBLIC_URL_INPUT"
ok "PUBLIC_URL = ${PUBLIC_URL_FINAL}"

# --- write/update .env -------------------------------------------------------

if [[ -f .env ]]; then
  # .env exists — preserve secrets, only patch PUBLIC_URL if it differs.
  existing_url="$(grep -E '^PUBLIC_URL=' .env | head -1 | cut -d= -f2- || true)"
  if [[ "$existing_url" == "$PUBLIC_URL_FINAL" ]]; then
    ok ".env already has the correct PUBLIC_URL — leaving secrets intact."
  else
    say "Updating PUBLIC_URL in existing .env (was: ${existing_url:-<unset>})…"
    if grep -qE '^PUBLIC_URL=' .env; then
      sed -i "s|^PUBLIC_URL=.*|PUBLIC_URL=${PUBLIC_URL_FINAL}|" .env
    else
      echo "PUBLIC_URL=${PUBLIC_URL_FINAL}" >> .env
    fi
    ok "PUBLIC_URL updated. Other secrets preserved."
  fi
else
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
    # Maintained fork of containrrr/watchtower. The upstream was abandoned in
    # 2023; its :latest image ships a Docker client too old for modern daemons
    # (API < 1.40) and crashes in a restart loop on Ubuntu 24.04+/Docker 26+.
    image: nickfedor/watchtower:latest
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

# --- post-up health check ---------------------------------------------------

say "Waiting up to 60s for all services to stabilize…"
all_good=0
for i in 1 2 3 4 5 6 7 8 9 10 11 12; do
  sleep 5
  running=$(docker compose ps --status running --quiet 2>/dev/null | wc -l | tr -d ' ')
  total=$(docker compose ps --quiet 2>/dev/null | wc -l | tr -d ' ')
  if [[ "$total" -gt 0 && "$running" == "$total" ]]; then
    all_good=1
    break
  fi
done

if [[ $all_good -eq 1 ]]; then
  ok "All $total services running."
else
  warn "Some services are not in 'running' state after 60s."
  warn "Inspect with:"
  warn "    cd ${INSTALL_DIR} && sudo docker compose ps"
  warn "    cd ${INSTALL_DIR} && sudo docker compose logs --tail=50"
fi

echo ""
docker compose ps

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

     If "Test Connection" times out, your ISP (residential connections especially)
     may block outbound SMTP ports. See the diagnostic + workarounds at:
       https://docs.senddock.dev/guide/smtp#diagnosing-port-issues

${b}Updates${n}

  When a new SendDock release lands, an "Update available" badge appears in
  the dashboard nav. Click it → "Update now" → Watchtower pulls the new
  image and recreates the container. Postgres + Redis volumes are preserved
  (your data stays). Takes ~30–60s.

  Note: Watchtower mounts /var/run/docker.sock and is therefore root-equivalent
  on the host. It's label-scoped to only ever touch the SendDock container —
  not your other services. To opt out, see docker-compose.override.yml.

EOF
