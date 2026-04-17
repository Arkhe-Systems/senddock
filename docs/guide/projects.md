# Projects

A project is an isolated workspace in SendDock. Each project has its own subscribers, templates, SMTP configuration, and API keys.

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
- Delete the project (requires typing the project name to confirm)

## API

See [Projects API](/api/projects) for the full REST API reference.
