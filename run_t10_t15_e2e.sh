#!/usr/bin/env bash
set -u
umask 077
LOT=/home/ubuntu/forgelocal_self_validation_v6/t10_t15_e2e
REPO=/home/ubuntu/forgelocal_self_validation_v6/repo
DASH=$REPO/forge-dashboard
EVID=$LOT/evidence/T10-T15-E2E-VALIDATION
mkdir -p "$EVID"
RUN_ROOT=$(mktemp -d /tmp/forgelocal-t10-t15-run.XXXXXX)
BASE_DIR="$RUN_ROOT/base"
TOKEN_FILE="$RUN_ROOT/token"
CORE_BIN="$RUN_ROOT/forge-core"
CORE_LOG="$RUN_ROOT/core.log"
DASH_LOG="$RUN_ROOT/dashboard.log"
TEST_LOG="$RUN_ROOT/playwright.log"
TOKEN=e2e_t10_t15_$(od -An -N12 -tx1 /dev/urandom | tr -d ' \n')
printf '%s\n' "$TOKEN" > "$TOKEN_FILE"
chmod 600 "$TOKEN_FILE"
mkdir -p "$BASE_DIR"
cp "$REPO/config.default.json" "$BASE_DIR/config.json"
sed -i 's#browsers/browseforge-chromium/chrome#/usr/bin/chromium#' "$BASE_DIR/config.json"

cleanup() {
  local rc=$?
  set +e
  if [ -n "${DASH_PID:-}" ]; then kill -- -"$DASH_PID" 2>/dev/null || kill "$DASH_PID" 2>/dev/null; fi
  if [ -n "${CORE_PID:-}" ]; then kill -- -"$CORE_PID" 2>/dev/null || kill "$CORE_PID" 2>/dev/null; fi
  for _ in $(seq 1 20); do
    alive=0
    if [ -n "${DASH_PID:-}" ] && kill -0 "$DASH_PID" 2>/dev/null; then alive=1; fi
    if [ -n "${CORE_PID:-}" ] && kill -0 "$CORE_PID" 2>/dev/null; then alive=1; fi
    [ "$alive" -eq 0 ] && break
    sleep 0.25
  done
  if [ -n "${DASH_PID:-}" ] && kill -0 "$DASH_PID" 2>/dev/null; then kill -- -"$DASH_PID" 2>/dev/null || kill -KILL "$DASH_PID" 2>/dev/null; fi
  if [ -n "${CORE_PID:-}" ] && kill -0 "$CORE_PID" 2>/dev/null; then kill -- -"$CORE_PID" 2>/dev/null || kill -KILL "$CORE_PID" 2>/dev/null; fi
  # Redact the deterministic temporary token if any process emitted it unexpectedly.
  for f in "$CORE_LOG" "$DASH_LOG" "$TEST_LOG"; do
    [ -f "$f" ] && sed -i "s/$TOKEN/[REDACTED_TEMP_TOKEN]/g" "$f"
  done
  cp "$CORE_LOG" "$EVID/CORE_LOOPBACK_REDACTED.log" 2>/dev/null || true
  cp "$DASH_LOG" "$EVID/DASHBOARD_LOOPBACK_REDACTED.log" 2>/dev/null || true
  cp "$TEST_LOG" "$EVID/PLAYWRIGHT_T10_T15_SEQUENTIAL.log" 2>/dev/null || true
  rm -f "$TOKEN_FILE"
  rm -rf "$BASE_DIR"
  # Keep only evidence; remove the temporary process root after copying logs.
  rm -rf "$RUN_ROOT"
  {
    echo "test_exit=$TEST_EXIT"
    echo "core_pid=${CORE_PID:-unset} dashboard_pid=${DASH_PID:-unset}"
    echo "token_file_removed=$([ ! -e "$TOKEN_FILE" ] && echo yes || echo no)"
    echo "base_dir_removed=$([ ! -e "$BASE_DIR" ] && echo yes || echo no)"
    echo "port_19280_after_cleanup=$([ -z "$(ss -ltn '( sport = :19280 )' 2>/dev/null | tail -n +2)" ] && echo closed || echo open)"
    echo "port_3000_after_cleanup=$([ -z "$(ss -ltn '( sport = :3000 )' 2>/dev/null | tail -n +2)" ] && echo closed || echo open)"
    echo "run_root_removed=$([ ! -e "$RUN_ROOT" ] && echo yes || echo no)"
  } > "$EVID/CLEANUP_RESULT.log"
  exit "$rc"
}
TEST_EXIT=125
trap cleanup EXIT

# Do not start if either requested loopback port is already occupied.
if [ -n "$(ss -ltn '( sport = :19280 or sport = :3000 )' 2>/dev/null | tail -n +2)" ]; then
  echo 'PORT_PRECONDITION_FAILED' > "$EVID/EXECUTION_RESULT.log"
  TEST_EXIT=125
  exit 125
fi

(cd "$REPO" && /usr/local/go1.25.13/bin/go build -o "$CORE_BIN" ./cmd/server) >"$EVID/CORE_BUILD.log" 2>&1
setsid env BROWSEFORGE_TOKEN="$TOKEN" BROWSEFORGE_SKIP_BROWSEFORGE_CHROMIUM_AUTO_UPDATE=true "$CORE_BIN" --base-dir "$BASE_DIR" --config "$BASE_DIR/config.json" serve --host 127.0.0.1 --port 19280 --no-open --no-sandbox >"$CORE_LOG" 2>&1 &
CORE_PID=$!

for _ in $(seq 1 60); do
  if curl -fsS --max-time 1 http://127.0.0.1:19280/api/health >/dev/null 2>&1; then break; fi
  sleep 1
done
if ! curl -fsS --max-time 2 http://127.0.0.1:19280/api/health >/dev/null 2>&1; then
  echo 'CORE_LOOPBACK_HEALTH_FAILED' > "$EVID/EXECUTION_RESULT.log"
  TEST_EXIT=126
  exit 126
fi

cd "$DASH"
setsid pnpm exec vite --host 127.0.0.1 --port 3000 >"$DASH_LOG" 2>&1 &
DASH_PID=$!
for _ in $(seq 1 60); do
  if curl -fsS --max-time 1 http://127.0.0.1:3000/ >/dev/null 2>&1; then break; fi
  sleep 1
done
if ! curl -fsS --max-time 2 http://127.0.0.1:3000/ >/dev/null 2>&1; then
  echo 'DASHBOARD_LOOPBACK_HEALTH_FAILED' > "$EVID/EXECUTION_RESULT.log"
  TEST_EXIT=127
  exit 127
fi

FORGELOCAL_CORE_BASE_URL=http://127.0.0.1:19280 \
FORGELOCAL_DASHBOARD_URL=http://127.0.0.1:3000 \
FORGELOCAL_BINARY="$CORE_BIN" \
FORGELOCAL_BASE_DIR="$BASE_DIR" \
FORGELOCAL_TOKEN_PATH="$TOKEN_FILE" \
BROWSEFORGE_TOKEN="$TOKEN" \
pnpm exec playwright test tests/proxies-t10.spec.ts tests/automation-t15.spec.ts --config=playwright.config.ts --workers=1 --reporter=line >"$TEST_LOG" 2>&1
TEST_EXIT=$?
{
  echo "test_exit=$TEST_EXIT"
  echo "core_health=200"
  echo "dashboard_health=200"
  echo "playwright_workers=1"
  echo "camoufox_started=no"
  echo "real_proxy_started=no"
  echo "real_cookie_or_user_data=no"
  echo "production_runtime=no"
} > "$EVID/EXECUTION_RESULT.log"
exit "$TEST_EXIT"
