#!/usr/bin/env bash
# ForgeLocal — Dashboard T05 : origine HTTP loopback, Core explicitement local.
set -euo pipefail

DASHBOARD_DIR="${DASHBOARD_DIR:?DASHBOARD_DIR requis}"
CORE_PORT="${FORGELOCAL_E2E_PORT:-19280}"
DASHBOARD_PORT="${FORGELOCAL_DASHBOARD_PORT:-3001}"

test -f "$DASHBOARD_DIR/package.json"
cd "$DASHBOARD_DIR"
export VITE_CORE_BASE_URL="http://127.0.0.1:${CORE_PORT}"
exec ./node_modules/.bin/vite --host=127.0.0.1 --port="$DASHBOARD_PORT" --strictPort
