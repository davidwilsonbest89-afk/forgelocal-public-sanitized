#!/usr/bin/env bash
set -euo pipefail

# T30 full-race replay. This script is intentionally read-only with respect
# to product data: it runs Go tests only and never starts a browser/runtime.

required_kib=$((5 * 1024 * 1024))
available_kib=$(df -Pk . | awk 'NR==2 {print $4}')

printf 'started_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
printf 'cwd=%s\n' "$PWD"
printf 'git_head=%s\n' "$(git rev-parse HEAD)"
printf 'available_kib=%s\n' "$available_kib"
printf 'required_kib=%s\n' "$required_kib"

if (( available_kib < required_kib )); then
  printf '%s\n' 'status=BLOCKED_INSUFFICIENT_DISK_FOR_FULL_RACE_REPLAY'
  printf 'exit_code=%s\n' '64'
  exit 64
fi

printf '%s\n' 'command=go test -count=1 -race ./...'
set +e
go test -count=1 -race ./...
rc=$?
set -e
printf 'exit_code=%s\n' "$rc"
printf 'completed_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
exit "$rc"
