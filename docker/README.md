# BrowseForge Docker

一鍵部署 BrowseForge + KasmVNC 遠端桌面。

## 使用方式

建議正式部署 pin 版本 tag，避免 `latest` 在重啟或重新拉取時非預期升級：

```bash
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -e VNC_PASSWORD=browseforge \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.5
```

本地從原始碼 build：

```bash
cd docker
docker compose up -d --build
```

compose 預設會建置 `v1.7.5` release image。要測其他版本：

```bash
BROWSEFORGE_VERSION=v1.7.5 docker compose up -d --build
```

首次啟動需要 3-5 分鐘下載瀏覽器引擎（~440MB）。

## 連線

| 服務 | URL |
|------|-----|
| Dashboard + REST API + Playwright proxy | http://localhost:19280 |
| MCP Streamable HTTP | http://localhost:19281 |
| 遠端桌面 (KasmVNC) | http://localhost:6901 |
| VNC 帳號 | `user` / 環境變數 `VNC_PASSWORD`（預設 `browseforge`） |

## 取得 API Token

```bash
docker compose logs | grep "API Token"
# 或
docker compose exec browseforge /app/BrowseForge token
```

## 特性

- **KasmVNC** — 比 noVNC 更好的剪貼簿支援（Chrome 上 seamless）、IME 中文輸入
- **WebGL 完整偽裝** — GLX + 軟體渲染 + 完整 WebGL 指紋
- **Docker 自動偵測** — 自動啟用 `0.0.0.0` 綁定和 `--no-sandbox`
- **Playwright proxy** — Dashboard + API + Playwright WebSocket proxy 都走 19280
- **MCP HTTP** — Streamable HTTP MCP 走 19281，使用與 REST API 相同的 Bearer Token

## 注意事項

- GHCR Docker image 目前發佈 `linux/amd64`。Apple Silicon 或 ARM server 會透過 emulation 執行。
- VNC 用於觀看瀏覽器畫面和基本操作
- 中文輸入和剪貼簿在 Chrome 瀏覽器上 seamless 支援
- 瀏覽器引擎和 Profile 資料存在 Docker volumes 中，重建容器不會遺失

## Apple Silicon (M1/M2/M3)

docker-compose.yml 已設定 `platform: linux/amd64`，透過 Rosetta/QEMU 模擬執行。
原生 `linux/arm64` image 會在 KasmVNC、Camoufox、CloakBrowser runtime 都驗證完成後再開啟。
