# WebGL 指紋策略

## 問題

Camoufox 支援完整的 C++ 層級 WebGL 偽裝，但需要提供**完整**的 WebGL 資料才能正確運作：

- `webGl:renderer` — GPU 名稱字串
- `webGl:vendor` — 廠商字串
- `webGl:supportedExtensions` — 支援的 WebGL 擴充清單
- `webGl:parameters` — 所有 GL enum 參數值
- `webGl:shaderPrecisionFormats` — Shader 精度格式

如果只提供 renderer/vendor 而不提供其他欄位，Camoufox 只會偽裝字串，但 `getParameter()` 等 API 仍回傳真實值。這會造成**不一致**（宣稱是 NVIDIA GPU 但參數是軟體渲染器），被 WAF 偵測。

## BrowseForge 的策略

```
有完整 WebGL profile（含 supportedExtensions + parameters）
  → 全部傳給 Camoufox

只有 renderer/vendor（不完整）
  → 不傳任何 WebGL 欄位
  → Camoufox 自動用 BrowserForge 生成完整一致的指紋
```

這確保了：
- 有完整資料時：使用我們指定的 GPU 指紋
- 沒有完整資料時：Camoufox 自己生成一致的指紋（不會被偵測為不一致）

## 如何收集完整 WebGL Profile

1. 在有真實 GPU 的機器上開啟 `scripts/webgl-collector.html`
2. 點擊「擷取 WebGL 指紋」
3. 將 JSON 存入 `data/webgl-profiles/webgl-profiles.json`
4. 重新執行 `node scripts/generate-fingerprints.js` 生成指紋池

## WebGL Profile 格式

```json
[
  {
    "match": "Apple",
    "webGl:supportedExtensions": [...],
    "webGl:contextAttributes": {...},
    "webGl:parameters": {...},
    "webGl:shaderPrecisionFormats": {...}
  }
]
```

`match` 欄位用於配對 — 當指紋的 `webGl:renderer` 包含此字串時，附加完整的 WebGL 資料。

## CloakBrowser 的差異

CloakBrowser (Chromium) 使用 `--fingerprint=SEED` 種子機制，在渲染管線層面生成一致的假 WebGL 輸出，不需要真實 GPU 資料。這是架構上的根本差異。

## 參考

- [Camoufox WebGL 文件](https://camoufox.com/fingerprint/webgl/)
- [Camoufox WebGL 研究進度](https://camoufox.com/webgl-research/)
