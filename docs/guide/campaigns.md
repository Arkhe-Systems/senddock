# Campaigns

Campaigns let you schedule email broadcasts for a future time. Instead of sending immediately, you create a campaign that pairs a template with a scheduled delivery time. When the time arrives, SendDock broadcasts the template to all active subscribers in the project.

## How Campaigns Work

1. You create a campaign with a template, a name, and a scheduled time
2. The campaign sits in `scheduled` status until the scheduled time
3. A background worker checks every 30 seconds for campaigns ready to send
4. When the time arrives, the worker broadcasts the template to all active subscribers
5. The campaign status moves through `sending` and finally to `sent` (or `failed` if something goes wrong)

## Campaign Statuses

| Status | Description |
|--------|-------------|
| `scheduled` | Waiting for the scheduled time |
| `sending` | Currently broadcasting to subscribers |
| `sent` | All emails have been sent |
| `failed` | An error occurred during sending |

## Create a Campaign

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/campaigns \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "uuid",
    "name": "April Newsletter",
    "scheduled_at": "2026-04-20T09:00:00Z"
  }'
```

The `scheduled_at` field must be in RFC 3339 format and must be in the future.

## List Campaigns

```bash
curl https://your-instance.com/api/v1/projects/{id}/campaigns \
  -H "Authorization: Bearer sk_your_api_key"
```

Returns all campaigns for the project, ordered by scheduled time.

## Cancel a Campaign

You can only delete/cancel a campaign while it is in `scheduled` status:

```bash
curl -X DELETE https://your-instance.com/api/v1/projects/{id}/campaigns/{campaignId} \
  -H "Authorization: Bearer sk_your_api_key"
```

Campaigns that are `sending`, `sent`, or `failed` cannot be deleted.

## How the Worker Operates

The campaign worker runs as part of the SendDock backend process. It polls every 30 seconds for campaigns whose `scheduled_at` time has passed and whose status is still `scheduled`. When it finds one, it:

1. Sets the status to `sending`
2. Loads the template and all active subscribers
3. Sends the email to each subscriber (with variable replacement and unsubscribe URL injection)
4. Sets the status to `sent` (or `failed` if errors occurred)

No additional configuration is needed -- the worker starts automatically with the backend.

## Tips

- Use descriptive campaign names so you can identify them later in the list
- Schedule campaigns at least a few minutes in the future to allow time for review
- The template's subject line is used for the campaign emails
- Subscriber variables (`{{name}}`, `{{email}}`, `{{unsubscribe_url}}`) are replaced per recipient, just like broadcast
