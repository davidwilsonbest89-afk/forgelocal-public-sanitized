#!/usr/bin/env python3
"""Validate independent ForgeLocal release-traceability chains without executing artifacts."""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
import tarfile
from pathlib import Path


class ValidationError(Exception):
    """Raised when a traceability invariant is not satisfied."""


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as handle:
        for chunk in iter(lambda: handle.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def load_json(path: Path) -> dict:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise ValidationError(f"invalid_json:{path}:{error}") from error
    if not isinstance(value, dict):
        raise ValidationError(f"json_object_required:{path}")
    return value


def require_file(root: Path, relative: str) -> Path:
    path = root / relative
    if not path.is_file():
        raise ValidationError(f"required_file_missing:{relative}")
    return path


def verify_hash(root: Path, relative: str, expected: str) -> Path:
    path = require_file(root, relative)
    actual = sha256_file(path)
    if actual != expected:
        raise ValidationError(f"sha256_mismatch:{relative}:expected={expected}:actual={actual}")
    return path


def archive_member(bundle: tarfile.TarFile, member_name: str) -> tarfile.TarInfo:
    matches = [name for name in bundle.getnames() if name == member_name or name.endswith(f"/{member_name}")]
    if len(matches) != 1:
        raise ValidationError(f"archive_member_resolution_failed:{member_name}:matches={len(matches)}")
    return bundle.getmember(matches[0])


def extract_json_from_archive(archive: Path, member_name: str) -> dict:
    try:
        with tarfile.open(archive, "r:gz") as bundle:
            member = archive_member(bundle, member_name)
            extracted = bundle.extractfile(member)
            if extracted is None:
                raise ValidationError(f"archive_member_unreadable:{archive.name}:{member_name}")
            payload = extracted.read().decode("utf-8")
    except (OSError, tarfile.TarError, UnicodeDecodeError) as error:
        raise ValidationError(f"archive_member_missing_or_invalid:{archive.name}:{member_name}:{error}") from error
    try:
        value = json.loads(payload)
    except json.JSONDecodeError as error:
        raise ValidationError(f"archive_json_invalid:{archive.name}:{member_name}:{error}") from error
    if not isinstance(value, dict):
        raise ValidationError(f"archive_json_object_required:{archive.name}:{member_name}")
    return value


def validate_e2e(root: Path, chain: dict) -> None:
    e2e = chain.get("e2e")
    if not isinstance(e2e, dict):
        raise ValidationError(f"e2e_object_required:{chain.get('chain_id')}")
    evidence_path = e2e.get("evidence")
    expected_hash = e2e.get("sha256")
    expected_runtime = e2e.get("runtime_version_expected")
    if not all(isinstance(item, str) and item for item in (evidence_path, expected_hash, expected_runtime)):
        raise ValidationError(f"e2e_fields_missing:{chain.get('chain_id')}")
    evidence = verify_hash(root, evidence_path, expected_hash)
    text = evidence.read_text(encoding="utf-8")
    for required in (expected_runtime, "AC-BACK-01 runtime relaunch started", "about:blank", "AC-BACK-01 runtime relaunch stopped cleanly", "profile_lock_cleanup=verified", "PASS"):
        if required not in text:
            raise ValidationError(f"e2e_assertion_missing:{chain.get('chain_id')}:{required}")
    if e2e.get("reusable_for_other_runtime") is not False:
        raise ValidationError(f"e2e_reuse_must_be_false:{chain.get('chain_id')}")


def validate_candidate(root: Path, chain: dict) -> None:
    artifact = chain["artifact"]
    archive = verify_hash(root, artifact["file"], artifact["sha256"])
    sbom = chain.get("sbom")
    manifest = chain.get("external_release_manifest")
    runtime = chain.get("runtime")
    signature = chain.get("signature")
    if not all(isinstance(item, dict) for item in (sbom, manifest, runtime, signature)):
        raise ValidationError("candidate_metadata_objects_required")

    manifest_path = verify_hash(root, manifest["file"], manifest["sha256"])
    external = load_json(manifest_path)
    if external.get("artifact", {}).get("sha256") != artifact["sha256"]:
        raise ValidationError("candidate_external_manifest_artifact_hash_mismatch")
    if external.get("commit") != artifact["source_commit"]:
        raise ValidationError("candidate_external_manifest_source_commit_mismatch")
    if external.get("sbom", {}).get("sha256") != sbom.get("sha256"):
        raise ValidationError("candidate_external_manifest_sbom_hash_mismatch")
    if external.get("signature", {}).get("status") != signature.get("status"):
        raise ValidationError("candidate_signature_status_mismatch")
    if signature.get("status") != "UNSIGNED_REQUIRES_MAINTAINER_KEY":
        raise ValidationError("candidate_signature_gate_unexpected")

    archive_sbom = extract_json_from_archive(archive, sbom["file_inside_artifact"])
    if archive_sbom.get("spdxVersion") != sbom.get("format"):
        raise ValidationError("candidate_sbom_format_mismatch")
    with tarfile.open(archive, "r:gz") as bundle:
        member = archive_member(bundle, sbom["file_inside_artifact"])
        extracted = bundle.extractfile(member)
        if extracted is None:
            raise ValidationError("candidate_sbom_unreadable")
        sbom_digest = hashlib.sha256(extracted.read()).hexdigest()
    if sbom_digest != sbom.get("sha256"):
        raise ValidationError("candidate_sbom_hash_mismatch")

    lock = runtime.get("runtime_lock")
    if not isinstance(lock, dict):
        raise ValidationError("candidate_runtime_lock_required")
    lock_path = verify_hash(root, lock["file"], lock["sha256"])
    lock_json = load_json(lock_path)
    if lock_json.get("installed_runtime", {}).get("binary_sha256") != runtime.get("binary_sha256"):
        raise ValidationError("candidate_runtime_binary_hash_mismatch")
    expected_packages = {(item["name"], item["version"], item["sha256"]) for item in runtime.get("packages", [])}
    observed_packages = {(item.get("name"), item.get("version"), item.get("deb_sha256")) for item in lock_json.get("packages", [])}
    if expected_packages != observed_packages:
        raise ValidationError("candidate_runtime_package_lock_mismatch")


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    parser.add_argument(
        "--index",
        type=Path,
        default=Path("release/back01-minimal/RELEASE_TRACEABILITY_INDEX.json"),
    )
    args = parser.parse_args()
    root = args.root.resolve()
    index_path = args.index if args.index.is_absolute() else root / args.index

    try:
        index = load_json(index_path)
        if index.get("format") != "forgelocal-release-traceability-index/v1":
            raise ValidationError("unsupported_index_format")
        policy = index.get("policy")
        chains = index.get("chains")
        if not isinstance(policy, dict) or not isinstance(chains, list) or len(chains) != 2:
            raise ValidationError("two_chain_policy_required")
        if policy.get("public_release_decision") != "PUBLIC_RELEASE_BLOCKED":
            raise ValidationError("public_release_must_remain_blocked")
        if policy.get("independent_chain_required") is not True or policy.get("evidence_reuse_across_runtime_candidates_forbidden") is not True:
            raise ValidationError("independent_chain_policy_required")

        chain_ids: set[str] = set()
        e2e_paths: set[str] = set()
        runtime_versions: set[str] = set()
        candidate_count = 0
        for chain in chains:
            if not isinstance(chain, dict):
                raise ValidationError("chain_object_required")
            chain_id = chain.get("chain_id")
            artifact = chain.get("artifact")
            runtime = chain.get("runtime")
            if not isinstance(chain_id, str) or chain_id in chain_ids:
                raise ValidationError("chain_id_must_be_unique")
            if not isinstance(artifact, dict) or not isinstance(runtime, dict):
                raise ValidationError(f"chain_fields_missing:{chain_id}")
            chain_ids.add(chain_id)
            verify_hash(root, artifact["file"], artifact["sha256"])
            validate_e2e(root, chain)
            e2e_path = chain["e2e"]["evidence"]
            if e2e_path in e2e_paths:
                raise ValidationError("e2e_evidence_reused_across_chains")
            e2e_paths.add(e2e_path)
            version = chain["e2e"]["runtime_version_expected"]
            if version in runtime_versions:
                raise ValidationError("runtime_version_reused_across_chains")
            runtime_versions.add(version)
            if chain.get("public_release_eligible") is not False:
                raise ValidationError(f"public_release_eligibility_forbidden:{chain_id}")
            if chain.get("kind") == "controlled_release_candidate":
                candidate_count += 1
                validate_candidate(root, chain)
        if candidate_count != 1:
            raise ValidationError("exactly_one_candidate_required")
    except ValidationError as error:
        print(json.dumps({"valid": False, "error": str(error)}, sort_keys=True))
        return 1

    print(json.dumps({"valid": True, "chains": 2, "public_release_decision": "PUBLIC_RELEASE_BLOCKED"}, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
