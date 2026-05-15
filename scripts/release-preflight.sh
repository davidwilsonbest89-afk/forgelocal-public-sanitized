#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: scripts/release-preflight.sh vX.Y.Z

Environment:
  CAMOUFOX_PATH=/path/to/camoufox   Override Camoufox binary path.
  SKIP_CAMOUFOX=1                   Skip the Camoufox runtime spike.
  SKIP_DOCKER=1                     Skip the linux/amd64 Docker build.
USAGE
}

die() {
  echo "error: $*" >&2
  exit 1
}

run() {
  echo "+ $*"
  "$@"
}

version="${1:-}"
if [[ -z "$version" || "${version}" == "-h" || "${version}" == "--help" ]]; then
  usage
  exit 1
fi

[[ "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]] || die "version must match vX.Y.Z, got: $version"

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

[[ "$(git status --short)" == "" ]] || die "working tree must be clean before release"

current_branch="$(git branch --show-current)"
[[ "$current_branch" == "main" ]] || die "release must be prepared from main, current branch: $current_branch"

if git rev-parse "$version" >/dev/null 2>&1; then
  die "local tag already exists: $version"
fi

head_sha="$(git rev-parse HEAD)"
origin_main_sha="$(git rev-parse origin/main 2>/dev/null || true)"
if [[ -n "$origin_main_sha" && "$head_sha" != "$origin_main_sha" ]]; then
  die "HEAD ($head_sha) does not match origin/main ($origin_main_sha); push main first"
fi

command -v go >/dev/null || die "go is required"
command -v node >/dev/null || die "node is required"
command -v docker >/dev/null || die "docker is required"
command -v rg >/dev/null || die "ripgrep (rg) is required"

if ! rg -q "BROWSEFORGE_VERSION:-${version}" docker/docker-compose.yml; then
  die "docker/docker-compose.yml default BROWSEFORGE_VERSION is not ${version}"
fi

if ! rg -q "ghcr.io/nczz/browseforge:${version}" docker/README.md docs/linux-server.md; then
  die "Docker docs must reference ghcr.io/nczz/browseforge:${version}"
fi

run go test -count=1 ./...
run go vet ./...
run node --check extension/sidebar/app.js
run node -e "JSON.parse(require('fs').readFileSync('extension/manifest.json','utf8')); JSON.parse(require('fs').readFileSync('extension/_locales/en/messages.json','utf8')); JSON.parse(require('fs').readFileSync('extension/_locales/zh_TW/messages.json','utf8')); console.log('extension json ok')"
run node -e "const fs=require('fs'); const html=fs.readFileSync('internal/api/dashboard.html','utf8'); const m=html.match(/<script>([\\s\\S]*)<\\/script>/); if(!m) throw new Error('script not found'); new Function(m[1]); console.log('dashboard js ok')"

if command -v ruby >/dev/null; then
  run ruby -e "require 'yaml'; YAML.load_file('.github/workflows/release.yml'); puts 'workflow yaml ok'"
else
  echo "warning: ruby not found; skipping workflow YAML parse"
fi

run docker compose -f docker/docker-compose.yml config

if [[ "${SKIP_CAMOUFOX:-}" != "1" ]]; then
  camoufox_path="${CAMOUFOX_PATH:-$repo_root/browsers/camoufox/Camoufox.app/Contents/MacOS/camoufox}"
  [[ -x "$camoufox_path" ]] || die "Camoufox binary not found or not executable: $camoufox_path"
  run env CAMOUFOX_PATH="$camoufox_path" go test -count=1 -run '^TestPlaywrightBindEndpointWithCamoufox$' -v ./internal/spike
else
  echo "warning: skipping Camoufox runtime spike because SKIP_CAMOUFOX=1"
fi

if [[ "${SKIP_DOCKER:-}" != "1" ]]; then
  run docker build \
    --platform linux/amd64 \
    -f docker/Dockerfile.run \
    --build-arg "BROWSEFORGE_VERSION=${version}" \
    --build-arg BROWSEFORGE_ARCH=linux-x64 \
    -t "browseforge:verify-${version}" \
    docker
  run docker run --rm --platform linux/amd64 --entrypoint /bin/bash "browseforge:verify-${version}" -n /entrypoint.sh
else
  echo "warning: skipping Docker build because SKIP_DOCKER=1"
fi

echo "release preflight passed for ${version}"
