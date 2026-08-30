# TypeScript SDK

The official SDK wraps the project-scoped REST API in a typed client — no hand-rolled `fetch`, no copy-pasted auth headers, no guessing response shapes.

```bash
npm install @senddock/sdk
```

Source lives at [Arkhe-Systems/senddock-js](https://github.com/Arkhe-Systems/senddock-js). Zero runtime dependencies, ESM + CJS, Node 18+.

::: tip Server-side only
API keys carry full project access — keep them in your backend (API routes, workers, crons), never in browser code.
:::

## Setup

```ts
import { SendDock } from '@senddock/sdk'

const senddock = new SendDock({
  baseUrl: 'https://senddock.example.com',
  projectId: 'your-project-uuid',
  apiKey: process.env.SENDDOCK_API_KEY,
})
```

Create the key under **Project → Settings → API Keys**. Keys are project-scoped, so the client always talks to the project the key belongs to. Optional knobs: `maxRetries` (default `2`, set `0` to disable) and `timeoutMs` (default `30000`).

## Why the SDK

The same call, both ways. Without the SDK, a production-grade send needs URL building, auth, error extraction and retry handling:

```ts
const res = await fetch(`${base}/api/v1/projects/${projectId}/send`, {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${key}`, 'Content-Type': 'application/json' },
  body: JSON.stringify({ to, template_id: templateId, data }),
})
if (!res.ok) {
  let message = `senddock request failed: ${res.status}`
  try { message = (await res.json()).error ?? message } catch {}
  throw new Error(message)
}
const result = await res.json()
```

…and that still has no retries, no `Retry-After` handling, no timeout, and no types. With the SDK:

```ts
const result = await senddock.send({ to, template_id: templateId, data })
```

What you get on every call:

- **Full typing.** Requests are validated by the compiler — an invalid body doesn't build. Responses are typed, including the discriminated union for suppressed recipients.
- **One error type.** Anything that goes wrong throws `SendDockError` carrying the API's own message — a failed fetch never leaks a raw `TypeError` or `SyntaxError`.
- **Retries built in.** `429` waits for `Retry-After` (capped at 60s); `5xx` and network failures back off exponentially; `4xx` fails fast.
- **Timeouts.** Every request carries an abort signal, so a hung connection can't block your process.

## Sending

`send()` accepts the three shapes the [send endpoint](./sending) supports:

```ts
await senddock.send({ to: 'ada@example.com', template_id: 'tpl-id', data: { name: 'Ada' } })

await senddock.send({ subscriber_id: 'sub-id', template_id: 'tpl-id' })

await senddock.send({ to: 'ada@example.com', subject: 'Hi', html_body: '<h1>Hello</h1>' })
```

A suppressed recipient does not throw — the API answers `200` with `{ message: 'suppressed', suppressed: 1 }` and the SDK passes that through.

### Batch

Up to 500 recipients per call, each with their own variables:

```ts
const result = await senddock.sendBatch({
  template_id: 'tpl-id',
  recipients: [
    { to: 'ada@example.com', data: { name: 'Ada' } },
    { to: 'alan@example.com', data: { name: 'Alan' } },
  ],
})
// { sent: 2, failed: 0, suppressed: 0 }
```

### Broadcast

All active subscribers, or a saved [segment](./segments):

```ts
await senddock.broadcast({ template_id: 'tpl-id' })
await senddock.broadcast({ template_id: 'tpl-id', segment_id: 'seg-id' })
```

Broadcasts require the instance to have a [public URL](/guide/instance-settings#public-url) configured — recipients need a working unsubscribe link.

## Importing subscribers

The server-side way to add subscribers ([the single-create endpoint is cookie-only](./subscribers)). One row or tens of thousands, with validation and an exhaustive report:

```ts
const report = await senddock.importSubscribers(
  [{ email: 'ada@example.com', name: 'Ada', tags: ['beta'], fields: { plan: 'pro' } }],
  { validate_mx: true, validate_disposable: true },
)
// { imported, duplicates, syntax_invalid, no_mx, disposable, suppressed, rejected: [...] }
```

## Stats

```ts
const stats = await senddock.stats()
// { total, sent, failed, bounced, suppressed, opened }
```

## Verifying webhooks

SendDock signs outbound [webhooks](./webhooks) with `X-SendDock-Signature: t=<unix>,v1=<hmac>`. Verify against the **raw** body, before any JSON parsing:

```ts
import { verifyWebhookSignature } from '@senddock/sdk'

app.post('/webhooks/senddock', express.raw({ type: 'application/json' }), (req, res) => {
  const valid = verifyWebhookSignature({
    payload: req.body,
    signature: req.get('X-SendDock-Signature') ?? '',
    secret: process.env.SENDDOCK_WEBHOOK_SECRET,
  })
  if (!valid) return res.status(401).end()
  const event = JSON.parse(req.body)
  res.status(200).end()
})
```

The check is timing-safe and rejects timestamps older than 5 minutes (configurable via `toleranceSeconds`). The secret is shown once, when the webhook is created.

## Error handling

Every non-2xx response and every transport failure throws the same typed error:

```ts
import { SendDockError } from '@senddock/sdk'

try {
  await senddock.send({ to: 'ada@example.com', template_id: 'tpl-id' })
} catch (err) {
  if (err instanceof SendDockError) {
    err.status         // 401, 404, 429... or 0 for network failures
    err.message        // the API's error string, e.g. "missing or invalid api key"
    err.isAuthError    // convenience flags: isAuthError, isNotFound, isRateLimit, isNetworkError
  }
}
```

## What the SDK covers

API keys intentionally cover the runtime surface: sending, importing subscribers and stats. Configuration — templates, segments, webhooks, campaigns — is managed from the dashboard, where role-based permissions apply. See [Authentication → Endpoints that accept API keys](./authentication#endpoints-that-accept-api-keys).

Found a gap or a bug? Open an issue on [senddock-js](https://github.com/Arkhe-Systems/senddock-js/issues).
