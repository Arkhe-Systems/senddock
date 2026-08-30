# Webhooks API

Manage webhook endpoints and inspect their delivery history. Cookie auth required. Part of the free Core — no license needed.

For payload format, signature verification, retry policy and event reference, see the [Webhooks guide](/guide/webhooks).

## Create a webhook

```
POST /api/v1/projects/{id}/webhooks
```

```json
{
  "url": "https://example.com/webhooks/senddock",
  "events": ["email.sent", "email.opened"]
}
```

| Field | Required | Description |
|---|---|---|
| `url` | yes | Absolute http/https URL of your endpoint. |
| `events` | no | Subset of allowed event types. Omit or pass `[]` to subscribe to **all** events. |

Allowed events: `email.sent`, `email.failed`, `email.bounced`, `email.opened`, `email.clicked`, `subscriber.created`, `subscriber.unsubscribed`, `subscriber.newsletter_unsubscribed`.

**Response — 201 Created**

```json
{
  "id": "uuid",
  "url": "https://example.com/webhooks/senddock",
  "secret": "9f0a3b…64-char-hex",
  "events": ["email.sent", "email.opened"],
  "active": true,
  "created_at": "2026-04-29T05:12:00Z"
}
```

::: warning The secret is only returned here
The signing secret is shown **once**, on creation. SendDock never returns it again — subsequent `GET` calls return the secret field empty (or omit it). If you lose it, delete the webhook and create a new one.
:::

## List webhooks

```
GET /api/v1/projects/{id}/webhooks
```

```json
{
  "webhooks": [
    {
      "id": "uuid",
      "url": "https://example.com/webhooks/senddock",
      "secret": "",
      "events": ["email.sent", "email.opened"],
      "active": true,
      "created_at": "2026-04-29T05:12:00Z"
    }
  ]
}
```

## Get a webhook

```
GET /api/v1/projects/{id}/webhooks/{webhookId}
```

Same shape as a list element. The `secret` field is empty.

## Update a webhook

```
PATCH /api/v1/projects/{id}/webhooks/{webhookId}
```

Only the `active` flag can be patched today — pause or resume delivery without losing the configured events.

```json
{"active": false}
```

Returns the updated webhook.

A paused webhook (`active=false`) does **not** buffer events. Any deliveries that try to fire while it's paused are marked `failed` on the next dispatcher tick.

## Delete a webhook

```
DELETE /api/v1/projects/{id}/webhooks/{webhookId}
```

Returns `204 No Content`. The webhook and all its delivery history are removed immediately.

## List recent deliveries

```
GET /api/v1/projects/{id}/webhooks/{webhookId}/deliveries?limit=50
```

| Param | Default | Max |
|---|---|---|
| `limit` | 50 | 200 |

```json
{
  "deliveries": [
    {
      "id": "uuid",
      "event_type": "email.opened",
      "status": "delivered",
      "attempts": 1,
      "last_status_code": 200,
      "delivered_at": "2026-04-29T05:12:34Z",
      "created_at": "2026-04-29T05:12:33Z"
    },
    {
      "id": "uuid",
      "event_type": "email.failed",
      "status": "pending",
      "attempts": 2,
      "last_status_code": 502,
      "last_error": "Bad Gateway",
      "next_attempt_at": "2026-04-29T05:25:00Z",
      "created_at": "2026-04-29T05:11:00Z"
    }
  ]
}
```

### Status values

<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 760 360" role="img" aria-label="Webhook delivery state machine" style="width:100%;max-width:760px;margin:1rem 0;color:var(--vp-c-text-1);">
  <defs>
    <marker id="ds-a" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse"><path d="M 0 0 L 10 5 L 0 10 z" fill="currentColor" opacity="0.7"/></marker>
  </defs>
  <g style="font-family: ui-sans-serif, system-ui, sans-serif">
    <g transform="translate(40,140)"><rect x="0" y="0" width="130" height="60" rx="30" fill="none" stroke="currentColor" stroke-opacity="0.55"/><text x="65" y="36" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">pending</text></g>
    <line x1="172" y1="170" x2="232" y2="170" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ds-a)"/>
    <text x="202" y="160" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.7">claim</text>
    <g transform="translate(240,140)"><rect x="0" y="0" width="140" height="60" rx="30" fill="none" stroke="currentColor" stroke-opacity="0.9" stroke-width="1.5"/><text x="70" y="36" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">inflight</text></g>
    <line x1="382" y1="170" x2="452" y2="170" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" marker-end="url(#ds-a)"/>
    <text x="417" y="160" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.7">2xx response</text>
    <g transform="translate(460,140)"><rect x="0" y="0" width="160" height="60" rx="30" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="2"/><text x="80" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">delivered</text><text x="80" y="48" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">terminal</text></g>
    <path d="M 310 200 L 310 290 L 452 290" stroke="currentColor" stroke-opacity="0.7" stroke-width="1.5" fill="none" marker-end="url(#ds-a)"/>
    <text x="324" y="226" font-size="11" fill="currentColor" fill-opacity="0.7">5 attempts</text>
    <text x="324" y="244" font-size="11" fill="currentColor" fill-opacity="0.6">or webhook inactive</text>
    <g transform="translate(460,260)"><rect x="0" y="0" width="160" height="60" rx="30" fill="none" stroke="currentColor" stroke-opacity="0.95" stroke-width="2"/><text x="80" y="30" text-anchor="middle" font-size="13" font-weight="600" fill="currentColor">failed</text><text x="80" y="48" text-anchor="middle" font-size="10" font-weight="600" letter-spacing="0.06em" text-transform="uppercase" fill="currentColor" fill-opacity="0.55">terminal</text></g>
    <path d="M 240 200 C 240 252, 170 252, 170 200" stroke="currentColor" stroke-opacity="0.5" stroke-width="1.5" stroke-dasharray="5 4" fill="none" marker-end="url(#ds-a)"/>
    <text x="205" y="276" text-anchor="middle" font-size="11" fill="currentColor" fill-opacity="0.65">non-2xx · attempts &lt; 5</text>
  </g>
</svg>

The same data is also visible from the dashboard — open the **Deliveries** panel on any webhook row.

| Status | Meaning |
|---|---|
| `pending` | Waiting for next attempt. `next_attempt_at` is populated. |
| `inflight` | Claimed by a dispatcher worker, in flight to your endpoint. |
| `delivered` | Your endpoint returned a 2xx. `delivered_at` is populated. Terminal. |
| `failed` | Either the webhook was inactive when it ran, or 5 attempts all failed. Terminal. |

The `attempts` counter increments on every retry; combined with the [retry schedule](/guide/webhooks#retries) it tells you how far through the backoff a `pending` delivery is.

## Errors

| Code | When |
|---|---|
| `400` | Body fails validation — invalid URL, unknown event type, missing required field. |
| `403` | Authenticated user does not own this project, or the role lacks `webhooks:write`. |
| `404` | Project or webhook not found. |
| `500` | Server error. |

All errors share the shape `{"error": "human-readable message"}`.
