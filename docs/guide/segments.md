# Segments

A segment is a saved filter over your subscribers. Instead of always sending to *everyone*, you can target "only customers", "people tagged `vip`", "subscribers whose `plan_tier` is `pro`", or any combination. Segments are re-evaluated every time you use them, so membership is always current — there is no stale list to refresh.

Segments live under each project at **Project → Segments**, and are part of the free Core.

## Building blocks

A segment combines rules with a **match mode**:

- **All rules (AND)** — a subscriber must satisfy every rule.
- **Any rule (OR)** — a subscriber matching any single rule is included.

Each rule targets one of three things:

| Rule on | Operators | Example |
|---|---|---|
| **Status** | is / is not | status *is* `active` |
| **Tags** | has any of / has all of / has none of | tags *has any of* `vip`, `beta` |
| **Custom fields** | is / is not / contains / greater than / less than | `plan_tier` *is* `pro`, `signup_date` *after* a date |

Tags and [custom fields](/guide/subscribers#custom-fields) are what make segments expressive — status alone only gets you so far. Custom-field rules read the typed values you defined per project, so a `number` field supports greater/less-than, an `enum` field offers its allowed options, and so on.

## Creating a segment

1. Go to **Project → Segments** and click **+ New Segment**.
2. Give it a name and pick the match mode.
3. Add rules. As you edit, a live counter shows how many **active** subscribers currently match.
4. Save.

![The segment builder combining a status, tag and custom-field rule, with a live match count](/screenshots/segments-builder.png)

## Using a segment

When you send a broadcast (**Overview → Send Email → Newsletter**), a **segment selector** appears above the template picker. Leave it on *All active subscribers* to send to everyone, or pick a segment to send only to its members. Segment sends always exclude unsubscribed and suppressed addresses, exactly like a normal broadcast.

On Pro deployments, the [Analytics](/guide/analytics) dashboard gains a segment filter too, so you can read open/click performance for just that audience.

## API

Segments are fully scriptable — create, update, preview (live count) and target them from broadcasts. See the [Segments API](/api/segments).
