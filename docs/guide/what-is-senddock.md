# What is SendDock

SendDock is an open-source, self-hosted email marketing and transactional platform built around **fully isolated projects**. Each project is its own sending identity — its own SMTP, sending domain, subscribers, templates, API keys and analytics — so one install runs every client, brand or app at once, with no data bleed between them. You bring your own SMTP provider (Mailgun, SES, Postmark, your VPS — anything that speaks SMTP), point SendDock at it, and run unlimited sends from your own infrastructure with no per-contact or per-email markup.

![A project dashboard in SendDock](/screenshots/hero.png)

## The model

- **Open core, AGPL-3.0.** The Community edition is fully usable without a license.
- **API-first.** Every dashboard action has a REST endpoint. Cookie auth for the UI, per-project API keys (`Authorization: Bearer sk_...`) for everything else.
- **Single-binary Go backend + Vue dashboard, deployed by Docker Compose.** One container for the app, Postgres for storage, Redis for rate limits.
- **Pro and Team are tier flags on the same binary.** Activate a license from the dashboard (**Instance → License**) and the validator unlocks the corresponding endpoints. No separate build.

## How you use it

1. Create a **workspace** (free for single-user; multi-member on Team).
2. Create a **project** — the unit of isolation. Each project has its own SMTP credentials, subscribers, templates, API keys, suppression list and (optionally) bounce mailbox.
3. Configure **SMTP** under the project's Settings.
4. Add **subscribers** manually, via the API, or by [importing CSV/JSON](./subscribers#import) with email validation.
5. Build **templates** in the visual editor or write raw HTML.
6. **Send** — `/send` for transactional, `/send/batch` for fan-out, `/broadcast` for full-list campaigns. Suppressed addresses and bounced recipients are skipped automatically.

## Community vs Pro vs Team

| Feature | Community | Pro | Team |
|---------|:---------:|:---:|:----:|
| Projects, subscribers, sends | Unlimited | Unlimited | Unlimited |
| BYO SMTP per project | ✓ | ✓ | ✓ |
| HTML templates with variables | ✓ | ✓ | ✓ |
| CSV/JSON import with email validation | ✓ | ✓ | ✓ |
| [Custom fields](./subscribers#custom-fields) + [tags](./subscribers#tags) | ✓ | ✓ | ✓ |
| [Segments](./segments) (saved filters, broadcast targeting) | ✓ | ✓ | ✓ |
| Per-project [suppression list](./suppressions) | ✓ | ✓ | ✓ |
| [Bounce ingestion](./bounces) (5xx + webhook + IMAP) | ✓ | ✓ | ✓ |
| API keys with per-key rate limits | ✓ | ✓ | ✓ |
| Open + click tracking | ✓ | ✓ | ✓ |
| One-click unsubscribe (RFC 8058) | ✓ | ✓ | ✓ |
| [Webhooks](./webhooks) — dispatcher + management UI & API | ✓ | ✓ | ✓ |
| Scheduled campaigns | ✓ | ✓ | ✓ |
| Single-user workspace | ✓ | ✓ | ✓ |
| Basic stats endpoint | ✓ | ✓ | ✓ |
| [Analytics dashboard](./analytics) (Overview, Campaigns, Audience, Engagement tabs) | ✓ | ✓ | ✓ |
| [Deliverability](./deliverability) (domain health, per-provider rates, spam rate) | — | ✓ | ✓ |
| [Report builder](./reports) (pivots, segment filters, saved reports, CSV) | — | ✓ | ✓ |
| [Audit log](./audit-log) | — | ✓ | ✓ |
| [Multi-member workspaces](./members) | — | — | ✓ |
| Roles: `owner`, `admin`, `developer`, `viewer` | — | — | ✓ |
| Admin "Create user" flow | — | — | ✓ |

### Pricing (self-hosted license)

| Plan | Monthly | Annual |
|------|---------|--------|
| Community | Free (AGPL-3.0) | — |
| Pro | $9 | $90 |
| Team | $29 | $290 |

On self-hosted you activate the license from the dashboard under **Instance → License** ([Instance settings](./instance-settings#pro-license)); it's stored in the database and validated against Lemon Squeezy, re-checked periodically. Without a key the binary stays in Community mode and Pro/Team endpoints return `402 Payment Required` until one is activated. See [Configuration](../self-hosting/configuration).

## License

The SendDock Core repository is **AGPL-3.0**. You can self-host, modify and redistribute. If you offer a modified version as a hosted service to third parties, you must publish your modifications. The Pro/Team feature set lives in a separate, non-public repository and is unlocked by a license key — it is not subject to AGPL.
