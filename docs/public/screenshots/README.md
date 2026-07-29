# Screenshots

These images are referenced from the project README and the docs site.

GitHub renders this folder via relative paths (`docs/public/screenshots/foo.png`), and VitePress serves it under `/screenshots/foo.png` on docs.senddock.dev.

## Required captures

Use a clean demo dataset (10–20 fake subscribers, 2–3 templates, 1 finished campaign). 1440×900 viewport in a Chromium-based browser, dark mode. Trim to the relevant UI area and export as PNG (target < 500 KB each, use [squoosh.app](https://squoosh.app) if needed).

| Filename | What to capture | Where |
|---|---|---|
| `hero.png` | Project dashboard with subscribers count, templates count, recent activity. The single most-important visual — this is what people see at the top of the README. | `/projects/{id}` Overview tab |
| `projects.png` | Same overview, but slightly different angle (or a cleaner zoom). | `/projects/{id}` Overview tab |
| `editor.png` | Template editor in Visual tab with a decent-looking email loaded (header, body, CTA, footer). | `/projects/{id}` Templates → open a template |
| `campaigns.png` | Campaigns list with at least one completed broadcast and one scheduled. | `/projects/{id}` Campaigns tab |
| `analytics.png` | Analytics dashboard (now **free** — no license needed). Overview tab with the KPI tiles and the opens/clicks chart. | `/projects/{id}/analytics` |

## v0.8 captures (new)

These are referenced by the refreshed docs and don't exist yet. Same house style as above (1440×900, dark, trimmed, < 500 KB).

| Filename | What to capture | Where | Notes |
|---|---|---|---|
| `analytics-overview.png` | The Analytics **Overview** tab: KPI tiles (incl. Acceptance and Spam rate), trend pills, the opens/clicks time series, send-status donut. | `/projects/{id}/analytics` | Free. Use an instance with real sends so the tiles aren't zero. |
| `send-rich-text.png` | The **Send Email** modal with a variable toggled to **Rich** — the WYSIWYG toolbar visible and a bit of formatted content (a bold word + a bullet list) in the editor. | `/projects/{id}` Overview → **Send Email** | Free. Needs a template with a custom `{{placeholder}}` and SMTP configured (so the button shows). |
| `instance-settings.png` | The **Instance** settings screen: Public URL field, Session timeout, and the License section. | dashboard → **Instance** | **Self-host mode only** — this screen is hidden when `CLOUD=true`. Capture from a self-hosted (non-cloud) dev instance. |
| `deliverability.png` | The **Deliverability** tab: domain-health pass/warn/fail rows + the per-provider table with rates. | `/projects/{id}/analytics` → Deliverability | **Pro** — needs a license/plan active so it isn't the paywall. |
| `reports.png` | The **Reports** builder mid-report: dataset/measure/dimension controls on the left, a chart or pivot preview on the right. | `/projects/{id}/reports` | **Pro** — needs a license/plan active. |

## Tips

- Hide the browser chrome (use F11 fullscreen, then screenshot the content area only).
- Sanitize all email addresses — use `demo+ana@example.com`, `demo+luis@example.com`, etc. No real user data.
- For the hero, leave some empty space around the content — a screenshot crammed edge-to-edge feels claustrophobic.
- If the analytics screenshot requires Pro and you don't have a license loaded, you can drop the `analytics.png` row from the README screenshots table for now.
