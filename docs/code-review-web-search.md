# Strict Code Review: MCP `web_search` / `web_explore`

Review scope: current uncommitted BrowseForge changes for MCP-backed web search/explore sessions.

## Executive Summary

The current architecture is now aligned with BrowseForge's anti-detect browser coordinator role:

- Chrome/CloakBrowser profiles are primary for MCP web sessions.
- `browser.Manager` owns the persistent profile browser and Bind endpoint.
- `mcp.SessionPool` owns only agent session metadata and independent pages.
- Each agent session opens one isolated page with `connectedBrowser.NewPage()`.
- GC/destroy/shutdown close agent pages only and do not close profile browsers.

No Critical or High severity blockers remain after the cleanup edits in this branch.

## Severity Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | 0 | None found after cleanup |
| High | 0 | None found after cleanup |
| Medium | 3 | 3 completed, live HTTP MCP smoke still environment-dependent |
| Low | 3 | 2 completed/documented, 1 future enhancement |

## Findings

### Critical

None.

The previous lifecycle concern—agent session cleanup potentially closing a `browser.Manager`-owned browser through a connected browser handle—has been fixed. `CloseAll` and profile-pool refresh no longer call `Browser.Close()` on the connected profile browser handle; session lifecycle code closes only pages and metadata.

### High

None.

The older `Searcher`/Camoufox mismatch is no longer applicable. The current implementation uses `SessionPool`, `profile.Store`, `browser.Manager`, and Chromium/CloakBrowser profile sessions through Playwright Bind.

### Medium

#### M1 — Runtime behavior still needs real browser MCP smoke coverage

**Status:** Completed for current-code stdio MCP path. A `dev-smoke` Linux binary was built from the current worktree, copied into the existing `browseforge` container, and run with `/app/BrowseForge.current --mcp` against the container's real `DISPLAY=:1`, CloakBrowser install, Playwright runtime, config, and Chromium profile `prof_dc1530a98b16`.

**Observed result:** `initialize`/`tools/list`, `create_session`, `web_search`, `web_explore`, and `destroy_session` all passed. `web_search` launched the Chromium profile browser via Bind and returned three structured Google results; `web_explore` extracted the `Example Domain` page; `destroy_session` closed the returned agent page session.

**HTTP MCP limitation:** In-place replacement of the long-running container service could not be used for current-code HTTP MCP smoke because the existing container restores `/app/BrowseForge` v1.7.7. MCP is now intended to be mounted at `/mcp` on the main BrowseForge service port, so the gated HTTP smoke script should be run against an upgraded service once deployed:

```bash
RUN_MCP_SMOKE=1 MCP_PROFILE_ID=<profile-id> ./scripts/mcp-web-smoke.sh
```

Confirm the profile browser remains alive until explicitly closed by browser/session APIs.

#### M2 — Search extraction remains Google-DOM dependent

**Status:** Completed for current scope. `extractGoogleResults` now detects Google consent/captcha/unusual-traffic interstitials and returns explicit errors. It also includes a fallback organic anchor/heading extraction path when primary card selectors return no results.

**Residual risk:** Google markup and anti-bot behavior can still change; the new explicit errors and fallback reduce silent empty results but do not remove external SERP fragility.

#### M3 — `LastAccessed` is updated after successful operations only

**Status:** Completed. `WebSearch` and `WebExplore` now refresh `LastAccessed` before navigation so failed active attempts are not GC-stale solely because success did not occur.

### Low

#### L1 — Connected browser handles are retained but not explicitly disconnected

**Status:** Documented. `SessionPool` comments now explain that connected browser handles are non-owning, `Browser.Close()` may close the `browser.Manager`-owned profile browser, and BrowseForge has no verified non-closing disconnect path yet. Session cleanup closes pages only.

**Future enhancement:** If Playwright Go exposes a verified non-closing disconnect API for connected remote browsers, use it.

#### L2 — API response content is text-wrapped JSON rather than a pure structured result

**Status:** Resolved by compatibility rationale. `API.md` now documents that MCP `content` text blocks are retained for client compatibility while session metadata is exposed as top-level `session_id`, `profile_id`, and `session_created` fields.

**Future enhancement:** Consider adding explicit top-level structured result fields such as `results` or `page` in addition to the text content.

#### L3 — Documentation intentionally mentions old terms only as negatives

**Impact:** `docs/session-arch-design.md` mentions `BrowserContext`, `Searcher`, and Camoufox only in non-goal/review-checklist context. This is acceptable, but grep-based stale-design checks must distinguish negative context from active design.

**Recommendation:** Keep these mentions only where they clarify non-goals; avoid reintroducing old active flow terms such as `NewContext`, `Context.Close`, `agentBrowser`, or `agentPW`.

## Architecture Assessment

The architecture is reasonable and appropriate for BrowseForge:

```text
Agent MCP call
  -> MCP tool handler
  -> SessionPool
  -> browser.Manager profile browser
  -> Playwright Bind endpoint
  -> connectedBrowser.NewPage()
  -> independent per-agent Page
```

This design preserves the anti-detect profile browser as the source of truth while giving each agent session an independent page. It avoids the previous shared-page race problem and avoids starting a separate non-profile browser for search.

## Documentation Alignment

Updated docs now align with implementation:

- `API.md` documents session tools, parameters, response fields, and error codes.
- `docs/session-arch-design.md` documents page-only Bind sessions and ownership boundaries.
- `docs/mcp-call-flow.html` shows `connectedBrowser.NewPage()` and `WebSession.Page` cleanup.
- This review replaces the obsolete `Searcher`/Camoufox review.

## Validation Completed

- `gofmt -w cmd/server/main.go internal/browser/manager.go internal/mcp/server.go internal/mcp/web_search.go`
- `go test ./...`
- `go build ./...`
- `GO_IMAGE=golang:1.26-alpine ./scripts/docker-go-validate.sh`
- Current-code container MCP stdio smoke using `/app/BrowseForge.current --mcp` inside the existing `browseforge` container:
  - `initialize` / `tools/list` returned version `dev-smoke` and the web session tools.
  - `create_session` succeeded for Chromium profile `prof_dc1530a98b16`.
  - `web_search` returned three structured Google results.
  - `web_explore` extracted `https://example.com`.
  - `destroy_session` succeeded for the returned session.

Docker validation runs normal package tests and `go build ./...`; it compile-checks `internal/spike` without executing browser-dependent spike tests because those require host Playwright/browser assets.

## Remaining Manual External Verification

1. After deploying the current build with MCP mounted at `/mcp` on the main service port, run the gated HTTP MCP smoke against the upgraded live service:
   ```bash
   RUN_MCP_SMOKE=1 MCP_PROFILE_ID=<profile-id> ./scripts/mcp-web-smoke.sh
   ```
2. Confirm the profile browser remains alive after `destroy_session` and is only closed by browser/session APIs.
3. Optional future enhancement: add top-level structured `results` / `page` payloads while preserving MCP `content` compatibility.
