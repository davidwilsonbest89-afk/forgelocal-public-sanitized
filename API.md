# BrowseForge API Reference

[繁體中文](API.zh-TW.md)

## Connection Information

| Service | URL | Purpose |
|------|-----|---------|
| REST API | `http://127.0.0.1:19280/api` | Profile and browser automation |
| Dashboard | `http://127.0.0.1:19280` | Web management UI |
| MCP Streamable HTTP | `http://127.0.0.1:19280/mcp` | AI agent integration on the main service port |

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

## MCP Tools

BrowseForge exposes MCP tools at `http://127.0.0.1:19280/mcp` (Streamable HTTP transport, JSON-RPC 2.0). MCP is mounted on the main BrowseForge HTTP service port rather than a separate listener.

Migration note: older clients configured for a separate `:19281` MCP listener should update to the main service port plus `/mcp`.

### Authentication

All MCP requests require Bearer token authentication:

```http
Authorization: Bearer <token>
```

### Tool List

| Tool | Description |
|------|-------------|
| `list_profiles` | List all browser profiles |
| `create_profile` | Create a new browser profile |
| `delete_profile` | Delete a profile |
| `update_profile` | Update profile settings |
| `open_browser` | Open a browser session for a profile |
| `close_browser` | Close a browser session |
| `navigate` | Navigate to a URL |
| `click` | Click an element |
| `type_text` | Type text into an element |
| `screenshot` | Take a screenshot |
| `get_content` | Get page content |
| `evaluate` | Execute JavaScript |
| `new_tab` | Open a new tab |
| `list_tabs` | List all tabs |
| `switch_tab` | Switch to a tab |
| `close_tab` | Close a tab |
| `web_search` | Search Google using a profile-bound agent session |
| `web_explore` | Explore a webpage using a profile-bound agent session |
| `create_session` | Create an agent web session for a Chromium profile |
| `destroy_session` | Destroy an agent web session and close its page |
| `list_sessions` | List active agent web sessions |
| `gc_sessions` | Trigger web session garbage collection |

### Agent Web Sessions

`web_search` and `web_explore` run through profile-bound agent sessions:

- Only Chromium/CloakBrowser profiles are accepted for these tools.
- One profile has one persistent browser instance owned by `browser.Manager`.
- `SessionPool` connects to that browser through the profile's Playwright Bind endpoint.
- Each agent session opens one independent `Page` via `connectedBrowser.NewPage()`; it does not create a separate browser or `BrowserContext`.
- `session_id` pins later calls to the same page.
- GC/destroy/shutdown close idle agent pages and metadata only; they do **not** close the profile browser.

Defaults:

| Setting | Default |
|---------|---------|
| Idle TTL | 5 minutes |
| GC sweep interval | 1 minute |
| Max sessions per profile | 10 |

MCP response-shape compatibility note:

- BrowseForge keeps the standard MCP `content` text block as the primary human/client-compatible payload.
- For session-aware web tools, machine-readable session metadata is also exposed as top-level result fields: `session_id`, `profile_id`, and, when applicable, `session_created`.
- `web_search` also exposes top-level `extraction_mode`, `results`, and, when structured extraction has no results, `raw_fallback` for LLM-friendly SERP interpretation.
- This avoids breaking clients that only read `content` while still allowing agents to reuse pages reliably without parsing the text payload.

MCP error codes used by these tools:

| Code | Meaning |
|------|---------|
| `-32602` | Missing required argument, e.g. `query`, `url`, `profile_id`, or `session_id` |
| `-32000` | Runtime/session failure, e.g. session pool unavailable, profile not found, non-Chromium profile, browser launch/connect failure, navigation/search failure |

### `web_search`

Search Google and return structured results with title, URL, and snippet. If Google's DOM shape changes and structured extraction returns no results, BrowseForge returns an LLM-friendly raw SERP fallback containing page text and candidate links while preserving explicit captcha/consent/unusual-traffic errors.

**Parameters:**

| Parameter | Type | Required | Description |
|------|------|------|-------------|
| `query` | string | Yes | Search query |
| `profile_id` | string | Required if `session_id` is omitted | Chromium/CloakBrowser profile to use |
| `session_id` | string | No | Existing agent session to reuse; when omitted, a new session/page is created |
| `max_results` | number | No | Maximum results. Default `10`; values above `30` are clamped by `WebSearch` |

**Result shape:** `content[0].text` is a text prefix followed by pretty-printed JSON. Top-level result fields include `session_id`, `profile_id`, `session_created`, `extraction_mode`, `results`, and optional `raw_fallback`.

```json
{
  "content": [{
    "type": "text",
    "text": "Found 5 results for \"Go programming language tutorial\" (mode: structured):\n{...}"
  }],
  "session_id": "sess_search_0123abcd",
  "profile_id": "prof_abc123",
  "session_created": true,
  "extraction_mode": "structured",
  "results": [
    {"title": "Result title", "url": "https://example.com", "snippet": "Result snippet"}
  ]
}
```

When structured extraction is empty, `extraction_mode` is `raw_fallback` and `raw_fallback` contains:

```json
{
  "page_title": "Google Search",
  "text": "visible SERP text for LLM interpretation...",
  "candidate_links": [{"text": "Candidate", "url": "https://example.com"}]
}
```

### `web_explore`

Navigate to a URL and extract structured content: URL, title, optional meta description, visible text, and links.

**Parameters:**

| Parameter | Type | Required | Description |
|------|------|------|-------------|
| `url` | string | Yes | URL to explore; `https://` is prepended when no `http://` or `https://` prefix is provided |
| `profile_id` | string | Required if `session_id` is omitted | Chromium/CloakBrowser profile to use |
| `session_id` | string | No | Existing agent session to reuse; when omitted, a new session/page is created |
| `max_text_length` | number | No | Maximum text length. Default `3000`; extraction clamps to `10000` |
| `max_links` | number | No | Maximum links to extract. Default `50`; extraction clamps to `200` |

**Result shape:** `content[0].text` is pretty-printed JSON with `url`, `title`, `text`, `links`, and optional `description`. Top-level result fields also include `session_id`, `profile_id`, and `session_created`.

```json
{
  "content": [{
    "type": "text",
    "text": "{\n  \"url\": \"https://example.com\",\n  \"title\": \"Example Domain\",\n  \"text\": \"Example Domain...\",\n  \"links\": []\n}"
  }],
  "session_id": "sess_search_0123abcd",
  "profile_id": "prof_abc123",
  "session_created": true
}
```

### `create_session`

Create an agent web session for a Chromium/CloakBrowser profile without performing a search or page exploration.

**Parameters:** `profile_id` string, required.

**Returns:** text confirmation plus top-level `session_id` and `profile_id`.

```json
{
  "content": [{
    "type": "text",
    "text": "Session created: sess_search_0123abcd (profile: prof_abc123, browser: sess_prof_abc123)"
  }],
  "session_id": "sess_search_0123abcd",
  "profile_id": "prof_abc123"
}
```

### `destroy_session`

Destroy a session and close its agent page.

**Parameters:** `session_id` string, required.

**Returns:** text confirmation plus top-level `session_id`.

```json
{
  "content": [{
    "type": "text",
    "text": "Session destroyed: sess_search_0123abcd"
  }],
  "session_id": "sess_search_0123abcd"
}
```

### `list_sessions`

List active agent web sessions. Optional `profile_id` filters the result.

**Parameters:** `profile_id` string, optional.

**Returns:** `content[0].text` as pretty-printed JSON array of session info objects: `id`, `profile_id`, `browser_id`, `created_at`, `last_accessed`, and `idle_seconds`.

### `gc_sessions`

Run session GC immediately.

**Parameters:** none.

**Returns:** text confirmation such as `GC completed: closed 2 sessions`.

### Common Patterns

**Search and explore first result with the same agent session:**

```json
// Step 1: Search. Save returned session_id.
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "web_search",
    "arguments": {
      "query": "latest Go release notes",
      "profile_id": "prof_abc123"
    }
  }
}

// Step 2: Explore the first result URL using the same page/session.
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "web_explore",
    "arguments": {
      "url": "<url_from_step_1>",
      "session_id": "<session_id_from_step_1>"
    }
  }
}
```

## Security Notes

- Do not expose `19280` or `6901` directly to the public internet.
- Treat tokens, profiles, cookies, backups, and exported profile ZIPs as sensitive.
- Use SSH tunnels, VPN, or a hardened HTTPS reverse proxy for remote access.
