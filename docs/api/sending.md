# Email Sending API

Email endpoints accept both cookie auth and API key auth (`Authorization: Bearer sk_...`).

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

Or for subscriber sends:

```json
{"sent": 1, "failed": 0}
```

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
{"sent": 3, "failed": 0}
```

## Broadcast

```
POST /api/v1/projects/{id}/broadcast
```

Sends a template to **all active subscribers** in the project. Separated from `/send` for safety.

```json
{"template_id": "uuid"}
```

Variables are replaced per subscriber. The `{{unsubscribe_url}}` is injected automatically with a link to the public unsubscribe page.

**Response**

```json
{"sent": 150, "failed": 2}
```

## Unsubscribe

```
GET /unsubscribe/{projectId}/{subscriberId}
```

Public endpoint (no auth required). Shows a confirmation page and sets the subscriber's status to `unsubscribed`. The URL is auto-generated and injected via `{{unsubscribe_url}}` in broadcast and subscriber sends.

## Test SMTP

```
POST /api/v1/projects/{id}/smtp/test
```

Sends a test email to verify SMTP configuration. Cookie auth only.

## Open Tracking

```
GET /t/{logId}.gif
```

Public endpoint (no auth required). Returns a 1x1 transparent GIF pixel. This pixel is automatically injected into emails sent to subscribers and via broadcast. When the recipient's email client loads the image, SendDock records the `opened_at` timestamp on the corresponding email log entry. Only the first open is recorded.

## Email Logs

```
GET /api/v1/projects/{id}/logs?limit=50&offset=0
```

Cookie auth only.

### Filters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `status` | Filter by delivery status | `?status=sent` or `?status=failed` |
| `from` | Start date (RFC 3339) | `?from=2026-01-01T00:00:00Z` |
| `to` | End date (RFC 3339) | `?to=2026-02-01T00:00:00Z` |

Example with filters:

```
GET /api/v1/projects/{id}/logs?status=sent&from=2026-01-01T00:00:00Z&to=2026-02-01T00:00:00Z&limit=50&offset=0
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
      "sent_at": "2026-01-01T00:00:00Z",
      "opened_at": "2026-01-01T01:23:45Z"
    }
  ],
  "total": 1520
}
```

## Stats

```
GET /api/v1/projects/{id}/stats
```

```json
{"total": 1520, "sent": 1500, "failed": 20, "opened": 980}
```
