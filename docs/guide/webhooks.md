# Webhooks <Badge type="warning" text="Pro" />

Webhooks let SendDock call your own HTTP endpoint every time something interesting happens in a project — an email is sent, a recipient opens it, a subscriber unsubscribes, and so on. They are the right way to keep a CRM, a usage table, or a Slack channel in sync without polling the API.

Webhook delivery, signing and retries ship in the open-source Core; the **management UI and API endpoints** (creating, listing, pausing, deleting webhooks) live in Pro and require a license in cloud mode.

## Events

SendDock emits six event types today:

| Type | When it fires |
|---|---|
| `email.sent` | Email handed to the SMTP relay successfully |
| `email.failed` | Email rejected by the SMTP relay or returned an error |
| `email.opened` | First time a recipient loads the open-tracking pixel |
| `email.clicked` | First time a recipient clicks any tracked link in the email |
| `subscriber.created` | A subscriber is added (UI, API, import, or waitlist signup) |
| `subscriber.unsubscribed` | A subscriber's status changes to `unsubscribed` |

Open and click events fire only on the **first** open/click per email, so a subscriber clicking the same link twice produces a single `email.clicked` event.

## Payload shape

Every delivery is an HTTP `POST` with a JSON body shaped like an envelope:

```json
{
  "id": "evt_a14d2…",
  "type": "email.opened",
  "created_at": "2026-04-29T05:12:33Z",
  "data": { … }
}
```

`data` depends on the event:

```json
// email.sent / email.failed
{
  "log_id": "uuid",
  "project_id": "uuid",
  "to_email": "user@example.com",
  "subject": "Welcome",
  "error": "smtp: 550 mailbox unavailable"   // only on email.failed
}

// email.opened
{
  "log_id": "uuid",
  "project_id": "uuid",
  "opened_at": "2026-04-29T05:12:33Z"
}

// email.clicked
{
  "log_id": "uuid",
  "project_id": "uuid",
  "url": "https://example.com/blog/launch",
  "clicked_at": "2026-04-29T05:12:34Z"
}

// subscriber.created / subscriber.unsubscribed
{
  "subscriber_id": "uuid",
  "project_id": "uuid",
  "email": "user@example.com",
  "name": "Sebastián",
  "status": "active"   // omitted on subscriber.unsubscribed
}
```

`log_id` corresponds to the `email_logs.id` row, so you can join webhook events with whatever the `/logs` endpoint returns.

## Verifying the signature

Each delivery carries a `X-SendDock-Signature` header in this format:

```
X-SendDock-Signature: t=1714368753,v1=9d5a…ef
```

- `t` — Unix timestamp at the moment the request was signed (UTC seconds).
- `v1=` — lowercase hex of `HMAC_SHA256(secret, "<t>.<raw_body>")`.

The signing string is literally `<t>` then a dot then the raw request body — no header reordering, no whitespace trimming, no JSON re-serialisation.

To verify on your side, recompute the HMAC with the same secret and compare in constant time. Reject any request whose `t` is too far in the past (a few minutes is sensible — replay protection).

### Node.js

```js
import crypto from 'node:crypto'

function verify(rawBody, header, secret) {
  const [tPart, vPart] = header.split(',')
  const t = tPart.replace(/^t=/, '')
  const sig = vPart.replace(/^v1=/, '')

  const mac = crypto.createHmac('sha256', secret)
  mac.update(`${t}.${rawBody}`)
  const expected = mac.digest('hex')

  return crypto.timingSafeEqual(Buffer.from(sig), Buffer.from(expected))
}
```

### Go

```go
func verify(rawBody []byte, header, secret string) bool {
    parts := strings.Split(header, ",")
    if len(parts) != 2 {
        return false
    }
    ts := strings.TrimPrefix(parts[0], "t=")
    sig := strings.TrimPrefix(parts[1], "v1=")

    mac := hmac.New(sha256.New, []byte(secret))
    fmt.Fprintf(mac, "%s.", ts)
    mac.Write(rawBody)
    expected := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(sig), []byte(expected))
}
```

### Python

```python
import hmac, hashlib

def verify(raw_body: bytes, header: str, secret: str) -> bool:
    t_part, v_part = header.split(",")
    t = t_part.removeprefix("t=")
    sig = v_part.removeprefix("v1=")

    mac = hmac.new(secret.encode(), f"{t}.".encode() + raw_body, hashlib.sha256)
    return hmac.compare_digest(sig, mac.hexdigest())
```

::: warning Use the raw body, not the parsed JSON
Most frameworks parse the body before your handler runs. The signature is computed over the **bytes that arrived on the wire**, so re-serialising the parsed JSON will produce a different (mismatching) string. Express needs `express.raw({ type: 'application/json' })`, FastAPI needs `Request.body()`, etc.
:::

## Retries

Deliveries that don't return a `2xx` are retried with exponential backoff. The schedule is fixed:

| Attempt | Delay before retry |
|---|---|
| 1 → 2 | 30 seconds |
| 2 → 3 | 2 minutes |
| 3 → 4 | 10 minutes |
| 4 → 5 | 30 minutes |
| 5 → fail | gives up |

After 5 attempts the delivery is marked `failed` and stored on the webhook for visibility. The dispatcher polls every 10 seconds and claims up to 20 ready deliveries per tick using `FOR UPDATE SKIP LOCKED`, so multiple SendDock instances behind the same database don't double-send.

A delivery to a paused webhook (active=false) is marked `failed` immediately on the next tick — pausing a webhook does not buffer events for later replay.

## Managing webhooks

From the dashboard:

1. Open a project, click **Webhooks** in the sidebar.
2. **New webhook**: paste your endpoint URL, pick which event types to subscribe to (all by default), submit.
3. The signing secret appears **once** on creation. Copy it — you cannot retrieve it again later. If you lose it, delete the webhook and create a new one.
4. Each row offers **Pause/Resume**, **Delete**, and **Deliveries** — the latter opens a panel showing the most recent delivery attempts with status, attempt count, last HTTP code, and error message if any.

The same operations are available over the REST API — see [Webhooks API](/api/webhooks).

## Headers SendDock sends

| Header | Value |
|---|---|
| `Content-Type` | `application/json` |
| `User-Agent` | `SendDock-Webhooks/1.0` |
| `X-SendDock-Signature` | `t=<unix>,v1=<hmac-sha256-hex>` |

That's it. No custom auth — the signature is the auth. Make sure your endpoint accepts requests from any IP (SendDock's worker doesn't sit on a fixed range) and validates the signature before trusting the payload.

## Idempotency

The `id` field on the envelope (`evt_<uuid>`) is unique per event. Treat it as an idempotency key: if your handler receives the same `id` twice (a retried delivery, a re-sent webhook from the UI, etc.), it should be a no-op.

## Local development

For testing webhooks against a SendDock running on `localhost`, expose your handler with [ngrok](https://ngrok.com) or [cloudflared](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/) and paste the public URL into the webhook form.

If you want to inspect what SendDock is sending without writing a handler, point the webhook at [webhook.site](https://webhook.site) — every delivery shows up there with the full payload and headers.

## Licensing

Webhook **management** (CRUD endpoints, the UI section) is gated by `SENDDOCK_LICENSE_KEY` in cloud mode. The **dispatcher** runs in Core regardless of license — webhooks created before a license expires keep firing — but new webhooks cannot be created without a valid key.

In **self-hosted** mode an empty key keeps everything unlocked locally, which is the right default for development. See [Configuration → Pro license](/self-hosting/configuration#pro-license).
