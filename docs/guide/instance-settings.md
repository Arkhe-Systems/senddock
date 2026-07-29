# Instance settings

On a self-hosted deploy, the things that used to live in your `.env` — the public URL and the Pro license — are now configured from the dashboard, under **Instance**. They're stored in the database, apply **live** (no container restart), and survive upgrades because they travel with your Postgres backup.

![Instance settings](/screenshots/instance-settings.png)

::: info Self-hosted only
This screen is for self-hosted instances. On the managed cloud (`CLOUD=true`) these values are owned by the environment and this screen is hidden — we manage them for you.
:::

Open it at **dashboard → Instance** (visible to the instance owner).

## Public URL

The single most important setting. It's the address recipients' mail clients reach back to, and it drives three things:

- The **unsubscribe** links inside every outgoing email.
- The **open-tracking pixel** and **click-redirect** URLs (`/t/...`, `/c/...`) — without a reachable public URL, no opens or clicks are ever recorded.
- The **CORS origin** the dashboard is served from.

Set it to your public HTTPS domain, e.g. `https://mail.example.com`. Changes take effect immediately.

::: warning A loopback URL disables broadcasts
If the public URL is `localhost`, `127.0.0.1` or `0.0.0.0`, broadcasts to your whole list are blocked — every recipient would get a broken unsubscribe link, which is both illegal and a strong spam signal. One-off sends to a specific address still work for testing. Point it at a real domain to enable broadcasts.
:::

## Session timeout

Signs users out after a stretch of inactivity. Enter any whole number of **minutes between 5 and 1440** (24 hours); the default is 120. Lower it on shared or high-sensitivity instances; raise it if constant re-logins get in the way.

## Pro license

If you bought Pro or Team, paste the license key you received by email into the **License** section and activate it. SendDock validates it, stores it **encrypted** in the database, and unlocks the paid features **immediately** — the validator rebuilds in place, no restart. The current plan (Free / Pro / Team) is shown once it's active.

See [Plans and licensing](/self-hosting/configuration#plans-and-licensing) for what each tier unlocks.

## Migrating from environment variables

Earlier versions read `PUBLIC_URL` and `SENDDOCK_LICENSE_KEY` from the environment. On a self-hosted instance those are now **deprecated**:

- If a value is still present in your `.env` and the database has none yet, SendDock **imports it once** on boot and logs a deprecation notice.
- From then on, the dashboard value is authoritative — editing the `.env` no longer changes anything.
- Support for both env vars will be **removed in v0.9**. Move them into this screen at your convenience.

You'll see the notice on stdout, e.g. `DEPRECATION: PUBLIC_URL is now configured from the dashboard under Instance Settings…`.
