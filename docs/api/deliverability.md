# Deliverability API <Badge type="warning" text="Pro" />

Domain-health checks and per-provider send breakdowns that back the [Deliverability](../guide/deliverability) tab. Pro-gated: without a valid Pro license these endpoints return `402 Payment Required`.

Cookie auth only — like the rest of the analytics surface, these read project-scoped aggregates and require membership of the project's workspace, so a project API key can't call them.

## Domain health

```
GET /api/v1/projects/{id}/deliverability/domain-health
```

Resolves the SPF, DKIM and DMARC DNS records for the project's sending domain and grades each.

### Response

```json
{
  "domain": "acme.com",
  "checks": [
    { "name": "SPF",   "status": "pass", "detail": "SPF present with a strict -all policy", "value": "v=spf1 include:..." },
    { "name": "DKIM",  "status": "warn", "detail": "No DKIM key found for the common selectors", "fix": "Publish a DKIM key with your sending provider (or use a custom selector we can't auto-detect)." },
    { "name": "DMARC", "status": "fail", "detail": "No DMARC record found", "fix": "Add a TXT record at _dmarc: v=DMARC1; p=none; rua=mailto:you@your-domain" }
  ]
}
```

`name` is `SPF`, `DKIM` or `DMARC`. `status` is one of `pass` / `warn` / `fail`. `value` carries the record that was read and `fix` the suggested action; both are omitted when empty, so a pass has no `fix`.

## Per-provider breakdown

```
GET /api/v1/projects/{id}/deliverability/providers?from=...&to=...
```

Groups the project's email logs by mailbox provider (inferred from the recipient domain) and computes volumes and rates per provider. Accepts the same `from` / `to` window as the [Analytics](./analytics) endpoints.

### Response

```json
{
  "from": "2026-05-31T00:00:00Z",
  "to": "2026-06-30T00:00:00Z",
  "total_bounced": 41,
  "providers": [
    {
      "provider": "Gmail",
      "sent": 4120, "failed": 4, "bounced": 33, "opened": 1890, "clicked": 402, "complained": 3,
      "hard_bounces": 20, "soft_bounces": 13,
      "acceptance_pct": 99.11, "bounce_rate_pct": 0.79,
      "open_rate_pct": 45.87, "click_rate_pct": 9.76, "complaint_rate_pct": 0.07
    }
  ]
}
```

The envelope echoes the resolved `from` / `to` window and `total_bounced`, the bounced rows in it.

| Field | Meaning |
|---|---|
| `provider` | `Gmail`, `Outlook`, `Yahoo`, `Apple`, `AOL`, `Proton`, or `Other` (`Unknown` when the log row carries no recipient domain). |
| `failed` | Emails rejected at send time. |
| `hard_bounces` / `soft_bounces` | Bounce split, classified from the bounce reason text. |
| `complained` | Spam complaints reported for this provider via the [complaint webhook](../guide/deliverability#wiring-the-complaint-webhook) (FBL). |
| `acceptance_pct` | Sent ÷ attempted (`sent + failed + bounced`). |
| `bounce_rate_pct` | Bounced ÷ attempted. |
| `open_rate_pct` / `click_rate_pct` | Opened / clicked ÷ sent. |
| `complaint_rate_pct` | Complaints ÷ sent — 0 unless a [complaint webhook](../guide/deliverability#wiring-the-complaint-webhook) is feeding you FBL data. |

### Errors

| Status | Cause |
|---|---|
| `400` | Domain health only: the project has no from address, so there is no sender domain to resolve. |
| `401` | Missing cookie session. |
| `402` | No valid Pro license. |
| `403` | The authenticated user isn't a member of the project's workspace (also returned for a project that doesn't exist). |
| `404` | The project id in the path isn't a valid UUID. |

## See also

- [Deliverability guide](../guide/deliverability) — how to read domain health and the per-provider table.
- [Bounces](./bounces) — the bounce webhook that feeds the bounce numbers. The complaint webhook (`POST /webhooks/complaints/{projectId}`) is documented in the [Deliverability guide](../guide/deliverability#wiring-the-complaint-webhook).
