#!/usr/bin/env bash
# Verify a ForgeLocal release-manifest signature using an externally supplied public key only.
set -euo pipefail

MANIFEST_PATH="${1:?usage: verify-release-manifest.sh <release-manifest.json> <signature.asc> <maintainer-public-key.asc>}"
SIGNATURE_PATH="${2:?usage: verify-release-manifest.sh <release-manifest.json> <signature.asc> <maintainer-public-key.asc>}"
PUBLIC_KEY_PATH="${3:?usage: verify-release-manifest.sh <release-manifest.json> <signature.asc> <maintainer-public-key.asc>}"
EXPECTED_FINGERPRINT="${FORGELOCAL_RELEASE_SIGNING_FINGERPRINT:?set FORGELOCAL_RELEASE_SIGNING_FINGERPRINT to the approved public-key fingerprint}"

command -v gpg >/dev/null || { echo "gpg is required" >&2; exit 2; }
for path in "$MANIFEST_PATH" "$SIGNATURE_PATH" "$PUBLIC_KEY_PATH"; do
  [[ -f "$path" ]] || { echo "required file does not exist: $path" >&2; exit 3; }
done

normalized_expected="$(printf '%s' "$EXPECTED_FINGERPRINT" | tr -d '[:space:]' | tr '[:lower:]' '[:upper:]')"
[[ "$normalized_expected" =~ ^[0-9A-F]{40}$ ]] || { echo "release signing fingerprint must be exactly 40 hexadecimal characters" >&2; exit 4; }

verification_home="$(mktemp -d)"
trap 'rm -rf "$verification_home"' EXIT
chmod 0700 "$verification_home"
export GNUPGHOME="$verification_home"

gpg --batch --import "$PUBLIC_KEY_PATH" >/dev/null
public_fingerprints="$(gpg --batch --with-colons --list-keys "$normalized_expected" 2>/dev/null | awk -F: '$1 == "fpr" { print toupper($10) }')"
printf '%s\n' "$public_fingerprints" | grep -Fx "$normalized_expected" >/dev/null || {
  echo "approved maintainer public key is unavailable after import" >&2
  exit 5
}

if gpg --batch --with-colons --list-secret-keys | awk -F: '$1 == "sec" { found = 1 } END { exit(found ? 0 : 1) }'; then
  echo "verification keyring unexpectedly contains a private key" >&2
  exit 6
fi

status_file="$(mktemp)"
trap 'rm -rf "$verification_home" "$status_file"' EXIT
gpg --batch --status-fd 1 --verify "$SIGNATURE_PATH" "$MANIFEST_PATH" >"$status_file" 2>&1 || {
  cat "$status_file" >&2
  exit 7
}

if ! awk -v expected="$normalized_expected" '$1 == "[GNUPG:]" && $2 == "VALIDSIG" { primary = toupper($NF); signing = toupper($3); if (primary == expected || signing == expected) valid = 1 } END { exit(valid ? 0 : 1) }' "$status_file"; then
  cat "$status_file" >&2
  echo "signature is not bound to the approved maintainer fingerprint" >&2
  exit 8
fi

printf 'verified_manifest=%s\nsignature=%s\nverification_fingerprint=%s\nprivate_keys_present=false\n' \
  "$MANIFEST_PATH" "$SIGNATURE_PATH" "$normalized_expected"
