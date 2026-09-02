# Projects API

All endpoints require cookie authentication.

## Create Project

```
POST /api/v1/projects
```

```json
{
  "workspace_id": "uuid",
  "name": "My Project",
  "description": "Optional description"
}
```

`workspace_id` is **required** — every project belongs to exactly one workspace. List the workspaces you have access to with `GET /api/v1/workspaces` and pick the right one. Posting without it returns `400 {"error": "workspace_id is required"}`.

`description` is optional.

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
  "unsubscribe_template_id": null,
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

All four fields — `smtp_host`, `smtp_port`, `smtp_user`, `smtp_password` — are required on every update; omitting `smtp_password` returns `400`, it does **not** keep the stored credential. To change only the host/port/user, re-send the current password alongside them.

## Test SMTP

```
POST /api/v1/projects/{id}/smtp/test
```

Sends a single test message to the project's `from_email` (falling back to the SMTP username). Returns `200` with `{"message":"SMTP connection successful. Test email sent."}` on acceptance; a connection or delivery failure returns `400` with the SMTP error in the `error` field.

## Unsubscribe page template

```
PUT /api/v1/projects/{id}/unsubscribe-template
```

```json
{"template_id": "uuid"}
```

Sets the [page template](/api/templates#template-types) rendered on the public unsubscribe pages (both the confirmation and the done page). The template must belong to the project and have `type: "page"` — anything else returns `400`. Send `{"template_id": ""}` to clear the setting and fall back to the built-in page. Requires the `project:settings` capability. Deleting the template clears the setting automatically.

Invalid or tampered unsubscribe links always render the built-in error page, never the branded template.
