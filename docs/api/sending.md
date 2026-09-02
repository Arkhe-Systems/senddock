# Email Sending API

`/send`, `/send/batch`, `/broadcast` and `/stats` accept both cookie auth and API key auth (`Authorization: Bearer sk_...`). `/logs` and `/smtp/test` are cookie-only.

## Send

```
POST /api/v1/projects/{id}/send
```

One endpoint for all individual sends. The behavior depends on the fields you provide.

### Send template to any email

```json
{
  "to": "user@example.com",
  "template_id": "uuid",
  "subject": "Optional override",
  "data": {
    "name": "John Doe",
    "email": "john@example.com"
  }
}
```

Sends a template to any email address. No subscriber needed. Variables like `{{name}}` in the template are replaced with values from `data`. If `subject` is provided, it overrides the template's subject.

### Send template to a subscriber

```json
{
  "subscriber_id": "uuid",
  "template_id": "uuid"
}
```

Sends a template to a specific subscriber. Variables `{{name}}`, `{{email}}`, and `{{unsubscribe_url}}` are replaced automatically with the subscriber's data.

### Send raw HTML (no template)

```json
{
  "to": "user@example.com",
  "subject": "Password Reset",
  "html_body": "<p>Click <a href='...'>here</a> to reset.</p>"
}
```

Sends a one-off email without a template. All three fields are required.

### Response

```json
{"message": "sent"}
```

For the subscriber-send variant (`subscriber_id` + `template_id`):

```json
{"sent": 1, "failed": 0}
```

When the recipient is on the project's [suppression list](./suppressions), the endpoint **does not** error — it returns `200 OK` with:

```json
{"message": "suppressed", "suppressed": 1}
```

The send is recorded with status `suppressed` in the email log; no SMTP attempt is made. This is the same shape as the failure-counted version below — when present, `suppressed` always counts in addition to `sent` + `failed`.

::: info Error contract
A `template_id` or `subscriber_id` that doesn't exist in the project, a project without SMTP configured, or a subscriber that isn't `active` all return `400` with an `{error}` body. An SMTP-level rejection (the relay refused the message) also returns `400` with the SMTP error in `{error}` — the log row is marked `failed` or `bounced` accordingly. Only suppression is a "soft" `200`, shown above.
:::

## Batch Send

```
POST /api/v1/projects/{id}/send/batch
```

```json
{
  "template_id": "uuid",
  "subject": "Optional override",
  "recipients": [
    {"to": "user1@example.com", "data": {"name": "John"}},
    {"to": "user2@example.com", "data": {"name": "Jane"}},
    {"to": "user3@example.com", "data": {"name": "Bob"}}
  ]
}
```

Sends a template to multiple recipients in one request. Each recipient can have its own `data` for variable replacement. Ideal for campaigns, notifications, or migrating from external systems that provide a list of emails.

**Response**

```json
{"sent": 3, "failed": 0, "suppressed": 1}
```

`suppressed` is omitted when zero. Recipients on the project's suppression list are skipped without an SMTP attempt and do **not** count towards `failed`.

## Broadcast

```
POST /api/v1/projects/{id}/broadcast
```

Sends a template to **all active subscribers** in the project. Separated from `/send` for safety.

```json
{"template_id": "uuid"}
```

The audience can be narrowed with exactly one of two optional fields:

```json
{"template_id": "uuid", "segment_id": "uuid"}
```

```json
{"template_id": "uuid", "newsletter_id": "uuid"}
```

`segment_id` targets the active subscribers matching a [segment](/api/segments); `newsletter_id` targets the active subscribers opted in to a [newsletter](/api/newsletters). Sending both returns `400`.

Variables are replaced per subscriber. The `{{unsubscribe_url}}` is injected automatically with a link to the public unsubscribe page — scoped to the newsletter when `newsletter_id` was used. Newsletter sends can also use `{{newsletter_name}}` in the template body.

**Response**

```json
{"sent": 150, "failed": 2, "suppressed": 8}
```

`suppressed` is the count of subscribers skipped because they were on the project's suppression list. They do **not** count towards `failed`. Field is omitted when zero.

::: info Public URL required
The instance must have a [public URL](../guide/instance-settings#public-url) configured — without it the endpoint refuses with `400` because unsubscribe links can't be built. See [A public URL is required](../guide/sending#a-public-url-is-required-for-broadcasts-and-campaigns).
:::

## Unsubscribe

```
GET  /unsubscribe/{projectId}/{subscriberId}?t=...
POST /unsubscribe/{projectId}/{subscriberId}?t=...
```

Public endpoints (no auth required). The URL is auto-generated and injected via `{{unsubscribe_url}}` in broadcast and subscriber sends; the `t` token is HMAC-signed against `JWT_SECRET` so it can't be forged or reused for a different recipient.

- **`GET`** — renders a confirmation page with an unsubscribe button. Used by anyone who clicks the unsubscribe link in an email. Nothing changes until the reader confirms.
- **`POST`** — performs the unsubscribe. The confirmation page's button posts here, and so do Gmail / Outlook for [RFC 8058](https://www.rfc-editor.org/rfc/rfc8058) **one-click unsubscribe**: the email's `List-Unsubscribe-Post: List-Unsubscribe=One-Click` header tells the client to POST to the URL when the user clicks the inbox-level unsubscribe button.

A project-wide unsubscribe flips the subscriber's status to `unsubscribed` and adds them to the project [suppression list](./suppressions) with reason `unsubscribe`, so subsequent sends from `/send`, `/send/batch` and `/broadcast` skip them.

### Per-newsletter unsubscribe

Links generated for a [newsletter](/api/newsletters) send carry an extra `n` parameter, and the token covers it:

```
/unsubscribe/{projectId}/{subscriberId}?n={newsletterId}&t=...
```

- With `n`, the POST only marks that newsletter's membership as unsubscribed — the subscriber's status and the suppression list are untouched, and other newsletters keep delivering. It also emits the `subscriber.newsletter_unsubscribed` [webhook](/guide/webhooks#events).
- The confirmation page names the newsletter and offers a separate **Unsubscribe from all emails** action, which performs the project-wide unsubscribe above.
- Tokens verify exactly the URL they were issued for: stripping or altering `n` invalidates the link. Old links without `n` keep verifying forever and stay project-wide. If the newsletter was deleted, its links degrade to a project-wide unsubscribe.

Both pages can be branded with a [page template](/api/templates#template-types); without one, SendDock renders a neutral built-in page.

## Test SMTP

```
POST /api/v1/projects/{id}/smtp/test
```

Sends a test email to verify SMTP configuration. Cookie auth only.

## Open Tracking

```
GET /t/{logId}
```

Public endpoint (no auth required, no file extension on the path). Returns a 1×1 transparent GIF in the response body with `Content-Type: image/gif`. This pixel is automatically injected into **every** outgoing email — subscriber sends, broadcasts, and raw transactional sends to a non-subscriber address (`/send`, `/send/batch`). When the recipient's email client loads the image, SendDock records the `opened_at` timestamp on the corresponding email log entry. Only the first open is recorded.

## Click Tracking

```
GET /c/{logId}/{payload}
```

Public endpoint (no auth required). `{payload}` has the form `<base64url-encoded-URL>.<hmac-token>` and is generated automatically when SendDock rewrites `<a href>` tags in outgoing emails.

The handler verifies the HMAC against `JWT_SECRET`, records the click (first click sets `clicked_at` on the email log; every click appends a row to `email_clicks` with URL, user agent and IP), then returns `302 Found` to the original URL.

Tampered or invalid tokens return `400 Bad Request`. Click tracking is on by default for every outgoing email; it does not need to be enabled per project.

## Email Logs

```
GET /api/v1/projects/{id}/logs?limit=50&offset=0
```

Cookie auth only.

### Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `status` | Filter by delivery status. Valid: `sent`, `failed`, `bounced`, `suppressed`. | `?status=bounced` |
| `template_id` | Restrict to logs that came from one template | `?template_id=...` |
| `newsletter_id` | Restrict to logs from sends targeted at one newsletter | `?newsletter_id=...` |
| `from` | Inclusive lower bound on `sent_at` (RFC 3339) | `?from=2026-01-01T00:00:00Z` |
| `to` | Inclusive upper bound on `sent_at` (RFC 3339) | `?to=2026-02-01T00:00:00Z` |
| `q` | Free-text match against `to_email` or `subject` (case-insensitive) | `?q=welcome` |
| `limit` | Page size (default 50, max 100) | `?limit=100` |
| `offset` | Pagination offset | `?offset=50` |

Status meanings: `sent` — accepted by the SMTP server; `bounced` — accepted, then a hard bounce returned (in-session, via webhook, or via IMAP); `failed` — rejected at send time (soft/transport errors); `suppressed` — skipped before any SMTP attempt because the address is on the suppression list.

### Export to CSV

```
GET /api/v1/projects/{id}/logs/export.csv
```

Same query parameters as `/logs` (no `limit`/`offset` — every matching row is exported). Returns `text/csv` with a `Content-Disposition: attachment` header so browsers download it directly. Columns: `id, to_email, subject, status, error, sent_at, opened_at, clicked_at, template_id, subscriber_id`.

Cookie auth only.

Example with filters:

```
GET /api/v1/projects/{id}/logs?status=bounced&from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z&limit=50&offset=0
```

```json
{
  "logs": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "subscriber_id": "uuid",
      "template_id": "uuid",
      "to_email": "user@example.com",
      "subject": "Welcome!",
      "status": "sent",
      "error": null,
      "sent_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 1520
}
```

`opened_at` and `clicked_at` are **not** included in the log row (the row tracks delivery, not engagement). For per-email open/click data, use the detail endpoint below.

### Email log detail

```
GET /api/v1/projects/{id}/logs/{logId}
```

Cookie auth only. Backs the right-side drawer in the Logs page. Returns the full log row **and** every click recorded against it:

```json
{
  "log": {
    "id": "uuid",
    "project_id": "uuid",
    "subscriber_id": "uuid",
    "template_id": "uuid",
    "to_email": "user@example.com",
    "subject": "Welcome!",
    "status": "sent",
    "error": null,
    "sent_at": "2026-01-01T00:00:00Z",
    "opened_at": "2026-01-01T00:02:14Z",
    "clicked_at": "2026-01-01T00:02:31Z"
  },
  "clicks": [
    {
      "url": "https://example.com/pricing",
      "user_agent": "Mozilla/5.0 ...",
      "ip_address": "203.0.113.42",
      "clicked_at": "2026-01-01T00:02:31Z"
    }
  ]
}
```

Returns `404` if the log doesn't belong to the project.

## Broadcasts

### List broadcasts

```
GET /api/v1/projects/{id}/broadcasts
```

Cookie auth only. Returns the broadcast history for the project, newest first, with live counters that update as the worker drains the queue. Optional query filters: `status` (one status value) and `newsletter_id`:

```json
{
  "broadcasts": [
    {
      "id": "uuid",
      "template_id": "uuid",
      "newsletter_id": null,
      "status": "sending",
      "total": 12500,
      "sent_count": 8472,
      "failed_count": 31,
      "created_at": "2026-01-15T10:00:00Z",
      "completed_at": null
    }
  ]
}
```

`status` is one of `pending`, `sending`, `completed`, `interrupted`. (`interrupted` is a legacy state from a mid-broadcast restart on older builds; current versions stay `sending` and resume from the job queue.) While a broadcast is `sending`, `sent_count` and `failed_count` reflect the live state of the per-recipient job queue. The dashboard polls this endpoint every five seconds to drive the live progress bars.

## Stats

```
GET /api/v1/projects/{id}/stats
```

```json
{"total": 1520, "sent": 1480, "failed": 12, "bounced": 8, "suppressed": 20, "opened": 980}
```

## Rate limits

Sending endpoints are rate-limited **per project** (shared across every API key of that project):

| Endpoint | Limit | Window | Hard cap per request |
|---|---|---|---|
| `POST /send` | 60 requests | 1 minute | — |
| `POST /send/batch` | 10 requests | 1 minute | 500 recipients |
| `POST /broadcast` | 5 requests | 1 hour | — |

Exceeding a limit returns `429 Too Many Requests` with a `Retry-After` header. See the [Sending guide](../guide/sending#rate-limits-and-abuse-prevention) for the reasoning and how to stay inside them.
