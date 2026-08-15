#!/usr/bin/env bash
set -euo pipefail

BASE_COMMIT="${1:?baseline T07 requise}"
TARGET_COMMIT="${2:-HEAD}"
CANDIDATE_ARCHIVE="${CANDIDATE_ARCHIVE:-/home/ubuntu/upload/camoflox-FINAL.zip}"
GITLEAKS="${GITLEAKS_BIN:-/tmp/forgelocal-gitleaks-8.18.4-verified/gitleaks}"
ROOT="$(git rev-parse --show-toplevel)"
TARGET="$(git -C "$ROOT" rev-parse "$TARGET_COMMIT")"
BASE="$(git -C "$ROOT" rev-parse "$BASE_COMMIT")"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="${T07_EVIDENCE_DIR:-/tmp/forgelocal-t07-evidence-${STAMP}}"
WORKTREE="${OUT}/worktree"
SNAPSHOT="${OUT}/delta-snapshot"
CANDIDATE_SOURCE="${OUT}/candidate-passive-source"
ARCHIVE="${OUT}.zip"

if [[ ! -f "$CANDIDATE_ARCHIVE" || ! -x "$GITLEAKS" ]]; then
  echo "candidate archive or pinned gitleaks binary unavailable" >&2
  exit 2
fi
if ! git -C "$ROOT" merge-base --is-ancestor "$BASE" "$TARGET"; then
  echo "baseline is not an ancestor of target" >&2
  exit 2
fi

rm -rf "$OUT" "$ARCHIVE"
mkdir -p "$OUT" "$SNAPSHOT" "$CANDIDATE_SOURCE"
trap 'git -C "$ROOT" worktree remove --force "$WORKTREE" >/dev/null 2>&1 || true' EXIT
git -C "$ROOT" worktree add --detach "$WORKTREE" "$TARGET" >/dev/null

printf 'baseline=%s\ntarget=%s\n' "$BASE" "$TARGET" > "$OUT/00_scope.txt"
printf 'gitleaks_version=' >> "$OUT/00_scope.txt"
"$GITLEAKS" version >> "$OUT/00_scope.txt"
printf 'candidate_archive_sha256=' >> "$OUT/00_scope.txt"
sha256sum "$CANDIDATE_ARCHIVE" | awk '{print $1}' >> "$OUT/00_scope.txt"
printf '%s\n' 'purpose=provenance-only; no candidate code installed, executed, imported, or integrated' >> "$OUT/00_scope.txt"

git -C "$ROOT" diff --name-only "$BASE..$TARGET" | LC_ALL=C sort > "$OUT/01_delta_files.txt"
if [[ ! -s "$OUT/01_delta_files.txt" ]]; then
  echo "empty T07 delta" >&2
  exit 2
fi
while IFS= read -r path; do
  mkdir -p "$SNAPSHOT/$(dirname "$path")"
  git -C "$ROOT" show "$TARGET:$path" > "$SNAPSHOT/$path"
done < "$OUT/01_delta_files.txt"
printf 'delta_file_count=' > "$OUT/01_coverage.txt"
wc -l < "$OUT/01_delta_files.txt" >> "$OUT/01_coverage.txt"
printf 'snapshot_file_count=' >> "$OUT/01_coverage.txt"
find "$SNAPSHOT" -type f | wc -l >> "$OUT/01_coverage.txt"

git -C "$ROOT" diff --name-only "$BASE..$TARGET" -- release/back01-minimal dist/back01-minimal > "$OUT/02_rc_delta.txt"
printf 'rc_delta_file_count=' > "$OUT/02_rc_delta_summary.txt"
wc -l < "$OUT/02_rc_delta.txt" >> "$OUT/02_rc_delta_summary.txt"
if [[ -s "$OUT/02_rc_delta.txt" ]]; then
  echo "frozen RC delta detected" >&2
  exit 1
fi

git -C "$WORKTREE" status --short > "$OUT/03_git_status_short.txt"
git -C "$WORKTREE" diff --check > "$OUT/04_git_diff_check.txt"

(
  cd "$WORKTREE"
  npm run check:component-rights
) > "$OUT/05_component_rights.out" 2>&1
(
  cd "$WORKTREE"
  npm run test:t07-provenance
) > "$OUT/06_t07_provenance_test.out" 2>&1
node --check "$WORKTREE/scripts/generate-t07-camoflox-sbom.mjs" > "$OUT/07_generator_syntax.out" 2>&1
node --check "$WORKTREE/scripts/generate-t07-camoflox-dependency-inventory.mjs" >> "$OUT/07_generator_syntax.out" 2>&1

set +e
"$GITLEAKS" detect --no-git --source "$SNAPSHOT" --redact --report-format json --report-path "$OUT/08_gitleaks_delta_redacted.json" > "$OUT/08_gitleaks_delta_redacted.out" 2>&1
delta_status=$?
set -e
printf 'gitleaks_exit=%s\n' "$delta_status" >> "$OUT/08_gitleaks_delta_redacted.out"
delta_findings=$( (grep -o '"RuleID"' "$OUT/08_gitleaks_delta_redacted.json" || true) | wc -l | tr -d ' ' )
printf 'gitleaks_findings=%s\n' "$delta_findings" > "$OUT/08_gitleaks_delta_findings.txt"
if [[ "$delta_status" -ne 0 || "$delta_findings" -ne 0 ]]; then
  echo "unexpected finding in T07 delta" >&2
  exit 1
fi

unzip -q "$CANDIDATE_ARCHIVE" -d "$CANDIDATE_SOURCE"
set +e
"$GITLEAKS" detect --no-git --source "$CANDIDATE_SOURCE" --redact --report-format json --report-path "$OUT/09_candidate_gitleaks_raw_redacted.json" > "$OUT/09_candidate_gitleaks_redacted.out" 2>&1
candidate_status=$?
set -e
sed "s|$CANDIDATE_SOURCE/||g" "$OUT/09_candidate_gitleaks_raw_redacted.json" > "$OUT/09_candidate_gitleaks_redacted.json"
rm -f "$OUT/09_candidate_gitleaks_raw_redacted.json"
printf 'gitleaks_exit=%s\n' "$candidate_status" >> "$OUT/09_candidate_gitleaks_redacted.out"
candidate_findings=$( (grep -o '"RuleID"' "$OUT/09_candidate_gitleaks_redacted.json" || true) | wc -l | tr -d ' ' )
printf 'gitleaks_findings=%s\n' "$candidate_findings" > "$OUT/09_candidate_gitleaks_findings.txt"
if [[ "$candidate_status" -ne 1 || "$candidate_findings" -ne 1 ]]; then
  echo "candidate scan no longer matches the documented single blocking alert" >&2
  exit 1
fi

printf '%s\n' \
  'PROV-01=PARTIAL: exact archive hash and private rights reference; source commit unavailable' \
  'PROV-02=PARTIAL: passive SBOM and locked dependency inventory only' \
  'PROV-03=PASS: Core remains unique; candidate remains non-integrated' \
  'PROV-04=BLOCKED: one redacted generic-api-key alert remains UNKNOWN' \
  'PROV-05=NOT_EVALUATED: candidate tests and binaries were not executed' \
  'PROV-06=PARTIAL: SBOM present; root license and notices absent' \
  'PROV-07=NOT_EVALUATED: no Go portage or product behavior introduced' \
  'DECISION=T07_PROVENANCE_BLOCKED_PENDING_EVIDENCE' > "$OUT/10_prov_status.txt"

cp "$WORKTREE/docs/T07_CAMOFLOX_PROVENANCE.md" "$OUT/11_t07_decision.md"
cp "$WORKTREE/docs/component-rights-register.json" "$OUT/12_component_rights_register.json"
cp "$WORKTREE/docs/T07_CAMOFLOX_PROVENANCE_SBOM.spdx.json" "$OUT/13_sbom.spdx.json"
cp "$WORKTREE/docs/T07_CAMOFLOX_DEPENDENCY_INVENTORY.json" "$OUT/14_dependency_inventory.json"
cp "$WORKTREE/docs/T07_EXTERNAL_SOURCES.md" "$OUT/15_external_sources.md"
cp "$WORKTREE/scripts/collect-t07-evidence.sh" "$OUT/16_collect_t07_evidence.sh"

rm -rf "$CANDIDATE_SOURCE" "$SNAPSHOT" "$WORKTREE"
trap - EXIT
PARENT="$(dirname "$OUT")"
OUT_NAME="$(basename "$OUT")"
PREARCHIVE="${OUT}.pre-rescan.zip"
VERIFY="${OUT}-verify"
(
  cd "$OUT"
  find . -type f ! -name SHA256SUMS -printf '%P\n' | LC_ALL=C sort | xargs -r sha256sum > SHA256SUMS
)
(
  cd "$PARENT"
  zip -q -r "$PREARCHIVE" "$OUT_NAME"
)
rm -rf "$VERIFY"
unzip -q "$PREARCHIVE" -d "$VERIFY"
(
  cd "$VERIFY/$OUT_NAME"
  sha256sum -c SHA256SUMS > /dev/null
)
set +e
"$GITLEAKS" detect --no-git --source "$VERIFY" --redact --report-format json --report-path "$OUT/19_archive_rescan_redacted.json" > "$OUT/19_archive_rescan_redacted.out" 2>&1
archive_status=$?
set -e
archive_findings=$( (grep -o '"RuleID"' "$OUT/19_archive_rescan_redacted.json" || true) | wc -l | tr -d ' ' )
printf 'gitleaks_exit=%s\ngitleaks_findings=%s\n' "$archive_status" "$archive_findings" >> "$OUT/19_archive_rescan_redacted.out"
if [[ "$archive_status" -ne 0 || "$archive_findings" -ne 0 ]]; then
  echo "unexpected finding in evidence archive" >&2
  exit 1
fi
rm -rf "$VERIFY" "$PREARCHIVE"
(
  cd "$OUT"
  find . -type f ! -name SHA256SUMS -printf '%P\n' | LC_ALL=C sort | xargs -r sha256sum > SHA256SUMS
)
(
  cd "$PARENT"
  zip -q -r "$ARCHIVE" "$OUT_NAME"
)
VERIFY="${OUT}-final-verify"
rm -rf "$VERIFY"
unzip -q "$ARCHIVE" -d "$VERIFY"
(
  cd "$VERIFY/$OUT_NAME"
  sha256sum -c SHA256SUMS > /dev/null
)
rm -rf "$VERIFY"
printf 'archive_sha256='; sha256sum "$ARCHIVE" | awk '{print $1}'
printf 'archive_path=%s\n' "$ARCHIVE"
