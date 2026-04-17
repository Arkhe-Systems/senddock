# Campaigns API

All endpoints accept both cookie auth and API key auth (`Authorization: Bearer sk_...`).

## Create Campaign

```
POST /api/v1/projects/{id}/campaigns
```

```json
{
  "template_id": "uuid",
  "name": "April Newsletter",
  "scheduled_at": "2026-04-20T09:00:00Z"
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `template_id` | string (UUID) | Yes | The template to send |
| `name` | string | Yes | Campaign name for identification |
| `scheduled_at` | string (RFC 3339) | Yes | When to send (must be in the future) |

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "template_id": "uuid",
  "name": "April Newsletter",
  "status": "scheduled",
  "scheduled_at": "2026-04-20T09:00:00Z",
  "sent_at": null,
  "created_at": "2026-04-16T12:00:00Z",
  "updated_at": "2026-04-16T12:00:00Z"
}
```

## List Campaigns

```
GET /api/v1/projects/{id}/campaigns
```

**Response**

```json
{
  "campaigns": [
    {
      "id": "uuid",
      "project_id": "uuid",
      "template_id": "uuid",
      "name": "April Newsletter",
      "status": "scheduled",
      "scheduled_at": "2026-04-20T09:00:00Z",
      "sent_at": null,
      "created_at": "2026-04-16T12:00:00Z",
      "updated_at": "2026-04-16T12:00:00Z"
    }
  ]
}
```

## Delete / Cancel Campaign

```
DELETE /api/v1/projects/{id}/campaigns/{campaignId}
```

Deletes a campaign. Only campaigns with `scheduled` status can be deleted. Attempting to delete a campaign that is `sending`, `sent`, or `failed` returns `400 Bad Request`.

**Response** `204 No Content`

## Campaign Statuses

| Status | Description |
|--------|-------------|
| `scheduled` | Waiting for the scheduled time to arrive |
| `sending` | Currently broadcasting to subscribers |
| `sent` | All emails have been delivered |
| `failed` | An error occurred during sending |

The background worker checks for due campaigns every 30 seconds. When a campaign's `scheduled_at` time has passed, it broadcasts the selected template to all active subscribers in the project, with per-subscriber variable replacement.
