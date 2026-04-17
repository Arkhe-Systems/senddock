# Email Sending

SendDock has two endpoints for sending emails. All require SMTP to be configured.

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

The `data` object replaces template variables: `{{name}}` becomes "John". You can use any key/value pairs. `subject` is optional — if provided, it overrides the template's subject.

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

Broadcast emails automatically inject `{{unsubscribe_url}}` — a public link where subscribers can opt out. Use it in your templates:

```html
<a href="{{unsubscribe_url}}">Unsubscribe</a>
```

The link takes the subscriber to a confirmation page and changes their status to `unsubscribed`.

## Scheduled Campaigns

For recurring or scheduled sends, use **Campaigns** instead of sending directly. A campaign ties a template to a scheduled time and broadcasts it to all active subscribers when the time arrives.

See the [Campaigns guide](/guide/campaigns) for details on creating and managing campaigns.

## Open Tracking

SendDock automatically injects a 1x1 transparent tracking pixel into emails sent to subscribers and via broadcast. When the recipient opens the email and their email client loads the pixel, SendDock records the open.

- The tracking pixel URL is `GET /t/{logId}.gif` (public, no auth)
- Only the first open is recorded (`opened_at` timestamp on the email log)
- The stats endpoint includes the `opened` count alongside `sent` and `failed`

Open tracking is automatic and requires no configuration.

## Sending from the UI

From the project **Overview**, click **Send Email** to:

- Select a template
- Choose "All subscribers" (broadcast) or "Specific email" (direct send)
- Send immediately

## CSS Inlining

SendDock automatically inlines CSS styles before sending. If your template uses `<style>` tags, they are converted to inline `style=""` attributes for compatibility with email clients like Gmail.

## Email Logs

Every email sent is logged. View logs from the project Overview or via API:

```bash
curl https://your-instance.com/api/v1/projects/{id}/logs?limit=50 \
  -H "Authorization: Bearer sk_your_api_key"
```

## Stats

```bash
curl https://your-instance.com/api/v1/projects/{id}/stats \
  -H "Authorization: Bearer sk_your_api_key"
```

```json
{"total": 1520, "sent": 1500, "failed": 20, "opened": 980}
```

## Authentication

All sending endpoints accept both cookie auth (from the UI) and API key auth (`Authorization: Bearer sk_...`).
