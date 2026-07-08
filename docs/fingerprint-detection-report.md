# Browser Fingerprint Detection Report

Date: 2026-07-07
Profile: `x@undo.im` (Chromium / CloakBrowser)
Spoofed Identity: Chrome 146.0.7680.177 on Windows 11
Actual Environment: macOS Apple Silicon (arm64)
IP: 123.193.234.206 (kbro CO. Ltd., Taipei, Taiwan)

---

## Test Sites Summary

| Site | Result | Score |
|------|--------|-------|
| [browserleaks.com](https://browserleaks.com) | ✅ Loaded | N/A (sub-page tools) |
| [bot.sannysoft.com](https://bot.sannysoft.com) | ✅ All Passed | WebDriver=missing, all PHANTOM tests ok |
| [CreepJS](https://abrahamjuliot.github.io/creepjs/) | ⚠️ 25% like headless | GPU confidence: moderate |
| [pixelscan.net](https://pixelscan.net/fingerprint-check) | ❌ Inconsistent | Masking detected, font mismatch |
| [iphey.com](https://iphey.com) | ✅ Trustworthy | All sections fine |
| [browserscan.net](https://www.browserscan.net) | ⚠️ 80% Authenticity | -5% WebGL, -5% Audio, -10% Incognito |

---

## Detailed Findings

### 1. BrowserScan (Authenticity: 80%)

**Deduction Breakdown:**

| Item | Deduction | Description |
|------|-----------|-------------|
| WebGL exception | -5% | WebGL fingerprint anomaly detected |
| Audio exception | -5% | "Audio appears to have been modified manually" |
| Incognito mode | -10% | Incognito window detected |

**Root Cause Analysis:**

- **WebGL (-5%)**: CloakBrowser injects a spoofed GPU string (`NVIDIA GeForce RTX 4050 Laptop GPU / Direct3D11`), but the actual rendering behavior (pixel output, parameter limits) doesn't fully match what a real RTX 4050 on Windows would produce. Detection scripts compare the declared renderer against the actual WebGL parameter set and rendering output.

- **Audio (-5%)**: CloakBrowser modifies the AudioContext fingerprint to prevent tracking, but the modification pattern is detectable. The OfflineAudioContext output has been altered in a way that doesn't match any known real-device audio stack.

- **Incognito (-10%)**: BrowseForge launches profiles in Playwright's persistent context with isolation features (no shared cookies, limited service worker scope). This triggers incognito detection heuristics that check for filesystem quota limits, availability of certain storage APIs, and cookie persistence behavior.

**Other Detected Values:**

| Field | Value |
|-------|-------|
| Browser | Chrome 146.0.7680.177 |
| Platform | Windows 11 |
| IP Timezone | Asia/Taipei |
| JS Timezone | Asia/Taipei |
| Screen | 1920×1080 |
| WebRTC | Disabled |
| Bot Detection | No |
| Proxy | No |
| Canvas Hash | 39D86154 |
| WebGL Hash | B808350B |
| Audio Hash | 3F356F8F |
| Visitor ID | 13AF8F36 |

---

### 2. Pixelscan (Result: Inconsistent)

**Issues Detected:**

| Issue | Details |
|-------|---------|
| Fingerprint Status | **Inconsistent** |
| Proxy Masking | Detected |
| Browser Detection | Chrome 146.0.0.0 on Windows **(Incognito Window)** |
| Font List | Only "Arial" detected |

**Root Cause Analysis:**

- **Proxy Masking Detected**: Pixelscan cross-references IP geolocation with other signals. While no proxy is actually in use, the fingerprint inconsistency between the spoofed Windows environment and the actual rendering behavior triggers this flag.

- **Font Mismatch (Critical)**: Only Arial is available in the font list. A genuine Windows 11 system would expose 50+ system fonts (Segoe UI, Calibri, Consolas, Times New Roman, etc.). Running a Windows-spoofed profile on macOS means only cross-platform fonts are available, which is a strong signal of environment mismatch.

- **Browser Version Inconsistency**: Pixelscan detects multiple browser feature support patterns that don't align with a single Chrome version. The feature detection matrix shows mixed signals between Chrome 22-28, Chrome 29+, and other browsers.

**Detected Values:**

| Field | Value |
|-------|-------|
| IP Address | 123.193.234.206 |
| Country | Taiwan |
| City | Taipei |
| ISP | kbro CO. Ltd. |
| Timezone (JS) | Asia/Taipei |
| Screen | 1920x1080 |
| Available Screen | 1920x1032 |
| Font Hash | 6fbda3a3567da6d01bc9da915e91d702 |
| Canvas Hash | 688cd2f8f7ebbe8247bbb2c7a0d02090 |
| WebGL Hash | 7d7eaf96048ba951cc66b71183061825 |
| AudioContext Hash | 9a135583c610257f095623e922936a25 |
| UA (HTTP) | Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 ... Chrome/146.0.0.0 Safari/537.36 |
| UA (JS) | Same as HTTP |
| Platform | Win32 |
| Hardware Concurrency | 8 |
| Languages | zh-TW |

---

### 3. CreepJS (25% Like Headless)

**Headless Detection Breakdown:**

| Check | Result | Triggered |
|-------|--------|-----------|
| noChrome | false | ❌ |
| hasPermissionsBug | false | ❌ |
| noPlugins | false | ❌ |
| noMimeTypes | false | ❌ |
| notificationIsDenied | false | ❌ |
| hasKnownBgColor | false | ❌ |
| prefersLightColor | true | ✅ (benign) |
| uaDataIsBlank | false | ❌ |
| pdfIsDisabled | false | ❌ |
| noTaskbar | false | ❌ |
| hasVvpScreenRes | false | ❌ |
| hasSwiftShader | false | ❌ |
| noWebShare | false | ❌ |
| **noContentIndex** | **true** | ✅ Headless signal |
| **noContactsManager** | **true** | ✅ Headless signal |
| **noDownlinkMax** | **true** | ✅ Headless signal |

**Stealth Detection (0% — All Passed):**

| Check | Result |
|-------|--------|
| hasIframeProxy | false |
| hasHighChromeIndex | false |
| hasBadChromeRuntime | false |
| hasToStringProxy | false |
| hasBadWebGL | false |

**Headless Detection (0% — All Passed):**

| Check | Result |
|-------|--------|
| webDriverIsOn | false |
| hasHeadlessUA | false |
| hasHeadlessWorkerUA | false |

**GPU Analysis:**

| Field | Value |
|-------|-------|
| Vendor | Google Inc. (NVIDIA) |
| Renderer | ANGLE (NVIDIA, NVIDIA GeForce RTX 4050 Laptop GPU (0x000028A1) Direct3D11 vs_5_0 ps_5_0, D3D11) |
| Confidence | **moderate** |

**Root Cause Analysis:**

- **noContentIndex / noContactsManager / noDownlinkMax**: These Web APIs (Content Index API, Contact Picker API, NetworkInformation.downlinkMax) are typically available in standard Chrome on Android/desktop but absent in Playwright-controlled browsers. Their absence is a known automation signal.

- **GPU confidence: moderate**: The declared GPU (RTX 4050 / D3D11) doesn't produce WebGL rendering output that fully matches known RTX 4050 signatures in CreepJS's database. The parameters are correct, but pixel-level rendering via ANGLE emulation differs from native D3D11 output.

**Other CreepJS Values:**

| Field | Value |
|-------|-------|
| FP ID | 3fa591041b616f41dca35ea35f24c7a629c3a3d2bc886f175ed988f7312b4351 |
| Timezone | Asia/Taipei (-480) |
| Intl | zh-TW |
| Screen | 1920×1080, avail 1920×1032 |
| Touch | false |
| Depth | 32 |
| Fonts loaded | 0/51 (Like undefined) |
| Speech voices (local) | Microsoft David, Microsoft Zira, Microsoft Mark |
| Speech voices (remote) | Google US English, Google UK English Female/Male |
| Features Version | v114-115 (JS/DOM), v114-115 (CSS) |
| Math Engine | Chromium |
| Canvas hash | 2dec16e9b7 |
| Audio sum | 124.04359175392165 |

---

### 4. iphey.com (Result: Trustworthy ✅)

iphey.com gave a **"Trustworthy"** verdict with no issues flagged in:
- Browser detection
- Location consistency
- Hardware fingerprint
- Software fingerprint

This suggests iphey's detection heuristics are less aggressive than BrowserScan/Pixelscan/CreepJS, or that CloakBrowser's fingerprint passes their specific threshold.

---

### 5. bot.sannysoft.com (Result: All Passed ✅)

| Test | Result |
|------|--------|
| User Agent | Mozilla/5.0 (Windows NT 10.0; Win64; x64) ... Chrome/146.0.0.0 |
| WebDriver (New) | missing ✅ |
| Chrome (New) | present ✅ |
| Permissions | prompt |
| Plugins Length | 5 |
| Languages | zh-TW |
| WebGL Vendor | Google Inc. (NVIDIA) |
| WebGL Renderer | ANGLE (NVIDIA, ...) |
| Broken Image | 16x16 ✅ |
| PHANTOM_UA | ok |
| PHANTOM_PROPERTIES | ok |
| PHANTOM_ETSL | ok |
| PHANTOM_LANGUAGE | ok |
| PHANTOM_WEBSOCKET | ok |
| MQ_SCREEN | ok |
| PHANTOM_OVERFLOW | ok |

CloakBrowser fully passes bot.sannysoft.com's Playwright/Puppeteer automation detection suite.

---

## Vulnerability Priority Matrix

| Priority | Weakness | Affected Sites | Root Cause | Fix Complexity |
|----------|----------|----------------|------------|----------------|
| 🔴 High | Incognito mode detection | BrowserScan (-10%) | Playwright persistent context isolation behavior | Medium |
| 🔴 High | Font list too sparse | Pixelscan (inconsistent) | macOS lacks Windows system fonts; only cross-platform fonts exposed | High |
| 🟡 Medium | Audio fingerprint modified | BrowserScan (-5%) | CloakBrowser AudioContext spoofing pattern detectable | Medium |
| 🟡 Medium | WebGL parameter mismatch | BrowserScan (-5%), CreepJS (moderate) | Declared GPU vs actual ANGLE rendering inconsistency | High |
| 🟡 Medium | Missing Web APIs | CreepJS (25% like headless) | ContentIndex, ContactsManager, downlinkMax absent in Playwright | Low-Medium |
| 🟠 Low | Browser feature matrix inconsistency | Pixelscan (inconsistent) | Feature detection doesn't perfectly align with declared Chrome version | High |

---

## Remediation Status (BrowseForge)

| Area | Status | Notes |
| --- | --- | --- |
| Platform coherence | Implemented | `runtimes.cloakbrowser.settings.fingerprint_platform` is a typed config field. `auto` follows CloakBrowser wrapper defaults: macOS host emits `macos`, other hosts emit `windows`. |
| Font coherence | Implemented with explicit input | `runtimes.cloakbrowser.settings.fonts_dir` emits `--fingerprint-fonts-dir=<dir>` after directory validation. Windows identities still need an operator-provided Windows font pack. |
| Storage quota | Implemented as explicit tradeoff | `runtimes.cloakbrowser.settings.storage_quota_mb` emits `--fingerprint-storage-quota=<MB>` when positive. Default `0` leaves CloakBrowser auto behavior unchanged. |
| Managed flag collisions | Implemented | `extra_args` rejects BrowseForge-managed fingerprint flags such as platform, timezone, locale, WebRTC IP, fonts dir, storage quota, screen, and hardware concurrency. |
| Camoufox WebGL completeness | Implemented | Incomplete `webGl:*` / `webGl2:*` profile fragments are removed before `CAMOU_CONFIG`; complete profiles are preserved. Large config payloads use `CAMOU_CONFIG_N` chunks. |
| Audio calibration | Upstream/native issue | BrowseForge no longer treats a non-existent `audio:seed` as a Camoufox control. CloakBrowser audio hash changes require upstream/native calibration or seed quarantine reporting. |
| Missing browser APIs | Platform-conditioned | Do not add JS stubs for ContentIndex, ContactsManager, or NetworkInformation.downlinkMax unless the target Chrome/platform genuinely exposes them. |

## Next Verification Run

Re-test with a coherent target profile:

- macOS host + macOS Chrome-like identity, or Linux/Windows host + Windows Chrome-like identity.
- If using Windows identity on macOS, configure `runtimes.cloakbrowser.settings.fonts_dir` with a licensed Windows-compatible font pack and accept the GPU/runtime mismatch risk explicitly.
- For BrowserScan incognito scoring, compare `storage_quota_mb=0` against `5000` and record FingerprintJS/BrowserScan tradeoff separately.

## Test Methodology

- **Tool**: BrowseForge MCP (BrowseForge v1.10.x)
- **Profile**: `x@undo.im` (Chromium/CloakBrowser engine)
- **Method**: Navigate to each site, wait for networkidle, extract page content via DOM
- **Date**: 2026-07-07
- **Network**: Direct connection (no proxy), kbro ISP, Taipei, Taiwan
