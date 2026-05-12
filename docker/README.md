# BrowseForge Docker

一鍵部署 BrowseForge + noVNC 遠端桌面。

## 使用方式

```bash
cd docker
docker compose up -d --build
```

首次啟動需要 3-5 分鐘下載瀏覽器引擎（~440MB）。

## 連線

| 服務 | URL |
|------|-----|
| Dashboard | http://localhost:19280 |
| 遠端桌面 (noVNC) | http://localhost:6080/vnc.html |
| VNC 密碼 | `browseforge`（可透過環境變數 `VNC_PASSWORD` 修改） |

## 取得 API Token

```bash
docker compose logs | grep "API Token"
# 或
docker compose exec browseforge /app/BrowseForge token
```

## 注意事項

- VNC 用於觀看瀏覽器畫面，操作建議透過 Dashboard 或 API
- 中文輸入和剪貼簿操作請使用 BrowseForge API（`/api/sessions/{id}/type`）
- BrowseForge 在 Docker 中自動啟用 `0.0.0.0` 綁定和 `--no-sandbox`
- 瀏覽器引擎和 Profile 資料存在 Docker volumes 中，重建容器不會遺失

## 自訂

```yaml
environment:
  - VNC_PASSWORD=your_password  # VNC 連線密碼
```

## Apple Silicon (M1/M2/M3)

docker-compose.yml 已設定 `platform: linux/amd64`，會透過 Rosetta/QEMU 模擬執行。
