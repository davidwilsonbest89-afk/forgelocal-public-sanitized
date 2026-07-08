# BrowseForge — 執行計劃

> Archive note: this execution plan reflects early Camoufox-first milestones. The current product contract is the dual-browser architecture in [dual-browser-architecture.md](dual-browser-architecture.md).

> 目標：編譯出一個執行檔，打開就能用。
> 分工：AI 執行開發 + 排錯 + 驗證，使用者做最終產品驗證。
> 每個 Phase 結束時有明確的「使用者驗證項目」。

---

## Phase 0：Spike 驗證（2-3 天）

### 0-1. 環境準備
- [ ] 安裝 Go 1.22+
- [ ] 安裝 Node.js 22 LTS
- [ ] 安裝 web-ext (`npm install -g web-ext`)
- [ ] 下載 Camoufox binary（FF146-pre 優先，v135 備用）
- [ ] 確認 Camoufox 可啟動：`./camoufox --version`

### 0-2. Spike：Container 隔離
- [ ] 在 Camoufox 中啟用 Container Tabs（about:config → privacy.userContext.enabled = true）
- [ ] 手動建立 2 個 Container，各自登入不同網站帳號
- [ ] 驗證 cookie 互不干擾
- [ ] 關閉重開，驗證 cookie 保留
- [ ] 記錄結果到 spike-results.md

### 0-3. Spike：proxy.onRequest + cookieStoreId
- [ ] 寫最小 WebExtension（manifest.json + background.js）
- [ ] proxy.onRequest listener 印出 requestInfo.cookieStoreId
- [ ] 驗證不同 container 的 cookieStoreId 不同
- [ ] 根據 cookieStoreId 回傳不同 proxy → httpbin.org/ip 驗證
- [ ] 記錄結果

### 0-4. Spike：Camoufox Portable 模式
- [ ] `./camoufox -profile ./test-data -no-remote` 測試
- [ ] 確認資料寫入 ./test-data/
- [ ] 搬移目錄 → 重啟 → 確認資料保留
- [ ] 測試未簽署 extension 載入（xpinstall.signatures.required = false）
- [ ] 記錄結果

### 0-5. Spike：Playwright 連接
- [ ] 安裝 playwright-go（先試 v0.5700.1）
- [ ] 寫 Go 測試程式：Launch Camoufox with ExecutablePath + Env
- [ ] 確認能 navigate + screenshot
- [ ] 若失敗，退回 playwright-go v0.5101.0 重試
- [ ] 記錄可用的版本組合

### 0-6. Spike：Content Script → Setter
- [ ] 寫測試 extension，content script (document_start) 呼叫 wrappedJSObject.setCanvasSeed()
- [ ] 在 Camoufox 中載入 → 訪問測試頁面
- [ ] 驗證 Canvas toDataURL hash 因 seed 不同而改變
- [ ] 記錄結果

### 0-7. Spike：Container 數量 + 記憶體
- [ ] 腳本批次建立 10/25/50 個 container + tab
- [ ] 記錄每階段記憶體佔用
- [ ] 確定產品建議帳號上限

### 0-8. Spike 總結
- [ ] 完成 spike-results.md
- [ ] 所有 PASS → 進入 Phase 1
- [ ] 有 FAIL → 啟動對應 fallback 方案

**🔍 使用者驗證：** 閱讀 spike-results.md，確認結果可接受。

---

## Phase 1：Firefox MVP（Week 1-5）

### Week 1：骨架

#### 1-1. Go 專案初始化
- [ ] `go mod init camoufoxmulti`
- [ ] 目錄結構：cmd/server/、internal/api/、internal/profile/、internal/session/、internal/browser/
- [ ] 安裝依賴：playwright-go、chi router、slog
- [ ] 建立 config.json 結構 + 載入邏輯

#### 1-2. Profile 系統（Go）
- [ ] Profile JSON 結構（v2 `runtime_id` provider 欄位；legacy `engine` 需先遷移）
- [ ] Profile CRUD：Create / Read / Update / Delete / Duplicate
- [ ] 檔案系統持久化（atomic write）
- [ ] 啟動時掃描 profiles/ 載入記憶體
- [ ] 單元測試

#### 1-3. Playwright 啟動 Camoufox
- [ ] Go server 啟動時用 playwright-go Launch Camoufox
- [ ] 傳入 CAMOU_CONFIG env var（基底指紋）
- [ ] 傳入 firefox_user_prefs（Container 啟用、WebRTC 關閉等）
- [ ] 持有 browser handle
- [ ] 關閉流程：browser disconnected → server 清理退出

#### 1-4. 啟動腳本
- [ ] start.sh（macOS/Linux）
- [ ] start.bat（Windows）
- [ ] 首次啟動初始化（建立目錄、預設 config）

#### 1-5. 日誌框架
- [ ] slog JSON 日誌
- [ ] HTTP request logging 中間件
- [ ] 日誌檔案輸出（logs/server.log）

**🔍 M1 使用者驗證：**
- `./start.sh` 啟動 → Camoufox 視窗出現
- `curl http://localhost:19280/api/status` 回傳 200
- `curl -X POST http://localhost:19280/api/profiles -d '{"name":"test"}'` 建立 profile 成功

---

### Week 2：Container + Proxy + 指紋

#### 2-1. WebExtension 骨架
- [ ] manifest.json（MV2，permissions: contextualIdentities, proxy, tabs, storage, cookies, sidebar_action）
- [ ] background.js（啟動時連線 Control Server WebSocket）
- [ ] sidebar/index.html + style.css + app.js（空殼）

#### 2-2. Container 管理
- [ ] 建立 profile → 建立 contextualIdentity → 記錄 cookieStoreId 映射
- [ ] 刪除 profile → browsingData.removeCookies → 刪除 contextualIdentity
- [ ] 開啟 profile → tabs.create({ cookieStoreId }) → 記錄 tab ID
- [ ] 關閉 profile → tabs.remove
- [ ] 啟動時一致性檢查（清理孤兒 container）

#### 2-3. Per-container Proxy 路由
- [ ] proxy.onRequest listener
- [ ] 從記憶體映射表查 cookieStoreId → proxy 設定
- [ ] proxyDNS: true
- [ ] webRequest.onAuthRequired 自動填入帳密
- [ ] 驗證：兩個 container 各自走不同 proxy → httpbin.org/ip 確認

#### 2-4. 指紋池生成
- [ ] Node.js 腳本呼叫 fingerprint-suite 批次生成
- [ ] Firefox + Chrome 各 500 組，按 OS 分類
- [ ] fingerprint-suite → Camoufox 格式轉換器
- [ ] 輸出 data/fingerprints-firefox-windows.json 等

#### 2-5. Per-container 指紋注入
- [ ] Content script (document_start)：讀 storage → wrappedJSObject.setXxx()
- [ ] Profile 建立時：指紋寫入 browser.storage.local（key: fp_{cookieStoreId}）
- [ ] GeoIP 地理調配（proxy IP → timezone/locale/language/geolocation）
- [ ] 驗證：兩個 container 訪問 browserleaks.com → Canvas hash 不同

#### 2-6. API 認證
- [ ] 啟動時生成隨機 token → 寫入 data/.api-token
- [ ] Authorization: Bearer {token} 中間件
- [ ] /api/status 免認證

**🔍 M2 使用者驗證：**
- Sidebar 顯示 profile 列表
- 點擊 profile → 開啟 Container Tab → 該 tab 走指定 proxy
- 登入網站 → 關閉 tab → 重開 → cookie 保留
- 兩個 profile 的 Canvas hash 不同

---

### Week 3：Control Server API + Sidebar

#### 3-1. Session 管理 API
- [ ] POST /api/sessions（開啟 profile 的 container tab）
- [ ] GET /api/sessions（列出活躍 session）
- [ ] DELETE /api/sessions/:id（關閉 tab）
- [ ] Session → Playwright page 映射

#### 3-2. 瀏覽器操作 API（Playwright）
- [ ] POST /api/sessions/:id/navigate → page.Goto()
- [ ] POST /api/sessions/:id/click → page.Click()
- [ ] POST /api/sessions/:id/type → page.Type()
- [ ] POST /api/sessions/:id/eval → page.Evaluate()
- [ ] GET /api/sessions/:id/screenshot → page.Screenshot()
- [ ] GET /api/sessions/:id/content → page.Content()
- [ ] POST /api/sessions/:id/wait → page.WaitForSelector()

#### 3-3. Cookie API
- [ ] GET /api/sessions/:id/cookies（透過 extension）
- [ ] POST /api/sessions/:id/cookies（透過 extension）

#### 3-4. Playwright endpoint 暴露
- [ ] GET /api/playwright/endpoint → ws_url

#### 3-5. Sidebar UI 完善
- [ ] Profile 列表（按分組摺疊、🟢/⚪ 狀態）
- [ ] 搜尋功能
- [ ] 點擊開啟/切換 tab
- [ ] 右鍵選單（編輯、複製、刪除）

#### 3-6. Tab suspend
- [ ] 非活躍 tab 自動 discard（15 分鐘）
- [ ] 手動 suspend（右鍵選單）
- [ ] 💤 狀態顯示

**🔍 M3 使用者驗證：**
- curl 完成完整流程：建立 profile → 開 session → navigate → screenshot → 關閉
- Sidebar 搜尋、分組、右鍵選單正常
- 無 token 的請求被拒絕

---

### Week 4：UI 完善 + 異常恢復

#### 4-1. Profile 操作介面
- [ ] 新增/編輯 Profile 表單（名稱、分組、proxy、指紋預覽）
- [ ] 指紋預覽面板（UA、Platform、Screen、Timezone、WebGL）
- [ ] 「測試指紋」按鈕 → 開啟 browserleaks.com
- [ ] 「重新生成指紋」按鈕

#### 4-2. 狀態監控
- [ ] 人工介入偵測（tab title 含「驗證」「CAPTCHA」→ 🔴 警示）
- [ ] Proxy 狀態顯示（🟢/🟡/🔴）
- [ ] 記憶體使用顯示

#### 4-3. Cookie 管理
- [ ] 匯出 container cookies 為 JSON
- [ ] 從 JSON 匯入 cookies

#### 4-4. 異常恢復
- [ ] 啟動時一致性修復（孤兒 container、殘留 session）
- [ ] WebSocket 斷線自動重連（指數退避）
- [ ] Profile JSON 損壞 fallback（.bak 備份）

#### 4-5. 備份/還原
- [ ] POST /api/profiles/:id/export → .zip
- [ ] POST /api/profiles/import → 匯入 .zip
- [ ] POST /api/backup → 全量備份
- [ ] POST /api/restore → 全量還原

### Week 5：打包 + 測試

#### 5-1. 打包腳本
- [ ] Makefile：download-camoufox、build-server、build-extension、package
- [ ] 三平台打包：BrowseForge-v0.1-{win64|macos-arm64|linux-x64}.zip
- [ ] Go binary strip（-ldflags "-s -w"）

#### 5-2. 測試
- [ ] Go 單元測試（profile CRUD、API handler、指紋生成）
- [ ] Extension lint（web-ext lint）
- [ ] E2E 測試腳本（curl 完整流程）
- [ ] Container 隔離驗證（Playwright 腳本）
- [ ] Proxy 路由驗證

#### 5-3. 文件
- [ ] README.md（安裝、使用、API 文件）
- [ ] 最佳實踐（proxy 選擇、帳號養號策略）

**🔍 M4 使用者驗證：**
- 下載 ZIP → 解壓 → `./start.sh` → 瀏覽器開啟
- 建立 3 個 profile（不同 proxy）→ 全部開啟 → 各自登入不同網站
- 關閉 → 重啟 → 登入狀態保留
- 異常關閉（kill）→ 重啟 → 系統恢復正常

---

## Phase 2：C++ 指紋 + VPN（Week 6-9）

### Week 6-7：Fork Camoufox

#### 6-1. Fork + Build 環境
- [ ] Fork daijro/camoufox
- [ ] 搭建 Gecko build 環境（Rust + C++ + Python + mach）
- [ ] 首次完整 build 成功
- [ ] 確認 build 產出的 binary 功能正常

#### 6-2. 閱讀 Camoufox Patch
- [ ] 逐一閱讀所有 fingerprint patch
- [ ] 記錄每個攔截點：API、檔案、攔截方式
- [ ] 分類：DOM 層 vs SpiderMonkey 層
- [ ] 確認 userContextId 取得路徑

#### 6-3. Per-container 指紋 Map
- [ ] C++ HashMap<uint32_t, FingerprintProfile>
- [ ] 啟動時從 profiles/ 載入
- [ ] WebExtension → Native Messaging → C++ 更新 map

#### 6-4. 攔截點改動（按難度）
- [ ] Canvas noise per-container
- [ ] WebGL vendor/renderer per-container
- [ ] Navigator 屬性 per-container
- [ ] User-Agent per-container
- [ ] Screen 屬性 per-container
- [ ] AudioContext per-container
- [ ] Font enumeration per-container
- [ ] Timezone per-container（最難，最後做）

#### 6-5. 驗證
- [ ] 兩個 container 的 Canvas hash 不同
- [ ] 兩個 container 的 WebGL renderer 不同
- [ ] creepjs：lies count = 0, trash count = 0
- [ ] browserleaks.com 無紅色警告

### Week 8：VPN + 加密

#### 8-1. VPN Per-profile
- [ ] wireguard-go 整合
- [ ] 每個 profile 可設定 WireGuard .conf
- [ ] 啟動 profile 時建立 userspace tunnel → 暴露為 local SOCKS5 port
- [ ] 接入現有 proxy 路由

#### 8-2. Profile 加密
- [ ] AES-256-GCM 加密 proxy 帳密和 cookie 備份
- [ ] 主密碼 → PBKDF2 派生金鑰
- [ ] 每個 profile 獨立 IV

### Week 9：測試 + 修 Bug

#### 9-1. 完整測試
- [ ] C++ 指紋隔離 E2E 測試
- [ ] VPN 連線測試
- [ ] 加密/解密測試
- [ ] 三平台 build + 測試
- [ ] CI/CD pipeline（GitHub Actions）

**🔍 使用者驗證：**
- 兩個 profile 訪問 creepjs → 指紋完全不同 → lies = 0
- VPN profile 的出口 IP 正確
- 加密後的 profile 無法被直接讀取

---

## Phase 3：Chromium + MCP（Week 10-14）

### Week 10：Chromium 引擎

#### 10-1. CloakBrowser 整合
- [ ] Binary 下載管理（偵測安裝、引導下載）
- [ ] Chromium launcher：--fingerprint=<seed> + --user-data-dir + --proxy-server + --remote-debugging-port
- [ ] Playwright ConnectOverCDP 取得 browser handle
- [ ] Session 管理（PID tracking、CDP port）

#### 10-2. 雙引擎 API
- [ ] Profile.runtime_id 欄位生效
- [ ] POST /api/sessions 根據 runtime provider capability 分派
- [ ] 所有瀏覽器操作 API 引擎無關（Playwright 抽象）
- [ ] Cookie API 引擎分派（Firefox: extension, Chromium: Playwright）

#### 10-3. 指紋池分引擎
- [ ] fingerprint-suite 生成 Chrome 指紋池
- [ ] 建立 profile 時根據 runtime provider 取用

### Week 11：Sidebar 雙引擎 + 完善

#### 11-1. Sidebar 更新
- [ ] 引擎 badge（🦊/🌐）
- [ ] 新增 Profile 時選擇引擎
- [ ] Chromium profile 開啟為獨立視窗
- [ ] 點擊已開啟的 Chromium profile → 聚焦視窗

#### 11-2. Local Proxy 統一
- [ ] 評估是否統一 Firefox/Chromium 的 proxy 方案為 local proxy process
- [ ] 如果統一：每個 profile 啟動一個 local proxy → 兩個引擎都用 --proxy-server

### Week 12：MCP + Workflow

#### 12-1. MCP Server
- [ ] Go 實作 MCP Server（HTTP POST transport）
- [ ] Tools：open_profile、navigate、click、type、screenshot、get_content、close_profile
- [ ] Resources：profile 列表、session 狀態
- [ ] 引擎無關（AI agent 不需要知道底層引擎）

#### 12-2. YAML Workflow
- [ ] YAML schema 定義（steps、conditions、loops）
- [ ] Go 解析器 + 執行引擎
- [ ] 排程支援（cron）
- [ ] 執行日誌

### Week 13-14：測試 + 文件 + 打包

#### 13-1. E2E 測試
- [ ] Firefox profile 完整流程
- [ ] Chromium profile 完整流程
- [ ] 混合引擎場景（同時開 Firefox + Chromium profile）
- [ ] MCP tool 測試
- [ ] Workflow 執行測試

#### 13-2. 打包更新
- [ ] 打包腳本加入 CloakBrowser 下載引導
- [ ] 更新 start.sh/start.bat
- [ ] 三平台打包測試

#### 13-3. 文件更新
- [ ] README 更新（雙引擎、MCP、Workflow）
- [ ] API 文件更新
- [ ] MCP tool 文件
- [ ] Workflow YAML 範例

**🔍 使用者驗證：**
- 建立 Firefox profile + Chromium profile → 同時開啟 → 各自正常
- MCP 連接 Claude → 「幫我用 profile A 登入 Facebook」→ AI 自動操作
- YAML workflow 排程執行

---

## Phase 4：自主 Fork + Cloud（依需求）

### 4-1. Fork Chromium（AI agent 輔助）
- [ ] 搭建 Chromium build 環境
- [ ] 參考 Camoufox patch，AI 輔助建立 Chromium 等價 patch
- [ ] Navigator、Screen、Canvas、WebGL、Audio、Timezone 攔截
- [ ] 保持 --fingerprint CLI 參數介面（Control Server 不需改）
- [ ] 三平台 build + CI/CD

### 4-2. Cloud Sync
- [ ] S3-compatible 後端（MinIO / AWS S3）
- [ ] Profile + Proxy + Group 同步
- [ ] E2E 加密（AES-GCM + Argon2）
- [ ] 衝突解決策略

### 4-3. 團隊協作
- [ ] Profile 匯出/匯入包
- [ ] 多人共用 profile（鎖定機制）
- [ ] 操作日誌（誰在什麼時候操作了哪個 profile）

**🔍 使用者驗證：**
- 自己的 Chromium fork binary 取代 CloakBrowser → 功能不變
- 兩台電腦同步 profile → 資料一致
- 團隊成員共用 profile → 鎖定機制正常

---

*文件版本：v1.0 | 建立：2026-04-23*
