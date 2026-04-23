# CamoufoxMulti API 文件

## 連線資訊

| 服務 | URL | 用途 |
|------|-----|------|
| REST API | `http://127.0.0.1:19280/api` | Profile 管理 + 瀏覽器操作 |
| Dashboard | `http://127.0.0.1:19280` | Web 管理介面 |
| MCP Server | `http://127.0.0.1:19281` | AI Agent 整合（JSON-RPC） |

## 認證

所有 REST API（除 `/api/status`）需要 Bearer Token：

```
Authorization: Bearer {token}
```

Token 位於 `data/.api-token`。Dashboard 首次開啟時會要求輸入。

---

## REST API

### 系統

#### GET /api/status
```bash
curl http://127.0.0.1:19280/api/status
```
```json
{"version": "0.2.0", "status": "ok"}
```

#### POST /api/shutdown
關閉所有瀏覽器並停止 server。

---

### Profile 管理

#### POST /api/profiles — 建立 Profile
```bash
curl -X POST http://127.0.0.1:19280/api/profiles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "FB Brand #1",
    "engine": "firefox",
    "group": "客戶A",
    "tags": ["facebook", "品牌"],
    "proxy": {
      "type": "socks5",
      "host": "proxy.example.com",
      "port": 1080,
      "username": "user",
      "password": "pass"
    }
  }'
```

回傳：
```json
{
  "data": {
    "id": "prof_a1b2c3d4e5f6",
    "name": "FB Brand #1",
    "engine": "firefox",
    "group": "客戶A",
    "fingerprint": {
      "navigator.userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:135.0) ...",
      "navigator.platform": "Win32",
      "screen.width": 1920,
      "screen.height": 1080,
      "canvas:seed": 3948271650,
      "..."
    },
    "proxy": {"type": "socks5", "host": "proxy.example.com", "port": 1080}
  }
}
```

- `engine`：`"firefox"`（Camoufox）或 `"chromium"`（CloakBrowser）
- `fingerprint`：未提供時自動從指紋池分配
- `proxy`：選填，支援 `socks5` 和 `http`

#### GET /api/profiles — 列出 Profile
```bash
# 全部
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:19280/api/profiles

# 篩選
curl -H "Authorization: Bearer $TOKEN" "http://127.0.0.1:19280/api/profiles?group=客戶A&tag=facebook"
```

#### GET /api/profiles/:id — 取得單一 Profile
```bash
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6
```

#### PUT /api/profiles/:id — 更新 Profile
```bash
curl -X PUT http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name": "新名稱", "group": "客戶B"}'
```

#### DELETE /api/profiles/:id — 刪除 Profile
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/profiles/prof_a1b2c3d4e5f6
```

#### POST /api/profiles/:id/duplicate — 複製 Profile
新 ID、新指紋，保留 proxy 和分組。

---

### Session（瀏覽器控制）

#### POST /api/sessions — 開啟瀏覽器
```bash
curl -X POST http://127.0.0.1:19280/api/sessions \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"profile_id": "prof_a1b2c3d4e5f6"}'
```
```json
{"data": {"session_id": "sess_prof_a1b2c3d4e5f6", "profile_id": "prof_a1b2c3d4e5f6", "engine": "firefox"}}
```

開啟一個帶有該 Profile 指紋和 Proxy 的瀏覽器視窗。

#### GET /api/sessions — 列出活躍 Session
```bash
curl -H "Authorization: Bearer $TOKEN" http://127.0.0.1:19280/api/sessions
```

#### DELETE /api/sessions/:id — 關閉瀏覽器
```bash
curl -X DELETE -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/sess_prof_a1b2c3d4e5f6
```

---

### 瀏覽器操作

所有操作透過 Playwright，引擎無關（Firefox/Chromium 同一 API）。

#### POST /api/sessions/:id/navigate
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/navigate \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url": "https://facebook.com", "wait_until": "load"}'
```
`wait_until`：`"load"` | `"domcontentloaded"` | `"networkidle"`

#### POST /api/sessions/:id/click
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/click \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "button#login"}'
```

#### POST /api/sessions/:id/type
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/type \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "input[name=email]", "text": "user@example.com", "delay": 50}'
```
`delay`：每個字元的間隔（毫秒），模擬打字速度。

#### POST /api/sessions/:id/eval — 執行 JavaScript
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/eval \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"script": "document.title"}'
```
```json
{"data": "Facebook - log in or sign up"}
```

#### GET /api/sessions/:id/screenshot
```bash
# 回傳 PNG binary
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/$SID/screenshot \
  -o screenshot.png

# 全頁截圖
curl -H "Authorization: Bearer $TOKEN" \
  "http://127.0.0.1:19280/api/sessions/$SID/screenshot?full_page=true" \
  -o full.png
```

#### GET /api/sessions/:id/content
```bash
# 整頁 HTML
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/$SID/content

# 指定元素的文字
curl -H "Authorization: Bearer $TOKEN" \
  "http://127.0.0.1:19280/api/sessions/$SID/content?selector=h1"
```

#### POST /api/sessions/:id/wait
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/wait \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"selector": "#result", "timeout": 10000}'
```

#### GET /api/sessions/:id/cookies
```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://127.0.0.1:19280/api/sessions/$SID/cookies
```

#### POST /api/sessions/:id/cookies — 匯入 Cookies
```bash
curl -X POST http://127.0.0.1:19280/api/sessions/$SID/cookies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '[{"name":"session","value":"abc123","domain":".facebook.com","path":"/"}]'
```

---

### 備份/還原

#### POST /api/profiles/:id/export — 匯出單一 Profile
回傳 .zip 檔案。

#### POST /api/profiles/import — 匯入 Profile
```bash
curl -X POST http://127.0.0.1:19280/api/profiles/import \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@profile.zip"
```

#### POST /api/backup — 全量備份
回傳包含所有 Profile 的 .zip。

#### POST /api/restore — 全量還原
```bash
curl -X POST http://127.0.0.1:19280/api/restore \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@backup.zip"
```

---

### Workflow

#### POST /api/workflow/run — 執行 Workflow
```bash
curl -X POST http://127.0.0.1:19280/api/workflow/run \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "自動登入",
    "steps": [
      {"name": "建立", "action": "create_profile", "params": {"name": "Auto", "engine": "firefox", "var": "p1"}},
      {"name": "開啟", "action": "open_browser", "profile_id": "$p1"},
      {"name": "導航", "action": "navigate", "profile_id": "$p1", "params": {"url": "https://facebook.com"}},
      {"name": "等待", "action": "sleep", "params": {"seconds": 5}},
      {"name": "關閉", "action": "close_browser", "profile_id": "$p1"}
    ]
  }'
```

支援的 action：
| Action | 說明 | 參數 |
|--------|------|------|
| `create_profile` | 建立 Profile | `name`, `engine`, `var`（變數名） |
| `open_browser` | 開啟瀏覽器 | `profile_id` |
| `close_browser` | 關閉瀏覽器 | `profile_id` |
| `navigate` | 導航 | `profile_id`, `url` |
| `click` | 點擊 | `profile_id`, `selector` |
| `type` | 輸入 | `profile_id`, `selector`, `text` |
| `eval` | 執行 JS | `profile_id`, `script` |
| `wait` | 等待元素 | `profile_id`, `selector` |
| `screenshot` | 截圖 | `profile_id` |
| `sleep` | 等待 | `seconds` |

變數：`create_profile` 的 `var` 參數定義變數名，後續步驟用 `$變數名` 引用 profile_id。

---

## MCP Server（AI Agent 整合）

端口：`http://127.0.0.1:19281`
協議：JSON-RPC 2.0 over HTTP POST
規格：MCP 2025-11-25

### 連線

```python
# Python (mcp 套件)
from mcp import ClientSession
async with ClientSession("http://127.0.0.1:19281") as session:
    await session.initialize()
    tools = await session.list_tools()
```

### 可用 Tools（11 個）

| Tool | 說明 | 參數 |
|------|------|------|
| `list_profiles` | 列出所有 Profile | — |
| `create_profile` | 建立 Profile | `name`, `engine`, `group` |
| `delete_profile` | 刪除 Profile | `profile_id` |
| `open_browser` | 開啟瀏覽器 | `profile_id` |
| `close_browser` | 關閉瀏覽器 | `profile_id` |
| `navigate` | 導航到 URL | `profile_id`, `url` |
| `click` | 點擊元素 | `profile_id`, `selector` |
| `type_text` | 輸入文字 | `profile_id`, `selector`, `text` |
| `screenshot` | 截圖 | `profile_id` |
| `get_content` | 取得頁面內容 | `profile_id`, `selector`（選填） |
| `evaluate` | 執行 JavaScript | `profile_id`, `script` |

### 範例：用 Claude 操作

```
User: 幫我建立一個 Firefox profile，然後開啟瀏覽器到 facebook.com

Claude: [呼叫 create_profile] → 建立了 prof_xxx
        [呼叫 open_browser] → 瀏覽器已開啟
        [呼叫 navigate] → 已導航到 facebook.com
```

AI Agent 不需要知道底層是 Firefox 還是 Chromium，MCP tools 是引擎無關的。

---

## 錯誤格式

所有錯誤回傳統一格式：
```json
{
  "error": {
    "code": "PROFILE_NOT_FOUND",
    "message": "profile not found: prof_xxx"
  }
}
```

常見錯誤碼：
| Code | HTTP | 說明 |
|------|------|------|
| `UNAUTHORIZED` | 401 | Token 無效或缺失 |
| `NOT_FOUND` | 404 | Profile 或 Session 不存在 |
| `MISSING_NAME` | 400 | 建立 Profile 時缺少 name |
| `LAUNCH_FAILED` | 500 | 瀏覽器啟動失敗 |
| `NAVIGATE_FAILED` | 500 | 導航失敗 |
