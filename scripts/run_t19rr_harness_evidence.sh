#!/usr/bin/env bash
set -euo pipefail

# T19-RR-HARNESS-EVIDENCE — qualification locale, sans runtime Camoufox,
# proxy réseau réel, SystemVault natif ni publication. Les logs excluent tout token.

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DASHBOARD_DIR="$REPO_DIR/forge-dashboard"
OUTPUT_DIR="${1:-$REPO_DIR/t19rr-harness-evidence-raw}"
CORE_BINARY="$OUTPUT_DIR/forge-core"
GO_BIN="${GO_BIN:-/usr/local/go/bin/go}"
RUNTIME_CACHE_DIR="${RUNTIME_CACHE_DIR:-}"
mkdir -p "$OUTPUT_DIR"

if [[ ! -x "$GO_BIN" ]]; then
  echo "GO_COMPILER_UNAVAILABLE path=$GO_BIN" >&2
  exit 127
fi

cleanup_processes() {
  for pid in "${CORE_PID:-}" "${DASHBOARD_PID:-}"; do
    if [[ -n "${pid:-}" ]] && kill -0 "$pid" 2>/dev/null; then
      kill "$pid" 2>/dev/null || true
      wait "$pid" 2>/dev/null || true
    fi
  done
  CORE_PID=""
  DASHBOARD_PID=""
}
trap cleanup_processes EXIT

wait_for_200() {
  local url="$1"
  local label="$2"
  local deadline=$((SECONDS + 30))
  while (( SECONDS < deadline )); do
    if [[ "$(curl -sS --connect-timeout 2 --max-time 5 -o /dev/null -w '%{http_code}' "$url" || true)" == "200" ]]; then
      return 0
    fi
    sleep 1
  done
  echo "${label}_UNAVAILABLE url=${url}" >&2
  return 1
}

run_logged_playwright() {
  local log_file="$1"
  shift
  local command_display="pnpm exec playwright test $*"
  set +e
  (
    set +e
    echo "timestamp_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
    echo "cwd=$DASHBOARD_DIR"
    echo "git_head=$(git -C "$REPO_DIR" rev-parse HEAD)"
    echo "core=$FORGELOCAL_CORE_BASE_URL dashboard=$FORGELOCAL_DASHBOARD_URL"
    echo "binary=$FORGELOCAL_BINARY base_dir=$FORGELOCAL_BASE_DIR"
    echo "command=$command_display"
    (cd "$DASHBOARD_DIR" && pnpm exec playwright test "$@")
    status=$?
    echo "exit_code=$status"
    exit "$status"
  ) >"$log_file" 2>&1
  status=$?
  set -e
  return "$status"
}

run_instance() {
  local name="$1"
  local core_port="$2"
  local dashboard_port="$3"
  shift 3
  local instance_dir="$OUTPUT_DIR/$name"
  local base_dir="$instance_dir/core-data"
  local token_file="$instance_dir/token.tmp"
  mkdir -p "$instance_dir"

  cleanup_processes
  rm -rf "$base_dir"
  umask 077
  export BROWSEFORGE_TOKEN
  BROWSEFORGE_TOKEN="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"
  printf '%s' "$BROWSEFORGE_TOKEN" >"$token_file"
  chmod 600 "$token_file"

  "$CORE_BINARY" --base-dir "$base_dir" init --force --json >"$instance_dir/core-init.log" 2>&1
  export FORGELOCAL_BINARY="$CORE_BINARY"
  export FORGELOCAL_BASE_DIR="$base_dir"
  export FORGELOCAL_CORE_BASE_URL="http://127.0.0.1:$core_port"
  export FORGELOCAL_DASHBOARD_URL="http://127.0.0.1:$dashboard_port"
  export FORGELOCAL_TOKEN_PATH="$token_file"

  if [[ -n "$RUNTIME_CACHE_DIR" && -d "$RUNTIME_CACHE_DIR/browsers" ]]; then
    rm -rf "$base_dir/browsers"
    ln -s "$RUNTIME_CACHE_DIR/browsers" "$base_dir/browsers"
  fi

  "$CORE_BINARY" --base-dir "$base_dir" serve --host 127.0.0.1 --port "$core_port" --no-open >"$instance_dir/core-server.log" 2>&1 &
  CORE_PID=$!
  (cd "$DASHBOARD_DIR" && exec env VITE_CORE_BASE_URL="$FORGELOCAL_CORE_BASE_URL" node ./node_modules/vite/bin/vite.js --host 127.0.0.1 --port "$dashboard_port") >"$instance_dir/dashboard-server.log" 2>&1 &
  DASHBOARD_PID=$!

  wait_for_200 "$FORGELOCAL_CORE_BASE_URL/api/health" "CORE"
  wait_for_200 "$FORGELOCAL_DASHBOARD_URL" "DASHBOARD"
  run_logged_playwright "$instance_dir/playwright.log" "$@"
  cleanup_processes
  if [[ -n "$RUNTIME_CACHE_DIR" && ! -e "$RUNTIME_CACHE_DIR/browsers" && -d "$base_dir/browsers" ]]; then
    mkdir -p "$RUNTIME_CACHE_DIR"
    mv "$base_dir/browsers" "$RUNTIME_CACHE_DIR/browsers"
  fi
  rm -f "$token_file"
  unset BROWSEFORGE_TOKEN
}

echo "timestamp_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)" >"$OUTPUT_DIR/qualification-metadata.log"
echo "repo=$REPO_DIR" >>"$OUTPUT_DIR/qualification-metadata.log"
echo "base_commit=$(git -C "$REPO_DIR" rev-parse HEAD)" >>"$OUTPUT_DIR/qualification-metadata.log"
echo "build_command=$GO_BIN build -o forge-core ./cmd/server" >>"$OUTPUT_DIR/qualification-metadata.log"
(cd "$REPO_DIR" && "$GO_BIN" build -o "$CORE_BINARY" ./cmd/server) >>"$OUTPUT_DIR/core-build.log" 2>&1

# Chaque invocation possède son Core, Dashboard, token et base-dir propres.
run_instance "run1-t15" 19290 3100 tests/automation-t15.spec.ts --workers=1
run_instance "run2-t14" 19291 3101 tests/runtime-t14.spec.ts --workers=1
run_instance "run3-full" 19280 3102 --workers=1

echo "qualification_exit_code=0" >>"$OUTPUT_DIR/qualification-metadata.log"
