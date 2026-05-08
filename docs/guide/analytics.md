# Analytics <Badge type="warning" text="Pro" />

The Analytics section is a Pro-gated dashboard that turns the raw `email_logs` table into something you can read at a glance — funnel, opens-over-time, top templates, top clicked links, and a small panel of auto-generated insights.

It lives under each project at **Project → Analytics**, and reads the same data that the Logs view shows, just rolled up.

## What it shows

Open `https://your-instance.com/projects/{id}/analytics` to get:

- **Six top cards** — sent, failed, opened, clicked, open rate, click rate. Each card carries a trend pill comparing the current period against the same-length previous period (▲/▼ with delta percentage).
- **Conversion funnel** — Sent → Delivered → Opened → Clicked, with deliverability/open-rate/click-rate percentages stacked on top.
- **Insights panel** — short bullet points generated from your data: best day this period, open rate vs the ~21% industry benchmark, high failure rates worth investigating, most-used template, and an empty-state hint when there are no sends yet.
- **Opens over time** — a smoothed area chart whose granularity adapts to the range you pick (hourly under 24h, daily for ≤90d, weekly for ≤1y, monthly for longer).
- **Top templates** and **Top clicked links** — bar lists ranked by sends and unique URL clicks respectively.
- **Send status donut** — sent vs failed split, plus the total active subscribers count.
- **Broadcasts in flight** — a live panel that appears at the top of the dashboard whenever there is at least one broadcast actively sending. It shows a progress bar per broadcast (`X / total`, percentage), elapsed time, and per-status counters (sent / failed / suppressed / pending). The panel polls `/analytics/overview` every five seconds while any broadcast is in flight and disappears as soon as the queue drains. Useful for watching large sends to 50k+ subscriber lists without leaving Analytics.

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

- An empty `SENDDOCK_LICENSE_KEY` leaves Pro locked. The Analytics dashboard returns `402 Payment Required`, which the UI renders as a paywall card linking to pricing. This applies to both self-hosted and cloud deployments — Core stays fully usable, Pro requires a license.
- A valid license unlocks Analytics. The validator caches its last successful result for 24 hours, so brief network blips with Lemon Squeezy don't lock you out.

See [Configuration → Pro license](/self-hosting/configuration#plans-and-licensing) for the env var.
