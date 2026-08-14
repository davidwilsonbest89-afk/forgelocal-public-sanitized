#!/usr/bin/env bash
# Sign a ForgeLocal release manifest with a maintainer-controlled OpenPGP key.
# The build never generates, imports or stores a private signing key.
set -euo pipefail

MANIFEST_PATH="${1:?usage: sign-release-manifest.sh <release-manifest.json>}"
EXPECTED_FINGERPRINT="${FORGELOCAL_RELEASE_SIGNING_FINGERPRINT:?set FORGELOCAL_RELEASE_SIGNING_FINGERPRINT to the approved public-key fingerprint}"
SIGNATURE_PATH="${MANIFEST_PATH}.asc"

command -v gpg >/dev/null || { echo "gpg is required" >&2; exit 2; }
[[ -f "$MANIFEST_PATH" ]] || { echo "manifest does not exist: $MANIFEST_PATH" >&2; exit 3; }

normalized_expected="$(printf '%s' "$EXPECTED_FINGERPRINT" | tr -d '[:space:]' | tr '[:lower:]' '[:upper:]')"
[[ "$normalized_expected" =~ ^[0-9A-F]{40}$ ]] || { echo "release signing fingerprint must be exactly 40 hexadecimal characters" >&2; exit 4; }
actual_fingerprints="$(gpg --batch --with-colons --list-secret-keys "$normalized_expected" 2>/dev/null | awk -F: '$1 == "fpr" { print toupper($10) }')"
printf '%s\n' "$actual_fingerprints" | grep -Fx "$normalized_expected" >/dev/null || {
  echo "approved release signing private key is unavailable in this user keyring" >&2
  exit 5
}

rm -f "$SIGNATURE_PATH"
gpg --batch --yes --armor --local-user "$normalized_expected" --detach-sign --output "$SIGNATURE_PATH" "$MANIFEST_PATH"
gpg --batch --verify "$SIGNATURE_PATH" "$MANIFEST_PATH"
printf 'signed_manifest=%s\nsignature=%s\nsigning_fingerprint=%s\n' "$MANIFEST_PATH" "$SIGNATURE_PATH" "$normalized_expected"
