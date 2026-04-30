# Workspaces API

Every project belongs to a workspace. Endpoints under `/workspaces` manage workspaces and their member list. All endpoints require cookie authentication. See the [Workspaces guide](../guide/workspaces) for the conceptual model.

All errors share the shape `{"error": "human-readable message"}`.

## List workspaces

```
GET /api/v1/workspaces
```

Returns every workspace the authenticated user is a member of, ordered by creation date (oldest first).

```json
{
  "workspaces": [
    {
      "id": "uuid",
      "name": "My Workspace",
      "created_by": "uuid",
      "created_at": "2026-04-30T12:00:00Z",
      "updated_at": "2026-04-30T12:00:00Z"
    }
  ]
}
```

## Create workspace

```
POST /api/v1/workspaces
```

```json
{ "name": "Acme Marketing" }
```

The caller becomes the workspace's first owner.

**Response — 201 Created**

```json
{
  "id": "uuid",
  "name": "Acme Marketing",
  "created_by": "uuid",
  "created_at": "2026-04-30T12:00:00Z",
  "updated_at": "2026-04-30T12:00:00Z",
  "role": "owner"
}
```

## Rename workspace

```
PATCH /api/v1/workspaces/{id}
```

Owner only.

```json
{ "name": "New Name" }
```

Returns the updated workspace.

## Delete workspace

```
DELETE /api/v1/workspaces/{id}
```

Owner only. Returns `204 No Content`. Returns `409 Conflict` (`{"error":"workspace still has projects"}`) if any project still belongs to the workspace — delete those first.

## List projects in a workspace

```
GET /api/v1/workspaces/{id}/projects
```

Returns the projects of the workspace that the caller has access to (i.e. all of them, since the caller is a member). Same shape as `GET /projects`.

## List members

```
GET /api/v1/workspaces/{id}/members
```

```json
{
  "members": [
    {
      "user_id": "uuid",
      "email": "alice@example.com",
      "name": "Alice",
      "role": "owner",
      "joined_at": "2026-04-30T12:00:00Z"
    }
  ]
}
```

## Add member

```
POST /api/v1/workspaces/{id}/members
```

Owner only.

```json
{ "email": "bob@example.com", "role": "member" }
```

The user must already have a SendDock account on this instance — v0.6 does not send invitation emails. Returns `404` (`{"error":"user not found"}`) if no account uses that email.

`role` defaults to `member`. Valid values: `owner`, `member`.

**Response — 201 Created**

```json
{
  "user_id": "uuid",
  "email": "bob@example.com",
  "name": "Bob",
  "role": "member",
  "joined_at": "2026-04-30T12:00:00Z"
}
```

If the user is already a member, the call updates their role and returns the same shape.

## Change member role

```
PATCH /api/v1/workspaces/{id}/members/{userId}
```

Owner only. Returns `204 No Content`.

```json
{ "role": "owner" }
```

Returns `409 Conflict` (`{"error":"cannot remove the last owner"}`) if the change would leave the workspace without an owner.

## Remove member

```
DELETE /api/v1/workspaces/{id}/members/{userId}
```

Owner can remove any member. A member can also remove themselves (leave the workspace). Returns `204 No Content`. Returns `409 Conflict` if removing the user would leave the workspace without an owner.

The removed user's access to every project in the workspace is revoked immediately.

## Errors

| Status | Meaning |
|---|---|
| `400` | Body fails validation, unknown role. |
| `401` | Missing / wrong session cookie. |
| `403` | Action requires owner role, or caller is not a member of the workspace. |
| `404` | Workspace, member or invited user not found. |
| `409` | Last-owner guard, or workspace still has projects on delete. |
| `500` | Server error. |

## Audit events

These events land in the project audit log only when scoped to a project; workspace-level events are recorded under `target_type=workspace`:

- `workspace.create`
- `workspace.rename`
- `workspace.delete`
- `workspace.member_added`
- `workspace.member_role_changed`
- `workspace.member_removed`
