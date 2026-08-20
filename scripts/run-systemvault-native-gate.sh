#!/usr/bin/env bash
# Execute the native Linux SystemVault smoke matrix from a real user session.
# It never accepts, prints, or stores secret material in this script.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_DIR="${OUT_DIR:-$ROOT/systemvault-native-evidence}"
SERVICE="${FORGELOCAL_VAULT_SERVICE:-ForgeLocal.Back01.Release.$(date -u +%Y%m%dT%H%M%SZ)}"

if [[ "${EUID}" -eq 0 ]]; then
  echo "refusing SystemVault release gate as root; run in the target desktop user session without sudo" >&2
  exit 2
fi
if [[ -f /.dockerenv ]] || grep -Eq '/(docker|containerd|kubepods|lxc)/' /proc/1/cgroup 2>/dev/null; then
  echo "refusing SystemVault release gate in a container; run on the target OS host" >&2
  exit 3
fi
if [[ -z "${DBUS_SESSION_BUS_ADDRESS:-}" ]] || [[ -z "${XDG_RUNTIME_DIR:-}" ]]; then
  echo "missing user D-Bus session; unlock the native desktop keyring and run from that user session" >&2
  exit 4
fi

mkdir -p "$OUT_DIR"
chmod 0700 "$OUT_DIR"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.25.13}"
export FORGELOCAL_VAULT_SERVICE="$SERVICE"

{
  echo "os_id=$( . /etc/os-release && printf '%s' "${ID:-unknown}" )"
  echo "os_version_id=$( . /etc/os-release && printf '%s' "${VERSION_ID:-unknown}" )"
  echo "kernel=$(uname -r)"
  echo "architecture=$(uname -m)"
  echo "backend_expected=Secret Service"
  echo "session_user_class=non-root-desktop-session"
  echo "service=$SERVICE"
  echo "executed_without_sudo=true"
  echo "executed_in_container=false"
} > "$OUT_DIR/systemvault-host-context.env"

(
  cd "$ROOT"
  go run ./cmd/systemvault-doctor > "$OUT_DIR/systemvault-matrix.json"
)

python3 - "$OUT_DIR/systemvault-matrix.json" <<'PY'
import json, sys
path = sys.argv[1]
record = json.load(open(path, encoding="utf-8"))
required = ("created_key", "read_key", "restart_read", "created_secret", "read_secret", "deleted", "absent_verified")
missing = [name for name in required if record.get(name) is not True]
if missing:
    raise SystemExit("SystemVault matrix is incomplete: " + ", ".join(missing))
PY

printf 'systemvault_native_matrix=passed\n' > "$OUT_DIR/SYSTEMVAULT_NATIVE_GATE_STATUS"
printf 'manual_cases_remaining=external_revocation,vault_locked_or_permission_denied,integrated_backup_anti_leak\n' >> "$OUT_DIR/SYSTEMVAULT_NATIVE_GATE_STATUS"
printf '%s\n' "SystemVault smoke matrix completed: $OUT_DIR/systemvault-matrix.json"
