# SMTP Setup

Each project requires SMTP configuration to send emails. SendDock connects to your SMTP server directly.

![The per-project SMTP settings form](/screenshots/smtp.png)

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

::: warning Why your local install times out
Most residential ISPs (Comcast, Claro, Movistar, BT, and many more) block outbound TCP ports 25, 465 and 587 at the network edge to prevent spam botnets. This means a SendDock instance running on your laptop or home server **cannot reach external SMTP providers** at all — it has nothing to do with SendDock or your credentials. You will see a 5-10s timeout error.

See [Diagnosing port issues](#diagnosing-port-issues) below to test which ports your network allows, and [Working around ISP blocks](#working-around-isp-blocks) for the unblocking paths.
:::

## Diagnosing port issues

Before assuming your SendDock is broken, test from inside the SendDock host whether the SMTP port is actually reachable. The check below probes the 4 most common ports against your SMTP server:

```bash
for port in 25 465 587 2525; do
  echo -n "port $port: "
  timeout 3 bash -c "echo > /dev/tcp/your-smtp-host.com/$port" 2>/dev/null \
    && echo "OPEN ✓" || echo "BLOCKED ✗"
done
```

Replace `your-smtp-host.com` with your actual server (`smtp.gmail.com`, `mail.yourdomain.com`, etc.).

A port marked **`BLOCKED`** can mean:
1. Your ISP filters outbound traffic on that port (most common cause on residential connections — see next section)
2. The SMTP server isn't listening on that port (if you control the server, you can add a listener — see [Self-hosted mail server](#self-hosted-mail-server-opening-extra-ports))
3. A firewall on the SMTP server's side drops connections from your IP

To go deeper and verify the port actually serves SMTP (not just accepts TCP), open a manual session — a healthy SMTP listener replies with a `220` banner within ~1s:

```bash
timeout 5 nc -v your-smtp-host.com 25
# Expected output starts with: 220 your-smtp-host.com ESMTP ...
```

## Working around ISP blocks

If the diagnostic shows everything BLOCKED (or only `25` open with no banner), the ISP is filtering. Five real fixes, ordered from easiest to most involved:

1. **Try port 25 explicitly.** Even when 465/587 are blocked, some ISPs leave 25 open. If your `nc` test above returned `25: OPEN` AND the SMTP server's manual session shows a banner, just set Port to `25` in the dashboard. SendDock auto-detects STARTTLS for that port.

2. **Use an SMTP provider that supports port 2525.** Brevo, Mailgun, SendGrid and Postmark all listen on `2525` specifically as an escape hatch for ISP-blocked users. Free tiers are enough for most self-hosters:
   - Brevo: `smtp-relay.brevo.com:2525` (300 emails/day free)
   - Mailgun: `smtp.mailgun.org:2525`
   - Postmark: `smtp.postmarkapp.com:2525`

3. **Run SendDock on a cloud server (DigitalOcean, Hetzner, AWS, etc.).** Cloud providers don't apply the residential SMTP block — port 587 + your existing provider just work. This is the canonical production deploy. See [Installation](/self-hosting/installation).

4. **Cloud SMTP relay.** Keep SendDock at home but spin up a tiny VPS ($4/mo on Hetzner) running Postfix in relay mode. SendDock → relay over a non-blocked port → relay forwards to your real provider on 587. Adds complexity but keeps email infra self-hosted.

5. **Local development only — use [Mailpit](https://mailpit.axllent.org/)** as a fake SMTP catcher. Accepts any auth on `localhost:1025`, shows captured emails in a web UI. Useful for the demo case but obviously doesn't deliver to real recipients:
   ```bash
   sudo docker run -d --name mailpit \
     --network senddock_default \
     -p 8025:8025 \
     axllent/mailpit
   ```
   Then in SMTP Settings: Host `mailpit`, Port `1025`, no auth. Inspect captured mail at `http://your-host:8025`.

## Self-hosted mail server — opening extra ports

If `mail.yourdomain.com` is **your own** mail server (Postal, Mailcow, Postfix, Stalwart, etc.), opening additional listening ports is often the cleanest fix — port 2525 isn't part of the standard SMTP block list and is usually allowed by residential ISPs.

### Postal

In `config/postal.yml`:

```yaml
smtp_server:
  enabled: true
  port: 25       # standard SMTP — keep for inter-server traffic
  bind_address: 0.0.0.0
  tls_enabled: true
  smtp_relay_port: 2525   # non-standard, residential-friendly
```

Reload Postal: `postal reload`. Don't forget to open the new port on your server's firewall (next section).

### Postfix

In `/etc/postfix/master.cf`, uncomment or add:

```
2525     inet  n       -       y       -       -       smtpd
  -o syslog_name=postfix/2525
  -o smtpd_tls_security_level=encrypt
  -o smtpd_sasl_auth_enable=yes
```

Restart: `sudo systemctl reload postfix`.

### Mailcow / Stalwart / others

Check their docs for "additional SMTP submission ports" or "alternate submission port". The principle is the same: add a listener on 2525 in addition to the defaults.

### Server-side firewall (UFW example)

After adding the listener on the mail server, allow inbound on that port:

```bash
sudo ufw allow 2525/tcp
sudo ufw status
```

Verify from a different network (or your SendDock host):

```bash
nc -zv mail.yourdomain.com 2525
# expected: Connection succeeded
```

## Host firewall (SendDock side)

Linux distros (Ubuntu, Debian, RHEL) ship with **outbound traffic wide open** by default. The SendDock host almost never blocks outbound SMTP — the problem is upstream (ISP) or downstream (mail server).

If you've explicitly enabled a host firewall in default-deny outbound mode (uncommon), allow the relevant ports:

```bash
# UFW
sudo ufw allow out 25/tcp
sudo ufw allow out 465/tcp
sudo ufw allow out 587/tcp
sudo ufw allow out 2525/tcp
```

::: tip Docker containers can reach the public internet
SendDock runs in a Docker container, but Docker's default bridge network NATs outbound traffic through your host's interface and inherits the host's connectivity. If the host can reach port 587, SendDock can too. No special Docker config is needed for outbound SMTP.
:::

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
