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
{"total": 1520, "sent": 1500, "failed": 20}
```

## Authentication

All sending endpoints accept both cookie auth (from the UI) and API key auth (`Authorization: Bearer sk_...`).
