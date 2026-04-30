# Workspaces

A workspace is a container for projects with its own list of members. Every project belongs to exactly one workspace, and authorization is scoped to membership: anyone in a workspace can use any project inside it, with the same permissions as anyone else in the same workspace.

## Why workspaces

Before workspaces, every project was tied to its creator's user account and only that user could access it. That made teamwork impossible without sharing logins, and lumping unrelated projects under one user mixed transactional and marketing concerns — same suppression list, same logs, same rate limits.

Workspaces fix the first problem and let you keep the second clean: split your *transactional* and *marketing* projects across two workspaces in the same SendDock instance, give one team access to one of them, and a bounce on marketing won't poison the suppression list of password-reset emails.

## Roles

Two roles, both workspace-wide:

| Role | Can do |
|---|---|
| **Owner** | Everything in the workspace, plus rename, delete, manage members, change roles. |
| **Member** | Full access to projects in the workspace (templates, subscribers, sends, broadcasts, settings, suppressions). Cannot manage members. |

Any owner can promote/demote other members. The system always guarantees at least one owner — the last owner cannot be removed or demoted.

## Default workspace

When you sign up — through the **Setup** screen on a fresh self-hosted instance, or **Register** in cloud mode — SendDock creates a workspace named `My Workspace` and makes you the owner. Existing instances were migrated the same way: every user got a `My Workspace`, and every project they had moved under it.

You can rename it from **Manage members** in the workspace switcher.

## The workspace switcher

The dashboard header has a switcher pinned to the left of `+ New Project`. It shows the **active** workspace name; click it to:

- Switch to another workspace you belong to.
- Open **Manage members** of the active workspace.
- Create a new workspace (`+ New workspace`).

The active workspace is remembered in your browser's `localStorage` (key `senddock.activeWorkspaceId`), so it survives refreshes and tab restarts. If the active one is deleted or you lose access to it, SendDock automatically falls back to the first workspace in your list.

## Adding members

From **Manage members** → `+ Add member`:

1. Enter the email of someone who **already has a SendDock account on this instance**. v0.6 does not send invitation emails — the user must register first (or be created via Setup), and then the owner adds them by email.
2. Pick a role (`member` by default).
3. The member appears in the table immediately and can sign in to see the workspace in their switcher.

If the email isn't registered yet, you'll see "no SendDock account uses that email yet". Have them register first; then add them.

## Changing roles & removing members

Owners can change any member's role from the inline `Role` select in the members table. Demoting the last owner to `member` is rejected — the workspace would have no owner left.

Removing a member revokes their access to every project in the workspace. The audit log records the action with the workspace ID and the affected user ID under `workspace.member_removed`.

## Deleting a workspace

Open the workspace, hit `Manage members → Rename` to change name, or use the API to `DELETE /workspaces/{id}`. SendDock refuses if the workspace still owns any project — move or delete the projects first. This guard exists to make a workspace deletion non-destructive: data only goes away when you explicitly delete the projects yourself.

## Pro features and licensing

The license check is deployment-wide, not per-workspace. A valid `SENDDOCK_LICENSE_KEY` unlocks Pro features (Analytics, Webhooks, Audit log) for **every** workspace on the instance. There is no per-workspace tier in v0.6 — that's reserved for cloud pricing later.

## What's next

- [Workspaces API](../api/workspaces) covers every endpoint with request/response shapes.
- [Code examples](../api/code-examples) shows multi-language snippets for the most common operations.
