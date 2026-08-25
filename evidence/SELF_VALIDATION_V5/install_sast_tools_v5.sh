#!/usr/bin/env bash
set -u
BASE=/home/ubuntu/forgelocal_self_validation_v5
TOOLS=$BASE/tools
E=$BASE/evidence
mkdir -p "$TOOLS" "$E"
python3 -m venv "$TOOLS/semgrep-venv"
"$TOOLS/semgrep-venv/bin/python" -m pip install --upgrade pip >/dev/null 2>&1
"$TOOLS/semgrep-venv/bin/pip" install semgrep >/tmp/semgrep-install-v5.log 2>&1
SEMVER=$($TOOLS/semgrep-venv/bin/semgrep --version 2>&1)
printf '%s\n' "semgrep=$SEMVER" > "$E/semgrep.version.log"
GRYPE_TAG=$(gh api repos/anchore/grype/releases/latest --jq '.tag_name')
GRYPE_VERSION=${GRYPE_TAG#v}
URL="https://github.com/anchore/grype/releases/download/${GRYPE_TAG}/grype_${GRYPE_VERSION}_linux_amd64.tar.gz"
curl -fL "$URL" -o "$TOOLS/grype_${GRYPE_VERSION}_linux_amd64.tar.gz" > "$E/grype.download.log" 2>&1
sha256sum "$TOOLS/grype_${GRYPE_VERSION}_linux_amd64.tar.gz" > "$E/grype_${GRYPE_VERSION}.tar.gz.sha256"
tar -xzf "$TOOLS/grype_${GRYPE_VERSION}_linux_amd64.tar.gz" -C "$TOOLS"
install -m 0755 "$TOOLS/grype" /home/ubuntu/bin/grype-${GRYPE_VERSION}
GRYPEVER=$(/home/ubuntu/bin/grype-${GRYPE_VERSION} version 2>&1)
printf '%s\n' "tag=$GRYPE_TAG" "binary=/home/ubuntu/bin/grype-${GRYPE_VERSION}" "$GRYPEVER" > "$E/grype.version.log"
