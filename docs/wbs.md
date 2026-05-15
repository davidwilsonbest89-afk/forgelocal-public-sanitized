# BrowseForge — 五層 WBS 開發計劃

> Archive note: this WBS includes early Camoufox-first and per-container research. The current product contract is the dual-browser architecture in [dual-browser-architecture.md](dual-browser-architecture.md).

> 版本：v0.3 | 更新：2026-04-23
> 層級：L1 開發項目 → L2 子系統 → L3 功能模組 → L4 實作任務 → L5 技術步驟

---

## Phase 0：開工前置作業（Week 0，1-2 天）

> ⚠️ 以下所有項目必須在寫任何正式程式碼之前完成。任何一個 spike 驗證失敗都可能改變整體架構。

### 0.1 技術驗證 Spike

> 目的：用最少程式碼驗證核心架構假設，每個 spike 獨立可執行，結果記錄在 `spike-results.md`。

#### 0.1.1 Container 隔離完整性驗證
- 0.1.1.1 手動驗證 cookie / storage 隔離
  - 0.1.1.1.1 在 Camoufox 中啟用 Container Tabs（about:config → `privacy.userContext.enabled = true`）
  - 0.1.1.1.2 建立 Container A 和 Container B
  - 0.1.1.1.3 Container A 登入 Facebook 帳號 X → Container B 登入 Facebook 帳號 Y → 確認互不干擾
  - 0.1.1.1.4 關閉 Camoufox → 重新開啟 → 確認兩個 container 的 cookie 都保留
- 0.1.1.2 驗證 localStorage / IndexedDB / cache 隔離
  - 0.1.1.2.1 Container A 寫入 localStorage → Container B 讀取 → 確認讀不到
  - 0.1.1.2.2 Container A 載入圖片 → Container B 載入同圖片 → 用 DevTools Network 確認是獨立請求
- **通過標準**：所有儲存機制完全隔離，重啟後保留
- **失敗 Fallback**：改用多實例方案（見 0.4 Fallback 方案）

#### 0.1.2 proxy.onRequest + cookieStoreId 驗證
- 0.1.2.1 寫最小 WebExtension 測試 proxy 路由
  - 0.1.2.1.1 manifest.json：只要 proxy + contextualIdentities 權限
  - 0.1.2.1.2 background.js：proxy.onRequest listener，印出 requestInfo.cookieStoreId
  - 0.1.2.1.3 在兩個不同 container tab 中各自發請求，確認 cookieStoreId 不同
  - 0.1.2.1.4 根據 cookieStoreId 回傳不同 proxy 設定，用 httpbin.org/ip 驗證 IP 不同
- **通過標準**：能根據 container 分流到不同 proxy，DNS 也走 proxy
- **失敗 Fallback**：研究 Firefox PAC script 或 per-container network.proxy 設定

#### 0.1.3 Camoufox Portable 模式驗證
- 0.1.3.1 測試 -profile 參數
  - 0.1.3.1.1 `./camoufox -profile ./test-data -no-remote`
  - 0.1.3.1.2 確認 Firefox profile 資料寫入 ./test-data/ 而非系統預設位置
  - 0.1.3.1.3 關閉 → 搬移整個目錄到另一個路徑 → 重新啟動 → 確認資料保留
- 0.1.3.2 測試未簽署 extension 載入
  - 0.1.3.2.1 在 test-data/prefs.js 加入 `xpinstall.signatures.required = false`
  - 0.1.3.2.2 將 spike 0.1.2 的測試 extension 放入 test-data/extensions/
  - 0.1.3.2.3 啟動 Camoufox → 確認 extension 自動載入且功能正常
- **通過標準**：portable 目錄可搬移，extension 自動載入
- **失敗 Fallback**：研究 Camoufox 的 Python launcher 是否有其他 profile 管理方式

#### 0.1.4 Playwright 連接驗證
- 0.1.4.1 測試 Playwright 連接正在跑的 Camoufox
  - 0.1.4.1.1 啟動 Camoufox with `--remote-debugging-port=9222`（或 Marionette port）
  - 0.1.4.1.2 Python: `playwright.firefox.connect_over_cdp("ws://localhost:9222")` 或對應協議
  - 0.1.4.1.3 確認能列出所有 tab、能操作特定 tab（navigate、click、screenshot）
- 0.1.4.2 測試 Playwright 操作 container tab
  - 0.1.4.2.1 手動開兩個 container tab → Playwright 連入 → 確認能分辨並操作不同 container 的 tab
- **通過標準**：Playwright 能連入並操作指定 container tab
- **失敗 Fallback**：Playwright 直連不可行時，所有操作走 REST API → WebExtension 中轉

#### 0.1.5 Container 數量與記憶體實測
- 0.1.5.1 壓力測試腳本
  - 0.1.5.1.1 用 WebExtension API 批次建立 container + 開 tab：10 / 25 / 50 / 100 個
  - 0.1.5.1.2 每個 tab 載入一個輕量頁面（如 example.com）
  - 0.1.5.1.3 記錄每個階段的：總記憶體、每 tab 平均記憶體、建立耗時、UI 回應速度
- 0.1.5.2 記錄結果
  - 0.1.5.2.1 填入 spike-results.md 的記憶體實測表格
  - 0.1.5.2.2 確定產品的建議帳號上限（基於 8GB / 16GB / 32GB RAM）
- **通過標準**：50 個 container tab 在 16GB RAM 機器上可正常運作
- **失敗 Fallback**：降低產品定位的帳號上限，或加入強制 tab suspend 策略

#### 0.1.7 Content Script 能否呼叫 Camoufox Setter（🔴 最高優先）
- 0.1.7.1 驗證 content script → main world setter 的可達性
  - 0.1.7.1.1 **首選方案：ISOLATED content script + `wrappedJSObject`**
    ```js
    // content-script.js (run_at: document_start, world: ISOLATED)
    // 1. 從 extension storage 讀取該 container 的指紋
    const fp = await browser.storage.local.get(cookieStoreId);
    // 2. 透過 wrappedJSObject 呼叫 main world 的 setter
    window.wrappedJSObject.setCanvasSeed(fp.canvasSeed);
    window.wrappedJSObject.setTimezone(fp.timezone);
    window.wrappedJSObject.setNavigatorPlatform(fp.platform);
    // ... 其他 setter
    ```
    - Firefox 49+ 支援 wrappedJSObject，Camoufox 基於 Firefox 128+ ESR，確認可用
    - wrappedJSObject 直接存取 main world，不受 CSP 限制
    - 注意：setter 是 self-destructing（呼叫後自動刪除），只能呼叫一次
  - 0.1.7.1.2 **備選方案 A：manifest `world: "MAIN"` + inline `<script>` 注入**
    - Firefox 128+ 支援 content_scripts 的 `world: "MAIN"` 參數
    - 但 MAIN world 無法存取 browser.storage API，需要先用 ISOLATED script 讀設定再注入
  - 0.1.7.1.3 **備選方案 B：inline `<script>` tag 注入**
    ```js
    // content-script.js (run_at: document_start)
    const fp = await browser.storage.local.get(cookieStoreId);
    const s = document.createElement('script');
    s.textContent = `window.setCanvasSeed(${fp.canvasSeed});window.setTimezone('${fp.timezone}');`;
    document.documentElement.appendChild(s);
    s.remove();
    ```
    - 同步執行，在 page JS 之前
    - 風險：可能被頁面 CSP 阻擋（但 Camoufox 可控制 CSP 設定）
  - 0.1.7.1.4 驗證步驟：
    1. 在 Camoufox 中載入測試 extension
    2. Content script 呼叫 `window.wrappedJSObject.setCanvasSeed(12345)`
    3. 頁面中執行 Canvas toDataURL → 記錄 hash A
    4. 換一個 seed 重複 → 記錄 hash B
    5. 確認 hash A ≠ hash B（setter 生效）
  - 0.1.7.1.5 驗證 timing：確認 document_start content script 執行時 setter 已存在
    - C++ 層在 window 物件建立時就定義 setter → 早於任何 JS 執行
    - document_start content script 在 page `<script>` 之前執行
    - 因此 setter 在 content script 執行時一定已存在
- **通過標準**：wrappedJSObject 方案能成功呼叫 setter 並改變指紋輸出
- **失敗 Fallback**：用 Playwright `page.evaluate()` 注入（見 0.1.8）

#### 0.1.8 Playwright 能否對 Container Tab 注入 InitScript
- 0.1.8.1 驗證 Playwright 對 container tab 的控制能力
  - 0.1.8.1.1 Control Server 用 Playwright 連入 Camoufox
  - 0.1.8.1.2 WebExtension 開啟一個 container tab
  - 0.1.8.1.3 Playwright 列出所有 page → 找到該 container tab
  - 0.1.8.1.4 對該 page 執行 `page.evaluate(() => window.setCanvasSeed(12345))`
  - 0.1.8.1.5 驗證 setter 是否生效（Canvas hash 改變）
- 0.1.8.2 驗證 Playwright addInitScript 對後續導航的持久性
  - 0.1.8.2.1 對 container tab 執行 `page.addInitScript()` 注入 setter 呼叫
  - 0.1.8.2.2 導航到新頁面 → 確認 setter 仍然生效
- **通過標準**：Playwright 能找到 container tab 並成功注入 setter
- **失敗 Fallback**：瀏覽器操作全部走 WebExtension 中轉（回到原設計）

#### 0.1.9 Camoufox 是否支援 CDP（--remote-debugging-port）
- 0.1.9.1 測試 CDP 端口
  - 0.1.9.1.1 啟動 Camoufox with `--remote-debugging-port=9222`
  - 0.1.9.1.2 嘗試 `curl http://localhost:9222/json/list` 看是否回傳 tab 列表
  - 0.1.9.1.3 如果支援 CDP，測試用 puppeteer 或 chrome-remote-interface 連入
- **通過標準**：CDP 端口可用且能列出 tab
- **失敗影響**：不影響核心功能（Playwright Juggler 是主要方案），但限制了第三方工具整合

#### 0.1.10 Spike 結果記錄
- 0.1.10.1 建立 spike-results.md
  - 0.1.10.1.1 每個 spike 記錄：假設、驗證方式、結果（PASS/FAIL）、發現的問題、決策
  - 0.1.10.1.2 若有 FAIL，記錄選擇的 Fallback 方案及對 WBS 的影響

### 0.2 元件間介面契約

> 目的：在開工前對齊所有元件之間的通訊格式，避免分頭開發後對不上。

#### 0.2.1 Control Server → WebExtension（WebSocket 推送）
- 0.2.1.1 定義完整 message type 清單
  - 0.2.1.1.1 撰寫 `docs/protocol-server-to-extension.md`
  - 0.2.1.1.2 必要 message types（僅 container 管理，瀏覽器操作走 Playwright）：

    ```
    open_tab        { profile_id, url? }           → Extension 開啟 container tab
    close_tab       { profile_id }                 → Extension 關閉 container tab
    close_all_tabs  {}                             → Extension 關閉所有 tab
    get_cookies     { profile_id }                 → Extension 取得 container cookies
    set_cookies     { profile_id, cookies[] }      → Extension 匯入 cookies
    update_profile  { profile_id, changes }        → Extension 更新 container 設定
    delete_profile  { profile_id }                 → Extension 刪除 container
    inject_fingerprint { profile_id, setter_calls } → Extension 注入指紋 setter（Phase 1）
    ```

    > ⚠️ 已移除：navigate、execute_script、take_screenshot — 這些改走 Playwright（§7.4）

  - 0.2.1.1.3 每個 message 定義 request/response 格式：
    ```
    Request:  { id: "req_001", type: "open_tab", payload: { ... } }
    Response: { id: "req_001", success: true, payload: { tab_id: 42 } }
    Error:    { id: "req_001", success: false, error: { code: "NOT_FOUND", message: "..." } }
    ```

#### 0.2.2 WebExtension → Control Server（HTTP 事件回報）
- 0.2.2.1 定義事件回報格式
  - 0.2.2.1.1 撰寫 `docs/protocol-extension-to-server.md`
  - 0.2.2.1.2 事件 types：

    ```
    POST /api/events
    tab_opened       { profile_id, tab_id, url }
    tab_closed       { profile_id, tab_id }
    tab_updated      { profile_id, tab_id, url, title, status }
    need_attention   { profile_id, tab_id, reason, title }
    proxy_error      { profile_id, proxy_host, error }
    extension_ready  { version }
    extension_error  { error, stack }
    ```

#### 0.2.3 REST API 回傳格式統一
- 0.2.3.1 定義統一的 API response envelope
  - 0.2.3.1.1 撰寫 `docs/api-conventions.md`
  - 0.2.3.1.2 成功：`{ data: { ... } }`
  - 0.2.3.1.3 錯誤：`{ error: { code: "PROFILE_NOT_FOUND", message: "..." } }`
  - 0.2.3.1.4 列表：`{ data: [...], total: 42 }`
  - 0.2.3.1.5 HTTP status code 規範：200 成功 / 201 建立 / 400 參數錯誤 / 401 未認證 / 404 不存在 / 500 內部錯誤

### 0.3 里程碑驗收標準

> 每個里程碑定義可觀察的行為，作為「做到什麼程度算完成」的判斷依據。

#### 0.3.1 M1: 骨架跑通（Week 1 末）
- 0.3.1.1 驗收項目
  - 0.3.1.1.1 `./start.sh` 能啟動 Control Server + Camoufox，無報錯
  - 0.3.1.1.2 `curl http://localhost:19280/api/status` 回傳 200 + JSON
  - 0.3.1.1.3 Camoufox 視窗出現，sidebar 顯示空的 profile 列表
  - 0.3.1.1.4 `POST /api/profiles` 能建立 profile，`GET /api/profiles` 能列出
  - 0.3.1.1.5 關閉 Camoufox → Control Server 自動退出

#### 0.3.2 M2: 單 Profile 完整流程（Week 2 末）
- 0.3.2.1 驗收項目
  - 0.3.2.1.1 Sidebar 點擊「新增 Profile」→ 填寫名稱和 proxy → 建立成功
  - 0.3.2.1.2 點擊 profile → 開啟 Container Tab → tab 走指定 proxy（httpbin.org/ip 驗證）
  - 0.3.2.1.3 在 tab 中登入某網站 → 關閉 tab → 重新點擊 profile 開啟 → cookie 保留，仍在登入狀態
  - 0.3.2.1.4 建立第二個 profile（不同 proxy）→ 兩個 tab 同時開啟 → IP 不同 → cookie 不互通
  - 0.3.2.1.5 Sidebar 顯示 🟢（已開啟）/ ⚪（未開啟）狀態正確

#### 0.3.3 M3: API 完整可用（Week 3 末）
- 0.3.3.1 驗收項目
  - 0.3.3.1.1 外部 curl 能完成：建立 profile → 開 session → navigate → screenshot → 關閉 session → 刪除 profile
  - 0.3.3.1.2 Playwright `firefox.connect()` 能連入並操作指定 tab
  - 0.3.3.1.3 指紋生成器產出的指紋通過一致性檢查（UA/platform/screen 合理配對）
  - 0.3.3.1.4 sidebar 搜尋、分組摺疊、右鍵選單功能正常
  - 0.3.3.1.5 API 認證生效：無 token 的請求被拒絕

#### 0.3.4 M4: 可發佈（Week 4 末）
- 0.3.4.1 驗收項目
  - 0.3.4.1.1 打包出 ZIP → 在乾淨機器（無 Go/Node）解壓 → `./start.sh` 啟動成功
  - 0.3.4.1.2 建立 5 個 profile，各自設不同 proxy → 全部開啟 → 各自登入不同網站 → 全部正常
  - 0.3.4.1.3 關閉 → 重啟 → 所有 profile 的登入狀態保留
  - 0.3.4.1.4 Tab suspend 功能正常（非活躍 tab 自動 suspend → 點擊恢復）
  - 0.3.4.1.5 異常關閉（kill process）→ 重啟 → 系統自動恢復，無孤兒 container
  - 0.3.4.1.6 Go 單元測試全部通過，E2E 測試腳本跑通

### 0.4 Fallback 方案

> 如果 spike 階段發現 container 方案有致命問題，以下是預定義的退路。

#### 0.4.1 Fallback A: 混合模式
- 0.4.1.1 低風險帳號用 container（省資源）
  - 0.4.1.1.1 同平台 ≤3 帳號、同區域 proxy → container 方案
- 0.4.1.2 高風險帳號用獨立 Camoufox 實例
  - 0.4.1.2.1 同平台 >3 帳號、跨區域 proxy → 獨立實例，各自有完整指紋
  - 0.4.1.2.2 Control Server 同時管理 container tab 和獨立實例
- 0.4.1.3 觸發條件
  - 0.4.1.3.1 Container 隔離驗證（0.1.1）部分通過但有邊界 case
  - 0.4.1.3.2 Container 數量上限 < 50

#### 0.4.2 Fallback B: 全面多實例 + 資源管理
- 0.4.2.1 放棄 container 方案，每個 profile 一個 Camoufox process
  - 0.4.2.1.1 用 process suspend（SIGSTOP/SIGCONT）管理非活躍實例的記憶體
  - 0.4.2.1.2 UI 改為多視窗管理器（類似 Donut Browser 的做法）
  - 0.4.2.1.3 指紋隔離天然完美（每個實例獨立設定）
- 0.4.2.2 觸發條件
  - 0.4.2.2.1 Container 隔離驗證（0.1.1）失敗
  - 0.4.2.2.2 proxy.onRequest 無法取得 cookieStoreId（0.1.2 失敗）

---

## 1. 瀏覽器核心整合

### 1.1 Portable 目錄結構設計
#### 1.1.1 根目錄佈局定義
- 1.1.1.1 定義頂層目錄結構（camoufox binary、extension/、profiles/、data/、config.json）
  - 1.1.1.1.1 建立目錄結構模板，寫入 README 說明每個目錄用途
- 1.1.1.2 設計跨平台路徑解析策略
  - 1.1.1.2.1 Windows backslash vs Unix forward slash 統一處理（Go filepath 套件）
  - 1.1.1.2.2 相對路徑解析：所有路徑相對於執行檔所在目錄

#### 1.1.2 Firefox Profile Data 目錄
- 1.1.2.1 data/ 目錄作為 Firefox -profile 指向的目標
  - 1.1.2.1.1 建立 prefs.js 預設模板（關閉自動更新、遙測、crash reporter）
  - 1.1.2.1.2 預設 about:config 設定（media.peerconnection.enabled=false 等安全設定）
- 1.1.2.2 profiles/ 目錄存放使用者 profile JSON
  - 1.1.2.2.1 每個 profile 一個子目錄：profiles/{profile_id}/profile.json + firefox-data/

#### 1.1.3 config.json 全域設定檔
- 1.1.3.1 定義設定結構（API port、Camoufox binary 路徑、預設 proxy 等）
  - 1.1.3.1.1 JSON Schema 定義 + 預設值 fallback 邏輯
  - 1.1.3.1.2 首次啟動時自動生成預設 config.json

### 1.2 啟動參數與流程配置
#### 1.2.1 Camoufox 啟動參數
- 1.2.1.1 -profile ./data 指向 portable profile 目錄
  - 1.2.1.1.1 啟動前檢查 data/ 目錄是否存在，不存在則建立並初始化
- 1.2.1.2 -no-remote 防止連到其他 Firefox 實例
  - 1.2.1.2.1 偵測是否已有實例在跑，給出明確錯誤訊息
- 1.2.1.3 其他必要參數（--new-instance、視窗大小等）
  - 1.2.1.3.1 從 config.json 讀取可自訂的啟動參數

#### 1.2.2 啟動順序編排
- 1.2.2.1 Control Server 啟動
  - 1.2.2.1.1 Go server 啟動，載入 config.json 和 profiles
  - 1.2.2.1.2 寫入 PID file，開放 health check endpoint
- 1.2.2.2 Control Server 透過 Playwright 啟動 Camoufox
  - 1.2.2.2.1 建構 CAMOU_CONFIG env var（基底指紋）
  - 1.2.2.2.2 呼叫 `playwright.firefox.launch(executable_path, env, firefox_user_prefs, args)`
  - 1.2.2.2.3 取得 browser handle，存入 server 全域狀態
  - 1.2.2.2.4 Extension 隨 Camoufox 啟動自動載入（透過 -profile 目錄中的 extensions 設定）
- 1.2.2.3 WebExtension 初始化，連線 Control Server
  - 1.2.2.3.1 Background script 啟動時建立 WebSocket 連線到 server
  - 1.2.2.3.2 Server 收到 extension_ready 事件後，系統進入就緒狀態

#### 1.2.3 關閉流程
- 1.2.3.1 優雅關閉：使用者關閉瀏覽器 → Playwright 偵測到 browser disconnected → server 清理
  - 1.2.3.1.1 Playwright browser.on('disconnected') 觸發清理
  - 1.2.3.1.2 Control Server 清理所有 session 狀態，釋放資源
  - 1.2.3.1.3 Control Server 自動退出
- 1.2.3.2 API 觸發關閉：POST /api/shutdown
  - 1.2.3.2.1 通知 WebExtension 清理 → Playwright `browser.close()` → server 退出

### 1.3 Extension 自動載入機制
#### 1.3.1 開發模式載入
- 1.3.1.1 about:debugging 臨時載入（開發期間）
  - 1.3.1.1.1 文件說明開發時如何載入未簽署 extension
- 1.3.1.2 xpinstall.signatures.required = false 設定
  - 1.3.1.2.1 寫入 prefs.js 預設模板

#### 1.3.2 正式部署載入
- 1.3.2.1 將 extension 放入 profile/extensions/ 目錄
  - 1.3.2.1.1 extension 打包為 .xpi 或使用目錄載入方式
  - 1.3.2.1.2 在 prefs.js 中設定 extensions.autoDisableScopes = 0 避免被停用

### 1.4 Phase 2: Fork Camoufox + CI/CD
#### 1.4.1 建立 Fork
- 1.4.1.1 Fork daijro/camoufox 到自己的 repo
  - 1.4.1.1.1 閱讀 Camoufox 的 patch 結構，理解它如何管理對 Firefox 的改動
  - 1.4.1.1.2 建立 branch 策略：main（穩定）、dev（開發）、upstream-sync（同步上游）
- 1.4.1.2 本地 build 環境搭建
  - 1.4.1.2.1 安裝 Gecko build 依賴（Rust、C++ toolchain、Python、mach）
  - 1.4.1.2.2 完成首次完整 build，確認產出可執行的 Camoufox binary

#### 1.4.2 CI/CD Pipeline
- 1.4.2.1 GitHub Actions 自動 build（Linux/macOS/Windows）
  - 1.4.2.1.1 設定三平台 build matrix
  - 1.4.2.1.2 build artifact 上傳到 GitHub Releases
- 1.4.2.2 上游同步自動化
  - 1.4.2.2.1 定期 fetch upstream，自動嘗試 rebase，衝突時通知

---

## 2. Container 隔離系統

### 2.1 Container 建立與刪除
#### 2.1.1 contextualIdentities API 操作
- 2.1.1.1 browser.contextualIdentities.create() 建立新 container
  - 2.1.1.1.1 設定 container name = profile name，color 按分組分配
  - 2.1.1.1.2 記錄 container cookieStoreId 與 profile ID 的映射關係
- 2.1.1.2 browser.contextualIdentities.remove() 刪除 container
  - 2.1.1.2.1 刪除前先關閉該 container 的所有 tab
  - 2.1.1.2.2 清除該 container 的 cookie 和 storage（browsingData.remove）

#### 2.1.2 Container 數量限制測試
- 2.1.2.1 測試 Firefox container 數量上限
  - 2.1.2.1.1 建立 100、500、1000 個 container，觀察效能和穩定性
  - 2.1.2.1.2 記錄記憶體佔用曲線
- 2.1.2.2 設計超限處理策略
  - 2.1.2.2.1 若有上限，實作 container 池回收機制（LRU 淘汰不活躍 container）

### 2.2 Container Tab 生命週期管理
#### 2.2.1 開啟 Container Tab
- 2.2.1.1 browser.tabs.create({ cookieStoreId }) 在指定 container 開 tab
  - 2.2.1.1.1 開啟後導航到 profile 設定的首頁（預設 about:blank）
  - 2.2.1.1.2 記錄 tab ID → session ID 映射
- 2.2.1.2 同一 container 多 tab 處理
  - 2.2.1.2.1 決定策略：一個 container 只允許一個 tab，或允許多 tab

#### 2.2.2 關閉 Container Tab
- 2.2.2.1 browser.tabs.remove() 關閉指定 tab
  - 2.2.2.1.1 關閉時更新 session 狀態為 inactive
  - 2.2.2.1.2 觸發 profile.last_used 時間戳更新
- 2.2.2.2 批次關閉（按分組、全部）
  - 2.2.2.2.1 查詢所有屬於目標 container 的 tab，逐一關閉

#### 2.2.3 Tab 狀態監控
- 2.2.3.1 browser.tabs.onUpdated 監聽 tab 狀態變化
  - 2.2.3.1.1 追蹤 loading / complete 狀態，更新 sidebar 顯示
- 2.2.3.2 browser.tabs.onRemoved 監聽 tab 被使用者手動關閉
  - 2.2.3.2.1 同步更新 session 狀態，通知 Control Server

### 2.3 Container-Profile 綁定機制
#### 2.3.1 映射表管理
- 2.3.1.1 維護 profileId ↔ cookieStoreId 雙向映射
  - 2.3.1.1.1 存儲在 browser.storage.local，啟動時載入
  - 2.3.1.1.2 Profile 刪除時同步刪除映射和 container
- 2.3.1.2 啟動時一致性檢查
  - 2.3.1.2.1 比對 profiles/ 目錄和 container 列表，清理孤兒 container

#### 2.3.2 Container 屬性同步
- 2.3.2.1 Profile 更名時同步更新 container name
  - 2.3.2.1.1 browser.contextualIdentities.update() 更新 container 屬性
- 2.3.2.2 Profile 分組變更時更新 container color
  - 2.3.2.2.1 定義分組 → 顏色映射表

### 2.4 跨 Container 隔離驗證
#### 2.4.1 Cookie 隔離驗證
- 2.4.1.1 在兩個 container 分別登入同一網站不同帳號
  - 2.4.1.1.1 驗證 cookie 不互相干擾
  - 2.4.1.1.2 驗證 localStorage / sessionStorage / IndexedDB 隔離
- 2.4.1.2 自動化測試腳本
  - 2.4.1.2.1 Playwright 腳本：開兩個 container tab → 各自設 cookie → 互相讀取確認隔離

#### 2.4.2 Cache 隔離驗證
- 2.4.2.1 驗證 HTTP cache 不跨 container 共享
  - 2.4.2.1.1 Container A 載入圖片 → Container B 載入同圖片 → 確認發出獨立請求

---

## 3. 指紋系統

### 3.1 指紋資料模型設計
#### 3.1.1 指紋欄位定義
- 3.1.1.1 定義完整指紋結構（UA、platform、screen、timezone、canvas seed、webgl、audio、fonts 等）
  - 3.1.1.1.1 參考 creepjs / browserleaks 的檢測項目，確保覆蓋所有主要指紋維度
  - 3.1.1.1.2 定義 JSON Schema，每個欄位有型別、範圍、預設值

#### 3.1.2 指紋一致性規則
- 3.1.2.1 定義欄位間的約束關係
  - 3.1.2.1.1 UA 版本 ↔ platform 一致性（Windows UA 不能配 MacIntel platform）
  - 3.1.2.1.2 Screen resolution ↔ device memory ↔ hardware concurrency 合理配對
  - 3.1.2.1.3 Timezone ↔ locale ↔ proxy IP 地理位置一致性
  - 3.1.2.1.4 WebGL vendor/renderer 必須是真實存在的顯卡型號

### 3.2 指紋生成器（基於 fingerprint-suite）
#### 3.2.1 fingerprint-suite 整合
- 3.2.1.1 建立指紋預生成工具
  - 3.2.1.1.1 Node.js 腳本：呼叫 `FingerprintGenerator.getFingerprint({ browsers: ['firefox'] })`
  - 3.2.1.1.2 批次生成：`--os windows --count 500` / `--os macos --count 500` / `--os linux --count 200`
  - 3.2.1.1.3 輸出為 JSON 檔案：`data/fingerprints-windows.json` 等
- 3.2.1.2 fingerprint-suite → Camoufox 格式轉換器
  - 3.2.1.2.1 映射 `fingerprint.navigator.userAgent` → `"navigator.userAgent"`
  - 3.2.1.2.2 映射 `fingerprint.screen.width` → `"screen.width"`
  - 3.2.1.2.3 映射 `fingerprint.videoCard.renderer` → `"webGl:renderer"`
  - 3.2.1.2.4 補充 fingerprint-suite 不生成的欄位：`canvas:seed`、`audio:seed`、`fonts:spacing_seed`（隨機 uint32）
  - 3.2.1.2.5 轉換器可用 Node.js 或 Go 實作（Go 更適合嵌入 Control Server）

#### 3.2.2 指紋池管理
- 3.2.2.1 Go Control Server 啟動時載入指紋池
  - 3.2.2.1.1 讀取 `data/fingerprints-*.json` → 按 OS 分類存入記憶體
  - 3.2.2.1.2 建立 profile 時從對應 OS 池中隨機取一組
  - 3.2.2.1.3 取用後標記為已使用（避免兩個 profile 用同一組指紋）
- 3.2.2.2 指紋池耗盡時的處理
  - 3.2.2.2.1 池中剩餘 < 10% 時在 sidebar 顯示警告
  - 3.2.2.2.2 提供 `make generate-fingerprints` 命令重新生成

#### 3.2.3 Proxy-Fingerprint 地理一致性（自動調配）
- 3.2.3.1 GeoIP 查詢
  - 3.2.3.1.1 建立/更新 profile 設定 proxy 時，自動查詢 proxy IP 的地理資訊
  - 3.2.3.1.2 查詢結果：country、city、timezone、latitude、longitude、language code
  - 3.2.3.1.3 使用 MaxMind GeoLite2（§13.3）或簡化版 IP-country 映射表
- 3.2.3.2 自動覆寫地理相關指紋欄位
  - 3.2.3.2.1 `timezone` → GeoIP 查到的 timezone（如 `"Asia/Tokyo"`）
  - 3.2.3.2.2 `locale:language` → 對應國家的主要語言（如 `"ja"`）
  - 3.2.3.2.3 `locale:region` → 國家代碼（如 `"JP"`）
  - 3.2.3.2.4 `navigator.language` → `"ja-JP"`（language-region 格式）
  - 3.2.3.2.5 `navigator.languages` → `["ja-JP", "ja", "en-US", "en"]`（主語言 + 英文 fallback）
  - 3.2.3.2.6 `headers.Accept-Language` → `"ja-JP,ja;q=0.9,en-US;q=0.8,en;q=0.7"`
  - 3.2.3.2.7 `geolocation:latitude` / `geolocation:longitude` → GeoIP 查到的城市座標
  - 3.2.3.2.8 `geolocation:accuracy` → 城市級精度（~10000 公尺）
- 3.2.3.3 不動的欄位（與地理無關，保持指紋池原值）
  - 3.2.3.3.1 screen、GPU、canvas/audio/font seed、navigator.platform、hardwareConcurrency、deviceMemory
- 3.2.3.4 Proxy 變更時自動重新調配
  - 3.2.3.4.1 PUT /api/profiles/:id 更新 proxy 時，自動觸發 GeoIP 查詢 + 覆寫地理欄位
  - 3.2.3.4.2 覆寫後通知 WebExtension 更新該 container 的 setter 參數
  - 3.2.3.4.3 若該 profile 有活躍 session，下次頁面導航時生效
- 3.2.3.5 無 proxy 時的處理
  - 3.2.3.5.1 使用系統 timezone 和 locale
  - 3.2.3.5.2 geolocation 欄位留空（不偽造）
- 3.2.3.6 Country → Language 映射表
  - 3.2.3.6.1 內建常用國家的主要語言映射（~50 國）：
    ```
    US → en-US, JP → ja-JP, KR → ko-KR, DE → de-DE,
    FR → fr-FR, BR → pt-BR, TW → zh-TW, CN → zh-CN,
    RU → ru-RU, TH → th-TH, VN → vi-VN, ID → id-ID, ...
    ```
  - 3.2.3.6.2 未知國家 fallback 到 `en-US`

### 3.3 Phase 1: Per-container 指紋注入（Camoufox Setter）
#### 3.3.1 Camoufox 啟動時設定基底指紋
- 3.3.1.1 用 CAMOU_CONFIG env var 設定 process-wide 基底指紋
  - 3.3.1.1.1 從指紋池取一組作為基底（所有 container 的 fallback）
  - 3.3.1.1.2 JSON 序列化 → 按 OS 限制分段為 CAMOU_CONFIG_1, CAMOU_CONFIG_2, ...
  - 3.3.1.1.3 傳入 Playwright `firefox.launch(env={...})` 的 env 參數

#### 3.3.2 Per-container 指紋覆寫
- 3.3.2.1 每個 Container Tab 開啟時注入 setter 呼叫
  - 3.3.2.1.1 **確認可行的注入機制**（基於 Firefox API 研究）：
    - 首選：ISOLATED content script（`run_at: document_start`）透過 `wrappedJSObject` 呼叫 setter
    - Firefox 128+ 的 wrappedJSObject 可從 content script 直接存取 main world 函數
    - C++ 層 setter 在 window 物件建立時就定義，早於 content script 執行，timing 安全
  - 3.3.2.1.2 注入流程：
    ```
    Container Tab 開啟 → 頁面開始載入
      → document_start content script 啟動
      → 從 browser.storage.local 讀取該 container 的指紋設定
      → 透過 wrappedJSObject 呼叫 setter：
        window.wrappedJSObject.setCanvasSeed(fp["canvas:seed"])
        window.wrappedJSObject.setAudioFingerprintSeed(fp["audio:seed"])
        window.wrappedJSObject.setFontSpacingSeed(fp["fonts:spacing_seed"])
        window.wrappedJSObject.setNavigatorPlatform(fp["navigator.platform"])
        window.wrappedJSObject.setNavigatorUserAgent(fp["navigator.userAgent"])
        window.wrappedJSObject.setNavigatorOscpu(fp["navigator.oscpu"])
        window.wrappedJSObject.setWebGLVendor(fp["webGl:vendor"])
        window.wrappedJSObject.setWebGLRenderer(fp["webGl:renderer"])
        window.wrappedJSObject.setScreenDimensions(fp["screen.width"], fp["screen.height"])
        window.wrappedJSObject.setScreenColorDepth(fp["screen.colorDepth"])
        window.wrappedJSObject.setTimezone(fp["timezone"])
        window.wrappedJSObject.setWebRTCIPv4(fp["webrtc:ipv4"])
        window.wrappedJSObject.setFontList(fp["fonts"].join(","))
      → setter 自動銷毀（self-destructing）
      → 頁面 JS 開始執行（此時指紋已生效）
    ```
  - 3.3.2.1.3 指紋設定寫入 storage 的時機：
    - Profile 建立時：Control Server 通知 Extension → Extension 將指紋寫入 `browser.storage.local`
    - Key 格式：`fp_{cookieStoreId}` → 指紋 JSON
    - Proxy 變更時：Control Server 重新計算地理欄位 → 通知 Extension 更新 storage

#### 3.3.3 Phase 1 已知限制
- 3.3.3.1 文件化 setter 方案的限制
  - 3.3.3.1.1 setter 是 JS 呼叫，呼叫痕跡可能被進階偵測發現
  - 3.3.3.1.2 WebGL parameters/extensions/shaderPrecisionFormats 等深層屬性無法透過 setter 覆寫（仍共用基底）
  - 3.3.3.1.3 Battery API、MediaDevices 等無 setter 的屬性仍共用

### 3.4 Phase 2: C++ Per-container 指紋偽造
#### 3.4.1 Camoufox Patch 分析
- 3.4.1.1 列出 Camoufox 所有指紋攔截 patch
  - 3.4.1.1.1 逐一閱讀 patch，記錄：攔截的 API、修改的檔案、攔截方式
  - 3.4.1.1.2 分類：DOM 層（容易取 BrowsingContext）vs SpiderMonkey 層（較難）
- 3.4.1.2 確認每個攔截點取得 userContextId 的路徑
  - 3.4.1.2.1 DOM 層：nsINode → Document → BrowsingContext → OriginAttributes → userContextId
  - 3.4.1.2.2 SpiderMonkey 層：JSContext → Realm → Global → 回到 DOM 取 context

#### 3.4.2 指紋 Profile Map 實作
- 3.4.2.1 C++ HashMap<uint32_t, FingerprintProfile> 全域 map
  - 3.4.2.1.1 定義 FingerprintProfile C++ struct（對應 JSON 欄位）
  - 3.4.2.1.2 啟動時從 profiles/ 目錄載入所有指紋到 map
- 3.4.2.2 動態更新機制
  - 3.4.2.2.1 WebExtension → Native Messaging → C++ 層更新 map
  - 3.4.2.2.2 新增/修改/刪除 profile 時即時同步

#### 3.4.3 攔截點改動（按難度排序）
- 3.4.3.1 Canvas noise（dom/canvas/）
  - 3.4.3.1.1 找到 toDataURL / toBlob / getImageData 攔截點
  - 3.4.3.1.2 取得當前 Document → BrowsingContext → userContextId
  - 3.4.3.1.3 從 map 取 canvas_noise_seed，替換全域 seed
- 3.4.3.2 WebGL getParameter（dom/canvas/WebGLContext*）
  - 3.4.3.2.1 攔截 UNMASKED_VENDOR_WEBGL / UNMASKED_RENDERER_WEBGL
  - 3.4.3.2.2 per-container 回傳不同 vendor/renderer 字串
- 3.4.3.3 Navigator 屬性（dom/base/Navigator.cpp）
  - 3.4.3.3.1 platform、hardwareConcurrency、deviceMemory per-container
- 3.4.3.4 User-Agent（netwerk/protocol/http/）
  - 3.4.3.4.1 HTTP header 層的 UA per-container
  - 3.4.3.4.2 navigator.userAgent JS 層 per-container
- 3.4.3.5 Screen 屬性（dom/base/Screen.cpp）
  - 3.4.3.5.1 width / height / availWidth / availHeight / colorDepth per-container
- 3.4.3.6 AudioContext（dom/media/webaudio/）
  - 3.4.3.6.1 OscillatorNode 輸出加 per-container noise
- 3.4.3.7 Timezone — 最難（js/src/builtin/intl/ + js/src/jsdate.cpp）
  - 3.4.3.7.1 Intl.DateTimeFormat.resolvedOptions().timeZone per-container
  - 3.4.3.7.2 Date.getTimezoneOffset() per-container
  - 3.4.3.7.3 需要從 SpiderMonkey JSContext 回溯到 DOM BrowsingContext 取 userContextId
- 3.4.3.8 Font enumeration
  - 3.4.3.8.1 攔截字型列舉 API，per-container 回傳不同字型子集

---

## 4. Proxy 系統

### 4.1 Per-container Proxy 路由
#### 4.1.1 proxy.onRequest 監聽器
- 4.1.1.1 實作 proxy.onRequest listener
  - 4.1.1.1.1 從 requestInfo.cookieStoreId 取得 container ID
  - 4.1.1.1.2 查詢 container → profile 映射，取得 proxy 設定
  - 4.1.1.1.3 回傳 { type, host, port, username, password, proxyDNS: true }
  - 4.1.1.1.4 無 proxy 設定時回傳 { type: "direct" }

#### 4.1.2 Proxy DNS 處理
- 4.1.2.1 確保 DNS 查詢走 proxy（防止 DNS leak）
  - 4.1.2.1.1 proxyDNS: true 強制 DNS 透過 proxy 解析
  - 4.1.2.1.2 驗證：用 DNS leak test 網站確認

### 4.2 Proxy 認證管理
#### 4.2.1 帳密認證
- 4.2.1.1 proxy.onRequest 回傳 username/password
  - 4.2.1.1.1 支援 HTTP Basic Auth 和 SOCKS5 認證
- 4.2.1.2 webRequest.onAuthRequired 處理認證彈窗
  - 4.2.1.2.1 攔截 407 回應，自動填入 proxy 帳密，不彈窗給使用者

#### 4.2.2 認證資訊安全儲存
- 4.2.2.1 Proxy 帳密存在 profile JSON 中
  - 4.2.2.1.1 Phase 1: 明文存儲（portable 優先）
  - 4.2.2.1.2 Phase 2: AES-256 加密存儲

### 4.3 Proxy 健康檢查
#### 4.3.1 連線測試
- 4.3.1.1 定期透過 proxy 發送測試請求
  - 4.3.1.1.1 GET http://httpbin.org/ip 確認 proxy 可用且 IP 正確
  - 4.3.1.1.2 記錄延遲和成功率
- 4.3.1.2 失敗處理
  - 4.3.1.2.1 Proxy 失敗時在 sidebar 顯示警告圖示
  - 4.3.1.2.2 可選：自動切換到備用 proxy

### 4.4 Phase 2: Proxy 池管理
#### 4.4.1 Proxy 池資料模型
- 4.4.1.1 支援 static（固定 IP 列表）和 rotating（旋轉端口）兩種模式
  - 4.4.1.1.1 定義 proxy pool JSON 結構
  - 4.4.1.1.2 Profile 可綁定到 pool 而非單一 proxy
- 4.4.1.2 自動分配邏輯
  - 4.4.1.2.1 新 profile 建立時從 pool 中分配未使用的 proxy
  - 4.4.1.2.2 Round-robin 或最少使用策略

---

## 5. Profile 系統

### 5.1 Profile 資料模型
#### 5.1.1 核心欄位定義
- 5.1.1.1 定義 Profile JSON 結構
  - 5.1.1.1.1 必要欄位：id、name、created_at、last_used
  - 5.1.1.1.2 指紋欄位：fingerprint 物件（**Camoufox 扁平格式**，直接存 `"navigator.userAgent": "..."` 等）
  - 5.1.1.1.3 Proxy 欄位：proxy 物件（type、host、port、username、password）
  - 5.1.1.1.4 分組欄位：group（字串）、tags（字串陣列）
  - 5.1.1.1.5 內部欄位：container_id、firefox_profile_dir
  - 5.1.1.1.6 指紋格式範例：
    ```json
    {
      "id": "prof_a1b2c3",
      "name": "FB Brand #1",
      "engine": "firefox",
      "fingerprint": {
        "navigator.userAgent": "Mozilla/5.0 ...",
        "navigator.platform": "Win32",
        "navigator.hardwareConcurrency": 8,
        "screen.width": 1920,
        "screen.height": 1080,
        "webGl:renderer": "ANGLE (NVIDIA, ...)",
        "webGl:vendor": "Google Inc. (NVIDIA)",
        "timezone": "America/New_York",
        "canvas:seed": 3948271650,
        "audio:seed": 1293847562,
        "fonts:spacing_seed": 2847361952
      },
      "proxy": { "type": "socks5", "host": "...", "port": 1080 },
      "group": "客戶A",
      "tags": ["facebook"]
    }
    ```
  - 5.1.1.1.7 `engine` 欄位：`"firefox"` | `"chromium"`（Phase 1 只有 firefox，Phase 3 加 chromium）
  - 5.1.1.1.8 Chromium profile 額外欄位：`fingerprint_seed`（uint32，CloakBrowser 用）
  - 5.1.1.1.9 好處：fingerprint 欄位可直接傳入 CAMOU_CONFIG 或 setter，無需格式轉換

#### 5.1.2 ID 生成策略
- 5.1.2.1 使用 prof_ 前綴 + 隨機短 ID
  - 5.1.2.1.1 Go: crypto/rand 生成 6 bytes → hex encode → prof_a1b2c3d4
  - 5.1.2.1.2 確保唯一性：生成後檢查 profiles/ 目錄是否已存在

### 5.2 檔案系統持久化
#### 5.2.1 儲存結構
- 5.2.1.1 每個 profile 一個目錄
  - 5.2.1.1.1 profiles/{profile_id}/profile.json — profile 設定
  - 5.2.1.1.2 profiles/{profile_id}/firefox-data/ — 該 profile 的 Firefox 資料（cookie、cache）
  - 5.2.1.1.3 profiles/{profile_id}/cookies-backup.json — cookie 匯出備份

#### 5.2.2 讀寫操作
- 5.2.2.1 Go 層：JSON 讀寫 + 檔案鎖
  - 5.2.2.1.1 讀取：os.ReadFile → json.Unmarshal
  - 5.2.2.1.2 寫入：json.Marshal → os.WriteFile（atomic write: 先寫 .tmp 再 rename）
  - 5.2.2.1.3 檔案鎖：sync.RWMutex per profile，防止並發寫入衝突

#### 5.2.3 啟動時載入
- 5.2.3.1 掃描 profiles/ 目錄，載入所有 profile.json 到記憶體
  - 5.2.3.1.1 建立 map[string]*Profile 索引（by ID）
  - 5.2.3.1.2 建立 group → []Profile 索引（by group）
  - 5.2.3.1.3 載入失敗的 profile 記錄 warning log，不阻斷啟動

### 5.3 Profile CRUD 操作
#### 5.3.1 Create
- 5.3.1.1 建立 profile 目錄 + 生成 profile.json
  - 5.3.1.1.1 生成 ID → 建立目錄 → 生成指紋（若未提供）→ 寫入 JSON
  - 5.3.1.1.2 通知 WebExtension 建立對應 container
  - 5.3.1.1.3 回傳完整 profile 物件

#### 5.3.2 Read
- 5.3.2.1 單一查詢 + 列表查詢
  - 5.3.2.1.1 GET /profiles/:id → 從記憶體 map 取
  - 5.3.2.1.2 GET /profiles?group=X&tag=Y → 過濾 + 排序

#### 5.3.3 Update
- 5.3.3.1 部分更新（PATCH 語意）
  - 5.3.3.1.1 只更新提供的欄位，未提供的保持不變
  - 5.3.3.1.2 指紋變更時通知 WebExtension 更新 container 設定
  - 5.3.3.1.3 Proxy 變更時即時生效（下次請求使用新 proxy）

#### 5.3.4 Delete
- 5.3.4.1 刪除 profile 及所有關聯資料
  - 5.3.4.1.1 關閉該 profile 的活躍 session
  - 5.3.4.1.2 通知 WebExtension 刪除 container
  - 5.3.4.1.3 刪除 profiles/{profile_id}/ 整個目錄
  - 5.3.4.1.4 從記憶體 map 移除

#### 5.3.5 Duplicate
- 5.3.5.1 複製 profile（新 ID、新指紋、保留 proxy 和分組）
  - 5.3.5.1.1 深拷貝 profile → 生成新 ID → 重新生成指紋 → 儲存

### 5.4 Profile 分組與標籤
#### 5.4.1 分組管理
- 5.4.1.1 分組為字串欄位，不需要預先定義
  - 5.4.1.1.1 從所有 profile 的 group 欄位動態收集分組列表
  - 5.4.1.1.2 支援拖拉變更分組（UI 層實作）

#### 5.4.2 標籤管理
- 5.4.2.1 tags 為字串陣列
  - 5.4.2.1.1 支援多標籤篩選（AND / OR）
  - 5.4.2.1.2 標籤自動完成（從現有標籤池建議）

### 5.5 Phase 2: Profile 加密
#### 5.5.1 加密方案
- 5.5.1.1 AES-256-GCM 加密 profile.json
  - 5.5.1.1.1 使用者設定主密碼 → PBKDF2 派生加密金鑰
  - 5.5.1.1.2 每個 profile 獨立 IV
  - 5.5.1.1.3 加密範圍：proxy 帳密、cookie 備份（指紋不加密，不敏感）

---

## 6. WebExtension 管理介面

### 6.1 Extension 骨架
#### 6.1.1 manifest.json
- 6.1.1.1 定義 permissions 和 API 存取
  - 6.1.1.1.1 必要 permissions: contextualIdentities, cookies, proxy, tabs, storage, webRequest, webRequestBlocking
  - 6.1.1.1.2 host_permissions: <all_urls>（proxy 路由需要）
  - 6.1.1.1.3 sidebar_action 指向 sidebar/index.html

#### 6.1.2 Background Script
- 6.1.2.1 Extension 核心邏輯入口
  - 6.1.2.1.1 啟動時連線 Control Server，載入 profile 列表
  - 6.1.2.1.2 註冊 proxy.onRequest listener
  - 6.1.2.1.3 註冊 tabs.onUpdated / onRemoved listener
  - 6.1.2.1.4 處理 sidebar ↔ background 的 message passing

### 6.2 Sidebar Panel UI
#### 6.2.1 UI 框架選擇
- 6.2.1.1 純 HTML/CSS/JS（無框架，保持輕量）
  - 6.2.1.1.1 sidebar/index.html — 主結構
  - 6.2.1.1.2 sidebar/style.css — 樣式（暗色主題，配合 Firefox）
  - 6.2.1.1.3 sidebar/app.js — 互動邏輯

#### 6.2.2 Profile 列表顯示
- 6.2.2.1 按分組摺疊顯示
  - 6.2.2.1.1 分組標題可展開/收合
  - 6.2.2.1.2 每個 profile 顯示：名稱、狀態圖示（🟢/⚪）、proxy IP 國旗
  - 6.2.2.1.3 未分組的 profile 放在「未分組」區塊
- 6.2.2.2 搜尋功能
  - 6.2.2.2.1 即時搜尋 profile 名稱、分組、標籤
  - 6.2.2.2.2 搜尋結果高亮匹配文字

#### 6.2.3 Profile 操作
- 6.2.3.1 點擊 profile → 開啟 Container Tab
  - 6.2.3.1.1 已開啟的 profile 點擊 → 切換到該 tab
- 6.2.3.2 右鍵選單
  - 6.2.3.2.1 編輯、複製、刪除、開啟/關閉、查看指紋
- 6.2.3.3 底部按鈕列
  - 6.2.3.3.1 [+ 新增 Profile] → 開啟新增表單
  - 6.2.3.3.2 [⚙ 設定] → 開啟設定頁面

### 6.3 Profile 操作介面
#### 6.3.1 新增/編輯 Profile 表單
- 6.3.1.1 表單欄位
  - 6.3.1.1.1 名稱（必填）、分組（下拉 + 自訂）、標籤（多選）
  - 6.3.1.1.2 Proxy 設定（type 下拉、host、port、username、password）
  - 6.3.1.1.3 指紋設定：「自動生成」或「手動設定」切換
  - 6.3.1.1.4 手動指紋：展開所有欄位供進階使用者調整

#### 6.3.2 指紋預覽面板
- 6.3.2.1 顯示當前 profile 的指紋摘要
  - 6.3.2.1.1 UA、Platform、Screen、Timezone、WebGL renderer
  - 6.3.2.1.2 「重新生成指紋」按鈕
  - 6.3.2.1.3 「測試指紋」按鈕 → 開啟 browserleaks.com 在該 container 中

### 6.4 狀態監控與通知
#### 6.4.1 人工介入通知
- 6.4.1.1 偵測需要人工介入的情況
  - 6.4.1.1.1 監聽 tab title 變化，偵測關鍵字（「驗證」「Verify」「Checkpoint」「CAPTCHA」）
  - 6.4.1.1.2 偵測到時在 sidebar 該 profile 旁顯示 🔴 警示
  - 6.4.1.1.3 可選：系統通知（browser.notifications.create）

#### 6.4.2 Proxy 狀態顯示
- 6.4.2.1 每個 profile 旁顯示 proxy 狀態
  - 6.4.2.1.1 🟢 正常 / 🟡 延遲高 / 🔴 斷線 / ⚪ 無 proxy
  - 6.4.2.1.2 hover 顯示 proxy IP 和延遲

#### 6.4.3 記憶體使用監控
- 6.4.3.1 顯示整體記憶體使用量
  - 6.4.3.1.1 從 Control Server /api/status 取得 memory_usage
  - 6.4.3.1.2 sidebar 底部顯示：「12 profiles | 1.8 GB RAM」

### 6.5 Cookie 管理介面
#### 6.5.1 Cookie 匯出
- 6.5.1.1 匯出指定 container 的所有 cookie 為 JSON
  - 6.5.1.1.1 browser.cookies.getAll({ storeId: cookieStoreId })
  - 6.5.1.1.2 格式：[{ name, value, domain, path, secure, httpOnly, sameSite, expirationDate }]
  - 6.5.1.1.3 下載為 {profile_name}_cookies.json

#### 6.5.2 Cookie 匯入
- 6.5.2.1 從 JSON 檔案匯入 cookie 到指定 container
  - 6.5.2.1.1 讀取 JSON → 逐一 browser.cookies.set({ storeId, ... })
  - 6.5.2.1.2 支援從其他反偵測瀏覽器匯出的格式（Netscape、JSON 通用格式）

---

## 7. Control Server

### 7.1 HTTP Server 骨架
#### 7.1.1 Go 專案結構
- 7.1.1.1 建立 Go module
  - 7.1.1.1.1 go mod init camoufoxmulti/control-server
  - 7.1.1.1.2 目錄結構：cmd/server/main.go、internal/api/、internal/profile/、internal/session/

#### 7.1.2 HTTP Router
- 7.1.2.1 使用 net/http + 輕量 router（chi 或 stdlib ServeMux）
  - 7.1.2.1.1 路由註冊：/api/profiles、/api/sessions、/api/status 等
  - 7.1.2.1.2 中間件：CORS（localhost only）、request logging、error recovery

#### 7.1.3 啟動與設定
- 7.1.3.1 從 config.json 讀取 port 和其他設定
  - 7.1.3.1.1 命令列參數覆蓋：--port、--config
  - 7.1.3.1.2 預設 port 19280
- 7.1.3.2 Health check endpoint
  - 7.1.3.2.1 GET /api/status → { version, uptime, active_sessions, memory_usage }

### 7.2 Profile CRUD API
#### 7.2.1 API 實作
- 7.2.1.1 POST /api/profiles — 建立
  - 7.2.1.1.1 解析 request body → 驗證必要欄位 → 生成 ID → 生成指紋 → 儲存 → 回傳
- 7.2.1.2 GET /api/profiles — 列表
  - 7.2.1.2.1 支援 query params: group、tag、sort、limit、offset
- 7.2.1.3 GET /api/profiles/:id — 詳情
  - 7.2.1.3.1 找不到回傳 404
- 7.2.1.4 PUT /api/profiles/:id — 更新
  - 7.2.1.4.1 部分更新，只覆蓋提供的欄位
- 7.2.1.5 DELETE /api/profiles/:id — 刪除
  - 7.2.1.5.1 先關閉活躍 session → 刪除檔案 → 通知 extension 刪除 container
- 7.2.1.6 POST /api/profiles/:id/duplicate — 複製
  - 7.2.1.6.1 深拷貝 + 新 ID + 新指紋

### 7.3 Session 管理 API
#### 7.3.1 Session 生命週期
- 7.3.1.1 POST /api/sessions — 開啟 session
  - 7.3.1.1.1 接收 profile_id → 通知 extension 開啟 container tab → 回傳 session_id、tab_id
  - 7.3.1.1.2 若該 profile 已有活躍 session，回傳現有 session（不重複開啟）
- 7.3.1.2 GET /api/sessions — 列出活躍 session
  - 7.3.1.2.1 回傳所有活躍 session 的 ID、profile_id、tab_id、開啟時間
- 7.3.1.3 DELETE /api/sessions/:id — 關閉 session
  - 7.3.1.3.1 通知 extension 關閉對應 tab
- 7.3.1.4 DELETE /api/sessions — 關閉所有
  - 7.3.1.4.1 逐一關閉所有活躍 session

### 7.4 瀏覽器操作 API（透過 Playwright）
#### 7.4.1 Playwright 連線管理
- 7.4.1.1 使用 playwright-go（版本依 spike 結果決定，見 §13.2）
  - 7.4.1.1.1 優先：`go get github.com/playwright-community/playwright-go@v0.5700.1`（最新）
  - 7.4.1.1.2 退回：`go get github.com/playwright-community/playwright-go@v0.5101.0`（匹配 Camoufox v135）
  - 7.4.1.1.2 啟動方式 A（推薦）：直接 launch
    ```go
    browser, _ := pw.Firefox.Launch(playwright.BrowserTypeLaunchOptions{
        ExecutablePath: playwright.String("./camoufox/firefox"),
        Env: map[string]string{
            "CAMOU_CONFIG_1": camoufoxConfigJSON,
        },
        FirefoxUserPrefs: map[string]any{
            "privacy.userContext.enabled": true,
            "media.peerconnection.enabled": false,
        },
    })
    ```
  - 7.4.1.1.3 啟動方式 B（備選）：連接 camoufox server
    ```go
    // 先啟動: python -m camoufox server --port 19281
    browser, _ := pw.Firefox.Connect("ws://localhost:19281")
    ```
  - 7.4.1.1.4 注意：playwright-go 內部會啟動一個 Node.js driver process（~50MB），這是架構限制
- 7.4.1.2 維護 session_id → Playwright Page 的映射
  - 7.4.1.2.1 WebExtension 開啟 container tab 後通知 server（tab URL/title）
  - 7.4.1.2.2 Server 透過 `browser.Contexts()[0].Pages()` 找到對應 page
  - 7.4.1.2.3 建立 session_id → page 映射，後續操作直接用 page handle

#### 7.4.2 頁面操作（REST API → Playwright）
- 7.4.2.1 POST /api/sessions/:id/navigate
  - 7.4.2.1.1 `page.goto(url, { waitUntil })` — 直接用 Playwright
- 7.4.2.2 POST /api/sessions/:id/click
  - 7.4.2.2.1 `page.click(selector)` — Playwright 原生支援 auto-wait
- 7.4.2.3 POST /api/sessions/:id/type
  - 7.4.2.3.1 `page.fill(selector, text)` 或 `page.type(selector, text, { delay })`
- 7.4.2.4 POST /api/sessions/:id/eval
  - 7.4.2.4.1 `page.evaluate(script)` — 回傳 JSON 結果
- 7.4.2.5 GET /api/sessions/:id/screenshot
  - 7.4.2.5.1 `page.screenshot({ fullPage })` — 回傳 PNG binary
- 7.4.2.6 GET /api/sessions/:id/content
  - 7.4.2.6.1 `page.content()` 或 `page.locator(selector).textContent()`
- 7.4.2.7 POST /api/sessions/:id/wait
  - 7.4.2.7.1 `page.waitForSelector(selector, { timeout })`

#### 7.4.3 Cookie 操作（透過 WebExtension）
- 7.4.3.1 GET /api/sessions/:id/cookies
  - 7.4.3.1.1 透過 WebExtension 取得該 container 的 cookie（Playwright 無法區分 container cookie）
- 7.4.3.2 POST /api/sessions/:id/cookies
  - 7.4.3.2.1 透過 WebExtension 匯入 cookie 到指定 container

### 7.5 Playwright Endpoint 暴露
#### 7.5.1 外部 Playwright 直連
- 7.5.1.1 Control Server 啟動 Camoufox 時用 Playwright server 模式
  - 7.5.1.1.1 `camoufox server` 或 Playwright `launch_server()` 暴露 WebSocket endpoint
  - 7.5.1.1.2 GET /api/playwright/endpoint → `{ ws_url: "ws://localhost:19281" }`
  - 7.5.1.1.3 外部 Playwright client 可直連，獲得完整控制能力

#### 7.5.2 Session-Page 映射 API
- 7.5.2.1 GET /api/sessions/:id/page-info
  - 7.5.2.1.1 回傳該 session 對應的 page index，讓外部 client 能找到正確的 page

### 7.6 與 WebExtension 通訊
#### 7.6.1 通訊協議
- 7.6.1.1 WebExtension → Control Server: HTTP fetch（localhost）
  - 7.6.1.1.1 extension background script 用 fetch() 呼叫 REST API
- 7.6.1.2 Control Server → WebExtension: WebSocket 推送
  - 7.6.1.2.1 extension 啟動時建立 WebSocket 連線到 server
  - 7.6.1.2.2 server 推送 container 管理指令（開 tab、關 tab、更新 profile 設定）
  - 7.6.1.2.3 **不再推送瀏覽器操作指令**（navigate/click/type 等走 Playwright）

---

## 8. 自動化整合

### 8.1 Playwright 整合（已內建於 §7.4-7.5）
#### 8.1.1 使用範例文件
- 8.1.1.1 撰寫 Python / Node.js 使用範例
  - 8.1.1.1.1 Python: 透過 REST API 建立 profile + 開 session + 操作
  - 8.1.1.1.2 Python: 透過 Playwright 直連操作特定 profile 的 tab
  - 8.1.1.1.3 Node.js: 同上

### 8.2 REST API 瀏覽器操作
#### 8.2.1 簡單操作封裝
- 8.2.1.1 REST API 已透過 Playwright 提供完整操作能力（§7.4）
  - 8.2.1.1.1 navigate、click、type、screenshot、eval 覆蓋 80% 自動化場景
  - 8.2.1.1.2 複雜場景引導使用者用 Playwright 直連

### 8.3 Phase 3: MCP Server Adapter
#### 8.3.1 MCP 協議實作
- 8.3.1.1 實作 MCP Server（Go 或 Node.js）
  - 8.3.1.1.1 定義 tools：open_profile、navigate、click、type、screenshot、get_content
  - 8.3.1.1.2 定義 resources：profile 列表、session 狀態
  - 8.3.1.1.3 MCP Server 內部呼叫 Control Server REST API

### 8.4 Phase 3: YAML Workflow 引擎
#### 8.4.1 Workflow 定義格式
- 8.4.1.1 設計 YAML workflow schema
  - 8.4.1.1.1 steps 陣列，每步定義 action + target profile + 參數
  - 8.4.1.1.2 支援條件分支、迴圈、等待
  - 8.4.1.1.3 支援排程（cron 語法）

#### 8.4.2 執行引擎
- 8.4.2.1 Go 層解析 YAML → 逐步執行
  - 8.4.2.1.1 每步呼叫對應的 REST API
  - 8.4.2.1.2 錯誤處理：重試 N 次 → 跳過 or 停止
  - 8.4.2.1.3 執行日誌記錄

---

## 8B. Chromium 引擎整合（Phase 3）

> 雙引擎架構：Firefox（Camoufox, container 方案）+ Chromium（CloakBrowser → 後期可換自己的 fork）
> 設計原則：Control Server 是引擎無關的，Playwright API 抽象了引擎差異，MCP/REST 呼叫端不需要知道底層引擎。

### 8B.1 Chromium 引擎選型
#### 8B.1.1 Phase 3 初期：CloakBrowser（現成 binary）
- 8B.1.1.1 CloakBrowser 整合
  - 8B.1.1.1.1 49+ C++ 層 patch，reCAPTCHA v3 得分 0.9，通過 Cloudflare Turnstile
  - 8B.1.1.1.2 指紋配置：`--fingerprint=<seed>`（deterministic，同 seed 同指紋）
  - 8B.1.1.1.3 Proxy：`--proxy-server=http://127.0.0.1:{port}`
  - 8B.1.1.1.4 隔離：`--user-data-dir=/profiles/{profile_id}/chromium-data`
  - 8B.1.1.1.5 授權：wrapper MIT，binary 免費使用但不可重新發佈 → 使用者自行下載
- 8B.1.1.2 Binary 管理
  - 8B.1.1.2.1 首次啟動偵測 CloakBrowser 是否已安裝
  - 8B.1.1.2.2 未安裝時引導使用者下載（提供下載連結 + 安裝路徑設定）
  - 8B.1.1.2.3 版本記錄在 `.cloakbrowser-version`，打包腳本不含 binary

#### 8B.1.2 後期可選：自己的 Chromium fork
- 8B.1.2.1 參考 Camoufox patch 模式，AI agent 輔助建立 Chromium 等價 patch set
  - 8B.1.2.1.1 攔截點對照：
    ```
    Camoufox (Gecko)                    → Chromium (Blink/V8)
    dom/base/Navigator.cpp              → blink/renderer/core/frame/navigator.cc
    dom/base/Screen.cpp                 → blink/renderer/core/frame/screen.cc
    dom/canvas/                         → blink/renderer/modules/canvas/
    dom/canvas/WebGLContext*             → gpu/command_buffer/service/
    dom/media/webaudio/                 → blink/renderer/modules/webaudio/
    js/src/jsdate.cpp (SpiderMonkey)    → v8/src/date/date.cc (V8)
    netwerk/protocol/http/ (UA)         → net/http/http_util.cc
    ```
  - 8B.1.2.1.2 保持與 CloakBrowser 相同的 CLI 參數介面 → Control Server 不需改動
  - 8B.1.2.1.3 觸發條件：CloakBrowser 停止維護、改授權、或功能不足時

### 8B.2 Chromium Profile 生命週期
#### 8B.2.1 建立 Chromium Profile
- 8B.2.1.1 建立 profile 時指定 engine: "chromium"
  - 8B.2.1.1.1 從 Chromium 指紋池取一組指紋（fingerprint-suite 生成 `browsers: ['chrome']`）
  - 8B.2.1.1.2 生成 fingerprint seed（uint32）→ 存入 profile.json
  - 8B.2.1.1.3 建立 `profiles/{profile_id}/chromium-data/` 目錄（Chromium user-data-dir）
  - 8B.2.1.1.4 GeoIP 調配地理欄位（同 §3.2.3，引擎無關）

#### 8B.2.2 啟動 Chromium Profile
- 8B.2.2.1 Go Control Server 啟動 CloakBrowser process
  - 8B.2.2.1.1 啟動 local proxy process（per-profile，同 Firefox 的 proxy 基礎設施）
  - 8B.2.2.1.2 組裝啟動參數：
    ```
    ./cloakbrowser
      --fingerprint=<seed>
      --fingerprint-platform=<windows|macos|linux>
      --user-data-dir=./profiles/{id}/chromium-data
      --proxy-server=http://127.0.0.1:{local_proxy_port}
      --remote-debugging-port=<random_port>
    ```
  - 8B.2.2.1.3 記錄 PID + CDP port → 存入 session 狀態
  - 8B.2.2.1.4 Playwright `pw.Chromium.ConnectOverCDP("http://localhost:{cdp_port}")` 取得 browser handle
  - 8B.2.2.1.5 建立 session_id → Playwright page 映射

#### 8B.2.3 關閉 Chromium Profile
- 8B.2.3.1 Playwright `browser.Close()` 或 kill process by PID
  - 8B.2.3.1.1 清理 session 狀態
  - 8B.2.3.1.2 停止 local proxy process
  - 8B.2.3.1.3 Chromium user-data-dir 保留（cookie/storage 持久化）

### 8B.3 雙引擎統一 API
#### 8B.3.1 REST API 引擎無關
- 8B.3.1.1 所有 API endpoint 不區分引擎
  - 8B.3.1.1.1 `POST /api/profiles` body 加 `engine` 欄位：`"firefox"` | `"chromium"`
  - 8B.3.1.1.2 `POST /api/sessions` → server 內部根據 profile.engine 分派到對應 launcher
  - 8B.3.1.1.3 `POST /api/sessions/:id/navigate` → 統一走 Playwright `page.Goto()`，引擎無關
  - 8B.3.1.1.4 `GET /api/sessions/:id/screenshot` → 統一走 Playwright `page.Screenshot()`

#### 8B.3.2 MCP tools 引擎無關
- 8B.3.2.1 MCP tool 定義不包含引擎參數
  - 8B.3.2.1.1 `open_profile { profile_id }` → server 自動判斷引擎
  - 8B.3.2.1.2 `navigate { profile_id, url }` → Playwright 抽象層處理
  - 8B.3.2.1.3 AI agent 不需要知道底層是 Firefox 還是 Chromium

#### 8B.3.3 引擎特有操作（少數例外）
- 8B.3.3.1 Cookie 操作的引擎差異
  - 8B.3.3.1.1 Firefox：透過 WebExtension `browser.cookies` API（per-container）
  - 8B.3.3.1.2 Chromium：透過 Playwright `context.cookies()` / `context.addCookies()`
  - 8B.3.3.1.3 REST API 統一介面，server 內部分派

### 8B.4 Sidebar UI 雙引擎支援
#### 8B.4.1 Profile 列表顯示
- 8B.4.1.1 每個 profile 旁顯示引擎 badge
  - 8B.4.1.1.1 🦊 Firefox | 🌐 Chromium
  - 8B.4.1.1.2 新增 Profile 時選擇引擎（下拉選單）
- 8B.4.1.2 Chromium profile 的狀態顯示
  - 8B.4.1.2.1 🟢 已開啟（獨立視窗）/ ⚪ 未開啟
  - 8B.4.1.2.2 點擊已開啟的 Chromium profile → 聚焦到該視窗（OS 層 window focus）

### 8B.5 指紋池雙引擎支援
#### 8B.5.1 分引擎生成指紋
- 8B.5.1.1 fingerprint-suite 分別生成 Firefox 和 Chrome 指紋
  - 8B.5.1.1.1 `--browser firefox --count 500` → `data/fingerprints-firefox-*.json`
  - 8B.5.1.1.2 `--browser chrome --count 500` → `data/fingerprints-chrome-*.json`
  - 8B.5.1.1.3 建立 profile 時根據 engine 從對應池取用

---

## 9. 打包與發佈

### 9.1 跨平台啟動腳本
#### 9.1.1 start.sh（Linux / macOS）
- 9.1.1.1 啟動 Control Server + Camoufox
  - 9.1.1.1.1 偵測 Camoufox binary 路徑（同目錄下）
  - 9.1.1.1.2 背景啟動 control-server，記錄 PID
  - 9.1.1.1.3 前景啟動 camoufox（使用者關閉瀏覽器 = 結束流程）
  - 9.1.1.1.4 Camoufox 退出後自動 kill control-server
  - 9.1.1.1.5 macOS: 處理 Gatekeeper 安全提示（xattr -cr）

#### 9.1.2 start.bat（Windows）
- 9.1.2.1 同等邏輯的 Windows 版本
  - 9.1.2.1.1 start /B control-server.exe 背景啟動
  - 9.1.2.1.2 前景啟動 camoufox.exe
  - 9.1.2.1.3 退出後 taskkill control-server.exe

### 9.2 Portable ZIP 打包
#### 9.2.1 打包腳本
- 9.2.1.1 自動化打包流程（Makefile 或 shell script）
  - 9.2.1.1.1 下載對應平台的 Camoufox release binary
  - 9.2.1.1.2 交叉編譯 Go control-server（GOOS=windows/darwin/linux）
  - 9.2.1.1.3 複製 extension/ 目錄
  - 9.2.1.1.4 建立空的 profiles/、data/ 目錄
  - 9.2.1.1.5 生成預設 config.json
  - 9.2.1.1.6 打包為 BrowseForge-v{version}-{platform}.zip

#### 9.2.2 檔案大小優化
- 9.2.2.1 Go binary strip symbols（-ldflags "-s -w"）
  - 9.2.2.1.1 預估：~15MB → ~10MB
- 9.2.2.2 UPX 壓縮（可選）
  - 9.2.2.2.1 進一步壓縮到 ~5MB，但啟動速度略慢

### 9.3 版本管理與更新機制
#### 9.3.1 版本號策略
- 9.3.1.1 SemVer（major.minor.patch）
  - 9.3.1.1.1 version 寫入 config.json 和 /api/status 回傳
  - 9.3.1.1.2 extension manifest.json 版本同步

#### 9.3.2 更新檢查（Phase 2）
- 9.3.2.1 啟動時檢查 GitHub Releases 最新版本
  - 9.3.2.1.1 GET https://api.github.com/repos/.../releases/latest
  - 9.3.2.1.2 有新版本時在 sidebar 顯示更新提示
  - 9.3.2.1.3 不自動更新，提供下載連結讓使用者手動替換

### 9.4 首次啟動初始化流程
#### 9.4.1 初始化檢測
- 9.4.1.1 偵測是否為首次啟動（data/ 目錄是否為空）
  - 9.4.1.1.1 首次啟動：初始化 Firefox profile、建立預設 config、顯示歡迎頁
  - 9.4.1.1.2 非首次：正常啟動流程

#### 9.4.2 歡迎頁面
- 9.4.2.1 首次啟動開啟內建歡迎頁
  - 9.4.2.1.1 簡要說明使用方式
  - 9.4.2.1.2 引導建立第一個 profile

---

## 10. 錯誤處理與日誌系統

### 10.1 統一日誌框架
#### 10.1.1 Control Server 日誌（Go）
- 10.1.1.1 結構化 JSON 日誌
  - 10.1.1.1.1 使用 slog（Go 1.21 標準庫），輸出 JSON 格式
  - 10.1.1.1.2 每筆日誌包含：timestamp、level、msg、request_id、component
  - 10.1.1.1.3 日誌等級：DEBUG / INFO / WARN / ERROR
- 10.1.1.2 日誌輸出目標
  - 10.1.1.2.1 stdout（預設，方便開發）
  - 10.1.1.2.2 檔案輸出：logs/control-server.log（config.json 可設定）
  - 10.1.1.2.3 日誌輪替：按大小（10MB）或按天輪替，保留最近 7 天

#### 10.1.2 WebExtension 日誌
- 10.1.2.1 background script 日誌收集
  - 10.1.2.1.1 封裝 logger 模組：log.info() / log.warn() / log.error()
  - 10.1.2.1.2 日誌寫入 browser.storage.local（環形緩衝區，最多 1000 筆）
  - 10.1.2.1.3 錯誤日誌同步推送到 Control Server（POST /api/logs）
- 10.1.2.2 sidebar 錯誤顯示
  - 10.1.2.2.1 sidebar 底部顯示最近一筆錯誤訊息（紅色 bar，可展開）
  - 10.1.2.2.2 點擊展開完整錯誤日誌列表

#### 10.1.3 HTTP Request 日誌
- 10.1.3.1 中間件記錄每個 API 請求
  - 10.1.3.1.1 記錄：method、path、status_code、duration_ms、request_id
  - 10.1.3.1.2 錯誤回應（4xx/5xx）記錄 request body 摘要（截斷至 500 字元）

### 10.2 異常恢復機制
#### 10.2.1 啟動時一致性修復
- 10.2.1.1 偵測上次是否異常關閉
  - 10.2.1.1.1 PID file 存在但 process 不存在 → 判定為異常關閉
  - 10.2.1.1.2 清理殘留 PID file
- 10.2.1.2 Container 狀態修復
  - 10.2.1.2.1 列出所有 container → 比對 profiles/ 目錄 → 刪除孤兒 container
  - 10.2.1.2.2 列出所有 profile → 比對 container 列表 → 為缺少 container 的 profile 重建
- 10.2.1.3 Session 狀態重置
  - 10.2.1.3.1 啟動時清空所有 session 記錄（上次的 tab 已不存在）

#### 10.2.2 運行時錯誤處理
- 10.2.2.1 Control Server panic recovery
  - 10.2.2.1.1 HTTP handler 層 recover middleware，捕獲 panic → 回傳 500 → 記錄 stack trace
- 10.2.2.2 WebExtension ↔ Server 斷線處理
  - 10.2.2.2.1 WebSocket 斷線後自動重連（指數退避：1s → 2s → 4s → 最大 30s）
  - 10.2.2.2.2 重連期間 sidebar 顯示「連線中...」狀態
  - 10.2.2.2.3 重連成功後重新同步 profile 列表和 session 狀態

#### 10.2.3 Profile 資料損壞處理
- 10.2.3.1 JSON parse 失敗時的 fallback
  - 10.2.3.1.1 嘗試讀取 profile.json.bak（每次成功寫入後保留上一版）
  - 10.2.3.1.2 backup 也失敗 → 記錄 error log → 跳過該 profile → sidebar 顯示損壞標記
  - 10.2.3.1.3 提供手動修復入口：sidebar 右鍵「修復 profile」→ 重新生成預設值

---

## 11. 安全機制

### 11.1 API 認證
#### 11.1.1 Token 認證
- 11.1.1.1 啟動時生成隨機 API token
  - 11.1.1.1.1 crypto/rand 生成 32 bytes → hex encode → 64 字元 token
  - 11.1.1.1.2 寫入 data/.api-token（檔案權限 0600）
  - 11.1.1.1.3 WebExtension 啟動時讀取 token（透過 fetch /api/token-file 或預設路徑）
- 11.1.1.2 請求驗證
  - 11.1.1.2.1 所有 API 請求需帶 `Authorization: Bearer {token}` header
  - 11.1.1.2.2 /api/status 免認證（health check 用）
  - 11.1.1.2.3 認證失敗回傳 401

#### 11.1.2 監聽範圍限制
- 11.1.2.1 預設只監聽 127.0.0.1（不監聽 0.0.0.0）
  - 11.1.2.1.1 config.json 可設定 bind_address，但預設最安全
  - 11.1.2.1.2 若設定為 0.0.0.0，啟動時顯示安全警告

### 11.2 敏感資料保護
#### 11.2.1 記憶體中的敏感資料
- 11.2.1.1 Proxy 密碼在 API 回傳時遮罩
  - 11.2.1.1.1 GET /api/profiles 回傳 proxy.password = "***"（不回傳明文）
  - 11.2.1.1.2 只有建立/更新時接受明文密碼寫入

#### 11.2.2 日誌中的敏感資料
- 11.2.2.1 日誌自動過濾敏感欄位
  - 11.2.2.1.1 password、token、cookie value 在日誌中替換為 [REDACTED]

---

## 12. 記憶體管理

### 12.1 Tab Suspend 機制
#### 12.1.1 自動 Suspend
- 12.1.1.1 非活躍 tab 自動卸載
  - 12.1.1.1.1 使用 browser.tabs.discard(tabId) 卸載 tab 內容（保留 tab 標籤）
  - 12.1.1.1.2 觸發條件：tab 超過 N 分鐘未被切換到（N 在 config.json 設定，預設 15 分鐘）
  - 12.1.1.1.3 排除條件：正在播放音訊/視訊的 tab 不自動 suspend
- 12.1.1.2 Suspend 狀態顯示
  - 12.1.1.2.1 sidebar 中 suspended tab 顯示 💤 圖示
  - 12.1.1.2.2 點擊 suspended profile → 自動 reload tab

#### 12.1.2 手動 Suspend
- 12.1.2.1 sidebar 右鍵選單「暫停此 profile」
  - 12.1.2.1.1 立即 discard 該 tab
- 12.1.2.2 批次操作「暫停所有非活躍 profile」
  - 12.1.2.2.1 除了當前 active tab 外，全部 discard

### 12.2 記憶體監控
#### 12.2.1 系統記憶體查詢
- 12.2.1.1 Go server 定期查詢 Camoufox process 記憶體用量
  - 12.2.1.1.1 Linux/macOS: 讀取 /proc/{pid}/status 或 ps 命令
  - 12.2.1.1.2 Windows: 用 syscall 查詢 process memory info
  - 12.2.1.1.3 每 30 秒更新一次，寫入 /api/status 回傳

#### 12.2.2 記憶體警告
- 12.2.2.1 超過閾值時觸發警告
  - 12.2.2.1.1 閾值：系統可用記憶體 < 1GB 或 Camoufox 佔用 > config 設定上限
  - 12.2.2.1.2 sidebar 顯示記憶體警告 banner
  - 12.2.2.1.3 建議使用者 suspend 不活躍的 profile

---

## 13. 開發環境與前置條件

### 13.1 開發環境搭建
#### 13.1.1 必要工具安裝
- 13.1.1.1 Go 1.22+（playwright-go 要求）
  - 13.1.1.1.1 macOS: `brew install go`（當前版本 1.26.2）
  - 13.1.1.1.2 驗證：`go version`
- 13.1.1.2 Node.js 22 LTS（fingerprint-suite、web-ext）
  - 13.1.1.2.1 macOS: `brew install node`（當前 LTS 22.22.2）
  - 13.1.1.2.2 安裝 web-ext：`npm install -g web-ext`（extension 開發/測試用）
- 13.1.1.3 Camoufox binary
  - 13.1.1.3.1 從 GitHub Releases 下載對應平台的 Camoufox
  - 13.1.1.3.2 解壓到專案根目錄的 camoufox/ 子目錄
  - 13.1.1.3.3 記錄目標版本號到 .camoufox-version 檔案

#### 13.1.2 開發模式啟動
- 13.1.2.1 Control Server 開發模式
  - 13.1.2.1.1 `go run cmd/server/main.go --config ./config.dev.json`
  - 13.1.2.1.2 可選：用 air 做 hot reload（`go install github.com/air-verse/air@latest`）
- 13.1.2.2 WebExtension 開發模式
  - 13.1.2.2.1 `web-ext run --source-dir ./extension --firefox ./camoufox/camoufox --firefox-profile ./data`
  - 13.1.2.2.2 web-ext 會自動 reload extension 當檔案變更
- 13.1.2.3 一鍵開發啟動腳本
  - 13.1.2.3.1 `make dev` 或 `./dev.sh`：同時啟動 Go server（hot reload）+ Camoufox（web-ext）

### 13.2 版本相容性矩陣（2026-04 驗證）

> ⚠️ 以下版本經過交叉驗證，確認相互相容。變更任何一項前必須重新驗證相容性。

#### 13.2.1 鎖定版本

| 技術 | 優先版本（先測試） | 退回版本（若不相容） | 關鍵約束 |
|------|------------------|-------------------|---------|
| **Camoufox** | FF146-pre（v146-hardware, 2026-03） | v135.0.1-beta.24（2025-03） | 需 spike 驗證 Playwright 協議相容性 |
| **playwright-go** | v0.5700.1（PW 1.57, 2026-02） | v0.5101.0（PW 1.51, 2025） | 必須匹配 Camoufox 的 Playwright 支援 |
| **Go** | 1.26.2（2026-04） | 1.22+（playwright-go 最低要求） | |
| **Node.js** | 22 LTS（22.22.2） | | fingerprint-suite、web-ext |
| **fingerprint-suite** | 2.1.82（2026-04） | | 每月自動更新指紋模型 |
| **web-ext** | 9.3.0 | | Extension 開發/測試 |
| **MaxMind GeoLite2** | 最新 DB | | 免費，需註冊帳號 |

> 版本策略：**優先測試最新版組合，spike 階段驗證相容性，不行再退回已知相容的舊版。**

#### 13.2.2 Spike 版本驗證（加入 Phase 0）

在 spike 0.1.4（Playwright 連接驗證）中同時測試兩組版本：
1. **最新組合**：Camoufox FF146-pre + playwright-go v0.5700.1 → 能連入並操作？
2. **保守組合**：Camoufox v135 + playwright-go v0.5101.0 → 作為 fallback 確認可用
3. 記錄結果到 spike-results.md，選擇通過的最新版本組合

#### 13.2.3 Phase 2 版本升級路徑

Fork Camoufox 後可自行更新 Playwright 支援：
1. 從 `microsoft/playwright` 取最新 `browser_patches/firefox/patches/bootstrap.diff`
2. 替換 `patches/playwright/0-playwright.patch`
3. 重新 build → 測試 → 即可支援最新 Playwright 版本
4. 這是 Camoufox 維護者的標準升級流程，不需要額外發明

#### 13.2.4 ⚠️ 版本相容性注意事項

**playwright-go 版本必須匹配 Camoufox 的 Playwright 協議：**
- Camoufox 的 Playwright 支援來自 `patches/playwright/0-playwright.patch`（源自 microsoft/playwright 的 Firefox patches）
- 不同 Playwright 版本的 Juggler 協議可能有差異
- Spike 階段必須實測確認，不能只看版本號

**Camoufox 基於 Firefox 135/146（非 ESR）：**
- 所有需要的 WebExtension API 在 Firefox 128+ 都可用，135/146 完全相容
- Phase 2 fork 後可自行選擇基於哪個 Firefox 版本

#### 13.2.3 WebExtension API 相容性確認（Firefox 135）

| API | 可用？ | MV2 支援 | 備註 |
|-----|:---:|:---:|------|
| contextualIdentities | ✅ | ✅ | Firefox 57+，無棄用計劃 |
| proxy.onRequest + cookieStoreId | ✅ | ✅ | cookieStoreId 確認在 RequestDetails 中 |
| content_scripts world: "MAIN" | ✅ | ✅ | Firefox 128+ 新增，Mozilla 明確回移到 MV2 |
| wrappedJSObject | ✅ | ✅ | Firefox 核心安全架構，無棄用計劃 |
| tabs.discard() | ✅ | ✅ | 無變更 |
| scripting.executeScript world: "MAIN" | ✅ | ✅ | Firefox 102+ MV2 可用，128+ 支援 MAIN world |
| sidebar_action | ✅ | ✅ | Firefox 專有，無棄用計劃 |
| browsingData.removeCookies + cookieStoreId | ✅ | ✅ | 可清除特定 container 的 cookie + localStorage |

**Manifest V2 狀態：Mozilla 明確承諾不棄用 MV2，無 sunset 時間表。**
- 2025-02 Mozilla 再次重申不會移除 MV2 支援
- Chrome 已於 2025-06 完全移除 MV2，Firefox 明確差異化
- BrowseForge 使用 MV2 是安全的長期選擇

### 13.3 Camoufox 版本管理
#### 13.3.1 版本鎖定
- 13.3.1.1 專案根目錄 `.camoufox-version` 記錄目標版本
  - 13.3.1.1.1 格式：`v135.0.1-beta.24`
  - 13.3.1.1.2 打包腳本和 CI 從此檔案讀取版本號
- 13.3.1.2 版本升級流程
  - 13.3.1.2.1 更新 .camoufox-version → 下載新版 → 確認 Playwright 協議版本 → 必要時更新 playwright-go 版本 → 跑完整測試 → 確認無 regression

#### 13.3.2 Phase 2 Build 環境（Fork 後才需要）
- 13.3.2.1 Gecko build 依賴
  - 13.3.2.1.1 Rust toolchain（rustup）
  - 13.3.2.1.2 C++ compiler（clang / gcc）
  - 13.3.2.1.3 Python 3.x（mach build system）
  - 13.3.2.1.4 首次 build 預估時間：30-60 分鐘（增量 build 5-10 分鐘）

### 13.3 GeoIP 資料庫
#### 13.3.1 資料來源
- 13.3.1.1 MaxMind GeoLite2-City 免費版
  - 13.3.1.1.1 註冊 MaxMind 帳號取得 license key
  - 13.3.1.1.2 下載 GeoLite2-City.mmdb（~70MB）
  - 13.3.1.1.3 Go 用 oschwald/maxminddb-golang 讀取
- 13.3.1.2 輕量替代方案
  - 13.3.1.2.1 內建簡化版 IP-country-timezone JSON 映射表（~500KB）
  - 13.3.1.2.2 覆蓋主要 proxy 供應商的 IP 段（US、UK、DE、JP、KR、SG 等）
  - 13.3.1.2.3 Phase 1 用簡化版，Phase 2 升級到 MaxMind

#### 13.3.2 更新機制
- 13.3.2.1 GeoIP 資料庫定期更新
  - 13.3.2.1.1 MaxMind 每週更新一次
  - 13.3.2.1.2 打包時嵌入最新版，使用者可手動替換

---

## 14. Profile 備份與還原

### 14.1 單一 Profile 匯出/匯入
#### 14.1.1 匯出
- 14.1.1.1 打包單一 profile 為 .zip
  - 14.1.1.1.1 包含：profile.json + cookies-backup.json（不含 firefox-data/，太大）
  - 14.1.1.1.2 匯出前自動備份當前 cookie 到 cookies-backup.json
  - 14.1.1.1.3 API: POST /api/profiles/:id/export → 回傳 .zip binary

#### 14.1.2 匯入
- 14.1.2.1 從 .zip 匯入 profile
  - 14.1.2.1.1 解壓 → 生成新 ID（避免衝突）→ 建立 container → 匯入 cookie
  - 14.1.2.1.2 API: POST /api/profiles/import（multipart form-data）

### 14.2 批次備份/還原
#### 14.2.1 全量備份
- 14.2.1.1 打包整個 profiles/ 目錄
  - 14.2.1.1.1 API: POST /api/backup → 生成 BrowseForge-backup-{date}.zip
  - 14.2.1.1.2 包含所有 profile.json + cookies-backup.json + config.json
  - 14.2.1.1.3 不含 firefox-data/（使用者可選擇包含，但檔案會很大）

#### 14.2.2 還原
- 14.2.2.1 從備份 .zip 還原
  - 14.2.2.1.1 API: POST /api/restore（multipart form-data）
  - 14.2.2.1.2 還原前確認：覆蓋現有 profile 或合併（同 ID 的 profile 如何處理）
  - 14.2.2.1.3 還原後重建所有 container 映射

---

## 15. 測試策略

### 15.1 單元測試
#### 15.1.1 Go Control Server
- 15.1.1.1 Profile 模組測試
  - 15.1.1.1.1 CRUD 操作：建立/讀取/更新/刪除 profile JSON
  - 15.1.1.1.2 指紋生成器：驗證生成的指紋符合一致性規則
  - 15.1.1.1.3 ID 生成：唯一性、格式正確性
- 15.1.1.2 API Handler 測試
  - 15.1.1.2.1 用 httptest 模擬 HTTP 請求，驗證回傳格式和狀態碼
  - 15.1.1.2.2 錯誤情境：缺少必要欄位、profile 不存在、重複 ID

#### 15.1.2 WebExtension 純函數測試
- 15.1.2.1 指紋一致性驗證函數
  - 15.1.2.1.1 輸入一組指紋 → 檢查 UA/platform/screen 是否合理配對
- 15.1.2.2 Container-Profile 映射邏輯
  - 15.1.2.2.1 映射建立、查詢、刪除、孤兒清理

### 15.2 整合測試
#### 15.2.1 Container 隔離測試
- 15.2.1.1 Playwright 腳本自動化驗證
  - 15.2.1.1.1 開兩個 container tab → 各自設 cookie → 互相讀取 → 確認隔離
  - 15.2.1.1.2 開兩個 container tab → 各自寫 localStorage → 互相讀取 → 確認隔離
  - 15.2.1.1.3 開兩個 container tab → 各自載入同一圖片 → 確認 HTTP cache 獨立

#### 15.2.2 Proxy 路由測試
- 15.2.2.1 驗證 per-container proxy 正確性
  - 15.2.2.1.1 Container A 設 proxy X → 訪問 httpbin.org/ip → 確認回傳 proxy X 的 IP
  - 15.2.2.1.2 Container B 設 proxy Y → 訪問 httpbin.org/ip → 確認回傳 proxy Y 的 IP
  - 15.2.2.1.3 Container C 無 proxy → 確認走 direct 連線

#### 15.2.3 Server ↔ Extension 通訊測試
- 15.2.3.1 API 呼叫 → Extension 動作 → 結果回傳
  - 15.2.3.1.1 POST /api/sessions → 確認 tab 被開啟
  - 15.2.3.1.2 DELETE /api/sessions/:id → 確認 tab 被關閉
  - 15.2.3.1.3 POST /api/sessions/:id/navigate → 確認 tab URL 變更

### 15.3 E2E 測試
#### 15.3.1 完整流程測試
- 15.3.1.1 從零開始的使用者流程
  - 15.3.1.1.1 啟動系統 → 建立 profile → 開啟 session → 導航到網站 → 截圖 → 關閉 → 刪除
  - 15.3.1.1.2 用 shell script + curl 自動化執行

#### 15.3.2 指紋隔離驗證（Phase 2）
- 15.3.2.1 跨 container 指紋比對
  - 15.3.2.1.1 兩個 container 各自訪問 browserleaks.com → 擷取指紋結果 → 比對差異
  - 15.3.2.1.2 驗證項目：Canvas hash、WebGL renderer、AudioContext hash、timezone、screen、UA
  - 15.3.2.1.3 所有項目必須不同（除非刻意設定相同）

### 15.4 CI 整合
#### 15.4.1 GitHub Actions
- 15.4.1.1 PR 觸發自動測試
  - 15.4.1.1.1 Go 單元測試：`go test ./...`
  - 15.4.1.1.2 Extension lint：`web-ext lint --source-dir ./extension`
  - 15.4.1.1.3 整合測試需要 Camoufox binary → CI 中下載 → headless 模式執行

---

## 完善度審閱

### ✅ 已補入的遺漏項目

| 項目 | 章節 | 狀態 |
|------|------|------|
| 錯誤處理與日誌系統 | §10 | ✅ 已展開至 L5 |
| API 安全認證 | §11 | ✅ 已展開至 L5 |
| Tab 記憶體管理 | §12 | ✅ 已展開至 L5 |
| 開發環境與前置條件 | §13 | ✅ 已展開至 L5 |
| Profile 備份與還原 | §14 | ✅ 已展開至 L5 |
| 測試策略 | §15 | ✅ 已展開至 L5 |

### 🟡 依賴關係（開發順序約束）

```
13. 開發環境 ─────────────────────────────────── 最先（前置條件）
         │
1. 瀏覽器核心整合 ──┐
                    ├→ 2. Container 隔離 ──┐
5. Profile 系統 ────┘                     ├→ 6. WebExtension UI
                                          │
4. Proxy 系統 ────────────────────────────┘
                                          │
3. 指紋系統 (Phase 1) ───────────────────┘
                                          │
10. 錯誤處理/日誌 ──→ 貫穿所有模組       │
11. 安全機制 ────────→ 嵌入 7            │
12. 記憶體管理 ──────→ 嵌入 2 + 6        │
                                          │
7. Control Server ←───────────────────────┘
         │
         ├→ 8. 自動化整合
         ├→ 14. Profile 備份/還原
         │
         └→ 9. 打包與發佈
                  │
                  └→ 15. 測試策略（CI 整合）
```

建議開發順序：

**Phase 1 — Firefox MVP（5-6 週）：**
1. **Week 0（2-3 天）**: Phase 0 全部 — spike 驗證（含 0.1.7/0.1.8/0.1.9）+ 介面契約 + 里程碑定義。**Spike 全部 PASS 才進入 Week 1**
2. **Week 1**: 13（開發環境）+ 1（瀏覽器核心，Playwright 啟動 Camoufox）+ 5（Profile 系統，含 engine 欄位）+ 10.1（日誌）
3. **Week 2**: 2（Container 隔離）+ 4（Proxy 路由）+ 3.2-3.3（fingerprint-suite 指紋池 + setter 注入）+ 11（API 認證）
4. **Week 3**: 7（Control Server API，Playwright 瀏覽器操作）+ 6.1-6.2（Extension 骨架 + Sidebar）+ 12（Tab suspend）
5. **Week 4**: 6.3-6.5（UI 完善）+ 10.2（異常恢復）+ 14（備份/還原）
6. **Week 5**: 9（打包）+ 15（測試）+ 修 bug + 文件

**Phase 2 — C++ 指紋 + VPN（3-4 週）：**
7. **Week 6-7**: Fork Camoufox + per-container C++ 指紋改動（§3.4）
8. **Week 8**: VPN per-profile（wireguard-go local tunnel）+ Profile 加密（§5.5）
9. **Week 9**: 測試 + creepjs 驗證 + 修 bug

**Phase 3 — Chromium 引擎 + 自動化生態（4-5 週）：**
10. **Week 10**: CloakBrowser 整合（§8B.1-8B.2）+ Chromium launcher + 雙引擎 API
11. **Week 11**: Sidebar 雙引擎 UI（§8B.4）+ 指紋池分引擎（§8B.5）
12. **Week 12**: MCP Server（§8.3）+ YAML Workflow（§8.4）
13. **Week 13-14**: E2E 測試 + 文件 + 打包更新

**Phase 4（可選）— 深度自主：**
14. Fork Chromium 取代 CloakBrowser（AI agent 輔助，參考 Camoufox patch 模式）
15. Cloud sync + 團隊協作

### ✅ 測試策略

已獨立為 §15，涵蓋單元測試、整合測試、E2E 測試、CI 整合。

### ✅ Camoufox 版本鎖定

已納入 §13.2，含版本鎖定檔案和升級流程。

### ✅ 覆蓋完整的部分

- 15 個 L1 開發項目，全部展開至 L5
- Profile 資料模型和 CRUD 流程完整
- Container 生命週期管理完整（含 suspend 機制）
- Proxy per-container 路由完整
- Phase 2 C++ 指紋改動路徑清晰（按難度排序）
- Control Server API 設計完整（含認證）
- 啟動/關閉流程有考慮異常情況（含恢復機制）
- 打包和跨平台策略明確
- 錯誤處理、日誌、安全、記憶體管理、測試策略齊備
- 開發環境搭建和前置條件明確
- Profile 備份/還原完整

---

---

## 附錄 B：開源專案深度解析 — 對 BrowseForge 的技術影響

> 以下是 WBS 中提及的所有開源專案的深度分析，提取對開發直接有用的技術細節。

---

### B.1 Camoufox（daijro/camoufox）— 核心引擎

#### 指紋配置：雙層架構

**Layer 1：C++ 層（process-wide，不可變）**
- 透過環境變數 `CAMOU_CONFIG_1`, `CAMOU_CONFIG_2`, ... 傳入 JSON 字串
- Windows 每段 2047 bytes，Linux/Mac 每段 32767 bytes（OS env var 長度限制）
- `MaskConfig.hpp` 用 `std::call_once` 解析，整個 process 生命週期只讀一次
- 所有 C++ 攔截點用 `MaskConfig::GetString("navigator.userAgent")` 查詢

**Layer 2：JS 層（per-context，可變）**
- `NewContext()` 透過 `addInitScript()` 注入 JS，呼叫 Camoufox 暴露的 self-destructing setter：
  ```
  window.setCanvasSeed(seed)
  window.setAudioFingerprintSeed(seed)
  window.setFontSpacingSeed(seed)
  window.setNavigatorPlatform(val)
  window.setNavigatorUserAgent(val)
  window.setWebGLVendor(val) / window.setWebGLRenderer(val)
  window.setScreenDimensions(w, h) / window.setScreenColorDepth(d)
  window.setTimezone(tz)
  window.setWebRTCIPv4(ip)
  window.setFontList(csv) / window.setSpeechVoices(csv)
  ```

**⚠️ 對 BrowseForge 的影響：**
- Phase 1 可以直接用 Layer 2 的 setter 做 per-container 指紋覆寫，不需要改 C++
- 但這些 setter 是 JS 層的，creepjs 的 prototype 偵測可能會抓到
- Phase 2 改 C++ 層時，可以參考 `cross-process-storage.patch` 的 IPC 機制（`RoverfoxStoragePut/Get`）做 per-container 指紋同步

#### 完整指紋設定 Schema（關鍵欄位）

```json
{
  "navigator.userAgent": "string",
  "navigator.platform": "Win32 | MacIntel | Linux x86_64",
  "navigator.hardwareConcurrency": "uint",
  "navigator.oscpu": "string",
  "navigator.language": "string",
  "navigator.languages": ["string"],
  "screen.width": "uint", "screen.height": "uint",
  "screen.availWidth": "uint", "screen.availHeight": "uint",
  "screen.colorDepth": "uint",
  "window.outerWidth": "uint", "window.outerHeight": "uint",
  "window.innerWidth": "uint", "window.innerHeight": "uint",
  "window.devicePixelRatio": "double",
  "webGl:vendor": "string",
  "webGl:renderer": "string",
  "webGl:supportedExtensions": ["string"],
  "webGl:parameters": {"GL_enum": "value"},
  "webGl:shaderPrecisionFormats": {"key": {"rangeMin": 127, "rangeMax": 127, "precision": 23}},
  "webrtc:ipv4": "string",
  "timezone": "America/Los_Angeles",
  "locale:language": "en", "locale:region": "US",
  "fonts": ["Arial", "..."],
  "fonts:spacing_seed": "uint32",
  "audio:seed": "uint32",
  "canvas:seed": "uint32",
  "humanize": "bool"
}
```

#### Patch 結構（Phase 2 必讀）

```
patches/
├── fingerprint-injection.patch    ← 核心：MaskConfig 整合到 DOM API
├── navigator-spoofing.patch       ← Navigator 屬性
├── screen-spoofing.patch          ← Screen 尺寸
├── webgl-spoofing.patch           ← WebGL 參數/擴展
├── audio-context-spoofing.patch   ← AudioContext
├── audio-fingerprint-manager.patch ← Audio noise
├── timezone-spoofing.patch        ← Timezone/Date
├── locale-spoofing.patch          ← Intl API
├── font-hijacker.patch            ← Font metrics noise
├── font-list-spoofing.patch       ← Font enumeration
├── webrtc-ip-spoofing.patch       ← WebRTC IP
├── cross-process-storage.patch    ← IPC 機制（per-container 可參考）
├── playwright/
│   ├── 0-playwright.patch         ← Juggler 協議（Playwright 用）
│   └── 1-leak-fixes.patch        ← 修復 Playwright 洩漏
└── (其他 debloat/security patches)
```

每個 patch 的攔截模式統一：
```cpp
#include "MaskConfig.hpp"
if (auto value = MaskConfig::GetString("navigator.userAgent"))
    return value.value();
// 否則 fallback 到原始 Firefox 實作
```

#### 啟動與 Playwright 整合

- 協議：**Juggler**（非 CDP、非 Marionette）— Playwright 專用的 Firefox 自動化協議
- 啟動方式：`playwright.firefox.launch(executable_path=camoufox_path, env={CAMOU_CONFIG_*})`
- 遠端連線：`camoufox server` 啟動 → `playwright.firefox.connect("ws://host:port")`
- Persistent context：`launch_persistent_context(user_data_dir="/path/to/profile")`

---

### B.2 fingerprint-suite（apify/fingerprint-suite）— 指紋生成

#### 核心演算法：Bayesian Generative Network

**不是硬編碼組合，是從真實數據訓練的機率模型：**
1. 收集真實瀏覽器指紋數據
2. 訓練三個 Bayesian network：input-network → header-network → fingerprint-network
3. 生成時用 recursive backtracking 確保統計一致性

**生成流程：**
```
使用者約束（browser=firefox, os=windows）
  → input-network 選擇相容的 browser+HTTP 版本
  → header-network 生成 HTTP headers
  → fingerprint-network 生成完整指紋（基於 UA 的條件機率分佈）
  → 解包 *STRINGIFIED* 複合欄位
  → 輸出 BrowserFingerprintWithHeaders JSON
```

#### 整合方案

| 方案 | 適用場景 | 工作量 |
|------|---------|--------|
| **Option A: Node.js sidecar** | 開發期間快速驗證 | 低 — npm install + 呼叫 API |
| **Option B: Port 到 Go** | 長期方案 | 高 — 需移植 Bayesian sampling + 解析 network 定義檔 |
| **Option C: 預生成 JSON 池** | Phase 1 推薦 | 中 — Node.js 批次生成 1000+ 指紋存 JSON，Go 隨機取用 |

**Phase 1 建議用 Option C**：
```bash
# 一次性生成指紋池
node generate-pool.js --browser firefox --os windows --count 500 > fingerprints-win.json
node generate-pool.js --browser firefox --os macos --count 500 > fingerprints-mac.json
```

#### 指紋輸出格式（關鍵欄位）

```json
{
  "headers": { "user-agent": "...", "accept-language": "...", "sec-ch-ua": "..." },
  "fingerprint": {
    "screen": { "width": 1920, "height": 1080, "availWidth": 1920, "availHeight": 1040, "colorDepth": 24, "devicePixelRatio": 1.0, "innerWidth": 1903, "innerHeight": 919, "outerWidth": 1920, "outerHeight": 1040 },
    "navigator": { "userAgent": "...", "platform": "Win32", "hardwareConcurrency": 8, "deviceMemory": 8, "language": "en-US", "languages": ["en-US", "en"] },
    "videoCard": { "renderer": "ANGLE (NVIDIA, NVIDIA GeForce GTX 1080 Ti Direct3D11 ...)", "vendor": "Google Inc. (NVIDIA)" },
    "fonts": ["Arial", "Courier New", "Georgia", "..."],
    "audioCodecs": { "ogg": "probably", "mp3": "probably" },
    "videoCodecs": { "h264": "probably", "webm": "probably" }
  }
}
```

**注意**：fingerprint-suite 不處理 Canvas/Audio noise seed、TLS 指紋、TCP 指紋。這些由 Camoufox 在 C++ 層處理。

---

### B.3 Donut Browser（zhom/donutbrowser）— 架構參考

#### 架構決策對比

| 決策 | Donut Browser | BrowseForge |
|------|--------------|---------------|
| 隔離方式 | 多實例（每 profile 一個 OS process） | Container（單實例多 tab） |
| UI 框架 | Tauri v2（Rust + Next.js） | WebExtension（零依賴） |
| 指紋注入 | CAMOU_CONFIG env var（啟動時） | 同左（Phase 1）+ C++ per-container（Phase 2） |
| Proxy | 每 profile 一個 donut-proxy process | WebExtension proxy.onRequest |
| 自動化 | CDP WebSocket | Juggler（Playwright）+ REST API |
| 引擎 | Camoufox + Wayfern（雙引擎） | Camoufox only |
| API 認證 | Bearer token | Bearer token（同） |
| 付費模式 | Browser interaction 需付費 | 全免費開源 |

#### 可借鏡的設計

1. **指紋生成器**：Donut 也用 Bayesian network + WebGL SQLite 資料庫，與 fingerprint-suite 同源
2. **Profile 資料模型**：UUID 目錄 + profile.json + browser-data 子目錄，與我們的設計幾乎相同
3. **Launch hook**：`launch_hook` 欄位 — 啟動前呼叫 HTTP URL 取得動態 proxy，適合 rotating proxy 場景
4. **Ephemeral profile**：臨時 profile 用完即刪，適合一次性任務
5. **MCP Server**：41+ tools 的完整 MCP 實作，可作為 Phase 3 的參考

#### 不應照搬的設計

1. **多實例架構**：記憶體效率差，這是我們的核心差異化
2. **Tauri 依賴**：增加了 runtime 依賴，不符合 portable 目標
3. **付費 gating**：browser interaction 需付費，我們全開源

---

### B.4 creepjs（abrahamjuliot/creepjs）— 指紋驗證基準

#### 偵測機制概要

creepjs 不用單一分數，用多維度判斷：
- **Lies count**：偵測到的 prototype 篡改數量
- **Trash count**：可疑/無效值數量
- **Headless %**：headless 瀏覽器特徵匹配度

#### 🔴 絕對不能做的事（會被立即標記）

**1. 不能用 JS Proxy 做 API spoofing**
creepjs 有 15+ 種 Proxy 偵測：
- `toString()` 驗證（必須回傳 `function X() { [native code] }`）
- Property descriptor 列舉（必須只有 `length` 和 `name`）
- `hasOwnProperty` 檢查 arguments/caller/prototype/toString
- Recursive prototype chain 測試
- `instanceof` + stack trace 驗證
- `Object.create().toString()` + stack trace 分析

→ **所有 spoofing 必須在 C++ 層**。Camoufox 的做法是正確的。

**2. 不能注入 Canvas noise**
creepjs 寫入隨機 pixel → 讀回 → 偵測任何修改。也檢查 `measureText('')` 是否回傳浮點數。
→ 用 Camoufox 的 `canvas:seed` 在 C++ 層做 deterministic noise，不要在 JS 層加。

**3. 不能注入 Audio noise**
creepjs 用 trap value 注入 + readback 比對、`copyFromChannel` vs `getChannelData` 交叉驗證、已知 pattern lookup。
→ 用 Camoufox 的 `audio:seed` 在 C++ 層做。

**4. 跨 context 必須一致**
- `navigator.platform`、`userAgent`、`hardwareConcurrency` 在 main thread 和所有 Worker 中必須相同
- WebGL `UNMASKED_RENDERER` 在 main thread 和 OffscreenCanvas Worker 中必須相同
→ Camoufox 的 C++ 層攔截天然滿足這個要求（process-wide）。per-container 改動時要確保 Worker 也能取到正確的 container 指紋。

#### 🟡 需要注意的事

| 檢查項 | 要求 | BrowseForge 對策 |
|--------|------|-------------------|
| GPU 字串有效性 | 必須包含已知 GPU 廠商名稱，無多餘空格 | 從真實 GPU 資料庫取值 |
| Platform API 一致性 | Windows 要有 SharedWorker/EyeDropper，Mac 要有 BarcodeDetector | Camoufox 基於真實 Firefox，天然正確 |
| Timezone/Locale 一致性 | Worker timezone = main thread，Intl locale = navigator.language | per-container timezone 改動時要確保 Worker 同步 |
| Screen 尺寸合理性 | availHeight < height（要模擬 taskbar）、innerWidth ≠ screen.width | 指紋生成時確保 avail < total |
| navigator.webdriver | 必須是 `false`（不是 undefined） | Camoufox 已處理 |
| deviceMemory 有效值 | 必須是 [0.25, 0.5, 1, 2, 4, 8, 16, 32] 之一 | 指紋生成器約束 |

#### creepjs 通過標準

- Lies count = 0
- Trash count = 0
- Headless % < 10%
- GPU confidence = high
- 無 version lie 標記

---

### B.5 對 WBS 的具體修正建議

基於以上分析，WBS 需要調整的地方：

#### 修正 1：Phase 1 指紋策略調整
原計劃：「共用 Camoufox 指紋」或「JS 層攔截」
**新建議**：利用 Camoufox 已有的 `window.setXxx()` setter 做 per-container 指紋覆寫

```
Phase 1 流程：
1. 啟動 Camoufox 時設定一組基底指紋（CAMOU_CONFIG）
2. 每個 Container Tab 開啟時，透過 content script 呼叫 window.setCanvasSeed() 等 setter
3. 這些 setter 是 Camoufox C++ 層暴露的，比純 JS 攔截更安全
4. 但仍有風險：setter 本身是 JS 呼叫，可能被偵測到呼叫痕跡
```

這比原計劃的「完全共用指紋」好很多，也比「JS prototype 覆寫」安全。

#### 修正 2：指紋生成器用 Option C
原計劃：「內建合理硬體組合池」
**新建議**：用 fingerprint-suite 預生成大量指紋存 JSON

```
Phase 1：
1. 用 Node.js 腳本呼叫 fingerprint-suite 生成 1000+ 指紋
2. 按 OS 分類存為 JSON 檔案（fingerprints-windows.json 等）
3. Go Control Server 啟動時載入，建立 profile 時隨機取用
4. 將 fingerprint-suite 的輸出格式轉換為 Camoufox 的 config 格式
```

#### 修正 3：Playwright 整合方式
原計劃：「WebSocket reverse proxy」
**新發現**：Camoufox 用 Juggler 協議（非 CDP），Donut Browser 用 CDP

```
需要在 spike 0.1.4 中確認：
- Camoufox 是否同時支援 Juggler 和 CDP？
- 如果只支援 Juggler，Playwright 是唯一的自動化選項
- 如果也支援 --remote-debugging-port（CDP），可以用更通用的 CDP client
```

#### 修正 4：cross-process-storage.patch 是 Phase 2 的關鍵參考
原計劃：「改 MaskConfig 查詢邏輯」
**新發現**：Camoufox 已有 `RoverfoxStoragePut/Get` IPC 機制

```
Phase 2 可以：
1. 擴展 cross-process-storage 的 IPC 機制
2. 加入 per-userContextId 的 namespace
3. 每個 content process 啟動時，從 parent process 取得對應 container 的指紋
4. 這比直接改 MaskConfig 更乾淨，且不影響 Camoufox 原有的 patch
```

> 以下按領域分類，每個領域包含：所需知識範圍、對應 WBS 項目、以該領域專家視角給出的注意事項與驗證清單。

---

### 領域 1：瀏覽器指紋與反偵測 🔴 深度

**知識範圍：**
- 指紋維度全貌：Canvas、WebGL、AudioContext、Navigator、Screen、Timezone、Font、WebRTC、Battery
- 指紋一致性：欄位間的約束關係（UA ↔ platform ↔ screen ↔ GPU）
- 偵測方的手段：prototype chain 檢查、toString 檢查、timing attack、JS 一致性驗證
- 指紋熵與市場佔比：什麼組合「看起來正常」，什麼組合一眼假
- 學習來源：[creepjs](https://github.com/abrahamjuliot/creepjs)、[browserleaks.com](https://browserleaks.com)、[niespodd/browser-fingerprinting](https://github.com/niespodd/browser-fingerprinting)
- 對應 WBS：§3 全部

**⚠️ 專家建議：**

**1. 不要自己發明指紋組合，要從真實數據抄。**
自己拍腦袋配的指紋組合（例如 RTX 4090 + 1366x768 螢幕 + 2GB RAM）在統計上不存在，反而成為獨特標記。必須從 Steam Hardware Survey、StatCounter 等來源收集真實組合。

**2. Canvas noise 不是加越多越好。**
過大的 noise 會讓 Canvas hash 落在「不可能是真實瀏覽器」的分佈區間。noise seed 應該產生微小偏移（1-3 個 pixel 值的差異），讓 hash 不同但仍在合理範圍內。

**3. WebGL renderer 字串必須精確匹配真實顯卡。**
`ANGLE (NVIDIA, NVIDIA GeForce RTX 3060)` 是真實值，但 `NVIDIA GeForce RTX 3060` 少了 ANGLE 前綴就是假的。不同 OS 和驅動版本的格式也不同。建議直接從真實瀏覽器收集 renderer 字串，不要自己拼。

**4. 指紋不只是靜態值，還有行為指紋。**
Canvas 的 rendering 結果、WebGL 的 shader 精度、AudioContext 的 oscillator 波形——這些是硬體決定的行為，不是改一個字串就能偽造的。Phase 1 共用 Camoufox 指紋時，所有 container 的行為指紋必然相同，這是已知限制，不要試圖用 JS 層解決。

**5. 驗證清單：**
- [ ] 每個生成的指紋組合，在 browserleaks.com 上不觸發任何紅色警告
- [ ] Canvas hash 在不同 profile 之間不同（Phase 2）
- [ ] WebGL renderer 字串能在 GPU 資料庫中查到對應的真實顯卡
- [ ] UA 版本號與 Camoufox 實際的 Firefox 版本一致（不要用過時的版本號）
- [ ] navigator.platform 與 UA 中的 OS 一致
- [ ] screen resolution 在 StatCounter 前 20 名之內

---

### 領域 2：Firefox WebExtension API 🔴 深度

**知識範圍：**
- contextualIdentities API（Container 管理）
- proxy.onRequest API（per-request proxy 路由）
- tabs / cookies / storage / webRequest / sidebar_action API
- Background script ↔ Sidebar message passing
- 學習來源：[MDN WebExtensions](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions)
- 對應 WBS：§2、§4、§6

**⚠️ 專家建議：**

**1. Manifest V2 vs V3：Camoufox 用的是 Firefox，堅持用 Manifest V2。**
Firefox 的 MV3 實作與 Chrome 不同，且 `proxy.onRequest` 和 `webRequest.onAuthRequired` 在 MV3 中的行為可能有變。MV2 在 Firefox 上仍然完全支援且更穩定。不要為了「跟上潮流」用 MV3。

**2. `proxy.onRequest` 是同步 API，不能做 async 操作。**
listener 必須同步回傳 proxy 設定。這意味著 container → profile 的映射表必須在記憶體中，不能每次去 fetch Control Server。啟動時一次性載入，之後透過 WebSocket 增量更新。

**3. `contextualIdentities.remove()` 不會自動清除該 container 的 cookie。**
刪除 container 後，cookie 可能殘留。必須在刪除 container 前先呼叫 `browsingData.removeCookies({ cookieStoreId })` 明確清除。

**4. `browser.tabs.captureVisibleTab()` 只能截當前可見的 tab。**
如果要截一個背景 tab 的畫面，需要先 `browser.tabs.update(tabId, { active: true })` 切到前景，截完再切回去。或者用 Playwright 的 screenshot 替代。

**5. Sidebar 的生命週期：sidebar 關閉時 JS context 會被銷毀。**
不要在 sidebar 的 JS 中維護重要狀態。所有狀態放在 background script 中，sidebar 每次開啟時從 background 拉取。

**6. `browser.tabs.discard()` 的限制。**
discard 後 tab 的 URL 和 title 保留，但頁面內容被卸載。使用者點擊該 tab 時會自動 reload。但 discard 不能用在 active tab 上，必須先切到其他 tab。

**7. 驗證清單：**
- [ ] proxy.onRequest 的 requestInfo 確實包含 cookieStoreId（spike 0.1.2）
- [ ] 同一個 container 的所有請求（包括 XHR、fetch、img、CSS）都走同一個 proxy
- [ ] container 刪除後，用 `browser.cookies.getAll({ storeId })` 確認 cookie 已清空
- [ ] sidebar 關閉再開啟後，profile 列表和狀態正確恢復
- [ ] 100 個 container 同時存在時，contextualIdentities API 仍正常回應

---

### 領域 3：Gecko / SpiderMonkey 內部架構 🔴 深度（Phase 2）

**知識範圍：**
- Gecko DOM 層：Document → BrowsingContext → OriginAttributes → userContextId
- SpiderMonkey：JSContext → Realm → Global → 回溯到 DOM
- Camoufox patch 結構
- Gecko build 系統（mach、mozconfig）
- 學習來源：[Firefox Source Docs](https://firefox-source-docs.mozilla.org/)、Camoufox 原始碼
- 對應 WBS：§1.4、§3.4

**⚠️ 專家建議：**

**1. 第一件事：完整閱讀 Camoufox 的所有 patch，不要急著改。**
Camoufox 對 Firefox 的改動是以 patch 檔管理的。先花 1-2 天完整讀完每個 patch，理解它的攔截模式，再開始改。盲目改動會破壞 Camoufox 原有的反偵測能力。

**2. 取得 userContextId 的路徑因攔截點而異，沒有統一方法。**
- DOM 層（Canvas、WebGL、Navigator、Screen）：從 `nsINode` 或 `nsIDocument` 可以拿到 `BrowsingContext`，再取 `OriginAttributes.mUserContextId`。相對直接。
- SpiderMonkey 層（Timezone、Date）：JS 引擎內部沒有 DOM 概念。需要從 `JSContext` → `JS::Realm` → `JS::GetRealmGlobal()` → cast 回 `nsGlobalWindowInner` → 取 `Document` → 取 `BrowsingContext`。這條路徑長且脆弱，Firefox 版本更新可能改變。

**3. Gecko 完整 build 第一次要 30-60 分鐘，增量 build 5-10 分鐘。**
開發迴圈很慢。建議：先在一個攔截點（如 Canvas）上完成 per-container 改動並驗證，確認模式可行後再批量套用到其他攔截點。不要一次改完所有攔截點再 build。

**4. 不要改 Camoufox 的 patch 檔，要加新的 patch。**
保持 Camoufox 原有 patch 不動，你的 per-container 改動作為額外的 patch 疊加上去。這樣上游更新時只需要 rebase 你的 patch，不會和 Camoufox 的 patch 衝突。

**5. Timezone 是最難的攔截點，建議最後做。**
`js/src/jsdate.cpp` 中的 `DateTimeInfo` 是 thread-local 的，不是 per-context 的。要改成 per-BrowsingContext 需要在 SpiderMonkey 和 DOM 之間建立新的橋接。如果時間不夠，可以先跳過 timezone，用「所有帳號用同區域 proxy」的使用策略繞過。

**6. 驗證清單：**
- [ ] 修改後的 Camoufox build 能正常啟動，原有反偵測功能不受影響
- [ ] 兩個不同 container 的 Canvas toDataURL 結果不同
- [ ] 兩個不同 container 的 WebGL getParameter(UNMASKED_RENDERER_WEBGL) 回傳不同值
- [ ] 修改後通過 creepjs 檢測，無新增的紅色警告
- [ ] 上游 Camoufox 更新後，你的 patch 能乾淨 rebase（或衝突可在 30 分鐘內解決）

---

### 領域 4：Go 後端開發 🟡 熟練

**知識範圍：**
- net/http 或 chi router、JSON 處理、檔案系統操作
- WebSocket server、並發控制（sync.RWMutex）
- 結構化日誌（slog）、交叉編譯
- 對應 WBS：§5、§7、§10、§11、§14

**⚠️ 專家建議：**

**1. Profile JSON 寫入必須是 atomic 的。**
先寫入 `.tmp` 檔再 `os.Rename` 到正式路徑。直接 `os.WriteFile` 如果中途 crash 會留下半寫的損壞檔案。Go 的 `os.Rename` 在同一檔案系統上是 atomic 的。

**2. WebSocket 連線管理要處理好 concurrent write。**
gorilla/websocket 的 `WriteMessage` 不是 goroutine-safe 的。需要用 channel 或 mutex 序列化寫入。常見 pattern：一個 goroutine 專門負責寫，其他 goroutine 透過 channel 送 message。

**3. 不要用 gin/echo 這類重框架，用 chi 或 stdlib。**
Control Server 的 API 很簡單（~15 個 endpoint），不需要重框架。chi 或 Go 1.22 的 stdlib `http.ServeMux`（支援 method routing）就夠了。依賴越少，binary 越小，跨平台問題越少。

**4. 驗證清單：**
- [ ] Profile JSON 寫入中途 kill process → 重啟後 profile 不損壞（atomic write 生效）
- [ ] 10 個並發 API 請求不會造成 race condition（`go test -race`）
- [ ] 交叉編譯的 binary 在目標平台能直接執行（無 CGO 依賴）

---

### 領域 5：Playwright 自動化 🟡 熟練

**知識範圍：**
- Playwright Firefox protocol（connect vs launch）
- 連接已存在的瀏覽器實例
- 多 tab / 多 context 操作
- 學習來源：[Playwright docs](https://playwright.dev/)
- 對應 WBS：§8.1、§0.1.4

**⚠️ 專家建議：**

**1. Playwright 連接 Firefox 和連接 Chromium 的方式不同。**
Chromium 用 CDP（Chrome DevTools Protocol），Firefox 用 Marionette 或 Juggler（Playwright 自己的 Firefox 協議）。`firefox.connect()` 需要的是 Playwright 自己的 WebSocket endpoint，不是標準的 remote debugging port。Camoufox 是否暴露這個 endpoint 需要在 spike 中驗證。

**2. Playwright 的 BrowserContext ≠ Firefox Container。**
Playwright 的 `browser.newContext()` 建立的是 Playwright 層面的隔離 context，不是 Firefox 的 contextualIdentity。連接到已存在的 Camoufox 時，Playwright 看到的是 browser 的 default context，container tab 都在裡面。你需要透過 tab URL 或 title 來辨識哪個 tab 屬於哪個 profile。

**3. 如果 Playwright 直連不可行，REST API 中轉是完全可用的 Plan B。**
navigate、click、type、screenshot 這些操作透過 WebExtension 的 `browser.tabs.executeScript` 都能實現。Playwright 直連是「錦上添花」，不是必要條件。不要在這裡卡太久。

**4. 驗證清單：**
- [ ] Playwright 能 connect 到正在跑的 Camoufox（spike 0.1.4）
- [ ] 連入後能列出所有 tab 並操作指定 tab
- [ ] 長時間連線（>1 小時）不會斷線
- [ ] Playwright 操作不會破壞 container 的 cookie 隔離

---

### 領域 6：網路代理協議 🟡 熟練

**知識範圍：**
- SOCKS5 協議（認證、DNS 解析模式）
- HTTP CONNECT proxy
- DNS leak 原理與防護
- Residential vs Datacenter proxy 差異
- 對應 WBS：§4 全部

**⚠️ 專家建議：**

**1. DNS leak 是最常被忽略的洩漏管道。**
即使 HTTP 流量走了 proxy，DNS 查詢可能走本機 ISP 的 DNS。這會暴露你的真實地理位置。`proxy.onRequest` 回傳時必須設 `proxyDNS: true`，且要用 DNS leak test 網站驗證。

**2. SOCKS5 認證有兩種模式，要都支援。**
- Username/Password 認證（RFC 1929）：大多數 proxy 供應商用這個
- 無認證（IP 白名單模式）：部分供應商用 IP 白名單代替帳密
- `proxy.onRequest` 回傳的 username/password 為空時，Firefox 會用無認證模式

**3. Proxy 407 認證彈窗會破壞使用者體驗。**
HTTP proxy 認證失敗時 Firefox 會彈出帳密輸入框。必須用 `webRequest.onAuthRequired` 攔截，自動填入帳密。如果帳密錯誤，要捕獲連續 407 避免無限迴圈（最多重試 2 次）。

**4. Residential proxy 的 IP 會輪替，不要 cache IP 檢查結果。**
Rotating residential proxy 每次請求可能是不同 IP。proxy 健康檢查不能只檢查一次就 cache「正常」，要定期重新檢查。

**5. 驗證清單：**
- [ ] 每個 container 的出口 IP 確實不同（httpbin.org/ip）
- [ ] DNS leak test 通過（dnsleaktest.com）— 所有 DNS 查詢走 proxy
- [ ] SOCKS5 + 帳密認證正常運作
- [ ] HTTP proxy + 帳密認證正常運作，無 407 彈窗
- [ ] Proxy 斷線時，該 container 的請求失敗而非 fallback 到直連（不能洩漏真實 IP）

---

### 領域 7：前端開發（Vanilla JS）🟡 熟練

**知識範圍：**
- DOM 操作、CSS layout、事件處理
- fetch API、WebSocket client
- 對應 WBS：§6

**⚠️ 專家建議：**

**1. Sidebar 寬度只有 280px，UI 設計要極度精簡。**
不要試圖塞太多資訊。每個 profile 項目只顯示：名稱 + 狀態圖示 + proxy 國旗。詳細資訊放在 hover tooltip 或右鍵選單。

**2. 不要用任何 UI 框架。**
WebExtension sidebar 的 JS context 很輕量，引入 React/Vue 會拖慢載入速度且增加 extension 體積。280px sidebar 的 UI 複雜度用 vanilla JS 完全可控。

**3. 長列表要做虛擬滾動。**
如果有 100+ profile，全部渲染 DOM 節點會卡。當 profile 數量 > 50 時，只渲染可見區域的 DOM 節點。可以用簡單的 IntersectionObserver 或固定高度 + scroll offset 計算。

**4. 驗證清單：**
- [ ] 100 個 profile 時 sidebar 滾動流暢（無明顯卡頓）
- [ ] sidebar 關閉再開啟，狀態完整恢復
- [ ] 暗色主題在 Firefox 預設暗色模式下視覺一致

---

### 領域 8：跨平台開發 🟡 熟練

**知識範圍：**
- Windows / macOS / Linux 路徑、process 管理、shell 差異
- 對應 WBS：§9、§1.2

**⚠️ 專家建議：**

**1. macOS 的 Gatekeeper 會阻擋未簽署的 binary。**
使用者下載 ZIP 解壓後，macOS 會標記所有檔案為「來自網路」。必須在 start.sh 中加入 `xattr -cr .` 清除標記，或在 README 中說明如何手動允許。

**2. Windows 的路徑長度限制（260 字元）。**
`profiles/prof_a1b2c3d4/firefox-data/storage/default/https+++www.facebook.com/...` 這種路徑很容易超過 260 字元。Go 在 Windows 上用 `\\?\` 前綴可以繞過，但 Firefox 本身可能不行。Profile 目錄路徑要盡量短。

**3. start.bat 中不要用 Unix 風格的命令。**
`kill`、`grep`、`sleep` 在 Windows 上不存在。用 `taskkill`、`findstr`、`timeout /t`。或者考慮用 Go 寫一個跨平台的 launcher 取代 shell script。

**4. 驗證清單：**
- [ ] 在 Windows、macOS、Linux 上各跑一次完整流程
- [ ] Windows 上 profile 路徑不超過 200 字元（留 60 字元給 Firefox 內部子目錄）
- [ ] macOS 上 start.sh 執行後不被 Gatekeeper 阻擋

---

### 領域 9-12：基本了解即可 🟢

| 領域 | 關鍵注意事項 | 對應 WBS |
|------|-------------|---------|
| **密碼學** | 不要自己實作加密演算法，用 Go 標準庫 `crypto/aes` + `crypto/cipher`。IV 必須每次隨機生成，不能重複使用 | §5.5 |
| **GeoIP** | MaxMind GeoLite2 免費版精度到城市級，對 timezone 映射夠用。注意 license 要求：必須每 30 天更新一次資料庫 | §13.3 |
| **CI/CD** | Gecko build 在 CI 上很慢（>30 分鐘），Phase 2 的 CI 建議用 self-hosted runner 或只在 release 時 build | §1.4.2、§15.4 |
| **REST API** | 用 PUT 做完整更新、PATCH 做部分更新。本專案建議統一用 PUT + 部分更新語意（簡化實作） | §7.2-7.4 |

---

### 領域 13：社群平台反偵測機制 🟢 領域知識

**知識範圍：**
- Facebook / Instagram / Twitter 的多帳號偵測策略
- 行為指紋（滑鼠軌跡、打字節奏、瀏覽模式）
- 帳號養號策略
- 對應 WBS：產品驗證、使用建議文件

**⚠️ 專家建議：**

**1. 指紋只是偵測的一部分，行為模式更重要。**
Facebook 的風控系統（Sigma）不只看指紋，還看：
- 登入時間模式（10 個帳號都在同一秒登入 → 明顯自動化）
- 操作節奏（每個帳號都是「登入 → 發文 → 登出」完全相同的流程）
- 滑鼠/觸控軌跡（直線移動 vs 自然曲線）
- 網路請求時序（多個帳號的請求間隔完全一致）

**2. 新帳號需要 warm-up 期。**
剛建立的帳號立刻大量操作會被標記。建議使用者：
- 前 3 天只做瀏覽，不發文不互動
- 第 4-7 天開始少量互動（按讚、留言）
- 第 2 週後才開始正常操作
- 這些建議應該寫在產品的使用指南中

**3. 同一 proxy IP 上不要跑太多帳號。**
即使指紋不同，同一個 IP 上出現 5+ 個帳號登入，平台會標記該 IP。Residential proxy 比 datacenter proxy 安全，但也不是無限的。建議每個 IP 最多 2-3 個帳號。

**4. 這些不是技術問題，是產品文件問題。**
BrowseForge 的 README 或使用指南中應該包含「最佳實踐」章節，告訴使用者怎麼用才不會被封。這不是 WBS 的開發任務，但是產品成功的關鍵。

---

### 技能缺口評估矩陣

| # | 領域 | Phase | 深度 | 自評程度 |
|---|------|-------|------|---------|
| 1 | 瀏覽器指紋與反偵測 | 1+2 | 🔴 深 | _____ |
| 2 | Firefox WebExtension API | 1 | 🔴 深 | _____ |
| 3 | Gecko / SpiderMonkey C++ | 2 | 🔴 深 | _____ |
| 4 | Go 後端開發 | 1 | 🟡 熟練 | _____ |
| 5 | Playwright 自動化 | 1 | 🟡 熟練 | _____ |
| 6 | 網路代理協議 | 1 | 🟡 熟練 | _____ |
| 7 | 前端 Vanilla JS | 1 | 🟡 熟練 | _____ |
| 8 | 跨平台開發 | 1 | 🟡 熟練 | _____ |
| 9 | 密碼學 | 2 | 🟢 基本 | _____ |
| 10 | GeoIP | 1 | 🟢 基本 | _____ |
| 11 | CI/CD | 1 | 🟢 基本 | _____ |
| 12 | REST API 設計 | 1 | 🟢 基本 | _____ |
| 13 | 社群平台反偵測機制 | 1 | 🟢 領域 | _____ |

> Phase 1 最大瓶頸：#2（WebExtension API）— Firefox-only 的 contextualIdentities 和 proxy.onRequest 文件少、範例少，需要靠實驗摸索。
> Phase 2 最大瓶頸：#3（Gecko C++）— 學習曲線最陡，且開發迴圈慢（每次 build 5-10 分鐘）。

---

*文件版本：v1.1 | 最後更新：2026-04-23*
