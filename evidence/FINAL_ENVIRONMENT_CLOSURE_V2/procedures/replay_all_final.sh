#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-environment-closure
REPO="$WORK/repo"
SRC="$WORK/source-only-final"
OUT="$WORK/final-results"
BIN="$WORK/tools/bin"
SEM="$WORK/tools/semgrep-venv/bin/semgrep"
LOG="$WORK/FINAL_ENVIRONMENT_TESTS_RAW.log"
export PATH="$WORK/tools/go/bin:$BIN:$PATH"
export GOTOOLCHAIN=local
rm -rf "$SRC" "$OUT"
mkdir -p "$SRC" "$OUT"
(cd "$REPO" && tar --exclude=.git --exclude=evidence --exclude=docs --exclude=artifacts --exclude=snapshots --exclude=.github --exclude=node_modules -cf - .) | tar -xf - -C "$SRC"
: > "$LOG"
run() {
  name="$1"; shift
  echo "=== $name ===" >> "$LOG"
  echo "COMMAND=$*" >> "$LOG"
  date -u --iso-8601=seconds >> "$LOG"
  timeout 300s "$@" >> "$LOG" 2>&1
  echo "EXIT_CODE=$?" >> "$LOG"
}
cd "$REPO"
run GO_TEST_RACE go test -count=1 -race ./cmd/... ./internal/...
run GO_VET go vet ./cmd/... ./internal/...
run GO_BUILD go build ./cmd/... ./internal/...
run PNPM_DASHBOARD_CHECK pnpm --dir forge-dashboard run check
run GIT_DIFF_CHECK git diff --check
run GOSEC_SOURCE_ONLY "$BIN/gosec" -fmt json -out "$OUT/gosec.json" "$SRC/cmd/..." "$SRC/internal/..."
run GITLEAKS_SOURCE_ONLY "$BIN/gitleaks" detect --source "$SRC" --no-git --no-banner --redact --report-format json --report-path "$OUT/gitleaks-source.json"
run GITLEAKS_EXTRACTION_NO_GIT "$BIN/gitleaks" detect --source "$SRC" --no-git --no-banner --redact --report-format json --report-path "$OUT/gitleaks-extraction.json"
run GOVULNCHECK_GO "$BIN/govulncheck" -json ./cmd/... ./internal/...
run OSV_GO_AND_PNPM "$BIN/osv-scanner" scan source -r "$SRC" --format json
run OSV_PNPM "$BIN/osv-scanner" scan source -r "$SRC/forge-dashboard" --format json
run TRIVY_FILESYSTEM "$BIN/trivy" fs --scanners vuln,secret,misconfig --format json --output "$OUT/trivy.json" --skip-dirs "$SRC/.git" "$SRC"
run SYFT_JSON "$BIN/syft" scan "dir:$SRC" -o "json=$OUT/syft.json"
run SYFT_CDX "$BIN/syft" scan "dir:$SRC" -o "cyclonedx-json=$OUT/syft.cdx.json"
run GRYPE_SYFT "$BIN/grype" sbom:"$OUT/syft.json" -o json
run SEMGREP "$SEM" scan --config auto --json --output "$OUT/semgrep.json" "$SRC"
find "$SRC" -type f -name '*.sh' -print0 | sort -z | xargs -0 -r shellcheck -f json > "$OUT/shellcheck.json" 2> "$OUT/shellcheck.stderr"; echo "SHELLCHECK_EXIT_CODE=$?" >> "$LOG"
find "$SRC" -type f \( -name '*.yml' -o -name '*.yaml' \) -print0 | sort -z | xargs -0 -r yamllint -f parsable > "$OUT/yamllint.txt" 2>&1; echo "YAMLLINT_EXIT_CODE=$?" >> "$LOG"
