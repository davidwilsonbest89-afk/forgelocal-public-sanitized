#!/usr/bin/env bash
set -euo pipefail

BASE_COMMIT="${1:?baseline requise}"
TARGET_COMMIT="${2:-HEAD}"
ATTESTATION_FILE="${T07R_ATTESTATION_FILE:?chemin de l’attestation redacted privée requis}"
INBOX_ROOT="${T07R_INBOX_ROOT:-/home/ubuntu/forgelocal-private-evidence/t07-r-inbox}"
GITLEAKS="${GITLEAKS_BIN:-/tmp/forgelocal-gitleaks-8.18.4-verified/gitleaks}"
ROOT="$(git rev-parse --show-toplevel)"
BASE="$(git -C "$ROOT" rev-parse "$BASE_COMMIT")"
TARGET="$(git -C "$ROOT" rev-parse "$TARGET_COMMIT")"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="${T07R_EVIDENCE_DIR:-/tmp/forgelocal-t07-r-evidence-${STAMP}}"
ARCHIVE="${OUT}.zip"
SNAPSHOT="${OUT}/delta-snapshot"
WORKTREE="${OUT}/worktree"
INBOX_REAL="$(readlink -f "$INBOX_ROOT")"
ATTESTATION_REAL="$(readlink -f "$ATTESTATION_FILE")"

if [[ ! -x "$GITLEAKS" || ! -f "$ATTESTATION_REAL" ]]; then
  echo "pinned gitleaks binary or private attestation unavailable" >&2
  exit 2
fi
if [[ "$ATTESTATION_REAL" != "$INBOX_REAL"/* ]]; then
  echo "attestation must remain inside the controlled private inbox" >&2
  exit 2
fi
if ! git -C "$ROOT" merge-base --is-ancestor "$BASE" "$TARGET"; then
  echo "baseline is not an ancestor of target" >&2
  exit 2
fi

rm -rf "$OUT" "$ARCHIVE"
mkdir -p "$OUT" "$SNAPSHOT"
trap 'git -C "$ROOT" worktree remove --force "$WORKTREE" >/dev/null 2>&1 || true' EXIT
git -C "$ROOT" worktree add --detach "$WORKTREE" "$TARGET" >/dev/null

# The validator emits only status, field paths and generic error codes. The private input is never copied.
if ! node "$WORKTREE/scripts/validate-t07-r-attestation.mjs" "$ATTESTATION_REAL" > "$OUT/01_attestation_completeness_redacted.json"; then
  echo "attestation is incomplete or inconsistent; no evidence archive created" >&2
  exit 1
fi

set +e
"$GITLEAKS" detect --no-git --source "$ATTESTATION_REAL" --redact --report-format json --report-path "$OUT/02_attestation_gitleaks_raw_redacted.json" > "$OUT/02_attestation_gitleaks_redacted.out" 2>&1
attestation_scan_status=$?
set -e
sed "s|$INBOX_REAL/||g" "$OUT/02_attestation_gitleaks_raw_redacted.json" > "$OUT/02_attestation_gitleaks_redacted.json"
rm -f "$OUT/02_attestation_gitleaks_raw_redacted.json"
attestation_findings="$( (grep -o '"RuleID"' "$OUT/02_attestation_gitleaks_redacted.json" || true) | wc -l | tr -d ' ' )"
printf 'gitleaks_exit=%s\ngitleaks_findings=%s\n' "$attestation_scan_status" "$attestation_findings" >> "$OUT/02_attestation_gitleaks_redacted.out"
if [[ "$attestation_scan_status" -ne 0 || "$attestation_findings" -ne 0 ]]; then
  echo "private attestation has a security finding; no evidence archive created" >&2
  exit 1
fi

printf 'baseline=%s\ntarget=%s\n' "$BASE" "$TARGET" > "$OUT/00_scope.txt"
printf '%s\n' \
  'purpose=receipt-completeness-only' \
  'private_attestation_copied=false' \
  'candidate_source_copied=false' \
  't08_authorized=false' > "$OUT/00_constraints.txt"
git -C "$ROOT" diff --name-only "$BASE..$TARGET" | LC_ALL=C sort > "$OUT/03_delta_files.txt"
if [[ ! -s "$OUT/03_delta_files.txt" ]]; then
  echo "empty delta" >&2
  exit 2
fi
while IFS= read -r path; do
  mkdir -p "$SNAPSHOT/$(dirname "$path")"
  git -C "$ROOT" show "$TARGET:$path" > "$SNAPSHOT/$path"
done < "$OUT/03_delta_files.txt"
printf 'delta_file_count=' > "$OUT/03_coverage.txt"
wc -l < "$OUT/03_delta_files.txt" >> "$OUT/03_coverage.txt"
git -C "$ROOT" diff --name-only "$BASE..$TARGET" -- release/back01-minimal dist/back01-minimal > "$OUT/04_rc_delta.txt"
if [[ -s "$OUT/04_rc_delta.txt" ]]; then
  echo "frozen RC delta detected" >&2
  exit 1
fi
git -C "$WORKTREE" status --short > "$OUT/05_git_status_short.txt"
git -C "$WORKTREE" diff --check > "$OUT/06_git_diff_check.txt"
(
  cd "$WORKTREE"
  npm run check:component-rights
  npm run test:t07-provenance
) > "$OUT/07_provenance_controls.out" 2>&1
set +e
"$GITLEAKS" detect --no-git --source "$SNAPSHOT" --redact --report-format json --report-path "$OUT/08_delta_gitleaks_redacted.json" > "$OUT/08_delta_gitleaks_redacted.out" 2>&1
delta_scan_status=$?
set -e
delta_findings="$( (grep -o '"RuleID"' "$OUT/08_delta_gitleaks_redacted.json" || true) | wc -l | tr -d ' ' )"
printf 'gitleaks_exit=%s\ngitleaks_findings=%s\n' "$delta_scan_status" "$delta_findings" >> "$OUT/08_delta_gitleaks_redacted.out"
if [[ "$delta_scan_status" -ne 0 || "$delta_findings" -ne 0 ]]; then
  echo "unexpected delta finding" >&2
  exit 1
fi

rm -rf "$SNAPSHOT" "$WORKTREE"
trap - EXIT
PARENT="$(dirname "$OUT")"
OUT_NAME="$(basename "$OUT")"
(
  cd "$OUT"
  find . -type f ! -name SHA256SUMS -printf '%P\n' | LC_ALL=C sort | xargs -r sha256sum > SHA256SUMS
)
(
  cd "$PARENT"
  zip -q -r "$ARCHIVE" "$OUT_NAME"
)
VERIFY="${OUT}-verify"
rm -rf "$VERIFY"
unzip -q "$ARCHIVE" -d "$VERIFY"
(
  cd "$VERIFY/$OUT_NAME"
  sha256sum -c SHA256SUMS > /dev/null
)
set +e
"$GITLEAKS" detect --no-git --source "$VERIFY" --redact --report-format json --report-path "$OUT/09_archive_rescan_redacted.json" > "$OUT/09_archive_rescan_redacted.out" 2>&1
archive_scan_status=$?
set -e
archive_findings="$( (grep -o '"RuleID"' "$OUT/09_archive_rescan_redacted.json" || true) | wc -l | tr -d ' ' )"
printf 'gitleaks_exit=%s\ngitleaks_findings=%s\n' "$archive_scan_status" "$archive_findings" >> "$OUT/09_archive_rescan_redacted.out"
if [[ "$archive_scan_status" -ne 0 || "$archive_findings" -ne 0 ]]; then
  echo "unexpected archive finding" >&2
  exit 1
fi
(
  cd "$OUT"
  find . -type f ! -name SHA256SUMS -printf '%P\n' | LC_ALL=C sort | xargs -r sha256sum > SHA256SUMS
)
rm -f "$ARCHIVE"
(
  cd "$PARENT"
  zip -q -r "$ARCHIVE" "$OUT_NAME"
)
rm -rf "$VERIFY"
printf 'archive_sha256='; sha256sum "$ARCHIVE" | awk '{print $1}'
printf 'archive_path=%s\n' "$ARCHIVE"
