#!/bin/bash
# Démarrage du Core ForgeLocal pour la suite E2E dashboard T16.
# Loopback seulement, port 19280, token via BROWSEFORGE_TOKEN, base-dir isolé.
set -euo pipefail
umask 077
export PATH=/home/ubuntu/.local/go1.25.13/bin:$PATH
export BROWSEFORGE_TOKEN="$(head -c 32 /dev/urandom | od -v -An -tx1 | tr -d ' \n')"
BASE="/tmp/forge-e2e-base"
rm -rf "$BASE"
mkdir -p "$BASE"
cd /home/ubuntu/forgebaseline-reimpl
go build -o /tmp/forge-core-e2e ./cmd/server
printf '%s\n' "$BROWSEFORGE_TOKEN" > /tmp/forge-e2e-token.txt
exec /tmp/forge-core-e2e serve --host 127.0.0.1 --port 19280 --no-open --no-sandbox
