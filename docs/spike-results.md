# BrowseForge Spike Results

## 0.1.4 Playwright 連接驗證
- **日期**: 2026-04-23
- **假設**: playwright-go 能啟動 Camoufox 並透過 CAMOU_CONFIG 注入指紋
- **版本**: Camoufox v135.0.1-beta.24 (x86_64) + playwright-go v0.5700.1 (PW 1.57)
- **結果**: ✅ PASS
- **發現**:
  - playwright-go v0.5700.1（PW 1.57）與 Camoufox v135 完全相容，不需要退回舊版
  - CAMOU_CONFIG env var 成功注入 navigator.userAgent 和 navigator.platform
  - headless 模式正常
  - 首次執行會下載 Playwright driver（~100MB），後續不需要
- **決策**: 使用 playwright-go v0.5700.1（最新版）

## 0.1.1 Container 隔離 — 待測
## 0.1.2 proxy.onRequest + cookieStoreId — 待測
## 0.1.3 Camoufox Portable 模式 — 待測
## 0.1.5 Container 數量 + 記憶體 — 待測
## 0.1.7 Content Script → Setter
- **日期**: 2026-04-23
- **假設**: Camoufox v135 暴露 window.setXxx() setter 供 per-context 指紋覆寫
- **結果**: ❌ FAIL — v135 沒有任何 setter 函數
- **發現**:
  - 14 個 setter（setCanvasSeed, setNavigatorPlatform 等）全部不存在
  - 這些 setter 來自 cloverlabs-camoufox fork，不在 daijro/camoufox v135 中
  - v146 pre-release 可能有，但只有 ARM64 binary（本機 x86_64 無法測試）
- **決策**: Phase 1 改用**多實例方案**（每 profile 一個 Camoufox process + 獨立 CAMOU_CONFIG）
  - CAMOU_CONFIG 已驗證可用（spike 0.1.4）
  - 每個實例天然隔離（cookie、storage、指紋全部獨立）
  - 記憶體較高但架構簡單可靠
  - Phase 2 fork 後再加 setter 或 per-container 支援

## Phase 2 Gecko Build
- **日期**: 2026-04-23
- **假設**: 可以在 Docker on macOS 中 build Gecko
- **結果**: ❌ FAIL — 多次環境問題（clang 版本、libclang、Node.js、WASM、header 大小寫）
- **發現**:
  - Camoufox 官方 CI 用 Ubuntu 24.04 原生 + `make mozbootstrap` + `multibuild.py`
  - Docker on macOS 的 filesystem 和環境差異導致反覆失敗
- **決策**: Phase 2 改用 **GitHub Actions CI build**
  - Fork Camoufox repo → 加入 ContainerConfig.hpp + per-container patch
  - Push 到 GitHub → CI 自動 build 三平台 binary
  - 下載 CI 產出的 binary 替換 dist/browsers/camoufox/
  - 不在本機 build Gecko
## 0.1.8 Playwright Container Tab — 待測
## 0.1.9 CDP 支援 — 待測
