#!/usr/bin/env bash
# Build the explicit ForgeLocal BACK-01 minimal distribution profile.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
export GOTOOLCHAIN="${GOTOOLCHAIN:-go1.25.13}"
export GOFLAGS="${GOFLAGS:--mod=readonly}"

VERSION="${VERSION:-0.1.0-back01-dev}"
OUT_ROOT="${OUT_ROOT:-$ROOT/dist/back01-minimal}"
STAGE="$OUT_ROOT/forgelocal-back01-core-$VERSION"
ARCHIVE="$OUT_ROOT/forgelocal-back01-core-$VERSION-linux-amd64.tar.gz"
SOURCE_DATE_EPOCH="${SOURCE_DATE_EPOCH:-$(git log -1 --format=%ct)}"
FORBIDDEN='^forgelocal/internal/(browser|fingerprint|humanize|mcp|runtime|workflow)(/|$)'
RUNTIME_EVIDENCE="${RUNTIME_EVIDENCE:-$ROOT/validation_back01_integration/final/ac_back_01_with_chromium.out}"
RUNTIME_APPROVAL="${RUNTIME_APPROVAL:-$ROOT/release/back01-minimal/RUNTIME_QA_APPROVAL.md}"
RUNTIME_PROVENANCE="${RUNTIME_PROVENANCE:-$ROOT/validation_back01_integration/final/chromium_runtime_provenance.out}"
case "${AC_BACK_01_RUNTIME_RELAUNCH_PROVEN:-false}" in
  true) runtime_relaunch_proven=true ;;
  false) runtime_relaunch_proven=false ;;
  *) echo "AC_BACK_01_RUNTIME_RELAUNCH_PROVEN must be true or false" >&2; exit 5 ;;
esac

if [[ "$runtime_relaunch_proven" == true ]]; then
  for evidence in "$RUNTIME_EVIDENCE" "$RUNTIME_APPROVAL" "$RUNTIME_PROVENANCE"; do
    if [[ ! -s "$evidence" ]]; then
      echo "refusing build: missing runtime relaunch evidence: $evidence" >&2
      exit 6
    fi
  done
fi

rm -rf "$STAGE"
mkdir -p "$STAGE" "$STAGE/migrations" "$STAGE/licenses"
if [[ "$runtime_relaunch_proven" == true ]]; then
  mkdir -p "$STAGE/evidence"
fi

actual_go="$(go version | awk '{print $3}')"
if [[ "$actual_go" != "go1.25.13" ]]; then
  echo "refusing build: expected go1.25.13, got $actual_go" >&2
  exit 2
fi

GOSEC_BIN="${GOSEC_BIN:-}"
if [[ -z "$GOSEC_BIN" ]]; then
  GOSEC_BIN="$(command -v gosec || true)"
fi
if [[ -z "$GOSEC_BIN" && -x "$HOME/go/bin/gosec" ]]; then
  GOSEC_BIN="$HOME/go/bin/gosec"
fi
if [[ -z "$GOSEC_BIN" || ! -x "$GOSEC_BIN" ]]; then
  echo "refusing build: gosec not found; set GOSEC_BIN or install gosec" >&2
  exit 7
fi

mapfile -t internal_deps < <(go list -deps ./cmd/back01-core | grep '^forgelocal/' | sort -u)
if printf '%s\n' "${internal_deps[@]}" | grep -E "$FORBIDDEN" >/dev/null; then
  echo "refusing build: minimal Core imports forbidden package(s)" >&2
  printf '%s\n' "${internal_deps[@]}" | grep -E "$FORBIDDEN" >&2
  exit 3
fi
expected=$'forgelocal/cmd/back01-core\nforgelocal/internal/backup\nforgelocal/internal/profile\nforgelocal/internal/secrets'
actual="$(printf '%s\n' "${internal_deps[@]}")"
if [[ "$actual" != "$expected" ]]; then
  echo "refusing build: internal dependency closure changed; review PROFILE.md" >&2
  printf '%s\n' "${internal_deps[@]}" >&2
  exit 4
fi

go test -race ./internal/backup ./internal/profile ./internal/secrets ./cmd/back01-core
"$GOSEC_BIN" ./internal/backup/... ./internal/profile/... ./internal/secrets/... ./cmd/back01-core/...

go build -trimpath -buildvcs=true -ldflags='-s -w' -o "$STAGE/forgelocal-back01-core" ./cmd/back01-core
install -m 0644 internal/backup/migrations/0001_back01.sql "$STAGE/migrations/0001_back01.sql"
install -m 0644 release/back01-minimal/PROFILE.md "$STAGE/PROFILE.md"
install -m 0644 LICENSE "$STAGE/LICENSE"
if [[ "$runtime_relaunch_proven" == true ]]; then
  install -m 0644 "$RUNTIME_APPROVAL" "$STAGE/RUNTIME_QA_APPROVAL.md"
  install -m 0644 "$RUNTIME_EVIDENCE" "$STAGE/evidence/AC_BACK_01_RUNTIME_RELAUNCH.out"
  install -m 0644 "$RUNTIME_PROVENANCE" "$STAGE/evidence/RUNTIME_PROVENANCE.out"
fi

# Only modules reached by the compiled minimal command belong in this artifact.
# The raw package graph is build-only metadata: the distributed inventory is
# THIRD_PARTY_MODULES.txt plus the copied licence notices.
dependency_graph="$(mktemp)"
trap 'rm -f "$dependency_graph"' EXIT
go list -deps -json ./cmd/back01-core > "$dependency_graph"
python3 - "$dependency_graph" "$STAGE/THIRD_PARTY_MODULES.txt" "$STAGE/licenses" <<'PY'
import json, os, shutil, sys
source, inventory, license_dir = sys.argv[1:]
raw = open(source, encoding="utf-8").read()
decoder = json.JSONDecoder(); i = 0; modules = {}
while i < len(raw):
    while i < len(raw) and raw[i].isspace(): i += 1
    if i >= len(raw): break
    package, i = decoder.raw_decode(raw, i)
    module = package.get("Module") or {}
    path, version, directory = module.get("Path"), module.get("Version"), module.get("Dir")
    if path and path != "forgelocal":
        modules[(path, version or "")] = directory
missing = []
with open(inventory, "w", encoding="utf-8") as out:
    for (path, version), directory in sorted(modules.items()):
        out.write(f"{path}\t{version}\n")
        candidates = ["LICENSE", "LICENSE.md", "COPYING", "NOTICE"]
        license_path = next((os.path.join(directory, item) for item in candidates if directory and os.path.isfile(os.path.join(directory, item))), None)
        if not license_path:
            missing.append(f"{path}@{version}")
            continue
        safe = (path + "@" + version).replace("/", "_").replace("@", "_")
        shutil.copyfile(license_path, os.path.join(license_dir, safe + ".txt"))
if missing:
    raise SystemExit("missing license notice for: " + ", ".join(missing))
PY

commit="$(git rev-parse HEAD)"
built_at="$(date -u -d "@$SOURCE_DATE_EPOCH" +%Y-%m-%dT%H:%M:%SZ)"
python3 - "$STAGE/MANIFEST.json" "$VERSION" "$commit" "$actual_go" "$built_at" "$runtime_relaunch_proven" <<'PY'
import hashlib, json, os, sys
path, version, commit, go_version, built_at, runtime_relaunch_proven = sys.argv[1:]
runtime_relaunch_proven = runtime_relaunch_proven == "true"
manifest = {
  "format": "forgelocal-back01-minimal/v1",
  "version": version,
  "commit": commit,
  "toolchain": go_version,
  "built_at": built_at,
  "entrypoint": "forgelocal-back01-core",
  "bind_default": "127.0.0.1:45100",
  "api": ["POST /api/v1/profiles/{id}/backups", "POST /api/v1/backups/{id}/restore", "GET /healthz"],
  "internal_dependencies": ["internal/backup", "internal/profile", "internal/secrets"],
  "excluded": ["internal/browser", "internal/fingerprint", "internal/humanize", "internal/mcp", "internal/runtime", "internal/workflow"],
  "runtime_included": False,
  "ac_back_01_runtime_relaunch_proven": runtime_relaunch_proven,
}
if runtime_relaunch_proven:
  evidence_dir = os.path.join(os.path.dirname(path), "evidence")
  evidence = {
    "qa_approval": "RUNTIME_QA_APPROVAL.md",
    "relaunch_test": "evidence/AC_BACK_01_RUNTIME_RELAUNCH.out",
    "runtime_provenance": "evidence/RUNTIME_PROVENANCE.out",
  }
  for key, relative_path in list(evidence.items()):
    if key == "qa_approval":
      evidence[key] = relative_path
      continue
    full_path = os.path.join(os.path.dirname(path), relative_path)
    evidence[key + "_sha256"] = hashlib.sha256(open(full_path, "rb").read()).hexdigest()
  manifest["runtime_relaunch_evidence"] = evidence
open(path, "w", encoding="utf-8").write(json.dumps(manifest, indent=2, sort_keys=True) + "\n")
PY

(
  cd "$STAGE"
  find . -type f ! -name SHA256SUMS -print0 | sort -z | xargs -0 sha256sum > SHA256SUMS
)
rm -f "$ARCHIVE"
tar --sort=name --mtime="@$SOURCE_DATE_EPOCH" --owner=0 --group=0 --numeric-owner -C "$OUT_ROOT" -czf "$ARCHIVE" "$(basename "$STAGE")"
sha256sum "$ARCHIVE" > "$ARCHIVE.sha256"

echo "created $ARCHIVE"
