# Changelog

All notable changes to SendDock are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Releases are also published on [GitHub](https://github.com/arkhe-systems/senddock/releases) and the `ghcr.io/arkhe-systems/senddock` image carries the matching tag.

## [Unreleased]

Targeted for the **0.6.0** release. Currently on the `dev` branch.

### Added

- **Workspaces.** A workspace is a container for projects with its own member list. Authorization for project endpoints now scopes by workspace membership; existing single-user installs are migrated transparently (every user got a `My Workspace` with their projects under it). Workspace CRUD, listing your own workspaces and listing members stay free in Core.
- **Roles & capabilities.** Four roles — `owner`, `admin`, `developer`, `viewer` — with a fixed capability matrix enforced at the handler level. Developers can `POST /send` (transactional) but not broadcast or edit templates; viewers are read-only; admins do everything except member management; owners do everything. The legacy `member` role is kept for backward compatibility (same access as `admin`).
- **Team plan tier.** Multi-member workspaces, role management, and the new admin "Create user" endpoint (`POST /workspaces/{id}/users`) belong to a new **Team** plan above Pro. Without a Team-entitled `SENDDOCK_LICENSE_KEY` the endpoints return `402 Payment Required` and the Members page renders a paywall — the single-user flow stays free.
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

[Unreleased]: https://github.com/arkhe-systems/senddock/compare/v0.5.2...HEAD
[0.5.2]: https://github.com/arkhe-systems/senddock/compare/v0.5.1...v0.5.2
[0.5.1]: https://github.com/arkhe-systems/senddock/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/arkhe-systems/senddock/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/arkhe-systems/senddock/releases/tag/v0.4.0
