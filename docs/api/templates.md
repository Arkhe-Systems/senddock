# Templates API

All endpoints require cookie authentication.

## Create Template

```
POST /api/v1/projects/{id}/templates
```

```json
{
  "name": "Welcome Email",
  "subject": "Welcome {{name}}!",
  "html_body": "<h1>Hello {{name}}</h1>",
  "text_body": ""
}
```

Only `name` is required.

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "name": "Welcome Email",
  "subject": "Welcome {{name}}!",
  "html_body": "<h1>Hello {{name}}</h1>",
  "text_body": "",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

## List Templates

```
GET /api/v1/projects/{id}/templates
```

Returns an array of templates ordered by last updated.

## Get Template

```
GET /api/v1/projects/{id}/templates/{templateId}
```

## Update Template

```
PUT /api/v1/projects/{id}/templates/{templateId}
```

```json
{
  "name": "Welcome Email v2",
  "subject": "Welcome!",
  "html_body": "<h1>Hello {{name}}</h1><p>Welcome aboard.</p>",
  "text_body": ""
}
```

## Delete Template

```
DELETE /api/v1/projects/{id}/templates/{templateId}
```

Returns `204 No Content`.
