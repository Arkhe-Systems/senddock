# Subscribers API

All endpoints require cookie authentication. The authenticated user must own the project.

## Add Subscriber

```
POST /api/v1/projects/{id}/subscribers
```

```json
{
  "email": "user@example.com",
  "name": "John Doe",
  "status": "active",
  "fields": {"plan_tier": "pro", "country": "CO"},
  "tags": ["vip", "beta"]
}
```

`status` is optional, defaults to `active`. Valid values: `active`, `pending`, `unsubscribed`. `pending` means the address registered but hasn't been confirmed; pending subscribers are excluded from broadcasts until set to `active` (there's no automatic confirmation link — activate them via [Update Status](#update-status) or the dashboard).

`fields` is optional — a map of [custom field](#custom-fields) values, validated against the project's field definitions. Unknown keys return `400`. `tags` is optional — an array of free-form labels (deduplicated and trimmed).

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "email": "user@example.com",
  "name": "John Doe",
  "status": "active",
  "fields": {"plan_tier": "pro", "country": "CO"},
  "tags": ["vip", "beta"],
  "subscribed_at": "2026-01-01T00:00:00Z",
  "unsubscribed_at": null,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Returns `409 Conflict` if the email already exists in the project, or `400` if a custom field value fails validation.

## List Subscribers

```
GET /api/v1/projects/{id}/subscribers?limit=50&offset=0
```

Optional filters, combinable:

| Parameter | Description | Example |
|-----------|-------------|---------|
| `status` | Filter by status: `active`, `pending`, `unsubscribed` | `?status=active` |
| `tag` | Only subscribers carrying this tag | `?tag=vip` |
| `newsletter_id` | Only opted-in members of this [newsletter](/api/newsletters) | `?newsletter_id=...` |

**Response**

```json
{
  "subscribers": [...],
  "total": 150
}
```

`total` counts the rows matching the filters, not the whole project.

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
| `fields` | no | Map of [custom field](#custom-fields) values. Validated per-row; a row whose fields fail validation is rejected (with the validation message as its `reason`) rather than aborting the whole import. |
| `tags` | no | Array of tags to attach to the row. |

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
| `action` | string | yes | One of `delete`, `update_status`, `add_tags`, `remove_tags`, `add_newsletter`, `remove_newsletter`. |
| `subscriber_ids` | string[] | yes | Non-empty list of subscriber UUIDs. |
| `status` | string | required for `update_status` | One of `active`, `pending`, `unsubscribed`. |
| `tags` | string[] | required for `add_tags` / `remove_tags` | Non-empty list of tags to add to or remove from every selected subscriber. |
| `newsletter_id` | string | required for `add_newsletter` / `remove_newsletter` | The [newsletter](/api/newsletters) to add every selected subscriber to (clearing any opt-out) or remove them from. |

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

The same endpoint also updates [custom field](#custom-fields) values — pass `fields` instead of (or alongside) `status`:

```json
{"fields": {"plan_tier": "team"}}
```

`fields` **replaces** the subscriber's stored values. Values are validated against the project's field definitions; unknown keys return `400`.

## Custom Fields

Custom fields are typed, project-scoped attributes stored on each subscriber beyond `email`/`name`/`status` (e.g. `plan_tier`, `country`, `birthday`).

Custom fields work in **two steps** — this is the part that trips people up:

1. **Define the field once, at the project level** (`POST /projects/{id}/fields`). This is what "applies to every subscriber" — it declares a typed column available across the whole project. There is no per-subscriber field creation.
2. **Set the value per subscriber** through the `fields` map on the [Add Subscriber](#add-subscriber), [Bulk Import](#bulk-import) or [Update](#update-status) endpoints. There is **no** separate "add value" endpoint — values ride along with the subscriber write.

A value for a key that has no definition is rejected with `400` (no schema drift). So always create the definition first, then write values.

Once defined, values are validated on write against the definition, and are available in templates as <span v-pre>`{{custom.KEY}}`</span> and in [segment](/api/segments) rules as `custom.KEY`.

### List field definitions

```
GET /api/v1/projects/{id}/fields
```

```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "key": "plan_tier",
    "label": "Plan tier",
    "field_type": "enum",
    "options": ["free", "pro", "team"],
    "required": false,
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

### Create a field definition

```
POST /api/v1/projects/{id}/fields
```

```json
{
  "key": "plan_tier",
  "label": "Plan tier",
  "field_type": "enum",
  "options": ["free", "pro", "team"],
  "required": false
}
```

| Field | Required | Description |
|---|---|---|
| `key` | yes | Machine key. Must start with a letter and contain only letters, numbers or underscores. Unique per project. Immutable after creation. |
| `label` | no | Human label for the UI. Defaults to `key`. |
| `field_type` | yes | One of `string`, `number`, `date`, `boolean`, `enum`. |
| `options` | required for `enum` | Allowed values for `enum` fields. |
| `required` | no | When `true`, subscriber writes must supply a non-empty value for this key. Defaults to `false`. |

Returns `400` for an invalid key/type or an `enum` with no options, `409` if the key already exists.

Value validation by type: `string` → string, `number` → JSON number, `boolean` → JSON boolean, `date` → `YYYY-MM-DD` string, `enum` → one of `options`.

### Update a field definition

```
PATCH /api/v1/projects/{id}/fields/{fieldId}
```

```json
{"label": "Subscription tier", "options": ["free", "pro", "team", "enterprise"], "required": true}
```

`key` and `field_type` cannot change. Send the full `options` list for `enum` fields.

### Delete a field definition

```
DELETE /api/v1/projects/{id}/fields/{fieldId}
```

Returns `204`. Existing subscriber values for that key stay stored until the subscriber's next write.

## Tags

Tags are free-form labels on a subscriber. Attach them at create/import time (`tags`), in bulk via the [Bulk Action](#bulk-action) endpoint (`add_tags` / `remove_tags`), or set the full list on a single subscriber.

### Set a subscriber's tags

```
PUT /api/v1/projects/{id}/subscribers/{subscriberId}/tags
```

```json
{"tags": ["vip", "beta"]}
```

Replaces the subscriber's entire tag list (deduplicated and trimmed). Returns the updated subscriber.

### List all tags in a project

```
GET /api/v1/projects/{id}/tags
```

Returns the distinct set of tags used across the project's subscribers, sorted:

```json
["beta", "customer", "vip"]
```

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
- `template_id` must reference an `email`-type template (page templates are reserved for unsubscribe pages); the subscriber's address is available to it as `{{email}}`
- If `template_id` is provided, sends that template as a confirmation email
- `template_id` is optional — omit it to just collect the email without sending a confirmation. The email is informational: it does **not** activate the subscriber.
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
