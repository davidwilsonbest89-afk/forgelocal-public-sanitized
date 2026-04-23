# CamoufoxMulti

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

```bash
# 下載 release
unzip CamoufoxMulti-v1.0.0-lite-macos-x64.zip
cd CamoufoxMulti-lite

# 啟動（首次會自動下載瀏覽器 ~440MB）
./CamoufoxMulti
# → Dashboard 自動開啟 http://127.0.0.1:19280
```

### Linux Server

```bash
sudo apt install -y xvfb
xvfb-run ./CamoufoxMulti
```

## 從原始碼 Build

```bash
# 需要 Go 1.22+ 和 Node.js 22+
git clone https://github.com/nczz/CamoufoxMulti.git
cd CamoufoxMulti

# 生成指紋池
npm install
node scripts/generate-fingerprints.js --browser firefox --os windows --count 500
node scripts/generate-fingerprints.js --browser firefox --os macos --count 500
node scripts/generate-fingerprints.js --browser chrome --os windows --count 500
node scripts/generate-fingerprints.js --browser chrome --os macos --count 500

# Build
go build -ldflags="-s -w" -o CamoufoxMulti ./cmd/server

# 啟動
./CamoufoxMulti
```

### Docker Build（不裝 Go）

```bash
make build    # 交叉編譯 macOS/Linux binary
make package  # 打包 ZIP
```

## API 文件

詳見 [API.md](API.md)

## 架構

```
CamoufoxMulti (Go binary)
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
