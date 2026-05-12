# Linux Server 部署指南

BrowseForge 在無桌面環境的 Linux Server 上有三種運行模式。

## 模式一：純 API（無畫面）

適合全自動化場景，不需要看到瀏覽器畫面。

```bash
sudo apt install -y xvfb
xvfb-run ./BrowseForge
```

透過 REST API 或 MCP 操作瀏覽器，所有操作在虛擬螢幕中執行。

## 模式二：noVNC 遠端桌面（推薦）

適合需要人工介入的場景（驗證碼、手動登入）。用瀏覽器即可遠端看到和操作瀏覽器畫面。

### 安裝

```bash
sudo apt install -y xvfb x11vnc novnc websockify
```

### 啟動

```bash
# 1. 啟動虛擬螢幕
export DISPLAY=:99
Xvfb :99 -screen 0 1920x1080x24 &

# 2. 啟動 VNC Server（連接到虛擬螢幕）
x11vnc -display :99 -nopw -forever -shared &

# 3. 啟動 noVNC（Web 介面，端口 6080）
websockify --web /usr/share/novnc 6080 localhost:5900 &

# 4. 啟動 BrowseForge
./BrowseForge
```

### 連線

用任何瀏覽器開啟：
```
http://YOUR_SERVER_IP:6080/vnc.html
```

點「Connect」即可看到遠端桌面。BrowseForge 開啟的瀏覽器視窗都會顯示在這裡。

### 設定密碼（建議）

```bash
x11vnc -display :99 -passwd YOUR_PASSWORD -forever -shared &
```

連線時輸入密碼。

### 一鍵啟動腳本

```bash
#!/bin/bash
# start-server.sh — 啟動 BrowseForge + noVNC 遠端桌面
export DISPLAY=:99

# 清理舊進程
pkill -f Xvfb; pkill -f x11vnc; pkill -f websockify; pkill -f BrowseForge
sleep 1

# 啟動虛擬螢幕（1920x1080）
Xvfb :99 -screen 0 1920x1080x24 &
sleep 1

# 啟動 VNC
x11vnc -display :99 -passwd ${VNC_PASSWORD:-browseforge} -forever -shared -bg

# 啟動 noVNC Web 介面
websockify --web /usr/share/novnc 6080 localhost:5900 &

# 啟動 BrowseForge
echo "========================================="
echo "  BrowseForge Server"
echo "  Dashboard:  http://0.0.0.0:19280"
echo "  Remote VNC: http://0.0.0.0:6080/vnc.html"
echo "  VNC 密碼:   ${VNC_PASSWORD:-browseforge}"
echo "========================================="
./BrowseForge
```

使用：
```bash
chmod +x start-server.sh
VNC_PASSWORD=mypassword ./start-server.sh
```

## 模式三：Docker Compose（最簡單）

BrowseForge v1.5+ 內建 Docker 自動偵測，會自動啟用 `0.0.0.0` 綁定和 `--no-sandbox`。

```yaml
# docker-compose.yml
services:
  browseforge:
    image: ubuntu:24.04
    platform: linux/amd64
    ports:
      - "19280:19280"   # Dashboard + API
      - "19281:19281"   # MCP Server
      - "6080:6080"     # noVNC 遠端桌面
    volumes:
      - ./profiles:/app/profiles
      - ./data:/app/data
      - ./browsers:/app/browsers
    environment:
      - DISPLAY=:99
      - VNC_PASSWORD=browseforge
    command: >
      bash -c "
        apt-get update && apt-get install -y xvfb x11vnc novnc websockify fluxbox curl unzip &&
        Xvfb :99 -screen 0 1920x1080x24 & sleep 2 &&
        fluxbox & 
        x11vnc -display :99 -passwd $$VNC_PASSWORD -forever -shared -bg &&
        websockify --web /usr/share/novnc 6080 localhost:5900 &
        cd /app && ./BrowseForge
      "
```

```bash
docker compose up -d
# Dashboard: http://localhost:19280
# 遠端桌面: http://localhost:6080/vnc.html
# 取得 Token: docker compose exec browseforge /app/BrowseForge token
```

## 防火牆設定

| 端口 | 用途 | 是否必要 |
|------|------|---------|
| 19280 | Dashboard + REST API | ✅ 必要 |
| 19281 | MCP Server | 選用（AI Agent 整合時） |
| 6080 | noVNC 遠端桌面 | 選用（需要看畫面時） |
| 5900 | VNC 原生端口 | ❌ 不需對外開放（noVNC 內部使用） |

```bash
# UFW 範例
sudo ufw allow 19280/tcp
sudo ufw allow 6080/tcp
```

## 安全建議

- **不要**把 19280 直接暴露到公網，用 SSH tunnel 或 VPN 存取
- noVNC 務必設定密碼（`VNC_PASSWORD`）
- API Token 存在 `data/.api-token`，不要外洩
- 建議用 nginx reverse proxy + HTTPS 包裝

```bash
# SSH tunnel 方式（最安全）
ssh -L 19280:localhost:19280 -L 6080:localhost:6080 user@server
# 然後本機開 http://localhost:19280 和 http://localhost:6080/vnc.html
```
