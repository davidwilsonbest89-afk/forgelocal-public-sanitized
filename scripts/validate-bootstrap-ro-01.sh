#!/usr/bin/env bash
# BOOTSTRAP-RO-01: harness local assaini. Ne jamais activer set -x.
set -euo pipefail

BIN="${FORGELOCAL_BINARY:?FORGELOCAL_BINARY est requis}"
BASE_DIR="${FORGELOCAL_BASE_DIR:?FORGELOCAL_BASE_DIR est requis}"
BASE_URL="${FORGELOCAL_BASE_URL:?FORGELOCAL_BASE_URL est requis}"
EVIDENCE_FILE="${BOOTSTRAP_RO_EVIDENCE_FILE:?BOOTSTRAP_RO_EVIDENCE_FILE est requis}"
VERIFY_EXPIRY="${VERIFY_EXPIRY:-0}"
NONLOOPBACK_BASE_URL="${FORGELOCAL_NONLOOPBACK_BASE_URL:-}"
NONLOOPBACK_ISSUANCE_URL="${FORGELOCAL_NONLOOPBACK_ISSUANCE_URL:-}"
NONLOOPBACK_BINARY="${FORGELOCAL_NONLOOPBACK_BINARY:-}"
NONLOOPBACK_BASE_DIR="${FORGELOCAL_NONLOOPBACK_BASE_DIR:-}"
NONLOOPBACK_EVIDENCE_FILE="${BOOTSTRAP_RO_NONLOOPBACK_EVIDENCE_FILE:-}"

tmp_dir="$(mktemp -d)"
trap 'rm -rf "$tmp_dir"' EXIT

fail() {
  printf 'BOOTSTRAP_RO_HARNESS_FAIL: %s\n' "$1" >&2
  exit 1
}

require_status() {
  local expected="$1"
  local actual="$2"
  local label="$3"
  [[ "$actual" == "$expected" ]] || fail "$label: statut $actual, attendu $expected"
}

json_string() {
	local field="$1"
	sed -n "s/.*\"${field}\"[[:space:]]*:[[:space:]]*\"\([^\"]*\)\".*/\1/p"
}

json_number() {
	local field="$1"
	sed -n "s/.*\"${field}\"[[:space:]]*:[[:space:]]*\([0-9][0-9]*\).*/\1/p"
}

issue_code() {
  "$BIN" --base-dir "$BASE_DIR" readonly-session code --base-url "$BASE_URL" --json
}

issue_nonloopback_code() {
  "$NONLOOPBACK_BINARY" --base-dir "$NONLOOPBACK_BASE_DIR" readonly-session code --base-url "$NONLOOPBACK_ISSUANCE_URL" --json
}

exchange_code() {
	local code="$1"
	exchange_code_at "$BASE_URL" "$code"
}

exchange_code_at() {
	local url="$1"
	local code="$2"
	curl --silent --show-error --noproxy '*' \
		--request POST "$url/api/v1/readonly/session/bootstrap" \
    --header 'Content-Type: application/json' \
    --header 'X-Request-ID: bootstrap-ro-harness' \
    --data "{\"code\":\"${code}\"}" \
    --write-out $'\n%{http_code}'
}

read_endpoint() {
  local token="$1"
  local endpoint="$2"
  curl --silent --show-error --noproxy '*' \
    --header "Authorization: Bearer ${token}" \
    --header 'X-Request-ID: bootstrap-ro-harness-read' \
    --write-out $'\n%{http_code}' \
    "$BASE_URL${endpoint}"
}

write_attempt() {
  local token="$1"
  curl --silent --show-error --noproxy '*' \
    --request POST "$BASE_URL/api/profiles" \
    --header "Authorization: Bearer ${token}" \
    --header 'Content-Type: application/json' \
    --header 'X-Request-ID: bootstrap-ro-harness-write-denied' \
    --data '{"name":"bootstrap-ro-write-must-not-execute"}' \
    --write-out $'\n%{http_code}'
}

split_response() {
  local response="$1"
  response_status="${response##*$'\n'}"
  response_body="${response%$'\n'*}"
}

forbidden_redacted_fields='"(proxy|proxy_secret_ref|password|username|fingerprint|profile_path|user_data_dir|cookie|cookies)"[[:space:]]*:'

mkdir -p "$(dirname "$EVIDENCE_FILE")"
: > "$EVIDENCE_FILE"
printf 'decision=BOOTSTRAP_RO_HARNESS_RUNNING\n' >> "$EVIDENCE_FILE"
printf 'base_url=%s\n' "$BASE_URL" >> "$EVIDENCE_FILE"
printf 'verification_expiry_reelle=%s\n' "$VERIFY_EXPIRY" >> "$EVIDENCE_FILE"

issuance="$(issue_code)" || fail 'émission de code impossible'
code="$(printf '%s' "$issuance" | json_string code)"
code_expires="$(printf '%s' "$issuance" | json_string expires_at)"
[[ "$code" =~ ^[a-f0-9]{64}$ ]] || fail 'code de bootstrap non conforme'
[[ -n "$code_expires" ]] || fail 'expiration de code absente'
code_ttl="$(( $(date -u -d "$code_expires" +%s) - $(date -u +%s) ))"
(( code_ttl >= 590 && code_ttl <= 600 )) || fail 'TTL de code hors contrat'
printf 'control_01_code_forme_64_hex=PASS\ncontrol_01_code_ttl_seconds=%s\n' "$code_ttl" >> "$EVIDENCE_FILE"

exchange="$(exchange_code "$code")" || fail 'échange loopback impossible'
split_response "$exchange"
require_status 200 "$response_status" 'échange loopback'
token="$(printf '%s' "$response_body" | json_string token)"
token_expires="$(printf '%s' "$response_body" | json_string expires_at)"
[[ "$token" =~ ^[a-f0-9]{64}$ ]] || fail 'token court non conforme'
[[ -n "$token_expires" ]] || fail 'expiration de token absente'
token_ttl="$(( $(date -u -d "$token_expires" +%s) - $(date -u +%s) ))"
(( token_ttl >= 890 && token_ttl <= 900 )) || fail 'TTL de token hors contrat'
printf 'control_02_exchange_loopback=PASS\ncontrol_02_token_ttl_seconds=%s\n' "$token_ttl" >> "$EVIDENCE_FILE"

summary_before="$(read_endpoint "$token" '/api/v1/readonly/summary')" || fail 'lecture summary avant écriture impossible'
split_response "$summary_before"
require_status 200 "$response_status" 'summary avant écriture'
profiles_before="$(printf '%s' "$response_body" | json_number profiles)"
[[ "$profiles_before" =~ ^[0-9]+$ ]] || fail 'compteur profiles absent'

for endpoint in '/api/v1/readonly/health' '/api/v1/readonly/summary' '/api/v1/readonly/profiles?limit=50'; do
  read_result="$(read_endpoint "$token" "$endpoint")" || fail "lecture ${endpoint} impossible"
  split_response "$read_result"
  require_status 200 "$response_status" "lecture ${endpoint}"
  if printf '%s' "$response_body" | grep -Eq "$forbidden_redacted_fields"; then
    fail "champ sensible présent dans ${endpoint}"
  fi
done
printf 'control_09_lectures_redacted_health_summary_profiles=PASS\n' >> "$EVIDENCE_FILE"

write_result="$(write_attempt "$token")" || fail "tentative d'écriture impossible"
split_response "$write_result"
require_status 401 "$response_status" "tentative d'écriture avec token court"
summary_after="$(read_endpoint "$token" '/api/v1/readonly/summary')" || fail 'lecture summary après écriture impossible'
split_response "$summary_after"
require_status 200 "$response_status" 'summary après écriture'
profiles_after="$(printf '%s' "$response_body" | json_number profiles)"
[[ "$profiles_before" == "$profiles_after" ]] || fail 'le nombre de profils a changé après une écriture refusée'
printf 'control_10_aucune_ecriture=PASS\ncontrol_10_profiles_avant_apres=%s/%s\n' "$profiles_before" "$profiles_after" >> "$EVIDENCE_FILE"

replay_issuance="$(issue_code)" || fail 'émission replay impossible'
replay_code="$(printf '%s' "$replay_issuance" | json_string code)"
[[ "$replay_code" =~ ^[a-f0-9]{64}$ ]] || fail 'code replay non conforme'
replay_first="$(exchange_code "$replay_code")" || fail 'premier échange replay impossible'
split_response "$replay_first"
require_status 200 "$response_status" 'premier échange replay'
replay_second="$(exchange_code "$replay_code")" || fail 'rejeu du code impossible'
split_response "$replay_second"
require_status 401 "$response_status" 'rejeu du code'
printf '%s' "$response_body" | grep -q 'INVALID_BOOTSTRAP_CODE' || fail 'code erreur rejet replay absent'
printf 'control_03_rejeu_code_refuse=PASS\n' >> "$EVIDENCE_FILE"

if [[ -n "$NONLOOPBACK_BASE_URL" ]]; then
		[[ -x "$NONLOOPBACK_BINARY" && -n "$NONLOOPBACK_BASE_DIR" && -n "$NONLOOPBACK_ISSUANCE_URL" ]] || fail 'configuration contrôle hors loopback incomplète'
		nonloop_issuance="$(issue_nonloopback_code)" || fail 'émission contrôle hors loopback impossible'
		nonloop_code="$(printf '%s' "$nonloop_issuance" | json_string code)"
		[[ "$nonloop_code" =~ ^[a-f0-9]{64}$ ]] || fail 'code hors loopback non conforme'
		nonloop_exchange="$(exchange_code_at "$NONLOOPBACK_BASE_URL" "$nonloop_code")" || fail 'contrôle hors loopback impossible'
		split_response "$nonloop_exchange"
		require_status 403 "$response_status" 'échange hors loopback'
		printf '%s' "$response_body" | grep -q 'LOOPBACK_REQUIRED' || fail 'code erreur hors loopback absent'
		printf 'control_05_hors_loopback_refuse=PASS\n' >> "$EVIDENCE_FILE"
		if [[ -n "$NONLOOPBACK_EVIDENCE_FILE" ]]; then
			mkdir -p "$(dirname "$NONLOOPBACK_EVIDENCE_FILE")"
			printf 'control_05_hors_loopback_refuse=PASS\nhttp_status=403\nerror_code=LOOPBACK_REQUIRED\n' > "$NONLOOPBACK_EVIDENCE_FILE"
		fi
else
		printf 'control_05_hors_loopback_refuse=NOT_RUN\n' >> "$EVIDENCE_FILE"
		[[ -z "$NONLOOPBACK_EVIDENCE_FILE" ]] || printf 'control_05_hors_loopback_refuse=NOT_RUN\n' > "$NONLOOPBACK_EVIDENCE_FILE"
fi

if [[ "$VERIFY_EXPIRY" == '1' ]]; then
  expiry_issuance="$(issue_code)" || fail 'émission expiration impossible'
  expiry_code="$(printf '%s' "$expiry_issuance" | json_string code)"
  [[ "$expiry_code" =~ ^[a-f0-9]{64}$ ]] || fail 'code expiration non conforme'
  sleep 605
  expired_exchange="$(exchange_code "$expiry_code")" || fail 'échange après expiration impossible'
  split_response "$expired_exchange"
  require_status 401 "$response_status" 'code expiré'
  printf '%s' "$response_body" | grep -q 'INVALID_BOOTSTRAP_CODE' || fail 'code erreur expiration absent'
  printf 'control_04_code_expire_refuse=PASS\n' >> "$EVIDENCE_FILE"
else
  printf 'control_04_code_expire_refuse=NOT_RUN\n' >> "$EVIDENCE_FILE"
fi

printf 'decision=BOOTSTRAP_RO_HARNESS_PASS\n' >> "$EVIDENCE_FILE"
printf 'BOOTSTRAP_RO_HARNESS_PASS evidence=%s\n' "$EVIDENCE_FILE"
