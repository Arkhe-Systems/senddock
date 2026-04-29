# Webhooks <Badge type="warning" text="Pro" />

Webhooks let SendDock call your own HTTP endpoint every time something interesting happens in a project — an email is sent, a recipient opens it, a subscriber unsubscribes, and so on. They are the right way to keep a CRM, a usage table, or a Slack channel in sync without polling the API.

Webhook delivery, signing and retries ship in the open-source Core; the **management UI and API endpoints** (creating, listing, pausing, deleting webhooks) live in Pro and require a license in cloud mode.

## How a delivery works

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 280" role="img" aria-label="Webhook delivery pipeline" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="wf-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.6"/></marker>
    <marker id="wf-ag" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#10b981"/></marker>
    <marker id="wf-aa" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="#f59e0b"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, -apple-system, sans-serif">
    <g transform="translate(20,108)"><rect x="0" y="0" width="130" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="65" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Event fires</text><text x="65" y="48" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">email.sent / opened / …</text></g>
    <line x1="152" y1="140" x2="188" y2="140" stroke="currentColor" stroke-opacity="0.6" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(190,108)"><rect x="0" y="0" width="130" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="65" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Enqueue</text><text x="65" y="48" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">webhook_deliveries</text></g>
    <line x1="322" y1="140" x2="358" y2="140" stroke="currentColor" stroke-opacity="0.6" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(360,108)"><rect x="0" y="0" width="130" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.85"/><text x="65" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Dispatcher</text><text x="65" y="48" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">tick every 10s</text></g>
    <line x1="492" y1="140" x2="528" y2="140" stroke="currentColor" stroke-opacity="0.6" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(530,108)"><rect x="0" y="0" width="150" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="75" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">POST + HMAC</text><text x="75" y="48" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">your endpoint</text></g>
    <path d="M 605 108 L 605 70" stroke="#10b981" stroke-width="1.6" fill="none" marker-end="url(#wf-ag)"/>
    <g transform="translate(560,28)"><rect x="0" y="0" width="90" height="32" rx="8" fill="none" stroke="#10b981"/><text x="45" y="20" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="#10b981">2xx · delivered</text></g>
    <path d="M 605 172 L 605 218" stroke="#f59e0b" stroke-width="1.6" stroke-dasharray="5 4" fill="none" marker-end="url(#wf-aa)"/>
    <g transform="translate(550,220)"><rect x="0" y="0" width="110" height="32" rx="8" fill="none" stroke="#f59e0b"/><text x="55" y="20" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="#f59e0b">non-2xx · retry</text></g>
    <path d="M 550 236 L 255 236 L 255 174" stroke="#f59e0b" stroke-width="1.6" stroke-dasharray="5 4" fill="none" marker-end="url(#wf-aa)"/>
    <text x="402" y="228" text-anchor="middle" font-size="11" fill="#f59e0b">attempts &lt; 5 → backoff</text>
  </g>
</svg>

The dispatcher runs inside the Core process, polls every 10 seconds, and claims up to 20 ready deliveries per tick using `FOR UPDATE SKIP LOCKED` — so two SendDock containers behind the same Postgres won't double-send.

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

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 200" role="img" aria-label="Retry backoff timeline" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <line x1="40" y1="100" x2="720" y2="100" stroke="currentColor" stroke-opacity="0.4" stroke-width="1.5"/>
    <g><line x1="60" y1="92" x2="60" y2="108" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5"/><circle cx="60" cy="100" r="5" fill="currentColor" fill-opacity="0.7"/><text x="60" y="80" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#1</text><text x="60" y="130" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">0s</text></g>
    <g><line x1="160" y1="92" x2="160" y2="108" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5"/><circle cx="160" cy="100" r="5" fill="currentColor" fill-opacity="0.7"/><text x="160" y="80" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#2</text><text x="160" y="130" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">30s</text></g>
    <text x="110" y="158" text-anchor="middle" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.75">+30s</text>
    <g><line x1="290" y1="92" x2="290" y2="108" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5"/><circle cx="290" cy="100" r="5" fill="currentColor" fill-opacity="0.7"/><text x="290" y="80" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#3</text><text x="290" y="130" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">2m 30s</text></g>
    <text x="225" y="158" text-anchor="middle" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.75">+2m</text>
    <g><line x1="450" y1="92" x2="450" y2="108" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5"/><circle cx="450" cy="100" r="5" fill="currentColor" fill-opacity="0.7"/><text x="450" y="80" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#4</text><text x="450" y="130" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">12m 30s</text></g>
    <text x="370" y="158" text-anchor="middle" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.75">+10m</text>
    <g><line x1="640" y1="92" x2="640" y2="108" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5"/><circle cx="640" cy="100" r="5" fill="currentColor" fill-opacity="0.7"/><text x="640" y="80" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#5</text><text x="640" y="130" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">42m 30s</text></g>
    <text x="545" y="158" text-anchor="middle" font-size="10" font-weight="600" fill="currentColor" fill-opacity="0.75">+30m</text>
    <g><line x1="710" y1="92" x2="710" y2="108" stroke="#ef4444" stroke-width="1.5"/><circle cx="710" cy="100" r="5" fill="#ef4444"/><text x="710" y="80" text-anchor="middle" font-size="11" font-weight="600" letter-spacing="0.05em" fill="#ef4444">FAIL</text><text x="710" y="130" text-anchor="middle" font-size="11" fill="#ef4444">~2h 12m</text></g>
    <text x="675" y="158" text-anchor="middle" font-size="10" font-weight="600" fill="#f59e0b">+2h</text>
    <text x="380" y="190" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">5 attempts max — terminal &quot;failed&quot; if all return non-2xx</text>
  </g>
</svg>

| Attempt | Delay before retry |
|---|---|
| 1 → 2 | 30 seconds |
| 2 → 3 | 2 minutes |
| 3 → 4 | 10 minutes |
| 4 → 5 | 30 minutes |
| 5 → fail | gives up |

After 5 attempts the delivery is marked `failed` and stored on the webhook for visibility.

A delivery to a paused webhook (active=false) is marked `failed` immediately on the next tick — pausing a webhook does not buffer events for later replay.

## Managing webhooks

From the dashboard:

1. Open a project, click **Webhooks** in the sidebar.
2. **New webhook**: paste your endpoint URL, pick which event types to subscribe to (all by default), submit.
3. The signing secret appears **once** on creation. Copy it — you cannot retrieve it again later. If you lose it, delete the webhook and create a new one.
4. Each row offers **Pause/Resume**, **Delete**, and **Deliveries** — the latter opens a panel showing the most recent delivery attempts with status, attempt count, last HTTP code, and error message if any.

<img src="/screenshots/webhooks-list.png" alt="Webhooks list view with one active webhook subscribed to all six event types" style="width:100%;max-width:900px;margin:1rem 0;border-radius:12px;border:1px solid rgba(120,120,128,0.25);" />

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
