# 反偵測機制技術文件

BrowseForge 整合多套反偵測瀏覽器 runtime，各有不同的指紋偽裝架構。穩定路線仍可使用 Camoufox / CloakBrowser；`browseforge-chromium` 是新的 source-level Chromium alpha runtime，適合先做 opt-in validation，不應在未簽章前宣稱 production-ready。

> Current architecture source: [Dual-Browser Anti-Detection Architecture](dual-browser-architecture.md).

---

## BrowseForge Chromium Runtime — Source-level Chromium alpha

### 架構

`browseforge-chromium` 是 BrowseForge 專用的 Chromium-family runtime。BrowseForge 仍負責 profile、REST API、MCP、dashboard、workflow、backup/restore 與 session orchestration；runtime artifact 則負責 Chromium binary、source-level stealth patches、native persona config、detector evidence、SBOM/provenance 與 release manifests。

目前 runtime identity：

| 欄位 | 值 |
|------|----|
| `runtime_id` | `browseforge-chromium` |
| `family` | `chromium` |
| BrowseForge 最低版本 | `v2.0.0` |
| Runtime 狀態 | `v0.1.1-alpha.0` |
| 發佈狀態 | alpha validation artifact；尚未 release-grade 簽章 |

### 使用端整合方式

1. 下載或放置對應平台的 `browseforge-runtime-chromium-v0.1.1-alpha.0-<platform>.zip`。
2. 解壓到 BrowseForge runtime 目錄，例如 `browsers/browseforge-chromium/`。
3. 在 `config.json` 啟用 runtime：

```json
{
  "default_runtime_id": "camoufox",
  "runtimes": {
    "browseforge-chromium": {
      "enabled": true,
      "binary_path": "browsers/browseforge-chromium/chrome",
      "family": "chromium",
      "display_name": "BrowseForge Chromium",
      "settings": {
        "auto_safe_gpu_fallback": true,
        "isolated_runtime_cache": true,
        "repair_transient_cache_on_launch_failure": true,
        "fingerprint_platform": "auto",
        "target_platform_policy": "warn",
        "native_mode": "enabled",
        "plugins_pdf": "enabled",
        "extra_args": []
      }
    }
  }
}
```

`binary_path` 要指向 artifact 內的瀏覽器 binary，不是 standalone wrapper：

| 平台 | `binary_path` 指向 |
|------|--------------------|
| Linux x64 | `browsers/browseforge-chromium/chrome` |
| macOS arm64 | `browsers/browseforge-chromium/Chromium.app/Contents/MacOS/Chromium` |
| Windows x64 | `browsers/browseforge-chromium/chrome.exe` |

建立 profile 時指定 `runtime_id`：

```bash
curl -X POST http://127.0.0.1:19280/api/profiles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Chromium Alpha Profile",
    "runtime_id": "browseforge-chromium",
    "proxy": {
      "type": "socks5",
      "host": "proxy.example.com",
      "port": 1080,
      "username": "user",
      "password": "pass",
      "region": "us-ny"
    }
  }'
```

`proxy.region` 是去識別化地理標籤，用來讓 native WebRTC persona 和 proxy 出口地區保持一致；不要填入 raw IP、帳密或可識別個資。

### BrowseForge 傳入的 native switches

| Switch | 來源 | 說明 |
|--------|------|------|
| `--fingerprint=SEED` | `profile.FingerprintSeed` | 穩定身份主種子 |
| `--fingerprint-user-agent` / UA-CH switches | 指紋池 + Chromium version | UA、full version、platform、architecture、bitness 一致性 |
| `--fingerprint-timezone` | proxy/GeoIP profile | JS timezone 與出口地區一致性 |
| `--fingerprint-locale` / `--fingerprint-accept-language` | proxy/GeoIP profile | navigator language 與 HTTP header 一致性 |
| `--fingerprint-hardware-concurrency` / `--fingerprint-device-memory` | 指紋池 | CPU / memory plausible range |
| `--fingerprint-screen-*` | 指紋池 | screen / avail / viewport coherence |
| `--fingerprint-canvas-noise` | 指紋池 / seed | Canvas 穩定 noise |
| `--fingerprint-audio-noise` | 指紋池 / seed | AudioContext 穩定 noise |
| `--fingerprint-fonts-list` | 指紋池 / font pack | 字型清單與目標 OS 一致性 |
| `--fingerprint-webgl-vendor` / `--fingerprint-webgl-renderer` | 指紋池 / GPU profile | WebGL vendor/renderer |
| `--browseforge-stealth-config` | BrowseForge native persona JSON | persona hash、origin salt、proxy/WebRTC metadata |
| `--browseforge-stealth-mode=enabled` | `runtimes.browseforge-chromium.settings.native_mode` | 啟用 native stealth substrate |

### 啟用前驗證

使用端啟用前至少確認：

- `GET /api/runtimes` 會列出 `browseforge-chromium` 且 `enabled=true`。
- `POST /api/profiles` 可建立 `runtime_id=browseforge-chromium` 的 profile。
- `POST /api/sessions` 回傳的 session `runtime_id` 仍是 `browseforge-chromium`。
- Playwright Bind endpoint 可連線第二個 client。
- MCP `list_runtimes`、`create_profile`、`open_browser` 可用。
- workflow `create_profile` 會原樣保留 `runtime_id`。
- Docker / server 部署時，artifact 已 seed 到 `/app/browsers/browseforge-chromium` 或 config 指到正確 binary。

### 已知限制

- 目前 artifact 是 unsigned alpha；可用於 validation / dogfood，不可宣稱 production-ready。
- release-grade 仍需要 Linux release asset signing policy、macOS Developer ID + notarization、Windows Authenticode。
- `download_url` 正式化前，自動 installer 應使用本地路徑、Docker seed 或明確的 alpha asset URL。
- 不建議把 `browseforge-chromium` 設成 `default_runtime_id`，除非 operator 明確接受 alpha runtime 風險。

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
