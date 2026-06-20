#!/usr/bin/env bash
set -euo pipefail

# Gated smoke test for MCP web sessions. It requires a running BrowseForge HTTP
# service with MCP mounted at /mcp and a Chromium/CloakBrowser profile that can
# launch via browser.Manager.
# Usage:
#   RUN_MCP_SMOKE=1 MCP_PROFILE_ID=<profile-id> ./scripts/mcp-web-smoke.sh

if [[ "${RUN_MCP_SMOKE:-}" != "1" ]]; then
  echo "Skipping MCP web smoke test. Set RUN_MCP_SMOKE=1 and MCP_PROFILE_ID=<profile-id> to run."
  exit 0
fi

MCP_URL="${MCP_URL:-http://127.0.0.1:19280/mcp}"
MCP_PROFILE_ID="${MCP_PROFILE_ID:?MCP_PROFILE_ID is required}"
MCP_SEARCH_QUERY="${MCP_SEARCH_QUERY:-BrowseForge MCP}"
MCP_SEARCH_ENGINE="${MCP_SEARCH_ENGINE:-google}"
MCP_EXPLORE_URL="${MCP_EXPLORE_URL:-https://example.com}"
MCP_AUTH_TOKEN="${MCP_AUTH_TOKEN:-}"

SESSION_ID=""
REQ_ID=0

rpc_call() {
  local tool="$1"
  local args_json="$2"
  REQ_ID=$((REQ_ID + 1))
  local payload
  payload="$(python3 - "$REQ_ID" "$tool" "$args_json" <<'PY'
import json, sys
req_id = int(sys.argv[1])
tool = sys.argv[2]
args = json.loads(sys.argv[3])
print(json.dumps({
    "jsonrpc": "2.0",
    "id": req_id,
    "method": "tools/call",
    "params": {"name": tool, "arguments": args},
}))
PY
)"

  local headers=(-H 'Content-Type: application/json')
  if [[ -n "$MCP_AUTH_TOKEN" ]]; then
    headers+=(-H "Authorization: Bearer ${MCP_AUTH_TOKEN}")
  fi

  curl -fsS "${headers[@]}" -d "$payload" "$MCP_URL"
}

require_no_error() {
  python3 - "$1" <<'PY'
import json, sys
resp = json.loads(sys.argv[1])
if 'error' in resp and resp['error']:
    raise SystemExit(f"MCP error: {resp['error']}")
print(json.dumps(resp.get('result', {})))
PY
}

cleanup() {
  if [[ -n "$SESSION_ID" ]]; then
    echo "Destroying MCP web session: $SESSION_ID"
    rpc_call destroy_session "$(python3 - "$SESSION_ID" <<'PY'
import json, sys
print(json.dumps({"session_id": sys.argv[1]}))
PY
)" >/dev/null || true
  fi
}
trap cleanup EXIT

create_resp="$(rpc_call create_session "$(python3 - "$MCP_PROFILE_ID" <<'PY'
import json, sys
print(json.dumps({"profile_id": sys.argv[1]}))
PY
)")"
create_result="$(require_no_error "$create_resp")"
SESSION_ID="$(python3 - "$create_result" <<'PY'
import json, sys
print(json.loads(sys.argv[1]).get('session_id', ''))
PY
)"
if [[ -z "$SESSION_ID" ]]; then
  echo "create_session did not return session_id" >&2
  exit 1
fi

echo "Created MCP web session: $SESSION_ID"

search_args="$(python3 - "$MCP_PROFILE_ID" "$SESSION_ID" "$MCP_SEARCH_QUERY" "$MCP_SEARCH_ENGINE" <<'PY'
import json, sys
print(json.dumps({"profile_id": sys.argv[1], "session_id": sys.argv[2], "query": sys.argv[3], "engine": sys.argv[4], "max_results": 3}))
PY
)"
require_no_error "$(rpc_call web_search "$search_args")" >/dev/null
echo "web_search succeeded ($MCP_SEARCH_ENGINE)"

explore_args="$(python3 - "$MCP_PROFILE_ID" "$SESSION_ID" "$MCP_EXPLORE_URL" <<'PY'
import json, sys
print(json.dumps({"profile_id": sys.argv[1], "session_id": sys.argv[2], "url": sys.argv[3], "max_text_length": 1000, "max_links": 10}))
PY
)"
require_no_error "$(rpc_call web_explore "$explore_args")" >/dev/null
echo "web_explore succeeded"

rpc_call destroy_session "$(python3 - "$SESSION_ID" <<'PY'
import json, sys
print(json.dumps({"session_id": sys.argv[1]}))
PY
)" >/dev/null
SESSION_ID=""
echo "destroy_session succeeded"
