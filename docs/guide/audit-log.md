# Audit log <Badge type="warning" text="Pro" />

The audit log records every sensitive action taken on a project — who did it, when, against what, and from which IP. Available on Pro and Team. Without a Pro license the endpoint and UI tab are hidden, and the underlying actions still happen, they just aren't recorded.

## What gets recorded

| Category | Actions |
|---|---|
| Project lifecycle | `project.create`, `project.update`, `project.delete` |
| SMTP & deliverability config | `smtp.update`, `bounce_imap.update`, `bounce_token.rotate` |
| API keys | `api_key.create`, `api_key.revoke` |
| Webhooks | `webhook.create`, `webhook.delete` |
| Suppressions | `suppression.add`, `suppression.delete` |
| Workspace (Team) | `workspace.create`, `workspace.rename`, `workspace.delete`, `workspace.member_added`, `workspace.member_removed`, `workspace.member_role_changed`, `workspace.user_created` |
| Sends | `broadcast.send` |

Each entry stores:

- **`created_at`** — UTC, ISO 8601.
- **`user_id`** — UUID of the user who took the action. Resolve to email via [`GET /workspaces/{id}/members`](../api/workspaces#list-members) when displaying.
- **`action`** — the string from the table above.
- **`target_type`** + **`target_id`** — the entity affected (`project`, `api_key`, `webhook`, `suppression`, `workspace`).
- **`metadata`** — action-specific JSON (e.g. for `smtp.update` you get `smtp_host`, `smtp_user`, `from_email` — never the password).
- **`ip_address`** + **`user_agent`** — request context, IP parsed from `X-Forwarded-For` if present.

## Where to see it

In the project sidebar, open **Audit Log**. Filter by action, by actor, or by date range. Each entry expands to show the metadata JSON.

## API

```bash
curl -H "Authorization: Bearer sk_..." \
  "https://your-instance.com/api/v1/projects/{project_id}/audit-log?limit=50&action=smtp.update"
```

Query parameters:

- `limit` — max rows (default 50, max 200).
- `offset` — pagination offset (default 0).
- `action` — filter by exact action string.
- `from`, `to` — RFC 3339 timestamps to bound the window.

Response:

```json
{
  "entries": [
    {
      "id": "01H...",
      "user_id": "01H...",
      "action": "smtp.update",
      "target_type": "project",
      "target_id": "01H...",
      "metadata": {
        "smtp_host": "smtp.mailgun.org",
        "smtp_user": "postmaster@acme.com",
        "from_email": "hello@acme.com"
      },
      "ip_address": "203.0.113.42",
      "user_agent": "Mozilla/5.0 ...",
      "created_at": "2026-04-30T14:22:11Z"
    }
  ],
  "total": 1842
}
```

The endpoint is cookie-only — API keys can't read the audit log because the actor identity is mandatory and a project-scoped key has none. Resolve `user_id` to a display name / email via [`GET /workspaces/{id}/members`](../api/workspaces#list-members) when rendering the list.

## What it doesn't record

The audit log is for **administrative** actions, not data-plane traffic. It does **not** record:

- Individual `/send` calls (those are in the email log — millions of rows).
- Open or click events (those are in the analytics tables).
- Read operations (`GET /subscribers`, etc.).
- Login attempts (logged separately at the auth layer; surfaced in v0.7+).

If you need a full request log, use your reverse proxy's access log — Nginx, Caddy and Traefik all log per-request and are the right tool for that. The audit log is the curated subset of events worth showing in a "Who changed our SMTP password?" investigation.

## Retention

Entries are kept indefinitely by default. There's no built-in pruning — it's a small append-only table and on a typical install 12 months of changes is on the order of a few thousand rows. If you want a rolling window you can run a `DELETE FROM audit_log WHERE created_at < now() - interval '180 days'` from `psql`.

## See also

- [Webhooks](./webhooks) — webhook create/delete is in the audit log.
- [Suppressions](./suppressions) — suppression add/remove is in the audit log.
- [Workspaces](./workspaces) — workspace and member changes (Team) are in the audit log.
