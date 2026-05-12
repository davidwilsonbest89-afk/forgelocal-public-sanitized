# 反偵測機制技術文件

BrowseForge 整合兩套反偵測瀏覽器引擎，各有不同的指紋偽裝架構。

---

## CloakBrowser (Chromium) — Seed 驅動

### 架構

CloakBrowser 使用**單一種子（seed）**在 C++ 層面自動生成所有指紋值。一個 seed 產生一組完整、一致的身份，包含 GPU、Canvas、WebGL、Audio、Fonts、Screen、Hardware 等所有可偵測面向。

### BrowseForge 傳入的 Flags

| Flag | 來源 | 說明 |
|------|------|------|
| `--fingerprint=SEED` | `profile.FingerprintSeed` | 主種子，決定所有指紋值 |
| `--fingerprint-timezone=TZ` | GeoIP 偵測 | 時區（如 `Asia/Taipei`） |
| `--fingerprint-locale=LOCALE` | GeoIP 偵測 | 語系（如 `zh-TW`） |
| `--fingerprint-platform=windows` | Linux 自動加 | OS 偽裝（Linux→Windows） |
| `--fingerprint-webrtc-ip=auto` | 有 proxy 時自動加 | WebRTC IP 偽裝 |
| `--fingerprint-hardware-concurrency=N` | 指紋池 | CPU 核心數 |
| `--fingerprint-screen-width=W` | 指紋池 | 螢幕寬度 |
| `--fingerprint-screen-height=H` | 指紋池 | 螢幕高度 |
| `--fingerprint-fonts-dir=PATH` | Docker 自動偵測 | 字型目錄（對抗 Kasada/Akamai） |
| `--no-sandbox` | Docker 自動偵測 | 停用 kernel sandbox |

### 自動生成的值（由 seed 決定，不需外部傳入）

- GPU vendor/renderer
- Canvas noise
- WebGL parameters + extensions + shader precision
- Audio context fingerprint
- Client rects noise
- Device memory (預設 8GB)
- Font metrics

### 特性

- **無 GPU 環境也能通過 WebGL 偵測** — seed 在渲染管線層面生成一致的假輸出
- **同一 seed = 同一身份** — 適合「回訪者」場景
- **不同 seed = 不同身份** — 每個 profile 自動分配不同 seed
- **49+ C++ 層級 patch** — canvas, WebGL, audio, fonts, GPU, screen, WebRTC, network timing, CDP

### 已知限制

- macOS binary 只有 26 patches（Linux/Windows 有 57）
- `--fingerprint-storage-quota` 未設定（FingerprintJS vs BrowserScan 的 incognito 偵測取捨）

---

## Camoufox (Firefox) — Config 屬性注入

### 架構

Camoufox 透過 `CAMOU_CONFIG` 環境變數接收 JSON，在 C++ 層面攔截對應的 API 回傳值。每個屬性需要**明確提供**，未提供的屬性使用真實系統值。

### BrowseForge 傳入的 CAMOU_CONFIG

| 屬性 | 來源 | 說明 |
|------|------|------|
| `navigator.userAgent` | 指紋池 | UA 字串 |
| `navigator.platform` | 指紋池 | 平台 |
| `navigator.hardwareConcurrency` | 指紋池 | CPU 核心數 |
| `navigator.language` | 指紋池 | 語言 |
| `navigator.languages` | 指紋池 | 語言清單 |
| `navigator.oscpu` | 指紋池 | OS/CPU 資訊 |
| `screen.width/height/avail*` | 指紋池 | 螢幕尺寸 |
| `screen.colorDepth` | 指紋池 | 色深 |
| `window.outer/inner/devicePixelRatio` | 指紋池 | 視窗尺寸 |
| `canvas:seed` | 指紋池（隨機） | Canvas noise 種子 |
| `fonts:spacing_seed` | 指紋池（隨機） | 字型間距隨機化種子 |
| `fonts` | 指紋池（隨機子集） | 字型清單（60-80% 子集） |
| `timezone` | GeoIP 偵測 | 時區 |
| `locale:language` / `locale:region` | GeoIP 偵測 | 語系 |
| `webGl:*`（完整時） | 指紋池 + WebGL profile | 完整 WebGL 指紋 |

### WebGL 策略（關鍵）

```
有完整 WebGL profile（含 supportedExtensions + parameters + shaderPrecisionFormats）
  → 全部傳給 Camoufox

只有 renderer/vendor（不完整）
  → 不傳任何 WebGL 欄位
  → Camoufox 自動用 BrowserForge 生成完整一致的指紋
```

**原因**：只傳 renderer/vendor 會導致字串宣稱是 NVIDIA 但 `getParameter()` 回傳軟體渲染器的值，被 WAF 偵測為不一致。

### 不傳給 Camoufox 的欄位

以下欄位曾經存在於指紋池但已移除（Camoufox 不認識）：
- ~~`headers.User-Agent`~~ — Camoufox 從 `navigator.userAgent` 自動生成
- ~~`headers.Accept-Language`~~ — Camoufox 從 locale 自動生成
- ~~`_meta`~~ — 內部標記
- ~~`audio:seed`~~ — 非有效 Camoufox key

### 環境變數傳遞

```go
Env: map[string]string{
    "CAMOU_CONFIG": string(configJSON),
    "DISPLAY":      os.Getenv("DISPLAY"),  // Docker/Xvfb 需要
    "HOME":         os.Getenv("HOME"),
}
```

注意：Playwright 的 `Env` 欄位是**替換**而非合併，必須明確傳入 `DISPLAY` 和 `HOME`。

### 已知限制

- WebGL 渲染一致性依賴完整的 WebGL profile 資料
- Camoufox v135-beta.24 有維護空窗期，部分指紋不一致問題尚未修復
- 無 GPU 環境中，即使有完整 WebGL profile，Canvas 渲染 hash 仍可能不一致

---

## 指紋池生成

### 工具

- `scripts/generate-fingerprints.js` — 使用 fingerprint-generator (Apify) 的貝葉斯網路生成
- `scripts/webgl-collector.html` — 從真實 GPU 收集完整 WebGL 資料
- `data/webgl-profiles/webgl-profiles.json` — WebGL profile 資料庫

### 生成指令

```bash
node scripts/generate-fingerprints.js --browser firefox --os macos --count 500
node scripts/generate-fingerprints.js --browser firefox --os windows --count 500
node scripts/generate-fingerprints.js --browser chrome --os macos --count 500
node scripts/generate-fingerprints.js --browser chrome --os windows --count 500
```

### 指紋分配

- Profile 建立時從指紋池隨機分配一組固定指紋
- 存在 `profiles/prof_xxx/profile.json` 的 `fingerprint` 欄位
- 同一 profile 每次開啟使用相同指紋（身份一致性）
- CloakBrowser profile 額外有 `fingerprint_seed`（uint32）

---

## 偵測測試網站

| 網站 | 測試重點 |
|------|---------|
| https://bot.sannysoft.com | 自動化偵測（Playwright 特徵） |
| https://browserleaks.com | 全面指紋（Canvas、WebGL、字型、WebRTC） |
| https://abrahamjuliot.github.io/creepjs/ | 指紋一致性分析 |
| https://pixelscan.net | 指紋一致性 + 異常偵測 |
| https://iphey.com | 綜合反偵測評分 |
| https://www.browserscan.net | 指紋洩漏 + Proxy 偵測 |

---

## 參考連結

- [CloakBrowser GitHub](https://github.com/CloakHQ/CloakBrowser)
- [CloakBrowser 文件](https://cloakbrowser.dev/)
- [Camoufox 指紋文件](https://camoufox.com/fingerprint/)
- [Camoufox WebGL 研究](https://camoufox.com/webgl-research/)
- [BrowserForge](https://github.com/daijro/browserforge)
