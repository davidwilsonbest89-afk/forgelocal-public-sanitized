#!/usr/bin/env bash
# Capture reproducible release evidence for an externally supplied Chromium runtime.
# This script does not install or execute a package and does not require sudo.
set -euo pipefail

PACKAGE_NAME="${PACKAGE_NAME:-chromium}"
PACKAGE_VERSION="${PACKAGE_VERSION:?set PACKAGE_VERSION to the exact candidate version}"
PACKAGE_ARCH="${PACKAGE_ARCH:-amd64}"
REPOSITORY_URL="${REPOSITORY_URL:-https://ppa.launchpadcontent.net/xtradeb/apps/ubuntu}"
REPOSITORY_SUITE="${REPOSITORY_SUITE:-noble}"
REPOSITORY_COMPONENT="${REPOSITORY_COMPONENT:-main}"
KEYRING_PATH="${KEYRING_PATH:-/etc/apt/trusted.gpg.d/xtradeb-apps.gpg}"
EXPECTED_KEY_FINGERPRINT="${EXPECTED_KEY_FINGERPRINT:-5301FA4FD93244FBC6F6149982BB6851C64F6880}"
OUT_DIR="${OUT_DIR:-$PWD/runtime-release-evidence/${PACKAGE_NAME}_${PACKAGE_VERSION}_${PACKAGE_ARCH}}"

for command in apt-get apt-cache curl dpkg-deb gpg gpgv gzip sha256sum; do
  command -v "$command" >/dev/null || { echo "missing required command: $command" >&2; exit 2; }
done
[[ -r "$KEYRING_PATH" ]] || { echo "keyring is not readable: $KEYRING_PATH" >&2; exit 3; }

actual_fingerprints="$(gpg --show-keys --with-colons "$KEYRING_PATH" 2>/dev/null | awk -F: '$1 == "fpr" { print $10 }')"
printf '%s\n' "$actual_fingerprints" | grep -Fxq "$EXPECTED_KEY_FINGERPRINT" || {
  echo "expected signing key fingerprint is absent from $KEYRING_PATH" >&2
  exit 4
}

mkdir -p "$OUT_DIR"
: > "$OUT_DIR/RELEASE_CAPTURE_STATUS"
printf 'capture_status=started\n' >> "$OUT_DIR/RELEASE_CAPTURE_STATUS"
printf 'package=%s\nversion=%s\narchitecture=%s\nrepository=%s\nsuite=%s\ncomponent=%s\nexpected_key_fingerprint=%s\n' \
  "$PACKAGE_NAME" "$PACKAGE_VERSION" "$PACKAGE_ARCH" "$REPOSITORY_URL" "$REPOSITORY_SUITE" "$REPOSITORY_COMPONENT" "$EXPECTED_KEY_FINGERPRINT" \
  > "$OUT_DIR/REQUESTED_RUNTIME.env"

# Capture the local APT view before downloading. This is evidence, not a substitute
# for the signed repository metadata copied below.
apt-cache policy "$PACKAGE_NAME" > "$OUT_DIR/APT_POLICY.txt" || true
apt-cache show "${PACKAGE_NAME}=${PACKAGE_VERSION}" > "$OUT_DIR/APT_EXACT_VERSION.txt" 2>&1 || true

inrelease_url="${REPOSITORY_URL}/dists/${REPOSITORY_SUITE}/InRelease"
curl --fail --location --proto '=https' --tlsv1.2 --max-time 60 --silent --show-error \
  "$inrelease_url" -o "$OUT_DIR/InRelease"
gpgv --keyring "$KEYRING_PATH" "$OUT_DIR/InRelease" > "$OUT_DIR/INRELEASE_SIGNATURE.txt" 2>&1
printf '%s  %s\n' "$EXPECTED_KEY_FINGERPRINT" "signing-key-fingerprint" > "$OUT_DIR/SIGNING_KEY_FINGERPRINT.txt"
cp "$KEYRING_PATH" "$OUT_DIR/SIGNING_KEYRING.gpg"
sha256sum "$OUT_DIR/InRelease" "$OUT_DIR/SIGNING_KEYRING.gpg" > "$OUT_DIR/INDEX_AND_KEYRING_SHA256SUMS"

# Verify the exact binary index against the signed InRelease before using the
# package checksum published in that index. Both the compressed index and its
# decompressed form are retained as evidence.
packages_relative="${REPOSITORY_COMPONENT}/binary-${PACKAGE_ARCH}/Packages.gz"
packages_url="${REPOSITORY_URL}/dists/${REPOSITORY_SUITE}/${packages_relative}"
curl --fail --location --proto '=https' --tlsv1.2 --max-time 60 --silent --show-error \
  "$packages_url" -o "$OUT_DIR/Packages.gz"
expected_packages_sha256="$(awk -v target="$packages_relative" '
  $1 == "SHA256:" { section=1; next }
  section && NF == 0 { exit }
  section && $3 == target { print $1; exit }
' "$OUT_DIR/InRelease")"
[[ -n "$expected_packages_sha256" ]] || { echo "Packages.gz hash missing from signed InRelease" >&2; exit 8; }
actual_packages_sha256="$(sha256sum "$OUT_DIR/Packages.gz" | awk '{print $1}')"
[[ "$actual_packages_sha256" == "$expected_packages_sha256" ]] || { echo "Packages.gz hash does not match signed InRelease" >&2; exit 9; }
printf '%s  Packages.gz\n' "$actual_packages_sha256" > "$OUT_DIR/PACKAGES_INDEX_SHA256SUMS"
gzip -dc "$OUT_DIR/Packages.gz" > "$OUT_DIR/Packages"

# `apt-get download` is intentionally unprivileged: it only retrieves the exact
# binary package. A missing historical version is a hard failure, not permission
# to substitute the current candidate.
(
  cd "$OUT_DIR"
  apt-get download "${PACKAGE_NAME}=${PACKAGE_VERSION}"
)

deb_file="$(find "$OUT_DIR" -maxdepth 1 -type f -name "${PACKAGE_NAME}_*_${PACKAGE_ARCH}.deb" -print -quit)"
[[ -n "$deb_file" ]] || { echo "exact .deb not retrieved: ${PACKAGE_NAME}=${PACKAGE_VERSION}" >&2; exit 5; }
actual_version="$(dpkg-deb -f "$deb_file" Version)"
actual_architecture="$(dpkg-deb -f "$deb_file" Architecture)"
[[ "$actual_version" == "$PACKAGE_VERSION" ]] || { echo "retrieved a different version: $actual_version" >&2; exit 6; }
[[ "$actual_architecture" == "$PACKAGE_ARCH" ]] || { echo "retrieved a different architecture: $actual_architecture" >&2; exit 7; }

dpkg-deb -I "$deb_file" > "$OUT_DIR/DEB_CONTROL.txt"
sha256sum "$deb_file" > "$OUT_DIR/DEB_SHA256SUMS"
expected_deb_sha256="$(awk -v package="$PACKAGE_NAME" -v version="$PACKAGE_VERSION" -v arch="$PACKAGE_ARCH" '
  BEGIN { RS=""; FS="\\n" }
  {
    found_package = found_version = found_arch = 0; hash = ""
    for (i = 1; i <= NF; i++) {
      line = $i
      if (line == "Package: " package) found_package = 1
      if (line == "Version: " version) found_version = 1
      if (line == "Architecture: " arch) found_arch = 1
      if (line ~ /^SHA256: /) { sub(/^SHA256: /, "", line); hash = line }
    }
    if (found_package && found_version && found_arch && hash != "") { print hash; exit }
  }
' "$OUT_DIR/Packages")"
[[ -n "$expected_deb_sha256" ]] || { echo "deb checksum missing from signed Packages index" >&2; exit 10; }
actual_deb_sha256="$(sha256sum "$deb_file" | awk '{print $1}')"
[[ "$actual_deb_sha256" == "$expected_deb_sha256" ]] || { echo "deb checksum does not match signed Packages index" >&2; exit 11; }
printf 'package_index_checksum_match=true\nsha256=%s\n' "$actual_deb_sha256" > "$OUT_DIR/DEB_INDEX_SHA256_VERIFICATION.txt"
sha256sum "$OUT_DIR"/InRelease "$OUT_DIR"/Packages.gz "$OUT_DIR"/Packages "$OUT_DIR"/SIGNING_KEYRING.gpg "$deb_file" > "$OUT_DIR/RELEASE_EVIDENCE_SHA256SUMS"
printf 'capture_status=complete\n' >> "$OUT_DIR/RELEASE_CAPTURE_STATUS"
printf '%s\n' "captured release evidence in $OUT_DIR"
