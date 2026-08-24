#!/usr/bin/env bash
set +e
export PATH=/home/ubuntu/toolchain/go1.25.13/bin:/home/ubuntu/go/bin:/home/ubuntu/bin:/home/ubuntu/.nvm/versions/node/v22.13.0/bin:$PATH
export GOTOOLCHAIN=local CGO_ENABLED=1 GIT_LFS_SKIP_SMUDGE=1
repo=/home/ubuntu/forgelocal-prehuman-fresh-20260824
out=/home/ubuntu/forgelocal-final-check-20260824
log=$out/PREHUMAN_FINAL_LINTER_JSON_RAW.log
: > "$log"
run_in() {
  label=$1; root=$2; shift 2
  printf 'started_utc=%s\ncwd=%s\nhead=%s\ncommand=' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$root" "$(git -C "$root" rev-parse HEAD 2>&1)" >> "$log"
  printf '%q ' "$@" >> "$log"; printf '\n--- raw_output ---\n' >> "$log"
  (cd "$root" && "$@") >> "$log" 2>&1; rc=$?
  printf 'exit_code=%s\ncompleted_utc=%s\ncheck=%s\n---\n' "$rc" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$label" >> "$log"
}
parent=$(git -C "$repo" rev-parse 't00-t27-complete-20260820^{commit}')
base=$(mktemp -d)
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" init --quiet
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" fetch --quiet "$repo" "$parent:refs/heads/baseline"
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" checkout --quiet --detach "$parent"
run_in staticcheck_baseline "$base" bash -c 'staticcheck -f json ./... > "$1"' bash "$out/PREHUMAN_FINAL_STATICCHECK_BASELINE.jsonl"
run_in golangci_baseline "$base" bash -c 'golangci-lint run --out-format json > "$1"' bash "$out/PREHUMAN_FINAL_GOLANGCI_BASELINE.json"
run_in staticcheck_head "$repo" bash -c 'staticcheck -f json ./... > "$1"' bash "$out/PREHUMAN_FINAL_STATICCHECK_HEAD.jsonl"
run_in golangci_head "$repo" bash -c 'golangci-lint run --out-format json > "$1"' bash "$out/PREHUMAN_FINAL_GOLANGCI_HEAD.json"
rm -rf "$base"
