# Your account & security

Per-user settings live on the **Account** and **Billing** pages, reachable from the avatar menu at the bottom of the sidebar. They cover password changes, two-factor authentication, recovery codes, and your license / plan.

![The account and security page](/screenshots/account.png)

## Account page (`/account`)

Open the avatar menu → **Account**. The page is split into three cards:

### Profile

Read-only summary of your identity: name, email, current plan badge (Free / Pro / Team for self-host, Free / Starter / Growth / Scale for cloud) and the date you created the account.

Name and email edits aren't supported yet — they need an email re-verification flow that isn't in the binary. For now, create a new user (Team license required) or restore from a Postgres dump if you really need to change either.

### Two-factor authentication

TOTP (Time-based One-Time Password) with single-use recovery codes. Works the same self-hosted or on cloud. **Enable it on every account that has dashboard access** — it's the single biggest improvement you can make to your instance's security posture.

#### Enable 2FA

1. On the Account page, click **Enable** in the Two-factor authentication card.
2. The setup modal generates a TOTP secret and renders a QR code.
3. Scan the QR with any authenticator app — Google Authenticator, 1Password, Authy, Bitwarden, Aegis. Or copy the base32 secret and paste it in manually.
4. Enter the 6-digit code your app shows and click **Verify and enable**.
5. **Save the recovery codes shown next** — this is the only time they'll appear. Copy them to a password manager, download the `senddock-recovery-codes-YYYY-MM-DD.txt` file, or print them. Each code unlocks the account once.

After enabling, every subsequent login becomes a two-step flow: password first, then either the current TOTP code or one recovery code.

#### Disable 2FA

Click **Disable** on the Account page. You must enter a valid TOTP code **or** a recovery code to confirm — there is no admin override and no "I lost my phone" shortcut. This is intentional: if it were easier to disable, the protection would be worthless.

#### Regenerate recovery codes

Click **Regenerate codes** to invalidate the existing set and generate ten fresh ones. You'll need a current TOTP code to authorize the change. Do this after using a recovery code so you stay at ten unused ones, and after any incident where you suspect the old set may have been seen.

#### Lost your authenticator device and recovery codes?

For self-hosted instances, the only recovery path is direct database access:

```sql
-- Connect to Postgres and clear the 2FA columns on the affected user.
UPDATE users
SET totp_secret = NULL,
    totp_enabled = false,
    totp_verified_at = NULL
WHERE email = 'you@example.com';

-- Then invalidate any active recovery codes for that user.
DELETE FROM user_recovery_codes WHERE user_id = (SELECT id FROM users WHERE email = 'you@example.com');
```

The next login will skip the 2FA step. Re-enable 2FA immediately from the Account page and save the new recovery codes somewhere safer this time.

For cloud accounts, contact support — there's no SSH-into-the-database escape hatch.

### Change password

Three-field form: current password, new password, confirm. The new password is checked against the same rules as registration:

- ≥ 8 characters
- at least one uppercase letter
- at least one digit
- at least one special character

The current password is verified server-side with bcrypt before the change applies. Changing the password does **not** sign out existing sessions on other devices — clear them by deleting the session cookies in those browsers, or wait for the JWT to expire.

## Billing page (`/billing`)

Open the avatar menu → **Billing**. What you see depends on whether you're self-hosting or on cloud.

### Self-hosted

- **Current plan** — Free, Pro or Team, from the license you've activated.
- When a license is active: `expires_at` (when the subscription renews or lapses) and `last_check` (when the validator last reached Lemon Squeezy). The validator caches the last-good response for 24 h, so a brief network outage won't lock you out.
- When no key is set: paywall cards for Pro ($9/mo) and Team ($29/mo) linking directly to Lemon Squeezy checkout.
- A note clarifying that SendDock charges **only** for features. BYO SMTP means you pay your SMTP provider for delivery — never SendDock.

To upgrade, click the checkout link, complete the purchase, then activate the key from the dashboard under **Instance → License** — it's stored in the database and unlocks the new tier **immediately**, no restart. See [Instance settings → Pro license](./instance-settings#pro-license). (The old `SENDDOCK_LICENSE_KEY` env var still works but is deprecated; support ends in v0.9.)

### Cloud

- **Current plan** — Free, Starter, Growth or Scale, computed from the active subscription.
- Subscriber usage gauge (e.g. `1,847 / 10,000 subscribers`) and the upgrade prompt when you cross 80%.
- One-click checkout to upgrade to a higher tier; portal access to change card / cancel / update billing email.

Both paths skip the per-send fees other tools charge — pricing is by self-host license tier or by subscriber count on cloud, never per email.

## Login flow with 2FA enabled

```
POST /api/v1/auth/login        { email, password }
→ 200 { needs_2fa: true, intermediate_token }
POST /api/v1/auth/2fa          { intermediate_token, code }      # TOTP or recovery code
→ 200 + session cookie
```

The intermediate token is short-lived (a few minutes) and only valid for completing the 2FA step. If you're integrating against the API rather than the dashboard, expect this two-step response shape whenever the target account has 2FA on.

## Related endpoints

| Endpoint | Purpose |
|---|---|
| `POST /api/v1/me/password` | Change password (current + new + confirm). |
| `POST /api/v1/me/2fa/setup` | Generate TOTP secret + recovery codes. Returns `otpauth_url`, base32 secret, ten recovery codes. |
| `POST /api/v1/me/2fa/verify` | Confirm setup with a 6-digit code. Enables 2FA. |
| `POST /api/v1/me/2fa/disable` | Disable 2FA. Requires a TOTP code or a recovery code. |
| `POST /api/v1/me/2fa/recovery-codes` | Regenerate the recovery code set. Requires a TOTP code. |
| `POST /api/v1/auth/2fa` | Login second step. Trades `intermediate_token` + code for a session. |
| `GET /api/v1/me` | Current user (`user_id`, `email`, `name`, `plan`, `created_at`). |

See also [Security Checklist](/self-hosting/configuration#security-checklist) for the broader hardening pass before you expose SendDock to the internet.
