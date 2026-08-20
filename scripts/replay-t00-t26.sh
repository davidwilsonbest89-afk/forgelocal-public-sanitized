#!/usr/bin/env bash
set -euo pipefail

# Local-only ForgeLocal continuation replay. No runtime, proxy, provider or release is started.
ROOT="${1:?artifact root required}"
OUT="${2:-$PWD/forgelocal-t00-t26-replay}"
WORK="$OUT/core-t26"
VERIFY="$OUT/verify"
LOG="$OUT/BASELINE_DISCOVERY_RAW.log"
WRAPPER="$ROOT/forgelocal-continuation-t00-t23-20260819.zip"
T26_BUNDLE="$ROOT/forgelocal-core-t26-simulated-provider-930003c.bundle"

mkdir -p "$OUT" "$VERIFY"
exec > >(tee "$LOG") 2>&1
run() { printf 'started_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"; printf 'cwd=%s\ncommand=' "$PWD"; printf '%q ' "$@"; printf '\n'; "$@"; printf 'exit_code=0\ncompleted_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"; }

run bash -c "cd '$ROOT' && sha256sum -c forgelocal-continuation-t00-t23-20260819.zip.sha256"
run unzip -t "$WRAPPER"
run bash -c "cd '$ROOT' && sha256sum -c forgelocal-core-t26-simulated-provider-930003c.bundle.sha256 && sha256sum -c forgelocal-t26-simulated-provider-930003c.zip.sha256"
run git -C "$VERIFY" init -q
run git -C "$VERIFY" bundle verify "$T26_BUNDLE"
test ! -e "$WORK" || { echo "worktree already exists: $WORK"; exit 2; }
run git clone "$T26_BUNDLE" "$WORK"
run git -C "$WORK" checkout --detach t26-simulated-proxy-provider-2026-08-20
run git -C "$WORK" rev-parse HEAD
run git -C "$WORK" fsck --full
run env PATH=/usr/local/go/bin:$PATH go -C "$WORK" test -count=1 -race ./...
run env PATH=/usr/local/go/bin:$PATH go -C "$WORK" vet ./...
run env PATH=/usr/local/go/bin:$PATH go -C "$WORK" build ./...
echo 'decision=BASELINE_T26_QUALIFIED_FOR_NEXT_AUTHORIZED_LOT_ONLY'
