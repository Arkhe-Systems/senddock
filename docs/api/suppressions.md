# Suppressions API

Manage the per-project suppression list — addresses that `/send`, `/send/batch` and `/broadcast` will skip. See the [Suppressions guide](../guide/suppressions) for the conceptual model and how the list interacts with bounces and unsubscribes.

**Cookie auth only.** Suppression management requires the `suppressions:write` capability that an API key (project-scoped, identity-less) does not carry. Use the dashboard, or call from your own UI built on the same login flow.

The cURL examples below use `-b cookies.txt` to indicate the cookie jar from a prior `POST /api/v1/auth/login`.

## List suppressions

```
GET /api/v1/projects/{id}/suppressions
```

### Query parameters

| Parameter | Type | Description |
|---|---|---|
| `limit` | int (default `50`, max `200`) | Page size. |
| `offset` | int (default `0`) | Pagination offset. |
| `reason` | string | Filter to a single reason (`hard_bounce`, `unsubscribed`, `manual`). Optional. |

### Response

```json
{
  "suppressions": [
    {
      "id": "01H...",
      "project_id": "01H...",
      "email": "user@example.com",
      "reason": "hard_bounce",
      "source": "webhook ingest: 550 5.1.1 mailbox unavailable",
      "created_at": "2026-04-30T14:22:11Z",
      "last_seen_at": "2026-05-02T09:00:03Z"
    }
  ],
  "total": 1284
}
```

`source` is optional free-text recorded at insert time (e.g. the SMTP error code that caused the bounce). `last_seen_at` is updated whenever a new send attempt to the same address bumps into the list.

### Example

```bash
curl -G "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/suppressions" \
  -b cookies.txt \
  --data-urlencode "reason=hard_bounce" \
  --data-urlencode "limit=100"
```

## Add suppressions

```
POST /api/v1/projects/{id}/suppressions
```

Add one or more addresses to the list. Already-suppressed entries are silently de-duplicated.

### Request body

```json
{
  "emails": ["user@example.com", "OTHER@example.com"],
  "reason": "manual",
  "source": "imported from legacy ESP"
}
```

| Field | Type | Required | Description |
|---|---|---|---|
| `emails` | string[] | yes | One or more addresses. Whitespace trimmed, lower-cased server-side. |
| `reason` | string | no | One of `hard_bounce`, `unsubscribed`, `manual`. Defaults to `manual`. |
| `source` | string | no | Free-text note (≤ 255 chars), shown in the UI's `source` column and in audit log entries. |

### Response

```json
{ "added": 2, "skipped": 0 }
```

`skipped` includes both rejected rows (no `@`, empty after trim) and entries that were already on the list.

### Capability

Requires the `suppressions:write` capability — owners, admins and developers can call this endpoint; viewers cannot. API keys bypass the role check.

### Example

```bash
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/suppressions" \
  -b cookies.txt \
  -H "Content-Type: application/json" \
  -d '{"emails":["user@example.com"],"reason":"manual","source":"imported from legacy ESP"}'
```

## Remove a suppression

```
DELETE /api/v1/projects/{id}/suppressions/{suppressionId}
```

Removes the entry. The next send to that address will go through normally — useful when you've confirmed a bounce was a typo or the recipient asked back in.

### Response

`204 No Content` on success.

### Capability

Requires `suppressions:write`.

### Example

```bash
curl -X DELETE "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/suppressions/01H..." \
  -b cookies.txt
```

## Errors

| Status | Cause |
|---|---|
| `400` | `emails` array is empty / missing, invalid project id, invalid suppression id. |
| `401` | Missing / invalid API key. |
| `403` | The role doesn't have `suppressions:write`. |
| `404` | Project not found, or suppression id not part of this project. |

## Audit

`POST` and `DELETE` are recorded in the [Pro audit log](./audit-log) with the `suppression.add` and `suppression.delete` actions respectively. The metadata for `add` includes `{ "added": N, "reason": "..." }`.
