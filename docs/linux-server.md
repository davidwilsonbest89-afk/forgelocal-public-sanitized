# Linux Server Deployment

[繁體中文](linux-server.zh-TW.md)

## Recommended: Docker

```bash
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -e VNC_PASSWORD=browseforge \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.0
```

Pin a version tag such as `v1.7.0` for production deployments. Use `latest` only for short trials, because pulling or restarting later may upgrade unexpectedly.

> The current GHCR Docker image is `linux/amd64`. Apple Silicon and ARM servers run it through emulation. Native `linux/arm64` Docker images should only be enabled after KasmVNC, Camoufox, and CloakBrowser runtime checks pass inside an ARM container.

| Service | URL |
|------|-----|
| Dashboard + REST API + Playwright proxy | `http://YOUR_SERVER:19280` |
| MCP Streamable HTTP | `http://YOUR_SERVER:19281` |
| KasmVNC remote desktop | `http://YOUR_SERVER:6901` |
| VNC credentials | `user` / `VNC_PASSWORD` |

## Get the API Token

```bash
docker logs browseforge | grep "API Token"
docker exec browseforge /app/BrowseForge token
```

## Persistent Data

```bash
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -e VNC_PASSWORD=browseforge \
  -v browseforge-profiles:/app/profiles \
  -v browseforge-data:/app/data \
  -v browseforge-browsers:/app/browsers \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.0
```

## Docker Compose

```yaml
services:
  browseforge:
    image: ghcr.io/nczz/browseforge:v1.7.0
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

## Firewall

| Port | Purpose | Required |
|------|---------|----------|
| 19280 | Dashboard + REST API + Playwright WebSocket proxy | Required |
| 19281 | MCP Streamable HTTP | Optional, only for remote MCP |
| 6901 | KasmVNC remote desktop | Optional, only when visual access is needed |

```bash
sudo ufw allow 19280/tcp
sudo ufw allow 19281/tcp
sudo ufw allow 6901/tcp
```

## Security

- Do not expose `19280`, `19281`, or `6901` directly to the public internet.
- Use SSH tunnels, VPN, or a hardened HTTPS reverse proxy.
- KasmVNC uses Basic Auth through `user` and `VNC_PASSWORD`.
- API and MCP share the Bearer token stored in `data/.api-token`.
- Treat profiles, backup ZIPs, exported profiles, cookies, and tokens as sensitive.

Recommended SSH tunnel:

```bash
ssh -L 19280:localhost:19280 -L 19281:localhost:19281 -L 6901:localhost:6901 user@server
```

Then open:

- `http://localhost:19280`
- `http://localhost:19281`
- `http://localhost:6901`

## Upgrade

```bash
docker pull ghcr.io/nczz/browseforge:v1.7.0
docker stop browseforge && docker rm browseforge
docker run -d --name browseforge \
  -p 19280:19280 -p 19281:19281 -p 6901:6901 \
  -v browseforge-profiles:/app/profiles \
  -v browseforge-data:/app/data \
  -v browseforge-browsers:/app/browsers \
  -e VNC_PASSWORD=browseforge \
  --restart unless-stopped \
  ghcr.io/nczz/browseforge:v1.7.0
```

Profiles, tokens, and browser engines remain in Docker volumes.

## Runtime Features

- KasmVNC remote desktop with browser viewing/control.
- GLX and software rendering for WebGL behavior in containerized environments.
- Docker auto-detection for `0.0.0.0` binding and `--no-sandbox`.
- Playwright WebSocket proxy through port `19280`.
- MCP Streamable HTTP through port `19281` with Bearer token authentication.
