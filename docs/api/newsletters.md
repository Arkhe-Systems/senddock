# Newsletters API

Newsletters are named publications inside a project. Subscribers keep a single project-level status, and on top of that hold per-newsletter memberships they can join and leave individually — unsubscribing from one newsletter never touches the others, the project status, or the suppression list. Cookie auth required; mutations need the `subscribers:write` role capability. Part of the free Core.

See the [Newsletters guide](/guide/newsletters) for concepts and the unsubscribe behavior.

## List newsletters

```
GET /api/v1/projects/{id}/newsletters
```

```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "name": "Dev Tips",
    "description": "Weekly software development digest",
    "active_count": 128,
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
]
```

`active_count` is the number of **active** subscribers currently opted in (membership present and not unsubscribed from this newsletter).

## Create a newsletter

```
POST /api/v1/projects/{id}/newsletters
```

```json
{"name": "Dev Tips", "description": "Weekly software development digest"}
```

Returns `201` with the created newsletter. Names are unique per project; a duplicate returns `400`. `description` is optional.

## Update a newsletter

```
PATCH /api/v1/projects/{id}/newsletters/{newsletterId}
```

```json
{"name": "Dev Tips Weekly", "description": "Now every Friday"}
```

## Delete a newsletter

```
DELETE /api/v1/projects/{id}/newsletters/{newsletterId}
```

Returns `204`. Memberships are removed; subscribers stay in the project untouched. Unsubscribe links already sent for this newsletter keep working — they fall back to a project-wide unsubscribe.

## Subscriber memberships

### List a subscriber's newsletters

```
GET /api/v1/projects/{id}/subscribers/{subscriberId}/newsletters
```

```json
[
  {"id": "uuid", "name": "Dev Tips", "unsubscribed_at": null},
  {"id": "uuid", "name": "AI Weekly", "unsubscribed_at": "2026-02-01T00:00:00Z"}
]
```

Only newsletters the subscriber has a membership row for are returned. `unsubscribed_at: null` means opted in; a timestamp means they left that newsletter.

### Set a subscriber's newsletters

```
PUT /api/v1/projects/{id}/subscribers/{subscriberId}/newsletters
```

```json
{"newsletter_ids": ["uuid", "uuid"]}
```

Full replacement, like subscriber tags: the subscriber ends up a member of exactly these newsletters. Re-adding a newsletter the subscriber had left clears their opt-out — use deliberately.

### Bulk add / remove

The [bulk subscribers endpoint](/api/subscribers#bulk-action) accepts two newsletter actions:

```json
{"action": "add_newsletter", "newsletter_id": "uuid", "subscriber_ids": ["uuid", "uuid"]}
```

```json
{"action": "remove_newsletter", "newsletter_id": "uuid", "subscriber_ids": ["uuid", "uuid"]}
```

`add_newsletter` upserts memberships (clearing any opt-out); `remove_newsletter` deletes the membership rows.

## Sending to a newsletter

Pass `newsletter_id` to the [broadcast endpoint](/api/sending#broadcast). Recipients are the active subscribers opted in to that newsletter. `newsletter_id` and `segment_id` are mutually exclusive — sending both returns `400`.

```json
{"template_id": "uuid", "newsletter_id": "uuid"}
```

Emails sent this way carry a per-newsletter unsubscribe link (and matching `List-Unsubscribe` header): the reader leaves that one newsletter, stays subscribed to everything else, and the confirmation page offers a separate "unsubscribe from all emails" option. [Campaigns](/api/campaigns) accept the same `newsletter_id` field.

## Filtering by newsletter

Once sends are attributed to a newsletter, these endpoints accept a `newsletter_id` query parameter:

- `GET /api/v1/projects/{id}/analytics/overview`, `…/analytics/campaigns`, `…/analytics/engagement` — scope metrics to one newsletter ([Analytics API](/api/analytics))
- `GET /api/v1/projects/{id}/logs` and `…/logs/export.csv` — filter email logs
- `GET /api/v1/projects/{id}/broadcasts` — filter history (also accepts `status`)
- `GET /api/v1/projects/{id}/subscribers` — list members of one newsletter (also accepts `status` and `tag`)

## Webhook

Leaving a newsletter emits `subscriber.newsletter_unsubscribed` with `subscriber_id`, `newsletter_id`, `newsletter_name` and the subscriber's `email`. See [Webhooks](/guide/webhooks#events).
