# UI Screenshots

The docs reference screenshots that live in this directory. Until you drop the images here, the `<img>` tags in the markdown will render broken-image icons.

## Required screenshots

Take these from a project on your dev or local instance. Use the dark theme (default). 1× zoom, browser at ~1280px wide is fine. PNG, no compression artifacts.

| File | Where in the docs | What to capture |
|---|---|---|
| `webhooks-list.png` | `guide/webhooks.md` | The Webhooks section of a project with at least one active webhook. Crop to the section, hide your URL secret if it's visible anywhere. |
| `webhook-deliveries.png` | `api/webhooks.md` | The "Recent deliveries" modal with at least 3 entries (mix of `delivered` and `pending`/`failed` if possible). |
| `analytics-dashboard.png` | `guide/analytics.md` | The full Analytics page with the 6 metric cards, funnel, opens chart, top templates and top clicked links visible. Use the 30d preset. |

## Tips

- Open dev tools → toggle device toolbar → set width to 1280–1440 for a clean crop.
- Hide your real domain in the URL bar (or screenshot just the content area, not the chrome).
- Don't include real subscriber emails — use a test project with seed data.
- Optional: run images through `pngquant --quality=80-95` or similar to keep the docs bundle light.

Drop the PNGs in this folder with the exact filename listed above.
