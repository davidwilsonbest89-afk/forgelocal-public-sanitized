#!/usr/bin/env bash
set -u
BASE=/home/ubuntu/forgelocal_self_validation_v5
REPO=$BASE/repo
E=$BASE/evidence
cd "$REPO"
export PATH=/usr/local/go1.25.13/bin:/home/ubuntu/bin:$PATH
export GOTOOLCHAIN=local
export CGO_ENABLED=1
{
  echo "utc=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "cwd=$PWD"
  echo "head=$(git rev-parse HEAD)"
  echo "go=$(/usr/local/go1.25.13/bin/go version)"
  echo "golangci=$(/home/ubuntu/bin/golangci-lint-2.13.1 version 2>&1)"
  echo '--- golangci-lint'
  /home/ubuntu/bin/golangci-lint-2.13.1 run --timeout 10m --output.json.path="$E/GOLANGCI_FINAL.json" > "$E/GOLANGCI_FINAL.stdout" 2> "$E/GOLANGCI_FINAL.stderr"
  echo "golangci_exit_code=$?"
  echo '--- go test shuffle count=3'
  /usr/local/go1.25.13/bin/go test -shuffle=on -count=3 ./... > "$E/GO_TEST_SHUFFLE.log" 2>&1
  echo "go_test_shuffle_exit_code=$?"
  echo '--- go test shuffle race count=3'
  /usr/local/go1.25.13/bin/go test -shuffle=on -count=3 -race ./... > "$E/GO_TEST_SHUFFLE_RACE.log" 2>&1
  echo "go_test_shuffle_race_exit_code=$?"
} > "$E/MANDATORY_V5_RAW.log" 2>&1
