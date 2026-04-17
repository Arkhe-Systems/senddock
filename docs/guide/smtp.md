# SMTP Setup

Each project requires SMTP configuration to send emails. SendDock connects to your SMTP server directly.

## Configuration

Go to **SMTP Settings** in the project sidebar and fill in:

| Field | Description | Example |
|-------|-------------|---------|
| SMTP Host | Your SMTP server hostname | `smtp.gmail.com` |
| Port | SMTP port (usually 587 for TLS) | `587` |
| Username | SMTP authentication username | `you@gmail.com` |
| Password | SMTP password or app-specific password | `xxxx xxxx xxxx xxxx` |
| From Name | Display name for the sender (optional) | `My Newsletter` |
| From Email | Email shown as sender (optional, defaults to username) | `noreply@mydomain.com` |

## Testing

After saving, click **Test Connection**. SendDock will send a test email to the configured from address (or SMTP username) to verify the connection works.

## Common SMTP Providers

### Gmail
- Host: `smtp.gmail.com`
- Port: `587`
- Username: your Gmail address
- Password: [App Password](https://support.google.com/accounts/answer/185833) (not your regular password)

### Amazon SES
- Host: `email-smtp.{region}.amazonaws.com`
- Port: `587`
- Username: SES SMTP username (from IAM)
- Password: SES SMTP password (from IAM)

### Mailgun
- Host: `smtp.mailgun.org`
- Port: `587`
- Username: your Mailgun SMTP username
- Password: your Mailgun SMTP password

### Resend
- Host: `smtp.resend.com`
- Port: `465`
- Username: `resend`
- Password: your Resend API key

### Custom / Self-hosted (Postfix, etc.)
- Host: your server's hostname or IP
- Port: `25`, `465`, or `587`
- Username/Password: as configured on your server

## From Email vs SMTP Username

The **SMTP username** is used for authentication. The **From Email** is what recipients see as the sender address. These can be different if your SMTP provider allows it (e.g., sending from `noreply@yourdomain.com` while authenticating with `smtp-user@provider.com`).

If From Email is not set, the SMTP username is used as the sender address.
