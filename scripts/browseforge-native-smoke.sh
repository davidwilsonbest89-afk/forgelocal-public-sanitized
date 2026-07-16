#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BROWSEFORGE_BASE_URL:-http://127.0.0.1:19280}"
TOKEN="${BROWSEFORGE_API_TOKEN:-}"
BASE_DIR="${BROWSEFORGE_BASE_DIR:-}"
RUNTIME_ID="${BROWSEFORGE_RUNTIME_ID:-browseforge-chromium}"
PROFILE_NAME="${BROWSEFORGE_SMOKE_PROFILE_NAME:-native smoke}"

if [ -z "$TOKEN" ]; then
  echo "BROWSEFORGE_API_TOKEN is required" >&2
  exit 2
fi

auth_header="Authorization: Bearer ${TOKEN}"
json_header="Content-Type: application/json"
profile_payload=$(python3 - <<'PY'
import json, os
print(json.dumps({
    "name": os.environ.get("BROWSEFORGE_SMOKE_PROFILE_NAME", "native smoke"),
    "runtime_id": os.environ.get("BROWSEFORGE_RUNTIME_ID", "browseforge-chromium"),
    "group": "smoke",
    "tags": ["native", "smoke"],
}))
PY
)

profile_response=$(curl -fsSL -H "$auth_header" -H "$json_header" -d "$profile_payload" "$BASE_URL/api/profiles")
profile_id=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["id"])' <<<"$profile_response")
profile_dir=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["data"].get("profile_dir", ""))' <<<"$profile_response")

session_payload=$(python3 - <<PY
import json
print(json.dumps({"profile_id": "$profile_id"}))
PY
)
session_response=$(curl -fsSL -H "$auth_header" -H "$json_header" -d "$session_payload" "$BASE_URL/api/sessions")
session_id=$(python3 -c 'import json,sys; print(json.load(sys.stdin)["data"]["session_id"])' <<<"$session_response")

read -r -d '' smoke_js <<'JS' || true
(() => {
  const canvas = document.createElement('canvas');
  const gl = canvas.getContext('webgl') || canvas.getContext('experimental-webgl');
  const dbg = gl && gl.getExtension('WEBGL_debug_renderer_info');
  return {
    url: location.href,
    userAgent: navigator.userAgent,
    webdriver: navigator.webdriver,
    language: navigator.language,
    languages: navigator.languages,
    intlDateLocale: Intl.DateTimeFormat().resolvedOptions().locale,
    intlNumberLocale: Intl.NumberFormat().resolvedOptions().locale,
    intlCollatorLocale: Intl.Collator().resolvedOptions().locale,
    timezone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    screen: {
      width: screen.width,
      height: screen.height,
      availWidth: screen.availWidth,
      availHeight: screen.availHeight,
      colorDepth: screen.colorDepth,
      devicePixelRatio: devicePixelRatio,
      innerWidth: innerWidth,
      innerHeight: innerHeight,
      outerWidth: outerWidth,
      outerHeight: outerHeight,
      visualViewportWidth: visualViewport && visualViewport.width,
      visualViewportHeight: visualViewport && visualViewport.height,
    },
    webgl: gl ? {
      vendor: dbg ? gl.getParameter(dbg.UNMASKED_VENDOR_WEBGL) : null,
      renderer: dbg ? gl.getParameter(dbg.UNMASKED_RENDERER_WEBGL) : null,
      version: gl.getParameter(gl.VERSION),
      shadingLanguageVersion: gl.getParameter(gl.SHADING_LANGUAGE_VERSION),
    } : null,
  };
})()
JS

eval_payload=$(python3 - <<PY
import json
print(json.dumps({"script": """$smoke_js"""}))
PY
)
eval_response=$(curl -fsSL -H "$auth_header" -H "$json_header" -d "$eval_payload" "$BASE_URL/api/sessions/$session_id/eval")

native_config=""
if [ -n "$BASE_DIR" ] && [ -n "$profile_dir" ]; then
  candidate="$BASE_DIR/$profile_dir/browser-data/BrowseForgeNative/persona.json"
  if [ -f "$candidate" ]; then
    native_config=$(python3 -c 'import json,sys; print(json.dumps(json.load(open(sys.argv[1])), sort_keys=True))' "$candidate")
  fi
fi

python3 - <<PY
import json
payload = {
    "ok": True,
    "runtime_id": "$RUNTIME_ID",
    "profile_id": "$profile_id",
    "session_id": "$session_id",
    "profile_dir": "$profile_dir",
    "observed": json.loads('''$eval_response''').get("data"),
}
if '''${native_config:+1}''':
    payload["native_config"] = json.loads('''${native_config:-{}}''')
print(json.dumps(payload, indent=2, sort_keys=True))
PY
