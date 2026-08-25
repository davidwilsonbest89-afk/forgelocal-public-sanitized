#!/usr/bin/env bash
set -u
BASE_DIR=/home/ubuntu/forgelocal_self_validation_v5
REPO=$BASE_DIR/repo
EVIDENCE=$BASE_DIR/evidence
BASE_TAG=t00-t27-complete-20260820
cd "$REPO"
HEAD=$(git rev-parse HEAD)
BASE=$(git rev-parse "$BASE_TAG")
LIST="$EVIDENCE/explicit_commit_trees.list"
git log --format='%H' --reverse "$BASE_TAG..$HEAD" > "$LIST"
TMP=$(mktemp -d /tmp/forgelocal-v5-gitleaks-trees.XXXXXX)
trap 'rm -rf "$TMP"' EXIT
TOTAL=0
PASS=0
FAIL=0
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$REPO"
  echo "base=$BASE"
  echo "head=$HEAD"
  echo "commits=$(wc -l < "$LIST")"
  echo 'method=git archive each commit tree; gitleaks detect --no-git --source <tree>'
  while IFS= read -r commit; do
    TOTAL=$((TOTAL+1))
    tree="$TMP/$commit"
    mkdir -p "$tree"
    GIT_LFS_SKIP_SMUDGE=1 git -c filter.lfs.required=false archive "$commit" | tar -x -C "$tree"
    report="$EVIDENCE/GITLEAKS_TREE_${TOTAL}_${commit}.json"
    log="$EVIDENCE/GITLEAKS_TREE_${TOTAL}_${commit}.log"
    /home/ubuntu/bin/gitleaks detect --no-banner --redact --no-git --source "$tree" --report-format json --report-path "$report" > "$log" 2>&1
    rc=$?
    echo "commit=$commit tree_index=$TOTAL exit_code=$rc"
    if [ "$rc" -eq 0 ]; then PASS=$((PASS+1)); else FAIL=$((FAIL+1)); fi
  done < "$LIST"
  echo "total=$TOTAL"
  echo "pass=$PASS"
  echo "fail=$FAIL"
} > "$EVIDENCE/GITLEAKS_EXPLICIT_TREE_SCAN.log"
