# Subscribers

Subscribers are the people who receive your emails. Each subscriber belongs to a project and has a unique email within that project.

## Adding Subscribers

### Via the UI

Go to **Subscribers** in the project sidebar and click **+ Add Subscriber**. Provide an email and optional name.

### Via the API

```bash
curl -X POST https://your-instance.com/api/v1/projects/{id}/subscribers \
  -H "Authorization: Bearer sk_your_api_key" \
  -H "Content-Type: application/json" \
  -d '{"email": "user@example.com", "name": "John Doe"}'
```

## Import

Bulk-import subscribers from a CSV or JSON file. Open the **Subscribers** tab, click **Import**, and either drop a file onto the dropzone or pick one with the file picker.

### File formats

- **CSV** — first row is the header. Recognized columns: `email` (required), `name`, `status`. Extra columns are ignored.
- **JSON** — an array of `{ "email": "...", "name": "...", "status": "..." }` objects.

### Email validation

Every row goes through three checks before it lands in the database:

1. **Syntax** — the address must parse as a valid mailbox (RFC 5322).
2. **MX record** — SendDock resolves the domain's MX records. Domains with no MX (typo, dead domain) are rejected.
3. **Disposable-domain block-list** — a built-in list of throwaway providers (Mailinator, 10minutemail, etc.) is rejected by default.

Rows that fail any check are skipped. The **import results modal** shows five outcome cards:

- **Imported** — added to the project.
- **Updated** — already existed, name/status was refreshed.
- **Duplicates** — appeared more than once in the same file.
- **Suppressed** — on the project's [suppression list](./suppressions); skipped.
- **Rejected** — failed validation. Each rejected row is listed below with the reason (`invalid_syntax`, `no_mx`, `disposable`).

You can fix rejected rows in your source file and re-import — already-imported rows are deduplicated, so re-running the same file is safe.

## Subscriber Status

| Status | Description |
|--------|-------------|
| `active` | Receives emails normally |
| `pending` | Registered but not yet confirmed |
| `unsubscribed` | Opted out, will not receive emails |

Only `active` subscribers receive broadcast emails. Unsubscribed subscribers are also added to the project's [suppression list](./suppressions) so transactional sends skip them too.

## Managing Subscribers

From the subscribers table you can:

- **Activate** a pending or unsubscribed subscriber
- **Unsubscribe** an active subscriber
- **Delete** a subscriber permanently

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

This creates a subscriber with `pending` status and optionally sends a confirmation email. No authentication needed — safe to call from frontend JavaScript.

Use it to build pre-launch waitlists, beta signups, or email collection forms.

## API

See [Subscribers API](/api/subscribers) for the full REST API reference.
