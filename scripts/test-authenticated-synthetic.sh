#!/usr/bin/env bash
set -u
set -o pipefail

ROOT=$(mktemp -d "${TMPDIR:-/tmp}/forgelocal-auth.XXXXXX")
BIN="$ROOT/server"
BASE="$ROOT/valid"
EXPIRED="$ROOT/expired"
OUT="$ROOT/server.stdout"
ERR="$ROOT/server.stderr"
PORT="${FORGELOCAL_AUTH_TEST_PORT:-19380}"
EXPIRED_PORT="${FORGELOCAL_AUTH_EXPIRED_TEST_PORT:-19381}"
SENTINEL="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
SERVER_PID=""

cleanup() {
  if [ -n "$SERVER_PID" ] && kill -0 "$SERVER_PID" 2>/dev/null; then
    kill -TERM "$SERVER_PID" 2>/dev/null || true
    for _ in $(seq 1 30); do
      kill -0 "$SERVER_PID" 2>/dev/null || break
      sleep 0.1
    done
    kill -KILL "$SERVER_PID" 2>/dev/null || true
    wait "$SERVER_PID" 2>/dev/null || true
  fi
  SERVER_PID=""
  rm -rf "$ROOT"
}
trap cleanup EXIT

fail() { printf 'AUTH_SYNTHETIC_FAIL=%s\n' "$1" >&2; exit 1; }
contains_sentinel() { grep -Fq -- "$SENTINEL" "$1" 2>/dev/null; }
assert_absent() { ! contains_sentinel "$1" || fail "$2_SENTINEL_FOUND"; }

build() {
  go build -o "$BIN" ./cmd/server || fail BUILD
  chmod 700 "$BIN"
}

wait_ready() {
  local port="$1" code
  for _ in $(seq 1 40); do
    code=$(curl -sS --max-time 1 -o /dev/null -w '%{http_code}' "http://127.0.0.1:$port/api/status" 2>/dev/null)
    [ "$code" = 200 ] && return 0
    sleep 0.25
  done
  return 1
}

start_server() {
  local base="$1" port="$2" out="$3" err="$4"
  "$BIN" --base-dir "$base" init --force --json >/dev/null 2>&1 || return 1
  umask 077
  mkdir -p "$base/data"
  printf '%s\n' "$SENTINEL" > "$base/data/.api-token"
  chmod 0600 "$base/data/.api-token"
  "$BIN" --base-dir "$base" serve --host 127.0.0.1 --port "$port" --no-open --no-runtime > "$out" 2> "$err" &
  SERVER_PID=$!
  wait_ready "$port"
}

http_get() {
  local out="$1" url="$2" header="${3:-}"
  if [ -n "$header" ]; then
    curl -sS --max-time 3 -o "$out" -w '%{http_code}' -H "$header" "$url" 2>/dev/null
  else
    curl -sS --max-time 3 -o "$out" -w '%{http_code}' "$url" 2>/dev/null
  fi
}

http_post() {
  local out="$1" url="$2" header="$3" origin="$4"
  curl -sS --max-time 3 -X POST -o "$out" -w '%{http_code}' -H "$header" -H "Origin: $origin" "$url" 2>/dev/null
}

build
mkdir -p "$BASE" "$EXPIRED"
start_server "$BASE" "$PORT" "$OUT" "$ERR" || fail SERVER_START
mode=$(stat -c '%a' "$BASE/data/.api-token")
[ "$mode" = 600 ] || fail TOKEN_MODE
code=$(http_get "$ROOT/positive.json" "http://127.0.0.1:$PORT/api/runtimes" "Authorization: Bearer $SENTINEL")
[ "$code" = 200 ] || fail "AUTHENTICATED_REQUEST_HTTP_$code"
assert_absent "$ROOT/positive.json" AUTH_RESPONSE
code=$(http_get "$ROOT/missing.json" "http://127.0.0.1:$PORT/api/runtimes")
[ "$code" = 401 ] || fail "MISSING_HTTP_$code"
grep -Fq '"reason":"missing"' "$ROOT/missing.json" || fail MISSING_REASON
assert_absent "$ROOT/missing.json" MISSING_RESPONSE
code=$(http_get "$ROOT/invalid.json" "http://127.0.0.1:$PORT/api/runtimes" 'Authorization: Bearer invalid-synthetic-token')
[ "$code" = 401 ] || fail "INVALID_HTTP_$code"
grep -Fq '"reason":"invalid"' "$ROOT/invalid.json" || fail INVALID_REASON
assert_absent "$ROOT/invalid.json" INVALID_RESPONSE
code=$(http_post "$ROOT/revoke.json" "http://127.0.0.1:$PORT/api/auth/revoke" "Authorization: Bearer $SENTINEL" "http://127.0.0.1:$PORT")
[ "$code" = 204 ] || fail "REVOKE_HTTP_$code"
code=$(http_get "$ROOT/revoked.json" "http://127.0.0.1:$PORT/api/runtimes" "Authorization: Bearer $SENTINEL")
[ "$code" = 401 ] || fail "REVOKED_HTTP_$code"
grep -Fq '"reason":"revoked"' "$ROOT/revoked.json" || fail REVOKED_REASON
assert_absent "$ROOT/revoked.json" REVOKED_RESPONSE
kill -TERM "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""
rm -f "$BASE/data/.api-token" "$BASE/data/.api-token.meta"

start_server "$EXPIRED" "$EXPIRED_PORT" "$ROOT/expired-first.stdout" "$ROOT/expired-first.stderr" || fail EXPIRED_SETUP
kill -TERM "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""
META="$EXPIRED/data/.api-token.meta"
sed -i -E 's/"issued_at": "[^"]+"/"issued_at": "2000-01-01T00:00:00Z"/; s/"expires_at": "[^"]+"/"expires_at": "2000-01-01T01:00:00Z"/' "$META" || fail EXPIRED_METADATA
"$BIN" --base-dir "$EXPIRED" serve --host 127.0.0.1 --port "$EXPIRED_PORT" --no-open --no-runtime > "$ROOT/expired.stdout" 2> "$ROOT/expired.stderr" &
SERVER_PID=$!
wait_ready "$EXPIRED_PORT" || fail EXPIRED_SERVER_START
code=$(http_get "$ROOT/expired.json" "http://127.0.0.1:$EXPIRED_PORT/api/runtimes" "Authorization: Bearer $SENTINEL")
[ "$code" = 401 ] || fail "EXPIRED_HTTP_$code"
grep -Fq '"reason":"expired"' "$ROOT/expired.json" || fail EXPIRED_REASON
assert_absent "$ROOT/expired.json" EXPIRED_RESPONSE

if [ -r "/proc/$SERVER_PID/environ" ] && tr '\0' '\n' < "/proc/$SERVER_PID/environ" | grep -Fq -- "$SENTINEL"; then fail PROCESS_ENV_SENTINEL; fi
kill -TERM "$SERVER_PID" 2>/dev/null || true
wait "$SERVER_PID" 2>/dev/null || true
SERVER_PID=""
rm -f "$BASE/data/.api-token" "$BASE/data/.api-token.meta" "$EXPIRED/data/.api-token" "$EXPIRED/data/.api-token.meta"
assert_absent "$OUT" SERVER_STDOUT
assert_absent "$ERR" SERVER_STDERR
assert_absent "$ROOT/expired.stdout" EXPIRED_STDOUT
assert_absent "$ROOT/expired.stderr" EXPIRED_STDERR
if grep -RIl -F -- "$SENTINEL" "$ROOT" >/dev/null 2>&1; then fail ARTIFACT_SENTINEL; fi
printf '%s\n' 'AUTHENTICATED_SYNTHETIC_REQUEST=PASS'
printf '%s\n' 'MISSING_TOKEN=PASS'
printf '%s\n' 'INVALID_TOKEN=PASS'
printf '%s\n' 'EXPIRED_TOKEN=PASS'
printf '%s\n' 'REVOKED_TOKEN=PASS'
printf '%s\n' 'SENTINEL_ABSENT=PASS'
printf '%s\n' 'TOKEN_FILE_MODE_0600=PASS'
printf '%s\n' 'PROCESS_AND_TEMP_CLEANUP=PASS'
