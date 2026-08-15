#!/usr/bin/env bash
# ForgeLocal — collecteur T00 à T05 BOOTSTRAP-RO-01.
# Les preuves d'exécution restent locales et exclues de Git. Aucun code ou token
# n'est écrit dans les sorties : les assertions ne produisent que des statuts.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO="${FORGELOCAL_GO:-/home/ubuntu/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.13.linux-amd64/bin/go}"
TARGET_COMMIT="${FORGELOCAL_TARGET_COMMIT:-5123dc4}"
DASHBOARD_DIR="${DASHBOARD_DIR:?DASHBOARD_DIR requis}"
EVIDENCE_DIR="${BOOTSTRAP_RO_EVIDENCE_DIR:?BOOTSTRAP_RO_EVIDENCE_DIR requis}"
CORE_PORT="${FORGELOCAL_E2E_PORT:-19280}"
NONLOOP_PORT="${FORGELOCAL_NONLOOP_PORT:-19281}"
DASHBOARD_PORT="${FORGELOCAL_DASHBOARD_PORT:-3001}"
SOURCE_DIR="$EVIDENCE_DIR/source"
CORE_BASE_DIR="$EVIDENCE_DIR/core"
NONLOOP_BASE_DIR="$EVIDENCE_DIR/nonloop"

case "$EVIDENCE_DIR" in
  /tmp/forgelocal-*|/home/ubuntu/ForgeLocal/.local-evidence/*) ;;
  *) printf 'Le dossier de preuves doit rester hors livraison : /tmp/forgelocal-* ou .local-evidence/\n' >&2; exit 2 ;;
esac
test -f "$DASHBOARD_DIR/package.json"
test -x "$GO"
mkdir -p "$EVIDENCE_DIR"
for port in "$CORE_PORT" "$NONLOOP_PORT" "$DASHBOARD_PORT"; do
  if ss -ltn "sport = :$port" | grep -q LISTEN; then
    printf 'Le port de preuve %s est déjà occupé ; choisir un triplet libre.\n' "$port" >&2
    exit 2
  fi
done
if [[ -e "$SOURCE_DIR" ]]; then
  printf 'Le worktree de preuve existe déjà : %s\n' "$SOURCE_DIR" >&2
  exit 2
fi

core_pid=""
nonloop_pid=""
dashboard_pid=""
cleanup() {
  for pid in "$dashboard_pid" "$nonloop_pid" "$core_pid"; do
    [[ -n "$pid" ]] && kill "$pid" 2>/dev/null || true
  done
  [[ -d "$SOURCE_DIR" ]] && git -C "$ROOT" worktree remove --force "$SOURCE_DIR" 2>/dev/null || true
}
trap cleanup EXIT

git -C "$ROOT" worktree add --detach "$SOURCE_DIR" "$TARGET_COMMIT" >"$EVIDENCE_DIR/T00_worktree.log" 2>&1
{
  printf 'target_commit=%s\n' "$TARGET_COMMIT"
  printf 'source_commit=%s\n' "$(git -C "$SOURCE_DIR" rev-parse HEAD)"
  printf 'dashboard_commit=%s\n' "$(git -C "$DASHBOARD_DIR" rev-parse HEAD 2>/dev/null || printf non_git)"
  printf 'go_version=%s\n' "$(GOTOOLCHAIN=local "$GO" version)"
  printf 'node_version=%s\n' "$(node --version)"
  printf 'playwright_version=%s\n' "$(cd "$DASHBOARD_DIR" && pnpm exec playwright --version)"
} >"$EVIDENCE_DIR/T00_environment.log"

export GOTOOLCHAIN=local
"$GO" -C "$SOURCE_DIR" test -list 'TestReadOnly(Bootstrap|Session|Routes)' ./internal/api >"$EVIDENCE_DIR/T01_test_list.log"
"$GO" -C "$SOURCE_DIR" test ./internal/api -run 'TestReadOnly(Bootstrap|Session|Routes)' -count=1 -v -json >"$EVIDENCE_DIR/T01_api_tests.json"
"$GO" -C "$SOURCE_DIR" test -race ./internal/api -run 'TestReadOnly(Bootstrap|Session)' -count=1 -v -json >"$EVIDENCE_DIR/T02_api_race.json"
"$GO" -C "$SOURCE_DIR" test ./cmd/server -run 'Test(CLIReadOnlySession|DashboardOpenURL)' -count=1 -v -json >"$EVIDENCE_DIR/T03_cli_tests.json"
node "$SOURCE_DIR/scripts/check-ci-provenance-workflow.mjs" >"$EVIDENCE_DIR/T04_provenance.log"
node "$SOURCE_DIR/scripts/test-release-separation.mjs" >"$EVIDENCE_DIR/T04_release_separation.log"

FORGELOCAL_SOURCE_DIR="$SOURCE_DIR" FORGELOCAL_E2E_BASE_DIR="$CORE_BASE_DIR" FORGELOCAL_E2E_PORT="$CORE_PORT" \
  "$ROOT/scripts/core-dev.sh" >"$EVIDENCE_DIR/T05_core.log" 2>&1 &
core_pid="$!"
FORGELOCAL_SOURCE_DIR="$SOURCE_DIR" FORGELOCAL_E2E_BASE_DIR="$NONLOOP_BASE_DIR" FORGELOCAL_E2E_PORT="$NONLOOP_PORT" FORGELOCAL_E2E_HOST="0.0.0.0" \
  "$ROOT/scripts/core-dev.sh" >"$EVIDENCE_DIR/T05_nonloop_core.log" 2>&1 &
nonloop_pid="$!"
for _ in $(seq 1 60); do
  if curl --fail --silent --noproxy '*' "http://127.0.0.1:${CORE_PORT}/api/status" >/dev/null \
    && curl --fail --silent --noproxy '*' "http://127.0.0.1:${NONLOOP_PORT}/api/status" >/dev/null; then break; fi
  sleep 1
done
curl --fail --silent --noproxy '*' "http://127.0.0.1:${CORE_PORT}/api/status" >/dev/null
curl --fail --silent --noproxy '*' "http://127.0.0.1:${NONLOOP_PORT}/api/status" >/dev/null

FORGELOCAL_E2E_BASE_DIR="$CORE_BASE_DIR" FORGELOCAL_E2E_PORT="$CORE_PORT" \
  "$ROOT/scripts/dashboard-dev.sh" >"$EVIDENCE_DIR/T05_dashboard.log" 2>&1 &
dashboard_pid="$!"
for _ in $(seq 1 60); do
  if curl --fail --silent --noproxy '*' "http://127.0.0.1:${DASHBOARD_PORT}" >/dev/null; then break; fi
  sleep 1
done
curl --fail --silent --noproxy '*' "http://127.0.0.1:${DASHBOARD_PORT}" >/dev/null

ss -ltnp | grep -E ":(${CORE_PORT}|${NONLOOP_PORT}|${DASHBOARD_PORT})\\b" | \
  sed -E 's/pid=[0-9]+/pid=<redacted>/g' >"$EVIDENCE_DIR/T05_loopback_sockets.log"
grep -Eq "127\.0\.0\.1:${CORE_PORT}" "$EVIDENCE_DIR/T05_loopback_sockets.log"
grep -Eq "127\.0\.0\.1:${DASHBOARD_PORT}" "$EVIDENCE_DIR/T05_loopback_sockets.log"
printf 'T05_SOCKET_SCOPE: principal=loopback dashboard=loopback control_nonloop=separate\n' >>"$EVIDENCE_DIR/T05_loopback_sockets.log"

ip_addr="$(ip -4 -o addr show scope global | awk 'NR==1 {split($4, part, "/"); print part[1]}')"
[[ -n "$ip_addr" ]] || { printf 'Adresse IPv4 non loopback indisponible\n' >&2; exit 1; }
nonloop_issuance="$($NONLOOP_BASE_DIR/forgelocal --base-dir "$NONLOOP_BASE_DIR" readonly-session code --base-url "http://127.0.0.1:${NONLOOP_PORT}" --json)"
nonloop_code="$(printf '%s' "$nonloop_issuance" | sed -n 's/.*"code"[[:space:]]*:[[:space:]]*"\([a-f0-9]*\)".*/\1/p')"
[[ "$nonloop_code" =~ ^[a-f0-9]{64}$ ]] || { printf 'Émission nonloop invalide\n' >&2; exit 1; }
nonloop_result="$(curl --silent --noproxy '*' --request POST "http://${ip_addr}:${NONLOOP_PORT}/api/v1/readonly/session/bootstrap" --header 'Content-Type: application/json' --data "{\"code\":\"${nonloop_code}\"}" --write-out $'\n%{http_code}')"
nonloop_status="${nonloop_result##*$'\n'}"
nonloop_body="${nonloop_result%$'\n'*}"
[[ "$nonloop_status" == 403 ]] && printf '%s' "$nonloop_body" | grep -q 'LOOPBACK_REQUIRED'
printf 'T05_NONLOOPBACK: PASS status=403 code=LOOPBACK_REQUIRED\n' >"$EVIDENCE_DIR/T05_nonloopback.log"

FORGELOCAL_BINARY="$CORE_BASE_DIR/forgelocal" FORGELOCAL_BASE_DIR="$CORE_BASE_DIR" FORGELOCAL_BASE_URL="http://127.0.0.1:${CORE_PORT}" \
  FORGELOCAL_NONLOOPBACK_BASE_URL="" BOOTSTRAP_RO_EVIDENCE_FILE="$EVIDENCE_DIR/T05_core_contract.log" VERIFY_EXPIRY=1 \
  "$SOURCE_DIR/scripts/validate-bootstrap-ro-01.sh" >"$EVIDENCE_DIR/T05_core_contract_runner.log" 2>&1

DASHBOARD_DIR="$DASHBOARD_DIR" FORGELOCAL_E2E_BASE_DIR="$CORE_BASE_DIR" FORGELOCAL_E2E_PORT="$CORE_PORT" \
  FORGELOCAL_DASHBOARD_PORT="$DASHBOARD_PORT" "$ROOT/scripts/test-bootstrap-ro-playwright.sh" >"$EVIDENCE_DIR/T05_playwright.log" 2>&1

if grep -Eiq 'authorization:|bearer[[:space:]]+[a-f0-9]{16,}|"token"[[:space:]]*:' "$EVIDENCE_DIR"/T05_*.log; then
  printf 'Les logs T05 contiennent un motif d’autorisation interdit\n' >&2
  exit 1
fi
printf 'T05_LOG_REDACTION: PASS authorization=absent bearer=absent token_value=absent\n' >"$EVIDENCE_DIR/T05_log_redaction.log"

git -C "$SOURCE_DIR" diff --name-only "$TARGET_COMMIT" -- release/back01-minimal dist/back01-minimal >"$EVIDENCE_DIR/T06_rc_paths.log"
test ! -s "$EVIDENCE_DIR/T06_rc_paths.log"
printf 'T06_RC_PATHS: PASS changed_files=0\n' >>"$EVIDENCE_DIR/T06_rc_paths.log"
find "$EVIDENCE_DIR" -maxdepth 1 -type f ! -name 'SHA256SUMS' -printf '%f\n' | sort | while read -r name; do sha256sum "$EVIDENCE_DIR/$name"; done >"$EVIDENCE_DIR/SHA256SUMS"
printf 'BOOTSTRAP_RO_EXECUTION_EVIDENCE: PASS target_commit=%s\n' "$TARGET_COMMIT"
