#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

version="${1:-}"
if [[ -z "$version" ]]; then
  version="$(grep 'BrowseForgeChromiumVersion' internal/browser/download.go | sed -E 's/.*"([^"]+)".*/\1/' | head -n 1)"
fi

if [[ -z "$version" ]]; then
  echo "Could not read BrowseForgeChromiumVersion from internal/browser/download.go" >&2
  exit 1
fi

base_url="${BROWSEFORGE_CHROMIUM_RELEASE_BASE_URL:-https://github.com/nczz/browseforge-runtime-chromium/releases/download}"
base_url="${base_url%/}"

missing=0
checksums_url="${base_url}/${version}/checksums.txt"
echo "Checking ${checksums_url}"
checksums=""
if ! checksums="$(curl -fsSL "$checksums_url" 2>/dev/null)"; then
  echo "BrowseForge Chromium runtime metadata unavailable: ${checksums_url}" >&2
  missing=1
fi

manifest_url="${base_url}/${version}/runtime.manifest.json"
echo "Checking ${manifest_url}"
if ! curl -fsI "$manifest_url" >/dev/null; then
  echo "BrowseForge Chromium runtime metadata unavailable: ${manifest_url}" >&2
  missing=1
fi

for suffix in linux-x64 linux-arm64 macos-arm64 macos-x64 windows-x64; do
  filename="browseforge-runtime-chromium-${version}-${suffix}.zip"
  url="${base_url}/${version}/${filename}"
  echo "Checking ${url}"
  if ! curl -fsI "$url" >/dev/null; then
    echo "BrowseForge Chromium runtime artifact unavailable: ${url}" >&2
    missing=1
  fi
  if [[ -n "$checksums" && "$checksums" != *"$filename"* ]]; then
    echo "BrowseForge Chromium runtime checksum entry missing: ${filename}" >&2
    missing=1
  fi
done

exit "$missing"
