# Email Logs

Every email SendDock sends — transactional, broadcast or campaign — is recorded in the project's **Logs**. It's the source of truth for "what happened to that message", and where opens, clicks and bounces surface per recipient.

Open it from **Logs** in the project sidebar.

![The email logs with the status filter](/screenshots/bounces.png)

## What each row shows

| Column | Meaning |
|---|---|
| **To** | The recipient address. |
| **Subject** | The rendered subject line. |
| **Status** | `sent`, `failed`, `bounced` or `suppressed` (see [Sending → statuses](./sending) and [Bounces](./bounces)). |
| **Engagement** | Whether the message was opened and/or clicked. |
| **Date** | When it was processed. |

Click a row to open its **detail drawer** — the full metadata for that send, including the template used, the broadcast it belonged to (if any), and the open/click timeline.

## Filtering

A filter bar sits above the table:

- **Status chips** — All / Sent / Failed / Bounced / Suppressed.
- **Search** — match by recipient email or subject.
- **Template** — narrow to a single template.
- **From / To** — a date range.
- **Clear** — reset every filter at once.

Filters combine, so "Bounced + last 7 days + a given template" is one click each.

## Export

**Export CSV** (top right) downloads the currently-filtered logs — handy for auditing a specific send or handing a bounce list to someone. The export honours whatever filters you have applied.

## API

The same data is available programmatically at `GET /api/v1/projects/{id}/logs` (with the same filter query params) and `GET /logs/{logId}` for a single row. The CSV export maps to `/logs/export.csv`. There's no separate "logs" API page — the shapes live alongside [Sending](/api/sending).
