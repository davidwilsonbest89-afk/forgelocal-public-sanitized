# Agent Integration

BrowseForge supports local MCP stdio and remote MCP HTTP. Use stdio when the agent runs on the same machine. Use HTTP when BrowseForge runs on a server or container.

## Local Stdio

```bash
BrowseForge init
BrowseForge browsers install
BrowseForge mcp-config stdio --json
```

Give the generated config to the agent client.

## Remote HTTP

Start BrowseForge and keep the token persisted in `data/.api-token`.

For GHCR/container deployments, set `BROWSEFORGE_PUBLIC_BASE_URL` before starting BrowseForge so MCP `screenshot` returns agent-fetchable `screenshot_url` links:

```bash
export BROWSEFORGE_PUBLIC_BASE_URL=${BROWSEFORGE_PUBLIC_BASE_URL:-http://localhost:19280}
```

```bash
TOKEN=$(BrowseForge token)
BrowseForge smoke mcp --token "$TOKEN" --wait --json
BrowseForge mcp-config http --url http://YOUR_SERVER:19280/mcp --token "$TOKEN" --json
```

MCP `screenshot_url` links are temporary, unauthenticated, random URLs. Fetch them before `expires_at`; no Bearer header is required for the screenshot download URL itself.

Do not expose `19280` publicly without a trusted network boundary, SSH tunnel, VPN, or hardened HTTPS reverse proxy.

## Agent Readiness Flow

1. `BrowseForge capabilities --json`
2. `BrowseForge doctor --strict --json`
3. `BrowseForge status --json`
4. `BrowseForge smoke rest --wait --json`
5. `BrowseForge smoke mcp --wait --json`

Agents should observe page state after browser actions, wait on conditions instead of fixed sleeps, and clean up temporary web sessions after search/explore workflows. See [Agent Prompt Guide](agent-prompt-guide.md).
