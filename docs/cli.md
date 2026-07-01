# BrowseForge CLI Reference

BrowseForge ships as a single binary. The CLI is the stable entry point for local operators, containers, CI jobs, and agents that need to discover or validate a BrowseForge runtime before using REST, MCP, or Playwright.

## Global Flags

| Flag | Description |
|------|-------------|
| `--base-dir DIR` | Runtime directory for `config.json`, `profiles/`, `data/`, `logs/`, and downloaded browser engines. Defaults to the binary directory. |
| `--config PATH` | Config path. Relative paths are resolved from `--base-dir`. Defaults to `config.json`. |
| `--help`, `-h` | Print usage and exit `0`. Works at the root and after subcommands. |
| `--version` | Print the BrowseForge version and exit `0`. |

## Exit Codes

| Code | Meaning |
|------|---------|
| `0` | Command succeeded. |
| `1` | Runtime validation, IO, server, smoke, or workflow execution failed. |
| `2` | CLI usage error, unknown command, invalid flags, or invalid positional arguments. |

Unknown commands fail with exit code `2` and print usage. They do not start the server.

## Commands

### `serve`

Starts the Dashboard, REST API, MCP HTTP endpoint, and Playwright proxy.

```bash
BrowseForge serve
BrowseForge serve --host 0.0.0.0 --port 19280 --no-sandbox --no-open
```

| Flag | Description |
|------|-------------|
| `--host HOST` | Override the config listen address. Docker auto-detects `0.0.0.0` when the config uses `127.0.0.1`. |
| `--port PORT` | Override the REST/Dashboard/MCP HTTP port. |
| `--no-sandbox` | Disable Chromium sandbox. Required in many Docker runtimes. |
| `--no-open` | Do not open the local dashboard browser after startup. Recommended for agents, CI, and remote servers. |

Running `BrowseForge` with no subcommand is kept as a compatibility alias for `BrowseForge serve`.

### `mcp-stdio`

Starts the MCP server over stdio for local agent clients.

```bash
BrowseForge mcp-stdio
BrowseForge --mcp
```

`--mcp` is kept for compatibility. Stdio MCP reuses the existing config and does not download browsers during startup. If `config.json` is missing, it creates the default config.

### `init`

Creates the runtime directories and a default config.

```bash
BrowseForge init
BrowseForge init --force --json
BrowseForge --base-dir /srv/browseforge init
```

| Flag | Description |
|------|-------------|
| `--force` | Overwrite an existing config file. |
| `--json` | Print a machine-readable result. |

JSON shape:

```json
{
  "ok": true,
  "base_dir": "/srv/browseforge",
  "config": "/srv/browseforge/config.json"
}
```

### `config show`

Prints the effective config loaded from `--config`.

```bash
BrowseForge config show --json
```

### `config validate`

Validates required config fields.

```bash
BrowseForge config validate
BrowseForge config validate --json
```

JSON shape:

```json
{
  "ok": true,
  "config": "/path/to/config.json"
}
```

### `token`

Prints the REST and MCP Bearer token from `data/.api-token`.

```bash
BrowseForge token
BrowseForge token --json
```

JSON shape:

```json
{
  "ok": true,
  "token": "redacted-example",
  "path": "/srv/browseforge/data/.api-token"
}
```

Treat this output as sensitive. The token is created on first server start and should be persisted by mounting or backing up `data/`.

### `doctor`

Checks local runtime readiness.

```bash
BrowseForge doctor
BrowseForge doctor --strict --json
```

| Flag | Description |
|------|-------------|
| `--strict` | Treat missing display, browser engines, token, and Docker sandbox requirements as failures instead of warnings. |
| `--json` | Print a machine-readable report. |

JSON shape:

```json
{
  "version": "v1.8.1",
  "base_dir": "/srv/browseforge",
  "checks": [
    {"name": "config", "status": "ok", "message": "/srv/browseforge/config.json"}
  ],
  "ok": true
}
```

Check statuses are `ok`, `warn`, or `fail`.

### `capabilities`

Prints the integration surface supported by the binary.

```bash
BrowseForge capabilities
BrowseForge capabilities --json
```

Use this in agents and installers before deciding whether to use REST, MCP HTTP, MCP stdio, or Playwright proxy.

### `smoke rest`

Checks the REST status endpoint.

```bash
BrowseForge smoke rest --base-url http://127.0.0.1:19280 --json
```

If `--base-url` is omitted, the CLI uses the host and port from the config. `0.0.0.0` is converted to `127.0.0.1` for local checks.

### `smoke mcp`

Checks the MCP HTTP endpoint by sending an `initialize` request.

```bash
TOKEN=$(BrowseForge token)
BrowseForge smoke mcp --base-url http://127.0.0.1:19280 --token "$TOKEN" --json
```

If `--token` is omitted, the CLI reads `data/.api-token`.

### `workflow run FILE`

Runs a YAML workflow through the server-backed workflow engine.

```bash
TOKEN=$(BrowseForge token)
BrowseForge workflow run examples/multi-login.yaml --token "$TOKEN" --json
```

The server must already be running. If `--base-url` is omitted, the CLI uses the config host and port.

### `profiles list`

Lists profiles through the REST API.

```bash
BrowseForge profiles list --token "$TOKEN" --json
```

### `sessions list`

Lists active sessions through the REST API.

```bash
BrowseForge sessions list --token "$TOKEN" --json
```

## Agent Integration Checklist

1. Run `BrowseForge --help` or `BrowseForge capabilities --json` to confirm the binary supports the expected integration surface.
2. Run `BrowseForge init` for a new runtime directory.
3. Start the server with `BrowseForge serve --no-open`.
4. Read the token with `BrowseForge token --json`.
5. Validate REST with `BrowseForge smoke rest --json`.
6. Validate MCP HTTP with `BrowseForge smoke mcp --token "$TOKEN" --json`.
7. Use `doctor --strict --json` in CI or deployment health checks.

## Data Persistence

For production and containers, persist at least:

- `profiles/`: profile metadata and browser user data.
- `data/`: API token and fingerprint data.
- `browsers/`: downloaded browser engines.
- `logs/`: server logs, useful for diagnosis.

When using Docker, prefer host bind mounts for these directories if you want normal filesystem backups. See [Linux Server Deployment](linux-server.md).
