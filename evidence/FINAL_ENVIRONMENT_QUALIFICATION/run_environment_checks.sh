#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-env-qualification
REPO="$WORK/repo"
FIX="$WORK/fixtures"
LOG="$WORK/FINAL_ENVIRONMENT_TESTS_RAW.log"
mkdir -p "$FIX"
printf '%s\n' '<!doctype html><meta charset="utf-8"><title>ForgeLocal synthetic loopback</title><body>synthetic-loopback-ok</body>' > "$FIX/index.html"
run() {
  name="$1"; shift
  {
    echo "=== $name ==="
    echo "COMMAND=$*"
    date -u --iso-8601=seconds
    "$@"
    rc=$?
    echo "EXIT_CODE=$rc"
    echo
  } >> "$LOG" 2>&1
}
{
  echo '=== BROWSER/PLATFORM ENVIRONMENT CHECKS ==='
  date -u --iso-8601=seconds
  echo 'All targets are synthetic or loopback only; no credentials, cookies, external providers, Camoufox or real proxy are used.'
} >> "$LOG"
python3 -m http.server 18765 --bind 127.0.0.1 --directory "$FIX" > "$WORK/loopback-http-server.log" 2>&1 &
server_pid=$!
cleanup() {
  kill "$server_pid" 2>/dev/null || true
  wait "$server_pid" 2>/dev/null || true
  rm -rf "$WORK/chromium-profile-a" "$WORK/chromium-profile-b"
}
trap cleanup EXIT
sleep 1
CHROMIUM=$(command -v chromium || command -v chromium-browser || true)
if [ -n "$CHROMIUM" ]; then
  run CHROMIUM_LOOPBACK_PROFILE_A timeout 30s "$CHROMIUM" --headless=new --no-sandbox --disable-gpu --user-data-dir="$WORK/chromium-profile-a" --dump-dom http://127.0.0.1:18765/index.html
  run CHROMIUM_LOOPBACK_PROFILE_B timeout 30s "$CHROMIUM" --headless=new --no-sandbox --disable-gpu --user-data-dir="$WORK/chromium-profile-b" --dump-dom http://127.0.0.1:18765/index.html
  run CHROMIUM_PROXY_FAIL_CLOSED timeout 20s "$CHROMIUM" --headless=new --no-sandbox --disable-gpu --user-data-dir="$WORK/chromium-profile-a" --proxy-server=http://127.0.0.1:1 --proxy-bypass-list= --dump-dom http://example.invalid/
else
  echo '=== CHROMIUM ===' >> "$LOG"
  echo 'STATUS=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE' >> "$LOG"
  echo 'REASON=Chromium binary absent' >> "$LOG"
fi
if command -v docker >/dev/null 2>&1; then
  run DOCKER_VERSION docker version
  run DOCKER_INFO docker info
  run DOCKER_BUILDX docker buildx version
  run DOCKER_CONTEXTS docker context ls
else
  {
    echo '=== DOCKER ==='
    echo 'STATUS=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'
    echo 'REASON=Docker client and daemon absent in environment'
  } >> "$LOG"
fi
if command -v firefox >/dev/null 2>&1; then
  run FIREFOX_VERSION firefox --version
  run FIREFOX_HEADLESS timeout 30s firefox --headless --version
else
  {
    echo '=== FIREFOX ==='
    echo 'STATUS=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'
    echo 'REASON=Firefox standard binary absent; Camoufox intentionally not downloaded'
  } >> "$LOG"
fi
{
  echo '=== SYSTEMVAULT_NATIVE ==='
  if command -v secret-tool >/dev/null 2>&1 && command -v busctl >/dev/null 2>&1; then
    echo 'STATUS=BLOCKED_ENVIRONMENT_REQUIRED'
    echo 'REASON=Presence of clients alone is insufficient; no verified active native Secret Service/keyring test was authorized in this run.'
    secret-tool --version 2>&1 || true
    busctl --user status org.freedesktop.secrets 2>&1 || true
  else
    echo 'STATUS=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'
    echo 'REASON=Native Secret Service/keyring tooling unavailable'
  fi
  echo 'VERDICT=NATIVE_SYSTEMVAULT_NOT_TESTED'
} >> "$LOG"
{
  echo '=== RESIDUAL PROCESS CHECK ==='
  pgrep -af 'chromium.*forgelocal-final-environment-qualification' || true
  pgrep -af 'http.server 18765' || true
  echo 'CLEANUP_EXPECTED=PASS'
} >> "$LOG"
