# Local Quickstart

Use this path when BrowseForge runs on your workstation and you want manual dashboard access, local agent control, or development integration.

## Manual Use

```bash
BrowseForge init
BrowseForge browsers install
BrowseForge serve
```

In another terminal:

```bash
BrowseForge status
BrowseForge open
```

`open` reads `data/.api-token` and opens the dashboard with the token in the URL fragment.

## Local Agent Use

```bash
BrowseForge doctor --strict --json
BrowseForge mcp-config stdio --json
```

Use the generated MCP config in the agent client. Stdio mode keeps traffic local to the machine.

## Local Development

```bash
BrowseForge serve --no-open
TOKEN=$(BrowseForge token)
BrowseForge smoke rest --wait --json
BrowseForge smoke mcp --token "$TOKEN" --wait --json
BrowseForge status --json
```

Use REST at `http://127.0.0.1:19280/api`, MCP HTTP at `http://127.0.0.1:19280/mcp`, and the Playwright proxy endpoints from the REST API.
