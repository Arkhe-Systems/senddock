# Reports API <Badge type="warning" text="Pro" />

The engine behind the [Reports](../guide/reports) builder: a catalog of what you can query, a runner that executes one report configuration, and CRUD for saved reports. Pro-gated: without a valid Pro license these endpoints return `402 Payment Required`.

Cookie auth only, scoped to workspace membership: any member of the project's workspace can call these endpoints, and no workspace role is checked.

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
| `measure` | `count` on either dataset. `emails` also takes the counts `sent`, `opened`, `clicked`, `bounced`, `complained` and the rates `open_rate`, `click_rate`, `bounce_rate`, `spam_rate`, `acceptance`. |
| `dimensions` | One or two. Two produces a pivot. Allowed keys depend on the dataset — see the schema response. A `custom.<key>` dimension is validated against `^[a-zA-Z0-9_]+$`. |
| `filter` | Optional [segment predicate](./segments#predicate-shape). |
| `window` | Optional date range + `day` / `week` / `month` granularity for time dimensions. |
| `viz` | `table`, `pivot`, `bar`, `line`, `area`, `donut`, `pie`. Accepted but not read: the response shape depends only on how many dimensions you pass. |

### Response

One dimension returns a flat breakdown; two return a pivot. Both echo the configuration that ran:

```json
{
  "dataset": "emails",
  "measure": "open_rate",
  "dimensions": ["provider"],
  "rows": [ { "label": "Gmail", "value": 45.9 }, { "label": "Outlook", "value": 38.2 } ]
}
```

```json
{
  "dataset": "emails",
  "measure": "open_rate",
  "dimensions": ["provider", "send_time"],
  "pivot": {
    "columns": ["2026-W22", "2026-W23"],
    "rows": [ { "label": "Gmail", "cells": { "2026-W22": 44.1, "2026-W23": 47.0 } } ]
  }
}
```

`rows` is present on a flat report, `pivot` on a two-dimension one. A pivot cell is `0` for a combination with no rows.

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
| `403` | The authenticated user isn't a member of the project's workspace (also returned for a project that doesn't exist). |
| `404` | Report not found, or the project id in the path isn't a valid UUID. |

## See also

- [Reports guide](../guide/reports) — the builder, saved reports and CSV export.
- [Segments API](./segments) — the predicate shape reused by the `filter` field.
