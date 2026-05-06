# Campaigns

Campaigns let you schedule email broadcasts for a future time. Instead of sending immediately, you create a campaign that pairs a template with a scheduled delivery time. When the time arrives, SendDock broadcasts the template to all active subscribers in the project.

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

From the **Newsletters** tab in your project, click **+ New Campaign**, pick a template, set the scheduled date and (optionally) override the subject. The full request shape (including `variables` for per-campaign substitutions) lives in the [Campaigns API reference](/api/campaigns#create-campaign).

::: tip Cookie auth only
Campaigns mutate workspace state and require role-based capabilities (`campaigns:write`). API keys, which are project-scoped and identity-less, can't call these endpoints — you'll get `401`. Schedule from the dashboard, or call the endpoints from your own UI built on the same cookie-session login.
:::

## List Campaigns

The **Newsletters** tab shows every campaign in the project, ordered by scheduled time, with their current status and live `sent_count` / `failed_count`. The same data is exposed at `GET /api/v1/projects/{id}/campaigns` — see the [API reference](/api/campaigns#list-campaigns).

## Cancel a Campaign

You can only delete or reschedule a campaign while its status is `scheduled`. From the dashboard, open the campaign row and use **Cancel** or **Edit**. Once it transitions to `sending`, `sent` or `failed`, the row becomes immutable.

## How the Worker Operates

The campaign worker runs as part of the SendDock backend process. It polls every 30 seconds for campaigns whose `scheduled_at` time has passed and whose status is still `scheduled`. When it finds one, it:

1. Sets the status to `sending`
2. Loads the template and all active subscribers
3. Sends the email to each subscriber (with variable replacement and unsubscribe URL injection)
4. Sets the status to `sent` (or `failed` if errors occurred)

No additional configuration is needed -- the worker starts automatically with the backend.

## Tips

- Use descriptive campaign names so you can identify them later in the list
- Schedule campaigns at least a few minutes in the future to allow time for review
- The template's subject line is used for the campaign emails
- Subscriber variables (`{{name}}`, `{{email}}`, `{{unsubscribe_url}}`) are replaced per recipient, just like broadcast
