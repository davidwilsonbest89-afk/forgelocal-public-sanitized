#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-environment-closure
REPO="$WORK/repo"
SRC="$WORK/source-only"
OUT="$WORK/static-analysis"
SEM="$WORK/tools/semgrep-venv/bin/semgrep"
rm -rf "$SRC" "$OUT"
mkdir -p "$SRC" "$OUT"
(cd "$REPO" && tar --exclude=.git --exclude=evidence --exclude=docs --exclude=artifacts --exclude=snapshots --exclude=.github --exclude=node_modules -cf - .) | tar -xf - -C "$SRC"
rm -rf "$SRC/.git" "$SRC/evidence" "$SRC/docs" "$SRC/artifacts" "$SRC/snapshots" "$SRC/.github" "$SRC/node_modules"
{
 echo '=== STATIC ANALYSIS PREFLIGHT ==='; date -u --iso-8601=seconds
 echo "source=$SRC"
 echo "semgrep=$($SEM --version 2>&1)"
 shellcheck --version | head -4
 yamllint --version
 echo '=== SEMGREP ==='
 "$SEM" scan --config auto --json --output "$OUT/semgrep.json" "$SRC"; echo "SEMGREP_EXIT_CODE=$?"
 echo '=== SHELLCHECK ==='
 find "$SRC" -type f -name '*.sh' -print0 | sort -z | xargs -0 -r shellcheck -f json > "$OUT/shellcheck.json" 2> "$OUT/shellcheck.stderr"; echo "SHELLCHECK_EXIT_CODE=$?"
 echo '=== YAMLLINT ==='
 find "$SRC" -type f \( -name '*.yml' -o -name '*.yaml' \) -print0 | sort -z | xargs -0 -r yamllint -f parsable > "$OUT/yamllint.txt" 2>&1; echo "YAMLLINT_EXIT_CODE=$?"
 echo '=== FILE INVENTORY ==='
 find "$SRC" -type f \( -name '*.sh' -o -name '*.yml' -o -name '*.yaml' \) -printf '%P\n' | sort
} > "$OUT/static-analysis-raw.log" 2>&1
cat "$OUT/static-analysis-raw.log"
