# Webhooks API <Badge type="warning" text="Pro" />

Manage webhook endpoints and inspect their delivery history. Cookie auth required.

In **cloud** mode every endpoint on this page returns `402 Payment Required` (`{"error":"license required for webhooks"}`) when the deployment has no valid `SENDDOCK_LICENSE_KEY`. In **self-hosted** mode an empty license unlocks everything locally — see [Configuration → Pro license](/self-hosting/configuration#pro-license).

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

Allowed events: `email.sent`, `email.failed`, `email.opened`, `email.clicked`, `subscriber.created`, `subscriber.unsubscribed`.

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

<img src="/diagrams/delivery-states.svg" alt="Webhook delivery state machine" style="width:100%;max-width:760px;margin:1rem 0;" />

The same data is also visible from the dashboard — open the **Deliveries** panel on any webhook row:

<img src="/screenshots/webhook-deliveries.png" alt="Recent deliveries panel showing email.sent, email.opened and email.clicked events all delivered with HTTP 200" style="width:100%;max-width:600px;margin:1rem 0;border-radius:12px;border:1px solid rgba(120,120,128,0.25);" />

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
| `402` | License required (cloud mode without a valid `SENDDOCK_LICENSE_KEY`). |
| `403` | Authenticated user does not own this project. |
| `404` | Project or webhook not found. |
| `500` | Server error. |

All errors share the shape `{"error": "human-readable message"}`.
