#!/usr/bin/env bash
set -u
BASE=/home/ubuntu/forgelocal_self_validation_v5
ROOT=$BASE/repo
E=$BASE/evidence
WRAPPER=$ROOT/evidence/forgelocal-t00-t42-self-validation-v5-enhanced.zip
SIDECAR=$ROOT/evidence/forgelocal-t00-t42-self-validation-v5-enhanced.zip.portable.sha256
EXTRACT=$BASE/fresh_extract_v5
rm -rf "$EXTRACT"
mkdir -p "$EXTRACT"
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "wrapper=$WRAPPER"
  echo '--- unzip test'
  unzip -t "$WRAPPER"
  echo "unzip_test_exit_code=$?"
  echo '--- sidecar'
  (cd "$(dirname "$WRAPPER")" && sha256sum -c "$(basename "$SIDECAR")")
  echo "sidecar_verify_exit_code=$?"
  echo '--- fresh extraction'
  unzip -q "$WRAPPER" -d "$EXTRACT"
  echo "fresh_extract_exit_code=$?"
  echo "fresh_file_count=$(find "$EXTRACT" -type f | wc -l)"
} > "$BASE/SELF_VALIDATION_V5_WRAPPER_VERIFY.log" 2>&1
set +e
/home/ubuntu/bin/gitleaks detect --no-banner --redact --no-git --source "$EXTRACT" --report-format json --report-path "$E/GITLEAKS_V5_WRAPPER_FRESH_RESCAN.json" > "$E/GITLEAKS_V5_WRAPPER_FRESH_RESCAN.log" 2>&1
RC=$?
printf 'gitleaks_wrapper_fresh_rescan_exit_code=%s\n' "$RC" >> "$BASE/SELF_VALIDATION_V5_WRAPPER_VERIFY.log"
printf 'wrapper_verify_script_exit_code=0\n' >> "$BASE/SELF_VALIDATION_V5_WRAPPER_VERIFY.log"
