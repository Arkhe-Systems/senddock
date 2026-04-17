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
| `{{unsubscribe_url}}` | Unsubscribe link (planned) |

Variables are replaced per subscriber when sending.

## API

See [Templates API](/api/templates) for the full REST API reference.
