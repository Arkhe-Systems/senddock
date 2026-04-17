# Subscribers

Subscribers are the people who receive your emails. Each subscriber belongs to a project and has a unique email within that project.

## Adding Subscribers

### Via the UI

Go to **Subscribers** in the project sidebar and click **+ Add Subscriber**. Provide an email and optional name.

### Via the API

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/subscribers \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "name": "John Doe"}'
```

## Subscriber Status

| Status | Description |
|--------|-------------|
| `active` | Receives emails normally |
| `pending` | Registered but not yet confirmed |
| `unsubscribed` | Opted out, will not receive emails |

Only `active` subscribers receive broadcast emails.

## Managing Subscribers

From the subscribers table you can:

- **Activate** a pending or unsubscribed subscriber
- **Unsubscribe** an active subscriber
- **Delete** a subscriber permanently

## Constraints

- Email must be unique per project (the same email can exist in different projects)
- Deleting a project deletes all its subscribers

## API

## Waitlist

SendDock includes a public waitlist endpoint for collecting emails from landing pages without exposing API keys.

```
POST /api/v1/projects/{id}/waitlist
```

```json
{"email": "user@example.com", "template_id": "uuid"}
```

This creates a subscriber with `pending` status and optionally sends a confirmation email. No authentication needed — safe to call from frontend JavaScript.

Use it to build pre-launch waitlists, beta signups, or email collection forms.

## API

See [Subscribers API](/api/subscribers) for the full REST API reference.
