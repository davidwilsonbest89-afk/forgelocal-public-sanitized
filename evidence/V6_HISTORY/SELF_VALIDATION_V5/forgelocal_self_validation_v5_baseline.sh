#!/usr/bin/env bash
set -u
BASE_DIR=/home/ubuntu/forgelocal_self_validation_v5
REPO=$BASE_DIR/repo
EVIDENCE=$BASE_DIR/evidence
REMOTE=https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized.git
BRANCH=audit/t00-t42-self-validation-synthetic-e2e
BASE_TAG=t00-t27-complete-20260820
mkdir -p "$BASE_DIR"
rm -rf "$REPO" "$EVIDENCE"
mkdir -p "$EVIDENCE"
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$(pwd)"
  echo "df=$(df -h /home/ubuntu | tail -1)"
  echo "git_version=$(git --version)"
  echo "gitleaks_version=$(/home/ubuntu/bin/gitleaks version 2>&1 || true)"
  echo "go_version=$(/usr/local/go1.25.13/bin/go version 2>&1 || true)"
  echo "semgrep_version=$(semgrep --version 2>&1 || true)"
  echo "grype_version=$(grype version 2>&1 || true)"
} > "$EVIDENCE/01_baseline_environment.log"
GIT_LFS_SKIP_SMUDGE=1 gh repo clone davidwilsonbest89-afk/forgelocal-public-sanitized "$REPO" -- --branch "$BRANCH" > "$EVIDENCE/02_clone.log" 2>&1
cd "$REPO"
git fetch --tags origin "$BASE_TAG" > "$EVIDENCE/03_fetch_baseline.log" 2>&1
HEAD=$(git rev-parse HEAD)
BASE=$(git rev-parse "$BASE_TAG")
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "branch=$BRANCH"
  echo "base_tag=$BASE_TAG"
  echo "base=$BASE"
  echo "head=$HEAD"
  printf 'ancestor_rc='; git merge-base --is-ancestor "$BASE_TAG" "$HEAD"; echo "$?"
  printf 'commit_count='; git rev-list --count "$BASE_TAG..$HEAD"
} > "$EVIDENCE/04_refs_and_ancestor.log"
git rev-list --format='%H' "$BASE_TAG..$HEAD" > "$EVIDENCE/05_commit_list.log"
git log --format='%H' "$BASE_TAG..$HEAD" > "$EVIDENCE/06_commit_hashes.log"
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "head=$HEAD"
  echo "base=$BASE"
  echo '--- gitleaks historical range'
  set +e
  /home/ubuntu/bin/gitleaks detect --no-banner --redact --source . --log-opts="$BASE_TAG..$HEAD" --report-format json --report-path "$EVIDENCE/GITLEAKS_IMMUTABLE_RANGE.json" > "$EVIDENCE/GITLEAKS_IMMUTABLE_RANGE.log" 2>&1
  RC1=$?
  echo "gitleaks_historical_exit_code=$RC1"
  echo '--- gitleaks fresh checkout'
  /home/ubuntu/bin/gitleaks detect --no-banner --redact --no-git --source . --report-format json --report-path "$EVIDENCE/GITLEAKS_FRESH_CHECKOUT.json" > "$EVIDENCE/GITLEAKS_FRESH_CHECKOUT.log" 2>&1
  RC2=$?
  echo "gitleaks_fresh_checkout_exit_code=$RC2"
  set -e
  echo '--- gitleaks historical tail'
  tail -20 "$EVIDENCE/GITLEAKS_IMMUTABLE_RANGE.log"
  echo '--- gitleaks fresh tail'
  tail -20 "$EVIDENCE/GITLEAKS_FRESH_CHECKOUT.log"
} > "$EVIDENCE/07_gitleaks_replay.log"
