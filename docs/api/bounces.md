# Bounces API

Configure how SendDock detects bounces for a project. See the [Bounces guide](../guide/bounces) for the conceptual model — these endpoints just expose the per-project knobs.

There are two configurable sources, plus a public ingest endpoint that providers post to:

- **IMAP poller** — SendDock logs into a bounce mailbox every 5 minutes and parses DSNs.
- **Webhook ingest** — SendDock receives a POST from your provider (Mailgun, SES via SNS, custom).

In-session SMTP detection (5xx on `RCPT TO`) needs no configuration and is always on.

All three detection sources write the same suppression reason `bounce`; the parsed SMTP/DSN text is stored in the entry's `source` field (e.g. `webhook ingest: 550 …`, `imap dsn poll`, or `smtp 550 during rcpt: …`).

All `/api/v1/...` endpoints on this page require **cookie auth** — bounce configuration is dashboard-managed, and project API keys cannot call it. The public ingest at `/webhooks/bounces/{projectId}` uses a per-project token instead and is the only endpoint here that is **not** under `/api/v1`.

## Get bounce IMAP config

```
GET /api/v1/projects/{id}/bounce-imap
```

### Response

```json
{
  "enabled": false,
  "host": "imap.mailgun.org",
  "port": 993,
  "user": "bounces@acme.com",
  "folder": "INBOX",
  "password_set": true
}
```

The password is never returned; `password_set` reports whether one is stored for this mailbox.

## Update bounce IMAP config

```
PUT /api/v1/projects/{id}/bounce-imap
```

### Request body

```json
{
  "enabled": true,
  "host": "imap.mailgun.org",
  "port": 993,
  "user": "bounces@acme.com",
  "password": "...",
  "folder": "INBOX"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `enabled` | bool | yes | When `false`, the poller stops scanning this mailbox. |
| `host` | string | yes if `enabled` | TLS-only — plaintext IMAP is rejected. |
| `port` | int | yes if `enabled` | Typically `993`. |
| `user` | string | yes if `enabled` | Mailbox account. |
| `password` | string | yes if `enabled` (or omit to keep current) | Encrypted at rest with the same key SMTP uses. |
| `folder` | string | no (default `INBOX`) | Folder name to scan. |

The poller logs in over TLS, scans for unread messages, parses `Final-Recipient` (RFC 3464) first, falls back to scanning for `5xx` lines, adds matched recipients to the [suppression list](./suppressions), and marks each message `\Seen` (never deletes).

### Response

`200 OK` with the new config (same shape as `GET`).

### Audit

Recorded as `bounce_imap.update`.

## Get bounce webhook config

```
GET /api/v1/projects/{id}/bounce-webhook
```

### Response

```json
{
  "project_id": "uuid",
  "bounce_token": "uuid",
  "path": "/webhooks/bounces/uuid?token=uuid"
}
```

`path` is the ingest path with the current token, relative to your SendDock origin — prepend your instance URL before pasting it into your provider's webhook settings. `bounce_token` is the token on its own.

## Rotate the bounce webhook token

```
POST /api/v1/projects/{id}/bounce-webhook/rotate
```

Generates a new token and invalidates the old one. The `path` returned by `GET /bounce-webhook` after this call carries the new token.

### Response

```json
{
  "project_id": "uuid",
  "bounce_token": "new-uuid",
  "path": "/webhooks/bounces/uuid?token=new-uuid"
}
```

### Audit

Recorded as `bounce_token.rotate`.

::: warning Rotate updates your provider too
Old URL stops working immediately. Update the webhook destination on your provider's side **before** rotating, or have the new URL queued so you can paste it in right after.
:::

## Public ingest endpoint

```
POST /webhooks/bounces/{projectId}?token=<bounce-token>
```

This is the URL you give to your email provider. It's not under `/api/v1` because it isn't authenticated by API key — the URL token is the credential. The `projectId` and `token` together identify the project.

### Request body — generic

A single object (or an array of objects) with at least `email` and optional `reason` / `type`. `reason` is free-text stored as the entry's `source`; if you send only `type` (e.g. `permanent`), it's used as that text. Every reported bounce is suppressed — SendDock does not distinguish permanent vs transient here, so only report hard bounces.

```json
{ "email": "user@example.com", "reason": "550 mailbox unavailable", "type": "permanent" }
```

### Request body — Mailgun event-data webhook

Send the verbatim payload Mailgun POSTs for `permanent_failure` (or `failed`) events. SendDock parses `event-data.recipient` and `event-data.delivery-status` automatically:

```json
{
  "event-data": {
    "event": "failed",
    "severity": "permanent",
    "recipient": "user@example.com",
    "reason": "bounce",
    "delivery-status": { "code": 550, "message": "User unknown" }
  }
}
```

### Body size limit

64 KiB. Larger payloads return `413 Payload Too Large`.

### Response

| Status | Body | Meaning |
|---|---|---|
| `200` | `{"status":"accepted","email":"<addr>"}` | Recipient extracted and added to the suppression list. |
| `400` | `{"error":"invalid project id"}` | The `projectId` in the path is not a UUID. |
| `400` | `{"error":"could not find email in payload"}` | The body matched neither the generic nor the Mailgun shape. |
| `401` | `{"error":"missing or invalid token"}` | The `token` query parameter is missing or not a UUID. |
| `401` | `{"error":"invalid project or token"}` | Token doesn't match this project. |
| `413` | `{"error":"payload too large"}` | Body > 64 KiB. |

The endpoint is **idempotent at the suppression layer** — re-posting the same recipient is a no-op (already on the list). The rate limit is the global per-IP cap (600 req/min behind a proxy that sets `X-Forwarded-For`).

### Configure your provider

| Provider | Where to set the URL |
|---|---|
| Mailgun | Sending → Domain settings → Webhooks → Permanent failure |
| SES via SNS | Subscribe an SNS topic; use a small Lambda or AWS API Destination to translate the SNS message into the generic `{ email, reason }` shape and POST to `/webhooks/bounces/{id}`. |
| Postmark | Bounce webhook → JSON body; Postmark's payload contains `Email` (capital), so use a small adapter. |

## See also

- [Bounces guide](../guide/bounces) — when each detection source fires and how it interacts with suppressions and webhooks.
- [Suppressions API](./suppressions) — what the bounce sources write to.
- [Webhooks API](./webhooks) — `email.bounced` event fired downstream when a hard bounce is recorded.
