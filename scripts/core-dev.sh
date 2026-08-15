#!/usr/bin/env bash
# ForgeLocal — Core T05 : boucle locale, état temporaire, aucun runtime.
set -euo pipefail

ROOT="${FORGELOCAL_SOURCE_DIR:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
GO="${FORGELOCAL_GO:-/home/ubuntu/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.13.linux-amd64/bin/go}"
BASE_DIR="${FORGELOCAL_E2E_BASE_DIR:?FORGELOCAL_E2E_BASE_DIR requis}"
PORT="${FORGELOCAL_E2E_PORT:-19280}"
HOST="${FORGELOCAL_E2E_HOST:-127.0.0.1}"

case "$BASE_DIR" in
  /tmp/forgelocal-*) ;;
  *) printf 'core-dev refuse un répertoire hors /tmp/forgelocal-*\n' >&2; exit 2 ;;
esac

export GOTOOLCHAIN=local
test -f "$ROOT/go.mod"
mkdir -p "$BASE_DIR"
"$GO" -C "$ROOT" build -o "$BASE_DIR/forgelocal" ./cmd/server
"$BASE_DIR/forgelocal" --base-dir "$BASE_DIR" init --json >/dev/null
exec "$BASE_DIR/forgelocal" --base-dir "$BASE_DIR" serve --host "$HOST" --port "$PORT" --no-runtime --no-open
