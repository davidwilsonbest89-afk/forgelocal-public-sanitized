#!/usr/bin/env bash
set -euo pipefail

# Run with SENTINEL_VALUE supplied only by the private test process.
: "${SENTINEL_VALUE:?SENTINEL_VALUE must be supplied privately}"
ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
TMP=$(mktemp -d)
cleanup() {
  pkill -P $$ 2>/dev/null || true
  rm -rf "$TMP"
}
trap cleanup EXIT
mkdir -p "$TMP/bin" "$TMP/app/data" "$TMP/app/profiles" "$TMP/app/browsers" "$TMP/app/logs" "$TMP/app/backups"
cat > "$TMP/bin/Xvnc" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
cat > "$TMP/bin/openbox" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
cat > "$TMP/bin/vncpasswd" <<'STUB'
#!/usr/bin/env bash
cat >/dev/null
exit 0
STUB
cat > "$TMP/app/BrowseForge" <<'STUB'
#!/usr/bin/env bash
exit 0
STUB
chmod +x "$TMP/bin/Xvnc" "$TMP/bin/openbox" "$TMP/bin/vncpasswd" "$TMP/app/BrowseForge"
printf '%s' "$SENTINEL_VALUE" > "$TMP/app/data/.api-token"
chmod 600 "$TMP/app/data/.api-token"
sed -e "s|/app|$TMP/app|g" -e "s|/usr/bin/Xvnc|$TMP/bin/Xvnc|g" "$ROOT/docker/entrypoint.sh" > "$TMP/entrypoint-present.sh"
chmod +x "$TMP/entrypoint-present.sh"
env PATH="$TMP/bin:$PATH" BROWSEFORGE_SEED_BROWSERS=0 VNC_PASSWORD=private-test-password timeout 15s bash "$TMP/entrypoint-present.sh" >"$TMP/present.out" 2>"$TMP/present.err"
if grep -Fq "$SENTINEL_VALUE" "$TMP/present.out" || grep -Fq "$SENTINEL_VALUE" "$TMP/present.err"; then
  exit 41
fi
[[ "$(stat -c '%a' "$TMP/app/data/.api-token")" == 600 ]]
cmp -s <(printf '%s' "$SENTINEL_VALUE") "$TMP/app/data/.api-token"
grep -Fq '[redacted' "$TMP/present.out"
sed -e 's/seq 1 60/seq 1 2/g' -e 's/sleep 2/sleep 0.01/g' "$TMP/entrypoint-present.sh" > "$TMP/entrypoint-absent.sh"
rm -f "$TMP/app/data/.api-token"
env PATH="$TMP/bin:$PATH" BROWSEFORGE_SEED_BROWSERS=0 VNC_PASSWORD=private-test-password timeout 15s bash "$TMP/entrypoint-absent.sh" >"$TMP/absent.out" 2>"$TMP/absent.err"
if grep -Fq "$SENTINEL_VALUE" "$TMP/absent.out" || grep -Fq "$SENTINEL_VALUE" "$TMP/absent.err"; then
  exit 42
fi
! grep -Fq 'API Token:' "$TMP/absent.out"
! grep -Fq 'API Token:' "$TMP/absent.err"
bash -n "$ROOT/docker/entrypoint.sh" "$ROOT/scripts/start.sh"
if grep -Fq 'API Token: $TOKEN' "$ROOT/docker/entrypoint.sh" "$ROOT/scripts/start.sh"; then
  exit 43
fi
if pgrep -f "$TMP/(entrypoint-present|entrypoint-absent|BrowseForge|Xvnc|openbox)" >/dev/null 2>&1; then
  exit 44
fi
printf '%s\n' 'ENTRYPOINT_SECRET_REDACTION_TEST=PASS'
printf '%s\n' 'TOKEN_VALUE_NOT_PRINTED=TRUE'
printf '%s\n' 'TOKEN_FILE_PERMISSION_TEST=PASS'
printf '%s\n' 'TOKEN_ABSENT_CASE=PASS'
printf '%s\n' 'PROCESS_CLEANUP=PASS'
