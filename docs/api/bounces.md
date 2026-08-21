# Bounces API

Configure how SendDock detects bounces for a project. See the [Bounces guide](../guide/bounces) for the conceptual model — these endpoints just expose the per-project knobs.

There are two configurable sources, plus a public ingest endpoint that providers post to:

- **IMAP poller** — SendDock logs into a bounce mailbox every 5 minutes and parses DSNs.
- **Webhook ingest** — SendDock receives a POST from your provider (Mailgun, SES via SNS, custom).

In-session SMTP detection (5xx on `RCPT TO`) needs no configuration and is always on.

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
  "last_polled_at": "2026-05-06T15:30:00Z",
  "last_error": ""
}
```

The password is never returned. `last_polled_at` and `last_error` reflect the most recent poller run.

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
  "url": "https://your-host.example.com/webhooks/bounces/01H...?token=01H...",
  "rotated_at": "2026-04-30T14:22:11Z"
}
```

`url` is the full ingest URL with the current token, ready to paste into your provider's webhook settings. The token is part of the URL — there is no separate field.

## Rotate the bounce webhook token

```
POST /api/v1/projects/{id}/bounce-webhook/rotate
```

Generates a new token and invalidates the old one. The `url` returned by `GET /bounce-webhook` after this call carries the new token.

### Response

```json
{
  "url": "https://your-host.example.com/webhooks/bounces/01H...?token=NEW...",
  "rotated_at": "2026-05-06T16:00:00Z"
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

A single object (or an array of objects) with at least `email` and optional `reason` / `type`:

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
| `400` | `{"error":"could not find email in payload"}` | The body matched neither the generic nor the Mailgun shape. |
| `401` | `{"error":"missing or invalid token"}` | URL token doesn't match this project. |
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
