# Workspaces

A workspace is a container for projects with its own list of members. Every project belongs to exactly one workspace, and authorization is scoped to membership: anyone in a workspace can use any project inside it, with the same permissions as anyone else in the same workspace.

## Why workspaces

Before workspaces, every project was tied to its creator's user account and only that user could access it. That made teamwork impossible without sharing logins, and lumping unrelated projects under one user mixed transactional and marketing concerns — same suppression list, same logs, same rate limits.

Workspaces fix the first problem and let you keep the second clean: split your *transactional* and *marketing* projects across two workspaces in the same SendDock instance, give one team access to one of them, and a bounce on marketing won't poison the suppression list of password-reset emails.

## Roles & capabilities

Four assignable roles, with a fixed capability matrix:

| Capability | Owner | Admin | Developer | Viewer |
|---|---|---|---|---|
| Manage members & delete workspace | ✓ | — | — | — |
| Project settings / SMTP / Bounce config | ✓ | ✓ | — | — |
| Create / edit templates | ✓ | ✓ | — | — |
| Create / edit / import subscribers | ✓ | ✓ | — | — |
| Manage API keys | ✓ | ✓ | — | — |
| Send transactional (`POST /send`) | ✓ | ✓ | ✓ | — |
| Broadcast / batch send / campaigns | ✓ | ✓ | — | — |
| Manage suppressions | ✓ | ✓ | — | — |
| Manage webhooks (Pro) | ✓ | ✓ | — | — |
| Read templates, subscribers, logs, analytics, audit log | ✓ | ✓ | ✓ | ✓ |

A few notes:

- **Developer** is the role for the rest of your team's services. They can call `/send` from a backend job for password resets and one-off transactional emails, but they can't broadcast to your subscriber list, edit your branded templates, or rotate API keys.
- **Viewer** is read-only — useful for support staff who need to look up an email log or analytics chart without any risk of writing.
- **API keys are project-scoped, not user-scoped.** Anyone with a key has full project access on the API surface that supports keys (sending, broadcast, batch). Roles only constrain UI / cookie-auth users, so carve up developer access by what API keys you hand out, not by what role they have.
- The system always guarantees at least one owner — the last owner cannot be removed or demoted.

## Default workspace

When you sign up — through the **Setup** screen on a fresh self-hosted instance, or **Register** in cloud mode — SendDock creates a workspace named `My Workspace` and makes you the owner. Existing instances were migrated the same way: every user got a `My Workspace`, and every project they had moved under it.

You can rename it from **Manage members** in the workspace switcher.

## The workspace switcher

The dashboard header has a switcher pinned to the left of `+ New Project`. It shows the **active** workspace name; click it to:

- Switch to another workspace you belong to.
- Open **Manage members** of the active workspace.
- Create a new workspace (`+ New workspace`).

The active workspace is remembered in your browser's `localStorage` (key `senddock.activeWorkspaceId`), so it survives refreshes and tab restarts. If the active one is deleted or you lose access to it, SendDock automatically falls back to the first workspace in your list.

## Adding members <Badge type="warning" text="Team" />

Member management is gated on a valid `SENDDOCK_LICENSE_KEY`. Without one, every workspace is a single-user surface — you can still create as many workspaces as you want to organize your projects, you just can't share them.

There are two paths to add a member, both from **Manage members**:

### Add an existing user

If the person already has a SendDock account on this instance:

1. `+ Add existing` → enter their email → pick role.
2. They appear in the table immediately and the workspace shows up in their switcher.

If the email isn't registered, the form returns "no SendDock account uses that email yet" and you fall through to the second path.

### Create a new user

For self-hosted, public registration is disabled — so the *only* way for a new teammate to get an account is for an owner to create it:

1. `+ Create user` → email + name + temporary password + role.
2. SendDock creates the user account, hashes the password, and adds them to the workspace at the chosen role in a single transaction.
3. Pass the temporary password to the user out of band (1Password, signed message, in person). They can change it later.

## Changing roles & removing members

Owners can change any member's role from the inline `Role` select in the members table. Demoting the last owner to `member` is rejected — the workspace would have no owner left.

Removing a member revokes their access to every project in the workspace. The audit log records the action with the workspace ID and the affected user ID under `workspace.member_removed`.

## Deleting a workspace

Open the workspace, hit `Manage members → Rename` to change name, or use the API to `DELETE /workspaces/{id}`. SendDock refuses if the workspace still owns any project — move or delete the projects first. This guard exists to make a workspace deletion non-destructive: data only goes away when you explicitly delete the projects yourself.

## Plans & licensing

The license check is deployment-wide, not per-workspace. SendDock has two paid tiers, both unlocked through the same `SENDDOCK_LICENSE_KEY` (the validator decides which features the key entitles you to):

| Plan | What it unlocks |
|---|---|
| **Pro** | Analytics dashboard, Webhooks, Audit log. |
| **Team** | Everything in Pro **plus** multi-member workspaces, role management and admin user creation. |

Without any license, member management endpoints return `402 Payment Required` and the Members page renders a paywall card. The single-user flow stays free: you can still have multiple workspaces, organize your projects, and use Core features unchanged.

## What's next

- [Workspaces API](../api/workspaces) covers every endpoint with request/response shapes.
- [Code examples](../api/code-examples) shows multi-language snippets for the most common operations.
