# Monitoring

SendDock emits **structured JSON logs** and a **Prometheus metrics endpoint** so you can watch the send pipeline in production — queue depth, send outcomes, SMTP failures by class, and bounce/complaint ingest.

## Structured logs

Logs are emitted as JSON on stdout, one object per line, with a consistent `time`, `level` and `msg`. Pipeline events on the hot paths add fields like `project_id` and message counts. Any log aggregator that parses JSON (Loki, Elastic, Datadog, `docker compose logs` piped to `jq`) picks them up without a parser rule.

```json
{"time":"2026-08-19T10:00:00Z","level":"INFO","msg":"bounce imap poller processed project","project_id":"…","messages_with_suppressions":2,"messages_total":5}
```

## Metrics endpoint

`GET /metrics` serves the Prometheus exposition format on the same port as the app (default `8080`), alongside `/health`.

```bash
curl -s http://localhost:8080/metrics | grep senddock_
```

::: warning Unauthenticated by design
`/metrics` is not behind auth — it's the Prometheus convention, and the endpoint exposes only operational counters, no message content. Scope it at the network layer: keep port `8080` private and scrape from inside your network, or expose only `/metrics` through your reverse proxy with an allow-list. Don't publish it to the open internet.
:::

A minimal Prometheus scrape config:

```yaml
scrape_configs:
  - job_name: senddock
    metrics_path: /metrics
    static_configs:
      - targets: ['senddock:8080']
```

### Metrics exposed

| Metric | Type | Labels | Meaning |
|---|---|---|---|
| `senddock_broadcast_queue_depth` | gauge | — | Broadcast jobs waiting or in flight (`pending`, `retry`, `sending`). |
| `senddock_email_send_attempts_total` | counter | — | Emails handed to the SMTP send path. |
| `senddock_emails_sent_total` | counter | — | Emails accepted by the upstream SMTP server. |
| `senddock_emails_failed_total` | counter | `reason` (`bounce`, `error`) | Emails that failed to send. |
| `senddock_smtp_errors_total` | counter | `class` (`4xx`, `5xx`, `other`) | SMTP errors by response class. |
| `senddock_webhook_deliveries_total` | counter | `result` (`success`, `retry`, `failed`) | Outbound webhook delivery outcomes. |
| `senddock_bounce_ingest_total` | counter | `source` (`webhook`, `imap`) | Bounces ingested. |
| `senddock_complaint_ingest_total` | counter | — | Spam complaints ingested. |
| `senddock_bounce_poller_tick_seconds` | histogram | — | Duration of a bounce IMAP poller tick. |

The endpoint also carries the standard Go runtime and process collectors (`go_*`, `process_*`) from the Prometheus client.

### Useful signals

- **Deliverability degrading:** a rising `rate(senddock_smtp_errors_total{class="5xx"}[5m])` or `senddock_emails_failed_total{reason="bounce"}` — recipients or your sending reputation are rejecting mail.
- **Queue backing up:** `senddock_broadcast_queue_depth` climbing and not draining means the broadcast workers can't keep pace (SMTP throttling, or a stuck provider).
- **Webhooks failing:** `senddock_webhook_deliveries_total{result="failed"}` rising means downstream endpoints are down or misconfigured.
