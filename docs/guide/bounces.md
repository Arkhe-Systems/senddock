# Bounces

A **bounce** is when an email cannot be delivered. SendDock detects bounces three different ways and feeds all of them into the same project [suppression list](./suppressions), so a recipient that hard-bounces once stops receiving sends — without you having to do anything.

## What gets classified as a bounce

| Bounce type | What happens |
|---|---|
| **Hard** (5xx — mailbox doesn't exist, domain doesn't exist, address rejected) | Recipient is added to the suppression list, log row is marked `bounced`, `email.bounced` webhook fires. |
| **Soft** (4xx — temporary deferral, mailbox full, greylisting) | Log row marked `failed`, **not** added to the suppression list. Retry on your side if you want. |

SendDock only suppresses on hard (`5xx`) bounces. Soft bounces are surfaced in the email log so you can decide what to do with them.

## How SendDock detects bounces

There are three sources, all enabled per project. Each one independently writes to the suppression list and dispatches webhooks.

### 1. In-session SMTP (5xx during RCPT TO)

This one is on by default and needs zero configuration. When SendDock talks to your SMTP server and the destination MTA rejects a recipient with a 5xx code on `RCPT TO`, SendDock catches the response in the same connection, classifies it, and:

- marks the email log row as `bounced`,
- adds the recipient to the project suppression list with reason `hard_bounce`,
- fires an `email.bounced` webhook (if configured).

This catches the cleanest case: addresses that don't exist on a major provider (Gmail, Outlook, your customer's company server) and respond synchronously.

### 2. Public webhook endpoint

For providers that do **out-of-band** bounce reporting (Mailgun, Postmark, SES via SNS, etc.), each project gets a public ingest URL:

```
POST https://<your-host>/webhooks/bounces/{project_id}?token=<bounce_token>
```

The endpoint accepts:

- **Generic JSON** — `{"recipient": "user@example.com", "reason": "..."}` (or an array of those).
- **Mailgun event payloads** — sent verbatim from a Mailgun event-data webhook with `event=permanent_failure` or `event=failed`.

Configure the destination in your provider's UI to point to that URL. The bounce token is shown in the project's **Settings → Bounces** tab and can be rotated from the same screen.

### 3. IMAP poller

For SMTP providers without webhooks (your own VPS, an old shared host), SendDock can scan a bounce mailbox over IMAP every 5 minutes.

In **Settings → Bounces**, configure:

- IMAP host + port (TLS only)
- Mailbox username + password
- Folder to scan (defaults to `INBOX`)

The poller logs in over TLS, scans for unread messages, and for each one:

1. Looks for an **RFC 3464 DSN** (`Final-Recipient: rfc822;user@example.com`) — the standardized bounce format.
2. Falls back to scanning the body for a `5xx` line and extracting the address there.
3. Adds the recipient to the suppression list and marks the IMAP message `\Seen` so it's not processed again.

The poller never deletes messages — it only flags them as read.

## Where bounces show up

| Where | What you see |
|---|---|
| **Email logs** | The send row's status changes from `sent` to `bounced`. |
| **Suppression list** | Bounced recipients appear with reason `hard_bounce`. |
| **Project stats** | `bounced` and `suppressed` are tracked separately from `sent` and `failed`. The dashboard's outcome cards show all four. |
| **Webhooks** | `email.bounced` fires once with `{ "log_id", "to", "reason" }`. |
| **Audit log (Pro)** | Bounce-mailbox config changes and bounce-token rotations are recorded. |

## Webhook payload

```json
{
  "id": "evt_01H...",
  "type": "email.bounced",
  "created_at": "2026-04-30T14:22:11Z",
  "data": {
    "log_id": "01H...",
    "project_id": "01H...",
    "to": "user@example.com",
    "subject": "Welcome",
    "reason": "550 5.1.1 mailbox unavailable"
  }
}
```

Signed with `X-SendDock-Signature: sha256=<hex>` over the raw body, using the webhook's secret. See [Webhooks](./webhooks#verifying-signatures).

## Why this matters

Hard-bouncing the same address twice is how senders get flagged by Gmail/Outlook reputation systems. SendDock's three detection sources plus the suppression list mean once an address bounces, it never gets sent to again from that project — so your sender reputation stays clean without manual list hygiene.

## Resending after a bounce

If you've fixed the underlying issue (typo in the address, recipient was on vacation responder, etc.), remove the entry from **Suppressions** in the UI or via `DELETE /projects/{id}/suppressions/{id}`. The next send to that recipient will go through.

## See also

- [Suppression list](./suppressions) — how skipped recipients are tracked and managed.
- [Webhooks](./webhooks) — full event reference and signature verification.
- [Sending](./sending) — how `/send`, `/send/batch` and `/broadcast` interact with bounces and suppressions.
