# Subscribers

Subscribers are the people who receive your emails. Each subscriber belongs to a project and has a unique email within that project.

![The subscribers table with custom-field columns (Plan tier, Country, Signup date) and tags](/screenshots/subscribers-fields-tags.png)

## Adding Subscribers

### Via the UI

Go to **Subscribers** in the project sidebar and click **+ Add Subscriber**. Enter an email and an optional name — and, right there in the same dialog, add **tags** and fill in any **[custom fields](#custom-fields)** the project defines, so the subscriber lands fully populated in one step.

![The Add Subscriber dialog with tags and custom fields](/screenshots/add-subscriber-modal.png)

### Via the API

For programmatic ingestion of multiple subscribers (e.g. from a CRM sync, an existing user database, or a CSV upload in your own UI), use [`POST /api/v1/projects/{id}/subscribers/import`](/api/subscribers#bulk-import) — it accepts both cookie auth and `Authorization: Bearer sk_...` API keys, takes a top-level JSON array, and runs every row through the same email validation as the dashboard import.

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/subscribers/import \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '[
    {"email": "user@example.com", "name": "John Doe"},
    {"email": "other@example.com", "name": "Jane Smith"}
  ]'
```

The body is the array itself — **not** wrapped in `{ "rows": [...] }`. Pass `?validate_mx=false` and/or `?validate_disposable=false` as query params to relax validation for sources you already trust.

The single-recipient `POST /subscribers` is cookie-only and used by the dashboard's "+ Add Subscriber" button — for a one-shot programmatic add, just send a one-row import array.

For waitlist forms on landing pages, use the public [`/waitlist` endpoint](/api/subscribers#waitlist-public) — no auth required, and it sets the subscriber's status to `pending` instead of `active`.

## Import

Bulk-import subscribers from a **CSV** file. Open the **Subscribers** tab, click **Import**, and either drop a `.csv` onto the dropzone or pick one with the file picker.

![The CSV import dialog](/screenshots/import-modal.png)

### CSV format

- The first row is the **header**. Recognized columns: `email` (required) and `name`. Any extra column whose header matches a [custom field](#custom-fields) key or label is imported into that field; unmatched extra columns are ignored.
- New subscribers are created with status `active`.

::: tip Importing JSON, or from your own code
The dashboard importer takes CSV only. To import a JSON array (with `fields` and `tags` per row), or to sync from a CRM/database, use the [import API](/api/subscribers#bulk-import) instead — see [Via the API](#via-the-api) above.
:::

### Email validation

Every row goes through three checks before it lands in the database:

1. **Syntax** — the address must parse as a valid mailbox (RFC 5322).
2. **MX record** — SendDock resolves the domain's MX records. Domains with no MX (typo, dead domain) are rejected.
3. **Disposable-domain block-list** — a built-in list of throwaway providers (Mailinator, 10minutemail, etc.) is rejected by default.

Rows that fail any check are skipped. The **import results modal** breaks the input down into:

- **Imported** — new subscribers actually inserted.
- **Duplicates** — emails that already exist on the project. Silently skipped (not re-fetched, not updated). Re-importing the same file is therefore safe.
- **Suppressed** — emails on the project's [suppression list](./suppressions); skipped without insert.
- **Rejected** — failed validation. Each rejected row is listed below with the reason (`syntax_invalid`, `no_mx`, `disposable`).

The four categories above — imported, duplicates, suppressed and rejected — sum exactly to the number of rows in your file, so every input row is accounted for.

You can fix rejected rows in your source file and re-import. Existing rows are deduplicated, so re-running the same file just makes the new rows go through and the rest count as duplicates.

## Subscriber Status

| Status | Description |
|--------|-------------|
| `active` | Receives emails normally |
| `pending` | Registered but not yet confirmed |
| `unsubscribed` | Opted out, will not receive emails |

Only `active` subscribers receive broadcast emails. Unsubscribed subscribers are also added to the project's [suppression list](./suppressions) so transactional sends skip them too.

A `pending` subscriber never receives broadcasts and does not become `active` on its own — activate them from the subscribers table, the bulk action, or the API. There is no automatic confirmation link: `pending` is a manual holding state for leads you haven't approved yet.

Status is the project-wide switch. On top of it, subscribers can hold per-[newsletter](./newsletters) memberships — leaving one newsletter doesn't change the status or suppress anything.

## Managing Subscribers

From the subscribers table, on a single row you can:

- **Activate** a pending or unsubscribed subscriber
- **Unsubscribe** an active subscriber
- **Edit** a subscriber's name, custom fields, tags and [newsletter](./newsletters) memberships
- **Delete** a subscriber permanently

The table filters by **status**, **tag** and **newsletter**, combinable with each other.

### Bulk actions

Tick the checkboxes on several rows (or the header checkbox to select the page) and a bulk-action bar appears. From it you can, across every selected subscriber at once:

- **Change status** — set them all to active, unsubscribed or pending.
- **Tags** — add or remove tags (see [Tags](#tags)).
- **Newsletter** — add them all to or remove them all from one [newsletter](./newsletters).
- **Delete** — remove them all permanently (asks you to confirm first).

## Custom Fields

Out of the box a subscriber has `email`, `name` and `status`. **Custom fields** let you store typed attributes beyond that — `plan_tier`, `country`, `birthday`, `signup_source`, whatever your use case needs.

::: tip Two levels — this is the key thing to understand
A custom field is **defined once for the whole project** (that's the part that "applies to all subscribers"), and then each subscriber holds its **own value** for it. You don't create fields from the Subscribers table — you create the *definition* in **Settings**, and the Subscribers table then gains a column and an input for it.
:::

![The Custom Fields section in project Settings, listing each definition with its key and type](/screenshots/custom-fields-settings.png)

You define fields once per project under **Project → Settings → Custom Fields**. Each definition has:

- a **key** (machine name, e.g. `plan_tier`) and a **label** for the UI;
- a **type** — `string`, `number`, `date`, `boolean`, or `enum` (a dropdown with a fixed option list);
- an optional **required** flag.

![Creating a custom field: key, label, type picker and required toggle](/screenshots/custom-fields-modal.png)

Once defined, the field shows up as an input in the add/edit subscriber modal (with the right control per type), as a column in the subscribers table, and as a mappable column in the CSV importer. Values are validated on write — a `number` field rejects non-numbers, an `enum` rejects values outside its options, unknown keys are rejected entirely.

![The edit-subscriber modal rendering each custom field with a type-appropriate input, plus a tag editor](/screenshots/subscribers-edit-modal.png)

Custom fields pay off in two places:

- **Templates** — reference them as <span v-pre>`{{custom.KEY}}`</span>, e.g. <span v-pre>`Hi {{name}}, your {{custom.plan_tier}} plan renews soon.`</span> See [Templates → variables](/guide/templates#template-variables).
- **Segments** — filter on them (`plan_tier is pro`, `signup_date after …`). See [Segments](/guide/segments).

Deleting a definition leaves existing subscriber values in place until each subscriber's next write.

## Tags

Tags are lightweight, free-form labels — `vip`, `beta`, `paid`, `newsletter`. Unlike custom fields there's nothing to define first: type a tag and it exists.

Add or edit tags from the add/edit subscriber modal, or select rows in the subscribers table and use the **Tags** bulk action to add or remove tags across many subscribers at once. Tags are a first-class [segment](/guide/segments) filter (has any of / has all of / has none of), which makes them the quickest way to carve out an audience for a broadcast.

## Constraints

- Email must be unique per project (the same email can exist in different projects)
- Deleting a project deletes all its subscribers

## Waitlist

SendDock includes a public waitlist endpoint for collecting emails from landing pages without exposing API keys.

```
POST /api/v1/projects/{id}/waitlist
```

```json
{"email": "user@example.com", "template_id": "uuid"}
```

This creates a subscriber with `pending` status. Include `template_id` — the ID of an `email`-type template — to send that template as a confirmation email (the address is available to the template as the `{{email}}` variable); omit it to create the subscriber silently. The email is informational: it does **not** flip the subscriber to `active`, which stays a manual step. No authentication needed — safe to call from frontend JavaScript.

Use it to build pre-launch waitlists, beta signups, or email collection forms.

## API

See [Subscribers API](/api/subscribers) for the full REST API reference.
