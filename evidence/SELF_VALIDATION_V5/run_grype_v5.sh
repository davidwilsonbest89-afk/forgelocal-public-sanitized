#!/usr/bin/env bash
set -u
BASE=/home/ubuntu/forgelocal_self_validation_v5
REPO=$BASE/repo
E=$BASE/evidence
GRYPE=/home/ubuntu/bin/grype-0.117.0
cd "$REPO"
CYCLONE=evidence/SELF_VALIDATION_V4/SELF_VALIDATION_SBOM.cdx.json
SPDX=evidence/SELF_VALIDATION_V4/SELF_VALIDATION_SBOM.spdx.json
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "grype=$($GRYPE version 2>&1)"
  echo "cyclonedx=$CYCLONE"
  echo "spdx=$SPDX"
  set +e
  "$GRYPE" "sbom:$CYCLONE" -o json > "$E/GRYPE_CYCLONEDX.json" 2> "$E/GRYPE_CYCLONEDX.stderr"
  RC1=$?
  echo "grype_cyclonedx_exit_code=$RC1"
  "$GRYPE" "sbom:$SPDX" -o json > "$E/GRYPE_SPDX.json" 2> "$E/GRYPE_SPDX.stderr"
  RC2=$?
  echo "grype_spdx_exit_code=$RC2"
  set -e
  echo '--- cyclonedx stderr tail'
  tail -20 "$E/GRYPE_CYCLONEDX.stderr"
  echo '--- spdx stderr tail'
  tail -20 "$E/GRYPE_SPDX.stderr"
} > "$E/GRYPE_RAW.log" 2>&1
