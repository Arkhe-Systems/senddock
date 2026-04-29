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

## Phase 10: Team Members — PRO
- [ ] Team members table + migration
- [ ] Invite system
- [ ] Role-based access (owner, admin, member, viewer)
- [ ] Team member management endpoints
- [ ] Team management UI

## Phase 11: Advanced SMTP — PRO
- [ ] Multi-SMTP configuration per project
- [ ] Automatic failover between SMTP providers
- [ ] SMTP health checking
- [ ] Weighted SMTP routing

## Phase 12: Admin Panel
- [ ] Admin users table + migration
- [ ] Admin authentication (separate from user auth)
- [ ] User management (list, view, ban/unban)
- [ ] Financial dashboard (subscriptions, revenue)
- [ ] System health monitoring
- [ ] Activation codes for plan upgrades

## Phase 13: Deployment & Self-hosting
- [x] Production Dockerfile (multi-stage: Go build + Vue build)
- [x] Docker Compose for self-hosting (app + postgres + redis)
- [x] Go serves Vue static files (single binary/container)
- [ ] Environment configuration documentation
- [x] Health check endpoints for container orchestration
- [ ] Graceful shutdown handling
- [ ] CLI tool for admin tasks (create admin, reset password)

## Phase 14: Enterprise Features — PRO
- [ ] SSO / LDAP integration
- [ ] Audit logs
- [ ] White-label (remove SendDock branding)
- [ ] Custom domain support
- [ ] Data export tools
- [ ] SLA monitoring
- [ ] Internationalization (English + Spanish)

---

Items marked **PRO** live in the private `senddock-pro` repository and are not included in the community edition.
