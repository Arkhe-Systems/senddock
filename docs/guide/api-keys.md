# API Keys

API keys allow external applications to authenticate with SendDock's API. Each key is scoped to a single project.

## Creating a Key

Go to **Settings** in the project sidebar, find the **API Keys** section, and click **+ Create Key**. Give it a descriptive name.

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
curl https://your-instance.com/api/v1/projects/{id}/subscribers \
  -H "Authorization: Bearer sk_your_full_key_here"
```

## Key Scope

An API key grants access to all operations within its project: subscribers, templates, email sending, and stats. It does not grant access to other projects or account-level operations.

## Last Used

SendDock tracks when each key was last used. Check this in Settings to identify unused keys.

## Revoking a Key

Click **Revoke** next to the key in Settings. This is immediate and permanent. Any application using that key will stop working.

## API

See [API Keys API](/api/api-keys) for the REST API reference.
