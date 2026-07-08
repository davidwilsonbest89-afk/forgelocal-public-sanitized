# 反偵測機制技術文件

BrowseForge 整合兩套反偵測瀏覽器引擎，各有不同的指紋偽裝架構。

> Current architecture source: [Dual-Browser Anti-Detection Architecture](dual-browser-architecture.md).

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
| `--fingerprint-platform=PLATFORM` | `runtimes.cloakbrowser.settings.fingerprint_platform` 或平台預設 | CloakBrowser 指紋平台；`auto` 時 macOS→`macos`，其他平台→`windows` |
| `--fingerprint-webrtc-ip=auto` | 有 proxy 時自動加 | WebRTC IP 偽裝 |
| `--fingerprint-hardware-concurrency=N` | 指紋池 | CPU 核心數 |
| `--fingerprint-screen-width=W` | 指紋池 | 螢幕寬度 |
| `--fingerprint-screen-height=H` | 指紋池 | 螢幕高度 |
| `--fingerprint-fonts-dir=PATH` | `runtimes.cloakbrowser.settings.fonts_dir` 或 Linux `/usr/share/fonts` | 目標平台字型目錄；Windows 身份需真正 Windows 字型以改善 Pixelscan/CreepJS 字型一致性 |
| `--fingerprint-storage-quota=MB` | `runtimes.cloakbrowser.settings.storage_quota_mb` | Storage quota 覆寫；可改善 BrowserScan non-incognito，但可能和 FingerprintJS 取捨 |
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
- **57 個 source-level fingerprint patches（CloakBrowser v146 Linux/Windows）** — canvas, WebGL, audio, fonts, GPU, screen, WebRTC, storage quota, CDP 等。BrowseForge 使用公開 wrapper/flag contract；CloakBrowser public repo 不含完整 Chromium C++ patch source。

### 已知限制與取捨

- macOS fingerprint profile 對 aggressive detector 有已知不一致；CloakBrowser wrapper 預設 macOS 使用 `--fingerprint-platform=macos`，Linux/Windows 使用 `windows`。
- `--fingerprint-storage-quota` 是取捨：預設 auto 約 500MB 偏向通過 FingerprintJS，但 BrowserScan 可能仍扣 non-incognito；較高值（例如 5000MB）可通過 BrowserScan，但可能觸發 FingerprintJS。
- `safe_gpu`、`auto_safe_gpu_fallback`、`isolated_runtime_cache` 是啟動穩定性策略，不是 stealth 加分策略；無 GPU VM 優先用 `auto_safe_gpu_fallback`，避免預設關 GPU 造成 WebGL authenticity 下降。

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

### Camoufox schema 注意事項

Camoufox v135 的 `settings/camoucfg.jvv` 確認下列欄位有效：

- `headers.User-Agent`、`headers.Accept-Language`
- `navigator.userAgent`、`navigator.platform`、`navigator.oscpu`
- `navigator.language`、`navigator.languages`
- `fonts`、`fonts:spacing_seed`
- `AudioContext:sampleRate`、`AudioContext:outputLatency`、`AudioContext:maxChannelCount`
- `webGl:*` / `webGl2:*` 的 extensions、parameters、shader precision、context attributes

`audio:seed` 不是有效 Camoufox key。需要 audio 控制時應使用 `AudioContext:*` 欄位，且值必須來自真實裝置分布或保守平台預設。

### 環境變數傳遞

```go
Env: camoufoxEnv(configJSON, map[string]string{
    "DISPLAY": os.Getenv("DISPLAY"),  // Docker/Xvfb 需要
    "HOME":    os.Getenv("HOME"),
})
```

Camoufox 的 `MaskConfig` 會優先讀取 `CAMOU_CONFIG_1`、`CAMOU_CONFIG_2` ...，再 fallback 到單一 `CAMOU_CONFIG`。BrowseForge 對大型 WebGL/voices/fonts 設定使用 chunked env，避免單一環境變數過長。

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
