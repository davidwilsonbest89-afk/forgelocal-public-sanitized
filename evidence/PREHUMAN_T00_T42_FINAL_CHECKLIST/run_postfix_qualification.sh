#!/usr/bin/env bash
set +e
export PATH=/home/ubuntu/toolchain/go1.25.13/bin:/home/ubuntu/go/bin:/home/ubuntu/bin:/home/ubuntu/.nvm/versions/node/v22.13.0/bin:$PATH
export GOTOOLCHAIN=local CGO_ENABLED=1 GIT_LFS_SKIP_SMUDGE=1
repo=/home/ubuntu/forgelocal-postfix-fresh-20260824
out=/home/ubuntu/forgelocal-postfix-final-check-20260824
mkdir -p "$out"
log="$out/POSTFIX_QUALIFICATION_RAW.log"
: > "$log"
run() {
  label=$1; shift
  printf 'started_utc=%s\ncwd=%s\nhead=%s\ncommand=' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$PWD" "$(git -C "$repo" rev-parse HEAD 2>&1)" >> "$log"
  printf '%q ' "$@" >> "$log"
  printf '\n--- raw_output ---\n' >> "$log"
  "$@" >> "$log" 2>&1
  rc=$?
  printf 'exit_code=%s\ncompleted_utc=%s\ncheck=%s\nfree_after_kib=%s\n---\n' "$rc" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$label" "$(df -Pk / | awk 'NR==2 {print $4}')" >> "$log"
}
cd "$repo"
run worktree git status --short --branch
run head git rev-parse HEAD
run fsck git fsck --full
run lfs_fsck git lfs fsck
run go_version go version
run go_mod_verify go mod verify
run go_list_modules go list -m -json all
run go_test_race go test -count=1 -race ./...
run go_vet go vet ./...
run go_build go build ./...
run staticcheck bash -c 'staticcheck -f json ./... > "$1"; exit $?' bash "$out/POSTFIX_STATICCHECK.jsonl"
run golangci_lint bash -c 'golangci-lint run --out-format json ./... > "$1"; exit $?' bash "$out/POSTFIX_GOLANGCI.json"
run govulncheck govulncheck ./...
run osv_scanner osv-scanner scan source -r .
run trivy_fs trivy fs --scanners vuln,secret,misconfig .
run syft_cyclonedx bash -c 'syft dir:"$1" -o cyclonedx-json="$2"' bash "$repo" "$out/POSTFIX_SBOM.cdx.json"
run syft_spdx bash -c 'syft dir:"$1" -o spdx-json="$2"' bash "$repo" "$out/POSTFIX_SBOM.spdx.json"
run gitleaks_inventory bash -c 'git diff --name-status "t00-t27-complete-20260820..HEAD"' bash
run gitleaks_cumulative bash -c 'git diff --binary "t00-t27-complete-20260820..HEAD" | gitleaks detect --pipe --no-banner --redact --report-format json --report-path "$1/POSTFIX_GITLEAKS.json"' bash "$out"
base=$(mktemp -d)
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" init --quiet
parent=$(git -C "$repo" rev-parse 't00-t27-complete-20260820^{commit}')
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" fetch --quiet "$repo" "$parent:refs/heads/baseline"
GIT_LFS_SKIP_SMUDGE=1 git -C "$base" checkout --quiet --detach "$parent"
run gosec_baseline bash -c 'cd "$1" && gosec -fmt=json -out="$2/POSTFIX_GOSEC_BASELINE.json" ./...' bash "$base" "$out"
run gosec_head bash -c 'cd "$1" && gosec -fmt=json -out="$2/POSTFIX_GOSEC_HEAD.json" ./...' bash "$repo" "$out"
cp "$out/POSTFIX_GOSEC_BASELINE.json" "$out/PREHUMAN_FINAL_GOSEC_BASELINE.json"
cp "$out/POSTFIX_GOSEC_HEAD.json" "$out/PREHUMAN_FINAL_GOSEC_HEAD.json"
run gosec_normalized bash -c 'python3 /home/ubuntu/normalize_final_gosec.py "$1"; rm -f "$1/PREHUMAN_FINAL_GOSEC_BASELINE.json" "$1/PREHUMAN_FINAL_GOSEC_HEAD.json"' bash "$out"
rm -rf "$base"
run gate_text grep -R -n -E 'PUBLIC_RELEASE_BLOCKED|SCAN_BLOCKED_UNKNOWN|NATIVE_SYSTEMVAULT_NOT_TESTED|camoflox_execution_authorized=false|t08_authorized=false|release_authorized=false' --exclude-dir=.git --exclude-dir=node_modules .
run prohibited_names bash -c 'find evidence -type f | grep -Ei "(^|/)(\.env|.*(cookie|token|secret|sqlite|playwright|node_modules|dist|build|\.key|\.pem))" || true' bash
if [ -f forge-dashboard/package.json ]; then
  cd forge-dashboard
  run dashboard_install pnpm install --frozen-lockfile
  run dashboard_tsc pnpm exec tsc --noEmit
  run dashboard_build pnpm run build
  run dashboard_audit pnpm audit --prod
  run dashboard_playwright pnpm exec playwright test --workers=1
fi
cd "$repo"
run final_worktree git status --short --branch
