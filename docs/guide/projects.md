# Projects

A project is the isolation boundary inside a [workspace](./workspaces). Each project has its own subscribers, templates, SMTP credentials, API keys, suppression list, audit log and (optionally) bounce mailbox. Sends from one project never see another project's data.

In a single-user (Free or Pro) install, you have one workspace with one or more projects in it. On Team, multiple users share workspaces and roles control access to all projects within.

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
- Manage API keys
- Configure the [bounce webhook](./bounces#2-public-webhook-endpoint) URL and IMAP poller
- Delete the project (requires typing the project name to confirm)

Each project also gets its own tabs in the sidebar for [Suppressions](./suppressions), [Webhooks](./webhooks) (Pro) and [Audit Log](./audit-log) (Pro) — they're scoped to the current project, not the workspace.

## API

See [Projects API](/api/projects) for the full REST API reference.
