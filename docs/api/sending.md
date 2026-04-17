# Email Sending API

Email endpoints accept both cookie auth and API key auth (`Authorization: Bearer sk_...`).

## Send to Subscriber

```
POST /api/v1/projects/{id}/send
```

```json
{"subscriber_id": "uuid", "template_id": "uuid"}
```

Sends the template to a specific subscriber. Variables like `{{name}}` and `{{email}}` are replaced with the subscriber's data. The subscriber must be `active`.

**Response**

```json
{"sent": 1, "failed": 0}
```

## Broadcast

```
POST /api/v1/projects/{id}/broadcast
```

```json
{"template_id": "uuid"}
```

Sends the template to all active subscribers in the project. Variables are replaced per subscriber.

**Response**

```json
{"sent": 150, "failed": 2}
```

## Direct Send

```
POST /api/v1/projects/{id}/send/direct
```

```json
{
  "to": "user@example.com",
  "subject": "Password Reset",
  "html_body": "<p>Click <a href='...'>here</a> to reset your password.</p>"
}
```

Sends a one-off email without a template or subscriber. All three fields are required.

**Response**

```json
{"message": "sent"}
```

## Test SMTP

```
POST /api/v1/projects/{id}/smtp/test
```

Sends a test email to verify SMTP configuration. Cookie auth only.

## Email Logs

```
GET /api/v1/projects/{id}/logs?limit=50&offset=0
```

Cookie auth only.

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

## Stats

```
GET /api/v1/projects/{id}/stats
```

```json
{"total": 1520, "sent": 1500, "failed": 20}
```
