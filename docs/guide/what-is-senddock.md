# What is SendDock

SendDock is an open-source, BYOSMTP (Bring Your Own SMTP) email marketing platform for developers and businesses that want full control over their email infrastructure.

## Key Principles

- **Self-hostable** — Install on your own server, keep your data private
- **API-first** — Every feature works through the REST API
- **No vendor lock-in** — Export your data, migrate freely, fork the code
- **Open core** — Community edition is free and fully functional

## Architecture

SendDock is a monorepo with three components:

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Backend | Go (stdlib net/http) | REST API, auth, business logic |
| Frontend | Vue 3 + TypeScript | Dashboard, template editor, settings |
| Database | PostgreSQL | Data storage |

### How it works

1. Create a **project** (isolated workspace)
2. Configure **SMTP** (your email server credentials)
3. Add **subscribers** (manually or via API)
4. Build **templates** (visual editor or HTML code)
5. **Send emails** — to individuals, broadcast to all, or direct transactional

## Community vs Pro

| Feature | Community | Pro |
|---------|-----------|-----|
| Projects | Unlimited | Unlimited |
| Subscribers | Unlimited | Unlimited |
| Email sending | Unlimited | Unlimited |
| Template builder | Code + Visual | Code + Visual |
| API keys | Yes | Yes |
| Email logs | Yes | Yes |
| Open tracking | Yes | Yes |
| Scheduled campaigns | Yes | Yes |
| SMTP per project | 1 | Multiple + failover |
| Team members | 1 admin | Unlimited + roles |
| Analytics | Sent/Failed | Opens, clicks, bounces |
| Webhooks | No | Yes |
| SSO/LDAP | No | Yes |
| White-label | No | Yes |

## License

AGPL-3.0. Free to use and self-host. If you modify SendDock and offer it as a hosted service, you must open-source your modifications.
