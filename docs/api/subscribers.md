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
