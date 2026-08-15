#!/usr/bin/env bash
# ForgeLocal — collecteur T06 Groupes/Runtimes en lecture seule.
# Chaque preuve est produite dans un worktree propre au commit ciblé. Les
# rapports ne contiennent ni code de bootstrap, ni Bearer, ni sentinelle brute.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO="${FORGELOCAL_GO:-/home/ubuntu/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.25.13.linux-amd64/bin/go}"
GITLEAKS="${FORGELOCAL_GITLEAKS:-/tmp/forgelocal-gitleaks-8.18.4-verified/gitleaks}"
TARGET_COMMIT="${FORGELOCAL_TARGET_COMMIT:-$(git -C "$ROOT" rev-parse HEAD)}"
DASHBOARD_DIR="${DASHBOARD_DIR:?DASHBOARD_DIR requis}"
DASHBOARD_PORT="${FORGELOCAL_T06_DASHBOARD_PORT:-3012}"
EVIDENCE_DIR="${T06_EVIDENCE_DIR:?T06_EVIDENCE_DIR requis}"

case "$EVIDENCE_DIR" in
  /tmp/forgelocal-*|"$ROOT"/.local-evidence/*) ;;
  *) printf 'Le dossier de preuves T06 doit rester sous /tmp/forgelocal-* ou .local-evidence/.\n' >&2; exit 2 ;;
esac
test -x "$GO"
test -x "$GITLEAKS"
test -f "$DASHBOARD_DIR/package.json"

SOURCE_DIR="$EVIDENCE_DIR/source"
PREARCHIVE_DIR="$EVIDENCE_DIR/prearchive-extracted"
FINAL_ARCHIVE="${EVIDENCE_DIR%/}.zip"
FINAL_EXTRACTED="${EVIDENCE_DIR}-extracted"
dashboard_pid=""

stop_dashboard() {
  if [[ -n "$dashboard_pid" ]]; then
    kill "$dashboard_pid" 2>/dev/null || true
    wait "$dashboard_pid" 2>/dev/null || true
    dashboard_pid=""
  fi
}
cleanup() {
  stop_dashboard
  [[ -d "$SOURCE_DIR" ]] && git -C "$ROOT" worktree remove --force "$SOURCE_DIR" 2>/dev/null || true
}
trap cleanup EXIT

if [[ -e "$EVIDENCE_DIR" || -e "$FINAL_ARCHIVE" || -e "$FINAL_EXTRACTED" ]]; then
  printf 'Le chemin de preuve T06 existe déjà : %s\n' "$EVIDENCE_DIR" >&2
  exit 2
fi
for port in "$DASHBOARD_PORT"; do
  if ss -ltn "sport = :$port" | grep -q LISTEN; then
    printf 'Le port de preuve T06 %s est déjà occupé.\n' "$port" >&2
    exit 2
  fi
done

mkdir -p "$EVIDENCE_DIR"
git -C "$ROOT" worktree add --detach "$SOURCE_DIR" "$TARGET_COMMIT" >/dev/null
parent_commit="$(git -C "$SOURCE_DIR" rev-parse "${TARGET_COMMIT}^")"

git -C "$SOURCE_DIR" status --short >"$EVIDENCE_DIR/T00_git_status_short.log"
git -C "$SOURCE_DIR" diff --check >"$EVIDENCE_DIR/T00_git_diff_check.log"
git -C "$SOURCE_DIR" diff --name-only "$parent_commit" "$TARGET_COMMIT" -- release/back01-minimal dist/back01-minimal >"$EVIDENCE_DIR/T00_rc_paths.log"
test ! -s "$EVIDENCE_DIR/T00_git_status_short.log"
test ! -s "$EVIDENCE_DIR/T00_git_diff_check.log"
test ! -s "$EVIDENCE_DIR/T00_rc_paths.log"
{
  printf 'target_commit=%s\n' "$TARGET_COMMIT"
  printf 'target_parent=%s\n' "$parent_commit"
  printf 'dashboard_commit=%s\n' "$(git -C "$DASHBOARD_DIR" rev-parse HEAD)"
  printf 'go_version=%s\n' "$(GOTOOLCHAIN=local "$GO" version)"
  printf 'node_version=%s\n' "$(node --version)"
  printf 'playwright_version=%s\n' "$(cd "$DASHBOARD_DIR" && pnpm exec playwright --version)"
  printf 'gitleaks_version=%s\n' "$("$GITLEAKS" version)"
  printf 'git_status_short=PASS bytes=%s\n' "$(wc -c < "$EVIDENCE_DIR/T00_git_status_short.log")"
  printf 'git_diff_check=PASS bytes=%s\n' "$(wc -c < "$EVIDENCE_DIR/T00_git_diff_check.log")"
  printf 'rc_paths_against_target_parent=PASS changed_files=%s\n' "$(wc -l < "$EVIDENCE_DIR/T00_rc_paths.log")"
} >"$EVIDENCE_DIR/T00_environment.log"

export GOTOOLCHAIN=local
"$GO" -C "$SOURCE_DIR" test ./internal/api -list 'Test.*ReadOnly.*Catalog|Test.*Group.*Redact|Test.*Runtime.*Redact' >"$EVIDENCE_DIR/T06_1_api_test_list.log"
"$GO" -C "$SOURCE_DIR" test ./internal/api -run 'Test(ReadOnlyCatalog|SQLiteReadOnly)' -count=1 -v -json >"$EVIDENCE_DIR/T06_1_api_tests.json"
"$GO" -C "$SOURCE_DIR" test ./internal/backup -run 'TestSQLiteReadOnlyCatalog' -count=1 -v -json >"$EVIDENCE_DIR/T06_2_sqlite_tests.json"
"$GO" -C "$SOURCE_DIR" test -race ./internal/api ./internal/backup -run 'Test(ReadOnlyCatalog|SQLiteReadOnly)' -count=1 >"$EVIDENCE_DIR/T06_2_race.log"

(
  cd "$DASHBOARD_DIR"
  pnpm check
  pnpm build
) >"$EVIDENCE_DIR/T06_3_dashboard_build.log" 2>&1
(
  cd "$DASHBOARD_DIR"
  pnpm exec vite --host 127.0.0.1 --port "$DASHBOARD_PORT"
) >"$EVIDENCE_DIR/T06_3_dashboard_server.log" 2>&1 &
dashboard_pid="$!"
for _ in $(seq 1 60); do
  curl --fail --silent --noproxy '*' "http://127.0.0.1:${DASHBOARD_PORT}" >/dev/null && break
  sleep 1
done
curl --fail --silent --noproxy '*' "http://127.0.0.1:${DASHBOARD_PORT}" >/dev/null
(
  cd "$DASHBOARD_DIR"
  FORGELOCAL_DASHBOARD_URL="http://127.0.0.1:${DASHBOARD_PORT}" pnpm exec playwright test tests/t06-groups-runtimes.spec.ts --reporter=line
) >"$EVIDENCE_DIR/T06_3_playwright.log" 2>&1
stop_dashboard

"$GITLEAKS" detect --source "$SOURCE_DIR" --log-opts="${parent_commit}..${TARGET_COMMIT}" --redact --report-format json --report-path "$EVIDENCE_DIR/T06_4_gitleaks_delta_redacted.json" >"$EVIDENCE_DIR/T06_4_gitleaks_delta.log" 2>&1
if grep -R -Eiq 't06-sentinel|/t06/private|T06-RUNTIME-HASH|authorization:|bearer[[:space:]]+[a-f0-9]{16,}' "$EVIDENCE_DIR"; then
  printf 'Une sentinelle ou un motif d’autorisation interdit est présent dans les preuves T06.\n' >&2
  exit 1
fi
printf 'T06_REDACTION: PASS sentinels=absent authorization=absent bearer=absent\n' >"$EVIDENCE_DIR/T06_4_redaction.log"

make_manifest() {
  (
    cd "$EVIDENCE_DIR"
    find . -maxdepth 1 -type f ! -name 'SHA256SUMS' -printf '%f\n' | LC_ALL=C sort | while IFS= read -r name; do
      sha256sum "$name"
    done
  ) >"$EVIDENCE_DIR/SHA256SUMS"
}
make_archive() {
  rm -f "$FINAL_ARCHIVE"
  (
    cd "$EVIDENCE_DIR"
    find . -maxdepth 1 -type f -printf '%f\n' | LC_ALL=C sort | zip -q "$FINAL_ARCHIVE" -@
  )
}

make_manifest
make_archive
mkdir -p "$PREARCHIVE_DIR"
unzip -q "$FINAL_ARCHIVE" -d "$PREARCHIVE_DIR"
"$GITLEAKS" detect --no-git --source "$PREARCHIVE_DIR" --redact --report-format json --report-path "$EVIDENCE_DIR/T06_5_gitleaks_archive_redacted.json" >"$EVIDENCE_DIR/T06_5_gitleaks_archive.log" 2>&1
rm -rf "$PREARCHIVE_DIR"
make_manifest
make_archive

mkdir -p "$FINAL_EXTRACTED"
unzip -t "$FINAL_ARCHIVE" >"$EVIDENCE_DIR/T06_5_archive_integrity.log"
unzip -q "$FINAL_ARCHIVE" -d "$FINAL_EXTRACTED"
(
  cd "$FINAL_EXTRACTED"
  sha256sum -c SHA256SUMS
) >"$EVIDENCE_DIR/T06_5_manifest_check.log"
"$GITLEAKS" detect --no-git --source "$FINAL_EXTRACTED" --redact --report-format json --report-path "$EVIDENCE_DIR/T06_5_gitleaks_final_extracted_redacted.json" >"$EVIDENCE_DIR/T06_5_gitleaks_final_extracted.log" 2>&1
archive_sha="$(sha256sum "$FINAL_ARCHIVE" | awk '{print $1}')"
{
  printf 'T06_EVIDENCE_ARCHIVE=PASS\n'
  printf 'archive=%s\n' "$FINAL_ARCHIVE"
  printf 'sha256=%s\n' "$archive_sha"
  printf 'manifest=PASS\n'
  printf 'archive_extract_gitleaks=PASS\n'
  printf 'public_release_status=PUBLIC_RELEASE_BLOCKED\n'
  printf 'scan_status=SCAN_BLOCKED_UNKNOWN\n'
} >"$EVIDENCE_DIR/T06_5_summary.log"
printf 'T06 evidence archive created: %s\n' "$FINAL_ARCHIVE"
