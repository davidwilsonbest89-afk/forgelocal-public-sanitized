#!/usr/bin/env bash
set +e
repo=/home/ubuntu/forgelocal-prehuman-fresh-20260824
out=/home/ubuntu/forgelocal-final-check-20260824
log=$out/PREHUMAN_FINAL_LICENSE_INVENTORY_RAW.log
json=$out/PREHUMAN_FINAL_LICENSE_INVENTORY.json
: > "$log"
printf 'started_utc=%s\ncwd=%s\nhead=%s\ncommand=pnpm dlx license-checker --production --json\n--- raw_output ---\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$repo/forge-dashboard" "$(git -C "$repo" rev-parse HEAD 2>&1)" >> "$log"
(cd "$repo/forge-dashboard" && pnpm dlx license-checker --production --json > "$json") >> "$log" 2>&1
rc=$?
printf 'exit_code=%s\ncompleted_utc=%s\n' "$rc" "$(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$log"
