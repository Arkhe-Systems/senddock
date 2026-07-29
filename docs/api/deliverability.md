# Deliverability API <Badge type="warning" text="Pro" />

Domain-health checks and per-provider send breakdowns that back the [Deliverability](../guide/deliverability) tab. Pro-gated: without a valid Pro license these endpoints return `402 Payment Required`.

Cookie auth only — like the rest of the analytics surface, these read project-scoped aggregates and require an authenticated owner, so a project API key can't call them.

## Domain health

```
GET /api/v1/projects/{id}/deliverability/domain-health
```

Resolves the SPF, DKIM and DMARC DNS records for the project's sending domain and grades each.

### Response

```json
{
  "domain": "acme.com",
  "from_email": "hello@acme.com",
  "checks": [
    { "record": "spf",   "status": "pass", "detail": "v=spf1 include:...", "fix": "" },
    { "record": "dkim",  "status": "warn", "detail": "No key on the common selectors", "fix": "Publish your DKIM key, or confirm the selector your relay uses." },
    { "record": "dmarc", "status": "fail", "detail": "No _dmarc record", "fix": "Add a TXT record at _dmarc.acme.com, e.g. v=DMARC1; p=none; rua=..." }
  ]
}
```

`status` is one of `pass` / `warn` / `fail`. `fix` is empty on a pass.

## Per-provider breakdown

```
GET /api/v1/projects/{id}/deliverability/providers?from=...&to=...
```

Groups the project's email logs by mailbox provider (inferred from the recipient domain) and computes volumes and rates per provider. Accepts the same `from` / `to` window as the [Analytics](./analytics) endpoints.

### Response

```json
{
  "providers": [
    {
      "provider": "Gmail",
      "sent": 4120, "bounced": 33, "opened": 1890, "clicked": 402, "complained": 3,
      "hard": 20, "soft": 13,
      "acceptance_pct": 99.2, "bounce_pct": 0.8,
      "open_pct": 45.9, "click_pct": 9.8, "spam_pct": 0.07
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `provider` | `Gmail`, `Outlook`, `Yahoo`, `Apple`, or `Other`. |
| `hard` / `soft` | Bounce split, classified from the bounce reason text. |
| `acceptance_pct` | Accepted ÷ attempted. |
| `spam_pct` | Complaints ÷ delivered — 0 unless a [complaint webhook](../guide/deliverability#wiring-the-complaint-webhook) is feeding you FBL data. |

### Errors

| Status | Cause |
|---|---|
| `401` | Missing cookie session. |
| `402` | No valid Pro license. |
| `403` | Authenticated user doesn't own this project. |
| `404` | Project not found. |

## See also

- [Deliverability guide](../guide/deliverability) — how to read domain health and the per-provider table.
- [Bounces](./bounces) — the bounce and complaint webhooks that feed these numbers.
