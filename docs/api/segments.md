# Segments API

Segments are saved filters over your subscribers, evaluated fresh on every use (no materialized membership). A broadcast can target a segment instead of "all active subscribers". Cookie auth required; mutations need the `subscribers:write` role capability. Part of the free Core.

See the [Segments guide](/guide/segments) for concepts and examples.

## The predicate

A segment stores a `predicate` — a match mode plus a list of rules:

```json
{
  "match": "all",
  "rules": [
    {"field": "status", "op": "eq", "value": "active"},
    {"field": "tags", "op": "includes_any", "value": ["vip", "customer"]},
    {"field": "custom.plan_tier", "op": "eq", "value": "pro"}
  ]
}
```

`match` is `all` (AND) or `any` (OR). Each rule is `{field, op, value}`.

| Field | Operators | Value |
|---|---|---|
| `status` | `eq`, `neq` | `active` / `pending` / `unsubscribed` |
| `tags` | `includes_any`, `includes_all`, `excludes` | array of tags |
| `custom.KEY` | `eq`, `neq`, `contains`, `gt`, `lt` | value matching the field type (`gt`/`lt` are numeric) |

`custom.KEY` rules read [custom field](/api/subscribers#custom-fields) values. An unknown field name or operator returns `400`.

## List segments

```
GET /api/v1/projects/{id}/segments
```

```json
[
  {
    "id": "uuid",
    "project_id": "uuid",
    "name": "Active pro customers",
    "predicate": { "match": "all", "rules": [ ... ] },
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
]
```

## Create a segment

```
POST /api/v1/projects/{id}/segments
```

```json
{"name": "Active pro customers", "predicate": { "match": "all", "rules": [ ... ] }}
```

Returns `201` with the created segment, or `400` for an invalid predicate.

## Update a segment

```
PATCH /api/v1/projects/{id}/segments/{segmentId}
```

```json
{"name": "Active pro & team", "predicate": { "match": "any", "rules": [ ... ] }}
```

## Delete a segment

```
DELETE /api/v1/projects/{id}/segments/{segmentId}
```

Returns `204`. Broadcasts already sent are unaffected.

## Preview a segment

```
POST /api/v1/projects/{id}/segments/preview
```

Counts how many **active** subscribers a predicate matches, without saving it — the dashboard uses this for the live count while you build rules.

```json
{"predicate": { "match": "all", "rules": [ ... ] }}
```

**Response**

```json
{"count": 128}
```

## Sending to a segment

Pass `segment_id` to the [broadcast endpoint](/api/sending#broadcast). Recipients are the active subscribers matching the segment; omit `segment_id` (or send `""`) to fall back to all active subscribers.

```json
{"template_id": "uuid", "segment_id": "uuid"}
```
