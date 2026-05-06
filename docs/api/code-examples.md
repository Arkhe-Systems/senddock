# Code examples

Copy-paste-ready snippets for the most common SendDock operations in **cURL, JavaScript (Node), Python, Java, C# and PHP**.

Every example uses three placeholders — replace them once and the rest works:

| Placeholder | What goes there |
|---|---|
| `YOUR_BASE_URL` | The base URL of your SendDock instance, e.g. `https://senddock.example.com`. |
| `YOUR_API_KEY` | A project-scoped API key (`sk_…`). Create one in **Settings → API Keys**, see the [API Keys API](./api-keys). |
| `YOUR_PROJECT_ID` | The project UUID. It's the segment after `/projects/` in the dashboard URL. |

The API root is always `${YOUR_BASE_URL}/api/v1`. Authentication is `Authorization: Bearer ${YOUR_API_KEY}` for every endpoint that supports API keys (sending, broadcast, batch, subscribers/import). Cookie-only endpoints are noted on each [reference page](./authentication#endpoints-supporting-api-key-auth).

All request and response bodies are `application/json` unless stated otherwise.

## Send a single email (template)

Sends a template to one address. Variables in the template like `{{name}}` get replaced from `data`.

::: code-group

```bash [cURL]
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/send" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "template_id": "YOUR_TEMPLATE_ID",
    "data": { "name": "Jane" }
  }'
```

```js [JavaScript]
const res = await fetch(
  `${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/send`,
  {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${YOUR_API_KEY}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      to: 'user@example.com',
      template_id: 'YOUR_TEMPLATE_ID',
      data: { name: 'Jane' },
    }),
  }
)
if (!res.ok) throw new Error(await res.text())
const result = await res.json()
```

```python [Python]
import requests

res = requests.post(
    f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/send",
    headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
    json={
        "to": "user@example.com",
        "template_id": "YOUR_TEMPLATE_ID",
        "data": {"name": "Jane"},
    },
    timeout=10,
)
res.raise_for_status()
result = res.json()
```

```java [Java]
// Java 11+, no extra dependencies
import java.net.URI;
import java.net.http.*;
import java.time.Duration;

var body = """
    {"to":"user@example.com","template_id":"YOUR_TEMPLATE_ID","data":{"name":"Jane"}}
    """;

var req = HttpRequest.newBuilder()
    .uri(URI.create(YOUR_BASE_URL + "/api/v1/projects/" + YOUR_PROJECT_ID + "/send"))
    .header("Authorization", "Bearer " + YOUR_API_KEY)
    .header("Content-Type", "application/json")
    .timeout(Duration.ofSeconds(10))
    .POST(HttpRequest.BodyPublishers.ofString(body))
    .build();

var res = HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString());
if (res.statusCode() >= 400) throw new RuntimeException(res.body());
```

```csharp [.NET / C#]
// .NET 6+, System.Net.Http + System.Text.Json
using System.Net.Http;
using System.Net.Http.Json;
using System.Net.Http.Headers;

using var http = new HttpClient { BaseAddress = new Uri($"{YOUR_BASE_URL}/api/v1/") };
http.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", YOUR_API_KEY);

var payload = new {
    to = "user@example.com",
    template_id = "YOUR_TEMPLATE_ID",
    data = new { name = "Jane" }
};

var res = await http.PostAsJsonAsync($"projects/{YOUR_PROJECT_ID}/send", payload);
res.EnsureSuccessStatusCode();
var result = await res.Content.ReadFromJsonAsync<Dictionary<string, object>>();
```

```php [PHP]
<?php
$ch = curl_init("$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/send");
curl_setopt_array($ch, [
    CURLOPT_RETURNTRANSFER => true,
    CURLOPT_POST           => true,
    CURLOPT_HTTPHEADER     => [
        "Authorization: Bearer $YOUR_API_KEY",
        "Content-Type: application/json",
    ],
    CURLOPT_POSTFIELDS => json_encode([
        "to"          => "user@example.com",
        "template_id" => "YOUR_TEMPLATE_ID",
        "data"        => ["name" => "Jane"],
    ]),
]);
$body   = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);
if ($status >= 400) throw new RuntimeException($body);
$result = json_decode($body, true);
```

:::

A successful call returns `{"message": "sent"}` (or `{"sent": 1, "failed": 0}` for the subscriber variant). Common failure modes:

- `400` — body doesn't match any of the three send shapes documented in [Sending → Send](./sending#send).
- `401` — missing / wrong API key.
- `429` — per-project rate limit hit (60 req/min on `/send`). Honor `Retry-After`.

## Send raw HTML (no template)

For one-off transactional emails without creating a template first.

::: code-group

```bash [cURL]
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/send" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "to": "user@example.com",
    "subject": "Password reset",
    "html_body": "<p>Click <a href=\"https://example.com/reset?t=abc\">here</a> to reset.</p>"
  }'
```

```js [JavaScript]
await fetch(`${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/send`, {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${YOUR_API_KEY}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    to: 'user@example.com',
    subject: 'Password reset',
    html_body: '<p>Click <a href="https://example.com/reset?t=abc">here</a> to reset.</p>',
  }),
})
```

```python [Python]
requests.post(
    f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/send",
    headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
    json={
        "to": "user@example.com",
        "subject": "Password reset",
        "html_body": '<p>Click <a href="https://example.com/reset?t=abc">here</a> to reset.</p>',
    },
).raise_for_status()
```

```java [Java]
var body = """
    {"to":"user@example.com","subject":"Password reset","html_body":"<p>Click <a href=\\"https://example.com/reset?t=abc\\">here</a>.</p>"}
    """;
// Same HttpClient setup as the previous example.
```

```csharp [.NET / C#]
await http.PostAsJsonAsync($"projects/{YOUR_PROJECT_ID}/send", new {
    to        = "user@example.com",
    subject   = "Password reset",
    html_body = "<p>Click <a href=\"https://example.com/reset?t=abc\">here</a> to reset.</p>",
});
```

```php [PHP]
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    "to"        => "user@example.com",
    "subject"   => "Password reset",
    "html_body" => '<p>Click <a href="https://example.com/reset?t=abc">here</a> to reset.</p>',
]));
```

:::

`subject`, `html_body` and `to` are all required when no `template_id` is provided. The response is `{"message": "sent"}`.

## Batch send

Send the same template to many recipients with per-recipient variables. Up to ~5,000 recipients per call is comfortable; for full lists use [Broadcast](#broadcast).

::: code-group

```bash [cURL]
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/send/batch" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "YOUR_TEMPLATE_ID",
    "recipients": [
      {"to": "a@example.com", "data": {"name": "Alice"}},
      {"to": "b@example.com", "data": {"name": "Bob"}}
    ]
  }'
```

```js [JavaScript]
const recipients = users.map(u => ({ to: u.email, data: { name: u.name } }))

await fetch(`${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/send/batch`, {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${YOUR_API_KEY}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ template_id: 'YOUR_TEMPLATE_ID', recipients }),
})
```

```python [Python]
recipients = [{"to": u.email, "data": {"name": u.name}} for u in users]

requests.post(
    f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/send/batch",
    headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
    json={"template_id": "YOUR_TEMPLATE_ID", "recipients": recipients},
).raise_for_status()
```

```java [Java]
// Build the JSON with your library of choice (Jackson, Gson, ...).
// Conceptually:
//   {"template_id":"...","recipients":[{"to":"...","data":{...}}, ...]}
// Then POST to /projects/{id}/send/batch with the same headers as before.
```

```csharp [.NET / C#]
var recipients = users.Select(u => new {
    to   = u.Email,
    data = new { name = u.Name }
}).ToArray();

await http.PostAsJsonAsync(
    $"projects/{YOUR_PROJECT_ID}/send/batch",
    new { template_id = "YOUR_TEMPLATE_ID", recipients }
);
```

```php [PHP]
$recipients = array_map(fn($u) => [
    "to"   => $u["email"],
    "data" => ["name" => $u["name"]],
], $users);

curl_setopt($ch, CURLOPT_URL, "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/send/batch");
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    "template_id" => "YOUR_TEMPLATE_ID",
    "recipients"  => $recipients,
]));
```

:::

The response is `{"sent": N, "failed": M}`. `failed` includes both SMTP rejections and recipients that were on the project's suppression list at send time.

## Broadcast to all subscribers

Send a template to **every active subscriber** in the project. The unsubscribe link is injected automatically per recipient.

::: code-group

```bash [cURL]
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/broadcast" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"template_id": "YOUR_TEMPLATE_ID"}'
```

```js [JavaScript]
await fetch(`${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/broadcast`, {
  method: 'POST',
  headers: {
    Authorization: `Bearer ${YOUR_API_KEY}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({ template_id: 'YOUR_TEMPLATE_ID' }),
})
```

```python [Python]
requests.post(
    f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/broadcast",
    headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
    json={"template_id": "YOUR_TEMPLATE_ID"},
).raise_for_status()
```

```java [Java]
var body = """
    {"template_id":"YOUR_TEMPLATE_ID"}
    """;
// Same HttpClient setup as the first example.
// POST to /projects/{YOUR_PROJECT_ID}/broadcast.
```

```csharp [.NET / C#]
await http.PostAsJsonAsync(
    $"projects/{YOUR_PROJECT_ID}/broadcast",
    new { template_id = "YOUR_TEMPLATE_ID" }
);
```

```php [PHP]
curl_setopt($ch, CURLOPT_URL, "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/broadcast");
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode(["template_id" => "YOUR_TEMPLATE_ID"]));
```

:::

Returns `{"sent": N, "failed": M}` once delivery has been attempted for every active subscriber. Broadcast is rate-limited to **5 calls per minute per project** to prevent runaway sends.

## Add a subscriber

::: code-group

```bash [cURL]
curl -X POST "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/subscribers" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","name":"John Doe"}'
```

```js [JavaScript]
const res = await fetch(
  `${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/subscribers`,
  {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${YOUR_API_KEY}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ email: 'user@example.com', name: 'John Doe' }),
  }
)
if (res.status === 409) {
  // already a subscriber on this project
} else if (!res.ok) {
  throw new Error(await res.text())
}
```

```python [Python]
res = requests.post(
    f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/subscribers",
    headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
    json={"email": "user@example.com", "name": "John Doe"},
)
if res.status_code == 409:
    # duplicate — already exists
    pass
else:
    res.raise_for_status()
```

```java [Java]
var body = """
    {"email":"user@example.com","name":"John Doe"}
    """;

var req = HttpRequest.newBuilder()
    .uri(URI.create(YOUR_BASE_URL + "/api/v1/projects/" + YOUR_PROJECT_ID + "/subscribers"))
    .header("Authorization", "Bearer " + YOUR_API_KEY)
    .header("Content-Type", "application/json")
    .POST(HttpRequest.BodyPublishers.ofString(body))
    .build();

var res = HttpClient.newHttpClient().send(req, HttpResponse.BodyHandlers.ofString());
if (res.statusCode() == 409) {
    // already exists
} else if (res.statusCode() >= 400) {
    throw new RuntimeException(res.body());
}
```

```csharp [.NET / C#]
var res = await http.PostAsJsonAsync(
    $"projects/{YOUR_PROJECT_ID}/subscribers",
    new { email = "user@example.com", name = "John Doe" }
);
if (res.StatusCode == HttpStatusCode.Conflict) {
    // already exists
} else {
    res.EnsureSuccessStatusCode();
}
```

```php [PHP]
curl_setopt($ch, CURLOPT_URL, "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/subscribers");
curl_setopt($ch, CURLOPT_POSTFIELDS, json_encode([
    "email" => "user@example.com",
    "name"  => "John Doe",
]));
$body   = curl_exec($ch);
$status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
if ($status === 409) {
    // already exists
}
```

:::

`409 Conflict` means the email is already on this project. Returns `201 Created` with the full subscriber object on success — see [Subscribers → Add](./subscribers#add-subscriber).

## List subscribers (paginated)

::: code-group

```bash [cURL]
curl -G "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/subscribers" \
  -H "Authorization: Bearer $YOUR_API_KEY" \
  --data-urlencode "limit=100" \
  --data-urlencode "offset=0"
```

```js [JavaScript]
async function* paginate() {
  const limit = 100
  let offset = 0
  while (true) {
    const res = await fetch(
      `${YOUR_BASE_URL}/api/v1/projects/${YOUR_PROJECT_ID}/subscribers?limit=${limit}&offset=${offset}`,
      { headers: { Authorization: `Bearer ${YOUR_API_KEY}` } }
    )
    const { subscribers, total } = await res.json()
    yield* subscribers
    offset += subscribers.length
    if (offset >= total || subscribers.length === 0) break
  }
}

for await (const sub of paginate()) {
  console.log(sub.email)
}
```

```python [Python]
def all_subscribers():
    limit, offset = 100, 0
    while True:
        res = requests.get(
            f"{YOUR_BASE_URL}/api/v1/projects/{YOUR_PROJECT_ID}/subscribers",
            headers={"Authorization": f"Bearer {YOUR_API_KEY}"},
            params={"limit": limit, "offset": offset},
        )
        res.raise_for_status()
        page = res.json()
        yield from page["subscribers"]
        offset += len(page["subscribers"])
        if offset >= page["total"] or not page["subscribers"]:
            return

for sub in all_subscribers():
    print(sub["email"])
```

```java [Java]
// Requires Jackson (com.fasterxml.jackson.databind) for JSON parsing —
// the JDK has no built-in JSON. Gson works the same way.
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

var mapper = new ObjectMapper();
var http = HttpClient.newHttpClient();
int limit = 100, offset = 0;

while (true) {
    var req = HttpRequest.newBuilder()
        .uri(URI.create(YOUR_BASE_URL + "/api/v1/projects/" + YOUR_PROJECT_ID
                        + "/subscribers?limit=" + limit + "&offset=" + offset))
        .header("Authorization", "Bearer " + YOUR_API_KEY)
        .build();
    var res = http.send(req, HttpResponse.BodyHandlers.ofString());
    JsonNode page = mapper.readTree(res.body());
    JsonNode subs = page.get("subscribers");
    for (JsonNode sub : subs) {
        System.out.println(sub.get("email").asText());
    }
    offset += subs.size();
    if (offset >= page.get("total").asInt() || subs.isEmpty()) break;
}
```

```csharp [.NET / C#]
int limit = 100, offset = 0;
while (true) {
    var page = await http.GetFromJsonAsync<JsonElement>(
        $"projects/{YOUR_PROJECT_ID}/subscribers?limit={limit}&offset={offset}"
    );
    var subs = page.GetProperty("subscribers");
    foreach (var sub in subs.EnumerateArray()) {
        Console.WriteLine(sub.GetProperty("email").GetString());
    }
    offset += subs.GetArrayLength();
    if (offset >= page.GetProperty("total").GetInt32() || subs.GetArrayLength() == 0) break;
}
```

```php [PHP]
$limit = 100; $offset = 0;
while (true) {
    curl_setopt($ch, CURLOPT_URL,
        "$YOUR_BASE_URL/api/v1/projects/$YOUR_PROJECT_ID/subscribers?limit=$limit&offset=$offset");
    curl_setopt($ch, CURLOPT_HTTPGET, true);
    $page = json_decode(curl_exec($ch), true);
    foreach ($page["subscribers"] as $sub) {
        echo $sub["email"] . "\n";
    }
    $offset += count($page["subscribers"]);
    if ($offset >= $page["total"] || empty($page["subscribers"])) break;
}
```

:::

The endpoint returns `{ "subscribers": [...], "total": N }`. Use `total` (not the page size) to decide when to stop — the last page may be smaller than `limit`.

## Verify webhook signatures <Badge type="warning" text="Pro" />

Every webhook delivery includes an `X-SendDock-Signature: t=<unix>,v1=<hex>` header. Recompute the HMAC over `<t>.<raw_body>` with your webhook's secret and compare in constant time. Reject anything older than ~5 minutes for replay protection.

The [Webhooks guide](../guide/webhooks#verifying-the-signature) has minimal Node.js, Go and Python versions. The implementations below add timestamp-skew protection so a captured request can't be replayed indefinitely — Python, Java, C# and PHP.

::: code-group

```python [Python]
import hmac, hashlib, time

def verify(raw_body: bytes, header: str, secret: str, max_skew_s: int = 300) -> bool:
    parts = dict(p.split("=", 1) for p in header.split(","))
    t, sig = parts.get("t", ""), parts.get("v1", "")
    if not t or not sig:
        return False
    if abs(time.time() - int(t)) > max_skew_s:
        return False
    signed = f"{t}.".encode() + raw_body
    expected = hmac.new(secret.encode(), signed, hashlib.sha256).hexdigest()
    return hmac.compare_digest(expected, sig)
```

```java [Java]
import java.security.MessageDigest;
import java.time.Instant;
import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.util.HexFormat;

public boolean verify(byte[] rawBody, String header, String secret) throws Exception {
    var parts = header.split(",", 2);
    if (parts.length != 2) return false;
    var t   = parts[0].replaceFirst("^t=",  "");
    var sig = parts[1].replaceFirst("^v1=", "");

    long ts = Long.parseLong(t);
    if (Math.abs(Instant.now().getEpochSecond() - ts) > 300) return false;

    var mac = Mac.getInstance("HmacSHA256");
    mac.init(new SecretKeySpec(secret.getBytes(), "HmacSHA256"));
    mac.update((t + ".").getBytes());
    mac.update(rawBody);
    var expected = HexFormat.of().formatHex(mac.doFinal());

    return MessageDigest.isEqual(expected.getBytes(), sig.getBytes());
}
```

```csharp [.NET / C#]
using System.Security.Cryptography;
using System.Text;

bool Verify(byte[] rawBody, string header, string secret) {
    var parts = header.Split(',', 2);
    if (parts.Length != 2) return false;
    var t   = parts[0].Replace("t=",  "");
    var sig = parts[1].Replace("v1=", "");

    var ts = long.Parse(t);
    if (Math.Abs(DateTimeOffset.UtcNow.ToUnixTimeSeconds() - ts) > 300) return false;

    using var mac = new HMACSHA256(Encoding.UTF8.GetBytes(secret));
    var prefix = Encoding.UTF8.GetBytes(t + ".");
    mac.TransformBlock(prefix, 0, prefix.Length, null, 0);
    mac.TransformFinalBlock(rawBody, 0, rawBody.Length);
    var expected = Convert.ToHexString(mac.Hash!).ToLowerInvariant();

    return CryptographicOperations.FixedTimeEquals(
        Encoding.UTF8.GetBytes(expected),
        Encoding.UTF8.GetBytes(sig)
    );
}
```

```php [PHP]
function verify(string $rawBody, string $header, string $secret, int $maxSkew = 300): bool {
    $parts = [];
    foreach (explode(",", $header) as $part) {
        [$k, $v] = array_pad(explode("=", $part, 2), 2, "");
        $parts[$k] = $v;
    }
    $t   = $parts["t"]  ?? "";
    $sig = $parts["v1"] ?? "";
    if ($t === "" || $sig === "") return false;
    if (abs(time() - (int)$t) > $maxSkew) return false;

    $expected = hash_hmac("sha256", $t . "." . $rawBody, $secret);
    return hash_equals($expected, $sig);
}
```

:::

::: warning Read the raw body, not the parsed JSON
Most frameworks parse JSON automatically and lose insignificant whitespace, which breaks the HMAC. Capture the bytes **before** any parser touches them — `request.body()` in Express, `request.body` in FastAPI's raw form, `Request.InputStream` in ASP.NET, `php://input` in PHP.
:::

## Errors

All errors return JSON of the shape `{"error": "human-readable message"}` with the appropriate HTTP status:

| Status | Meaning |
|---|---|
| `400` | Body fails validation. The error string usually points at the offending field. |
| `401` | Missing / wrong API key. Cookie session expired and refresh failed. |
| `402` | Pro feature requested without a valid `SENDDOCK_LICENSE_KEY`. |
| `403` | The authenticated user / API key doesn't own this project. |
| `404` | Project / template / subscriber not found. |
| `409` | Duplicate (e.g. subscriber email already exists). |
| `422` | The recipient is on the project's suppression list. The send is logged but no SMTP attempt happens. |
| `429` | Rate limited. Honor `Retry-After` (in seconds). |
| `5xx` | Server error. Retry with backoff. |

## What's next

- The full request/response schema for every endpoint lives under [API Reference](./authentication).
- [Sending API](./sending) documents the three send shapes and the `data` variable substitution rules.
- [Webhooks](./webhooks) covers HMAC signing, retry policy and event types end-to-end.
- [API Keys](./api-keys) explains key creation, revocation and the `last_used_at` audit trail.
