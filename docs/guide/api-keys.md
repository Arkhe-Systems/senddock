# API Keys

API keys allow external applications to authenticate with SendDock's API. Each key is scoped to a single project.

![The API Keys section in project Settings](/screenshots/api-keys.png)

## Creating a Key

Go to **Settings** in the project sidebar, find the **API Keys** section, and click **+ Create Key**. Give it a descriptive name.

![Creating an API key](/screenshots/new-api-key-modal.png)

The key is shown only once after creation. Copy it immediately.

## Key Format

Keys use the format `sk_` followed by 64 hex characters:

```
sk_a1b2c3d4e5f6...
```

The `sk_` prefix identifies it as a SendDock API key. Only the first 10 characters (prefix) are stored and shown in the UI. The full key is hashed with SHA-256 before storage.

## Using a Key

Pass the key in the `Authorization` header:

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/send \
  -H "Authorization: Bearer sk_your_full_key_here" \
  -H "Content-Type: application/json" \
  -d '{"to":"user@example.com","template_id":"YOUR_TEMPLATE_ID","data":{"name":"Jane"}}'
```

## Key Scope

API keys are deliberately narrow: they unlock only the endpoints you'd reach for from external code:

| Endpoint | What it does |
|---|---|
| `POST /api/v1/projects/{id}/send` | Single transactional send. |
| `POST /api/v1/projects/{id}/send/batch` | Same template, many recipients. |
| `POST /api/v1/projects/{id}/broadcast` | Send to every active subscriber. |
| `POST /api/v1/projects/{id}/subscribers/import` | CSV / JSON bulk import. |
| `GET /api/v1/projects/{id}/stats` | Read-only counts (sent, failed, bounced, suppressed, opened). |

**Everything else** — managing subscribers individually, editing templates, scheduling campaigns, configuring SMTP, reading the audit log — requires cookie auth (the dashboard, or your own UI built against the same login flow). The reason: those operations key off role-based capabilities tied to the user identity, which a project-scoped key has none of.

A key is also strictly scoped to its own project. It does not grant access to other projects, workspace-level operations, or your account.

## Last Used

SendDock tracks when each key was last used. Check this in Settings to identify unused keys.

## Revoking a Key

Click **Revoke** next to the key in Settings. This is immediate and permanent. Any application using that key will stop working.

## API

See [API Keys API](/api/api-keys) for the REST API reference.
