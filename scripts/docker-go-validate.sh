#!/usr/bin/env bash
set -euo pipefail

# Run the repository's Go validation in a clean Docker container.
# This validates normal Go packages plus full build. internal/spike contains
# browser/Playwright driver experiments that require host browser assets, so the
# Docker path compiles that package without executing its external-browser tests.
# Real MCP browser smoke tests still require a running BrowseForge server with a
# Chromium/CloakBrowser profile and are covered by scripts/mcp-web-smoke.sh.

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_IMAGE="${GO_IMAGE:-golang:1.25}"

mkdir -p "${ROOT_DIR}/.cache/go-build" "${ROOT_DIR}/.cache/go-mod"

docker run --rm \
  -v "${ROOT_DIR}:/workspace" \
  -v "${ROOT_DIR}/.cache/go-build:/root/.cache/go-build" \
  -v "${ROOT_DIR}/.cache/go-mod:/go/pkg/mod" \
  -w /workspace \
  "${GO_IMAGE}" \
  sh -lc 'export PATH="/usr/local/go/bin:${PATH}"; go version && pkgs="$(go list ./... | grep -v "/internal/spike$")" && go test ${pkgs} && go test -run "^$" ./internal/spike && go build ./...'
