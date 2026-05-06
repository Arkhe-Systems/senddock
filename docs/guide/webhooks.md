# Webhooks <Badge type="warning" text="Pro" />

Webhooks let SendDock call your own HTTP endpoint every time something interesting happens in a project — an email is sent, a recipient opens it, a subscriber unsubscribes, and so on. They are the right way to keep a CRM, a usage table, or a Slack channel in sync without polling the API.

Webhook delivery, signing and retries ship in the open-source Core; the **management UI and API endpoints** (creating, listing, pausing, deleting webhooks) live in Pro and require a license in cloud mode.

## How a delivery works

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 360" role="img" aria-label="Webhook delivery pipeline" style="width:100%;max-width:800px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="wf-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, -apple-system, sans-serif">
    <g transform="translate(20,150)"><rect x="0" y="0" width="150" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="75" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Event fires</text><text x="75" y="48" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.6">email · subscriber</text></g>
    <line x1="172" y1="182" x2="208" y2="182" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(210,150)"><rect x="0" y="0" width="150" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="75" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Enqueue</text><text x="75" y="48" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.6">webhook_deliveries</text></g>
    <line x1="362" y1="182" x2="398" y2="182" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(400,150)"><rect x="0" y="0" width="150" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.9" stroke-width="1.5"/><text x="75" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">Dispatcher</text><text x="75" y="48" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.6">tick every 10s</text></g>
    <line x1="552" y1="182" x2="588" y2="182" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.6" marker-end="url(#wf-a)"/>
    <g transform="translate(590,150)"><rect x="0" y="0" width="160" height="64" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45"/><text x="80" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">POST + HMAC</text><text x="80" y="48" text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.6">your endpoint</text></g>
    <path d="M 670 150 L 670 90" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.6" fill="none" marker-end="url(#wf-a)"/>
    <text x="684" y="125" font-size="11" fill="currentColor" fill-opacity="0.75">2xx</text>
    <g transform="translate(605,40)"><rect x="0" y="0" width="130" height="46" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.9" stroke-width="1.6"/><text x="65" y="22" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">delivered</text><text x="65" y="38" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">terminal</text></g>
    <path d="M 670 214 L 670 270" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.6" stroke-dasharray="5 4" fill="none" marker-end="url(#wf-a)"/>
    <text x="684" y="245" font-size="11" fill="currentColor" fill-opacity="0.65">non-2xx</text>
    <g transform="translate(605,272)"><rect x="0" y="0" width="130" height="46" rx="10" fill="none" stroke="currentColor" stroke-opacity="0.45" stroke-dasharray="5 4"/><text x="65" y="22" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">retry</text><text x="65" y="38" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">backoff 30s → 2h</text></g>
    <path d="M 605 295 L 285 295 L 285 215" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.6" stroke-dasharray="5 4" fill="none" marker-end="url(#wf-a)"/>
    <text x="445" y="287" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">loop back if attempts &lt; 5 · otherwise marked failed</text>
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

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 220" role="img" aria-label="Retry backoff timeline" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <line x1="40" y1="110" x2="720" y2="110" stroke="currentColor" stroke-opacity="0.35" stroke-width="1.5"/>
    <g><line x1="60" y1="102" x2="60" y2="118" stroke="currentColor" stroke-opacity="0.8" stroke-width="1.5"/><circle cx="60" cy="110" r="5" fill="currentColor" fill-opacity="0.8"/><text x="60" y="86" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#1</text><text x="60" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">0s</text></g>
    <g><line x1="160" y1="102" x2="160" y2="118" stroke="currentColor" stroke-opacity="0.8" stroke-width="1.5"/><circle cx="160" cy="110" r="5" fill="currentColor" fill-opacity="0.8"/><text x="160" y="86" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#2</text><text x="160" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">30s</text></g>
    <text x="110" y="172" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.04em" fill="currentColor" fill-opacity="0.55">+30s</text>
    <g><line x1="290" y1="102" x2="290" y2="118" stroke="currentColor" stroke-opacity="0.8" stroke-width="1.5"/><circle cx="290" cy="110" r="5" fill="currentColor" fill-opacity="0.8"/><text x="290" y="86" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#3</text><text x="290" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">2m 30s</text></g>
    <text x="225" y="172" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.04em" fill="currentColor" fill-opacity="0.55">+2m</text>
    <g><line x1="450" y1="102" x2="450" y2="118" stroke="currentColor" stroke-opacity="0.8" stroke-width="1.5"/><circle cx="450" cy="110" r="5" fill="currentColor" fill-opacity="0.8"/><text x="450" y="86" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#4</text><text x="450" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">12m 30s</text></g>
    <text x="370" y="172" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.04em" fill="currentColor" fill-opacity="0.55">+10m</text>
    <g><line x1="640" y1="102" x2="640" y2="118" stroke="currentColor" stroke-opacity="0.8" stroke-width="1.5"/><circle cx="640" cy="110" r="5" fill="currentColor" fill-opacity="0.8"/><text x="640" y="86" text-anchor="middle" font-size="12" font-weight="600" fill="currentColor">#5</text><text x="640" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.6">42m 30s</text></g>
    <text x="545" y="172" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.04em" fill="currentColor" fill-opacity="0.55">+30m</text>
    <g><line x1="710" y1="98" x2="710" y2="122" stroke="currentColor" stroke-width="2"/><rect x="694" y="98" width="32" height="24" fill="none" stroke="currentColor" stroke-width="1.6"/><text x="710" y="86" text-anchor="middle" font-size="11" font-weight="700" letter-spacing="0.06em" fill="currentColor">FAIL</text><text x="710" y="144" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.7">~2h 12m</text></g>
    <text x="675" y="172" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.04em" fill="currentColor" fill-opacity="0.55">+2h</text>
    <text x="380" y="206" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.55">5 attempts max — terminal &quot;failed&quot; if all return non-2xx</text>
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

An empty `SENDDOCK_LICENSE_KEY` keeps the management UI / API locked regardless of deployment mode — but the Core dispatcher keeps running, so any webhooks created earlier (from a Pro-licensed snapshot) continue firing. See [Configuration → Pro license](/self-hosting/configuration#plans-and-licensing).
