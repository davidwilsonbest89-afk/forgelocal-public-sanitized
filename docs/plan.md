# CamoufoxMulti — 跨平台多指紋反偵測瀏覽器

## 專案概述

CamoufoxMulti 是基於 Camoufox（Firefox 核心）的跨平台 portable 瀏覽器工具，核心目標：

- **每個分頁一個獨立指紋 profile**，實現多帳號社群營運管理
- **跨平台**（Windows / macOS / Linux）
- **Portable**（解壓即用，零安裝）
- **可程式化呼叫**（REST API + Playwright 協議），支援自動化控制

### 目標使用場景

- 多開 Facebook / Twitter / Instagram 等社群帳號
- 社群帳號營運管理（品牌、客戶分組）
- AI agent / MCP 整合自動化操作
- 排程發文、自動互動等 workflow

---

## 技術選型與核心決策

### 為什麼選 Camoufox？

| 特性 | Camoufox (Firefox) | Chromium 系競品 |
|------|-------------------|----------------|
| 指紋偽造層級 | C++ 層級（Gecko 引擎內部） | JS 注入（容易被偵測） |
| 偵測難度 | 高（原生 API 回傳假值） | 中（可被 JS 一致性檢查識破） |
| 指紋熵 | Firefox 指紋特徵，不易被歸類為反偵測瀏覽器 | Chromium 指紋特徵，反偵測瀏覽器用戶集中 |
| 開源 | ✅ 可 fork 客製化 | 多數競品閉源 |

### 多指紋隔離策略：Container-based 單實例

**不走多實例方案**（每分頁一個獨立 Camoufox process），改用 Firefox Container Tabs 機制：

- 一個 Camoufox 實例，多個 Container Tab
- 每個 Container 獨立 cookie / storage / cache / proxy
- 指紋偽造層改為 per-container（Phase 1 用 JS 攔截，Phase 2 改 C++ 層）

**效能對比：**

| | 多實例方案 | Container 方案（本專案） |
|---|---|---|
| 10 帳號 RAM | ~5 GB | ~1.5-2 GB |
| 20 帳號 RAM | ~10 GB | ~3-4 GB |
| 50 帳號 RAM | ~25 GB | ~8-10 GB |
| 啟動速度 | 每帳號 3-5 秒 | 開 tab 即用（<1 秒） |
| 指紋隔離度 | 100%（獨立 process） | ~99%（極少數邊界 case） |

### UI 框架：WebExtension（不用 Electron）

直接用 Camoufox 本身當 UI shell，管理介面以 WebExtension 實作：

- **零額外依賴** — 不需要 Electron / Tauri / 任何 runtime
- **天然跨平台** — Camoufox 本身有 Win / Mac / Linux build
- **天然 portable** — Firefox 支援 `-profile` 指定 portable 路徑
- **UI 就是 HTML/CSS/JS** — 標準 Web 技術

### Control Server：Go

- 單一 binary，跨平台編譯（`GOOS=windows/darwin/linux`）
- 不需要 runtime，binary ~10-15MB
- 內建 HTTP server，零外部依賴

---
## 系統架構

### 整體架構圖

```
使用者 / 外部程式 / AI Agent / MCP
              │
              ▼
┌──────────────────────────────┐
│      Control Server (Go)     │  ← HTTP/WebSocket, port 可設定
│                              │
│  REST API    Playwright Proxy│
│  (簡單操作)   (複雜自動化)    │
└──────────────┬───────────────┘
               │ Playwright Firefox Protocol
               ▼
┌──────────────────────────────┐
│       Camoufox Browser       │
│                              │
│  ┌────────┐ ┌────────┐      │
│  │Cont. 1 │ │Cont. 2 │ ... │
│  │FB #1   │ │FB #2   │      │
│  │Cookie A│ │Cookie B│      │
│  │Proxy X │ │Proxy Y │      │
│  │指紋 α  │ │指紋 β  │      │
│  └────────┘ └────────┘      │
│                              │
│  WebExtension (管理 UI)      │
│  C++ 指紋偽造層 (per-container)│
└──────────────────────────────┘
```

### 元件職責

| 元件 | 職責 |
|------|------|
| **Camoufox Browser** | 核心瀏覽器，提供 Container Tab 隔離 + C++ 指紋偽造 |
| **WebExtension** | Profile 管理 UI、Container 生命週期管理、Proxy 路由、指紋 JS 攔截 |
| **Control Server (Go)** | 對外 REST API、Playwright endpoint proxy、Profile CRUD、隨瀏覽器啟動/關閉 |
| **Profile Store** | 檔案系統上的 JSON + Firefox profile 目錄，portable 可搬移 |

### Portable 打包結構

```
CamoufoxMulti-v1.0-win64.zip  (~120MB 預估)
│
├── camoufox.exe              ← Camoufox 主程式 + Firefox DLLs
├── control-server.exe        ← Go API server (~15MB)
├── extension/                ← WebExtension 檔案
│   ├── manifest.json
│   ├── background.js
│   ├── sidebar/
│   │   ├── index.html
│   │   ├── style.css
│   │   └── app.js
│   └── lib/
│       ├── fingerprint.js    ← JS 層指紋攔截
│       ├── container.js      ← Container 管理
│       └── proxy.js          ← Per-container proxy 路由
├── profiles/                 ← 使用者 profile 資料（初始為空）
│   └── .gitkeep
├── data/                     ← Firefox profile 資料
├── config.json               ← 全域設定
├── start.bat                 ← Windows 啟動腳本
├── start.sh                  ← Linux/macOS 啟動腳本
└── README.md
```

---

## Profile 資料模型

每個 Profile 代表一個獨立的瀏覽器身份：

```json
{
  "id": "prof_a1b2c3",
  "name": "FB 品牌帳號 #1",
  "group": "客戶A",
  "tags": ["facebook", "品牌"],
  "created_at": "2026-04-23T12:00:00Z",
  "last_used": "2026-04-23T13:00:00Z",

  "fingerprint": {
    "user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:128.0) Gecko/20100101 Firefox/128.0",
    "platform": "Win32",
    "screen_width": 1920,
    "screen_height": 1080,
    "color_depth": 24,
    "timezone": "America/New_York",
    "locale": "en-US",
    "canvas_noise_seed": 48291,
    "webgl_vendor": "Google Inc. (NVIDIA)",
    "webgl_renderer": "ANGLE (NVIDIA, NVIDIA GeForce RTX 3060)",
    "audio_noise_seed": 73625,
    "fonts": ["Arial", "Helvetica", "Times New Roman", "Courier New"],
    "do_not_track": false,
    "hardware_concurrency": 8,
    "device_memory": 8
  },

  "proxy": {
    "type": "socks5",
    "host": "proxy.example.com",
    "port": 1080,
    "username": "user",
    "password": "pass"
  },

  "container_id": 3,
  "firefox_profile_dir": "profiles/prof_a1b2c3/firefox-data"
}
```

### 指紋生成規則

指紋不能亂數生成，必須是**合理的硬體組合**：

- Screen resolution 要符合市場佔比（1920x1080, 1366x768, 2560x1440 等）
- WebGL vendor/renderer 要是真實存在的顯卡
- User-Agent 版本要與 platform 一致
- Timezone 要與 proxy IP 的地理位置匹配
- Hardware concurrency 和 device memory 要合理配對

內建一個 fingerprint generator，從預定義的合理組合池中隨機選取。

---
## API 設計

### REST API（Control Server）

Base URL: `http://localhost:{port}/api`（port 在 config.json 設定，預設 `19280`）

#### Profile 管理

```
POST   /api/profiles
  Body: { name, group?, tags?, proxy?, fingerprint? }
  → 建立 profile，fingerprint 未提供則自動生成
  → Response: { id, name, fingerprint, ... }

GET    /api/profiles
  Query: ?group=xxx&tag=xxx
  → 列出所有 profile（支援篩選）

GET    /api/profiles/:id
  → 取得單一 profile 詳情

PUT    /api/profiles/:id
  Body: { name?, group?, proxy?, fingerprint? }
  → 更新 profile

DELETE /api/profiles/:id
  → 刪除 profile 及其所有資料

POST   /api/profiles/:id/duplicate
  → 複製 profile（生成新指紋，保留 proxy 和分組設定）
```

#### Session（分頁）控制

```
POST   /api/sessions
  Body: { profile_id }
  → 開啟該 profile 的 Container Tab
  → Response: { session_id, container_id, tab_id }

GET    /api/sessions
  → 列出所有活躍 session

DELETE /api/sessions/:id
  → 關閉該分頁

DELETE /api/sessions
  → 關閉所有分頁
```

#### 瀏覽器操作

```
POST   /api/sessions/:id/navigate
  Body: { url, wait_until?: "load"|"domcontentloaded"|"networkidle" }

POST   /api/sessions/:id/click
  Body: { selector }

POST   /api/sessions/:id/type
  Body: { selector, text, delay?: 100 }

POST   /api/sessions/:id/eval
  Body: { script }
  → Response: { result }

GET    /api/sessions/:id/screenshot
  Query: ?full_page=true
  → Response: PNG binary

GET    /api/sessions/:id/content
  Query: ?selector=body
  → Response: { text }

GET    /api/sessions/:id/cookies
  → Response: [{ name, value, domain, ... }]

POST   /api/sessions/:id/cookies
  Body: [{ name, value, domain, ... }]
  → 匯入 cookies

POST   /api/sessions/:id/wait
  Body: { selector, timeout?: 10000 }
```

#### Playwright 直連

```
GET    /api/playwright/endpoint
  → Response: { ws_url: "ws://localhost:19281/playwright" }
  → 外部 Playwright client 可直接連接，完整控制能力
```

#### 系統

```
GET    /api/status
  → { version, uptime, active_sessions, memory_usage }

POST   /api/shutdown
  → 優雅關閉所有 session 和瀏覽器
```

### Playwright 使用範例

```python
from playwright.async_api import async_playwright
import httpx

API = "http://localhost:19280/api"

# 建立 profile
profile = httpx.post(f"{API}/profiles", json={
    "name": "FB Brand #1",
    "group": "客戶A",
    "proxy": {"type": "socks5", "host": "proxy.example.com", "port": 1080}
}).json()

# 開啟 session
session = httpx.post(f"{API}/sessions", json={
    "profile_id": profile["id"]
}).json()

# 用 REST API 做簡單操作
httpx.post(f"{API}/sessions/{session['id']}/navigate", json={
    "url": "https://facebook.com"
})

# 或用 Playwright 做複雜自動化
async with async_playwright() as p:
    endpoint = httpx.get(f"{API}/playwright/endpoint").json()
    browser = await p.firefox.connect(endpoint["ws_url"])
    # 取得對應 session 的 page
    pages = browser.contexts[0].pages
    page = pages[session["tab_index"]]
    await page.fill("input[name='email']", "user@example.com")
    await page.click("button[type='submit']")
```

---
## 指紋隔離技術細節

### Phase 1：JS 層攔截（WebExtension content script）

透過 content script 在每個頁面注入指紋覆寫，根據 container ID 回傳不同值：

```javascript
// content script — 注入到每個 Container Tab
// 從 background script 取得該 container 的指紋 profile

const fp = await browser.runtime.sendMessage({ type: "get_fingerprint" });

// Canvas 指紋
const origToDataURL = HTMLCanvasElement.prototype.toDataURL;
HTMLCanvasElement.prototype.toDataURL = function(...args) {
    const ctx = this.getContext("2d");
    if (ctx) addCanvasNoise(ctx, fp.canvas_noise_seed);
    return origToDataURL.apply(this, args);
};

// Navigator 屬性
Object.defineProperty(navigator, 'platform', { get: () => fp.platform });
Object.defineProperty(navigator, 'language', { get: () => fp.locale });
Object.defineProperty(navigator, 'hardwareConcurrency', { get: () => fp.hardware_concurrency });
Object.defineProperty(screen, 'width', { get: () => fp.screen_width });
Object.defineProperty(screen, 'height', { get: () => fp.screen_height });

// Timezone
const origDateTimeFormat = Intl.DateTimeFormat;
Intl.DateTimeFormat = function(locale, options) {
    options = { ...options, timeZone: fp.timezone };
    return new origDateTimeFormat(locale, options);
};
```

**限制：** JS 層攔截可被頁面偵測（檢查 prototype chain、toString 等）。對大部分網站夠用，但 Facebook 的進階偵測可能識破。

### Phase 2：C++ 層 per-container 指紋（Fork Camoufox）

改動 Camoufox 的 C++ 指紋攔截層，加入 container 感知：

**核心改動邏輯：**

```
每個被攔截的 Web API call:
  1. 取得當前 BrowsingContext
  2. 從 OriginAttributes 取得 userContextId (= container ID)
  3. 查詢 fingerprint profile map[userContextId]
  4. 回傳該 profile 對應的假值
```

**需要改動的攔截點：**

| API | 檔案位置（Gecko） | 改動內容 |
|-----|------------------|---------|
| Canvas toDataURL/toBlob | `dom/canvas/` | noise seed per container |
| Canvas getImageData | `dom/canvas/` | 同上 |
| WebGL getParameter | `dom/canvas/WebGLContext*` | vendor/renderer per container |
| AudioContext | `dom/media/webaudio/` | noise per container |
| Navigator.platform | `dom/base/Navigator.cpp` | per container |
| Navigator.userAgent | `netwerk/protocol/http/` | per container |
| Screen.width/height | `dom/base/Screen.cpp` | per container |
| Intl.DateTimeFormat | `js/src/builtin/intl/` | timezone per container |
| Date.getTimezoneOffset | `js/src/jsdate.cpp` | 同上 |

**指紋 Profile 載入方式：**

啟動時從 `profiles/` 目錄讀取所有 `fingerprint.json`，建立 `HashMap<uint32_t, FingerprintProfile>` 供 C++ 層查詢。新增/修改 profile 時透過 WebExtension native messaging 通知 C++ 層更新 map。

---

## Proxy 管理

### Per-Container Proxy 路由

使用 WebExtension `proxy.onRequest` API：

```javascript
browser.proxy.onRequest.addListener((requestInfo) => {
    const containerId = requestInfo.cookieStoreId;
    // cookieStoreId 格式: "firefox-container-1", "firefox-container-2", ...
    const profile = getProfileByContainerId(containerId);

    if (profile?.proxy) {
        return {
            type: profile.proxy.type,    // "socks5" | "http"
            host: profile.proxy.host,
            port: profile.proxy.port,
            username: profile.proxy.username,
            password: profile.proxy.password,
            proxyDNS: true
        };
    }
    return { type: "direct" };
}, { urls: ["<all_urls>"] });
```

### Proxy 池管理（Phase 2）

```json
{
  "proxy_pools": {
    "us-residential": {
      "type": "rotating",
      "endpoint": "gate.proxy-provider.com",
      "port_range": [10000, 10100],
      "username": "user",
      "password": "pass",
      "health_check_interval": 300
    },
    "static-ips": {
      "type": "static",
      "proxies": [
        { "host": "1.2.3.4", "port": 1080 },
        { "host": "5.6.7.8", "port": 1080 }
      ]
    }
  }
}
```

---
## WebExtension 管理介面

### UI 設計

使用 Sidebar Panel 作為主要管理介面：

```
┌─ Camoufox 視窗 ─────────────────────────────────┐
│ ┌─ Sidebar (280px) ─┐ ┌─ 分頁列 ──────────────┐ │
│ │                    │ │ [FB#1] [FB#2] [TW#1]  │ │
│ │ 🔍 搜尋 profile    │ ├───────────────────────┤ │
│ │                    │ │                       │ │
│ │ 📁 客戶A           │ │                       │ │
│ │   ├ FB 品牌 #1  🟢 │ │   目前分頁的網頁內容    │ │
│ │   ├ FB 品牌 #2  ⚪ │ │                       │ │
│ │   └ TW 品牌 #1  🟢 │ │                       │ │
│ │                    │ │                       │ │
│ │ 📁 客戶B           │ │                       │ │
│ │   ├ IG 帳號 #1  ⚪ │ │                       │ │
│ │   └ FB 帳號 #1  ⚪ │ │                       │ │
│ │                    │ │                       │ │
│ │ [+ 新增 Profile]   │ │                       │ │
│ │ [⚙ 設定]          │ │                       │ │
│ └────────────────────┘ └───────────────────────┘ │
└──────────────────────────────────────────────────┘

🟢 = 已開啟（有活躍分頁）  ⚪ = 未開啟
```

### 功能清單

| 功能 | 說明 |
|------|------|
| Profile CRUD | 建立、編輯、刪除、複製 profile |
| 分組管理 | 拖拉分組、批次操作 |
| 一鍵開啟 | 點擊 profile 即開啟對應 Container Tab |
| 指紋預覽 | 顯示當前 profile 的指紋摘要 |
| Proxy 狀態 | 顯示 proxy 連線狀態、IP 位置 |
| Cookie 管理 | 匯出/匯入 cookies（JSON 格式） |
| 批次操作 | 全部開啟、全部關閉、按分組開啟 |

---

## 開發路線圖

### Phase 1 — Portable MVP（3-4 週）

**目標：** 可用的 portable 多帳號瀏覽器，JS 層指紋隔離

| 週次 | 工作項目 |
|------|---------|
| W1 | WebExtension 骨架：manifest.json、background script、sidebar UI |
| W1 | Container 管理：建立/刪除 container、開啟/關閉 container tab |
| W2 | Profile 資料模型 + 檔案系統持久化（JSON） |
| W2 | JS 層指紋攔截（content script，per-container） |
| W2 | Per-container proxy 路由（`proxy.onRequest`） |
| W3 | Go Control Server：HTTP server、Profile CRUD API、Session 管理 API |
| W3 | 瀏覽器操作 API（navigate、click、type、screenshot） |
| W4 | 跨平台打包腳本（Windows .bat、Linux/macOS .sh） |
| W4 | 指紋隨機生成器（合理硬體組合池） |
| W4 | 測試 + 修 bug + 文件 |

**交付物：**
- `CamoufoxMulti-v0.1-{platform}.zip`
- 支援 5-10 個同時開啟的 profile
- REST API 可用
- 基本 sidebar 管理 UI

### Phase 2 — 指紋強化 + 實用功能（3-4 週）

| 工作項目 | 說明 |
|---------|------|
| Fork Camoufox | 建立自己的 fork，設定 CI/CD |
| C++ per-container 指紋 | 改動 Canvas、WebGL、AudioContext 攔截點 |
| C++ Navigator/Screen | 改動 Navigator、Screen 屬性攔截 |
| C++ Timezone | 改動 Intl、Date timezone 處理 |
| Profile 加密 | AES-256 加密 profile 資料（密碼保護） |
| Cookie 匯入匯出 | JSON 格式，支援從其他瀏覽器匯入 |
| Proxy 池管理 | 支援 rotating proxy、健康檢查 |
| 指紋檢測自測 | 內建 fingerprint 檢測頁面，驗證隔離效果 |

**交付物：**
- C++ 層級指紋隔離，通過 creepjs / browserleaks 檢測
- Profile 加密保護
- Proxy 池管理

### Phase 3 — 自動化生態 + 進階功能（2-3 週）

| 工作項目 | 說明 |
|---------|------|
| Playwright endpoint | 暴露 Playwright WebSocket，外部可直連 |
| MCP Server adapter | 讓 AI agent 透過 MCP 協議控制瀏覽器 |
| 自動化 workflow 引擎 | 定義 YAML workflow，排程執行 |
| 帳號健康檢查 | 偵測帳號是否被限制/封鎖 |
| 團隊協作 | Profile 匯出/匯入包，多人共用 |
| 操作日誌 | 記錄每個 session 的操作歷史 |

---
## 技術風險與對策

| 風險 | 影響 | 對策 |
|------|------|------|
| **Camoufox 上游更新** | Fork 後需要持續 rebase | 最小化改動範圍，用 patch 管理，定期同步 |
| **JS 指紋攔截被偵測** | Phase 1 的 JS 層攔截可能被 Facebook 識破 | Phase 1 先驗證產品，Phase 2 升級到 C++ 層 |
| **Firefox Container 限制** | Container 數量或功能可能有上限 | 測試極限值；備案：混合多實例 + container |
| **記憶體瓶頸** | 大量 tab 仍然吃 RAM | 實作 tab suspend（非活躍 tab 釋放記憶體） |
| **平台政策風險** | Facebook/Twitter 禁止多帳號 | 定位為「帳號管理工具」，不主動規避封鎖 |
| **Screen/Timezone 隔離** | 這兩個屬性是 process-level，per-tab 隔離困難 | Phase 1 用 JS 攔截；Phase 2 在 C++ 層用 BrowsingContext 判斷 |
| **Playwright 與 Container 整合** | Playwright 可能無法直接操作特定 container tab | 透過 Control Server 做中間層轉譯 |

---

## 競品分析

| 功能 | Multilogin | GoLogin | AdsPower | **CamoufoxMulti** |
|------|-----------|---------|----------|-------------------|
| 核心引擎 | Chromium (Mimic) + Firefox (Stealthfox) | Chromium (Orbita) | Chromium | **Camoufox (Firefox)** |
| 指紋偽造層級 | JS 注入 + 部分 C++ | JS 注入 | JS 注入 | **C++ 層級** |
| 定價 | €99/月起 | $49/月起 | $9/月起 | **免費 / 開源** |
| 跨平台 | ✅ | ✅ | ✅ | ✅ |
| Portable | ❌（需安裝） | ❌ | ❌ | **✅** |
| API 自動化 | ✅（付費） | ✅（付費） | ✅（付費） | **✅（內建）** |
| Playwright 支援 | ❌ | ❌ | ❌ | **✅** |
| 開源 | ❌ | ❌ | ❌ | **✅** |
| 雲端同步 | ✅ | ✅ | ✅ | ❌（Phase 3 考慮） |
| 團隊管理 | ✅ | ✅ | ✅ | Phase 3 |

### 差異化定位

1. **開源 + 免費** — 競品全部是 SaaS 付費模式
2. **C++ 層級指紋** — 比 JS 注入更底層、更難偵測
3. **Portable** — 解壓即用，可放 USB 隨身碟
4. **Playwright 原生支援** — 開發者友善，可直接用 Playwright 生態
5. **可程式化優先** — API-first 設計，天然適合自動化和 AI agent 整合

---

## 資源需求

### 開發環境

- Camoufox 原始碼 build 環境（Rust + C++ toolchain）
- Go 1.21+
- Node.js（WebExtension 開發/打包）
- 測試用 residential proxy（驗證指紋隔離效果）

### 硬體建議（使用者端）

| 使用規模 | RAM | CPU | 儲存空間 |
|---------|-----|-----|---------|
| 5 帳號 | 8 GB | 4 核 | 2 GB |
| 10 帳號 | 16 GB | 4 核 | 5 GB |
| 20 帳號 | 32 GB | 8 核 | 10 GB |
| 50+ 帳號 | 64 GB | 16 核 | 20 GB |

---

## 附錄：啟動流程

```
使用者執行 start.sh / start.bat
  │
  ├─ 1. 啟動 Control Server (Go)
  │     → 載入 config.json
  │     → 載入所有 profiles
  │     → 開始監聽 HTTP port
  │
  ├─ 2. 啟動 Camoufox
  │     → -profile ./data
  │     → -no-remote（防止連到其他 Firefox 實例）
  │     → 自動載入 extension/ 目錄的 WebExtension
  │
  └─ 3. WebExtension 初始化
        → 與 Control Server 建立連線
        → 載入 profile 列表
        → 渲染 sidebar UI
        → 準備就緒
```

---

*文件版本：v0.1 | 最後更新：2026-04-23*
