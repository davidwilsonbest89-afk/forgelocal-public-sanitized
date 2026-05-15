# Playwright Driver Patch 狀態

[English](playwright-patches.md)

## 當前狀態

目前 BrowseForge 使用 Playwright 1.60 整合版，已移除先前為 Playwright 1.59.1 driver 準備的本地 hotfix。

| Patch | 舊位置 | 狀態 | 移除原因 |
|-------|--------|------|----------|
| WebSocket Bind path 缺少 `/` | `internal/browser/manager.go` → `patchDriverWSBind()` / `fixWSEndpoint()` | 已移除 | Playwright driver 1.60 已在 `browser.bind()` 產生正確 `/guid` WebSocket path |

## 詳細說明

### Bug 描述

`packages/playwright-core/src/server/browser.ts` 中：

```typescript
// 1.59.1 (有 bug):
endpoint = await this._wsServer.listen(options.port ?? 0, options.host, (0, import_utils.createGuid)());

// main 分支 (已修正):
endpoint = await this._wsServer.listen(options.port ?? 0, options.host, '/' + createGuid());
```

缺少 `/` 導致：
1. endpoint 回傳 `ws://host:PORTguid`（port 和 guid 黏在一起）
2. WebSocket upgrade handler 比對 `pathname !== path` 永遠失敗（因為 HTTP pathname 有 `/` 但 server 的 path 沒有）

### BrowseForge 曾經的修正

1. **`patchDriverWSBind()`** — 啟動時自動修正 driver JS 檔案（`{cacheDir}/ms-playwright-go/1.59.1/package/lib/server/browser.js`）
2. **`fixWSEndpoint()`** — 作為 fallback，修正回傳的 endpoint 格式（從尾部取 32 字元 GUID 插入 `/`）

這兩個修正只適用於 Playwright 1.59.1。升級到 1.60 後繼續保留反而會讓 BrowseForge 依賴過期 driver cache 路徑，並模糊真正的 runtime 行為。

### 驗證步驟

1. 檢查 `go.mod` 中 playwright-go 的版本和對應的 driver 版本
2. 查看 driver 檔案：
   ```bash
   grep "wsServer.listen" $(find ~/.cache ~/Library/Caches -path "*/ms-playwright-go/*/package/lib/server/browser.js" 2>/dev/null)
   ```
3. Playwright 1.60 driver 應包含 `"/" +` 或等價的 `/` path 產生邏輯。
4. BrowseForge 不應再出現 `patchDriverWSBind()` 或 `fixWSEndpoint()`。

### 相關檔案

- `internal/browser/manager.go` — 直接使用 `browser.Bind()` 回傳的 endpoint
- `go.mod` — playwright-go 版本（目前 fork: `nczz/playwright-go` 1.60 整合版，driver: 1.60.0）

### 上游追蹤

- Playwright 主倉庫：https://github.com/microsoft/playwright
- Playwright 1.60 已包含 WebSocket Bind path 修正
