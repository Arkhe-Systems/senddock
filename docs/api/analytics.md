# Analytics API

The endpoints behind the [Analytics dashboard](/guide/analytics). Cookie auth required.

Analytics is **free** — part of Core, no license needed. (The one exception is the Deliverability tab, which has its own [Pro endpoints](./deliverability).) The tabs map to sibling endpoints under `/api/v1/projects/{id}/analytics/`: `/overview` (documented below), `/campaigns`, `/campaigns/{broadcastId}` (a single broadcast's breakdown), `/audience`, and `/engagement`, plus a CSV export at `/analytics/export`. Each accepts the same `from` / `to` window and the `segment_id` / `newsletter_id` filters described below.

## Overview

```
GET /api/v1/projects/{id}/analytics/overview?from=...&to=...&segment_id=...&newsletter_id=...
```

| Query | Required | Format | Description |
|---|---|---|---|
| `from` | no | RFC 3339 | Start of the window (UTC). Defaults to 30 days before the current time. |
| `to` | no | RFC 3339 | End of the window (UTC). Defaults to the current time. |
| `granularity` | no | string | Bucket size for `opens_series` and `clicks_series` — one of `hour`, `day`, `week`, `month`, `auto`. Defaults to `auto`, which picks one from the range length. Any other value returns `400`. |
| `range_days` | no | int | Convenience alternative to `from`: start the window this many days before the current time. Values from `1` to `3650` are accepted; anything else is ignored. `from` wins when both are set. |
| `segment_id` | no | UUID | Restrict every metric to subscribers matching this [segment](/api/segments). Omit for all subscribers. Returns `404` if the segment doesn't exist in the project. |
| `newsletter_id` | no | UUID | Restrict every metric to sends attributed to this [newsletter](/api/newsletters). Combinable with `segment_id`. A malformed UUID returns `400`. |

The bucket granularity is `granularity` when the client sends one, and otherwise derived from the range length. The `previous` block in the response covers the same-length window immediately before `from`, used for trend comparisons. When `segment_id` is set it is echoed back in the response and applied to every metric except *broadcasts in flight* (which is not per-subscriber). `newsletter_id` is echoed back the same way; only emails sent after the newsletter feature shipped carry the attribution, so older sends fall outside any newsletter scope. The campaign-stats and engagement endpoints backing the dashboard's other tabs accept the same `newsletter_id` parameter.

**Response**

```json
{
  "from": "2026-03-30T00:00:00Z",
  "to":   "2026-04-29T00:00:00Z",
  "granularity": "day",
  "range_days": 30,

  "total_sent": 1500,
  "total_failed": 20,
  "total_opened": 980,
  "total_clicked": 142,

  "acceptance_pct": 98.7,
  "open_rate_pct": 65.3,
  "click_rate_pct": 9.5,
  "click_to_open_pct": 14.5,

  "active_subscribers": 4321,

  "opens_series": [
    { "bucket": "2026-04-28T00:00:00Z", "opens": 42 }
  ],

  "top_templates": [
    { "template_id": "uuid", "name": "Welcome", "sends": 320 }
  ],

  "top_clicked_links": [
    { "url": "https://example.com/launch", "clicks": 51 }
  ],

  "sends_by_status": [
    { "status": "sent",   "count": 1500 },
    { "status": "failed", "count":   20 }
  ],

  "previous": {
    "total_sent": 1410,
    "total_failed": 18,
    "total_opened": 905,
    "total_clicked": 130,
    "acceptance_pct": 98.7,
    "open_rate_pct": 64.2,
    "click_rate_pct": 9.2
  }
}
```

### Field notes

- All `total_*` and `*_pct` values cover the requested window.
- `opens_series`, `top_templates`, `top_clicked_links` and `sends_by_status` are always arrays — `[]` when the window has no data, never `null`.
- `click_to_open_pct` is `clicks / opens × 100` (engagement among readers), not a portion of total sends.
- `active_subscribers` is a project-wide snapshot (current count of `status=active`), not windowed.
- `previous` mirrors the headline metrics for the equivalent window immediately before `from`. If there is no data in that previous window, the values are `0`.
