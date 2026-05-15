# Linux Server 部署指南

[English](linux-server.md)

## 推薦：Docker 部署

```bash
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -e VNC_PASSWORD=browseforge \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.3
```

正式部署建議 pin 版本 tag，例如 `v1.7.3`。`latest` 可用於快速試用，但重啟或重新拉取時可能非預期升級。

> 目前 GHCR Docker image 發佈 `linux/amd64`。Apple Silicon 或 ARM server 會透過 emulation 執行；原生 `linux/arm64` image 會在 KasmVNC、Camoufox、CloakBrowser runtime 都驗證完成後再開啟。

| 服務 | URL |
|------|-----|
| Dashboard + API | http://YOUR_SERVER:19280 |
| MCP Streamable HTTP | http://YOUR_SERVER:19281 |
| 遠端桌面 (KasmVNC) | http://YOUR_SERVER:6901 |
| VNC 帳號 | `user` / 環境變數 `VNC_PASSWORD` |

### 取得 API Token

```bash
docker logs browseforge | grep "API Token"
# 或
docker exec browseforge /app/BrowseForge token
```

### 持久化資料

```bash
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -e VNC_PASSWORD=browseforge \
  -v browseforge-profiles:/app/profiles \
  -v browseforge-data:/app/data \
  -v browseforge-browsers:/app/browsers \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.3
```

### Docker Compose

```yaml
services:
  browseforge:
    image: ghcr.io/nczz/browseforge:v1.7.3
    platform: linux/amd64
    ports:
      - "19280:19280"
      - "19281:19281"
      - "6901:6901"
    volumes:
      - browseforge-profiles:/app/profiles
      - browseforge-data:/app/data
      - browseforge-browsers:/app/browsers
    environment:
      - VNC_PASSWORD=browseforge
    restart: unless-stopped

volumes:
  browseforge-profiles:
  browseforge-data:
  browseforge-browsers:
```

## 防火牆設定

| 端口 | 用途 | 是否必要 |
|------|------|---------|
| 19280 | Dashboard + REST API + Playwright WebSocket proxy | ✅ 必要 |
| 19281 | MCP Streamable HTTP | 選用（需要遠端 MCP 時） |
| 6901 | KasmVNC 遠端桌面 | 選用（需要看畫面時） |

```bash
sudo ufw allow 19280/tcp
sudo ufw allow 19281/tcp
sudo ufw allow 6901/tcp
```

## 安全建議

- **不要**把 19280、19281 和 6901 直接暴露到公網，用 SSH tunnel 或 VPN 存取
- KasmVNC 有 Basic Auth 保護（user/password）
- API/MCP Token 存在 `data/.api-token`，不要外洩
- 建議用 nginx reverse proxy + HTTPS 包裝

```bash
# SSH tunnel 方式（最安全）
ssh -L 19280:localhost:19280 -L 19281:localhost:19281 -L 6901:localhost:6901 user@server
# 然後本機開 http://localhost:19280、http://localhost:19281 和 http://localhost:6901
```

## 升級

```bash
docker pull ghcr.io/nczz/browseforge:v1.7.3
docker stop browseforge && docker rm browseforge
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -v browseforge-profiles:/app/profiles \
  -v browseforge-data:/app/data \
  -v browseforge-browsers:/app/browsers \
  -e VNC_PASSWORD=browseforge \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.3
```

Profiles、Token、瀏覽器引擎都在 volumes 中保留，只更新 BrowseForge。

## 特性

- **KasmVNC** — Chrome 上 seamless 剪貼簿、IME 中文輸入
- **WebGL 完整偽裝** — GLX + 軟體渲染 + 完整 WebGL 指紋
- **Docker 自動偵測** — 自動 `0.0.0.0` + `--no-sandbox`
- **Playwright WebSocket proxy** — 外部腳本透過 19280 port 連入操作瀏覽器
- **MCP Streamable HTTP** — 遠端 MCP client 透過 19281 port 連入，使用 Bearer Token
