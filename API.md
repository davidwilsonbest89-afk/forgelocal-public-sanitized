# BrowseForge API Reference

[繁體中文](API.zh-TW.md)

## Connection Information

| Service | URL | Purpose |
|------|-----|---------|
| REST API | `http://127.0.0.1:19280/api` | Profile and browser automation |
| Dashboard | `http://127.0.0.1:19280` | Web management UI |
| MCP Streamable HTTP | `http://127.0.0.1:19281` | AI agent integration |

## Authentication

All REST API endpoints except `/api/status` require a Bearer token:

```http
Authorization: Bearer <token>
```

The token is generated on first start and stored in `data/.api-token`. The Dashboard asks for the token on first use.

Error responses use stable, language-neutral `code` values:

```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "invalid or missing token"
  }
}
```

## Response Shape

Successful responses wrap data in `data`:

```json
{
  "data": {}
}
```

List endpoints may include `total`.

## System

### GET `/api/status`

Public health/status endpoint.

```bash
curl http://127.0.0.1:19280/api/status
```

```json
{
  "version": "1.7.0",
  "status": "ok"
}
```

### POST `/api/shutdown`

Closes browser sessions and stops the server.

```bash
curl -X POST http://127.0.0.1:19280/api/shutdown \
  -H "Authorization: Bearer $TOKEN"
```

## Profiles

### POST `/api/profiles`

Creates a browser profile.

```bash
curl -X POST http://127.0.0.1:19280/api/profiles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Brand Account 1",
    "engine": "firefox",
    "group": "Client A",
    "tags": ["facebook", "brand"],
    "proxy": {
      "type": "socks5",
      "host": "proxy.example.com",
      "port": 1080,
      "username": "user",
      "password": "pass"
    }
  }'
```

Fields:

| Field | Description |
|------|-------------|
| `name` | Required profile name |
| `engine` | `firefox` for Camoufox or `chromium` for CloakBrowser |
| `group` | Optional grouping label |
| `tags` | Optional string tags |
| `proxy` | Optional SOCKS5 or HTTP proxy configuration |
| `fingerprint` | Optional explicit fingerprint; auto-assigned when omitted |

### GET `/api/profiles`

Lists profiles.

```bash
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:19280/api/profiles
curl -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:19280/api/profiles?group=Client%20A&tag=facebook"
```

### GET `/api/profiles/{id}`

Gets a single profile.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6
```

### PUT `/api/profiles/{id}`

Updates profile metadata, proxy settings, or other mutable fields.

```bash
curl -X PUT http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "Renamed Profile", "group": "Client B"}'
```

### DELETE `/api/profiles/{id}`

Deletes a profile.

```bash
curl -X DELETE http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6 \
  -H "Authorization: Bearer $TOKEN"
```

### POST `/api/profiles/{id}/duplicate`

Creates a copy with a new profile ID and fingerprint while retaining group/proxy settings.

```bash
curl -X POST http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6/duplicate \
  -H "Authorization: Bearer $TOKEN"
```

### POST `/api/profiles/{id}/export`

Exports one profile as a ZIP archive.

```bash
curl -X POST http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6/export \
  -H "Authorization: Bearer $TOKEN" \
  -o profile.zip
```

### POST `/api/profiles/import`

Imports a profile ZIP.

```bash
curl -X POST http://127.0.0.1:19280/api/profiles/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@profile.zip"
```

## Sessions

### POST `/api/sessions`

Starts a browser session for a profile.

```bash
curl -X POST http://127.0.0.1:19280/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"profile_id": "prof_a1b2c3d4e5f6"}'
```

```json
{
  "data": {
    "session_id": "sess_prof_a1b2c3d4e5f6",
    "profile_id": "prof_a1b2c3d4e5f6",
    "engine": "firefox"
  }
}
```

### GET `/api/sessions`

Lists active sessions.

```bash
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:19280/api/sessions
```

### DELETE `/api/sessions/{id}`

Closes a browser session.

```bash
curl -X DELETE http://127.0.0.1:19280/api/sessions/sess_prof_a1b2c3d4e5f6 \
  -H "Authorization: Bearer $TOKEN"
```

## Browser Automation

All browser automation endpoints operate through Playwright and work across supported engines unless a runtime-specific limitation is documented.

### POST `/api/sessions/{id}/navigate`

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/navigate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com", "wait_until": "load"}'
```

`wait_until` accepts `load`, `domcontentloaded`, or `networkidle`.

### POST `/api/sessions/{id}/click`

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/click \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "button#login"}'
```

### POST `/api/sessions/{id}/type`

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/type \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "input[name=email]", "text": "user@example.com", "delay": 50}'
```

### POST `/api/sessions/{id}/eval`

Evaluates JavaScript in the page context.

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/eval \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"script": "document.title"}'
```

### GET `/api/sessions/{id}/screenshot`

Returns a PNG screenshot.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  "http://127.0.0.1:19280/api/sessions/$SID/screenshot?full_page=true" \
  -o screenshot.png
```

### GET `/api/sessions/{id}/content`

Returns full HTML or selected element text.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/$SID/content

curl -H "Authorization: Bearer $TOKEN" \
  "http://127.0.0.1:19280/api/sessions/$SID/content?selector=h1"
```

### POST `/api/sessions/{id}/wait`

Waits for a selector.

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/wait \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "#result", "timeout": 10000}'
```

### GET `/api/sessions/{id}/cookies`

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/$SID/cookies
```

### POST `/api/sessions/{id}/cookies`

Imports cookies into the session context.

```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/cookies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[{"name":"session","value":"abc123","domain":".example.com","path":"/"}]'
```

## Backup and Restore

### POST `/api/backup`

Exports all profiles as a ZIP archive.

```bash
curl -X POST http://127.0.0.1:19280/api/backup \
  -H "Authorization: Bearer $TOKEN" \
  -o browseforge-backup.zip
```

### POST `/api/restore`

Restores profiles from a backup ZIP. Existing profiles are not overwritten.

```bash
curl -X POST http://127.0.0.1:19280/api/restore \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@browseforge-backup.zip"
```

## Playwright Connect

### GET `/api/playwright/endpoint`

Lists Playwright-compatible endpoints for active sessions.

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/playwright/endpoint
```

### WebSocket `/api/playwright/ws/{session_id}`

Proxy endpoint for external Playwright clients.

```javascript
import { firefox } from 'playwright';

const browser = await firefox.connect(
  'ws://YOUR_SERVER:19280/api/playwright/ws/sess_prof_xxx',
  { headers: { Authorization: 'Bearer YOUR_TOKEN' } }
);
```

Compatibility notes:

- Use Playwright client `1.60.x`.
- Proxy mode requires only port `19280`.
- Bearer token authentication is required.

## Workflow

### POST `/api/workflow/run`

Runs a YAML workflow.

```bash
curl -X POST http://127.0.0.1:19280/api/workflow/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/yaml" \
  -d @examples/multi-login.yaml
```

Common actions:

- `create_profile`
- `open_browser`
- `close_browser`
- `navigate`
- `click`
- `type`
- `eval`
- `wait`
- `screenshot`
- `sleep`

## Security Notes

- Do not expose `19280`, `19281`, or `6901` directly to the public internet.
- Treat tokens, profiles, cookies, backups, and exported profile ZIPs as sensitive.
- Use SSH tunnels, VPN, or a hardened HTTPS reverse proxy for remote access.
