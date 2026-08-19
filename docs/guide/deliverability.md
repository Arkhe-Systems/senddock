# Deliverability <Badge type="warning" text="Pro" />

Deliverability is a Pro tab inside **Project → Analytics**. Where the free tabs tell you *what happened* to your sends, Deliverability tells you *why* — whether your domain is set up to be trusted, and how each mailbox provider is actually treating your mail.

::: tip Complaint handling is Core, not gated
Ingesting spam complaints (the FBL/complaint webhook `POST /webhooks/complaints/{projectId}`), **auto-suppressing** the complainer, and the **complaint rate** shown in [Analytics](/guide/analytics) are all part of **Core** — reputation-critical work is never behind the license. This Pro tab layers the *per-provider* breakdown and domain-health analysis on top.
:::

![Deliverability tab](/screenshots/deliverability.png)

It has two halves: **domain health** and a **per-provider breakdown**.

## Domain health

SendDock looks up the DNS records that inboxes check before trusting you, and grades each one **pass / warn / fail** with a short fix when something's off:

| Record | What it checks |
|---|---|
| **SPF** | A `TXT` record on your sending domain that authorizes SendDock's SMTP host to send for you. |
| **DKIM** | A public key that lets providers verify your messages weren't tampered with. SendDock probes the common selectors; if your provider uses an unusual selector it may show as *warn* even when configured. |
| **DMARC** | A `_dmarc` `TXT` policy that tells inboxes what to do with mail that fails SPF/DKIM, and where to send reports. |

The checks read live DNS, so fixing a record and re-opening the tab reflects it within your resolver's TTL. All three passing is the single biggest lever on whether you land in the inbox.

## Per-provider breakdown

Recipients are grouped by mailbox provider — Gmail, Outlook, Yahoo, Apple, and everything else as *Other* — inferred from the recipient domain. For each provider you get:

| Column | Meaning |
|---|---|
| **Sent / Bounced** | Volume and how much came back. |
| **Acceptance rate** | Accepted ÷ attempted — the relay-level counterpart to bounce rate. |
| **Bounce rate** | Split into **hard** (permanent — bad address) and **soft** (transient — full mailbox, greylisting) using the bounce reason text. |
| **Open / Click rate** | Engagement per provider. |
| **Spam rate** | Complaints ÷ delivered (see below). |

Reading providers separately matters because they behave differently: a bounce rate that looks fine overall can hide one provider quietly rejecting you. Sudden divergence on one provider is your earliest warning that something in your setup or list quality is off.

## Spam-complaint rate

When a recipient hits "report spam", their provider can notify the sender through a **feedback loop (FBL)**. SendDock ingests those via a complaint webhook, records a complaint on the original log (without changing its delivered status), and surfaces the resulting **spam rate** both on the Overview tab and per provider here.

Keep it well under **0.1%** — Gmail and others start throttling above roughly 0.3%.

### Wiring the complaint webhook

It mirrors the [bounce webhook](/guide/bounces#webhook). Each project has a signed URL:

```
POST https://your-instance.com/webhooks/complaints/{projectId}
```

Point your provider's or FBL's complaint notifications at it. SendDock matches the complaint to the original recipient, marks the log complained, and adds the address to the [suppression list](/guide/suppressions) so you stop mailing them. Without a complaint feed the spam rate simply reads 0% — it can't see reports it was never told about.

## Licensing

Deliverability is Pro. Without a license the tab renders a paywall; the free Analytics tabs are unaffected. See [Instance settings → Pro license](/guide/instance-settings#pro-license).
