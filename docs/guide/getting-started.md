# Getting Started

This is the guided tour: from a fresh install to your first sent email, every step done from the dashboard. Each step links to a deeper guide if you want the details.

Already installed? Open SendDock in your browser and follow along. If not, see [Installation](/self-hosting/installation) first.

## 1. Create your admin account

On first launch SendDock sees there are no users and shows a **Setup** screen. Enter your name, email and a password — this becomes the owner account. You're logged in automatically, and a default workspace named **My Workspace** is created for you.

::: tip Teams
Workspaces let you invite teammates with roles (a Team-tier feature). You can ignore them for now — everything below works in your default workspace. See [Workspaces](./workspaces) when you're ready.
:::

## 2. Create your first project

A **project** is one sending identity — its own SMTP, subscribers, templates and logs. Most people start with one.

From the dashboard, click **+ New Project**, give it a name (and an optional description), and click **Create Project**.

![The New Project dialog](/screenshots/new-project-modal.png)

Open the project and you land on its **Overview** — totals, recent activity and a one-click composer once you've sent something.

![A project Overview with send totals and recent activity](/screenshots/project-overview.png)

→ More in [Projects](./projects).

## 3. Connect your SMTP

SendDock doesn't send mail itself — it relays through **your** SMTP provider (Mailgun, SES, Postmark, your own server…). Nothing sends until this is set.

In the project sidebar, open **SMTP Settings**, fill in host, port, username and password, optionally a From name and address, then **Save**. Hit **Test Connection** to confirm it works.

![The SMTP settings form](/screenshots/smtp.png)

→ Provider-specific setup in [SMTP Setup](./smtp).

## 4. Add subscribers

Open **Subscribers**. Add people the way that fits you:

- **+ Add Subscriber** — enter an email (and name, tags, custom fields) by hand.
- **Import CSV** — bulk-import a list, mapping extra columns to [custom fields](./subscribers#custom-fields).

![The subscribers table with tags and custom fields](/screenshots/subscribers-fields-tags.png)

::: tip Developers
You can also add and manage subscribers programmatically — see the [Subscribers API](/api/subscribers).
:::

→ More in [Subscribers](./subscribers).

## 5. Build a template

Open **Templates → + New Template**. Write your email in the **Code** tab (HTML) or the **Visual** tab (drag-and-drop blocks), and use variables like `{{name}}` or your own `{{custom.plan}}` for personalization. The preview updates live. **Save** when it looks right.

![The visual drag-and-drop template editor](/screenshots/editor.png)

→ More in [Templates](./templates), including [rich-text variables](./templates#rich-html-variables-opt-in).

## 6. Send your first email

From the project **Overview**, click **Send Email**. Pick your template, fill in any variables, and choose who gets it — a **specific address** to test, or your **whole list** (optionally scoped to a [segment](./segments)).

![The Send Email composer](/screenshots/send-modal.png)

That's the full loop. Everything after this — scheduling, analytics, deliverability — builds on these six steps.

→ More in [Email Sending](./sending).

## 7. (Optional) Get an API key

If you want to drive SendDock from your own app — say, add a subscriber when someone signs up, or fire a transactional email — create a key under **Settings → API Keys → + Create Key**. Copy it once (it's shown only that one time) and send it as `Authorization: Bearer sk_...`.

![Creating an API key in project settings](/screenshots/api-keys.png)

→ More in [API Keys](./api-keys) and the [API reference](/api/authentication).

## Where to next

- [Campaigns](./campaigns) — schedule sends for later
- [Analytics](./analytics) — opens, clicks, bounces and engagement
- [Segments](./segments) — target subsets of your list
- [Your account & security](./account) — enable 2FA, manage your plan
- [Suppressions](./suppressions) & [Bounces](./bounces) — keep your list clean
