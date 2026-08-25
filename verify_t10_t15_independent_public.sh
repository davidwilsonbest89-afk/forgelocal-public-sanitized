#!/usr/bin/env bash
set -u
set -o pipefail
BASE=/home/ubuntu/forgelocal_self_validation_v6
REPO=https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized.git
BR=audit/t00-t42-t10-t15-independent-review-v6
EXPECTED=8b2b5f567b4bf19ca1172c3331b84a697e7059c3
SOURCE=8c38830b155f293c8b05b2789e60f7de4c45f565
PKG=evidence/t10-t15-independent-review-v6.zip
SIDE=evidence/t10-t15-independent-review-v6.zip.portable.sha256
BUNDLE=evidence/t10-t15-independent-review-v6.delta.bundle
BSIDE=evidence/t10-t15-independent-review-v6.delta.bundle.portable.sha256
FOLDER=evidence/T10-T15-INDEPENDENT-REVIEW
ROOT=$BASE/t10_t15_independent_review
WORK=$(mktemp -d /tmp/t10-t15-independent-public.XXXXXX)
CLONE=$WORK/clone
NEUTRAL=$WORK/neutral
EXTRACT=$WORK/extract
LOG=$BASE/T10_T15_INDEPENDENT_PUBLIC_VERIFY.log
: > "$LOG"
OVERALL=0
run() { local n=$1; shift; "$@" >>"$LOG" 2>&1; local r=$?; echo "${n}_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1; }
{
 echo "utc=$(date -u +%FT%TZ)"; echo "cwd=$(pwd)"; echo "command=$0"; echo "branch=$BR"; echo "expected_head=$EXPECTED"; echo "source=$SOURCE"; echo "go=$(/usr/local/go1.25.13/bin/go version 2>&1 || true)"; echo "node=$(node --version 2>&1 || true)"; echo "pnpm=$(pnpm --version 2>&1 || true)"; echo "playwright=$(cd /home/ubuntu/forgelocal_self_validation_v6/repo/forge-dashboard && pnpm exec playwright --version 2>&1 || true)"; echo "gitleaks=$(/home/ubuntu/bin/gitleaks version 2>&1 || true)"; echo "disk=$(df -h /home/ubuntu | tail -1)";
} >> "$LOG"
R=$(gh api "repos/davidwilsonbest89-afk/forgelocal-public-sanitized/git/ref/heads/$BR" --jq .object.sha 2>>"$LOG"); r=$?; echo "remote_api_exit=$r remote_head=$R" | tee -a "$LOG"; [ "$r" -eq 0 ] && [ "$R" = "$EXPECTED" ] || OVERALL=1
GIT_LFS_SKIP_SMUDGE=1 GIT_TERMINAL_PROMPT=0 git clone --depth=1 --filter=blob:none --no-checkout --branch "$BR" --single-branch "$REPO" "$CLONE" >>"$LOG" 2>&1; r=$?; echo "fresh_clone_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
run fresh_clone_fsck git -C "$CLONE" fsck --full
H=$(git -C "$CLONE" rev-parse HEAD 2>>"$LOG"); echo "clone_head=$H head_match=$([ "$H" = "$EXPECTED" ] && echo yes || echo no)" | tee -a "$LOG"; [ "$H" = "$EXPECTED" ] || OVERALL=1
run sparse_init git -C "$CLONE" sparse-checkout init --no-cone
run sparse_set git -C "$CLONE" sparse-checkout set evidence/T10-T15-INDEPENDENT-REVIEW "$PKG" "$SIDE" "$BUNDLE" "$BSIDE"
run exact_checkout git -C "$CLONE" checkout --detach "$EXPECTED"
run targeted_lfs git -C "$CLONE" lfs pull origin "$BR" --include="$PKG,$BUNDLE" --exclude=''
mkdir -p "$NEUTRAL" "$EXTRACT"
cp "$CLONE/$PKG" "$NEUTRAL/$(basename "$PKG")" >>"$LOG" 2>&1; r=$?; echo "zip_copy_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
cp "$CLONE/$SIDE" "$NEUTRAL/$(basename "$SIDE")" >>"$LOG" 2>&1; r=$?; echo "zip_sidecar_copy_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
sed -i "s#${PKG}#$(basename "$PKG")#" "$NEUTRAL/$(basename "$SIDE")"
(cd "$NEUTRAL" && sha256sum -c "$(basename "$SIDE")") >>"$LOG" 2>&1; r=$?; echo "zip_sidecar_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
run unzip_test unzip -t "$NEUTRAL/$(basename "$PKG")"
run fresh_extract unzip -q "$NEUTRAL/$(basename "$PKG")" -d "$EXTRACT"
EX="$EXTRACT/$(basename "$FOLDER")"
run internal_sha bash -c "cd '$EX' && sha256sum -c SHA256SUMS"
! grep -qE '^\| `(MANIFEST\.md|SHA256SUMS)` \|' "$EX/MANIFEST.md"; r=$?; echo "manifest_non_self_referential_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
cp "$CLONE/$BUNDLE" "$NEUTRAL/$(basename "$BUNDLE")" >>"$LOG" 2>&1; r=$?; echo "bundle_copy_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
cp "$CLONE/$BSIDE" "$NEUTRAL/$(basename "$BSIDE")" >>"$LOG" 2>&1; r=$?; echo "bundle_sidecar_copy_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
sed -i "s#${BUNDLE}#$(basename "$BUNDLE")#" "$NEUTRAL/$(basename "$BSIDE")"
(cd "$NEUTRAL" && sha256sum -c "$(basename "$BSIDE")") >>"$LOG" 2>&1; r=$?; echo "bundle_sidecar_exit=$r" | tee -a "$LOG"; [ "$r" -eq 0 ] || OVERALL=1
run bundle_verify git -C "$CLONE" bundle verify "$BUNDLE"
run clean_worktree git -C "$CLONE" status --porcelain
run nonempty_range git -C "$CLONE" rev-list --count "$SOURCE..$EXPECTED"
run gitleaks_extracted /home/ubuntu/bin/gitleaks detect --no-banner --redact --no-git --source "$EX" --report-format json --report-path "$ROOT/evidence/T10-T15-INDEPENDENT-REVIEW/GITLEAKS_PUBLIC_EXTRACTED.json"
run gitleaks_range /home/ubuntu/bin/gitleaks detect --no-banner --redact --source "$CLONE" --log-opts "$SOURCE..$EXPECTED" --report-format json --report-path "$ROOT/evidence/T10-T15-INDEPENDENT-REVIEW/GITLEAKS_PUBLIC_RANGE.json"
echo "overall_exit=$OVERALL" | tee "$LOG"
rm -rf "$WORK"
exit "$OVERALL"
