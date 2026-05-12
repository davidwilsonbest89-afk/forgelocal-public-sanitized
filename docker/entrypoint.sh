#!/bin/bash

# Clean up stale files from previous runs
rm -f /tmp/.X99-lock /tmp/.X11-unix/X99
find /app/profiles -name "SingletonLock" -delete 2>/dev/null
find /app/profiles -name "SingletonCookie" -delete 2>/dev/null
find /app/profiles -name "SingletonSocket" -delete 2>/dev/null

# Start virtual framebuffer
Xvfb :99 -screen 0 2560x1440x24 &
for i in $(seq 1 30); do
  [ -e /tmp/.X11-unix/X99 ] && break
  sleep 0.5
done

# Start window manager + clipboard sync
fluxbox &
autocutsel -fork &
autocutsel -selection PRIMARY -fork &

# Start VNC server
x11vnc -display :99 -passwd "$VNC_PASSWORD" -forever -shared -bg 2>/dev/null
sleep 1

# Start noVNC web interface
websockify --web /usr/share/novnc 6080 localhost:5900 &
sleep 1

# Start BrowseForge
/app/BrowseForge &
BF_PID=$!

# Wait for ready and show token
for i in $(seq 1 60); do
  if [ -f /app/data/.api-token ]; then
    TOKEN=$(cat /app/data/.api-token)
    echo "========================================="
    echo "  BrowseForge Docker"
    echo "  Dashboard:  http://0.0.0.0:19280"
    echo "  Remote VNC: http://0.0.0.0:6080/vnc.html"
    echo "  VNC Password: $VNC_PASSWORD"
    echo "  API Token: $TOKEN"
    echo "========================================="
    break
  fi
  sleep 2
done

wait $BF_PID
