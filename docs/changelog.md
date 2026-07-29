# Changelog

All notable changes to SendDock are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the project follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

For the canonical source, see [`CHANGELOG.md` in the repo](https://github.com/arkhe-systems/senddock/blob/main/CHANGELOG.md). Each release also has matching notes on [GitHub Releases](https://github.com/arkhe-systems/senddock/releases) and the `ghcr.io/arkhe-systems/senddock` image carries the same tag.

## How versions work

| Bump | When | Example |
|---|---|---|
| `0.x.0` (minor) | New features | Webhooks, Bounce ingestion |
| `0.x.y` (patch) | Bug fixes, doc updates | License gate hotfix |
| `1.0.0` (major) | First stable | When the API surface is frozen |

Pre-1.0 minor releases may contain breaking changes — check the version's notes before upgrading. Database migrations are forward-only and applied automatically by `goose` on container startup; rollback steps are documented in [Updating](./self-hosting/updating#rolling-back).

## [Unreleased]

_Nothing here yet. Track upcoming work on the [open issues](https://github.com/arkhe-systems/senddock/issues)._

## [0.8.0] — 2026-07-28

The v0.8 line: analytics graduates to free Core, self-host configuration moves into the dashboard, and the paid tier is redrawn around deliverability and reporting.

### Added

- **Analytics, rebuilt and free (Core).** The old single-endpoint funnel became a tabbed, charted dashboard — **Overview, Campaigns, Audience, Engagement** — with a date-range picker, per-campaign breakdowns (every email log is now tagged with its broadcast), a subscriber-growth view and a weekday×hour click **heatmap**. Bounces are recorded on the log, not just the suppression list.
- **Instance settings in the dashboard (self-host).** The public URL, a configurable **session inactivity timeout** (5–1440 min) and the **Pro license key** are now set from **Instance** in the dashboard, stored in the database, applied live with no restart. See [Instance settings](./guide/instance-settings).
- **Deliverability (Pro).** A new Analytics tab: domain health (SPF/DKIM/DMARC), a per-provider breakdown with acceptance/bounce/open/click rates, and a **spam-complaint rate** ingested from an FBL/complaint webhook.
- **Report builder (Pro).** A dynamic builder over subscribers and email events — dimensions/measures, 1–2 dimension pivots, segment filters, saved reports and CSV export, backed by an allowlist-only query engine.
- **Rich-text send.** The Send Email composer gained a **Text/Rich** toggle per template variable, backed by a WYSIWYG editor; rich values are sanitized server-side. On the API, mark fields via `html_fields`.
- **Delete a workspace from the UI.** Owner-only, from the workspace manage screen; refused while the workspace still owns projects.

### Changed

- **Analytics moved Pro → Core.** The descriptive analytics suite is now free and ungated. The paid tier is now **Deliverability + Reports + Audit log + Team**.
- **`DEPLOYMENT_MODE` → `CLOUD`.** The hosted flag is now a boolean `CLOUD`; the old string is still read until v0.9.

### Deprecated

- **`PUBLIC_URL` and `SENDDOCK_LICENSE_KEY` environment variables (self-host).** Both now live in the dashboard; a value left in the environment is imported once on boot and support for reading it is **removed in v0.9**.

## [0.7.0] — 2026-07-05

The "Marketing-ready" milestone — SendDock stops being send-to-everyone and becomes a real targeting tool — plus webhooks graduating to the free Core.

### Added

- **Custom fields for subscribers ([#72](https://github.com/arkhe-systems/senddock/issues/72), Core).** Typed, project-scoped attributes on top of the existing `subscribers.metadata` column: `string`, `number`, `date`, `boolean`, `enum`. Define them under **Settings → Custom Fields**; values are validated on write (unknown keys rejected), render in templates as <span v-pre>`{{custom.KEY}}`</span>, show as columns and per-type inputs in the subscribers table, and map from extra CSV columns on import. New endpoints under `/api/v1/projects/{id}/fields`.
- **Tags + segments ([#40](https://github.com/arkhe-systems/senddock/issues/40), Core).** Subscribers carry free-form tags (single, bulk, and inline-on-import). Segments are saved filters — a match-all/any predicate over `status`, `tags` and `custom.*` fields — with a live match count while you build. Broadcasts accept a `segment_id` to send to a subset instead of all active subscribers. New endpoints under `/api/v1/projects/{id}/segments` and `/tags`.
- **Segment filter on Pro Analytics.** The analytics overview accepts an optional `segment_id` to scope every metric to a segment's members.

### Changed

- **Webhooks are now free (Core).** The management UI and REST API (create/list/pause/delete, delivery history) moved out of the Pro tier — webhooks are developer table stakes. The dispatcher, HMAC signing and retries were already in Core; now nothing about webhooks requires a `SENDDOCK_LICENSE_KEY`. The paid tier is now Analytics + Audit log + Team.

### Security

- **Authentication hardening.** Per-account throttling on login and two-factor verification, all sessions revoked on password change, API keys constrained to their intended capabilities, and a safer default workspace role.

## [0.6.8] — 2026-06-22

Two new features and a chunky bag of UI fixes. Headline is the **community starter template library**: every SendDock instance — Cloud and self-hosted — now ships a "★ Browse library" modal on the Templates page that pulls templates from a separate community-maintained repo, clones one into your project on click, and opens it in the editor. Second headline is **Arch Linux support in the one-line installer** — `curl senddock.dev/install.sh | sudo bash` now works on Arch and its derivatives (Manjaro, EndeavourOS, CachyOS, Garuda) the same way it works on Ubuntu. Plus a fix for the long-standing "buttons stop responding after login" bug that turned out to be password-manager autofill extensions crashing on Vue Teleport. Drop-in upgrade — no migrations.

### Added

- **Community template library** ([#73](https://github.com/Arkhe-Systems/senddock/issues/73)). New "★ Browse library" button on the Templates page opens a modal with a categorized grid of starter templates (welcome flows, monthly digests, single-story newsletters, product launches, changelogs, weekly link roundups, password resets, email verifications, transactional receipts). Click a template, get a clone in your project, ready to edit. Library content lives in the public [`Arkhe-Systems/senddock-templates`](https://github.com/Arkhe-Systems/senddock-templates) repo — community PRs welcome, MIT-licensed. New backend service caches the manifest in Redis for 1 hour; per-template HTML is fetched on demand. Override the source with `TEMPLATE_LIBRARY_URL` to point at a private fork or curated internal gallery.
- **Arch Linux family support in the one-line installer** ([#81](https://github.com/Arkhe-Systems/senddock/issues/81)). `scripts/setup.sh` now branches on `/etc/os-release` and runs the pacman-based Docker install path on Arch, Manjaro, EndeavourOS, CachyOS, Garuda, or anything with `ID_LIKE=arch`. Ubuntu path unchanged. Tracked under the broader multi-distro umbrella ([#79](https://github.com/Arkhe-Systems/senddock/issues/79)); Debian, Fedora, RHEL and openSUSE remain TODO. macOS support tracked separately under [#82](https://github.com/Arkhe-Systems/senddock/issues/82).

### Changed

- **README hero replaced with the branded title card image.** New screenshot set (`docs/public/screenshots/`) added for hero + projects dashboard + template editor + campaigns list + analytics dashboard.

### Fixed

- **"Buttons stop responding after login" — password-manager autofill overlay crash** ([#80](https://github.com/Arkhe-Systems/senddock/issues/80)). Bitwarden / 1Password / LastPass inject overlay nodes adjacent to input fields. When a modal opened via `<Teleport to="body">` moved those inputs, the extension's stored reference became orphaned and the next overlay update threw `insertBefore on Node`, intercepting subsequent clicks until a full reload. Fixed by dropping `<Teleport>` from `AppModal` and adding defensive opt-out attributes on `AppInput` for all four major password managers.
- **"Send Email" button felt dead while it actually worked.** `openSendModal` fetched templates async before opening the modal with no visual feedback. Now the button shows "Loading…" and is disabled during the fetch, with an early-return guard against stacked clicks.
- **`fetch()` calls never timed out, leaving components stuck in loading state forever.** Added a 30s default timeout via `AbortController` in the API client, overridable per call via `timeoutMs`.
- **`AppModal` could leave `body.overflow: hidden` after unmount.** `onUnmounted` now resets the style as a safety net.

## [0.6.7] — 2026-06-18

Self-hosting UX release. Headline is a brand-new one-line installer (`curl senddock.dev/install.sh | sudo bash`) that brings up Docker, the compose stack, and one-click updates via Watchtower in under two minutes. Plus a real SMTP-troubleshooting section in the docs, a Cloudflare Tunnel guide as a first-class HTTPS option, and fixes for three bugs that bit users during install: a template editor that escaped `<` while typing, an installer that silently defaulted `PUBLIC_URL` to localhost, and an abandoned Watchtower image that crash-looped on Docker 26+. Drop-in upgrade — no migrations.

### Added

- **One-line installer for Ubuntu** (`scripts/setup.sh`, served at `https://senddock.dev/install.sh`). Installs Docker if missing, downloads the compose file, generates `.env` with random secrets, bundles Watchtower so the dashboard's "Update now" button works out of the box, runs a 60-second health check at the end. Attached as a release asset on every release.
- **Cloudflare Tunnel section in self-hosting docs.** First-class reverse-proxy path for users without a public IP or port forwarding. Documents the two non-obvious gotchas (use `Networking → Tunnels`, not Zero Trust → Connectors; include `http://` in the Service URL explicitly).
- **SMTP troubleshooting guide** (`/guide/smtp`). `/dev/tcp` probe loop to diagnose ISP blocks, five real workarounds for residential SMTP filtering, Postal / Postfix / Mailcow snippets for adding a 2525 listener.

### Changed

- **`docker-compose.image.yml` passes `SENDDOCK_WATCHTOWER_URL` and `SENDDOCK_WATCHTOWER_TOKEN` through** (empty defaults). Manual installs can wire Watchtower without modifying the base compose.
- **Installation docs restructured into four options** — one-line installer (recommended), manual Docker Compose, Dokploy (honest about the marketplace template PR still being open), build from source.
- **Update modal copy is platform-agnostic** when Watchtower isn't configured. Shows "Rebuild from your hosting panel" for Dokploy / Coolify / Portainer users alongside the direct docker compose command, instead of one-size-fits-all SSH advice.

### Fixed

- **Template editor escaped `<` to `&lt;`** while typing in the Code tab. The Visual tab's GrapesJS instance stayed mounted via `v-show` and re-emitted escaped content into the shared model. Switched to `v-if` so GrapesJS is only mounted when the user is on the Visual tab.
- **`setup.sh` silently defaulted `PUBLIC_URL` to localhost** when the user hit Enter on the empty prompt, and skipped re-prompting on subsequent runs. Now: empty prompt loops, localhost requires explicit `yes`, bare domains auto-prefix `https://`, and `SENDDOCK_PUBLIC_URL` always wins over an existing `.env` so a broken install can be patched in place.
- **Watchtower image (`containrrr/watchtower:latest`) was abandoned in 2023** and crash-looped on Docker 26+ daemons (`client version 1.25 is too old`). Switched to the maintained fork `nickfedor/watchtower:latest`. The "Update now" button works again.
- **`setup.sh` Docker check no longer passes when the daemon is dead** (CLI present but daemon down). Adds `docker info` to the probe, attempts `systemctl enable --now docker` before falling back to reinstall.
- **`docs/guide/environment.md`** no longer tells self-hosters to copy `.env.example` from `backend/` (that's source-build only).

### Removed

- **`OPEN_REGISTRATION` env var.** Self-hosted instances should never accept anonymous signups — registration is exclusively via the first-boot Setup screen and the admin "Create user" action. Cloud signup is unaffected (lives in the cloud package).

### Cloud

- Plan-aware Pro paywall — upgrade flow on cloud, self-host license copy on self-host.
- Hide the self-host "Update available" badge on the cloud build.
- Team paywall now selects the right checkout link based on the user's current plan.

## [0.6.5.1] — 2026-06-12

Security and authentication patch on top of [0.6.5](#065--2026-05-28). Adds opt-in two-factor authentication and hardens multi-tenant access checks across the dashboard API. Upgrading is a drop-in image bump; the 2FA migration is applied automatically on startup.

### Added

- **Two-factor authentication (TOTP) with recovery codes (#65).** Enable 2FA from the Account page: scan the QR with any authenticator app (Google Authenticator, 1Password, Authy, …), confirm a 6-digit code, and save the one-time recovery codes shown once at setup. Login becomes a two-step flow — password first, then the 6-digit code (or a recovery code if you lose the device), exchanged through a short-lived intermediate token. Recovery codes are single-use and bcrypt-hashed at rest, and disabling 2FA requires a valid code. Works the same self-hosted or on cloud.

### Security

- **Multi-tenancy hardening across project-scoped endpoints.** Authorization on the audit-log, webhooks and analytics sections now verifies workspace membership rather than project ownership alone, webhook delivery listings are constrained to the owning project, and the email service gained a defense-in-depth context guard so a project ID can only be acted on inside a request that was authorized for it. Internal-only webhook lookups are now clearly separated from tenant-scoped queries. No action is required on upgrade.

## [0.6.5] — 2026-05-28

Quality-of-life release: persistent profile sidebar, Account and Billing pages, one-click updates via Watchtower, an email log detail drawer, plus a fix that closes the loop on broadcast restart recovery from [0.6.4](#064--2026-05-08).

### Added

- **Persistent profile sidebar across every authenticated screen.** The dashboard now uses the same flex+sidebar layout as the project view, so the page doesn't look empty when you have one or two projects. The sidebar is sticky to the viewport (no longer scrolls away with the main content) and embeds a profile panel at the bottom with avatar, name, email, plan badge, and links to Account / Billing / Logout. Same panel appears under the project navigation when you're inside a project.
- **Account page (`/account`).** Read-only profile (name, email, plan, member since) plus a change-password form. Validates current password server-side and applies the same complexity rules as registration (≥8 chars, one uppercase, one digit, one special). Email and name changes are not supported yet — they need an email-verification flow that doesn't exist in the codebase yet.
- **Billing page (`/billing`).** Shows the current tier (Free / Pro / Team) with a colored badge, plus the license `expires_at` and `last_check` timestamps when a license is active. Free users see Pro and Team paywall cards with the existing checkout links. Notes explain that BYOS means SendDock doesn't charge for email volume — only for features.
- **One-click updates from the dashboard via Watchtower.** When `SENDDOCK_WATCHTOWER_URL` (and optional `SENDDOCK_WATCHTOWER_TOKEN`) are set and Watchtower's HTTP API is reachable, the update modal shows an "Update now" button that POSTs to Watchtower's `/v1/update`. The dashboard then polls `/api/v1/version` every 2 seconds, tolerates the network errors that happen while the container restarts, and shows a success toast once the new version is live. When Watchtower isn't configured (the default) the modal still shows the manual `docker compose pull && up -d` command as before — see [Updating → One-click updates](./self-hosting/updating#one-click-updates-from-the-dashboard-watchtower).
- **Email log detail drawer (#44).** Click a row in the Logs page and a right-side drawer slides in with the full lifecycle: a vertical timeline (sent → opened → clicked → bounced / failed / suppressed), per-click event list with URL + timestamp + user agent, error or suppression reason, and reference IDs. The list view also gained an **Engagement** column with compact `O` / `C` badges so you can spot opens and clicks without opening each row. Backed by a new `GET /api/v1/projects/{id}/logs/{logId}` endpoint that returns `{log, clicks}`.
- **`pg_trgm` gin indexes on `email_logs.to_email` and `email_logs.subject` (#44).** Substring `ILIKE '%x%'` searches now use a real index instead of sequential scan, which keeps the Logs search responsive past a few thousand rows. The migration enables `CREATE EXTENSION pg_trgm` if it isn't already.
- **Custom dark-themed scrollbars** across the whole app (subtle gray thumb, transparent track, lighter on hover, both WebKit and Firefox).

### Changed

- **`AppModal` now uses `Teleport` to `body`.** Any modal triggered from inside a sidebar (or any element with a `transform`) now correctly overlays the entire viewport instead of getting trapped inside its parent's bounding box. Fixes the "What's in vX.Y.Z" dialog rendering inside the sidebar.
- **`GET /api/v1/me`** now returns `email`, `name`, `plan` and `created_at` in addition to `user_id`. Needed by the new profile panel; no breaking change since the response was previously `{user_id}`.
- **Analytics, Webhooks and Audit log sections** check the user's license tier before fetching. Free users see the paywall immediately instead of waiting for a 402 / 404 round-trip and falling through to a generic "Couldn't load" error.

### Fixed

- **Broadcasts marked `interrupted` on restart even though they actually resumed sending (#41 follow-up).** A backend restart in the middle of a broadcast flipped its status to `'interrupted'` even though the per-recipient jobs were still being drained by the worker pool. The UI then told the user the broadcast had failed and the remaining recipients didn't get the email — which was wrong. The status now stays `'sending'` until every job has settled, then transitions to `'completed'`. Pro Analytics "broadcasts in flight" also recovers correctly across restarts now (it filters on `status='sending'` and was losing visibility because of this).

### API

- `GET /api/v1/me` — now returns `email`, `name`, `plan`, `created_at` alongside `user_id`.
- `POST /api/v1/me/password` — change the authenticated user's password. Body: `{current_password, new_password}`.
- `GET /api/v1/update/auto-status` — returns Watchtower availability: `{configured, healthy, url, last_check, last_error}`.
- `POST /api/v1/update/trigger` — fires the Watchtower `/v1/update` call and returns `202`. Returns `400` if Watchtower isn't configured, `502` if it's unreachable.
- `GET /api/v1/projects/{id}/logs/{logId}` — returns `{log, clicks}` for the email log detail drawer.

### Environment

- New optional `SENDDOCK_WATCHTOWER_URL` and `SENDDOCK_WATCHTOWER_TOKEN` — enable the dashboard's one-click update flow. When unset, the update modal falls back to the manual command.

### Migrations

- `20260528193111_add_email_logs_search_index.sql` — enables `pg_trgm` if missing and adds gin indexes on `email_logs.to_email` and `email_logs.subject`. Idempotent, no data changes.

## [0.6.4.1] — 2026-05-08

Hotfix on top of [0.6.4](#064--2026-05-08). No new features, no DB migrations — drop-in replacement for any 0.6.x deployment showing the "restart every ~5 hours" pattern.

### Fixed

- **Container restarted every ~5 hours after boot due to its own healthcheck being rate-limited.** `Redis.Increment` reset the key's TTL on every call (`INCR` followed unconditionally by `EXPIRE`). The Dockerfile healthcheck (`wget /health` every 30s) kept the `rl:::1` counter's key alive forever instead of expiring per-minute, so the counter accumulated cumulatively. After ~5 hours (600 hits ÷ 2/min), it crossed the 600 req/min threshold and `/health` started returning `429`. wget's `-q … || exit 1` turned that into exit 1; swarm killed the "unhealthy" container after 5 retries; the new container started fresh and the cycle repeated. Increment now uses an atomic Lua script that only sets the TTL when the key is created (`count == 1`), giving a real fixed-window counter.
- **`/health` was wrapped by the rate limiter, CORS, and body-size middlewares.** Healthchecks must never be subject to client-facing throttles. `/health` is now mounted on a separate root mux that bypasses all middleware.

### Added

- **`RATE_LIMIT_PER_MINUTE` env var (default `600`).** Used to be hard-coded. Lower it on small deployments or raise it for high-traffic apps without rebuilding the image.

## [0.6.4] — 2026-05-08

> **⚠️ Deprecated — upgrade to [0.6.4.1](#0641--2026-05-08).** Affected by a Redis rate-limiter bug that crashes the container in a restart loop ~5 hours after every boot. The fix lives in 0.6.4.1 with no migrations or breaking changes.



### Added

- **Per-recipient broadcast queue with retries (#41).** Broadcasts and scheduled campaigns no longer run in a single goroutine that loses everything on restart. Each recipient is enqueued as a row in a new `broadcast_jobs` table; five worker goroutines drain it concurrently using `SELECT … FOR UPDATE SKIP LOCKED`. Transient SMTP errors (4xx, network, DNS) reschedule the job with exponential backoff (30s, 2m, 8m, 30m, 1h), capping at five attempts before the recipient is marked `failed`. SMTP 5xx bounces are not retried — they are tagged `bounced` and the recipient is added to the suppression list, same as before. A backend restart mid-broadcast resets jobs that were `sending` to `retry` so sending resumes from where it left off, with no recipient sent to twice and none silently dropped. See [Sending → How a broadcast actually runs](./guide/sending#how-a-broadcast-actually-runs).
- **Live progress for newsletters and broadcasts.** The Newsletters list shows `sent_count` / `failed_count` updating every five seconds while a campaign is in `sending`, instead of jumping from `0/0` to the final number at the end. Same goes for the Broadcasts tab progress bars.
- **"Broadcasts in flight" panel on Pro Analytics (#41).** When at least one broadcast is actively sending, a card appears at the top of the Analytics dashboard with a live progress bar per broadcast (`X / total`, percentage), elapsed time, and per-status counters. Polls every five seconds and disappears as soon as the queue drains. Useful for watching big sends to 50k+ subscriber lists.
- **Email Logs filters and CSV export (#44).** Status filter is now a row of clickable chips (Sent / Failed / Bounced / Suppressed) with status-tinted highlight. Added a Template dropdown to filter logs to a single template's sends. New **Export CSV** button (top right of Logs) downloads every row matching the active filters — same query parameters as `/logs`, no `limit`/`offset`, served as `text/csv`. Columns: id, recipient, subject, status, error, sent_at, opened_at, clicked_at, template_id, subscriber_id. See [API → Email Logs](./api/sending#email-logs).
- **Dark color-scheme on native date pickers.** All `<input type="date">` widgets across the app now render their popup in dark mode (Chromium and Firefox), matching the rest of the UI.

### Fixed

- **`/health` no longer kills the container on transient Postgres latency.** v0.6.3 introduced a synchronous 2-second `PingContext` on every `/health` call. When Postgres took >2s for any reason (slow disk, GC pause, momentary IO contention shared with neighbors on a small VPS), healthcheck failed five times in a row over ~3 minutes and the orchestrator killed the container — only to start a new one which had the same problem. The endpoint is now backed by a background goroutine that pings Postgres every 10s with a 5s timeout and stores the last-success Unix timestamp atomically. `/health` reads the atomic and returns `503` only if no successful ping has happened in the last 60 seconds. Single blips no longer trigger crash-replace loops; sustained Postgres outages still surface to the orchestrator.
- **Newsletter shown as `sent` with `213/0` immediately on schedule.** With the new queue, `EmailService.Broadcast` returns in milliseconds (only enqueues), so the campaign worker — which used to call `Broadcast` and immediately mark the campaign as `sent` with the upfront recipient count — was claiming the campaign was finished before any recipient had been sent to. Campaigns now stay in `sending`, linked to their broadcast via a new `campaigns.broadcast_id` FK; when the broadcast worker drains the queue it cascades the real `sent_count` / `failed_count` back to the linked campaign and flips its status to `sent` with `sent_at` set.

### Documentation

- **New Docker Swarm / Dokploy troubleshooting section.** Documents the recurring "all services on the same host stop talking to each other" pattern — usually caused by accumulated dead containers and orphan services confusing the embedded DNS resolver `127.0.0.11:53`, which makes every service on the overlay network fail to resolve names simultaneously. Includes diagnostic commands and a safe cleanup sequence (`docker service rm` orphans, `docker container prune`, `docker network prune`) plus an explicit warning to never use `volume prune` or `system prune -a`. See [Troubleshooting → Docker Swarm / Dokploy](./self-hosting/troubleshooting#docker-swarm-dokploy).
- **Tag-cache `Image is up to date` workaround documented.** When Dokploy or Docker Swarm refuses to pull a new image because the tag (`:dev`, `:latest`) appears unchanged locally, enable Clean Cache in the UI or use `docker service update --force --image …` from the host.

## [0.6.3] — 2026-05-06

> **⚠️ Deprecated — upgrade to [0.6.4.1](#0641--2026-05-08).** Same restart-loop bug as 0.6.4 (rate limiter eventually 429-ing the container's own healthcheck), plus its own crash trigger from the synchronous DB ping introduced here. Both are fixed in 0.6.4.1.



### Fixed

- **`/health` now actually checks the database.** Previously it returned `{"status":"ok"}` without touching Postgres. When Postgres entered a stuck state (idle-in-transaction backlog, OOM, network partition), Docker still saw the container as healthy and the reverse proxy kept routing traffic to it. Each request that touched the database hung, and the proxy returned `502 Bad Gateway` to users. Restarting just the app didn't fix it because Postgres was the broken side. `/health` now runs a 2-second `PingContext`; if it fails, the endpoint returns `503` and the orchestrator can mark the container unhealthy and react.
- **Database connection pool capped.** Added `MaxOpenConns=25`, `MaxIdleConns=5`, `ConnMaxLifetime=5m`, `ConnMaxIdleTime=2m`. Without these, the default Go `sql.DB` lets connections accumulate without bound — and zombie sockets after a Postgres outage never get recycled. With these, dead connections age out within minutes even if `/health` somehow doesn't catch the issue first.

## [0.6.2] — 2026-05-06

### Fixed

- **Workspace role assignment was silently broken for `admin`, `developer` and `viewer`.** The internal `normalizeRole` helper that runs on every member assignment predated the v0.6 role expansion and only recognised `owner` and `member` — anything else came back as `400 invalid role`. Capability enforcement was already aware of all five roles, so the bug was strictly on the *assignment* path: an owner trying to invite a developer or change a member to viewer hit a hard `400`. Fixed in `service.normalizeRole`. Existing memberships are untouched.
- **`docker-compose.image.yml` healthcheck and license endpoint defaults.** The compose-level healthcheck overrides the image's; on hosts with `net.ipv6.bindv6only=1` it was still tripping the crash-replace loop fixed in v0.6.1 even after pulling the new image. Switched to `localhost` and longer grace (start-period `30s`, retries `5`). Removed the broken `SENDDOCK_LICENSE_ENDPOINT` default that pointed at a non-existent host; the binary now reaches its real default (`https://api.lemonsqueezy.com/v1`) when the env var is empty.

## [0.6.1] — 2026-05-06

### Fixed

- **Healthcheck reliability on IPv6-only hosts.** On hosts with `net.ipv6.bindv6only=1` (some VPS, Docker Swarm clusters and cloud providers) the Go server's `:8080` listener bound IPv6-only while the container `HEALTHCHECK` targeted `127.0.0.1:8080`, causing every healthcheck to fail. The orchestrator marked the task `unhealthy` after three retries, killed it and started a new one — a 60–90s crash-replace loop with no panic and no error log, just intermittent `Bad Gateway` from the reverse proxy. Fixed by binding the listener to `0.0.0.0:8080` explicitly and switching the `HEALTHCHECK` to `localhost`. Start-period grace was also raised to `30s` with `5` retries so a slow first-boot DB doesn't tip the container into a loop.

## [0.6.0] — 2026-04-30

### Added

- **Workspaces.** Multi-user collaboration with workspace-scoped projects. Existing single-user installs were backfilled — every user got a `My Workspace` with their projects under it. See [Workspaces guide](./guide/workspaces).
- **Roles & capabilities.** Four roles (`owner`, `admin`, `developer`, `viewer`) with a fixed capability matrix. Developers can call `/send` (transactional) but cannot broadcast or edit templates; viewers are read-only; admins do everything except member management.
- **Team plan tier.** Multi-member workspaces, role management and the admin "Create user" flow belong to a new **Team** plan above Pro. Without a Team license the endpoints return `402` and the Members page shows a paywall; single-user organization stays free.
- **In-app Subscribe links to Lemon Squeezy.** Pro and Team paywalls jump straight to checkout for the right tier; Pro pre-applies the launch promo while it is active.
- **Pro vs Team gating in the validator.** `Status.Tier` is exposed and `AllowsFeature(feature)` consults a feature → tier matrix, so a Pro license can no longer unlock Team-only endpoints.
- **Email validation on import (#43).** CSV/JSON imports now check syntax, MX records and a built-in disposable-domain list. Rejected rows surface in the import results modal with a per-row reason.
- **File picker and drag-and-drop for CSV imports.**
- **Per-project suppression list (#42).** New `suppressions` table; `/send`, `/send/batch` and `/broadcast` skip suppressed recipients and account for them in the result counts. Existing unsubscribed subscribers are backfilled. Manage entries from the Suppressions tab inside a project.
- **Per-project audit log (#47, Pro).** Records 12 sensitive actions (project create/delete, SMTP/IMAP/bounce updates, API key create/revoke, webhook create/delete, suppression add/remove, login). Exposed at `GET /audit-log` and surfaced in the Audit Log tab. Pro-gated.
- **Bounce ingestion (#38).**
  - **Part A — in-session 5xx detection.** RCPT TO failures with 5xx are classified, marked `bounced` in the email log, added to the suppression list and dispatched as `email.bounced` webhooks.
  - **Part B — public webhook endpoint.** `POST /webhooks/bounces/{projectId}?token=<bounce-token>` accepts generic JSON or Mailgun event payloads. Each project gets a rotatable token; the URL is shown in the Settings tab.
  - **Part C — IMAP poller.** Configure a bounce mailbox per project; a 5-minute poller logs in over TLS, scans `INBOX` for unread DSNs, extracts recipients (RFC 3464 `Final-Recipient` first, then 5xx-line fallback), adds them to the suppression list and marks the messages `\Seen`.
- **Standardized UI primitives.** `AppCheckbox` and `AppConfirmModal` replace inconsistent checkbox styles and native `confirm()` calls.

### Changed

- Suppressed sends are now a distinct outcome end-to-end — `bounced` and `suppressed` are tracked separately from `sent` and `failed` in stats and the broadcast result.
- Import results modal widened so the outcome cards and rejected-rows table fit without overflowing.

### Fixed

- `AppCheckbox` is now clickable outside `<label>` wrappers (the off-screen `peer sr-only` input did not catch clicks inside table cells).

## [0.5.2] — 2026-04-30

### Security

- **Critical: license bypass closed.** Versions 0.5.0 and 0.5.1 auto-unlocked every Pro feature on any self-hosted deployment with an empty `SENDDOCK_LICENSE_KEY`. The validator now returns `LockedFree` for an empty key in every deployment mode. Core stays fully usable; Pro endpoints return `402 Payment Required` until a valid key is set.

### Changed

- Documentation swept (README, configuration, installation, environment, analytics, webhooks, API webhooks) to remove the obsolete "self-hosted unlocks Pro locally" wording.
- Releases v0.5.0 and v0.5.1 marked as deprecated on GitHub.

## [0.5.1] — 2026-04-29

> Deprecated — contains the v0.5.0 license bypass. Upgrade to 0.5.2.

### Fixed

- Hide the update modal in cloud deployments. `/version` returns `enabled: false` when `DEPLOYMENT_MODE=cloud`.
- Render release notes as Markdown instead of raw text inside a `<pre>` block.
- Default upgrade command points to `docker compose pull && docker compose up -d`.
- Pricing links resolve again — three references to `senddock.com/pricing` switched to `senddock.dev/pricing`.

### Added

- The current-version label is always clickable, opening the same modal with the upgrade section hidden.

## [0.5.0] — 2026-04-29

> Deprecated — contains a license bypass. Upgrade to 0.5.2.

### Added

- **Webhooks (Core + Pro).** HMAC-SHA256 signed dispatcher with seven event types, exponential backoff retries (30s → 2m → 10m → 30m → 2h, 5 attempts), full management UI in Pro.
- **Click tracking (Core).** Every `<a href>` in outgoing emails routes through `/c/{logId}/{payload}` with HMAC-protected tokens.
- **Pro Analytics dashboard.** Funnel, opens-over-time, top templates, top clicked links, auto-generated insights, date presets, trend pills.
- **One-click unsubscribe (RFC 8058).**
- **Pro license validator** against Lemon Squeezy.
- **Container registry split.** Public tagged releases stay on `ghcr.io/arkhe-systems/senddock`; `:dev` builds moved to a private package.
- **Per-project rate limits** on `/send`, `/send/batch` and `/broadcast`.
- **SMTP test endpoint** with bounded 5s connect / 10s session timeouts.
- **`AppButton`** `size="sm"` and `variant="ghost"`.
- **Docs overhaul** with hand-crafted SVG diagrams that follow the VitePress theme toggle via `currentColor`.

## [0.4.0] — 2026-04-15

Foundation for the open-core release. Highlights:

- BYO SMTP per project with encrypted password storage.
- Subscribers, templates, sends, broadcasts, campaigns.
- API keys with hashed secrets and per-key rate limits.
- Open-tracking pixel and unsubscribe links.
- Initial dashboard, project switcher, settings UI.
