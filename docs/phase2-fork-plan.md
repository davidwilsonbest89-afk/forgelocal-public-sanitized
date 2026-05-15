# Phase 2: Fork Camoufox — Per-Container Fingerprint

> Archive note: this fork plan is retained as research background. The current product contract uses one isolated browser runtime per profile and is documented in [dual-browser-architecture.md](dual-browser-architecture.md).

## 目標
改動 Camoufox 的 C++ 層，讓 MaskConfig 查詢從 process-global 變成 per-userContextId。

## 前置條件
1. Fork https://github.com/daijro/camoufox
2. 搭建 Gecko build 環境（見下方 Docker）
3. 首次完整 build 成功

## 核心改動

### 1. 擴展 MaskConfig.hpp

現有：
```cpp
// additions/camoucfg/MaskConfig.hpp
static std::optional<std::string> GetString(const char* key) {
    // 從 process-global JSON config 查詢
    auto& config = GetConfig();  // std::call_once 載入的 CAMOU_CONFIG
    if (config.contains(key)) return config[key].get<std::string>();
    return std::nullopt;
}
```

改為：
```cpp
static std::optional<std::string> GetString(const char* key, uint32_t userContextId = 0) {
    // 先查 per-container config
    if (userContextId > 0) {
        auto& containerConfigs = GetContainerConfigs();
        auto it = containerConfigs.find(userContextId);
        if (it != containerConfigs.end() && it->second.contains(key)) {
            return it->second[key].get<std::string>();
        }
    }
    // Fallback 到 process-global config
    auto& config = GetConfig();
    if (config.contains(key)) return config[key].get<std::string>();
    return std::nullopt;
}

// Per-container config map
static std::unordered_map<uint32_t, nlohmann::json>& GetContainerConfigs() {
    static std::unordered_map<uint32_t, nlohmann::json> configs;
    return configs;
}

// 動態更新 container config（從 WebExtension 透過 IPC 呼叫）
static void SetContainerConfig(uint32_t userContextId, const std::string& jsonStr) {
    GetContainerConfigs()[userContextId] = nlohmann::json::parse(jsonStr);
}
```

### 2. 每個攔截點加入 userContextId 查詢

以 Navigator.cpp 為例：

現有（Camoufox patch）：
```cpp
if (auto value = MaskConfig::GetString("navigator.platform"))
    return value.value();
```

改為：
```cpp
uint32_t ucid = 0;
if (auto* doc = GetAssociatedDocument()) {
    if (auto* bc = doc->GetBrowsingContext()) {
        ucid = bc->OriginAttributesRef().mUserContextId;
    }
}
if (auto value = MaskConfig::GetString("navigator.platform", ucid))
    return value.value();
```

### 3. 攔截點改動清單（按難度排序）

| 攔截點 | 檔案 | 取得 userContextId 的方式 | 難度 |
|--------|------|--------------------------|------|
| Navigator 屬性 | dom/base/Navigator.cpp | Document → BrowsingContext | 低 |
| Screen 屬性 | dom/base/nsScreen.cpp | Document → BrowsingContext | 低 |
| Canvas noise | dom/canvas/ | CanvasRenderingContext → Document | 低 |
| WebGL params | dom/canvas/WebGLContext* | WebGLContext → Document | 低 |
| Audio noise | dom/media/webaudio/ | AudioContext → Document | 中 |
| Font metrics | layout/generic/ | Frame → Document | 中 |
| User-Agent header | netwerk/protocol/http/ | nsIChannel → LoadInfo → BrowsingContext | 中 |
| Timezone | js/src/jsdate.cpp | JSContext → Realm → Global → Document | 高 |
| Intl API | js/src/builtin/intl/ | 同上 | 高 |

### 4. IPC 機制（WebExtension → C++ 層）

利用 Camoufox 已有的 cross-process-storage.patch（RoverfoxStorage IPC）：

```
WebExtension background.js
  → browser.runtime.sendNativeMessage("camoufox", {
      type: "set_container_config",
      userContextId: 3,
      config: { "navigator.platform": "Win32", "canvas:seed": 12345, ... }
    })
  → C++ Native Messaging handler
  → MaskConfig::SetContainerConfig(3, configJson)
```

## Build 環境（Docker）

```dockerfile
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y \
    build-essential python3 python3-pip git curl \
    libgtk-3-dev libdbus-glib-1-dev libxt-dev \
    nasm yasm zip unzip
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y
ENV PATH="/root/.cargo/bin:${PATH}"
WORKDIR /build
# Clone and build:
# git clone https://github.com/YOUR_FORK/camoufox
# cd camoufox && make dir && make build
```

## 驗證

1. 兩個不同 userContextId 的 tab → Canvas toDataURL 結果不同
2. 兩個不同 userContextId 的 tab → navigator.platform 不同
3. creepjs: lies = 0, trash = 0
4. 原有 CAMOU_CONFIG（process-global）仍正常運作
