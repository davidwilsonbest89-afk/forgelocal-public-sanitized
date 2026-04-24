# 行為擬態功能 — 研究報告

> 研究日期：2026-04-24
> 目的：為 BrowseForge 實作行為擬態功能提供事實依據

---

## 一、偵測端分析的行為信號

現代反機器人系統（DataDome、Akamai、GeeTest、reCAPTCHA v3）建立「行為指紋」，分析以下維度：

### 滑鼠軌跡
- 路徑曲率（人類走弧線，不走直線）
- 速度剖面：加速 → 巡航 → 減速（Gaussian 速度分佈）
- 微修正（長距離移動中更頻繁）
- 過衝與修正（尤其對小/遠目標）
- 閒置時的微漂移（fidget movements）
- 點擊前的懸停時間
- 方向性差異：向上推 vs 向下拉的加速度不同（DMTG 論文發現）

### 鍵盤節奏
- 按鍵間隔遵循 **log-normal 分佈**（非均勻、非 Gaussian）
- 常見字母組合（th、ing）打得更快
- 不熟悉的字詞有更長停頓
- 打字錯誤率：**熟練打字者 2-5%**
- 錯誤修正模式：立即退格 or 多打幾個字後才退格
- 自然斷點處的思考停頓（句首、段落間）
- 雙字母第二個幾乎總是打得更快

### 捲動行為
- 速度隨內容複雜度變化
- 桌面端：滾輪增量步進
- 閱讀時間與內容長度相關
- 探索行為：偶爾的切線點擊、返回、重開分頁

### 點擊精度
- 自然的瞄準不精確（不會打中像素中心）
- 偶爾的點偏 + 修正
- 點擊前的猶豫（懸停時間）
- mousedown → mouseup 間隔自然變化

---

## 二、學術研究關鍵發現

### DMTG（arXiv:2410.18233）— 擴散模型滑鼠軌跡生成

**核心發現：**
- 人類軌跡存在於「純隨機」和「純直線」之間，有目的但帶有機變化
- 人類操作本質上不可重現 — 同一任務每次產生不同軌跡
- 複雜度 ≠ 真實度：GAN 生成的複雜曲線反而更容易被偵測

**偵測準確率比較（越低 = 越像人類）：**

| 生成器 | 獨立判別器準確率 | 統一判別器準確率 |
|--------|-----------------|-----------------|
| DMTG | 87-92% | 73-88% |
| ghost-cursor | 89-96% | 84-94% |
| GAN (BeCAPTCHA) | 99.4-99.9% | 93-97% |
| 純 Bézier | 96.2-98.7% | — |
| 直線 | 90.5-99.8% | — |

**分佈相似度（JSD，越低越好）：**
- DMTG: 0.3025
- ghost-cursor: 0.3565
- GAN: 0.5865

**結論：ghost-cursor 的 Bézier 方法已經相當好（JSD 僅比 DMTG 差 15%），且實作簡單得多。**

### BeCAPTCHA-Mouse（arXiv:2005.00890）— 神經運動模型

**Sigma-Lognormal 模型：** 人類滑鼠移動的速度剖面可分解為多個 Lognormal 形狀的原始筆劃。

每個筆劃 6 個參數：D（振幅）、t0（發生時間）、μ（對數時間延遲）、σ（對數響應時間）、θs（起始角）、θe（結束角）

**人類運動特徵：**
- 初始加速 + 末端減速（拮抗肌 vs 主動肌）
- 軌跡末端的精細修正（低速接近目標以提高精度）
- 單一軌跡即可達 93% 偵測準確率

**速度剖面：Gaussian 分佈最像人類**（先加速後減速），優於常數速度和對數速度。

### 鍵盤節奏攻防（arXiv:2601.17280）

- 用 LSTM 生成的偽造打字節奏，對五種分類器的逃避率達 **99.8%+**
- 說明鍵盤節奏模擬在技術上完全可行
- 關鍵：使用 log-normal 分佈而非均勻分佈

---

## 三、現有工具分析

### ghost-cursor（Node.js，Puppeteer/Playwright）

**滑鼠軌跡演算法：**
1. 三次 Bézier 曲線（4 控制點：起點、2 錨點、終點）
2. 兩個錨點在直線的**同一側**（避免 S 形，產生自然弧線）
3. 錨點偏移量 = clamp(距離, 2, 200) 像素
4. 步數由 Fitts's Law 決定：`steps = ceil((log₂(2*log₂(arcLen/width+1) + 1) + speed*25) * 3)`
5. 典型範圍：25-80 步

**過衝機制：**
- 距離 > 500px 時觸發
- 過衝點 = 目標 + 半徑 120px 圓內均勻隨機點
- 修正路徑使用 spread=10（很直的路徑）

**速度剖面：** 沿弧長均勻分佈（大致等速），無加速/減速曲線

**缺少的：** 無 jitter、無 Gaussian 噪聲、無微停頓、無速度 easing

**API：** 使用 CDP `Input.dispatchMouseEvent`（非 page.mouse.move）

### humanization-playwright（Python）

**滑鼠：** 三次 Bézier，控制點在 33%/66% 處，偏移 min(50, dist*0.2)px，Gaussian jitter σ=1px，5% 微停頓機率

**打字：** 預設 600 CPM（100ms/字），±20ms 偏移，空白鍵額外 50-100ms，keydown→keyup 100-140ms。**無打字錯誤模擬。**

**點擊：** mousedown→mouseup 50-80ms（快）/ 100-150ms（慢）

**捲動：** 均勻步進 ±10% 噪聲，**無慣性、無 easing**

**缺少的：** 無打字錯誤、無速度剖面、無捲動慣性、無標點特殊處理

---

## 四、playwright-go 可用 API

| API | 關鍵參數 | 時序控制 |
|-----|---------|---------|
| Mouse.Move(x, y) | Steps *int | Steps 控制插值點數 |
| Mouse.Click(x, y) | Button, ClickCount, Delay | Delay = down/up 間隔 ms |
| Mouse.Down() / Up() | Button, ClickCount | 無 |
| Mouse.Wheel(dX, dY) | deltaX, deltaY float64 | 無（fire and forget） |
| Keyboard.Press(key) | Delay *float64 | Delay = down/up 間隔 ms |
| Keyboard.Down(key) / Up(key) | key string | 無 |
| Keyboard.Type(text) | Delay *float64 | Delay = 每字間隔 ms |
| Locator.BoundingBox() | Timeout | 取得元素座標 |
| CDPSession.Send() | method, params | 完整 CDP 協議（僅 Chromium） |

**關鍵限制：** Mouse.Move 的 Steps 只做線性插值，不支援自訂路徑。要走 Bézier 曲線，必須自己算好每個點，逐點呼叫 Mouse.Move(x, y)。

**CDP：** 僅 Chromium 可用。Firefox 不支援 CDP，必須用 Playwright 原生 API。因此擬態實作應基於 Playwright API（Mouse.Move/Down/Up），確保兩引擎通用。

---

## 五、實作策略建議

### 取 ghost-cursor 的優點 + 補足缺陷

| 面向 | ghost-cursor 做法 | 改進 |
|------|-------------------|------|
| 路徑形狀 | Bézier 同側錨點 ✅ | 沿用 |
| 步數計算 | Fitts's Law ✅ | 沿用 |
| 過衝 | >500px 觸發 ✅ | 沿用，閾值可調 |
| 速度剖面 | 等速 ❌ | 改用 Gaussian easing（先慢後快再慢） |
| Jitter | 無 ❌ | 加 Gaussian(0, 1px) |
| 微停頓 | 無 ❌ | 加 5% 機率 30-70ms |
| 打字 | 無 | log-normal 間隔 + 2% 錯字率 |
| 捲動 | 基本 | 加 ease-out 減速 |

### 架構：在 Playwright API 層包裝

```
BrowseForge API/MCP 層
  → humanize 模組（新增）
    → Playwright API（Mouse.Move, Keyboard.Press, Mouse.Wheel）
      → Firefox / Chromium
```

擬態邏輯全部用 Playwright 原生 API 實作，不依賴 CDP，確保兩引擎通用。

### 預設啟用

- `humanize.enabled` 預設 `true`
- 提供 `"fast"` / `"normal"` / `"slow"` 預設速度
- 可在 profile 層級覆蓋全域設定
