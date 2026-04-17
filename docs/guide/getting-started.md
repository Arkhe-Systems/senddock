# Getting Started

After installing SendDock (see [Installation](/self-hosting/installation)), open it in your browser.

## Setup Screen

On first launch, SendDock detects there are no users and shows the setup screen. Create your admin account with name, email, and password. You'll be logged in automatically.

## Creating Your First Project

1. From the dashboard, click **+ New Project**
2. Give it a name and optional description
3. Click **Create Project**

## Configuring SMTP

Before you can send emails, configure your SMTP server:

1. Open your project
2. Go to **SMTP Settings** in the sidebar
3. Enter your SMTP host, port, username, and password
4. Optionally set a From Name and From Email
5. Click **Save Settings**
6. Click **Test Connection** to verify it works

See [SMTP Setup](/guide/smtp) for provider-specific instructions.

## Adding Subscribers

Go to **Subscribers** in the sidebar:

- Click **+ Add Subscriber** to add manually
- Or use the [API](/api/subscribers) to add them programmatically

## Building a Template

Go to **Templates** in the sidebar:

1. Click **+ New Template**
2. Use the **Code** tab to write HTML or the **Visual** tab for drag-and-drop
3. Use variables like `{{name}}` and `{{email}}` for personalization
4. The preview panel shows the rendered output in real time
5. Click **Save**

## Sending Emails

You can send emails via the [API](/api/sending):

- **Send to subscriber** — send a template to a specific subscriber
- **Broadcast** — send a template to all active subscribers
- **Direct send** — send a one-off email to any address

## Generating API Keys

To use the API from external applications:

1. Open your project
2. Go to **Settings** in the sidebar
3. Under **API Keys**, click **+ Create Key**
4. Copy the key immediately (it's only shown once)

Use it with `Authorization: Bearer sk_...` in your requests.

## Next Steps

- [Projects](/guide/projects) — managing multiple projects
- [Subscribers](/guide/subscribers) — subscriber statuses and management
- [Templates](/guide/templates) — code editor, visual editor, variables
- [Email Sending](/guide/sending) — send, broadcast, direct send
- [API Keys](/guide/api-keys) — authentication for external apps
- [SMTP Setup](/guide/smtp) — provider-specific configuration
- [Environment Variables](/guide/environment) — all configuration options
