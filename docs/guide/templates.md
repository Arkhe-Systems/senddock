# Templates

Templates define the content and structure of your emails. Each template belongs to a project.

## Creating a Template

Go to **Templates** in the project sidebar and click **+ New Template**. Give it a name, then use the editor to build the content.

## Editor Modes

### Code Editor

Write HTML directly with syntax highlighting powered by CodeMirror. A live preview panel shows the rendered output in real time.

### Visual Editor

Drag-and-drop email builder powered by GrapeJS. Available blocks:

**Layout:**
- Container (600px max-width email wrapper)
- Section
- 2 Columns / 3 Columns (table-based, email-safe)
- Divider
- Spacer

**Content:**
- Heading
- Text
- Image
- Button (table-based, works in all email clients)
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

Plus any custom keys you pass in the `data` map of `/send`, `/send/batch` or in a campaign's `variables` field — those are substituted by the same engine using the exact same `{{your_key}}` syntax.

Variables are replaced per recipient at send time.

### Safe by default — variables are HTML-escaped

When SendDock substitutes a variable inside the template body, the value runs through `html.EscapeString` first. So if a recipient's name happens to be `Bob <script>alert(1)</script>`, the rendered email shows the literal text — the `<script>` tag is encoded as `&lt;script&gt;` and never executes.

This applies to:

- The four built-in variables (`{{name}}`, `{{email}}`, `{{subscriber_id}}`, `{{unsubscribe_url}}`).
- Every key in the `data` / `variables` map you pass to a send or campaign.

The trade-off: you cannot inject HTML through a variable. If you genuinely need a dynamic chunk of HTML in your email (e.g. a different banner image per segment), build the HTML directly into the template body or split it into multiple templates rather than passing it as a variable.

The subject line is **not** escaped (subject is plain text, not HTML), but it is also not allowed to introduce headers — newlines are stripped to prevent SMTP header injection.

## API

See [Templates API](/api/templates) for the full REST API reference.
