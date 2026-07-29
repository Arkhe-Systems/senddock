# Reports API <Badge type="warning" text="Pro" />

The engine behind the [Reports](../guide/reports) builder: a catalog of what you can query, a runner that executes one report configuration, and CRUD for saved reports. Pro-gated: without a valid Pro license these endpoints return `402 Payment Required`.

Cookie auth only, project-owner scoped.

## Schema

```
GET /api/v1/projects/{id}/reports/schema
```

Returns the catalog that populates the builder's dropdowns — the datasets, the dimensions available for each (including the project's own custom fields and tags), the measures, and the visualization types.

## Run a report

```
POST /api/v1/projects/{id}/reports/run
```

Executes a configuration without saving it (this is what the live preview calls).

### Request body

```json
{
  "dataset": "emails",
  "measure": "open_rate",
  "dimensions": ["provider", "send_time"],
  "filter": { "op": "and", "rules": [ ... ] },
  "window": { "from": "2026-06-01", "to": "2026-07-01", "granularity": "week" },
  "viz": "pivot"
}
```

| Field | Notes |
|---|---|
| `dataset` | `subscribers` or `emails`. |
| `measure` | `count`, or for `emails` a rate: `open_rate`, `click_rate`, `bounce_rate`, `spam_rate`. |
| `dimensions` | One or two. Two produces a pivot. Allowed keys depend on the dataset — see the schema response. A `custom.<key>` dimension is validated against `^[a-zA-Z0-9_]+$`. |
| `filter` | Optional [segment predicate](./segments#predicate-shape). |
| `window` | Optional date range + `day` / `week` / `month` granularity for time dimensions. |
| `viz` | `table`, `pivot`, `bar`, `line`, `area`, `donut`, `pie`. |

### Response

One dimension returns a flat breakdown; two return a pivot:

```json
{
  "dimension": "provider",
  "rows": [ { "label": "Gmail", "value": 45.9 }, { "label": "Outlook", "value": 38.2 } ]
}
```

```json
{
  "row_dim": "provider",
  "col_dim": "send_time",
  "cols": ["2026-W22", "2026-W23"],
  "rows": [ { "label": "Gmail", "cells": { "2026-W22": 44.1, "2026-W23": 47.0 } } ]
}
```

## Saved reports

```
GET    /api/v1/projects/{id}/reports              # list
POST   /api/v1/projects/{id}/reports              # create { name, config }
PATCH  /api/v1/projects/{id}/reports/{reportId}   # rename / update config
DELETE /api/v1/projects/{id}/reports/{reportId}   # delete
```

A saved report is a `name` plus the same configuration object accepted by `/run`. Reports are keyed by `(id, project_id)`, so one project can't read or mutate another's.

### Errors

| Status | Cause |
|---|---|
| `400` | Invalid configuration (unknown dimension/measure, wrong dimension count, bad custom-field key). |
| `401` | Missing cookie session. |
| `402` | No valid Pro license. |
| `403` | Authenticated user doesn't own this project. |
| `404` | Project or report not found. |

## See also

- [Reports guide](../guide/reports) — the builder, saved reports and CSV export.
- [Segments API](./segments) — the predicate shape reused by the `filter` field.
