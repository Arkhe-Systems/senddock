# Audit log API <Badge type="warning" text="Pro" />

Read the per-project audit log — every sensitive action with actor, timestamp and metadata. Pro-gated: without a valid `SENDDOCK_LICENSE_KEY` this endpoint returns `402 Payment Required`. See the [Audit log guide](../guide/audit-log) for the catalog of tracked actions.

Cookie auth only on this endpoint — the audit log isn't readable via API key, the actor identity is mandatory and a project-scoped API key has none.

## List audit entries

```
GET /api/v1/projects/{id}/audit-log
```

### Query parameters

| Parameter | Type | Description |
|---|---|---|
| `limit` | int (default `50`, max `200`) | Page size. |
| `offset` | int (default `0`) | Pagination offset. |
| `action` | string | Filter to a single action string (e.g. `smtp.update`, `api_key.revoke`). |
| `from` | RFC 3339 timestamp | Inclusive lower bound on `created_at`. |
| `to` | RFC 3339 timestamp | Exclusive upper bound on `created_at`. |

### Response

```json
{
  "entries": [
    {
      "id": "01H...",
      "user_id": "01H...",
      "action": "smtp.update",
      "target_type": "project",
      "target_id": "01H...",
      "metadata": {
        "smtp_host": "smtp.mailgun.org",
        "smtp_user": "postmaster@acme.com",
        "from_email": "hello@acme.com"
      },
      "ip_address": "203.0.113.42",
      "user_agent": "Mozilla/5.0 ...",
      "created_at": "2026-04-30T14:22:11Z"
    }
  ],
  "total": 1842
}
```

| Field | Description |
|---|---|
| `id` | Unique entry id (ULID). |
| `user_id` | UUID of the user who took the action. Resolve to email via [`GET /workspaces/{id}/members`](./workspaces#list-members). |
| `action` | Dotted action string. See the [full catalog](../guide/audit-log#what-gets-recorded). |
| `target_type` / `target_id` | The entity acted on — `project`, `api_key`, `webhook`, `suppression`, `workspace`. |
| `metadata` | Action-specific JSON. SMTP and bounce-IMAP entries record host / user / from address; never password fields. |
| `ip_address` | IP of the request that triggered the action, parsed from `X-Forwarded-For` if present. |
| `user_agent` | Browser / client UA. |
| `created_at` | UTC, ISO 8601. |

### Errors

| Status | Cause |
|---|---|
| `401` | Missing cookie session. |
| `402` | No valid Pro license key. |
| `403` | The authenticated user doesn't own this project. |
| `404` | Project not found. |

### Examples

```bash
# Last 50 entries
curl "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/audit-log" \
  -b cookies.txt
```

```bash
# Just SMTP changes in the last 7 days
curl -G "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/audit-log" \
  -b cookies.txt \
  --data-urlencode "action=smtp.update" \
  --data-urlencode "from=$(date -u -d '7 days ago' +%Y-%m-%dT%H:%M:%SZ)"
```

## Retention

Entries are kept indefinitely. There is no built-in pruning — at typical install volumes 12 months of changes is a few thousand rows. If you want a rolling window, run from `psql`:

```sql
DELETE FROM audit_log WHERE created_at < now() - interval '180 days';
```

## See also

- [Audit log guide](../guide/audit-log) — full action catalog and what each metadata field carries.
- [Suppressions API](./suppressions) — `suppression.add` and `suppression.delete` are recorded here.
- [Webhooks API](./webhooks) — `webhook.create` and `webhook.delete` are recorded here.
