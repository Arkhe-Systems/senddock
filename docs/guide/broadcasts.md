# Broadcasts

A **broadcast** is a single send to many recipients at once — what happens when you send to *All subscribers* (or a [segment](./segments), or a [newsletter](./newsletters)) from the [Send Email](./sending) composer, or when a scheduled [campaign](./campaigns) fires. The **Broadcasts** tab is the history of those sends, with live progress while one is running.

Open it from **Broadcasts** in the project sidebar.

![The broadcasts history with per-send progress](/screenshots/broadcasts.png)

## The history table

Each row is one broadcast:

| Column | Meaning |
|---|---|
| **Subject** | The subject line that went out. |
| **Audience** | What the broadcast targeted: *All active subscribers*, a [newsletter](./newsletters) name, or a [segment](./segments) name. |
| **Status** | `sending` while in flight, `completed` when the queue drains. |
| **Progress** | A live bar and an `X / total` count that climbs in real time while sending. |
| **Started** | When the send began. |

The table filters by **status** and by **newsletter** (the selects appear top-right once the project has newsletters).

Rows **expand** (the ▸ chevron) to reveal the full breakdown: **sent / failed / suppressed / total**, how long it took, and the broadcast's ID. While a broadcast is in flight the panel polls every few seconds so you can watch a large list drain without leaving the page.

## How a broadcast actually runs

Broadcasts aren't sent inline — they go through a persistent, per-recipient queue with retries and crash-recovery, so a server restart mid-send never double-sends or drops anyone. The mechanics (worker pool, retry backoff, resumability) are documented in [Sending → How a broadcast actually runs](./sending#how-a-broadcast-actually-runs).

## Related

- [Email Sending](./sending) — how to start a broadcast and the queue internals.
- [Campaigns](./campaigns) — scheduled broadcasts.
- [Analytics → Campaigns tab](./analytics#tabs) — per-broadcast opens, clicks and rates.
