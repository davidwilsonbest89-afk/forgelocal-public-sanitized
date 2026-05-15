#!/bin/bash

: "${VNC_PASSWORD:=browseforge}"

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
