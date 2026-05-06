# Subscribers API

All endpoints require cookie authentication. The authenticated user must own the project.

## Add Subscriber

```
POST /api/v1/projects/{id}/subscribers
```

```json
{"email": "user@example.com", "name": "John Doe", "status": "active"}
```

`status` is optional, defaults to `active`. Valid values: `active`, `pending`, `unsubscribed`.

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "status": "active",
  "subscribed_at": "2026-01-01T00:00:00Z",
  "unsubscribed_at": null,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Returns `409 Conflict` if the email already exists in the project.

## List Subscribers

```
GET /api/v1/projects/{id}/subscribers?limit=50&offset=0
```

**Response**

```json
{
  "subscribers": [...],
  "total": 150
}
```

## Bulk Import

```
POST /api/v1/projects/{id}/subscribers/import
```

Imports many subscribers at once with per-row validation. Accepts both cookie auth and API key auth.

### Request body

The body is a **JSON array** at the top level (not wrapped in `{ "rows": [...] }`):

```json
[
  {"email": "user1@example.com", "name": "John", "status": "active"},
  {"email": "user2@example.com", "name": "Jane"},
  {"email": "BAD", "name": "Junk"}
]
```

| Field | Required | Description |
|---|---|---|
| `email` | yes | Recipient address. Validated for syntax, MX records and disposable-domain block-list (toggleable below). |
| `name` | no | Display name. Empty string allowed. |
| `status` | no | One of `active`, `pending`, `unsubscribed`. Defaults to `active`. |

### Query parameters

Both default to `true`. Pass `?...=false` to relax validation — useful when you're importing from a source you already trust:

| Parameter | Default | Effect when `false` |
|---|---|---|
| `validate_mx` | `true` | Skip the DNS lookup. Domains with no MX still get accepted. Saves ~50 ms per unique domain. |
| `validate_disposable` | `true` | Skip the built-in disposable-domain block-list. Mailinator / 10minutemail / etc. get imported. |

Syntax validation is always on — addresses without a valid `local@domain.tld` form are rejected regardless.

### Response

```json
{
  "imported": 2,
  "duplicates": 0,
  "syntax_invalid": 1,
  "no_mx": 0,
  "disposable": 0,
  "suppressed": 0,
  "rejected": [
    {"email": "BAD", "name": "Junk", "reason": "syntax_invalid"}
  ]
}
```

| Field | Meaning |
|---|---|
| `imported` | New subscribers actually inserted. |
| `duplicates` | Rows whose `email` was already on the project (silent skip — not an error). |
| `syntax_invalid` | Rows that failed RFC 5322 parsing. |
| `no_mx` | Rows whose domain has no MX record (only counted when `validate_mx=true`). |
| `disposable` | Rows whose domain is on the disposable block-list (only counted when `validate_disposable=true`). |
| `suppressed` | Rows whose email is on the project's [suppression list](./suppressions); skipped without insert. |
| `rejected` | Per-row breakdown of every row that didn't make it: `{email, name, reason}`. `reason` is one of `syntax_invalid`, `no_mx`, `disposable`, `suppressed`, `duplicate`. |

The five reject categories sum with `imported` to the input count, so the result is exhaustive — every row your client sent is accounted for.

## Bulk Action

```
POST /api/v1/projects/{id}/subscribers/bulk
```

Apply the same operation to many existing subscribers — the dashboard uses this for "select all → delete" or "select all → unsubscribe" flows.

**Request body**

```json
{
  "action": "update_status",
  "status": "unsubscribed",
  "subscriber_ids": ["01H...", "01H..."]
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `action` | string | yes | One of `delete` or `update_status`. |
| `subscriber_ids` | string[] | yes | Non-empty list of subscriber UUIDs. |
| `status` | string | required for `update_status` | One of `active`, `pending`, `unsubscribed`. |

Cookie auth only (the role must have `subscribers:write`). For ingesting fresh rows, use [Bulk Import](#bulk-import) — that endpoint takes raw `email`/`name` rows and accepts API keys; this one operates on already-stored subscriber ids.

**Response** `204 No Content`. Subscriber ids that don't belong to the project are silently skipped.

## Update Status

```
PATCH /api/v1/projects/{id}/subscribers/{subscriberId}
```

```json
{"status": "unsubscribed"}
```

Valid values: `active`, `pending`, `unsubscribed`. When set to `unsubscribed`, the `unsubscribed_at` timestamp is recorded.

## Delete Subscriber

```
DELETE /api/v1/projects/{id}/subscribers/{subscriberId}
```

Returns `204 No Content`.

## Waitlist (Public)

```
POST /api/v1/projects/{id}/waitlist
```

Public endpoint — no authentication required. Designed for landing page waitlist forms.

```json
{
  "email": "user@example.com",
  "template_id": "uuid"
}
```

- Creates a subscriber with `pending` status
- If `template_id` is provided, sends a confirmation email using that template
- `template_id` is optional — omit it to just collect the email without sending a confirmation
- Returns `409` if the email is already on the waitlist
- Rate limited (100 requests/min per IP)
- Email format is validated

**Response**

```json
{"message": "joined"}
```

**Quick test from the terminal:**

```bash
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/waitlist" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","template_id":"YOUR_TEMPLATE_ID"}'
```

**From a landing page in the browser:**

```javascript
const res = await fetch('https://your-instance.com/api/v1/projects/{id}/waitlist', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
        email: userEmail,
        template_id: 'uuid-of-confirmation-template'
    }),
});
```

No API key needed. The endpoint sets `Access-Control-Allow-Origin: *` so it's safe to call from frontend JavaScript on any domain.
