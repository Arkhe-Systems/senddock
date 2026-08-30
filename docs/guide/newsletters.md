# Newsletters

A newsletter is a named publication inside a project — "Dev Tips", "Product updates", "AI Weekly". One subscriber base, several publications: the same email address can be in two newsletters, leave one of them, and keep receiving the other. Before newsletters, unsubscribing was all-or-nothing; now a reader who is tired of one topic doesn't have to leave your whole list.

A project is home to **one subscriber base** and **as many newsletters as you like**. Even with several publications there's still a single list, a single [suppression list](/guide/suppressions) and a single project status for each subscriber — newsletters just add an extra *membership* on top. The branded [unsubscribe page](/guide/templates#page-templates) is likewise **one design per project**, shared by every newsletter (more on that below).

Newsletters live under each project at **Project → Newsletters**, and are part of the free Core.

## How membership works

Every subscriber keeps a single project-level status (`active` / `pending` / `unsubscribed`) — that's the legal on/off switch and nothing about newsletters changes it. On top of it, each subscriber holds **memberships**: which newsletters they belong to, and whether they've opted out of any of them.

So one subscriber can be `active` project-wide, a member of three newsletters, opted out of a fourth, and not a member of a fifth. Each of those is independent.

- A subscriber with no memberships behaves exactly as before — projects that never create a newsletter are unaffected.
- Unsubscribing from a newsletter only marks that membership; the subscriber stays `active` and keeps receiving your other newsletters and transactional email.
- The [suppression list](/guide/suppressions) still gates every send, always.

> Example: María is in **Dev Tips** and **Product updates**. She unsubscribes from **Dev Tips**. Her project status stays `active`, she keeps receiving **Product updates**, and transactional email is untouched. Re-adding her to **Dev Tips** later clears that opt-out.

## Creating and managing

1. Go to **Project → Newsletters** and click **+ New Newsletter**.
2. Name it (unique per project) and optionally describe it.
3. The table shows each newsletter's **active members** count and its ID — you'll need the ID for API sends.

Assign subscribers from **Project → Subscribers**: the edit modal has a checkbox per newsletter, and the bulk bar (select rows → **Newsletter**) adds or removes many at once. Re-adding someone who had left a newsletter clears their opt-out, so use it deliberately.

## Sending to a newsletter

In the broadcast composer (**Overview → Send Email**), pick a newsletter in the **audience** selector. Recipients are the active subscribers opted in to that newsletter. A send targets either a newsletter or a [segment](/guide/segments), not both. [Campaigns](/guide/campaigns) can also be tied to a newsletter.

## Per-newsletter unsubscribe

Emails sent to a newsletter carry an unsubscribe link scoped to it (the `List-Unsubscribe` header too, so Gmail's one-click unsubscribe behaves the same way). The confirmation page names the newsletter, and offers a secondary **Unsubscribe from all emails** option for readers who want out entirely.

All newsletters share **one unsubscribe page design** — the page template you pick under **Project → Settings → Unsubscribe page**. That's intentional: the design is reused, and the page itself is **per-newsletter at runtime**. The URL carries `?n={newsletterID}`, the heading reads `{{newsletter_name}}`, and the confirmation button says "Unsubscribe from {Newsletter}". So the same design renders correctly for every publication — you don't need one template per newsletter.

- Leaving a newsletter never touches the project status or the suppression list.
- Links sent before newsletters existed keep working and remain project-wide.
- If you delete a newsletter, its links degrade gracefully to a project-wide unsubscribe.

You can brand these pages with a page template — see [Templates](/guide/templates#page-templates).

## Filtering by newsletter

Once you send to newsletters, the rest of the dashboard follows:

- **Analytics** gains an *All project / newsletter* scope selector on the Overview, Campaigns and Engagement tabs.
- **Logs**, **Broadcasts** and **Subscribers** gain newsletter filters (subscribers also filter by status and tag).

## Newsletters vs segments

Both narrow a broadcast's audience, but they answer different questions. A **newsletter** is the reader's choice: a publication they joined and can leave. A **segment** is your query: a saved filter over status, tags and custom fields that the reader never sees. Use newsletters for distinct publications with their own unsubscribe; use segments to slice any audience by data.

## API

Newsletters are fully scriptable — CRUD, memberships, bulk actions, targeted sends and filters. See the [Newsletters API](/api/newsletters).
