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
| `analytics.png` | Pro analytics dashboard — funnel, opens-over-time chart. Requires `SENDDOCK_LICENSE_KEY` set for the dev instance. | `/projects/{id}/analytics` |

## Tips

- Hide the browser chrome (use F11 fullscreen, then screenshot the content area only).
- Sanitize all email addresses — use `demo+ana@example.com`, `demo+luis@example.com`, etc. No real user data.
- For the hero, leave some empty space around the content — a screenshot crammed edge-to-edge feels claustrophobic.
- If the analytics screenshot requires Pro and you don't have a license loaded, you can drop the `analytics.png` row from the README screenshots table for now.
