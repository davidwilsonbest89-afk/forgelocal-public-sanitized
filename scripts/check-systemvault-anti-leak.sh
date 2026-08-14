#!/usr/bin/env bash
# Scan ForgeLocal data for a non-production sentinel stored in a 0600 file.
# The sentinel never appears in command arguments, stdout, or the saved report.
set -euo pipefail

DATA_DIR="${FORGELOCAL_DATA_DIR:?set FORGELOCAL_DATA_DIR to the isolated test data directory}"
SENTINEL_FILE="${FORGELOCAL_TEST_SENTINEL_FILE:?set FORGELOCAL_TEST_SENTINEL_FILE to a mode-0600 file outside FORGELOCAL_DATA_DIR}"
OUT_FILE="${OUT_FILE:-$DATA_DIR/systemvault-anti-leak.json}"

[[ -d "$DATA_DIR" ]] || { echo "data directory does not exist" >&2; exit 2; }
[[ -f "$SENTINEL_FILE" ]] || { echo "sentinel file does not exist" >&2; exit 3; }
[[ "$(stat -c '%a' "$SENTINEL_FILE")" == "600" ]] || { echo "sentinel file must have mode 0600" >&2; exit 4; }
[[ "$(readlink -f "$SENTINEL_FILE")" != "$(readlink -f "$DATA_DIR")"/* ]] || {
  echo "sentinel file must be outside the scanned data directory" >&2
  exit 5
}

paths=()
for candidate in "$DATA_DIR/metadata.db" "$DATA_DIR/profiles" "$DATA_DIR/backups" "$DATA_DIR/logs"; do
  [[ -e "$candidate" ]] && paths+=("$candidate")
done
(( ${#paths[@]} > 0 )) || { echo "no expected ForgeLocal data paths found" >&2; exit 6; }

if grep -R --binary-files=text --fixed-strings --file="$SENTINEL_FILE" "${paths[@]}" >/dev/null 2>&1; then
  printf '{"anti_leak":false,"checked_paths":%d,"secret_material_emitted":false}\n' "${#paths[@]}" > "$OUT_FILE"
  echo "secret sentinel detected in ForgeLocal data; see local incident process without echoing the value" >&2
  exit 7
fi
printf '{"anti_leak":true,"checked_paths":%d,"secret_material_emitted":false}\n' "${#paths[@]}" > "$OUT_FILE"
printf '%s\n' "anti-leak scan passed; sanitized result: $OUT_FILE"
