# Analytics

Analytics turns the raw `email_logs` table into something you can read at a glance — sends, opens, clicks, bounces, subscriber growth and engagement, broken into tabs and drawn as real charts.

It lives under each project at **Project → Analytics**, reads the same data the Logs view shows (just rolled up), and is **free** — part of Core, no license required. The one exception is the **Deliverability** tab, which is a Pro feature (see below).

![Analytics overview](/screenshots/analytics-overview.png)

## Tabs

The dashboard is split into five tabs. The first four are open; the last is Pro.

| Tab | What it shows |
|---|---|
| **Overview** | Headline KPIs (sent, failed, opened, clicked, open rate, click rate, acceptance rate, spam rate), each with a trend pill vs the previous period; an opens-and-clicks time series; a send-status donut; and a *Broadcasts in flight* panel that appears live while a large send is running. |
| **Campaigns** | A per-broadcast breakdown. Every broadcast you run — from the UI or via `/broadcast` — shows up here as one row (a scheduled [campaign](/guide/campaigns) is just a broadcast sent later, by the same worker), with sent / opened / clicked and rates. |
| **Audience** | Subscriber growth over the window — sign-ups and unsubscribes over time, drawn from each subscriber's `created_at` / `unsubscribed_at`. |
| **Engagement** | A Sent → Opened → Clicked funnel, an opens/clicks series you can toggle between, a device and mail-client breakdown read from the click user-agent, and a weekday × hour **heatmap** of when your audience actually clicks. |
| **Deliverability** <Badge type="warning" text="Pro" /> | Domain health (SPF/DKIM/DMARC) and a per-provider breakdown with acceptance, bounce, open, click and spam rates. See [Deliverability](/guide/deliverability). |

The Overview tab's **acceptance rate** and **spam rate** are worth defining up front: acceptance rate = sends your SMTP relay accepted ÷ every send attempt; spam rate = spam complaints ÷ delivered. "Delivered" is not a separate log status — it means the log rows that settle in `sent` (accepted by the relay, and not later marked `bounced`). The send-status donut buckets by email-log status, with `bounced` and `failed` shown as separate slices.

All the chart math runs server-side; the dashboard just renders what each tab's endpoint returns.

The **Audience** tab charts list growth — sign-ups and unsubscribes over time:

![The Audience tab — subscriber growth over time](/screenshots/analytics-audience.png)

The **Engagement** tab shows the open/click funnel, a device and mail-client breakdown, and a weekday × hour click heatmap:

![The Engagement tab — funnel, devices and mail clients](/screenshots/analytics-engagement.png)

## Date ranges

A toolbar at the top lets you pick the window, and it applies to every tab:

| Preset | Range |
|---|---|
| 24h | Last 24 hours, hourly buckets |
| 7d | Last 7 days, daily buckets |
| 30d | Last 30 days, daily buckets (default) |
| 90d | Last 90 days, daily buckets |
| 1y | Last 365 days, weekly buckets |
| Custom | Any pair of dates you choose |

Picking **Custom** opens a small popover with `From` / `To` date inputs. The bucket granularity is decided server-side from the range length — a 60-day custom window gives daily buckets, a 200-day window gives weekly — so you never have to think about it.

## Segment filter

If the project has any [segments](/guide/segments), a dropdown next to the date presets scopes the whole dashboard to one of them. Pick a segment and every metric recomputes over just the subscribers that match it; switch back to *All subscribers* to see the project as a whole. The only panel that ignores the filter is *Broadcasts in flight*, since it tracks send queues rather than per-subscriber engagement.

## Newsletter scope

Projects with [newsletters](/guide/newsletters) get a second dropdown: **All project / \<newsletter\>**. It scopes the Overview, Campaigns and Engagement tabs to sends attributed to that newsletter, and combines with the segment filter. The Audience tab stays project-wide — it describes your subscriber base, not a send stream. Only emails sent after newsletters existed carry the attribution, so older history shows up under *All project* only.

## Trends

On the Overview tab, each KPI's pill compares the current window to a same-length window immediately before it. With the **30d** preset that reads "vs previous 30d"; with **Custom** it collapses to "vs previous period".

Direction (▲/▼) and colour follow whether the change is good or bad for that metric:

- Sent / Opened / Clicked / Open rate / Click rate / Acceptance — up is good (emerald), down is bad (red).
- Failed / Spam rate — up is bad (red), down is good (emerald).
- A change under ±0.5% renders as flat (zinc).

## Reading the numbers

Analytics is read-only. There's no setup — every email already passes through the logs and tracking pipelines, and Analytics just queries them on demand. As soon as you have sends, opens or clicks, the tabs populate.

If you don't see numbers you expect:

1. Confirm your public URL is set and reachable (dashboard → **Instance** — see [Instance settings](/guide/instance-settings)). Without it, the open-tracking pixel and click-redirect URLs in your emails point to a host recipients cannot reach, and no events are recorded.
2. Image proxies (Gmail's, Outlook's image cache) often pre-fetch the open pixel once on receipt, which inflates opens slightly. SendDock counts only the **first** open per email, so the inflation is bounded. The Engagement heatmap is built from **clicks**, not opens, precisely because clicks are far less affected by this.
3. Click events are recorded only for links that go through the tracked redirect (`/c/{logId}/{...}`). Emails built in the Email Editor get this automatically; for raw HTML sends, see [Email Sending → Click tracking](/guide/sending#click-tracking).

## Exporting

The **Campaigns** tab has an **Export CSV** button that downloads the per-campaign breakdown for the current date range — handy for a spreadsheet or a report. It respects the window and segment filter you have selected.

## API

Each tab is backed by its own endpoint under `…/analytics/…` (overview, campaigns, audience, engagement), plus the CSV export above. See the [Analytics API reference](/api/analytics) for the full payload schemas.
