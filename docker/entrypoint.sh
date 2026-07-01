#!/bin/bash

: "${VNC_PASSWORD:=browseforge}"
: "${BROWSEFORGE_SEED_BROWSERS:=1}"

mkdir -p /app/profiles /app/data /app/browsers /app/logs /app/backups

# Seed or update host-mounted browser cache from the image so browser engines
# follow the BrowseForge image version contract instead of drifting silently.
if [ "$BROWSEFORGE_SEED_BROWSERS" = "1" ] && [ -d /opt/browseforge/browsers ]; then
  for engine in camoufox cloakbrowser; do
    image_version=""
    current_version=""
    [ -f "/opt/browseforge/browsers/${engine}/.version" ] && image_version="$(cat "/opt/browseforge/browsers/${engine}/.version")"
    [ -f "/app/browsers/${engine}/.version" ] && current_version="$(cat "/app/browsers/${engine}/.version")"
    if [ -n "$image_version" ] && [ "$current_version" != "$image_version" ]; then
      echo "Seeding ${engine} into /app/browsers..."
      rm -rf "/app/browsers/${engine}"
      mkdir -p "/app/browsers/${engine}"
      cp -a "/opt/browseforge/browsers/${engine}/." "/app/browsers/${engine}/"
    fi
  done
elif [ "$BROWSEFORGE_SEED_BROWSERS" != "1" ]; then
  echo "Browser engine seeding disabled by BROWSEFORGE_SEED_BROWSERS=${BROWSEFORGE_SEED_BROWSERS}"
fi

if ! find /app/browsers -name ".version" -print -quit 2>/dev/null | grep -q .; then
  echo "Browser engines are not installed yet. First startup may spend several minutes downloading them before the dashboard becomes available."
fi

# Clean up stale files
rm -f /tmp/.X1-lock /tmp/.X11-unix/X1
find /app/profiles -name "SingletonLock" -delete 2>/dev/null
find /app/profiles -name "SingletonCookie" -delete 2>/dev/null
find /app/profiles -name "SingletonSocket" -delete 2>/dev/null

# Setup KasmVNC user with password
mkdir -p ~/.vnc
echo -e "${VNC_PASSWORD}\n${VNC_PASSWORD}\n" | vncpasswd -u user -wr

# Start KasmVNC X server (Basic Auth protects HTTP, VNC protocol uses None)
# +extension GLX +render enables WebGL support for Camoufox
/usr/bin/Xvnc :1 \
  -geometry 2560x1440 \
  -depth 24 \
  -sslOnly 0 \
  -SecurityTypes None \
  -AlwaysShared \
  -websocketPort 6901 \
  -interface 0.0.0.0 \
  +extension GLX \
  +render \
  -http-header Cross-Origin-Embedder-Policy=require-corp \
  -http-header Cross-Origin-Opener-Policy=same-origin \
  -httpd /usr/share/kasmvnc/www &

sleep 2

# Start window manager
export DISPLAY=:1
openbox &

# Start BrowseForge (force software GL so WebGL works in Camoufox)
export LIBGL_ALWAYS_SOFTWARE=1
/app/BrowseForge &
BF_PID=$!

# Wait for ready and show token
for i in $(seq 1 60); do
  if [ -f /app/data/.api-token ]; then
    TOKEN=$(cat /app/data/.api-token)
    echo "========================================="
    echo "  BrowseForge Docker (KasmVNC)"
    echo "  Dashboard:  http://0.0.0.0:19280"
    echo "  Remote VNC: http://0.0.0.0:6901"
    echo "  VNC User:   user"
    echo "  VNC Password: set via VNC_PASSWORD (default: browseforge)"
    echo "  API Token: $TOKEN"
    echo "========================================="
    break
  fi
  sleep 2
done

wait $BF_PID
