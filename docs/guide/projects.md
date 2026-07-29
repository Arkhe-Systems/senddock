# Projects

A project is the isolation boundary inside a [workspace](./workspaces). Each project has its own subscribers, templates, SMTP credentials, API keys, suppression list, audit log and (optionally) bounce mailbox. Sends from one project never see another project's data.

In a single-user (Free or Pro) install, you have one workspace with one or more projects in it. On Team, multiple users share workspaces and roles control access to all projects within.

![The dashboard listing your projects](/screenshots/dashboard.png)

## Creating a Project

From the dashboard, click **+ New Project** and provide a name and optional description.

## Project Overview

Each project dashboard shows:

- **Total Emails** — total emails sent from this project
- **Sent** — successfully delivered emails
- **Failed** — emails that failed to send (with error details)
- **Recent Activity** — last 10 emails sent

## SMTP Configuration

Each project needs its own SMTP configuration to send emails. Go to **SMTP Settings** in the sidebar and configure your SMTP server.

See [SMTP Setup](/guide/smtp) for details.

## Project Settings

In **Settings** you can:

- Edit project name and description
- Copy the project ID (for API usage)
- Manage [API keys](./api-keys) and [custom fields](./subscribers#custom-fields)
- Configure the [bounce webhook](./bounces#2-public-webhook-endpoint) URL and IMAP poller
- Delete the project from the **danger zone** (requires typing the project name to confirm)

Each project also gets its own tabs in the sidebar for [Suppressions](./suppressions), [Webhooks](./webhooks) and [Audit Log](./audit-log) (Pro) — they're scoped to the current project, not the workspace.

### Deleting a project

There are two ways to delete a project, both of which ask you to **type the project's name** to confirm (it's irreversible — subscribers, templates, logs and campaigns all go with it):

- From the **dashboard**, hover a project card and click **Delete**.
- From inside the project, **Settings → danger zone → Delete project**.

## API

See [Projects API](/api/projects) for the full REST API reference.
