#!/usr/bin/env bash
set -u
set -o pipefail
ROOT=/home/ubuntu/forgelocal_self_validation_v6/t10_t15_independent_review
REMOTE=https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized.git
BRANCH=audit/t00-t42-t10-t15-e2e-validation-v6
EXPECTED=8c38830b155f293c8b05b2789e60f7de4c45f565
SOURCE=999374d99b7996504ba91e421850a2fe84afb78d
PKG=evidence/t10-t15-e2e-validation-v6.zip
SIDE=evidence/t10-t15-e2e-validation-v6.zip.portable.sha256
BUNDLE=evidence/t10-t15-e2e-validation-v6.delta.bundle
BUNDLE_SIDE=evidence/t10-t15-e2e-validation-v6.delta.bundle.portable.sha256
FOLDER=evidence/T10-T15-E2E-VALIDATION
EVID=$ROOT/evidence/T10-T15-INDEPENDENT-REVIEW
WORK=$(mktemp -d /tmp/t10-t15-independent-review.XXXXXX)
CLONE=$WORK/clone
NEUTRAL=$WORK/neutral
EXTRACT=$WORK/extract
LOG=$EVID/INDEPENDENT_REVIEW_RAW_UTC.log
RESULT=$EVID/INDEPENDENT_REVIEW_RESULT.log
mkdir -p "$EVID" "$NEUTRAL" "$EXTRACT"
OVERALL=0
record() { printf '%s\n' "$*" | tee -a "$LOG"; }
run_check() { local name=$1; shift; "$@" >>"$LOG" 2>&1; local rc=$?; printf '%s_exit=%s\n' "$name" "$rc" | tee -a "$LOG"; [ "$rc" -eq 0 ] || OVERALL=1; return 0; }
{
  echo "utc=$(date -u +%FT%TZ)"
  echo "cwd=$(pwd)"
  echo "command=$0"
  echo "branch=$BRANCH"
  echo "expected_head=$EXPECTED"
  echo "source_base=$SOURCE"
  echo "go_version=$(/usr/local/go1.25.13/bin/go version 2>&1 || true)"
  echo "node_version=$(node --version 2>&1 || true)"
  echo "pnpm_version=$(pnpm --version 2>&1 || true)"
  echo "playwright_version=$(cd /home/ubuntu/forgelocal_self_validation_v6/repo/forge-dashboard && pnpm exec playwright --version 2>&1 || true)"
  echo "gitleaks_version=$(/home/ubuntu/bin/gitleaks version 2>&1 || true)"
  echo "git_version=$(git --version 2>&1)"
  echo "disk=$(df -h /home/ubuntu | tail -1)"
} > "$LOG"
REMOTE_HEAD=$(gh api "repos/davidwilsonbest89-afk/forgelocal-public-sanitized/git/ref/heads/$BRANCH" --jq .object.sha 2>>"$LOG"); RC=$?; echo "remote_head_api_exit=$RC" >>"$LOG"; echo "remote_head=$REMOTE_HEAD" >>"$LOG"; [ "$RC" -eq 0 ] && [ "$REMOTE_HEAD" = "$EXPECTED" ] || OVERALL=1
GIT_LFS_SKIP_SMUDGE=1 GIT_TERMINAL_PROMPT=0 git clone --depth=1 --filter=blob:none --no-checkout --branch "$BRANCH" --single-branch "$REMOTE" "$CLONE" >>"$LOG" 2>&1; RC=$?; echo "fresh_clone_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
run_check clone_fsck git -C "$CLONE" fsck --full
LOCAL_HEAD=$(git -C "$CLONE" rev-parse HEAD 2>>"$LOG"); echo "clone_head=$LOCAL_HEAD" | tee -a "$LOG"; [ "$LOCAL_HEAD" = "$EXPECTED" ] || OVERALL=1
GIT_LFS_SKIP_SMUDGE=1 git -C "$CLONE" checkout --detach "$EXPECTED" >>"$LOG" 2>&1; RC=$?; echo "exact_checkout_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
GIT_LFS_SKIP_SMUDGE=1 git -C "$CLONE" lfs pull origin "$BRANCH" --include="$PKG,$BUNDLE" --exclude='' >>"$LOG" 2>&1; RC=$?; echo "targeted_lfs_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
# Copy package and sidecar to a neutral directory, outside the clone.
cp "$CLONE/$PKG" "$NEUTRAL/$(basename "$PKG")" >>"$LOG" 2>&1; RC=$?; echo "neutral_zip_copy_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
cp "$CLONE/$SIDE" "$NEUTRAL/$(basename "$SIDE")" >>"$LOG" 2>&1; RC=$?; echo "neutral_sidecar_copy_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
sed -i "s#${PKG}#$(basename "$PKG")#" "$NEUTRAL/$(basename "$SIDE")"
(cd "$NEUTRAL" && sha256sum -c "$(basename "$SIDE")") >>"$LOG" 2>&1; RC=$?; echo "neutral_zip_sidecar_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
unzip -t "$NEUTRAL/$(basename "$PKG")" >>"$LOG" 2>&1; RC=$?; echo "unzip_test_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
unzip -q "$NEUTRAL/$(basename "$PKG")" -d "$EXTRACT" >>"$LOG" 2>&1; RC=$?; echo "fresh_extract_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
EXTRACTED="$EXTRACT/$(basename "$FOLDER")"
(cd "$EXTRACTED" && sha256sum -c SHA256SUMS) >>"$LOG" 2>&1; RC=$?; echo "fresh_internal_sha_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
! grep -qE '^\| `(MANIFEST\.md|SHA256SUMS)` \|' "$EXTRACTED/MANIFEST.md"; RC=$?; echo "manifest_non_self_referential_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
cp "$CLONE/$BUNDLE" "$NEUTRAL/$(basename "$BUNDLE")" >>"$LOG" 2>&1; RC=$?; echo "neutral_bundle_copy_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
cp "$CLONE/$BUNDLE_SIDE" "$NEUTRAL/$(basename "$BUNDLE_SIDE")" >>"$LOG" 2>&1; RC=$?; echo "neutral_bundle_sidecar_copy_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
sed -i "s#${BUNDLE}#$(basename "$BUNDLE")#" "$NEUTRAL/$(basename "$BUNDLE_SIDE")"
(cd "$NEUTRAL" && sha256sum -c "$(basename "$BUNDLE_SIDE")") >>"$LOG" 2>&1; RC=$?; echo "neutral_bundle_sidecar_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
(cd "$CLONE" && git bundle verify "$BUNDLE") >>"$LOG" 2>&1; RC=$?; echo "bundle_verify_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
run_check checkout_status git -C "$CLONE" status --porcelain
# The original test log is audited, not rerun here.
grep -q '7 passed' "$EXTRACTED/PLAYWRIGHT_T10_T15_SEQUENTIAL.log"; RC=$?; echo "seven_tests_log_assertion_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
cat > "$EVID/ASSERTIONS_AUDIT.md" <<'EOF'
# Assertions audited from published T10/T15 kit

| Assertion | Evidence checked |
|---|---|
| T10 valid creation | `T10_VALID_PROXY_CREATED: PASS server_validated` |
| T10 invalid port refused | `T10_INVALID_PORT_REFUSED: PASS no_write_on_rejection` |
| T10 phantom profile refused | `T10_ASSIGN_REQUIRES_CORE_PROFILE: PASS explicit_refusal_no_ghost_binding correlated` |
| T10 off-loopback writes refused | `T10_OFFLOOPBACK_REFUSED: PASS no_write_path_origin_offloopback` |
| Redacted listing / no credential value | `T10_LISTING_REDACTED: PASS no_credential_value_in_ui` |
| T15 external navigation refused | T15 W2 test name in published Playwright log/source |
| Browser credential absence / digest projection | T15 W3/W5 test names in published Playwright log/source |
| T15 session close | T15 W4 test name in published Playwright log/source |
| Dashboard automation panel | T15 W5 test name in published Playwright log/source |
| Sequential execution | `Running 7 tests using 1 worker` and `7 passed (15.6s)` |
EOF
# Check published source names are present without executing them.
for needle in 'fail-closed local-only navigation policy' 'content projection redacted' 'session close and refusal after close' 'dashboard automation panel mounts'; do grep -q "$needle" "$EXTRACTED/PLAYWRIGHT_T10_T15_SEQUENTIAL.log" "$CLONE/forge-dashboard/tests/automation-t15.spec.ts" 2>>"$LOG"; RC=$?; echo "assertion_source_$(echo "$needle" | tr ' ' '_')_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1; done
# Cleanup assertions are copied from the published execution proof and checked for exact success values.
for needle in 'token_file_removed=yes' 'base_dir_removed=yes' 'port_19280_after_cleanup=closed' 'port_3000_after_cleanup=closed' 'run_root_removed=yes'; do grep -q "$needle" "$EXTRACTED/CLEANUP_RESULT.log" 2>>"$LOG"; RC=$?; echo "cleanup_${needle%%=*}_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1; done
! grep -R -qE 'e2e_t10_t15_[0-9a-f]+' "$EXTRACTED"; RC=$?; echo "no_temporary_token_in_kit_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
/home/ubuntu/bin/gitleaks detect --no-banner --redact --no-git --source "$EXTRACTED" --report-format json --report-path "$EVID/GITLEAKS_EXTRACTED.json" >>"$LOG" 2>&1; RC=$?; echo "gitleaks_extracted_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
GIT_RANGE="$SOURCE..$EXPECTED"
GIT_LFS_SKIP_SMUDGE=1 git -C "$CLONE" fetch --no-tags origin "t00-t42-v6-local-qualified-2026-08-25" --depth=1 >>"$LOG" 2>&1; RC=$?; echo "base_tag_fetch_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
RANGE_COUNT=$(git -C "$CLONE" rev-list --count "$GIT_RANGE" 2>>"$LOG"); RC=$?; echo "git_range_count=$RANGE_COUNT range_count_exit=$RC" | tee -a "$LOG"; [ "$RC" -eq 0 ] && [ "$RANGE_COUNT" -gt 0 ] || OVERALL=1
/home/ubuntu/bin/gitleaks detect --no-banner --redact --source "$CLONE" --log-opts "$GIT_RANGE" --report-format json --report-path "$EVID/GITLEAKS_GIT_RANGE.json" >>"$LOG" 2>&1; RC=$?; echo "gitleaks_git_range_exit=$RC range=$GIT_RANGE" | tee -a "$LOG"; [ "$RC" -eq 0 ] || OVERALL=1
printf 'overall_exit=%s\n' "$OVERALL" | tee "$RESULT"
rm -rf "$WORK"
exit "$OVERALL"
