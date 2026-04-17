# Email Sending

SendDock supports three ways to send emails. All require SMTP to be configured for the project.

## Send to Subscriber

Send an email to a specific subscriber using a template. Template variables are replaced with the subscriber's data.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"subscriber_id": "uuid", "template_id": "uuid"}'
```

The subscriber must have `active` status.

## Broadcast

Send an email to all active subscribers in the project using a template. Variables are replaced per subscriber.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/broadcast \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"template_id": "uuid"}'
```

Response includes the count of sent and failed emails:

```json
{"sent": 150, "failed": 2}
```

## Direct Send

Send a one-off email to any address without a template or subscriber. Useful for transactional emails (password resets, confirmations, etc).

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send/direct \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Welcome",
    "html_body": "<h1>Hello!</h1>"
  }'
```

## Email Logs

Every email sent is logged with status (`sent` or `failed`), error details, recipient, and timestamp.

View logs from the project overview or via API:

```bash
curl https://your-instance.com/api/v1/projects/{id}/logs?limit=50&offset=0 \
  -H "Authorization: Bearer sk_your_api_key"
```

## Stats

Get aggregate stats for a project:

```bash
curl https://your-instance.com/api/v1/projects/{id}/stats \
  -H "Authorization: Bearer sk_your_api_key"
```

```json
{"total": 1520, "sent": 1500, "failed": 20}
```

## Authentication

Email endpoints accept both cookie auth (from the UI) and API key auth (`Authorization: Bearer sk_...`).

## API

See [Email Sending API](/api/sending) for the full reference.
