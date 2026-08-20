# BrowseForge Architecture Guide

This guide is the code-orientation entry point for agents and maintainers. It describes the current implementation boundaries so changes land in the existing framework instead of creating parallel paths.

## Runtime Shape

BrowseForge is a Go control server around isolated browser profiles:

1. `cmd/server` wires configuration, storage, browser management, REST API, MCP, and static UI.
2. `internal/profile` owns persisted profile metadata and profile directories.
3. `internal/browser` owns browser binaries, persistent browser sessions, Playwright Bind endpoints, and proxy relays.
4. `internal/api` exposes REST routes backed by profile storage and browser sessions.
5. `internal/mcp` exposes agent tools over stdio or Streamable HTTP and reuses the same profile/browser infrastructure.

Do not add a second browser/session framework for a new feature. New browser behavior should usually go through `internal/browser.Manager`, and new agent behavior should usually go through existing MCP tool handlers plus `SessionPool` when a page-level agent session is needed.

## Code Map

| Area | Primary files | Ownership |
| --- | --- | --- |
| Server composition | `cmd/server/main.go` | Process startup, ports, API/MCP wiring, background cleanup |
| Config | `internal/config/config.go` | Runtime config shape and defaults |
| Profiles | `internal/profile/store.go` | Profile CRUD, persistence, profile directory layout |
| Browser session lifecycle | `internal/browser/manager.go` | Public session API, Playwright driver lifecycle, Bind endpoint setup |
| Chromium launch | `internal/browser/launch_chromium.go` | CloakBrowser path, flags, downloads, proxy setup |
| Firefox launch | `internal/browser/launch_firefox.go` | Camoufox config, downloads, proxy setup |
| Launch helpers | `internal/browser/launch_helpers.go` | Shared launch error handling and stale lock cleanup |
| Browser downloads | `internal/browser/download.go` | Camoufox/CloakBrowser install and version expectations |
| REST routes | `internal/api/router.go`, `internal/api/sessions.go` | HTTP API routing and browser page operations |
| MCP transport/server | `internal/mcp/server.go`, `internal/mcp/protocol.go` | JSON-RPC handling, auth, request dispatch |
| MCP tool schemas | `internal/mcp/tool_schema.go` | Tool list and JSON schema definitions |
| MCP browser tools | `internal/mcp/web_tools.go`, `internal/mcp/advanced_tools.go` | Tool argument parsing, page utilities, cookies, downloads, workflow, diagnostics |
| MCP web session lifecycle | `internal/mcp/web_session_pool.go` | Per-agent page sessions, profile pool connection, GC |
| MCP search/explore behavior | `internal/mcp/web_search.go`, `internal/mcp/search_provider.go` | Search providers, navigation, extraction, fallback parsing |
| Humanized input | `internal/humanize` | Mouse/keyboard behavior wrappers |
| Workflow execution | `internal/workflow` | YAML workflow execution |
| Static UI | `web` | Dashboard assets |

## Change Rules

- Add a new MCP tool in `internal/mcp/tool_schema.go`, dispatch it from `internal/mcp/server.go`, and place tool-specific argument/result code in a focused file near related tools. Browser-page utility tools belong in `internal/mcp/advanced_tools.go`.
- Add browser launch behavior in `launch_chromium.go` or `launch_firefox.go`; keep `manager.go` focused on public session lifecycle.
- Add page-level agent behavior to `WebSession` methods in `web_search.go` or a sibling file. Use `SessionPool` for lifecycle and never close `browser.Manager` owned profile browsers from MCP cleanup paths.
- Add REST page operations in `internal/api/sessions.go` only when the behavior belongs to the public HTTP API. Do not call MCP handlers from REST or REST handlers from MCP.
- Add profile metadata through `internal/profile` first, then update API, UI, backup/restore, and docs as needed.
- Prefer Playwright APIs already used in the repo. Avoid CDP-only implementations unless the feature is explicitly Chromium-only and documented as such.

## Verification Gates

Use the narrowest fast check while iterating, then run the full gate before handoff:

```bash
go test ./internal/mcp
go test ./internal/browser
go test ./...
go vet ./...
go build ./...
git diff --check
```

For release-sensitive browser/runtime work, also use `scripts/release-preflight.sh`. For MCP web-session changes, use `scripts/mcp-web-smoke.sh` when a real Chromium/CloakBrowser profile is available.

## Related References

- [Session Architecture Design](session-arch-design.md)
- [Dual-Browser Anti-Detection Architecture](dual-browser-architecture.md)
- [Playwright Patch Status](playwright-patches.md)
