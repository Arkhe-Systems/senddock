# Email Sending

SendDock has two endpoints for sending emails. All require SMTP to be configured.

From the dashboard, **Project → Overview → Send Email** opens a composer that picks a template, fills its variables, and sends to a specific address or your whole list:

![The Send Email composer with template, variables and recipient options](/screenshots/send-modal.png)

The composer reads the `{{placeholders}}` out of the selected template and shows an input for each **custom** variable (the built-ins `{{name}}`, `{{email}}`, `{{unsubscribe_url}}` are listed as read-only chips, since they're filled per subscriber automatically). The **Send to** toggle switches between a single address and your whole list (optionally scoped to a [segment](/guide/segments)).

### Rich-text variables

By default a variable's value is inserted as plain text — see [Templates → Safe by default](/guide/templates#safe-by-default-variables-are-html-escaped). When you want a variable to carry **formatting** (a newsletter body behind a single `{{content}}` tag, say), flip that field's **Text / Rich** toggle to **Rich**. The plain input becomes a small WYSIWYG editor with bold, italic, headings, lists, quotes and links — no HTML to hand-write.

![The rich-text editor on a template variable](/screenshots/send-rich-text.png)

On the API, mark the same fields by listing their names in `html_fields`:

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "template_id": "uuid",
    "data": {"content": "<p>Hello <strong>there</strong></p>"},
    "html_fields": ["content"]
  }'
```

Rich values are **sanitized** server-side (formatting and safe links survive; scripts, event handlers and unknown tags are stripped), then inserted as HTML. Everything not in `html_fields` stays escaped, and subscriber-sourced values (`{{name}}`, `{{custom.*}}`) are never eligible — see the [warning in Templates](/guide/templates#rich-html-variables-opt-in). Broadcasts carry the same `html_fields`; scheduled campaigns don't yet.

## Send (`/send`)

One endpoint for all individual sends. What it does depends on the fields you provide.

### Template to any email (forms, transactional)

No subscriber needed. Ideal for contact forms, password resets, welcome emails.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "template_id": "uuid",
    "data": {"name": "John", "email": "john@example.com"}
  }'
```

The `data` object replaces template variables: `{{name}}` becomes "John". You can use any key/value pairs. `subject` is optional — if provided, it overrides the template's subject. `/broadcast` and campaigns also accept an optional `subject` field for the same purpose.

::: warning Sending without a subject
Emails sent with an empty subject line are treated as spam by most providers (Gmail, Outlook, ProtonMail, etc.) and routed straight to the spam folder. SendDock does not block empty-subject sends, but the dashboard surfaces a warning when the selected template has no subject and no override is provided. Always set a subject — on the template or as an override.
:::

### Template to a subscriber

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"subscriber_id": "uuid", "template_id": "uuid"}'
```

Variables `{{name}}`, `{{email}}`, and `{{unsubscribe_url}}` are replaced automatically.

### Raw HTML (no template)

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Welcome",
    "html_body": "<h1>Hello!</h1>"
  }'
```

## Batch Send (`/send/batch`)

Send a template to multiple recipients in one request. Each recipient can have its own data for variable replacement.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send/batch \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "uuid",
    "recipients": [
      {"to": "user1@example.com", "data": {"name": "John"}},
      {"to": "user2@example.com", "data": {"name": "Jane"}},
      {"to": "user3@example.com", "data": {"name": "Bob"}}
    ]
  }'
```

```json
{"sent": 3, "failed": 0}
```

Ideal for sending notifications or announcements to a known list of recipients without requiring them to be subscribers.

## Broadcast (`/broadcast`)

Send a template to **all active subscribers**. Separated from `/send` for safety — you can't accidentally broadcast by setting a wrong field.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/broadcast \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"template_id": "uuid"}'
```

Response includes the count of sent and failed:

```json
{"sent": 150, "failed": 2}
```

### Unsubscribe link

Every broadcast carries a working unsubscribe link, no matter what your template looks like:

- If your template uses the `{{unsubscribe_url}}` placeholder, SendDock replaces it with a per-recipient link signed with HMAC.
- If your template **does not** use the placeholder, SendDock auto-appends a small unsubscribe footer at the bottom of the email so the link is always present.
- The `List-Unsubscribe` and `List-Unsubscribe-Post: List-Unsubscribe=One-Click` headers are also added (RFC 8058), so Gmail and Outlook expose their native one-click "Unsubscribe" button next to the sender's name. Hitting that button POSTs directly to the unsubscribe endpoint without requiring the recipient to load a webpage.

When a recipient clicks the unsubscribe link inside the email body, SendDock shows a minimal confirmation page styled to match the app — the recipient must click "Confirm unsubscribe" before their status changes. That second step prevents accidental unsubscribes from email-client link prefetchers and corporate security scanners that follow links automatically.

The link is signed with a token derived from `JWT_SECRET` — recipients cannot be unsubscribed by guessing UUIDs, and tampering with the token returns a "Link expired or invalid" page.

```html
<a href="{{unsubscribe_url}}">Unsubscribe</a>
```

### A public URL is required for broadcasts and campaigns

Both `/broadcast` and the campaign scheduler will refuse to run when your public URL is unset or points at localhost. Without a public URL, recipients cannot reach the unsubscribe page, and that is a hard "no" for SendDock — sending without a working unsubscribe is the canonical spam pattern.

The error returned is:

```json
{
  "error": "your public URL is not set to a publicly reachable address. Newsletters need a working unsubscribe link before they can be sent. Set it under Settings → Instance"
}
```

To enable broadcasts:

1. Put SendDock behind a real domain (reverse proxy + DNS).
2. Set your public URL to `https://your-domain.com` under **Instance** in the dashboard — it applies immediately, no restart. See [Instance settings](/guide/instance-settings#public-url).

`/send` (single-recipient transactional) and `/send/batch` continue to work without a public URL, since their use cases are typically internal apps and explicit recipient lists, not list emails.

### How a broadcast actually runs

When you call `/broadcast` (or a scheduled campaign reaches its time), SendDock returns immediately with the total recipient count. The actual sending happens through a **persistent per-recipient queue**, not inline:

1. A `broadcasts` row is created with `status=sending` and `total_recipients=N`.
2. One row per recipient is inserted into `broadcast_jobs` with `status=pending`.
3. Five worker goroutines pull jobs concurrently using `SELECT … FOR UPDATE SKIP LOCKED`, so jobs never get sent twice even with multiple workers (or multiple SendDock instances on the same database).
4. As each job finishes, the worker increments the broadcast's `sent_count`, `failed_count`, or `suppressed_count`. When the queue drains, the broadcast flips to `status=completed` and any linked campaign moves from `sending` to `sent` with the real counts.

This design has three consequences worth knowing:

- **Server restart mid-send is safe.** Jobs that were in `sending` when the process died are rescheduled to `retry` on the next startup; nothing is lost and nothing is double-sent.
- **Per-recipient retries with exponential backoff.** Transient errors (DNS failures, network timeouts, SMTP 4xx) reschedule the job — backoff steps are 30s, 2m, 8m, 30m, 1h. After 5 attempts the job is marked `failed`. SMTP 5xx bounces are *not* retried (they are tagged `bounced` and the recipient is added to the suppression list immediately).
- **Live progress is visible while sending.** The Newsletters list, Broadcasts tab, and the "broadcasts in flight" panel in Analytics all read counts from the live broadcast row, refreshing every 5 seconds while at least one broadcast is in flight. You see `42/213 → 87/213 → 213/213`, not a sudden jump at the end.

If you need to inspect the queue directly:

```sql
SELECT status, COUNT(*) FROM broadcast_jobs WHERE broadcast_id = '...' GROUP BY status;
```

Statuses are `pending`, `retry`, `sending`, `sent`, `failed`, `bounced`, `suppressed`.

## Scheduled Campaigns

For recurring or scheduled sends, use **Campaigns** instead of sending directly. A campaign ties a template to a scheduled time and broadcasts it to all active subscribers when the time arrives.

See the [Campaigns guide](/guide/campaigns) for details on creating and managing campaigns.

## How tracking works

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 380" role="img" aria-label="Tracking sequence" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="tf-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <text x="80" y="32" text-anchor="middle" font-size="12" font-weight="600" letter-spacing="0.04em" text-transform="uppercase" fill="currentColor">Send API</text>
    <text x="280" y="32" text-anchor="middle" font-size="12" font-weight="600" letter-spacing="0.04em" text-transform="uppercase" fill="currentColor">SMTP</text>
    <text x="480" y="32" text-anchor="middle" font-size="12" font-weight="600" letter-spacing="0.04em" text-transform="uppercase" fill="currentColor">Recipient</text>
    <text x="680" y="32" text-anchor="middle" font-size="12" font-weight="600" letter-spacing="0.04em" text-transform="uppercase" fill="currentColor">/t · /c</text>
    <line x1="80" y1="50" x2="80" y2="350" stroke="currentColor" stroke-opacity="0.2" stroke-dasharray="4 4"/>
    <line x1="280" y1="50" x2="280" y2="350" stroke="currentColor" stroke-opacity="0.2" stroke-dasharray="4 4"/>
    <line x1="480" y1="50" x2="480" y2="350" stroke="currentColor" stroke-opacity="0.2" stroke-dasharray="4 4"/>
    <line x1="680" y1="50" x2="680" y2="350" stroke="currentColor" stroke-opacity="0.2" stroke-dasharray="4 4"/>
    <g transform="translate(0,80)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">1</text><text x="80" y="0" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">rewrite &lt;a href&gt;, inject pixel</text><path d="M 80 16 C 80 30 80 30 80 40" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none"/><circle cx="80" cy="16" r="3" fill="currentColor" fill-opacity="0.7"/><circle cx="80" cy="40" r="3" fill="currentColor" fill-opacity="0.7"/></g>
    <g transform="translate(0,140)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">2</text><line x1="80" y1="0" x2="270" y2="0" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#tf-a)"/><text x="180" y="-8" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">deliver</text></g>
    <g transform="translate(0,190)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">3</text><line x1="280" y1="0" x2="470" y2="0" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#tf-a)"/><text x="380" y="-8" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">email lands in inbox</text></g>
    <g transform="translate(0,242)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">4</text><line x1="480" y1="0" x2="670" y2="0" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#tf-a)"/><text x="580" y="-8" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">GET /t/{logId}</text><text x="580" y="14" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">→ opened_at = NOW()</text></g>
    <g transform="translate(0,302)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">5</text><line x1="480" y1="0" x2="670" y2="0" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#tf-a)"/><text x="580" y="-8" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">GET /c/{logId}/{...}</text><text x="580" y="14" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">→ clicked_at + email_clicks</text></g>
    <g transform="translate(0,344)"><text x="20" y="0" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.5">6</text><line x1="670" y1="0" x2="490" y2="0" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#tf-a)"/><text x="580" y="-8" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">302 → original URL</text></g>
  </g>
</svg>

Tracking is on by default for every send. The two touch points — the open pixel and the click redirect — are public endpoints on your SendDock instance, so your public URL must point at a host the recipient can reach.

## Open Tracking

SendDock automatically injects a 1x1 transparent tracking pixel into emails sent to subscribers and via broadcast. When the recipient opens the email and their email client loads the pixel, SendDock records the open.

- The tracking pixel URL is `GET /t/{logId}` (public, no auth — the response body is the 1×1 GIF, no file extension on the path)
- Only the first open is recorded (`opened_at` timestamp on the email log)
- The stats endpoint includes the `opened` count alongside `sent` and `failed`

Open tracking is automatic and requires no configuration.

## Click Tracking

For every link in an outgoing email, SendDock rewrites the `href` to point through a redirect endpoint hosted at the same instance:

```
GET /c/{logId}/{base64url(URL)}.{token}
```

When the recipient clicks the link, the redirect endpoint:

1. Verifies the token (HMAC of `<logId>.<URL>` using `JWT_SECRET`) — tampered or guessed links return 400.
2. Records the click — first click sets `clicked_at` on the email log; every click writes a row to `email_clicks` with the URL, user agent, and IP, which feeds the **Top clicked links** chart in [Analytics](/guide/analytics).
3. Issues a `302 Found` to the original URL.

Properties worth knowing:

- **No subscriber required**. Transactional sends (`/send` to a raw email, `/send/batch`) get tracked clicks too, not just broadcasts.
- **Tokens are unguessable**. Recipients can't fabricate a tracking link by knowing a `logId` — the HMAC is over the full URL plus the log ID.
- **The redirect is fast**. The DB write is best-effort; the redirect happens whether or not it lands. A failed write does not block the recipient from reaching their destination.
- **Only the first click counts** for the `email.clicked` webhook event and for the `clicked_at` log column. The `email_clicks` table records every click for analytics.

Click tracking is enabled by default for all outgoing emails. There is no opt-out at the project level today — links built directly with `https://...` in HTML get rewritten before send.

::: tip Plain-text URLs in HTML
Click tracking only rewrites links inside `<a href="…">` tags. URLs pasted as plain text (`Click here: https://example.com/x`) are left alone — most email clients auto-link them, but those auto-links are untracked. Always wrap a link you want measured in an explicit anchor.
:::

## Sending from the UI

From the project **Overview**, click **Send Email** to:

- Select a template
- Choose "All subscribers" (broadcast) or "Specific email" (direct send)
- Send immediately

## CSS Inlining

SendDock automatically inlines CSS styles before sending. If your template uses `<style>` tags, they are converted to inline `style=""` attributes for compatibility with email clients like Gmail.

## Email Logs

Every email SendDock attempts to send is recorded as one row in the `email_logs` table. View them in the dashboard under **Project → Logs**, or read them via API.

![The email logs view with per-message status](/screenshots/logs.png)

### In the dashboard

The **Logs** tab shows the most recent rows ordered newest-first, with filters at the top of the page:

| Filter | Options |
|---|---|
| **Status** | All · Sent · Failed · Bounced · Suppressed (clickable chips) |
| **Template** | Dropdown — narrow logs to a single template's sends |
| **From** | Date picker — inclusive lower bound on `sent_at` |
| **To** | Date picker — inclusive upper bound (end-of-day) on `sent_at` |

Each row shows the recipient, subject, status, and an **Engagement** column with compact `O` / `C` badges for opens and clicks. Click any row to open the detail drawer (full lifecycle timeline, per-click events with URL + user agent + timestamp, error or suppression reason, reference IDs). Failed and bounced rows expose the underlying SMTP error.

Pagination defaults to **50 rows per page** and accepts up to **100** via the `limit` query parameter; the total count is shown next to the page title. The **Export CSV** button (top right) downloads every row matching the active filters — no `limit`/`offset` cap — useful for ad-hoc analysis in a spreadsheet.

### Status values

| Status | When it lands |
|---|---|
| `sent` | SMTP relay accepted the message. |
| `failed` | Soft failure (4xx) or unexpected error during the send. |
| `bounced` | Hard failure (5xx) detected by [bounce ingestion](./bounces). The recipient is also added to the [suppression list](./suppressions). |
| `suppressed` | Send was skipped before it left SendDock — the recipient was already on the suppression list. |

`bounced` and `suppressed` are tracked **separately** from `failed` so you can tell list-hygiene issues apart from delivery problems.

### Via API

```bash
# Cookie auth only — log in first and pass the saved cookie jar.
curl -b cookies.txt \
  "https://your-instance.com/api/v1/projects/{id}/logs?limit=50&status=bounced"
```

The full request and response shape lives in [Email Sending API → Email Logs](/api/sending#email-logs).

## Stats

```bash
curl https://your-instance.com/api/v1/projects/{id}/stats \
  -H "Authorization: Bearer sk_your_api_key"
```

```json
{
  "total": 1520,
  "sent": 1480,
  "failed": 8,
  "bounced": 22,
  "suppressed": 10,
  "opened": 980
}
```

`/stats` is one of the five endpoints that accept API key auth (cached in Redis for 30 seconds). Click counts and per-template / per-link breakdowns live in the [Analytics dashboard](./analytics) instead.

## Authentication

The send-related endpoints use mixed auth:

- `POST /send`, `POST /send/batch`, `POST /broadcast`, `GET /stats` — accept **both** cookie session and `Authorization: Bearer sk_...` API keys.
- `GET /logs`, `POST /smtp/test` — **cookie only**. They tie to per-user role capabilities that an API key doesn't carry.

## Rate limits and abuse prevention

Sending endpoints are rate-limited per project to stop the API from being used as a spam loop. The limits are designed to leave room for normal application traffic (signup confirmations, password resets, policy notices) while making bulk-loop abuse impractical.

| Endpoint | Limit per project | Window | Hard cap per request |
|---|---|---|---|
| `POST /send` | 60 requests | 1 minute | — |
| `POST /send/batch` | 10 requests | 1 minute | 500 recipients |
| `POST /broadcast` | 5 requests | 1 hour | — |

When you exceed a limit the API returns:

```
HTTP/1.1 429 Too Many Requests
Retry-After: 60
Content-Type: application/json

{"error":"rate limit exceeded for this project on this endpoint. Slow down and retry later"}
```

The `Retry-After` header tells you how many seconds to wait before retrying.

### What "per project" means

Limits are tracked by project ID, not by API key or IP. If you generate three API keys for the same project, they share the project's quota. If you run two SendDock instances behind a load balancer (with shared Redis), they share the limit too.

### How to think about the limits

- **`/send` at 60/min** — fits a steady stream of transactional emails. A typical SaaS sends a confirmation email per signup; even at 1 signup per second this stays inside the limit.
- **`/send/batch` at 10/min × 500 recipients** — up to 5,000 known recipients per minute, intended for explicit recipient lists (announcements to a known set of users, notifications to a vendor list). Trying to use it as a hidden broadcast loop hits the cap.
- **`/broadcast` at 5/hour** — broadcasts to your subscriber list happen rarely on purpose. Five per hour leaves room for legitimate retries and segmented sends without being a spam vector.

### What the limits are not

These are abuse limits, not deliverability limits. Your SMTP provider applies its own quotas (SES at 14/sec by default, Mailgun depending on plan, etc.) which you must stay inside regardless of what SendDock allows. SendDock does not artificially throttle to match your provider — that is your responsibility.

If you need higher throughput than the per-project caps, the right answer is **subscribers + broadcast**, not loops over `/send`. A 50,000-subscriber broadcast counts as 1 broadcast call.

### Disabling Redis

If `REDIS_URL` is not set, rate limiting is bypassed (the cache is required to track counts across requests). Self-hosted instances exposed to the internet should run Redis even if you don't need it for caching, specifically so these limits stay enforced.
