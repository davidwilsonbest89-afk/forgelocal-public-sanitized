#!/usr/bin/env python3
"""Validate ForgeLocal BACK-01 release metadata without executing the binary."""
import hashlib
import json
import pathlib
import sys


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> int:
    if len(sys.argv) != 5:
        print("usage: validate-release-artifact.py <archive> <release-manifest> <internal-manifest> <sbom>", file=sys.stderr)
        return 2
    archive, release_path, internal_manifest_path, sbom_path = map(pathlib.Path, sys.argv[1:])
    release = json.loads(release_path.read_text(encoding="utf-8"))
    internal_manifest = json.loads(internal_manifest_path.read_text(encoding="utf-8"))
    sbom = json.loads(sbom_path.read_text(encoding="utf-8"))

    if release.get("format") != "forgelocal-back01-release-manifest/v1":
        raise ValueError("unexpected release manifest format")
    if internal_manifest.get("format") != "forgelocal-back01-minimal/v1":
        raise ValueError("unexpected internal manifest format")
    if sbom.get("spdxVersion") != "SPDX-2.3":
        raise ValueError("SBOM is not SPDX-2.3")
    if not sbom.get("packages"):
        raise ValueError("SBOM has no packages")
    if release["artifact"]["file"] != archive.name or release["artifact"]["sha256"] != sha256(archive):
        raise ValueError("archive hash does not match release manifest")
    if release["internal_manifest"]["sha256"] != sha256(internal_manifest_path):
        raise ValueError("internal manifest hash does not match release manifest")
    sbom_hash = sha256(sbom_path)
    if release["sbom"]["sha256"] != sbom_hash or internal_manifest["sbom"]["sha256"] != sbom_hash:
        raise ValueError("SBOM hash does not match both manifests")

    print(f"archive_sha256={sha256(archive)}")
    print(f"sbom_format={sbom['spdxVersion']}")
    print(f"sbom_packages={len(sbom['packages'])}")
    print(f"signature_status={release['signature']['status']}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
