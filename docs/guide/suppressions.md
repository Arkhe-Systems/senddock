# Suppression list

Each project has its own **suppression list** — a list of email addresses that should never receive sends from that project, even if you ask SendDock to send to them. It's the single source of truth that `/send`, `/send/batch` and `/broadcast` consult before every send.

## How it works

When you send to an address, SendDock first checks the suppression list. If the address is on it:

- the send is **not** attempted,
- the email log records the recipient as `suppressed` (a distinct outcome from `sent`, `failed`, `bounced`),
- batch and broadcast results show `suppressed` count separately, so you can see exactly how many recipients were skipped.

The list is **per-project**. Suppressing `user@example.com` in Project A does not affect Project B — different projects have different audiences and you may want them to behave independently.

## What populates the list

| Source | Reason | When |
|---|---|---|
| **Hard bounce** (in-session 5xx, webhook ingest, IMAP poll) | `hard_bounce` | Automatically, see [Bounces](./bounces). |
| **Unsubscribe** (one-click or manual) | `unsubscribed` | When a subscriber clicks unsubscribe or you flip their status. |
| **Manual add** | `manual` (or whatever reason you provide) | From the Suppressions tab or `POST /projects/{id}/suppressions`. |
| **Bulk import** | `manual` | Pasting a list in the Suppressions tab or sending an array to the API. |

Existing unsubscribed subscribers from before v0.6 were backfilled into the suppression list when you upgraded.

## Managing entries

In the project's **Suppressions** tab you can:

- **Filter** by reason (`hard_bounce`, `unsubscribed`, `manual`, ...).
- **Add** a single address with a reason.
- **Bulk import** a list of addresses (one per line or comma-separated).
- **Remove** an entry — useful if you confirmed the bounce was a typo or the user wants back in.

## API

### List

```bash
curl -H "Authorization: Bearer sk_..." \
  "https://your-instance.com/api/v1/projects/{project_id}/suppressions?reason=hard_bounce"
```

### Add

```bash
curl -X POST https://your-instance.com/api/v1/projects/{project_id}/suppressions \
  -H "Authorization: Bearer sk_..." \
  -H "Content-Type: application/json" \
  -d '{"emails": ["user@example.com", "other@example.com"], "reason": "manual"}'
```

Returns the count of newly suppressed addresses (already-suppressed addresses are silently de-duplicated).

### Remove

```bash
curl -X DELETE https://your-instance.com/api/v1/projects/{project_id}/suppressions/{suppression_id} \
  -H "Authorization: Bearer sk_..."
```

## Interaction with broadcast

For a broadcast to a 10,000-subscriber list where 800 are suppressed, the result looks like:

```json
{
  "queued": 9200,
  "suppressed": 800,
  "broadcast_id": "01H..."
}
```

Suppressed recipients never enter the queue. They don't count against rate limits, they don't generate webhook events, and they don't appear in the email log as `failed` — they appear as `suppressed`, which is the honest answer.

## Why per-project

Two projects under the same workspace can be a transactional product (where a hard bounce should be treated as a permanent block) and a marketing newsletter (where the bounce list is huge but should never affect transactional sends). Sharing a single global list would conflate them. Per-project suppression keeps each list focused on the audience that project actually serves.

## See also

- [Bounces](./bounces) — the three sources that automatically populate the list.
- [Sending](./sending) — how `suppressed` shows up in send results.
- [Audit log (Pro)](./audit-log) — `suppression.add` and `suppression.delete` are tracked actions.
