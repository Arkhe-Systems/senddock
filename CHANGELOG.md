# Changelog

All notable changes to SendDock are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases are also published on [GitHub](https://github.com/arkhe-systems/senddock/releases) and the `ghcr.io/arkhe-systems/senddock` image carries the matching tag.

## [Unreleased]

_Nothing here yet. Track upcoming work on the [open issues](https://github.com/arkhe-systems/senddock/issues)._

## [0.6.5] — 2026-05-08

### Fixed

- **Rate-limit counter never expired, eventually `429`-ing the container's own healthcheck and triggering a crash loop ~5 hours after every restart.** `Redis.Increment` reset the key's TTL on every call (`INCR` followed unconditionally by `EXPIRE`). Any IP that kept hitting the server — including `::1` from the Dockerfile's own `wget /health` healthcheck firing every 30s — kept the key alive forever, so the counter was effectively cumulative instead of per-window. After ~5 hours of healthchecks (600 hits ÷ 2/min), the counter crossed the 600 req/min threshold and `/health` started returning `429 Too Many Requests`. wget's `-q … || exit 1` turned that into exit 1, swarm killed the "unhealthy" container after 5 retries, the new container started fresh, and the cycle repeated. Increment now uses an atomic Lua script that only sets the TTL when the key is created (`count == 1`), giving a real fixed-window counter.
- **`/health` was wrapped by the rate limiter, CORS, and body-size middlewares.** Healthchecks must never be subject to client-facing throttles or origin checks. `/health` is now mounted on a separate root mux that bypasses all middleware; everything else still goes through the full pipeline. Defense-in-depth alongside the Increment fix above.

### Added

- **`RATE_LIMIT_PER_MINUTE` env var (default `600`).** Used to be hard-coded. Lower it on small deployments or raise it for high-traffic apps without rebuilding the image.

## [0.6.4] — 2026-05-08

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

- **Webhooks (Core + Pro).** HMAC-SHA256 signed dispatcher with six event types (`email.sent`, `email.failed`, `email.opened`, `email.clicked`, `subscriber.created`, `subscriber.unsubscribed`), exponential backoff retries (30s → 2m → 10m → 30m → 2h, 5 attempts), full management UI in Pro.
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
