# BrowseForge

🦊🌐 跨平台多指紋反偵測瀏覽器 — 雙引擎（Firefox + Chromium），每個 Profile 獨立指紋、Cookie、Proxy。

## 功能

- **雙引擎**：Firefox (Camoufox) + Chromium (CloakBrowser)，C++ 層級反偵測
- **獨立指紋**：每個 Profile 自動分配不同的瀏覽器指紋
- **獨立 Cookie**：各 Profile 登入狀態互不干擾，關閉重開後保留
- **獨立 Proxy**：每個 Profile 可設定不同的 SOCKS5/HTTP Proxy
- **GeoIP 調配**：根據出口 IP 自動調整 Timezone、Language、Geolocation
- **Web Dashboard**：瀏覽器開啟 `http://127.0.0.1:19280` 管理所有 Profile
- **REST API**：20 個 endpoint，完整程式化控制
- **MCP Server**：12 個 tools，AI Agent（Kiro/Claude）直接操作瀏覽器
- **YAML Workflow**：定義自動化流程，排程執行
- **Portable**：單一執行檔，首次啟動自動下載瀏覽器引擎

## 快速開始

### 從 Release 下載

到 [Releases](https://github.com/nczz/BrowseForge/releases) 下載對應平台的 ZIP：

| 平台 | 檔案 |
|------|------|
| macOS Intel | `BrowseForge-vX.X.X-lite-macos-x64.zip` |
| macOS Apple Silicon | `BrowseForge-vX.X.X-lite-macos-arm64.zip` |
| Linux x64 | `BrowseForge-vX.X.X-lite-linux-x64.zip` |
| Linux arm64 | `BrowseForge-vX.X.X-lite-linux-arm64.zip` |
| Windows x64 | `BrowseForge-vX.X.X-lite-windows-x64.zip` |

```bash
unzip BrowseForge-vX.X.X-lite-macos-x64.zip
cd BrowseForge-lite

# 啟動（首次會自動下載瀏覽器引擎 ~440MB）
./BrowseForge
# → Dashboard 自動開啟 http://127.0.0.1:19280
```

### Linux Server（無桌面環境）

```bash
sudo apt install -y xvfb
xvfb-run ./BrowseForge
```

需要遠端看到瀏覽器畫面（處理驗證碼等）？詳見 [Linux Server 部署指南](docs/linux-server.md)，支援 noVNC 遠端桌面和 Docker Compose 一鍵部署。

### Docker（最簡單）

```bash
docker run -d -p 19280:19280 -p 6080:6080 ghcr.io/nczz/browseforge:latest
```

或自行 build：
```bash
cd docker && docker compose up -d --build
```

BrowseForge 在 Docker 中會自動偵測並啟用 `--no-sandbox` 和 `0.0.0.0` 綁定。

### Windows

解壓後雙擊 `BrowseForge.exe`，首次啟動會自動下載瀏覽器引擎。

## 設定檔

首次啟動時自動生成 `config.json`：

```json
{
  "host": "127.0.0.1",
  "port": "19280",
  "profiles_dir": "profiles",
  "data_dir": "data",
  "log_file": "logs/server.log",
  "camoufox_path": "/path/to/browsers/camoufox/...",
  "cloakbrowser_path": "/path/to/browsers/cloakbrowser/...",
  "fingerprint_dir": "data"
}
```

| 欄位 | 說明 | 預設 |
|------|------|------|
| `host` | 監聽地址 | `127.0.0.1`（Docker 自動改為 `0.0.0.0`） |
| `port` | REST API + Dashboard 端口 | `19280` |
| `no_sandbox` | 停用 Chromium sandbox | `false`（Docker 自動啟用） |
| `profiles_dir` | Profile 資料目錄 | `profiles` |
| `data_dir` | 資料目錄（token、指紋池） | `data` |
| `log_file` | 日誌檔案路徑 | `logs/server.log` |
| `camoufox_path` | Camoufox 執行檔路徑（自動偵測） | 自動 |
| `cloakbrowser_path` | CloakBrowser 執行檔路徑（自動偵測） | 自動 |
| `fingerprint_dir` | 指紋池 JSON 目錄 | `data` |

瀏覽器路徑會在首次啟動時自動偵測或下載後填入。手動安裝瀏覽器時可直接修改路徑。

## CLI 命令

```bash
# 預設啟動（等同 serve）
./BrowseForge

# 指定監聽地址和停用 sandbox（Docker 部署）
./BrowseForge serve --host 0.0.0.0 --no-sandbox

# 顯示 API Token
./BrowseForge token

# 環境檢查
./BrowseForge doctor

# MCP stdio 模式（Kiro / Claude 整合）
./BrowseForge --mcp
```

### Docker 自動偵測

在 Docker 容器內執行時，BrowseForge 會自動：
- 將監聽地址改為 `0.0.0.0`
- 啟用 `--no-sandbox`（Chromium 在容器內無法使用 kernel sandbox）

無需額外設定，直接 `./BrowseForge` 即可。

## MCP Server（AI Agent 整合）

BrowseForge 內建 MCP Server，支援兩種模式：

### HTTP 模式（隨 server 自動啟動）

啟動 BrowseForge 後，MCP Server 自動在 `http://127.0.0.1:19281` 提供服務。

### stdio 模式（Kiro / Claude Desktop 整合）

```bash
# 註冊到 Kiro CLI
kiro-cli mcp add --name browseforge --command /path/to/BrowseForge --args "--mcp" --scope global

# 或手動編輯 ~/.kiro/settings/mcp.json
{
  "browseforge": {
    "command": "/path/to/BrowseForge",
    "args": ["--mcp"]
  }
}
```

Claude Desktop 的 `claude_desktop_config.json`：
```json
{
  "mcpServers": {
    "browseforge": {
      "command": "/path/to/BrowseForge",
      "args": ["--mcp"]
    }
  }
}
```

> **注意**：stdio 模式下 BrowseForge 會自動偵測瀏覽器路徑和載入 Profile。確保 `config.json` 和 `profiles/` 目錄與 BrowseForge 執行檔在同一目錄下。

### 可用 Tools（12 個）

| Tool | 說明 |
|------|------|
| `list_profiles` | 列出所有 Profile |
| `create_profile` | 建立 Profile（自動分配指紋） |
| `update_profile` | 更新 Profile 設定 |
| `delete_profile` | 刪除 Profile |
| `open_browser` | 開啟瀏覽器視窗 |
| `close_browser` | 關閉瀏覽器 |
| `navigate` | 導航到 URL |
| `click` | 點擊元素 |
| `type_text` | 輸入文字 |
| `screenshot` | 截圖 |
| `get_content` | 取得頁面內容 |
| `evaluate` | 執行 JavaScript |

使用範例（在 Kiro 或 Claude 中直接對話）：
```
建立一個 Firefox profile，開啟瀏覽器到 facebook.com
```

## REST API

Base URL: `http://127.0.0.1:19280/api`

認證：所有 API（除 `/api/status`）需要 Bearer Token：
```
Authorization: Bearer {token}
```
Token 在首次啟動時自動生成，存放在 `data/.api-token`。

完整 API 文件詳見 [API.md](API.md)。

## Playwright Connect（外部腳本直接操作）

BrowseForge 支援讓外部 Playwright 腳本直接連入操作瀏覽器。每個 session 啟動時自動暴露 WebSocket endpoint。

### 使用方式

```bash
# 1. 取得 connect endpoint
curl http://127.0.0.1:19280/api/playwright/endpoint \
  -H "Authorization: Bearer $TOKEN"
# → {"data":[{"session_id":"sess_prof_xxx","endpoint":"ws://...","proxy":"ws://..."}]}
```

**Node.js（透過 Proxy，推薦）：**
```javascript
import { firefox } from 'playwright';

const browser = await firefox.connect(
  'ws://YOUR_SERVER:19280/api/playwright/ws/sess_prof_xxx',
  { headers: { 'Authorization': 'Bearer YOUR_TOKEN' } }
);
const page = browser.contexts()[0].pages()[0];
await page.goto('https://example.com');
console.log(await page.title());
await browser.close();
```

**Python（透過 Proxy）：**
```python
from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.firefox.connect(
        "ws://YOUR_SERVER:19280/api/playwright/ws/sess_prof_xxx",
        headers={"Authorization": "Bearer YOUR_TOKEN"}
    )
    page = browser.contexts[0].pages[0]
    page.goto("https://example.com")
    print(page.title())
```

### 連線方式

| 方式 | URL | 適用場景 |
|------|-----|---------|
| **Proxy（推薦）** | `ws://host:19280/api/playwright/ws/{session_id}` | Docker、遠端、需要 token 驗證 |
| 直連 | `ws://host:{dynamic_port}/{guid}` | 本機、低延遲 |

Proxy 模式只需暴露 19280 port，自動帶 token 驗證。

### 限制

| 項目 | 說明 |
|------|------|
| Client 版本 | 必須使用 Playwright **1.59.x**（major.minor 需匹配） |
| 反偵測 | 不影響。使用 Playwright 內部協議，不暴露 CDP |

### 與 REST API 的差異

| | REST API | Playwright Connect |
|---|---------|-------------------|
| 功能範圍 | BrowseForge 定義的操作 | 完整 Playwright API |
| 適用場景 | 簡單自動化、AI Agent | 複雜腳本、測試框架 |
| 跨語言 | 任何 HTTP client | Playwright client (Node/Python/Java) |
| 版本依賴 | 無 | 需要 PW 1.59.x |
| Docker | 只需 19280 port | 只需 19280 port（透過 Proxy） |
| 驗證 | Bearer Token | Bearer Token（Proxy 模式） |

## YAML Workflow

定義自動化流程，透過 API 執行：

```yaml
name: 多帳號登入
steps:
  - name: 建立 Profile
    action: create_profile
    params: { name: "FB Account", engine: firefox, var: p1 }

  - name: 開啟瀏覽器
    action: open_browser
    profile_id: $p1

  - name: 導航
    action: navigate
    profile_id: $p1
    params: { url: "https://facebook.com" }

  - name: 等待
    action: sleep
    params: { seconds: 30 }

  - name: 關閉
    action: close_browser
    profile_id: $p1
```

執行：
```bash
TOKEN=$(cat data/.api-token)
curl -X POST http://127.0.0.1:19280/api/workflow/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d @examples/multi-login.yaml
```

支援的 action：`create_profile`、`open_browser`、`close_browser`、`navigate`、`click`、`type`、`eval`、`wait`、`screenshot`、`sleep`

## Profile 資料

每個 Profile 的資料存放在 `profiles/` 目錄下：

```
profiles/
  prof_abc123/
    profile.json        ← 指紋、Proxy、名稱、分組設定
    browser-data/       ← Cookie、localStorage、登入狀態
```

備份：打包 `profiles/` 目錄即可保存所有帳號資料。搬到另一台機器的 `profiles/` 下，啟動即恢復。

## Proxy 建議

| 類型 | 適用場景 | 偵測風險 |
|------|---------|---------|
| Residential Proxy | 社群多帳號 | 低 |
| ISP Proxy | 需要固定 IP | 低 |
| Mobile Proxy | 高風險帳號 | 最低 |
| Datacenter Proxy | 爬蟲、低風險場景 | 高（不建議用於社群） |

無 Proxy 時，BrowseForge 會自動偵測出口 IP 並調整指紋的 Timezone/Language（支援 VPN/WireGuard 場景）。

## 平台支援

| 平台 | BrowseForge | 🦊 Firefox (Camoufox) | 🌐 Chromium (CloakBrowser) |
|------|:---:|:---:|:---:|
| macOS x64 | ✅ | ✅ | ✅ |
| macOS arm64 | ✅ | ✅ | ✅ |
| Linux x64 | ✅ | ✅ | ✅ |
| Linux arm64 | ✅ | ✅ | ✅ |
| Windows x64 | ✅ | ✅ | ✅ |

詳見 [docs/platform-support.md](docs/platform-support.md)。

## 從原始碼 Build

```bash
# 需要 Go 1.22+ 和 Node.js 22+
git clone https://github.com/nczz/BrowseForge.git
cd BrowseForge

# 生成指紋池
npm install
node scripts/generate-fingerprints.js --browser firefox --os windows --count 500
node scripts/generate-fingerprints.js --browser firefox --os macos --count 500
node scripts/generate-fingerprints.js --browser chrome --os windows --count 500
node scripts/generate-fingerprints.js --browser chrome --os macos --count 500

# Build
go build -ldflags="-s -w" -o BrowseForge ./cmd/server

# 啟動
./BrowseForge
```

### Docker Build（不裝 Go）

```bash
make build    # 交叉編譯 macOS/Linux binary
make package  # 打包 ZIP
```

## 架構

```
BrowseForge (Go binary)
  ├── REST API (:19280)     ← Profile CRUD + 瀏覽器操作
  ├── MCP Server (:19281)   ← AI Agent 整合
  ├── Web Dashboard         ← 管理介面
  └── Playwright
       ├── Camoufox #1 (Profile A, 指紋α, Proxy X)
       ├── Camoufox #2 (Profile B, 指紋β, Proxy Y)
       └── CloakBrowser #3 (Profile C, 指紋γ, Proxy Z)
```

## 目錄結構

```
cmd/server/          Go 入口
internal/
  api/               REST API + Dashboard
  browser/           Playwright 瀏覽器管理 + 自動下載
  config/            設定載入
  fingerprint/       指紋池 + GeoIP 調配
  mcp/               MCP Server (HTTP + stdio)
  profile/           Profile CRUD + 持久化
  workflow/           YAML Workflow 引擎
extension/           WebExtension (sidebar + proxy + fingerprint injector)
scripts/             指紋生成腳本
data/                指紋池 JSON（gitignore，執行時生成）
docs/                Phase 2 fork 計劃
examples/            Workflow 範例
```

## License

MIT
