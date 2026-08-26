#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-env-qualification
REPO="$WORK/repo"
LOG="$WORK/FINAL_ENVIRONMENT_TESTS_RAW.log"
export PATH="$WORK/tools/go/bin:$PATH"
export GOTOOLCHAIN=local
cd "$REPO"
: > "$LOG"
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
    return "$rc"
  } >> "$LOG" 2>&1
}
run GO_TEST_RACE go test -count=1 -race ./cmd/... ./internal/... || true
run GO_VET go vet ./cmd/... ./internal/... || true
run GO_BUILD go build ./cmd/... ./internal/... || true
run GIT_DIFF_CHECK git diff --check || true
{
  echo '=== ENVIRONMENT AFTER GO INSTALL ==='
  date -u --iso-8601=seconds
  go version
  go env GOVERSION GOOS GOARCH GOPATH GOMODCACHE
  git status --short --branch
} >> "$LOG" 2>&1
cat "$LOG"
