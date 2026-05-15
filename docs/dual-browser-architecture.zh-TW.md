# 雙瀏覽器反偵測架構

BrowseForge 目前的核心產品契約是雙瀏覽器支援。早期 Camoufox-first 的研究仍有參考價值，但現行 runtime 是同時支援 Firefox/Camoufox 與 Chromium/CloakBrowser profile 的反偵測工作區。

## 現行 Runtime 契約

每個 profile 都有 `engine`：

- `firefox` 啟動 Camoufox。
- `chromium` 啟動 CloakBrowser。
- 空值或舊 profile 會預設為 `firefox`，維持升級相容。

REST API、MCP server、dashboard、workflow runner、session store、backup，以及 Playwright 連線模型都共用同一個 profile 契約。引擎差異集中在 browser manager 與 browser download layer。

## Firefox/Camoufox 路徑

Camoufox 是 Firefox-family runtime。BrowseForge 會為每個 profile 啟動獨立 browser process，使用 profile 專屬 data directory，並透過原生 `CAMOU_CONFIG` 套用指紋。

Camoufox 會接收 BrowseForge 指紋池中的完整 fingerprint object。內容可包含 navigator、screen、window、font、canvas seed、timezone、WebRTC，以及完整 WebGL 欄位。WebGL 採用 all-or-nothing：如果 BrowseForge 沒有完整 WebGL profile，會移除片段 renderer/vendor 欄位，讓 Camoufox 自行產生一致的原生結果。

這條路徑適合需要明確、可檢視 fingerprint 欄位，以及 Camoufox Firefox 層原生 masking 行為的 profile。

## Chromium/CloakBrowser 路徑

CloakBrowser 是 Chromium-family runtime。BrowseForge 會為每個 profile 啟動獨立 browser process，使用 profile 專屬 user data directory 與 CloakBrowser 原生 flags。

CloakBrowser 主要以 `fingerprint_seed` 作為身份來源。seed 會驅動 Chromium runtime 內部的原生 fingerprint surface。當 profile 有明確欄位時，BrowseForge 也可能傳入 timezone、locale、WebRTC、platform、fonts 等 selected override。

新的 Chromium profile 若沒有提供 seed，會自動產生 seed。這和 Firefox 路徑刻意不同：Chromium 的主要契約是 seed，明確欄位則是 override。

## Playwright 控制

兩個 engine 都透過專案整合的 Playwright 1.60 Bind endpoint 暴露 Playwright-compatible 控制。外部 client 連線到執行中的 profile 時，應使用 Playwright 1.60.x。

目前 release gate 已包含 Camoufox Bind runtime spike。BrowseForge 也保留 opt-in CloakBrowser Bind spike（`CLOAKBROWSER_SPIKE=1`），可在提供已驗證 CloakBrowser binary 的機器上執行。兩個測試都會啟動 persistent profile、建立 Playwright Bind endpoint、用第二個 Playwright client 連回去，並透過該 endpoint 開頁。

## 歷史資料

部分舊文件描述 Camoufox-only 或 per-container setter 設計。這些是規劃 archive，不是現行產品契約。現行實作不依賴單一共用 browser 搭配 per-container spoofing，而是每個 profile 一個隔離 runtime process；這比較容易推理，也能同時符合兩個 browser family。

除非明確更新，以下資料應視為歷史背景：

- `docs/plan.md`
- `docs/execution.md`
- `docs/wbs.md`
- `docs/phase2-fork-plan.md`
- `docs/spike-results.md`
- `extension/lib/fingerprint-injector.js`

## 目前缺口

- 提供已驗證 CloakBrowser binary 的 release machine 應使用 `REQUIRE_CLOAKBROWSER=1` 執行 release preflight。
- Release-critical workflow 維持 English-first public docs，並提供繁體中文對照。
- 持續替換不屬於相容性 persisted data 的舊內部命名，例如 `camoufoxmulti` 與 `cmfx`。
- 若未來從 CloakBrowser 改成自維護 Chromium fork，應文件化為 engine provider 變更，而不是 API contract 變更。
