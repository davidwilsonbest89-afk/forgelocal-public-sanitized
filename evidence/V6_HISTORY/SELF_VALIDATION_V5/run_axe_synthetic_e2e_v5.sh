#!/usr/bin/env bash
set +e
set -u
set -o pipefail
umask 077
BASE=/home/ubuntu/forgelocal_self_validation_v5
REPO=$BASE/repo
RUN_ROOT=$(mktemp -d /tmp/forgelocal-self-e2e.XXXXXX)
CORE_BASE=$(mktemp -d /tmp/forgelocal-self-core.XXXXXX)
TOKEN_FILE=$RUN_ROOT/token.tmp
AXE_RESULTS=$RUN_ROOT/axe-results.json
CORE_LOG=$BASE/SELF_VALIDATION_CORE_AXE_SYNTHETIC.log
DASH_LOG=$BASE/SELF_VALIDATION_DASHBOARD_AXE_SYNTHETIC.log
E2E_LOG=$BASE/SELF_VALIDATION_AXE_E2E_RAW.log
CLEANUP_LOG=$BASE/SELF_VALIDATION_AXE_CLEANUP_VERIFICATION.log
BIN=$RUN_ROOT/forgelocal-core
CORE_PID=''
DASH_PID=''
CSS_FILE="$REPO/forge-dashboard/client/src/index.css"
CSS_BACKUP="$RUN_ROOT/index.css.backup"
CORE_PORT=''
DASH_PORT=''
cleanup() {
  {
    echo '--- cleanup'; echo "completed_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)";
    if [ -n "$DASH_PID" ]; then kill "$DASH_PID" 2>/dev/null || true; pkill -TERM -P "$DASH_PID" 2>/dev/null || true; wait "$DASH_PID" 2>/dev/null || true; fi
    [ -n "$DASH_PORT" ] && pkill -f "vite --host 127.0.0.1 --port $DASH_PORT" 2>/dev/null || true
    [ -f "$CSS_BACKUP" ] && cp "$CSS_BACKUP" "$CSS_FILE"
    if [ -n "$CORE_PID" ]; then kill "$CORE_PID" 2>/dev/null || true; wait "$CORE_PID" 2>/dev/null || true; fi
    [ -n "$TOKEN_FILE" ] && rm -f "$TOKEN_FILE"
    if [ -f "$AXE_RESULTS" ]; then cp "$AXE_RESULTS" "$BASE/SELF_VALIDATION_AXE_RESULTS.json"; echo 'axe_results_preserved=yes'; else echo 'axe_results_preserved=no'; fi
    rm -rf "$CORE_BASE" "$RUN_ROOT"
    echo "token_file_exists=$(test -e "$TOKEN_FILE" && echo yes || echo no)"
    echo "core_base_exists=$(test -e "$CORE_BASE" && echo yes || echo no)"
    echo "run_root_exists=$(test -e "$RUN_ROOT" && echo yes || echo no)"
    if [ -n "$CORE_PORT" ]; then echo "core_port_listening=$(ss -ltn 2>/dev/null | grep -E ":${CORE_PORT}([[:space:]]|$)" >/dev/null && echo yes || echo no)"; fi
    if [ -n "$DASH_PORT" ]; then echo "dashboard_port_listening=$(ss -ltn 2>/dev/null | grep -E ":${DASH_PORT}([[:space:]]|$)" >/dev/null && echo yes || echo no)"; fi
    echo 'remaining_temp_processes:'
    ps -eo pid,args | grep -E 'forgelocal-self-(e2e|core)' | grep -v grep || true
  } > "$CLEANUP_LOG"
}
trap cleanup EXIT INT TERM
pick_port() {
  start=$1
  end=$2
  for p in $(seq "$start" "$end"); do
    if ! ss -ltn 2>/dev/null | awk '{print $4}' | grep -Eq "(:|\\])${p}$"; then printf '%s' "$p"; return 0; fi
  done
  return 1
}
CORE_PORT=$(pick_port 19280 19380)
DASH_PORT=$(pick_port 4173 4273)
if [ -z "$CORE_PORT" ] || [ -z "$DASH_PORT" ]; then echo 'PORT_ALLOCATION_FAILED' > "$E2E_LOG"; exit 1; fi
cd "$REPO" || exit 1
{
  echo "started_utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "head=$(git rev-parse HEAD)"
  echo "core_port=$CORE_PORT dashboard_port=$DASH_PORT"
  echo 'policy=loopback-only no-runtime synthetic data only'
  echo 'token_policy=temporary restrictive file never logged never archived'
  echo 'axe_results_path=temporary_run_root'
  echo '--- build command'
  echo 'GOTOOLCHAIN=local CGO_ENABLED=1 go build -o TEMP_BINARY ./cmd/server'
} > "$E2E_LOG"
TOKEN=$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')
printf '%s' "$TOKEN" > "$TOKEN_FILE"
chmod 600 "$TOKEN_FILE"
export BROWSEFORGE_TOKEN=$(cat "$TOKEN_FILE")
export GOTOOLCHAIN=local
export PATH=/usr/local/go1.25.13/bin:/home/ubuntu/bin:$PATH
CGO_ENABLED=1 /usr/local/go1.25.13/bin/go build -o "$BIN" ./cmd/server >> "$E2E_LOG" 2>&1
build_rc=$?
echo "build_exit_code=$build_rc" >> "$E2E_LOG"
if [ "$build_rc" -ne 0 ]; then exit "$build_rc"; fi
"$BIN" --base-dir "$CORE_BASE" --config "$CORE_BASE/config.json" serve --host 127.0.0.1 --port "$CORE_PORT" --no-sandbox --no-open --no-runtime > "$CORE_LOG" 2>&1 &
CORE_PID=$!
for i in $(seq 1 60); do
  if curl -fsS "http://127.0.0.1:${CORE_PORT}/api/status" >/dev/null 2>&1; then break; fi
  sleep 0.25
done
if ! curl -fsS "http://127.0.0.1:${CORE_PORT}/api/status" >/dev/null 2>&1; then echo 'CORE_READY=FAIL' >> "$E2E_LOG"; exit 1; fi
echo 'CORE_READY=PASS' >> "$E2E_LOG"
cd "$REPO/forge-dashboard" || exit 1
cp "$CSS_FILE" "$CSS_BACKUP"
sed -i '/fonts.googleapis.com/d' "$CSS_FILE"
echo 'external_fonts_import=temporarily_removed_for_loopback_policy' >> "$E2E_LOG"
printf '%s\n' 'temporary_axe_module_prepared_outside_product_manifests' > "$BASE/SELF_VALIDATION_DASHBOARD_INSTALL.log"
install_rc=$?
echo "dashboard_install_exit_code=$install_rc" >> "$E2E_LOG"
if [ "$install_rc" -ne 0 ]; then exit "$install_rc"; fi
VITE_CORE_BASE_URL="http://127.0.0.1:${CORE_PORT}" pnpm exec vite --host 127.0.0.1 --port "$DASH_PORT" --strictPort > "$DASH_LOG" 2>&1 &
DASH_PID=$!
for i in $(seq 1 80); do
  if curl -fsS "http://127.0.0.1:${DASH_PORT}/" >/dev/null 2>&1; then break; fi
  sleep 0.25
done
if ! curl -fsS "http://127.0.0.1:${DASH_PORT}/" >/dev/null 2>&1; then echo 'DASHBOARD_READY=FAIL' >> "$E2E_LOG"; exit 1; fi
echo 'DASHBOARD_READY=PASS' >> "$E2E_LOG"
FORGELOCAL_CORE_BASE_URL="http://127.0.0.1:${CORE_PORT}" \
FORGELOCAL_DASHBOARD_URL="http://127.0.0.1:${DASH_PORT}" \
FORGELOCAL_BINARY="$BIN" \
FORGELOCAL_BASE_DIR="$CORE_BASE" \
FORGELOCAL_TEST_TOKEN_FILE="$TOKEN_FILE" \
FORGELOCAL_AXE_RESULTS="$AXE_RESULTS" \
pnpm exec playwright test tests/self-validation-axe-e2e.spec.ts --workers=1 >> "$E2E_LOG" 2>&1
E2E_RC=$?
echo "playwright_exit_code=$E2E_RC" >> "$E2E_LOG"
exit "$E2E_RC"
