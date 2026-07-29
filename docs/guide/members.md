# Members & roles <Badge type="warning" text="Team" />

On the **Team** tier a [workspace](./workspaces) stops being single-user: you invite teammates, give each a role, and they get scoped access to every project in that workspace. Member management is gated on a valid Team license (activated under [Instance → License](/guide/instance-settings#pro-license)); without one, every workspace is a single-user surface — you can still create as many as you want to organize your projects, you just can't share them.

Everything below is done from **Manage workspace** in the workspace switcher.

![The workspace member list with per-member roles](/screenshots/workspace-members.png)

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
| Manage webhooks | ✓ | ✓ | — | — |
| Read templates, subscribers, logs, analytics, audit log | ✓ | ✓ | ✓ | ✓ |

A few notes:

- **Developer** is the role for the rest of your team's services. They can call `/send` from a backend job for password resets and one-off transactional emails, but they can't broadcast to your subscriber list, edit your branded templates, or rotate API keys.
- **Viewer** is read-only — useful for support staff who need to look up an email log or analytics chart without any risk of writing.
- **API keys are project-scoped, not user-scoped.** Anyone with a key has full access on the five endpoints that accept keys: `/send`, `/send/batch`, `/broadcast`, `/subscribers/import`, `/stats`. Roles only constrain cookie-auth users (the dashboard), so carve up backend access by which API keys you hand out — not by which role you give a teammate.
- The system always guarantees at least one owner — the last owner cannot be removed or demoted.

## Adding members

There are two paths to add a member, both from **Manage workspace**:

### Add an existing user

If the person already has a SendDock account on this instance:

1. `+ Add existing` → enter their email → pick role.
2. They appear in the table immediately and the workspace shows up in their switcher.

If the email isn't registered, the form returns "no SendDock account uses that email yet" and you fall through to the second path.

### Create a new user

For self-hosted, public registration is disabled — so the *only* way for a new teammate to get an account is for an owner to create it:

![Creating a user and assigning a role](/screenshots/team-create-user.png)

1. `+ Create user` → email + name + temporary password + role.
2. SendDock creates the user account, hashes the password, and adds them to the workspace at the chosen role in a single transaction.
3. Pass the temporary password to the user out of band (1Password, signed message, in person). They can change it later.

## Changing roles & removing members

Owners can change any member's role from the inline `Role` select in the members table. Demoting the last owner to any non-owner role is rejected — the workspace would have no owner left.

Removing a member revokes their access to every project in the workspace. The [audit log](./audit-log) records the action with the workspace ID and the affected user ID under `workspace.member_removed`.

## See also

- [Workspaces](./workspaces) — the (free) concept these members are scoped to.
- [Workspaces API](../api/workspaces) — every member endpoint with request/response shapes.
