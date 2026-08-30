# Templates

Templates define the content and structure of your emails. Each template belongs to a project.

![The templates list for a project](/screenshots/templates-list.png)

## Creating a Template

Two ways to start:

- **+ New Template** — opens a blank editor. Name it, write your HTML or build it visually.
- **★ Browse library** — pick a starter from the community library and clone it into your project. See [Template library](#template-library) below.

Either way, the template lives in your project and is fully editable from there.

When creating from scratch you also pick the template's **type**: an **Email** template is what you send, a **Page** template is rendered as a web page — see [Page templates](#page-templates). The type is fixed at creation.

Each template also has a **subject line**, set in the editor above the content. It's the default subject for every send and can be overridden per send, batch, broadcast or campaign via the optional `subject` field. Template variables (`{{name}}`, `{{custom.*}}`, …) are substituted in the subject just like the body, per recipient.

## Editor Modes

### Code Editor

Write HTML directly with syntax highlighting powered by CodeMirror. A live preview panel shows the rendered output in real time.

![The template editor: code on the left, live preview on the right, detected variables below](/screenshots/template-editor.png)

### Visual Editor

Drag-and-drop email builder powered by GrapeJS.

![The visual drag-and-drop template editor](/screenshots/editor.png)

Templates are sanitized when loaded into the editor (scripts, iframes and event handlers stripped), so pasted or imported HTML can't run in the canvas. Undo/redo works while you edit, and switching between the **Code** and **Visual** tabs no longer resets the canvas.

Available blocks:

**Layout:**
- Container (600px max-width email wrapper)
- Section
- 2 Columns / 3 Columns (div-based, so you can drag content in and out of them)
- Divider
- Spacer

**Content:**
- Heading
- Text
- Image
- Button (a single styled link — drags as one unit)
- Link
- List

**Pre-built Sections:**
- Header (dark banner with title)
- Footer (with unsubscribe link)
- CTA Section (call to action with button)

### Style Manager

When using the visual editor, select any element to edit:

- Typography (font family, size, weight, color, alignment)
- Background color
- Spacing (padding, margin)
- Size (width, height, max-width)
- Border (radius, width, style, color)

## Template Variables

Use double curly braces to insert dynamic content:

| Variable | Replaced with |
|----------|--------------|
| `{{name}}` | Subscriber's name |
| `{{email}}` | Subscriber's email |
| `{{subscriber_id}}` | Subscriber's UUID |
| `{{unsubscribe_url}}` | Per-recipient unsubscribe link, signed with HMAC |
| <span v-pre>`{{custom.KEY}}`</span> | A [custom field](/guide/subscribers#custom-fields) value for that subscriber, e.g. <span v-pre>`{{custom.plan_tier}}`</span> |

Plus any custom keys you pass in the `data` map of `/send`, `/send/batch` or in a campaign's `variables` field — those are substituted by the same engine using the exact same `{{your_key}}` syntax.

Below the editor, a variable panel lists **every** available token as a <span v-pre>`{{variable}}`</span> pill — the ones detected in the template, the built-in system variables, and the project's custom fields — grouped and clickable to insert at the cursor. Subscriber custom fields are hidden when you're editing a **page** template, since they aren't rendered on a public unsubscribe page.

Variables are replaced per recipient at send time.

### Safe by default — variables are HTML-escaped

When SendDock substitutes a variable inside the template body, the value runs through `html.EscapeString` first. So if a recipient's name happens to be `Bob <script>alert(1)</script>`, the rendered email shows the literal text — the `<script>` tag is encoded as `&lt;script&gt;` and never executes.

This is the default, and it always applies to:

- The four built-in variables (`{{name}}`, `{{email}}`, `{{subscriber_id}}`, `{{unsubscribe_url}}`).
- Custom field values (<span v-pre>`{{custom.KEY}}`</span>).
- Every escaped key in the `data` / `variables` map you pass to a send or campaign.

### Rich (HTML) variables — opt-in

If you genuinely need a formatted chunk — a newsletter body with bold, lists and links behind a single <span v-pre>`{{content}}`</span> tag — you can mark a specific variable as **rich**. In the Send Email dialog, each variable field has a **Text / Rich** toggle; switch it to Rich and a small WYSIWYG editor replaces the plain input. On the API you pass the field names in an `html_fields` list.

A rich value is **not** escaped — instead it runs through a strict HTML **sanitizer** (formatting and safe links survive; `<script>`, event handlers and unknown tags are stripped), then inserted as HTML. It's per-field and opt-in, so plain values stay escaped unless you deliberately choose otherwise.

::: warning Subscriber data is never rich
`{{name}}`, `{{email}}`, `{{subscriber_id}}` and every <span v-pre>`{{custom.KEY}}`</span> are **always escaped** — you cannot mark subscriber-sourced values as HTML. Only the values you type into the send are eligible to be rich.
:::

See [Email Sending → Rich-text variables](/guide/sending#rich-text-variables) for the full workflow.

The subject line is **not** escaped (subject is plain text, not HTML), but it is also not allowed to introduce headers — newlines are stripped to prevent SMTP header injection.

## Page templates

A template created with type **Page** is rendered as a public web page instead of an email. Today pages serve one purpose: **branded unsubscribe pages**. Assign one under **Project → Settings → Unsubscribe page** and both the unsubscribe confirmation and the "you're unsubscribed" page render with your design instead of the neutral built-in page.

This is **one design per project** — every [newsletter](./newsletters) uses the same page, but the page is **per-newsletter at runtime**: the URL carries `?n={newsletterID}`, `{{newsletter_name}}` names the publication, and the confirmation button reads "Unsubscribe from {Newsletter}". So you brand it once and it adapts to whichever newsletter the reader is leaving. See [Newsletters → Per-newsletter unsubscribe](./newsletters#per-newsletter-unsubscribe).

[Newsletter](./newsletters) links get the same treatment, with the newsletter's name in place.

Page templates support these placeholders:

| Placeholder | Renders as |
|---|---|
| `{{project_name}}` | The project's name |
| `{{email}}` | The recipient's address |
| `{{newsletter_name}}` | The newsletter being left; falls back to the project name on project-wide links |
| `{{confirm_button}}` | The unsubscribe action itself. If you leave it out, SendDock injects it anyway — a page can never lose the button |
| `{{unsubscribe_all_link}}` | On newsletter pages, a link to unsubscribe from all emails |

A few rules keep the page safe and predictable:

- The HTML is sanitized before rendering: scripts, iframes, forms, and event-handler attributes are stripped. Placeholder values are escaped, so they can't inject markup.
- Style with **inline `style` attributes** or a `<style>` block — both are preserved and inlined when the page renders. Images and tables work normally.
- An invalid or tampered unsubscribe link always shows the built-in error page, never your branded one.
- Page templates never appear in the broadcast or campaign template pickers, and email templates can't be assigned as pages.

The templates list has an **All / Email / Page** filter, and page templates carry a badge so the two kinds don't mix up.

## Managing templates

Back on the templates list, each row has two quick actions:

- **Copy ID** — copies the template's UUID to your clipboard. You need it when sending through the [API](/api/sending) (the `template_id` field) or referencing the template from your own code.
- **Delete** — removes the template. To avoid accidents you're asked to **type the template's name** to confirm. Deleting a template does not affect emails already sent with it; it only stops you from selecting it in future sends.

## Template library

Beyond writing templates from scratch, SendDock ships with a community-maintained starter library. Click **★ Browse library** on the Templates page to open it.

![The template library browser with categories](/screenshots/template-library.png)

The library covers common email scenarios — welcome flows, monthly digests, single-story newsletters, product launches, changelogs, weekly link roundups, password resets, email verifications, and transactional receipts. Each template uses Handlebars variables (so the personalization works out of the box) and inline CSS (so it renders correctly in Outlook, Gmail and the rest).

Pick one, click **Use template**, and SendDock clones the HTML into your project as a new editable template. From that point on it's yours — edit it freely; future changes to the library don't affect your copy.

### Filtering by category

The library is grouped into five categories:

| Category | When to use |
|---|---|
| **welcome** | First email a new subscriber or trial user receives |
| **newsletter** | Recurring content sends (monthly digest, single-story essay) |
| **announcement** | One-off broadcasts (product launches, release notes) |
| **digest** | Curated link roundups |
| **transactional** | Triggered by a user action (password resets, receipts, verifications). These don't include unsubscribe links — they're not marketing, recipients didn't subscribe to them. |

### Where templates live

The library is served from a separate public repo: [`Arkhe-Systems/senddock-templates`](https://github.com/Arkhe-Systems/senddock-templates). Every SendDock instance (Cloud and self-hosted) fetches the manifest from `raw.githubusercontent.com` and caches it in Redis for one hour. New templates added to the repo are live in every instance within an hour — no redeploy, no version bump.

Self-hosters who want to point at a private fork or a curated internal library can override the source with the [`TEMPLATE_LIBRARY_URL`](/guide/environment#optional) environment variable. The schema is documented in the public repo's `CONTRIBUTING.md`.

### Contributing a template

Templates are content, not code, and the repo is MIT-licensed and open to community PRs — **one template per PR**. The full guide lives in [CONTRIBUTING.md](https://github.com/Arkhe-Systems/senddock-templates/blob/main/CONTRIBUTING.md); the short version:

1. Pick a kebab-case `id` (e.g. `welcome-friendly`, `newsletter-startup-weekly`).
2. Add `templates/<id>.html` — inline CSS, responsive, no `<script>` tags, no external assets.
3. Add `templates/<id>.png` — 600 × 400 PNG thumbnail, ≤ 100 KB.
4. Add an entry to `index.json` with name, category, description, and declared variables.
5. Run `python3 scripts/validate.py` locally to catch errors.
6. Open the PR.

CI validates schema, file presence, size limits, URL conventions, and rejects anything with a `<script>` tag. A maintainer reviews the visual and the copy quality before merging.

This is the only SendDock repo that accepts external code contributions — the [core engine](https://github.com/Arkhe-Systems/senddock) is core-team-only by design — but templates are content, so the door is open.

## API

See [Templates API](/api/templates) for the full REST API reference.
