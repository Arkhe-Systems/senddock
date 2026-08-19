# Changelog

All notable changes to SendDock are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases are also published on [GitHub](https://github.com/arkhe-systems/senddock/releases) and the `ghcr.io/arkhe-systems/senddock` image carries the matching tag.

## [Unreleased]

_Nothing here yet. Track upcoming work on the [open issues](https://github.com/arkhe-systems/senddock/issues)._

## [0.8.0] — 2026-07-28

Analytics graduates to free Core, self-host configuration moves into the dashboard, and the paid tier is redrawn around deliverability and reporting.

### Added

- **Analytics, rebuilt and free (Core).** The old single-endpoint funnel became a tabbed, charted dashboard — **Overview, Campaigns, Audience, Engagement** — with a date-range picker, per-campaign breakdowns (every email log is now tagged with its `broadcast_id`), a subscriber-growth view and a weekday × hour click **heatmap**. Async bounces are now recorded on the log, not only the suppression list.
- **Instance settings in the dashboard (self-host).** The public URL, a configurable **session inactivity timeout** (5–1440 min) and the **Pro license key** are set from **Instance** in the dashboard, stored in the database, applied live with no restart.
- **Spam-complaint handling (Core).** An FBL/complaint webhook (`POST /webhooks/complaints/{projectId}`) ingests complaints and auto-suppresses the complainer, and the **complaint rate** shows in the free analytics — reputation-critical, so it stays in Core, ungated.
- **Deliverability (Pro).** A new Analytics tab: domain health (SPF/DKIM/DMARC) and a **per-provider breakdown** of acceptance/bounce/open/click and spam-complaint rates.
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

- **Custom fields for subscribers ([#72](https://github.com/arkhe-systems/senddock/issues/72), Core).** Typed, project-scoped attributes on top of the existing `subscribers.metadata` column: `string`, `number`, `date`, `boolean`, `enum`. Define them under **Settings → Custom Fields**; values are validated on write (unknown keys rejected), render in templates as `{{custom.KEY}}`, show as columns and per-type inputs in the subscribers table, and map from extra CSV columns on import. New endpoints under `/api/v1/projects/{id}/fields`.
- **Tags + segments ([#40](https://github.com/arkhe-systems/senddock/issues/40), Core).** Subscribers carry free-form tags (single, bulk, and inline-on-import). Segments are saved filters — a match-all/any predicate over `status`, `tags` and `custom.*` fields — with a live match count while you build. Broadcasts accept a `segment_id` to send to a subset instead of all active subscribers. New endpoints under `/api/v1/projects/{id}/segments` and `/tags`.
- **Segment filter on Pro Analytics.** The analytics overview accepts an optional `segment_id` to scope every metric to a segment's members.

### Changed

- **Webhooks are now free (Core).** The management UI and REST API (create/list/pause/delete, delivery history) moved out of the Pro tier — webhooks are developer table stakes. The dispatcher, HMAC signing and retries were already in Core; now nothing about webhooks requires a `SENDDOCK_LICENSE_KEY`. The paid tier is now Analytics + Audit log + Team.

### Security

- **Authentication hardening.** Per-account throttling on login and two-factor verification, all sessions revoked on password change, API keys constrained to their intended capabilities, and a safer default workspace role.

## [0.6.8] — 2026-06-22

Two new features and a chunky bag of UI fixes. Headline is the **community starter template library**: every SendDock instance — Cloud and self-hosted — now ships a "★ Browse library" modal on the Templates page that pulls templates from a separate community-maintained repo, clones one into your project on click, and opens it in the editor. The library is open to PRs (the only SendDock repo that accepts them by design). Second headline is **Arch Linux support in the one-line installer** — `curl senddock.dev/install.sh | sudo bash` now works on Arch and its derivatives (Manjaro, EndeavourOS, CachyOS, Garuda) the same way it works on Ubuntu. Plus a fix for the long-standing "buttons stop responding after login" bug that turned out to be password-manager autofill extensions crashing on Vue Teleport. Drop-in upgrade — no migrations.

### Added

- **Community template library** ([#73](https://github.com/Arkhe-Systems/senddock/issues/73)). New "★ Browse library" button on the Templates page opens a modal with a categorized grid of starter templates (welcome flows, monthly digests, single-story newsletters, product launches, changelogs, weekly link roundups, password resets, email verifications, transactional receipts). Click a template, get a clone in your project, ready to edit. Library content lives in the public [`Arkhe-Systems/senddock-templates`](https://github.com/Arkhe-Systems/senddock-templates) repo — community PRs welcome, MIT-licensed. New backend service caches the manifest in Redis for 1 hour; per-template HTML is fetched on demand. Override the source with `TEMPLATE_LIBRARY_URL` to point at a private fork or curated internal gallery.
- **Arch Linux family support in the one-line installer** ([#81](https://github.com/Arkhe-Systems/senddock/issues/81)). `scripts/setup.sh` now branches on `/etc/os-release` and runs the pacman-based Docker install path on Arch, Manjaro, EndeavourOS, CachyOS, Garuda, or anything with `ID_LIKE=arch`. Ubuntu path unchanged. The die message for unsupported distros now lists both supported families and links to the manual Compose path. Tracked under the broader multi-distro umbrella ([#79](https://github.com/Arkhe-Systems/senddock/issues/79)); Debian, Fedora, RHEL and openSUSE remain TODO. macOS support tracked separately under [#82](https://github.com/Arkhe-Systems/senddock/issues/82).

### Changed

- **README hero replaced with the branded title card image.** The `# SendDock` h1 is gone; the visual wordmark in the hero card serves as the title now. Repo-card / search-results display falls back to the repo name, no SEO regression. New screenshot set (`docs/public/screenshots/`) added for hero + projects dashboard + template editor + campaigns list + analytics dashboard.

### Fixed

- **"Buttons stop responding after login" — password-manager autofill overlay crash** ([#80](https://github.com/Arkhe-Systems/senddock/issues/80)). Bitwarden / 1Password / LastPass inject overlay nodes adjacent to input fields. When a modal opened via `<Teleport to="body">` moved those inputs to a new DOM position, the extension's stored reference became orphaned and the next overlay update threw `insertBefore on Node` — corrupting the extension's internal DOM tracking and intercepting subsequent click events on the page until the next full reload. Fixed by dropping `<Teleport>` from `AppModal` (the modal still uses `fixed inset-0 z-50` so it overlays the viewport the same way) and adding defensive opt-out attributes on `AppInput` for all four major password managers. Credential fields on login/register opt back in via explicit `autocomplete` values so they still autofill normally.
- **"Send Email" button felt dead while it actually worked.** `openSendModal` was fetching templates async before opening the modal, with no visual feedback on the button. Now the button shows "Loading…" and is disabled during the fetch; an early-return guard prevents stacked clicks from queuing multiple in-flight requests.
- **`fetch()` calls never timed out, leaving components stuck in loading state forever.** A hung backend (SMTP test against an unreachable host, slow tunnel, dead connection) would leave the calling component spinning indefinitely. Added a 30s default timeout via `AbortController` + signal in the API client, overridable per call via the `timeoutMs` option. This is the underlying cause of several "buttons stop working" reports.
- **`AppModal` could leave `body.overflow: hidden` after unmount.** If a modal unmounted while still shown (e.g. mid-navigation), the body could be left scroll-locked until the next reload. `onUnmounted` now resets the style as a safety net.
- **Truncated template names in the library grid had no way to read the full text.** Native `title` attribute on the card header — hover anywhere on the truncated name shows the full template name in the browser's native tooltip.

## [0.6.7] — 2026-06-18

Self-hosting UX release. Headline is a brand-new one-line installer (`curl senddock.dev/install.sh | sudo bash`) that brings up Docker, the compose stack, and one-click updates via Watchtower in under two minutes. Plus a real SMTP-troubleshooting section in the docs, a Cloudflare Tunnel guide as a first-class HTTPS option, and fixes for three bugs that bit users during install: a template editor that escaped `<` while typing, an installer that silently defaulted `PUBLIC_URL` to localhost, and an abandoned Watchtower image that crash-looped on Docker 26+. Drop-in upgrade — no migrations.

### Added

- **One-line installer for Ubuntu** (`scripts/setup.sh`, also served at `https://senddock.dev/install.sh`). Installs Docker from the official repo if missing, downloads `docker-compose.image.yml`, generates `.env` with random secrets (mode 600), prompts for `PUBLIC_URL`, bundles Watchtower in HTTP-API-only mode (no auto-polling, label-scoped to SendDock only) so the dashboard's "Update now" button works out of the box, brings up the stack, and runs a 60-second health check that warns if any service isn't running. The script is attached as a release asset on every published release.
- **Cloudflare Tunnel section in self-hosting docs.** First-class reverse-proxy path for users without a public IP, port forwarding, or Let's Encrypt setup. Documents the two non-obvious gotchas that took hours to debug the first time: use `Networking → Tunnels` (NOT Zero Trust → Connectors — that's a different feature), and include the `http://` scheme explicitly in the Service URL (without it Cloudflare defaults to HTTPS and the tunnel returns 502 because SendDock listens on plain HTTP internally).
- **SMTP troubleshooting guide** (`/guide/smtp`). New "Diagnosing port issues" section with a `/dev/tcp` probe loop that tells you exactly which ports your network actually allows. New "Working around ISP blocks" with five real fixes: try port 25 explicitly, use a provider that supports 2525 (Brevo, Mailgun, Postmark, SendGrid), run on a cloud server, run a cloud SMTP relay, or use Mailpit for local dev. New "Self-hosted mail server" section with config snippets for adding a 2525 listener to Postal, Postfix and Mailcow, plus the UFW commands to open the new port server-side.

### Changed

- **`docker-compose.image.yml` passes `SENDDOCK_WATCHTOWER_URL` and `SENDDOCK_WATCHTOWER_TOKEN` through** (empty defaults). Manual installs can now wire Watchtower via the override block in the docs without modifying the base compose file. Backwards-compatible for installs that don't set those vars.
- **Installation docs restructured into four options:** (1) one-line installer (recommended for Ubuntu), (2) manual Docker Compose (any Linux), (3) Dokploy (honest about [PR #821 to Dokploy templates](https://github.com/Dokploy/templates/pull/821) still being open for review), (4) build from source. Cross-doc references in `updating.md`, `configuration.md` and `environment.md` re-numbered to match.
- **Update modal copy is platform-agnostic when Watchtower isn't configured.** Previously the modal told everyone to run `docker compose pull && up -d` from the host — wrong advice for Dokploy / Coolify / Portainer users who don't manage the stack at that level. New layout shows two equally visible cards: "Rebuild from your hosting panel" (primary, for platform users) and "Direct Docker control" with the docker compose command (secondary, for bare-metal and `install.sh` users), plus a tip linking to the Watchtower setup for one-click updates from the dashboard.

### Fixed

- **Template editor escaped `<` to `&lt;` while typing in the Code tab.** The Visual tab's GrapesJS instance stayed mounted via `v-show` even when hidden. Its watcher fired on every CodeMirror keystroke and called `setComponents()` with the in-progress string. For incomplete HTML like a lone `<`, GrapesJS interpreted it as text content and re-emitted `&lt;` back into the shared `v-model`, feeding the bug character by character. Switched the Visual wrapper to `v-if` so GrapesJS is only mounted when the user is actually on that tab.
- **`scripts/setup.sh` silently defaulted `PUBLIC_URL` to localhost.** Hitting Enter at an empty prompt previously triggered a warning that was easy to miss in the install output and produced an instance with newsletters and unsubscribe links disabled — which the user discovered hours later in the dashboard with no breadcrumb to the cause. Worse, re-running the script on an existing `.env` skipped the prompt entirely, so there was no way to fix it through the installer. New behaviour: empty prompt loops until non-empty input or Ctrl+C; `localhost`/`127.0.0.1` requires explicit confirmation (or `SENDDOCK_ALLOW_LOCALHOST=yes` for non-interactive); bare domains auto-prefix `https://`; `SENDDOCK_PUBLIC_URL` env var always wins over existing `.env` so a broken install can be patched by re-running with the correct URL — secrets are preserved.
- **Watchtower image crashed in a restart loop on Docker 26+.** `containrrr/watchtower:latest` — the canonical image when `setup.sh` first shipped — has been abandoned since 2023 and bundles a Docker client too old (API 1.25) for modern Docker daemons (which require API 1.40+). Manifested as `client version 1.25 is too old. Minimum supported API version is 1.40` in the logs, silent restart loop in `docker compose ps`, and a non-working "Update now" button. Switched both `setup.sh` and the manual recipe in `updating.md` to `nickfedor/watchtower:latest`, the actively-maintained community fork. Inline comment explains the abandonment so future readers know why.
- **`scripts/setup.sh` no longer reports "Docker already installed" when the daemon is dead.** Previous check was `command -v docker && docker compose version` — both pass on hosts where the CLI is present but the daemon isn't running (the state left by an incomplete purge, snap-installed Docker, or some pre-bundled Ubuntu images). New check adds `docker info` and attempts `systemctl enable --now docker` before falling back to a clean reinstall.
- **`docs/guide/environment.md` no longer tells self-hosters to copy `.env.example` from `backend/`.** That instruction applies to source builds only; self-hosters now get pointed at the installation guide and the `.env.production.example` reference.

### Removed

- **`OPEN_REGISTRATION` env var.** It was a design footgun: self-hosted instances should never accept anonymous signups — that's the whole point of self-hosting your own list. Registration goes through the first-boot Setup screen and the admin "Create user" action exclusively. Cloud signup is in the separate cloud package and unaffected.

### Cloud

(carryover from the v0.6.6 work that never published as its own release)

- Plan-aware Pro paywall — shows the upgrade flow on cloud, the self-host license copy on self-host. No more confusing copy where cloud users see "enter your license key".
- Hide the self-host "Update available" badge on the cloud build (the badge is meaningless when the platform manages the version).
- Team plan paywall now selects the right checkout link based on the user's current plan.

## [0.6.5.1] — 2026-06-12

Security and authentication patch on top of [0.6.5](#065--2026-05-28). Adds opt-in two-factor authentication and hardens multi-tenant access checks across the dashboard API. Upgrading is a drop-in image bump; the 2FA migration is applied automatically on startup.

### Added

- **Two-factor authentication (TOTP) with recovery codes (#65).** Enable 2FA from the Account page: scan the QR with any authenticator app (Google Authenticator, 1Password, Authy, …), confirm a 6-digit code, and save the one-time recovery codes shown once at setup. Login becomes a two-step flow — password first, then the 6-digit code (or a recovery code if you lose the device), exchanged through a short-lived intermediate token. Recovery codes are single-use and bcrypt-hashed at rest, and disabling 2FA requires a valid code. Works the same self-hosted or on cloud.

### Security

- **Multi-tenancy hardening across project-scoped endpoints.** Authorization on the audit-log, webhooks and analytics sections now verifies workspace membership rather than project ownership alone, webhook delivery listings are constrained to the owning project, and the email service gained a defense-in-depth context guard so a project ID can only be acted on inside a request that was authorized for it. Internal-only webhook lookups are now clearly separated from tenant-scoped queries. No action is required on upgrade.

## [0.6.5] — 2026-05-28

Quality-of-life release: persistent profile sidebar, Account and Billing pages, one-click updates via Watchtower, an email log detail drawer, plus a fix that closes the loop on broadcast restart recovery from [0.6.4](#064--2026-05-08).

### Added

- **Persistent profile sidebar across every authenticated screen.** The dashboard now has the same flex+sidebar layout as the project view, so the page doesn't look empty when you have one or two projects. The sidebar is sticky to the viewport (no longer scrolls away with the main content) and embeds a new profile panel at the bottom with avatar, name, email, plan badge, and links to Account / Billing / Logout. Same panel appears under the project navigation when you're inside a project.
- **Account page (`/account`).** Read-only profile (name, email, plan, member since) plus a change-password form. Validates current password server-side and applies the same complexity rules as registration (≥8 chars, one uppercase, one digit, one special). Email and name changes are not supported yet — they need an email-verification flow that doesn't exist in the codebase yet.
- **Billing page (`/billing`).** Shows the current tier (Free / Pro / Team) with a colored badge, plus the license `expires_at` and `last_check` timestamps when a license is active. Free users see Pro and Team paywall cards with the existing LemonSqueezy checkout links. Notes explain that BYOS means SendDock doesn't charge for email volume — only for features.
- **One-click updates from the dashboard via Watchtower (#41 follow-up).** When `SENDDOCK_WATCHTOWER_URL` (and optional `SENDDOCK_WATCHTOWER_TOKEN`) are set and Watchtower's HTTP API is reachable, the update modal shows an "Update now" button that POSTs to Watchtower's `/v1/update`. The dashboard then polls `/api/v1/version` every 2 seconds, tolerates the network errors that happen while the container restarts, and shows a success toast once the new version is live. When Watchtower isn't configured (the default) the modal still shows the manual `docker compose pull && up -d` command as before — see the new ["One-click updates" section](https://docs.senddock.dev/self-hosting/updating#one-click-updates-from-the-dashboard-watchtower) in the updating docs.
- **Email log detail drawer (#44).** Click a row in the Logs page and a right-side drawer slides in with the full lifecycle: a vertical timeline (sent → opened → clicked → bounced / failed / suppressed), per-click event list with URL + timestamp + user agent, error or suppression reason, and reference IDs. The list view also gained an **Engagement** column with compact `O` / `C` badges so you can spot opens and clicks without opening each row. Backed by a new `GET /api/v1/projects/{id}/logs/{logId}` endpoint that returns `{log, clicks}`.
- **`pg_trgm` gin indexes on `email_logs.to_email` and `email_logs.subject` (#44).** Substring `ILIKE '%x%'` searches now use a real index instead of sequential scan, which keeps the Logs search responsive past a few thousand rows. The migration enables `CREATE EXTENSION pg_trgm` if it isn't already.
- **Custom dark-themed scrollbars** across the whole app (subtle gray thumb, transparent track, lighter on hover, both WebKit and Firefox).

### Changed

- **`AppModal` now uses `Teleport` to `body`.** Any modal triggered from inside a sidebar (or any element with a `transform`) now correctly overlays the entire viewport instead of getting trapped inside its parent's bounding box. Fixes the "What's in vX.Y.Z" dialog rendering inside the sidebar.
- **`AppModal` defaults to `size="lg"` in the update dialog**, with release-note `<pre>` blocks set to `pre-wrap` so long commands wrap instead of producing a horizontal scrollbar.
- **`GET /api/v1/me`** now returns `email`, `name`, `plan` and `created_at` in addition to `user_id`. Needed by the new profile panel; no breaking change since the response was previously `{user_id}`.
- **Analytics, Webhooks and Audit log sections** check `licenseStore.allowsPro` before fetching. Free users see the paywall immediately instead of waiting for a 402 / 404 round-trip and falling through to a generic "Couldn't load" error.

### Fixed

- **Broadcasts marked `interrupted` on restart even though they actually resumed sending (#41 follow-up).** `MarkInProgressBroadcastsInterrupted` flipped every `'sending'` broadcast to `'interrupted'` at startup, but the worker's `ResetStuckSendingJobs` was simultaneously bringing the per-recipient jobs back to `'retry'` — so sending continued in the background while the UI showed a misleading "did not receive the email, re-run the broadcast" warning and Pro Analytics lost visibility because `broadcasts_in_flight` filters on `status='sending'`. Replaced with `ReconcileStuckBroadcasts`, which only marks `'completed'` the broadcasts whose jobs have all settled. In-flight broadcasts stay `'sending'`, the worker keeps draining them, and the UI/Pro Analytics see the correct state.

### API

- `GET /api/v1/me` — now returns `email`, `name`, `plan`, `created_at` alongside `user_id`.
- `POST /api/v1/me/password` — change the authenticated user's password. Body: `{current_password, new_password}`. Verifies current via bcrypt, validates new against the registration password rules, then updates `users.password_hash`.
- `GET /api/v1/update/auto-status` — returns Watchtower availability: `{configured, healthy, url, last_check, last_error}`. Result is cached for 30 seconds.
- `POST /api/v1/update/trigger` — fires the Watchtower `/v1/update` call in a goroutine and returns `202` immediately. Returns `400` if Watchtower isn't configured, `502` if it's unreachable.
- `GET /api/v1/projects/{id}/logs/{logId}` — returns `{log, clicks}` for the email log detail drawer.

### Environment

- New optional `SENDDOCK_WATCHTOWER_URL` and `SENDDOCK_WATCHTOWER_TOKEN` — enable the dashboard's one-click update flow. When unset, the update modal falls back to the manual command.

### Migrations

- `20260528193111_add_email_logs_search_index.sql` — enables `pg_trgm` if missing and adds gin indexes on `email_logs.to_email` and `email_logs.subject`. Idempotent, no data changes.

## [0.6.4.1] — 2026-05-08

Hotfix on top of [0.6.4](#064--2026-05-08). No new features, no DB migrations — drop-in replacement for any 0.6.x deployment showing the "restart every ~5 hours" pattern.

### Fixed

- **Container restarted every ~5 hours after boot due to its own healthcheck being rate-limited.** `Redis.Increment` reset the key's TTL on every call (`INCR` followed unconditionally by `EXPIRE`). The Dockerfile healthcheck (`wget /health` every 30s) kept the `rl:::1` counter's key alive forever instead of expiring per-minute, so the counter accumulated cumulatively. After ~5 hours (600 hits ÷ 2/min), it crossed the 600 req/min threshold and `/health` started returning `429`. wget's `-q … || exit 1` turned that into exit 1; swarm killed the "unhealthy" container after 5 retries; the new container started fresh and the cycle repeated. Increment now uses an atomic Lua script that only sets the TTL when the key is created (`count == 1`), giving a real fixed-window counter.
- **`/health` was wrapped by the rate limiter, CORS, and body-size middlewares.** Healthchecks must never be subject to client-facing throttles. `/health` is now mounted on a separate root mux that bypasses all middleware; everything else still goes through the full pipeline. Defense-in-depth alongside the Increment fix above.

### Added

- **`RATE_LIMIT_PER_MINUTE` env var (default `600`).** Used to be hard-coded. Lower it on small deployments or raise it for high-traffic apps without rebuilding the image.

## [0.6.4] — 2026-05-08

> **⚠️ Deprecated — upgrade to [0.6.4.1](#0641--2026-05-08).** Affected by a Redis rate-limiter bug that crashes the container in a restart loop ~5 hours after every boot. The fix lives in 0.6.4.1 with no migrations or breaking changes.



### Added

- **Per-recipient broadcast queue with retries (#41).** Broadcasts and scheduled campaigns no longer run in a single goroutine that loses everything on restart. Each recipient is enqueued as a row in a new `broadcast_jobs` table; five worker goroutines drain it concurrently using `SELECT … FOR UPDATE SKIP LOCKED`. Transient SMTP errors (4xx, network, DNS) reschedule the job with exponential backoff (30s, 2m, 8m, 30m, 1h), capping at five attempts before the recipient is marked `failed`. SMTP 5xx bounces are not retried — they are tagged `bounced` and the recipient is added to the suppression list. A backend restart mid-broadcast resets jobs that were in `sending` to `retry` so sending resumes from where it left off, with no recipient sent to twice and none silently dropped.
- **Live progress for newsletters and broadcasts.** The Newsletters list shows `sent_count` / `failed_count` updating every five seconds while a campaign is in `sending`, instead of jumping from `0/0` to the final number at the end. Same goes for the Broadcasts tab progress bars.
- **"Broadcasts in flight" panel on Pro Analytics (#41).** When at least one broadcast is actively sending, a card appears at the top of the Analytics dashboard with a live progress bar per broadcast, elapsed time, and per-status counters. Polls every five seconds and disappears as soon as the queue drains.
- **Email Logs filters and CSV export (#44).** Status filter is now a row of clickable chips (Sent / Failed / Bounced / Suppressed) with status-tinted highlight. Added a Template dropdown to filter logs to a single template's sends. New `GET /api/v1/projects/{id}/logs/export.csv` endpoint and **Export CSV** button (top right of Logs) downloads every row matching the active filters — no `limit`/`offset`. Columns: id, recipient, subject, status, error, sent_at, opened_at, clicked_at, template_id, subscriber_id.
- **Dark color-scheme on native date pickers.** All `<input type="date">` widgets across the app now render their popup in dark mode in Chromium and Firefox, matching the rest of the UI.

### Fixed

- **`/health` no longer kills the container on transient Postgres latency.** v0.6.3 introduced a synchronous 2-second `PingContext` on every `/health` call. When Postgres took >2s for any reason (slow disk, GC pause, momentary IO contention shared with neighbors on a small VPS), healthcheck failed five times in a row over ~3 minutes and the orchestrator killed the container — only to start a new one which had the same problem. The endpoint is now backed by a background goroutine that pings Postgres every 10s with a 5s timeout and stores the last-success Unix timestamp atomically. `/health` reads the atomic and returns `503` only if no successful ping has happened in the last 60 seconds. Single blips no longer trigger crash-replace loops; sustained Postgres outages still surface to the orchestrator.
- **Newsletter shown as `sent` with `213/0` immediately on schedule.** With the new queue, `EmailService.Broadcast` returns in milliseconds (only enqueues), so the campaign worker — which used to call `Broadcast` and immediately mark the campaign as `sent` with the upfront recipient count — was claiming the campaign was finished before any recipient had been sent to. Campaigns now stay in `sending`, linked to their broadcast via a new `campaigns.broadcast_id` FK; when the broadcast worker drains the queue it cascades the real `sent_count` / `failed_count` back to the linked campaign and flips its status to `sent` with `sent_at` set.

### Documentation

- **New Docker Swarm / Dokploy troubleshooting section.** Documents the recurring "all services on the same host stop talking to each other" pattern — usually caused by accumulated dead containers and orphan services confusing the embedded DNS resolver `127.0.0.11:53`, which makes every service on the overlay network fail to resolve names simultaneously. Includes diagnostic commands and a safe cleanup sequence (`docker service rm` orphans, `docker container prune`, `docker network prune`) plus an explicit warning never to use `volume prune` or `system prune -a`.
- **Tag-cache `Image is up to date` workaround.** When Dokploy or Docker Swarm refuses to pull a new image because the tag (`:dev`, `:latest`) appears unchanged locally, enable Clean Cache in the UI or use `docker service update --force --image …` from the host.

## [0.6.3] — 2026-05-06

> **⚠️ Deprecated — upgrade to [0.6.4.1](#0641--2026-05-08).** Same restart-loop bug as 0.6.4 (rate limiter eventually 429-ing the container's own healthcheck), plus its own crash trigger from the synchronous DB ping introduced here. Both are fixed in 0.6.4.1.



### Fixed

- **Container reported `healthy` even when Postgres was zombie.** The `/health` endpoint returned `{"status":"ok"}` without touching the database. When Postgres entered a stuck state (idle-in-transaction backlog, OOM-killed worker, network partition), the Go binary's connection pool kept the old TCP sockets open and never noticed. Docker / Swarm therefore kept routing traffic to the container; the reverse proxy hit those queries, they hung, and Traefik returned `502 Bad Gateway` to the user. Restarting *just* the app didn't fix it because Postgres itself was the broken side; only restarting Postgres recovered the stack. `/health` now performs a `PingContext` against the database with a 2-second timeout, returning `503 db_unreachable` when the round-trip fails. The orchestrator can now react.
- **Database connection pool was unbounded.** The default `sql.DB` pool has no `MaxOpenConns` / `MaxIdleConns` / `ConnMaxLifetime`. Under traffic, idle connections accumulated indefinitely, and every connection that became zombie (because the Postgres side died) stayed in the pool forever — feeding directly into the symptom above. The pool is now capped at **25 open** / **5 idle** connections with **5-minute lifetime** and **2-minute idle timeout**, so dead sockets are recycled within minutes even if the unhealthy `/health` somehow doesn't trigger first.

## [0.6.2] — 2026-05-06

### Fixed

- **Role assignment accepted only `owner` and `member`.** The `normalizeRole` helper used by `AddMember`, `CreateUser` and `UpdateMember` predated the v0.6 role expansion and silently rejected `admin`, `developer` and `viewer` with `400 invalid role`. Capability enforcement (which gate each endpoint by role) was already aware of all five roles, so the bug was strictly on the assignment path. This shipped as part of the public Team launch but went unnoticed because the dashboard's default invite role is `developer` and most launch testing exercised owner/member rebalancing rather than fresh assignments. Two consequences are now closed:
  - `POST /api/v1/workspaces/{id}/members` and `POST /api/v1/workspaces/{id}/users` accept all five roles per the docs.
  - `PATCH /api/v1/workspaces/{id}/members/{userId}` can now move someone between roles other than just `owner ↔ member`.

  Existing memberships are unchanged. Owners on Team workspaces who tried to invite a developer/admin/viewer and saw `400` should retry now that 0.6.2 is live.

### Compose / image

- Same v0.6.1 healthcheck-on-IPv6-only-hosts class of bug, but in `docker-compose.image.yml` instead of the image. The compose-level healthcheck overrides the image's, so self-hosters using Option 1 of the install guide were still hitting the crash-replace loop on hosts with `net.ipv6.bindv6only=1` even after pulling the v0.6.1 image. Updated the compose file to target `localhost` and bumped the start grace.
- `SENDDOCK_LICENSE_ENDPOINT` no longer ships with a default that points at a non-existent host (`license.senddock.com`). The binary's own default (`https://api.lemonsqueezy.com/v1`) takes over when the env var is empty; deployments that need a custom endpoint can still set it explicitly.

## [0.6.1] — 2026-05-06

### Fixed

- **Healthcheck reliability on IPv6-only hosts.** On hosts with `net.ipv6.bindv6only=1` (some VPS, Docker Swarm clusters and cloud providers) the Go server's default `:8080` listener would bind IPv6-only, while the container `HEALTHCHECK` targeted `127.0.0.1:8080`. Every healthcheck failed, the container was marked `unhealthy` after three retries, the orchestrator killed the task and started a new one — a 60–90s crash-replace loop with no panic, no error log, just intermittent `Bad Gateway` from the reverse proxy. Two changes ship in 0.6.1:
  - The HTTP listener binds `0.0.0.0:8080` explicitly, guaranteeing IPv4 availability regardless of host sysctls.
  - The container `HEALTHCHECK` now uses `localhost` (resolves to both `127.0.0.1` and `::1` via `/etc/hosts`) and the start-period grace was raised to `30s` with `5` retries so a slow first-boot DB doesn't tip the container into a loop.

## [0.6.0] — 2026-04-30

The "team launch" release. Workspaces with members and roles, a new Team plan above Pro, an in-app subscribe path through Lemon Squeezy, plus a long list of v0.6 quality-of-life features.

### Added

- **Workspaces.** A workspace is a container for projects with its own member list. Authorization for project endpoints now scopes by workspace membership; existing single-user installs are migrated transparently (every user got a `My Workspace` with their projects under it). Workspace CRUD, listing your own workspaces and listing members stay free in Core.
- **Roles & capabilities.** Four roles — `owner`, `admin`, `developer`, `viewer` — with a fixed capability matrix enforced at the handler level. Developers can `POST /send` (transactional) but not broadcast or edit templates; viewers are read-only; admins do everything except member management; owners do everything. The legacy `member` role is kept for backward compatibility (same access as `admin`).
- **Team plan tier.** Multi-member workspaces, role management, and the new admin "Create user" endpoint (`POST /workspaces/{id}/users`) belong to a new **Team** plan above Pro. Without a Team-entitled `SENDDOCK_LICENSE_KEY` the endpoints return `402 Payment Required` and the Members page renders a paywall — the single-user flow stays free.
- **In-app Subscribe links.** Pro and Team paywalls now drive directly to the Lemon Squeezy checkout for the right tier instead of bouncing the user back to the landing page. Pro's CTA pre-applies the `LAUNCH3FREE` launch coupon while it is active.
- **Pro vs Team gating in the validator.** The Lemon Squeezy validator now classifies licenses by `variant_id` and returns the right tier in `Status.Tier`. `AllowsFeature(feature)` consults a feature → tier matrix so a Pro key cannot unlock Team features.
- **Email validation on import (#43).** CSV/JSON imports now check syntax, MX records and a built-in disposable-domain list. Rejected rows surface in the import results modal with a per-row reason.
- **File picker and drag-and-drop for CSV imports.** Replaces the textarea-only flow.
- **Per-project suppression list (#42).** New `suppressions` table; `/send`, `/send/batch` and `/broadcast` skip suppressed recipients and account for them in the result counts. Existing unsubscribed subscribers are backfilled. Manage entries from the Suppressions tab inside a project (list, filter by reason, manual add, bulk import, remove).
- **Per-project audit log (#47, Pro).** Records 12 sensitive actions (project create/delete, SMTP/IMAP/bounce updates, API key create/revoke, webhook create/delete, suppression add/remove, login). Exposed at `GET /audit-log` and surfaced in the Audit Log tab. Pro-gated.
- **Bounce ingestion (#38).**
  - **Part A — in-session 5xx detection.** RCPT TO failures with 5xx are classified, marked `bounced` in the email log, added to the suppression list and dispatched as `email.bounced` webhooks.
  - **Part B — public webhook endpoint.** `POST /webhooks/bounces/{projectId}?token=<bounce-token>` accepts generic JSON or Mailgun event payloads. Each project gets a rotatable token; the URL is shown in the Settings tab.
  - **Part C — IMAP poller.** Configure a bounce mailbox per project; a 5-minute poller logs in over TLS, scans `INBOX` for unread DSNs, extracts recipients (RFC 3464 `Final-Recipient` first, then 5xx-line fallback), adds them to the suppression list and marks the messages `\Seen`.
- **Standardized UI primitives.** `AppCheckbox` and `AppConfirmModal` replace five inconsistent checkbox styles and two native `confirm()` calls.

### Changed

- Suppressed sends are now a distinct outcome end-to-end: the `bounced` and `suppressed` columns are tracked separately from `sent` and `failed` in stats, the broadcast result, and the dashboard.
- The import results modal grew to `size="lg"` so the five outcome cards and the rejected-rows table fit without overflowing.

### Fixed

- `AppCheckbox` is now clickable outside `<label>` wrappers (the off-screen `peer sr-only` input did not catch clicks inside table cells).
- Rate-limit middleware no longer poisons the session-expiry flow. The global per-IP cap is 600/min, `X-Forwarded-For` is parsed correctly behind a proxy, and the frontend distinguishes 429 from 401 so a burst of polling no longer dumps the user back to the login screen.
- Project shell is responsive on mobile — sidebar collapses into a drawer below `md`, page-level headers wrap, every table sits inside `overflow-x-auto`.
- `GET /workspaces/{id}/projects` now wraps the response with the same shape as `GET /projects` (was returning raw Go field names, broke `project.created_at` and `project.name` on the dashboard).

## [0.5.2] — 2026-04-30

### Security

- **Critical: license bypass closed.** Versions 0.5.0 and 0.5.1 auto-unlocked every Pro feature (Analytics, Webhooks UI, Audit log) on any self-hosted deployment with an empty `SENDDOCK_LICENSE_KEY`. The validator now returns `LockedFree` for an empty key in every deployment mode. Core stays fully usable without a license; Pro endpoints return `402 Payment Required` until a valid key is set.

### Changed

- Documentation swept (README.md, `docs/self-hosting/configuration.md`, `docs/self-hosting/installation.md`, `docs/guide/environment.md`, `docs/guide/analytics.md`, `docs/guide/webhooks.md`, `docs/api/webhooks.md`) to remove the obsolete "self-hosted unlocks Pro locally" wording. The behaviour matrix now reads: empty key → Pro locked, valid key → Pro unlocked, regardless of `DEPLOYMENT_MODE`.
- Releases v0.5.0 and v0.5.1 marked as deprecated on GitHub with banners pointing to v0.5.2.

## [0.5.1] — 2026-04-29

> Deprecated — contains the v0.5.0 license bypass. Upgrade to 0.5.2.

### Fixed

- Hide the update modal in cloud deployments. `/version` returns `enabled: false` when `DEPLOYMENT_MODE=cloud`, so managed-product users no longer see a self-host upgrade prompt.
- Render release notes as Markdown instead of raw text inside a `<pre>` block.
- Default upgrade command points to `docker compose pull && docker compose up -d` (was `git pull && ./setup.sh`, source-build only).
- Pricing links resolve again — three references to the non-existent `senddock.com/pricing` switched to `senddock.dev/pricing`.

### Added

- The current-version label is always clickable, opening the same modal with the upgrade section hidden — useful for revisiting release notes after the fact.

## [0.5.0] — 2026-04-29

> Deprecated — contains a license bypass. Upgrade to 0.5.2.

### Added

- **Webhooks (Core + Pro).** HMAC-SHA256 signed dispatcher with seven event types (`email.sent`, `email.failed`, `email.bounced`, `email.opened`, `email.clicked`, `subscriber.created`, `subscriber.unsubscribed`), exponential backoff retries (30s → 2m → 10m → 30m → 2h, 5 attempts), full management UI in Pro.
- **Click tracking (Core).** Every `<a href>` in outgoing emails routes through `/c/{logId}/{payload}` with HMAC-protected tokens. First click stamps `clicked_at`; full URL hits land in `email_clicks`.
- **Pro Analytics dashboard.** Funnel (sent → delivered → opened → clicked), opens-over-time chart with adaptive bucket granularity, top templates, top clicked links, auto-generated insights, date presets and custom range, trend pills against the previous equivalent window.
- **One-click unsubscribe (RFC 8058).** Confirmation page on `GET /unsubscribe/{id}/{subscriberId}` and instant `POST` handling so Gmail/Outlook native buttons work without a click-through. Per-recipient HMAC-signed URLs.
- **Pro license validator.** `SENDDOCK_LICENSE_KEY` validated against Lemon Squeezy on startup and re-checked periodically.
- **Container registry split.** `ghcr.io/arkhe-systems/senddock` exclusively carries tagged releases (`:0.5.0`, `:0.5`, `:0`, `:latest`) and stays public. Pre-release `:dev` builds moved to a separate private package.
- **Per-project rate limits** on `/send`, `/send/batch` and `/broadcast` enforced via Redis.
- **SMTP test endpoint** with bounded 5s connect / 10s session timeouts.
- **`AppButton`** gains `size="sm"` and `variant="ghost"` for consistent modal action rows.
- **Docs overhaul** with hand-crafted monochromatic SVG diagrams that follow the VitePress theme toggle via `currentColor`.

## [0.4.0] — 2026-04-15

Earlier work and the foundation for the open-core release. See git history for the full list. Highlights:

- BYO SMTP per project with encrypted password storage.
- Subscribers, templates, sends, broadcasts, campaigns.
- API keys with hashed secrets and per-key rate limits.
- Open-tracking pixel and unsubscribe links.
- Initial dashboard, project switcher, settings UI.

[Unreleased]: https://github.com/arkhe-systems/senddock/compare/v0.6.3...HEAD
[0.6.3]: https://github.com/arkhe-systems/senddock/compare/v0.6.2...v0.6.3
[0.6.2]: https://github.com/arkhe-systems/senddock/compare/v0.6.1...v0.6.2
[0.6.1]: https://github.com/arkhe-systems/senddock/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/arkhe-systems/senddock/compare/v0.5.2...v0.6.0
[0.5.2]: https://github.com/arkhe-systems/senddock/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/arkhe-systems/senddock/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/arkhe-systems/senddock/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/arkhe-systems/senddock/releases/tag/v0.4.0
