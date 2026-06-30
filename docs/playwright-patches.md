# Playwright Driver Patch Status

[繁體中文](playwright-patches.zh-TW.md)

## Current Status

BrowseForge now uses the upstream `github.com/playwright-community/playwright-go` `v0.6000.0` Playwright 1.60 integration and no longer keeps the local hotfix that was previously required for the Playwright 1.59.1 driver.

| Patch | Old Location | Status | Reason Removed |
|-------|--------------|--------|----------------|
| Missing `/` in the WebSocket Bind path | `internal/browser/manager.go` via `patchDriverWSBind()` / `fixWSEndpoint()` | Removed | Playwright driver 1.60 produces the correct `/guid` WebSocket path from `browser.bind()` |

## Background

The Playwright 1.59.1 driver generated a WebSocket path without the leading slash:

```typescript
// 1.59.1 buggy behavior:
endpoint = await this._wsServer.listen(options.port ?? 0, options.host, (0, import_utils.createGuid)());

// Fixed behavior:
endpoint = await this._wsServer.listen(options.port ?? 0, options.host, '/' + createGuid());
```

The missing slash caused two failures:

1. The endpoint could look like `ws://host:PORTguid`, with the port and GUID joined together.
2. WebSocket upgrade handling failed because the HTTP pathname included `/` while the server path did not.

## Previous BrowseForge Hotfix

BrowseForge previously carried two local safeguards:

1. `patchDriverWSBind()` patched the cached driver JavaScript file at startup.
2. `fixWSEndpoint()` repaired malformed endpoints as a fallback.

Both were specific to Playwright 1.59.1. Keeping them after the 1.60 migration would hide the actual runtime behavior and keep BrowseForge coupled to an obsolete driver cache path.

## Verification

1. Check `go.mod` for the current `playwright-go` version.
2. Inspect the cached driver file when needed:

   ```bash
   grep "wsServer.listen" $(find ~/.cache ~/Library/Caches -path "*/ms-playwright-go/*/package/lib/server/browser.js" 2>/dev/null)
   ```

3. Playwright 1.60 should include `"/" +` or equivalent slash-prefixed path generation.
4. BrowseForge should not contain `patchDriverWSBind()` or `fixWSEndpoint()`.
5. Run the Camoufox Bind spike before release:

   ```bash
   CAMOUFOX_PATH=/path/to/camoufox go test -count=1 -run '^TestPlaywrightBindEndpointWithCamoufox$' -v ./internal/spike
   ```

## Related Files

- `internal/browser/manager.go` uses the endpoint returned by `browser.Bind()`.
- `go.mod` points to the upstream `github.com/playwright-community/playwright-go` `v0.6000.0` release.
- `internal/spike/bind_test.go` covers the Bind endpoint runtime path.

## Upstream Tracking

- Playwright upstream: https://github.com/microsoft/playwright
- Playwright Go upstream: https://github.com/mxschmitt/playwright-go
- Playwright 1.60 includes the WebSocket Bind path fix.
