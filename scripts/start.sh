#!/bin/bash
# BrowseForge Launcher
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

# API token is intentionally not read or printed by the launcher.

# Camoufox will be launched by Control Server via Playwright
# Just wait for server to exit
echo "BrowseForge is running. Press Ctrl+C to stop."
wait $SERVER_PID
