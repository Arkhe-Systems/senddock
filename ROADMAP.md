# SendDock Roadmap

## Phase 1: Core Backend
- [x] Project structure (Go modules, internal/, cmd/)
- [x] HTTP server with net/http stdlib
- [x] Environment config loading
- [x] PostgreSQL connection with database/sql
- [x] Docker Compose (PostgreSQL + Redis)
- [x] Database migrations with goose (users, projects)
- [x] sqlc setup and code generation
- [x] User registration with bcrypt password hashing
- [x] User login with JWT token generation
- [x] Health check endpoint
- [x] Auth middleware (JWT verification, extract user from token)
- [x] Protected routes (require authentication)
- [x] CRUD projects (create, list, get, delete)
- [x] CORS middleware (dynamic origin from config, credentials support)
- [x] Makefile with dev commands
- [x] Refresh token rotation (HttpOnly cookies, SHA-256 hashing)
- [x] Logout with token invalidation
- [ ] Validate project limits per plan (cloud mode)

## Phase 2: Frontend Foundation
- [x] Vue 3 + TypeScript + Vite project setup
- [x] Tailwind CSS 4 configuration
- [x] API client (fetch wrapper)
- [x] Auth store (Pinia)
- [ ] Update auth store to work with HttpOnly cookies (remove localStorage)
- [x] Vue Router with auth guards
- [x] Reusable UI components (AppInput, AppButton, AppAlert, AppCard)
- [x] Login page
- [x] Register page
- [x] Dashboard page (basic)
- [x] Logout functionality
- [x] Auth redirect with reason messages
- [x] Dashboard layout (sidebar, header)
- [x] Project list in dashboard
- [x] Create project modal/page
- [x] Project detail page

## Phase 3: Subscribers & Templates
- [x] Subscribers table + migration
- [x] CRUD subscribers (per project)
- [x] Bulk import subscribers (CSV/JSON)
- [x] Subscriber segmentation (active, pending, unsubscribed)
- [x] Templates table + migration
- [x] CRUD templates (per project)
- [x] Handlebars template rendering with dynamic variables
- [x] Subscriber management UI
- [x] Template editor UI

## Phase 4: Email Sending
- [x] SMTP configuration per project
- [x] SMTP password encryption/decryption
- [x] Transactional email endpoint (POST /api/v1/send)
- [x] Email worker with asynq (Redis-based job queue)
- [x] Broadcast endpoint (send to all subscribers)
- [ ] Email validation before sending
- [ ] Campaign builder UI (scheduled sends)

## Phase 5: Email Verification & Security
- [ ] Email verification on registration (send code)
- [ ] Verification page
- [ ] Onboarding flow (additional user info for cloud)
- [ ] Password reset flow
- [ ] Session expiration (7 days inactivity)
- [ ] Account lockout after failed attempts

## Phase 6: Tracking & Analytics
- [x] Open tracking (pixel injection) — Core
- [x] Click tracking (link rewriting + HMAC redirect) — Core
- [x] Unsubscribe handling (one-click + confirmation page, RFC 8058)
- [x] Analytics table + migration
- [x] Analytics endpoints (sent, failed, opened, clicked) — PRO
- [x] Logs table + migration
- [x] System logs endpoint
- [x] Analytics dashboard UI with charts — PRO

## Phase 7: API Keys & Security
- [x] API keys table + migration
- [x] API key generation (public pk_ / secret sdk_)
- [x] API key authentication middleware
- [x] Per-project rate limiting on `/send`, `/send/batch` and `/broadcast` (Redis-backed)
- [ ] Request logging
- [x] API keys management UI

## Phase 8: Payments & Plans (Cloud mode)
- [ ] Lemon Squeezy webhook handler
- [ ] Subscription lifecycle (created, updated, cancelled, expired)
- [ ] Plan upgrade/downgrade logic
- [ ] Coupon/discount support (handled by Lemon Squeezy)
- [ ] Monthly usage reset cron job
- [x] Deployment mode config (cloud vs self-hosted)
- [x] Feature gating based on plan (license validator + paywall states in UI)
- [x] License key activation/validation against Lemon Squeezy
- [ ] Billing page UI

## Phase 9: Webhooks
- [x] Webhook configuration per project — Pro UI/API, Core dispatcher
- [x] Webhook dispatcher (FOR UPDATE SKIP LOCKED, batch claim) — Core
- [x] Webhook retry logic with exponential backoff (30s → 2h, 5 attempts) — Core
- [x] Webhook signature verification (HMAC-SHA256, `X-SendDock-Signature: t=<ts>,v1=<hex>`) — Core
- [x] Webhook event types (email.sent/failed/opened/clicked, subscriber.created/unsubscribed) — Core
- [x] Webhooks management UI — PRO
- [x] Per-webhook deliveries view — PRO

## Phase 13: Deployment & Self-hosting
- [x] Production Dockerfile (multi-stage: Go build + Vue build)
- [x] Docker Compose for self-hosting (app + postgres + redis)
- [x] Go serves Vue static files (single binary/container)
- [x] Health check endpoints for container orchestration
- [ ] Graceful shutdown handling
- [ ] Environment configuration documentation

---

## Upcoming milestones

Phases 1–9 above document the foundation that shipped. The list below groups upcoming work into release-themed milestones rather than fixed version numbers — specific version numbers and dates are decided at release time as scope settles. 1.0 is reserved for when the platform is genuinely polished end-to-end; it is not the next release. Milestones ship in the order listed, but each one may end up as several smaller patch releases depending on how the work breaks down.

### Milestone: Marketing-ready

Typed subscriber data + the first piece of segmentation, so SendDock stops being "send to everyone" and starts being a real marketing tool.

- [ ] Tags + segmentation — [#40](https://github.com/Arkhe-Systems/senddock/issues/40) — Core
- [ ] Custom fields for subscribers — [#72](https://github.com/Arkhe-Systems/senddock/issues/72) — Core
- [ ] Starter template library — [#73](https://github.com/Arkhe-Systems/senddock/issues/73) — Core

### Milestone: Acquisition

Lower the friction of bringing users and existing subscriber lists into the platform.

- [ ] Google OAuth sign-in — [#70](https://github.com/Arkhe-Systems/senddock/issues/70) — Core
- [ ] Migration wizard (Mailchimp / EmailOctopus / ConvertKit) — [#74](https://github.com/Arkhe-Systems/senddock/issues/74) — Core
- [ ] Embeddable signup form widget — [#75](https://github.com/Arkhe-Systems/senddock/issues/75) — Core

### Milestone: Trust + automation foundation

What teams evaluate before committing their list. Plus the first phase of drips — linear sequences cover most automation demand on their own.

- [ ] Drip automations · Phase 1: linear sequences — [#77](https://github.com/Arkhe-Systems/senddock/issues/77) — Pro
- [ ] DNS deliverability check (SPF / DKIM / DMARC) — [#64](https://github.com/Arkhe-Systems/senddock/issues/64) — Pro
- [ ] GDPR / data subject tools — [#46](https://github.com/Arkhe-Systems/senddock/issues/46) — Core

### Milestone: Automation parity

Closes the marketing-automation gap with EmailOctopus and Mailchimp.

- [ ] Drip automations · Phase 2: branches, all triggers, goal tracking — [#77](https://github.com/Arkhe-Systems/senddock/issues/77) — Pro
- [ ] Polish + bug bash across previously shipped milestones

### Backlog (no committed milestone)

Picked up as demand surfaces or as bandwidth allows.

- [ ] `senddock-cli` (admin & ops CLI) — [#76](https://github.com/Arkhe-Systems/senddock/issues/76) — Core
- [ ] Multi-distro support for the one-line installer (Debian, Fedora, RHEL, openSUSE) — [#79](https://github.com/Arkhe-Systems/senddock/issues/79) — Core. Arch shipped in v0.6.8 ([#81](https://github.com/Arkhe-Systems/senddock/issues/81)); macOS tracked separately ([#82](https://github.com/Arkhe-Systems/senddock/issues/82)).
- [ ] Custom tracking domain / white-label — [#45](https://github.com/Arkhe-Systems/senddock/issues/45) — Pro / Enterprise
- [ ] Send-time heatmap and recommendation — [#51](https://github.com/Arkhe-Systems/senddock/issues/51) — Pro
- [ ] Email approval workflow for broadcasts — [#59](https://github.com/Arkhe-Systems/senddock/issues/59) — Team

### Deferred (no scheduled milestone)

Captured for visibility, not actively scoped. Will be promoted when there is a concrete customer ask.

- [ ] Multi-SMTP failover (was Phase 11) — Pro
- [ ] Admin panel for cloud operators (was Phase 12)
- [ ] SSO / LDAP integration (was Phase 14) — Enterprise
- [ ] Internationalization (English + Spanish) (was Phase 14)
- [ ] Resend campaign to non-openers — [#78](https://github.com/Arkhe-Systems/senddock/issues/78) — Pro

---

Items marked **Pro** or **Team** live in the private `senddock-pro` repository and are gated by `SENDDOCK_LICENSE_KEY`. They are not part of the AGPL Community edition.
