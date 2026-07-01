# Developer Integration

Use this path when another project needs to call BrowseForge through REST, MCP HTTP, workflows, or the Playwright proxy.

## Start a Local Runtime

```bash
BrowseForge init
BrowseForge browsers install
BrowseForge serve --no-open
```

## Discover Runtime Values

```bash
TOKEN=$(BrowseForge token)
BrowseForge status --json
BrowseForge capabilities --json
```

REST base URL:

```text
http://127.0.0.1:19280/api
```

MCP HTTP URL:

```text
http://127.0.0.1:19280/mcp
```

Authorization header:

```http
Authorization: Bearer <token>
```

## Smoke Checks

```bash
BrowseForge smoke rest --wait --json
BrowseForge smoke mcp --token "$TOKEN" --wait --json
```

## Common API Checks

```bash
curl http://127.0.0.1:19280/api/status

curl http://127.0.0.1:19280/api/profiles \
  -H "Authorization: Bearer $TOKEN"

curl http://127.0.0.1:19280/api/playwright/endpoint \
  -H "Authorization: Bearer $TOKEN"
```

## Workflows

```bash
BrowseForge workflow run examples/multi-login.yaml --token "$TOKEN" --json
```

Use workflows for repeatable procedures and MCP page tools for adaptive agent behavior.
