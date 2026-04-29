# Analytics <Badge type="warning" text="Pro" />

The Analytics section is a Pro-gated dashboard that turns the raw `email_logs` table into something you can read at a glance — funnel, opens-over-time, top templates, top clicked links, and a small panel of auto-generated insights.

It lives under each project at **Project → Analytics**, and reads the same data that the Logs view shows, just rolled up.

## What it shows

<img src="/screenshots/analytics-dashboard.png" alt="Analytics dashboard with metric cards, conversion funnel, opens-over-time chart, top templates and top clicked links" style="width:100%;max-width:900px;margin:1rem 0;border-radius:12px;border:1px solid rgba(120,120,128,0.25);" />

Open `https://your-instance.com/projects/{id}/analytics` to get:

- **Six top cards** — sent, failed, opened, clicked, open rate, click rate. Each card carries a trend pill comparing the current period against the same-length previous period (▲/▼ with delta percentage).
- **Conversion funnel** — Sent → Delivered → Opened → Clicked, with deliverability/open-rate/click-rate percentages stacked on top.
- **Insights panel** — short bullet points generated from your data: best day this period, open rate vs the ~21% industry benchmark, high failure rates worth investigating, most-used template, and an empty-state hint when there are no sends yet.
- **Opens over time** — a smoothed area chart whose granularity adapts to the range you pick (hourly under 24h, daily for ≤90d, weekly for ≤1y, monthly for longer).
- **Top templates** and **Top clicked links** — bar lists ranked by sends and unique URL clicks respectively.
- **Send status donut** — sent vs failed split, plus the total active subscribers count.

All the chart math runs server-side. The dashboard is a thin renderer over a single `/analytics/overview` call.

## Date ranges

A toolbar at the top of the page lets you pick the window:

| Preset | Range |
|---|---|
| 24h | Last 24 hours, hourly buckets |
| 7d | Last 7 days, daily buckets |
| 30d | Last 30 days, daily buckets (default) |
| 90d | Last 90 days, daily buckets |
| 1y | Last 365 days, weekly buckets |
| Custom | Any pair of dates you choose |

The "Custom" option opens a small popover with two date inputs (`From` / `To`). Picking Custom triggers a re-fetch and the trend pills then read "vs previous period" — same length as your selection.

The granularity is decided server-side from the range length, so picking a custom 60-day window gives you daily buckets but a custom 200-day window gives you weekly. You don't need to think about it.

## Trends

Each card's pill compares the current window to a same-length window immediately before it. So with the **30d** preset selected, "vs previous 30d" compares this month to the previous month. With **Custom** the comparison label collapses to "vs previous period".

Direction (▲/▼) and color (emerald/red/zinc) follow whether the change is good or bad for that metric:

- Sent / Opened / Clicked / Open rate / Click rate — up is good (emerald), down is bad (red).
- Failed — up is bad (red), down is good (emerald).
- A change under ±0.5% renders as flat (zinc).

## Authoring the page

Analytics is read-only. There's no setup — every email already passes through the logs and tracking pipelines, and Analytics just queries them on demand. As soon as you have sends, opens, or clicks, this page populates.

If you don't see numbers you expect:

1. Confirm `PUBLIC_URL` is set and reachable. Without it, the open-tracking pixel and click-redirect URLs in your emails point to a host the recipient cannot reach, and no events are recorded.
2. Image proxies (Gmail's, Outlook's image cache) often pre-fetch the open pixel once on receipt, which inflates the open count slightly. SendDock counts only the **first** open per email, so the inflation is bounded.
3. Click events are recorded only for links that go through the tracked redirect (`/c/{logId}/{...}`). Newsletters built in the Email Editor get this automatically; for raw HTML sends, see [Email Sending → Click tracking](/guide/sending#click-tracking).

## API

The Analytics page is one HTTP call:

```
GET /api/v1/projects/{id}/analytics/overview?from=...&to=...
```

See the [Analytics API reference](/api/analytics) for the full payload schema.

## Licensing

Analytics is gated by `SENDDOCK_LICENSE_KEY`:

- **Self-hosted** with empty license — Pro features are unlocked locally for development. Analytics works.
- **Self-hosted** with valid license — Pro unlocked.
- **Cloud** mode (the senddock.dev managed product) requires a valid key. An empty key returns `402 Payment Required` on `/analytics/overview`, which the UI renders as a paywall card linking to pricing.

See [Configuration → Pro license](/self-hosting/configuration#pro-license) for the env var.
