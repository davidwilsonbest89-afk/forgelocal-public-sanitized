#!/usr/bin/env bash
# ForgeLocal — exécution T05 : ne journalise aucun code ni token.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DASHBOARD_DIR="${DASHBOARD_DIR:?DASHBOARD_DIR requis}"
BASE_DIR="${FORGELOCAL_E2E_BASE_DIR:?FORGELOCAL_E2E_BASE_DIR requis}"
CORE_PORT="${FORGELOCAL_E2E_PORT:-19280}"
DASHBOARD_PORT="${FORGELOCAL_DASHBOARD_PORT:-3001}"

case "$BASE_DIR" in
  /tmp/forgelocal-*) ;;
  *) printf 'test-bootstrap-ro refuse un répertoire hors /tmp/forgelocal-*\n' >&2; exit 2 ;;
esac

test -x "$BASE_DIR/forgelocal"
test -f "$DASHBOARD_DIR/package.json"
cd "$DASHBOARD_DIR"
export FORGELOCAL_BINARY="$BASE_DIR/forgelocal"
export FORGELOCAL_BASE_DIR="$BASE_DIR"
export FORGELOCAL_CORE_BASE_URL="http://127.0.0.1:${CORE_PORT}"
export FORGELOCAL_DASHBOARD_URL="http://127.0.0.1:${DASHBOARD_PORT}"
pnpm exec playwright test tests/bootstrap-ro.spec.ts --reporter=line
