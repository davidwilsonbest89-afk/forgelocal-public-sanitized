#!/bin/bash
# CamoufoxMulti Launcher
set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$DIR"

# macOS Gatekeeper fix
if [[ "$(uname)" == "Darwin" ]]; then
  xattr -cr . 2>/dev/null || true
fi

# Start Control Server (background)
echo "Starting Control Server..."
./control-server --config config.json &
SERVER_PID=$!

# Wait for server ready
for i in $(seq 1 10); do
  if curl -s http://127.0.0.1:19280/api/status > /dev/null 2>&1; then
    echo "Control Server ready (PID: $SERVER_PID)"
    break
  fi
  sleep 1
done

# Read API token
TOKEN=$(cat data/.api-token 2>/dev/null || echo "")
echo "API Token: $TOKEN"

# Camoufox will be launched by Control Server via Playwright
# Just wait for server to exit
echo "CamoufoxMulti is running. Press Ctrl+C to stop."
wait $SERVER_PID
