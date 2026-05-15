# BrowseForge 平台支援矩陣

[English](platform-support.md)

> 維護此表以確保每個平台都有對應的瀏覽器引擎可用。
> 更新時機：瀏覽器引擎發佈新版本時。

## 當前版本

| 元件 | 版本 | 更新日期 |
|------|------|---------|
| BrowseForge | v1.7.3 | 2026-05-15 |
| Camoufox | v135.0.1-beta.24 | 2025-03-15 |
| CloakBrowser | chromium-v145.0.7632.109.2 | 2026-03-04 |

## 平台支援矩陣

| 平台 | BrowseForge | 🦊 Camoufox | 🌐 CloakBrowser | 備註 |
|------|:---:|:---:|:---:|------|
| macOS x64 (Intel) | ✅ | ✅ v135 | ✅ v145 | |
| macOS arm64 (Apple Silicon) | ✅ | ✅ v135 | ✅ v145 | |
| Linux x64 | ✅ | ✅ v135 | ✅ v145 | 需要 xvfb |
| Linux arm64 | ✅ | ✅ v135 | ⚠️ v146 only | CloakBrowser 需用 v146 版本 |
| Windows x64 | ✅ | ✅ v135 | ✅ v145 | |
| Windows i686 (32-bit) | ❌ | ✅ v135 | ❌ | BrowseForge 不提供 32-bit build |
| Linux i686 (32-bit) | ❌ | ✅ v135 | ❌ | 同上 |

## 下載 URL 對照表

### Camoufox (v135.0.1-beta.24)

```
macOS x64:    https://github.com/daijro/camoufox/releases/download/v135.0.1-beta.24/camoufox-135.0.1-beta.24-mac.x86_64.zip
macOS arm64:  https://github.com/daijro/camoufox/releases/download/v135.0.1-beta.24/camoufox-135.0.1-beta.24-mac.arm64.zip
Linux x64:    https://github.com/daijro/camoufox/releases/download/v135.0.1-beta.24/camoufox-135.0.1-beta.24-lin.x86_64.zip
Linux arm64:  https://github.com/daijro/camoufox/releases/download/v135.0.1-beta.24/camoufox-135.0.1-beta.24-lin.arm64.zip
Windows x64:  https://github.com/daijro/camoufox/releases/download/v135.0.1-beta.24/camoufox-135.0.1-beta.24-win.x86_64.zip
```

### CloakBrowser (chromium-v145.0.7632.109.2)

```
macOS x64:    https://github.com/CloakHQ/CloakBrowser/releases/download/chromium-v145.0.7632.109.2/cloakbrowser-darwin-x64.tar.gz
macOS arm64:  https://github.com/CloakHQ/CloakBrowser/releases/download/chromium-v145.0.7632.109.2/cloakbrowser-darwin-arm64.tar.gz
Linux x64:    https://github.com/CloakHQ/CloakBrowser/releases/download/chromium-v145.0.7632.109.2/cloakbrowser-linux-x64.tar.gz
Windows x64:  https://github.com/CloakHQ/CloakBrowser/releases/download/chromium-v145.0.7632.109.2/cloakbrowser-windows-x64.zip
```

### CloakBrowser Linux arm64 (chromium-v146.0.7680.177.3)

```
Linux arm64:  https://github.com/CloakHQ/CloakBrowser/releases/download/chromium-v146.0.7680.177.3/cloakbrowser-linux-arm64.tar.gz
```

## 升級檢查清單

更新瀏覽器版本時：

- [ ] 確認新版本在所有支援平台都有 binary
- [ ] 更新 `internal/browser/download.go` 中的版本號和 URL
- [ ] 更新此表
- [ ] 測試自動下載功能
- [ ] 測試反偵測能力（browserleaks、Sannysoft）
- [ ] 依照 [Release Process](release.md) 執行 `scripts/release-preflight.sh vX.Y.Z`
- [ ] 依照 [Release Process](release.md) 執行 `scripts/release-push.sh vX.Y.Z`
