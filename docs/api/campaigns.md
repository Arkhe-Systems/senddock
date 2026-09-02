# Campaigns API

Cookie auth only. Campaigns mutate workspace state and require role-based capabilities (`campaigns:write` for create / update / delete) that an API key does not carry — the role is bound to the user identity.

A campaign is a **scheduled broadcast** — delivered by the same worker as `POST /broadcast`; use `/broadcast` for immediate sends. Unlike `/broadcast`, campaigns accept `newsletter_id` but have no `segment_id` field.

## Create Campaign

```
POST /api/v1/projects/{id}/campaigns
```

```json
{
  "template_id": "uuid",
  "name": "April Newsletter",
  "subject": "Spring update — what's new",
  "scheduled_at": "2026-04-20T09:00:00Z",
  "variables": { "promo_code": "SPRING25" }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `template_id` | string (UUID) | Yes | The template to send. |
| `name` | string | Yes | Campaign name for the dashboard list. |
| `scheduled_at` | string (RFC 3339) | Yes | When to send. Must be in the future. |
| `subject` | string | No | Overrides the template's stored subject for this campaign only. |
| `variables` | object | No | Map of `string → string` injected into the template body in addition to per-subscriber `{{name}}` and `{{email}}`. |
| `newsletter_id` | string (UUID) | No | Target a [newsletter](/api/newsletters): only its active opted-in members receive the campaign, and unsubscribe links are scoped to it. Omit to send to all active subscribers. |

The deployment must have a public URL configured under **Instance** — campaigns inject unsubscribe links and SendDock refuses to schedule one without a known public URL.

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "template_id": "uuid",
  "newsletter_id": null,
  "name": "April Newsletter",
  "subject": "Spring update — what's new",
  "status": "scheduled",
  "scheduled_at": "2026-04-20T09:00:00Z",
  "sent_at": null,
  "sent_count": 0,
  "failed_count": 0,
  "created_at": "2026-04-16T12:00:00Z",
  "variables": { "promo_code": "SPRING25" }
}
```

## List Campaigns

```
GET /api/v1/projects/{id}/campaigns
```

Returns an array of campaigns ordered by most recent first.

```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "template_id": "uuid",
    "name": "April Newsletter",
    "subject": "Spring update — what's new",
    "status": "sent",
    "scheduled_at": "2026-04-20T09:00:00Z",
    "sent_at": "2026-04-20T09:00:14Z",
    "sent_count": 2341,
    "failed_count": 12,
    "created_at": "2026-04-16T12:00:00Z",
    "variables": { "promo_code": "SPRING25" }
  }
]
```

`sent_count` and `failed_count` update as the campaign worker processes recipients. Once `status` is `sent`, `sent_count + failed_count` covers the recipients actually attempted; suppressed addresses are skipped before any attempt and tracked separately on the underlying broadcast (visible in the Broadcasts tab), not in these two fields.

## Update Campaign

```
PATCH /api/v1/projects/{id}/campaigns/{campaignId}
```

Reschedules or replaces a `scheduled` campaign. Same body shape as Create:

```json
{
  "name": "Spring update — corrected date",
  "template_id": "uuid",
  "scheduled_at": "2026-04-22T10:00:00Z",
  "variables": { "promo": "SPRING25" }
}
```

`name`, `template_id` and `scheduled_at` are required (full replacement, not partial update). `scheduled_at` must be RFC 3339. `variables` and `newsletter_id` are optional and replace the previous values — omitting `newsletter_id` reverts the campaign to all active subscribers.

Only campaigns in `scheduled` status can be patched — once a campaign moves to `sending`, `sent` or `failed`, it's immutable. Cookie auth only; the role must have `campaigns:write` (owners, admins and developers, not viewers).

**Response** the updated campaign object.

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
| `failed` | The broadcast failed to start (for example an enqueue or claim error at trigger time) — not per-recipient failures, which only increment `failed_count`. |

A campaign reaches `sent` when its broadcast finishes draining, even if some recipients failed — those show up in `failed_count`. `failed` is reserved for the broadcast itself failing to start.

The background worker checks for due campaigns every 30 seconds. When a campaign's `scheduled_at` time has passed, it broadcasts the selected template to all active subscribers in the project, with per-subscriber variable replacement.
