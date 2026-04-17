# Projects API

All endpoints require cookie authentication.

## Create Project

```
POST /api/v1/projects
```

```json
{"name": "My Project", "description": "Optional description"}
```

**Response** `201`

```json
{
  "id": "uuid",
  "name": "My Project",
  "description": "Optional description",
  "from_name": null,
  "from_email": null,
  "smtp_host": null,
  "smtp_port": null,
  "smtp_user": null,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

## List Projects

```
GET /api/v1/projects
```

Returns an array of projects owned by the authenticated user.

## Get Project

```
GET /api/v1/projects/{id}
```

## Update Project

```
PUT /api/v1/projects/{id}
```

```json
{"name": "Updated Name", "description": "Updated description"}
```

## Delete Project

```
DELETE /api/v1/projects/{id}
```

Returns `204 No Content`. Deletes all associated subscribers, templates, API keys, and email logs.

## Update SMTP Settings

```
PUT /api/v1/projects/{id}/smtp
```

```json
{
  "smtp_host": "smtp.gmail.com",
  "smtp_port": 587,
  "smtp_user": "you@gmail.com",
  "smtp_password": "app-password",
  "from_name": "My Newsletter",
  "from_email": "noreply@mydomain.com"
}
```

Required fields: `smtp_host`, `smtp_port`, `smtp_user`, `smtp_password`.

## Test SMTP

```
POST /api/v1/projects/{id}/smtp/test
```

Sends a test email to verify the SMTP configuration is correct. Returns an error message if the connection fails.
