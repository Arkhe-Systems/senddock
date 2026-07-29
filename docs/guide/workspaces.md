# Workspaces

A workspace is a container for projects. Every project belongs to exactly one workspace, and it's the boundary that keeps unrelated work — and, on Team, unrelated teammates — cleanly separated. Organizing projects into workspaces is **free**; *sharing* a workspace with other people is a Team feature (see [Members & roles](./members)).

## Why workspaces

Before workspaces, every project was tied to its creator's user account and only that user could access it. That made teamwork impossible without sharing logins, and lumping unrelated projects under one user mixed transactional and marketing concerns — same suppression list, same logs, same rate limits.

Workspaces fix the first problem and let you keep the second clean: split your *transactional* and *marketing* projects across two workspaces in the same SendDock instance, give one team access to one of them, and a bounce on marketing won't poison the suppression list of password-reset emails.

## The workspace switcher

The dashboard header has a switcher pinned to the left of `+ New Project`. It shows the **active** workspace name; click it to:

- Switch to another workspace you belong to.
- Open **Manage workspace** for the active workspace.
- Create a new workspace (`+ New workspace`).

![The dashboard with the workspace switcher](/screenshots/dashboard.png)

The active workspace is remembered in your browser's `localStorage` (key `senddock.activeWorkspaceId`), so it survives refreshes and tab restarts. If the active one is deleted or you lose access to it, SendDock automatically falls back to the first workspace in your list.

## Default workspace

When you sign up — through the **Setup** screen on a fresh self-hosted instance, or **Register** in cloud mode — SendDock creates a workspace named `My Workspace` and makes you the owner. Existing instances were migrated the same way: every user got a `My Workspace`, and every project they had moved under it.

## Sharing with a team

On the **Team** tier you can invite other people into a workspace and give each of them a role (owner / admin / developer / viewer) that scopes what they can do across its projects. That whole flow — inviting, creating users, changing roles — lives in [Members & roles](./members).

## Renaming & deleting a workspace

Open the workspace from the dashboard's workspace switcher and choose **Manage workspace**. From that screen the **owner** can:

- **Rename** the workspace.
- **Delete** it — the red `Delete` button next to Rename. It's owner-only and hidden when it's your last workspace (so you're never left with none). A confirm dialog asks you to confirm before it goes.

SendDock **refuses to delete a workspace that still owns any project** — move or delete the projects first. This guard makes deletion non-destructive: data only goes away when you explicitly delete the projects yourself. The deletion is recorded in the [audit log](./audit-log) as `workspace.delete`.

The same operation is available on the API as `DELETE /workspaces/{id}` with the identical guard.

## Plans & licensing

The license check is deployment-wide, not per-workspace. SendDock has two paid tiers, both unlocked through the same license key (the validator decides which features the key entitles you to):

| Plan | What it unlocks |
|---|---|
| **Pro** | Deliverability, Report builder, Audit log. (The Analytics dashboard is free in Core.) |
| **Team** | Everything in Pro **plus** [multi-member workspaces, role management and admin user creation](./members). |

Without any license the single-user flow stays free: you can still have multiple workspaces, organize your projects, and use Core features unchanged.

## What's next

- [Members & roles](./members) — invite teammates and assign roles (Team).
- [Workspaces API](../api/workspaces) covers every endpoint with request/response shapes.
