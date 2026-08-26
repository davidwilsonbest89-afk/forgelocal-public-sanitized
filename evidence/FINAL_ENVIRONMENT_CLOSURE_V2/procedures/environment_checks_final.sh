#!/usr/bin/env bash
set -u
WORK=/home/ubuntu/forgelocal-final-environment-closure
OUT="$WORK/environment-final"
mkdir -p "$OUT"
LOG="$OUT/environment-checks-raw.log"
: > "$LOG"
record() { echo "$*" | tee -a "$LOG"; }
record '=== ENVIRONMENT CHECKS FINAL ==='; date -u --iso-8601=seconds | tee -a "$LOG"
record '=== CHROMIUM VERSION ==='; chromium --version 2>&1 | tee -a "$LOG"
for n in A B; do
  dir="$OUT/profile-$n"; rm -rf "$dir"; mkdir -m 700 "$dir"
  timeout 45s chromium --headless --no-sandbox --disable-gpu --user-data-dir="$dir" --dump-dom 'data:text/html,<title>ForgeLocal synthetic profile</title><p>profile-'"$n"'</p>' > "$OUT/chromium-$n.html" 2> "$OUT/chromium-$n.stderr"
  rc=$?; record "CHROMIUM_PROFILE_${n}_EXIT_CODE=$rc"
  if [[ $rc -eq 0 ]] && grep -q "profile-$n" "$OUT/chromium-$n.html"; then record "CHROMIUM_PROFILE_${n}=PASS"; else record "CHROMIUM_PROFILE_${n}=FAIL"; fi
done
if ! pgrep -af 'chromium.*profile-' >/dev/null; then record 'CHROMIUM_PROFILE_CLEANUP=PASS'; else record 'CHROMIUM_PROFILE_CLEANUP=FAIL'; fi
if command -v docker >/dev/null 2>&1; then docker info >> "$LOG" 2>&1; record "DOCKER_EXIT_CODE=$?"; else record 'DOCKER=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
if command -v firefox >/dev/null 2>&1; then firefox --version | tee -a "$LOG"; record 'FIREFOX=AVAILABLE_NOT_EXECUTED'; else record 'FIREFOX=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
if command -v camoufox >/dev/null 2>&1; then record 'CAMOUFOX=AVAILABLE_NOT_EXECUTED'; else record 'CAMOUFOX=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
if command -v secret-tool >/dev/null 2>&1 && { [[ -n "${DBUS_SESSION_BUS_ADDRESS:-}" ]] || command -v dbus-run-session >/dev/null 2>&1; }; then record 'SYSTEMVAULT=NATIVE_CHECK_REQUIRES_MANUAL_SECRET_SERVICE'; else record 'SYSTEMVAULT=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
if [[ -n "${DISPLAY:-}" || -n "${WAYLAND_DISPLAY:-}" ]]; then record "GUI_SESSION=AVAILABLE DISPLAY=${DISPLAY:-} WAYLAND=${WAYLAND_DISPLAY:-}"; else record 'GUI_SESSION=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
if command -v xdg-open >/dev/null 2>&1 && { [[ -n "${DISPLAY:-}" ]] || [[ -n "${WAYLAND_DISPLAY:-}" ]]; }; then record 'XDG_OPEN=AVAILABLE_NOT_EXECUTED'; else record 'XDG_OPEN=NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE'; fi
record '=== PROCESS RESIDUAL CHECK ==='; pgrep -af 'chromium.*profile-|forgelocal.*server|vite' | tee -a "$LOG" || true
