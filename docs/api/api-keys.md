# API Keys API

All endpoints require cookie authentication.

## Create API Key

```
POST /api/v1/projects/{id}/keys
```

```json
{"name": "Production API"}
```

**Response** `201`

```json
{
  "key": "sk_a1b2c3d4e5f6...",
  "api_key": {
    "id": "uuid",
    "project_id": "uuid",
    "name": "Production API",
    "key_prefix": "sk_a1b2c3d",
    "last_used_at": null,
    "created_at": "2026-01-01T00:00:00Z"
  }
}
```

The `key` field contains the full API key. It is only returned on creation and cannot be retrieved again. Store it securely.

## List API Keys

```
GET /api/v1/projects/{id}/keys
```

Returns an array of API keys. Only the prefix is shown, never the full key.

```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "name": "Production API",
    "key_prefix": "sk_a1b2c3d",
    "last_used_at": "2026-01-15T10:30:00Z",
    "created_at": "2026-01-01T00:00:00Z"
  }
]
```

## Revoke API Key

```
DELETE /api/v1/projects/{id}/keys/{keyId}
```

Returns `204 No Content`. The key is immediately invalidated.
