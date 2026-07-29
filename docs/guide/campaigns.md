# Campaigns

Campaigns let you schedule email broadcasts for a future time. Instead of sending immediately, you create a campaign that pairs a template with a scheduled delivery time. When the time arrives, SendDock broadcasts the template to all active subscribers in the project.

![The Newsletters view listing scheduled campaigns](/screenshots/campaigns.png)

## How Campaigns Work

1. You create a campaign with a template, a name, and a scheduled time
2. The campaign sits in `scheduled` status until the scheduled time
3. A background worker checks every 30 seconds for campaigns ready to send
4. When the time arrives, the worker broadcasts the template to all active subscribers
5. The campaign status moves through `sending` and finally to `sent` (or `failed` if something goes wrong)

## Campaign Statuses

| Status | Description |
|--------|-------------|
| `scheduled` | Waiting for the scheduled time |
| `sending` | Currently broadcasting to subscribers |
| `sent` | All emails have been sent |
| `failed` | An error occurred during sending |

## Create a Campaign

From the **Newsletters** tab in your project, click **+ New Campaign**, then:

![The New Campaign dialog](/screenshots/new-campaign-modal.png)


1. **Name** the campaign (for your own reference in the list).
2. Pick a **template**.
3. Optionally **override the subject**.
4. Fill in any **template variables** — each custom `{{placeholder}}` in the template gets its own input, applied to every recipient (subscriber values like `{{name}}` are filled per person automatically).
5. Choose **Send now** to broadcast immediately, or **Schedule** and set a date and time.

Click **Create** and it appears in the list. The programmatic equivalent, including the `variables` shape, is in the [Campaigns API reference](/api/campaigns#create-campaign).

::: tip Cookie auth only
Campaigns mutate workspace state and require role-based capabilities (`campaigns:write`). API keys, which are project-scoped and identity-less, can't call these endpoints — you'll get `401`. Schedule from the dashboard, or call the endpoints from your own UI built on the same cookie-session login.
:::

## List Campaigns

The **Newsletters** tab shows every campaign in the project, ordered by scheduled time, with their current status and live `sent_count` / `failed_count`. The same data is exposed at `GET /api/v1/projects/{id}/campaigns` — see the [API reference](/api/campaigns#list-campaigns).

## Edit or delete a campaign

Each row in the **Newsletters** tab has:

- **Edit** — only available while the campaign is still `scheduled`. Use it to change the template, subject, variables or scheduled time before it fires. Once it moves to `sending`, `sent` or `failed`, editing is closed.
- **Delete** — available in **any** status. Deleting a `scheduled` campaign cancels it before it sends; deleting a `sent`/`failed` one just removes it from the list (it doesn't unsend anything). A confirmation dialog spells out what will happen for the current status.

## How the Worker Operates

The campaign worker runs as part of the SendDock backend process. It polls every 30 seconds for campaigns whose `scheduled_at` time has passed and whose status is still `scheduled`. When it finds one, it:

1. Atomically claims the campaign (sets status to `sending`, prevents duplicate execution if multiple instances share a database)
2. Creates a broadcast and enqueues one job per active subscriber into the `broadcast_jobs` table
3. Links the campaign to that broadcast via `broadcast_id` and exits — actual sending is then driven by the broadcast worker pool
4. When the broadcast worker drains the queue, it cascades the final `sent_count` / `failed_count` back to the campaign and flips its status to `sent`

While the campaign is in `sending`, its `sent_count` and `failed_count` shown in the dashboard reflect the **live progress** of the underlying broadcast (read directly from the linked broadcast row each request). You will see the numbers climb in real time as the queue drains, not jump from 0 to total at the end.

If the backend process is killed mid-broadcast, the campaign stays in `sending`, jobs that were `sending` get reset to `retry` on the next startup, and sending resumes from where it left off — no recipient is sent to twice and no recipient is silently dropped. See the **How a broadcast actually runs** section in [Sending emails](/guide/sending#how-a-broadcast-actually-runs) for queue mechanics and retry behavior.

No additional configuration is needed -- the worker starts automatically with the backend.

## Per-campaign analytics

Every email a campaign sends is tagged with its broadcast, so each campaign gets its own engagement breakdown — not just a global roll-up. Open **Project → Analytics → Campaigns** to see, per broadcast, how many were sent, opened and clicked and the resulting rates. See [Analytics](/guide/analytics#tabs).

## Tips

- Use descriptive campaign names so you can identify them later in the list
- Schedule campaigns at least a few minutes in the future to allow time for review
- The template's subject line is used for the campaign emails
- Subscriber variables (`{{name}}`, `{{email}}`, `{{unsubscribe_url}}`) are replaced per recipient, just like broadcast
