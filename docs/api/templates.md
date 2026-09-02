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
  "text_body": "",
  "type": "email"
}
```

Only `name` is required. `text_body` is optional — when empty, SendDock generates a plain-text version from `html_body` at save time so every email ships both an HTML and a text part. `subject` and `text_body` apply to `type: "email"` templates only.

**Response** `201`

```json
{
  "id": "uuid",
  "project_id": "uuid",
  "name": "Welcome Email",
  "subject": "Welcome {{name}}!",
  "html_body": "<h1>Hello {{name}}</h1>",
  "text_body": "",
  "type": "email",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

## Template types

`type` is `email` (default) or `page`, set at creation and immutable afterwards — any other value returns `400`.

- **`email`** templates are what you send: broadcasts, campaigns and the API's send endpoints accept only this type.
- **`page`** templates are rendered as web pages, currently for [branded unsubscribe pages](/api/projects#unsubscribe-page-template). They never appear in send-time template pickers.

Page HTML is sanitized before rendering: scripts, iframes, form elements and event-handler attributes are stripped. Style with inline `style` attributes or a `<style>` block — both are preserved and inlined when the page renders. Available placeholders: `{{project_name}}`, `{{email}}`, `{{newsletter_name}}` (falls back to the project name on project-wide links), `{{confirm_button}}` (the unsubscribe action, injected automatically if omitted) and `{{unsubscribe_all_link}}`.

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
