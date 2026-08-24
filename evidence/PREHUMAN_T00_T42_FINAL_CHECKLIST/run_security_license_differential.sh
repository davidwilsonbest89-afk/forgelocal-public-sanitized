#!/usr/bin/env bash
set +e
export PATH=/home/ubuntu/toolchain/go1.25.13/bin:/home/ubuntu/go/bin:/home/ubuntu/bin:/home/ubuntu/.nvm/versions/node/v22.13.0/bin:$PATH
export GOTOOLCHAIN=local CGO_ENABLED=1 GIT_LFS_SKIP_SMUDGE=1
repo=/home/ubuntu/forgelocal-prehuman-fresh-20260824
out=/home/ubuntu/forgelocal-final-check-20260824
log=$out/PREHUMAN_FINAL_SECURITY_LICENSE_RAW.log
: > "$log"
run() {
  label=$1; shift
  printf 'started_utc=%s\ncwd=%s\nhead=%s\ncommand=' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$PWD" "$(git -C "$repo" rev-parse HEAD 2>&1)" >> "$log"
  printf '%q ' "$@" >> "$log"
  printf '\n--- raw_output ---\n' >> "$log"
  "$@" >> "$log" 2>&1
  rc=$?
  printf 'exit_code=%s\ncompleted_utc=%s\ncheck=%s\n---\n' "$rc" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$label" >> "$log"
}
parent=$(git -C "$repo" rev-parse 't00-t27-complete-20260820^{commit}')
base=$(mktemp -d)
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" init --quiet
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" fetch --quiet "$repo" "$parent:refs/heads/baseline"
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" checkout --quiet --detach "$parent"
run trivy_baseline bash -c 'trivy fs --scanners vuln,secret,misconfig --format json --output "$2" "$1"' bash "$base" "$out/PREHUMAN_FINAL_TRIVY_BASELINE.json"
run trivy_head bash -c 'trivy fs --scanners vuln,secret,misconfig --format json --output "$2" "$1"' bash "$repo" "$out/PREHUMAN_FINAL_TRIVY_HEAD.json"
rm -rf "$base"
cd "$repo/forge-dashboard"
run license_inventory bash -c 'pnpm dlx license-checker --production --json' bash
