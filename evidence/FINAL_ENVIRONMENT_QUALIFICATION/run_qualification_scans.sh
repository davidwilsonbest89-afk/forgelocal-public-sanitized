#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-env-qualification
REPO="$WORK/repo"
OUT="$WORK/scan-results"
TOOLS="$WORK/tools/bin"
SEM="$WORK/tools/semgrep-venv/bin/semgrep"
export PATH="$WORK/tools/go/bin:$TOOLS:$PATH"
export GOTOOLCHAIN=local
mkdir -p "$OUT"
rm -rf "$WORK/source-only"
mkdir -p "$WORK/source-only"
(cd "$REPO" && git archive --format=tar HEAD) | tar -xf - -C "$WORK/source-only"
rm -rf "$WORK/source-only/.git" "$WORK/source-only/evidence" "$WORK/source-only/docs" "$WORK/source-only/artifacts" "$WORK/source-only/snapshots" "$WORK/source-only/.github"
LOG="$WORK/FINAL_ENVIRONMENT_TESTS_RAW.log"
run() {
  name="$1"; shift
  {
    echo "=== $name ==="
    echo "COMMAND=$*"
    echo "SCOPE=source-only extraction at $WORK/source-only unless otherwise stated"
    date -u --iso-8601=seconds
    "$@"
    rc=$?
    echo "EXIT_CODE=$rc"
    echo
  } >> "$LOG" 2>&1
}
{
  echo '=== SCANNER INVENTORY ==='
  date -u --iso-8601=seconds
  go version
  "$TOOLS/gosec" -version 2>&1 || true
  "$TOOLS/gitleaks" version 2>&1 || true
  "$TOOLS/govulncheck" -version 2>&1 || true
  "$TOOLS/grype" version 2>&1 || true
  "$TOOLS/syft" version 2>&1 || true
  "$TOOLS/trivy" version 2>&1 || true
  "$TOOLS/osv-scanner" --version 2>&1 || true
  "$SEM" --version 2>&1 || true
  shellcheck --version 2>&1 | head -4 || true
  yamllint --version 2>&1 || true
} >> "$LOG" 2>&1
run GOSEC_SOURCE_ONLY "$TOOLS/gosec" -fmt json -out "$OUT/gosec-source-only.json" ./cmd/... ./internal/...
run GITLEAKS_SOURCE_ONLY "$TOOLS/gitleaks" detect --source "$WORK/source-only" --no-git --no-banner --redact --report-format json --report-path "$OUT/gitleaks-source-only.json"
run SEMGREP_SOURCE_ONLY "$SEM" scan --config auto --json --output "$OUT/semgrep-source-only.json" "$WORK/source-only"
find "$WORK/source-only" -type f -name '*.sh' -print0 | xargs -0 -r shellcheck -f json > "$OUT/shellcheck-source-only.json" 2> "$OUT/shellcheck-source-only.stderr"
sc=$?
echo "=== SHELLCHECK_SOURCE_ONLY ===" >> "$LOG"; echo "COMMAND=find source-only '*.sh' | xargs shellcheck -f json" >> "$LOG"; echo "SCOPE=source-only shell scripts" >> "$LOG"; cat "$OUT/shellcheck-source-only.json" >> "$LOG"; cat "$OUT/shellcheck-source-only.stderr" >> "$LOG"; echo "EXIT_CODE=$sc" >> "$LOG"
find "$WORK/source-only" -type f \( -name '*.yml' -o -name '*.yaml' \) -print0 | xargs -0 -r yamllint -f parsable > "$OUT/yamllint-source-only.txt" 2>&1
sc=$?
echo "=== YAMLLINT_SOURCE_ONLY ===" >> "$LOG"; echo "COMMAND=find source-only '*.yml' '*.yaml' | xargs yamllint -f parsable" >> "$LOG"; echo "SCOPE=source-only YAML" >> "$LOG"; cat "$OUT/yamllint-source-only.txt" >> "$LOG"; echo "EXIT_CODE=$sc" >> "$LOG"
run GOVULNCHECK_GO_SOURCE_ONLY "$TOOLS/govulncheck" -json ./cmd/... ./internal/...
run OSV_SOURCE_RECURSIVE "$TOOLS/osv-scanner" scan source -r "$WORK/source-only" --format json
run TRIVY_FILESYSTEM_SOURCE_ONLY "$TOOLS/trivy" fs --scanners vuln,secret,misconfig --format json --output "$OUT/trivy-filesystem-source-only.json" --skip-dirs "$WORK/source-only/.git" --skip-dirs "$WORK/source-only/evidence" --skip-dirs "$WORK/source-only/docs" "$WORK/source-only"
run SYFT_SOURCE_ONLY_JSON "$TOOLS/syft" scan "dir:$WORK/source-only" -o "json=$OUT/syft-source-only.json"
run SYFT_SOURCE_ONLY_CYCLONEDX "$TOOLS/syft" scan "dir:$WORK/source-only" -o "cyclonedx-json=$OUT/syft-source-only.cdx.json"
run GRYPE_SYFT_SOURCE_ONLY "$TOOLS/grype" sbom:"$OUT/syft-source-only.json" -o json
printf '=== SOURCE-ONLY FILE INVENTORY ===\n' >> "$LOG"
find "$WORK/source-only" -maxdepth 2 -type f -printf '%P\n' | sort >> "$LOG"
