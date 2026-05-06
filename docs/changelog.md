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

## [0.6.3] — 2026-05-06

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

- **Webhooks (Core + Pro).** HMAC-SHA256 signed dispatcher with six event types, exponential backoff retries (30s → 2m → 10m → 30m → 2h, 5 attempts), full management UI in Pro.
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
