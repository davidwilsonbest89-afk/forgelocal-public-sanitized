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
| Runtime 狀態 | `v0.1.5-alpha.0` |
| 發佈狀態 | alpha validation artifact；尚未 release-grade 簽章 |

### 使用端整合方式

1. 下載或放置對應平台的 `browseforge-runtime-chromium-v0.1.5-alpha.0-<platform>.zip`。
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
| `--fingerprint-user-agent` / UA-CH switches | resolved runtime platform + compatible 指紋池值 | UA、full version、platform、architecture、bitness 一致性；native Linux arm64 會輸出 Linux/arm/64，而不是 x86 pool 值 |
| `--fingerprint-timezone` | proxy/GeoIP profile | JS timezone 與出口地區一致性 |
| `--fingerprint-locale` / `--fingerprint-accept-language` | proxy/GeoIP profile + compatible 指紋池值 | navigator language 與 HTTP header 一致性；profile 語言和 GeoIP locale 不一致時以 GeoIP locale 為準 |
| `--fingerprint-hardware-concurrency` / `--fingerprint-device-memory` | 指紋池 | CPU / memory plausible range；同一組值也寫入 native persona JSON |
| `--fingerprint-screen-*` | 指紋池 | screen / avail / viewport coherence；avail 大於 screen 時會修正為不超過 screen |
| `--fingerprint-canvas-noise` | 指紋池 / seed | Canvas 穩定 noise |
| `--fingerprint-audio-noise` | 指紋池 / seed | AudioContext 穩定 noise |
| `--fingerprint-fonts-list` | 指紋池 / font pack | 字型清單與目標 OS 一致性 |
| `--fingerprint-webgl-vendor` / `--fingerprint-webgl-renderer` | 指紋池 / GPU profile | WebGL vendor/renderer；同一組值也寫入 native persona JSON |
| `--browseforge-stealth-config` | BrowseForge PersonaContract JSON | persona hash、origin salt，以及 browser/platform/locale/network/DNS/geolocation/hardware/screen/GPU/fonts/canvas/math/geometry/audio/plugins/media/permissions/WebRTC/storage/realm metadata |
| `--browseforge-stealth-mode=enabled` | `runtimes.browseforge-chromium.settings.native_mode` | 啟用 native stealth substrate |

BrowseForge Chromium 只接受 coherent native persona platform 組合：Windows=`Win32`/x86/64、macOS=`MacIntel`/x86 或 arm/64、Linux x64=`Linux x86_64`/x86/64、Linux arm64=`Linux aarch64`/arm/64。若指紋池值和 resolved runtime platform/locale 衝突，BrowseForge 會捨棄不相容值並由 canonical persona 重新產生 switch 與 native JSON。

### PersonaContract 與 detector regression

`browseforge-chromium` launch path 會先產生 PersonaContract，再寫入 `BrowseForgeNative/persona.json`。PersonaContract 是 BrowserLeaks / BrowserScan / CreepJS / SannySoft smoke 的單一比對基準，至少涵蓋：

- Browser identity：UA、Chrome/Chromium brands、full version list、UA-CH / Sec-CH-* platform、arch、bitness、mobile、model、form factors。
- Locality：timezone、timezone offset、locale、Accept-Language、navigator language/languages、Sec-CH-Lang。
- Network：proxy region、country/region metadata、DNS resolver policy、WebRTC redaction policy、geolocation policy。
- Device/rendering：hardwareConcurrency、deviceMemory、screen/avail/outer/inner/viewport/DPR/touch/orientation、GPU mode、WebGL/WebGL2/WebGPU profile、fonts/font metrics/emoji、canvas/text/emoji、math intrinsic sample、client rect geometry、audio、plugins/MIME/PDF、codecs/media devices、permissions、storage quota。
- Realm policy：top window、same-origin iframe、sandbox/nested iframe、workers、service worker、OffscreenCanvas worker 都必須使用同一份 contract。

Launch 前會 fail closed 下列不一致 tuple：UA 與 platform 不符、UA-CH platform/arch/bitness 與 JS platform 不符、locale / Accept-Language / navigator.languages / Sec-CH-Lang 不一致、screen/avail/inner/outer/viewport/DPR/touch tuple 不一致、desktop persona 宣告 mobile form factor、proxy persona 缺 region/country metadata、proxy persona 未使用 proxy-aligned DNS/geolocation/WebRTC direct-IP redaction、non-proxy persona 夾帶 proxy metadata、IP country 與 timezone 不一致、zh/ja/ko locale 缺 CJK font profile、啟用 PDF profile 時缺 Chromium plugin/MIME entries、macOS persona 宣告 SwiftShader renderer、缺 service-worker realm target、停用 stable math 或 client-rect policy。

Docker GPU mode 是 PersonaContract 的明確輸入，不允許自動猜測：`software` 產生 SwiftShader-aligned WebGL profile 並加上 SwiftShader launch flags；`native` 保留 browser-default GPU evidence；`passthrough` 僅表示 operator 已明確提供 host GPU passthrough。其他 `BROWSEFORGE_DOCKER_GPU_MODE` 值會在 entrypoint / launch 前 fail closed，避免 linux/arm64 detector evidence 在 software/native/passthrough 間靜默漂移。

半自動 detector harness：

```bash
node scripts/detector-harness.js collector > /tmp/browseforge-detector-collector.js
node scripts/detector-harness.js compare profiles/<profile>/browser-data/BrowseForgeNative/persona.json /tmp/detector-sample.json
node scripts/detector-harness.js selftest
node scripts/detector-harness.js matrix > /tmp/browseforge-detector-matrix.json
```

`collector` 產生可透過 BrowseForge REST eval / Playwright console 執行的 browser-side collector，收集：

- Realm：top window、same-origin iframe、sandbox iframe、fragment iframe、nested iframe、detached iframe（若 runtime 允許）、dedicated worker、shared worker、service worker controller、OffscreenCanvas/WebGL。
- Identity/locality：navigator、UAData high entropy（brands、fullVersionList、platformVersion、architecture、bitness、model、formFactors）、Intl/Date timezone、screen/viewport/DPR/orientation。
- Rendering/capability：canvas、math intrinsic hash、client-rect geometry hash、WebGL/WebGL2/WebGPU limits/extensions/shader precision、plugins/MIME、storage、geolocation API/permission state、media device API、codec support、Notification permission/query state、webdriver attributes、Selenium/WebDriver/CDP globals 與 `window.chrome` shape。

`compare` 只比對實際收集到且 PersonaContract 有宣告的欄位，並檢查 BrowserLeaks/BrowserScan 可抄錄的 optional HTTP headers（User-Agent、Accept-Language、Sec-CH-UA、Sec-CH-UA-Full-Version-List、Sec-CH-UA-Platform、Sec-CH-UA-Platform-Version、Sec-CH-UA-Arch、Sec-CH-UA-Bitness、Sec-CH-UA-Mobile、Sec-CH-UA-Model、Sec-CH-UA-Form-Factors、Sec-CH-Lang）與 realm parity：UA、platform、language/languages、hardwareConcurrency、deviceMemory、Intl timezone/locale、Date offset、DPR、WebGL vendor/renderer、canvas sample；optional realm 若 runtime 不允許存取會標成 unsupported 並略過。`matrix` 輸出每個 public detector target 的必收 artifact checklist，並附上 `document.readyState=complete`、DOM/resource stable window >= 3000ms、最小 text/node count 的穩定結果條件；沒有跑過 public detector 時不可宣稱 pass。

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
- BrowseForge Chromium 預設 font contract 是 `persona-default` deterministic font metadata：不設定 `runtimes.browseforge-chromium.settings.fonts_dir` 時，即使 fingerprint pool 內有 `fonts` 陣列，也不會輸出 `--fingerprint-fonts-list`，避免宣稱 runtime 尚未提供的 font metrics/raster coherence。只有 operator 明確提供 font corpus 目錄時才允許 ASCII、單一 family <=128 bytes、整體 <=8192 bytes 的 explicit font allowlist。
- BrowseForge Chromium proxy profile 必須提供 redacted `proxy.region`（例如 `us-ny`、`tw-taipei`），且 region 必須可映射到 country metadata；沒有 region 的 proxy 會在 browser launch 前 fail closed，避免 Playwright 實際走 proxy 但 PersonaContract/DNS/geolocation 仍宣告 local。

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

| 網站 / source | 測試重點 | 必收 artifact |
|------|---------|-------------|
| https://bot.sannysoft.com/ | HeadlessChrome、webdriver、Selenium/WebDriver/CDP globals、webdriver DOM attributes、plugins/MIME/PDF、permissions/Notification、language、WebGL、h264 | result table / screenshot |
| https://browserleaks.com/client-hints | UA-CH / Sec-CH 與 navigator UAData | visible fields / JSON |
| https://browserleaks.com/webgl | WebGL vendor/renderer、extensions、limits、shader precision | WebGL report |
| https://browserleaks.com/canvas | Canvas / text / emoji rendering stability | canvas hash/report |
| https://browserleaks.com/fonts | font set、font metrics、emoji font | font report |
| https://browserleaks.com/webrtc | WebRTC host/container/public IP leak | WebRTC report |
| https://browserleaks.com/javascript | JavaScript feature shape、screen、navigator、storage、geolocation API/permission、client rects | JavaScript report |
| https://www.browserscan.net/ | Browser、Location、IP、Hardware、Software trust coherence | report sections |
| https://www.browserscan.net/client-hints | UA-CH / JS UAData parity | Client Hints report |
| https://www.browserscan.net/dns-leak | DNS resolver geolocation/ASN leak | DNS leak report |
| https://www.browserscan.net/webrtc | WebRTC host/container/public IP leak | WebRTC report |
| https://www.browserscan.net/bot-detection | automation/headless signals | bot detection report |
| https://iphey.com/ | Browser、Location、IP、Hardware、Software trust score | main score sections |
| https://abrahamjuliot.github.io/creepjs/ | prototype lies、realm parity、fonts/canvas/audio/math/screen/client-rects | main result |
| https://abrahamjuliot.github.io/creepjs/tests/workers.html | worker parity | worker test result |
| https://abrahamjuliot.github.io/creepjs/tests/iframes.html | iframe parity | iframe test result |
| https://abrahamjuliot.github.io/creepjs/tests/prototype.html | prototype/native descriptor parity | prototype test result |
| https://pixelscan.net | 指紋一致性、IP / timezone / WebRTC / OS mismatch | fingerprint / bot checker result |

---

## 參考連結

- [CloakBrowser GitHub](https://github.com/CloakHQ/CloakBrowser)
- [CloakBrowser 文件](https://cloakbrowser.dev/)
- [Camoufox 指紋文件](https://camoufox.com/fingerprint/)
- [Camoufox WebGL 研究](https://camoufox.com/webgl-research/)
- [BrowserForge](https://github.com/daijro/browserforge)
