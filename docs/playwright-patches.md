# Playwright Driver Patch 狀態

## 當前 Patch

| Patch | 位置 | 原因 | 移除條件 |
|-------|------|------|---------|
| WebSocket Bind path 缺少 `/` | `internal/browser/manager.go` → `patchDriverWSBind()` | Playwright 1.59.1 driver 的 `browser.js` 中 `startServer` 呼叫 `wsServer.listen()` 時 path 沒有 `/` 前綴，導致 WebSocket upgrade 永遠 400 | Playwright driver ≥ 1.60（main 分支已修正） |

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

### BrowseForge 的修正

1. **`patchDriverWSBind()`** — 啟動時自動修正 driver JS 檔案（`{cacheDir}/ms-playwright-go/1.59.1/package/lib/server/browser.js`）
2. **`fixWSEndpoint()`** — 作為 fallback，修正回傳的 endpoint 格式（從尾部取 32 字元 GUID 插入 `/`）

### 驗證步驟（確認是否可以移除 patch）

1. 檢查 `go.mod` 中 playwright-go 的版本和對應的 driver 版本
2. 查看 driver 檔案：
   ```bash
   grep "wsServer.listen" $(find ~/.cache ~/Library/Caches -path "*/ms-playwright-go/*/package/lib/server/browser.js" 2>/dev/null)
   ```
3. 如果輸出包含 `"/" +`，表示已修正，可以移除：
   - `patchDriverWSBind()` 函數
   - `fixWSEndpoint()` 函數
   - `NewManager()` 中的 `patchDriverWSBind()` 呼叫

### 相關檔案

- `internal/browser/manager.go` — patch 和 fix 函數
- `go.mod` — playwright-go 版本（目前 fork: `nczz/playwright-go v0.5700.2`，driver: 1.59.1）

### 上游追蹤

- Playwright 主倉庫：https://github.com/microsoft/playwright
- 修正 commit 在 main 分支，尚未 release
- 預計 Playwright 1.60 會包含此修正
