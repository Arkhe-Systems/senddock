# Reports <Badge type="warning" text="Pro" />

The Analytics tabs answer fixed questions. **Reports** lets you ask your own — a small pivot-table-and-chart studio over your data. Pick what to measure, how to group it, how to filter it and how to draw it, and get a live answer you can save and export.

It's a Pro section in the project sidebar: **Project → Reports**.

![The report builder showing email volume by provider, split by status, as a stacked bar](/screenshots/reports.png)

## Building a report

A report is five choices, and the preview re-runs live as you change them:

1. **Dataset** — what you're counting.
   - **Subscribers** — your audience.
   - **Email events** — everything in the send logs.
2. **Measure** — the number.
   - Subscribers: a count.
   - Email events: count, or a rate (open, click, bounce, spam).
3. **Dimension(s)** — how to group. One dimension gives a breakdown; **two gives a pivot** (rows × columns).
   - Subscribers: status, **tag**, a **custom field** (`custom.<key>`), or sign-up time.
   - Email events: status, **provider**, template, campaign, or send time (day / week / month).
4. **Filter** *(optional)* — scope to a [segment](/guide/segments). For the email dataset this limits events to that segment's members.
5. **Visualization** — table, pivot matrix, bar, stacked bar, line, area, donut or pie.

A quick example: *Subscribers, count, by `custom.plan`* as a donut answers "how many subscribers do I have on each plan?" Add a second dimension — *by `provider` × send time* — and it becomes a pivot.

## Visualizations

The same query renders as whatever chart fits the question — switch it live from the **Chart** dropdown. A single dimension can be a **bar**, **donut**, **pie**, **line**, **area** or plain **table**; two dimensions unlock the **pivot matrix** and stacked/multi-series charts.

**Donut** — a composition, e.g. subscribers by plan:

![A report drawn as a donut chart](/screenshots/reports-donut.png)

**Area / line** — a multi-series trend over time. Add a second dimension and each value becomes its own series — here, sends per provider, week by week:

![A multi-series area chart, one series per provider over time](/screenshots/reports-area.png)

**Pivot** — a two-dimension matrix, e.g. open rate by provider across months:

![A two-dimension report drawn as a pivot table](/screenshots/reports-pivot.png)

The **stacked bar** at the top of this page (email volume per provider, split by status) and the **pie** and raw **table** round out the set. Pick the one that makes your point.

## Saving & exporting

- **Save** a report (name + its configuration) to pin a question you'll ask again; saved reports list in the panel and reload with one click.
- **Export CSV** downloads the current result — the breakdown or the full pivot — for a spreadsheet or a deck.

## How it stays safe

The builder never runs free-form SQL. Every dimension and measure maps to a fixed, vetted SQL expression from an allowlist; the only user-supplied value that reaches a query is a custom-field **key**, and that's validated against `^[a-zA-Z0-9_]+$` before use. Filters reuse the same predicate engine as [Segments](/guide/segments). So you get the flexibility of an ad-hoc report without opening an injection surface.

## Licensing

Reports is Pro. Without a license the section renders a paywall. See [Instance settings → Pro license](/guide/instance-settings#pro-license).
