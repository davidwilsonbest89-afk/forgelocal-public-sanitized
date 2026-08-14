#!/usr/bin/env python3
"""Validate ForgeLocal public-release gates, provenance and freshness metadata."""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import re
import sys
from pathlib import Path


SHA256_RE = re.compile(r"^[0-9a-f]{64}$")
COMMIT_RE = re.compile(r"^[0-9a-f]{40}$")
REQUIRED_GATE_IDS = {
    "SYSTEMVAULT_NATIVE_PER_TARGET",
    "SYSTEMVAULT_ANTI_LEAK_INTEGRATED_FLOW",
    "MAINTAINER_MANIFEST_SIGNATURE",
    "RUNTIME_LICENSE_AND_REDISTRIBUTION_REVIEW",
    "OS_COMPATIBILITY_EVIDENCE",
}


class GateError(Exception):
    pass


def load(path: Path) -> dict:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise GateError(f"invalid_json:{path}:{error}") from error
    if not isinstance(value, dict):
        raise GateError(f"json_object_required:{path}")
    return value


def require_date(value: object, name: str) -> None:
    if not isinstance(value, str):
        raise GateError(f"date_required:{name}")
    try:
        dt.date.fromisoformat(value)
    except ValueError as error:
        raise GateError(f"invalid_date:{name}") from error


def require_sha256(value: object, name: str) -> str:
    if not isinstance(value, str) or not SHA256_RE.fullmatch(value):
        raise GateError(f"sha256_required:{name}")
    return value


def relative_file(root: Path, value: object, name: str) -> Path:
    if not isinstance(value, str) or not value:
        raise GateError(f"file_required:{name}")
    path = (root / value).resolve()
    try:
        path.relative_to(root)
    except ValueError as error:
        raise GateError(f"path_escapes_repository:{name}") from error
    if not path.is_file():
        raise GateError(f"evidence_file_missing:{name}")
    return path


def sha256_file(path: Path) -> str:
    digest = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            digest.update(chunk)
    return digest.hexdigest()


def validate_gate(root: Path, gate: object, candidate: dict, allowed: set[str]) -> tuple[str, str]:
    if not isinstance(gate, dict):
        raise GateError("invalid_gate_entry")
    gate_id = gate.get("id")
    if not isinstance(gate_id, str) or gate_id not in REQUIRED_GATE_IDS:
        raise GateError("unknown_or_missing_gate_id")
    decision = gate.get("decision")
    if decision not in allowed:
        raise GateError(f"invalid_gate_decision:{gate_id}")
    owner = gate.get("owner")
    if not isinstance(owner, dict) or not isinstance(owner.get("role"), str) or not owner["role"].strip():
        raise GateError(f"owner_role_required:{gate_id}")
    if owner.get("assignment_status") not in {"ASSIGNED", "UNASSIGNED"}:
        raise GateError(f"owner_assignment_status_required:{gate_id}")
    require_date(gate.get("recorded_at"), f"{gate_id}.recorded_at")
    tested_commit = gate.get("tested_commit")
    if not isinstance(tested_commit, str) or not COMMIT_RE.fullmatch(tested_commit):
        raise GateError(f"tested_commit_required:{gate_id}")
    if tested_commit != candidate["artifact"]["source_commit"]:
        raise GateError(f"stale_tested_commit:{gate_id}")
    if gate.get("runtime_version") != candidate["runtime"]["version"]:
        raise GateError(f"stale_runtime_version:{gate_id}")
    freshness = gate.get("freshness")
    if not isinstance(freshness, dict):
        raise GateError(f"freshness_required:{gate_id}")
    if freshness.get("valid_for_artifact_sha256") != candidate["artifact"]["sha256"]:
        raise GateError(f"stale_artifact_hash:{gate_id}")
    if not isinstance(freshness.get("invalidated_by"), list) or not freshness["invalidated_by"]:
        raise GateError(f"invalidation_rules_required:{gate_id}")
    target = gate.get("target")
    if not isinstance(target, dict) or not {"os", "os_version", "architecture"}.issubset(target):
        raise GateError(f"target_metadata_required:{gate_id}")
    evidence = gate.get("evidence")
    if not isinstance(evidence, dict) or evidence.get("status") not in {"MISSING", "PRESENT"}:
        raise GateError(f"evidence_metadata_required:{gate_id}")
    review = gate.get("independent_review")
    if not isinstance(review, dict) or review.get("decision") not in allowed:
        raise GateError(f"independent_review_metadata_required:{gate_id}")

    if decision == "PASSED":
        if owner.get("assignment_status") != "ASSIGNED":
            raise GateError(f"passed_gate_owner_must_be_assigned:{gate_id}")
        require_date(gate.get("executed_at"), f"{gate_id}.executed_at")
        if not all(isinstance(target[key], str) and target[key].strip() for key in ("os", "os_version", "architecture")):
            raise GateError(f"passed_gate_target_required:{gate_id}")
        if evidence.get("status") != "PRESENT":
            raise GateError(f"passed_gate_evidence_missing:{gate_id}")
        evidence_path = relative_file(root, evidence.get("file"), f"{gate_id}.evidence")
        expected_hash = require_sha256(evidence.get("sha256"), f"{gate_id}.evidence")
        if sha256_file(evidence_path) != expected_hash:
            raise GateError(f"evidence_hash_mismatch:{gate_id}")
        if review.get("decision") != "PASSED":
            raise GateError(f"independent_review_must_pass:{gate_id}")
        reviewer_role = review.get("reviewer_role")
        if not isinstance(reviewer_role, str) or not reviewer_role.strip() or reviewer_role == owner["role"]:
            raise GateError(f"independent_reviewer_required:{gate_id}")
        require_date(review.get("reviewed_at"), f"{gate_id}.reviewed_at")
        if not isinstance(review.get("review_ref"), str) or not review["review_ref"].strip():
            raise GateError(f"independent_review_reference_required:{gate_id}")
    elif decision == "NOT_APPLICABLE":
        if not isinstance(gate.get("not_applicable_justification"), str) or not gate["not_applicable_justification"].strip():
            raise GateError(f"not_applicable_justification_required:{gate_id}")
        if review.get("decision") != "PASSED":
            raise GateError(f"not_applicable_review_must_pass:{gate_id}")
    else:
        if evidence.get("status") == "PRESENT":
            evidence_path = relative_file(root, evidence.get("file"), f"{gate_id}.evidence")
            expected_hash = require_sha256(evidence.get("sha256"), f"{gate_id}.evidence")
            if sha256_file(evidence_path) != expected_hash:
                raise GateError(f"evidence_hash_mismatch:{gate_id}")
        elif evidence.get("file") is not None or evidence.get("sha256") is not None:
            raise GateError(f"missing_evidence_must_not_contain_reference:{gate_id}")

    return gate_id, decision


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    root = args.root.resolve()
    state_path = root / "release/back01-minimal/PUBLIC_RELEASE_GATE_STATE.json"
    index_path = root / "release/back01-minimal/RELEASE_TRACEABILITY_INDEX.json"

    try:
        state = load(state_path)
        index = load(index_path)
        if state.get("format") != "forgelocal-public-release-gate-state/v2":
            raise GateError("unsupported_gate_state_format")
        decision = state.get("decision")
        if decision not in {"PUBLIC_RELEASE_BLOCKED", "PUBLIC_RELEASE_APPROVED"}:
            raise GateError("invalid_public_release_decision")
        require_date(state.get("recorded_at"), "state.recorded_at")
        artifact_chain = state.get("artifact_chain")
        if not isinstance(artifact_chain, dict) or not isinstance(artifact_chain.get("id"), str):
            raise GateError("artifact_chain_metadata_required")
        candidate = next((item for item in index.get("chains", []) if item.get("chain_id") == artifact_chain["id"]), None)
        if not isinstance(candidate, dict):
            raise GateError("candidate_chain_missing")
        if artifact_chain.get("artifact_sha256") != candidate.get("artifact", {}).get("sha256"):
            raise GateError("artifact_chain_hash_mismatch")
        if artifact_chain.get("source_commit") != candidate.get("artifact", {}).get("source_commit"):
            raise GateError("artifact_chain_source_commit_mismatch")
        if artifact_chain.get("runtime", {}).get("version") != candidate.get("runtime", {}).get("version"):
            raise GateError("artifact_chain_runtime_mismatch")

        schema = state.get("gate_schema")
        if not isinstance(schema, dict) or not isinstance(schema.get("allowed_decisions"), list):
            raise GateError("gate_schema_required")
        allowed = set(schema["allowed_decisions"])
        if allowed != {"PASSED", "FAILED", "PENDING", "NOT_APPLICABLE"}:
            raise GateError("gate_decision_enumeration_invalid")

        gates = state.get("gates")
        if not isinstance(gates, list) or len(gates) != len(REQUIRED_GATE_IDS):
            raise GateError("exactly_five_public_gates_required")
        validated = [validate_gate(root, gate, candidate, allowed) for gate in gates]
        gate_ids = {gate_id for gate_id, _ in validated}
        if gate_ids != REQUIRED_GATE_IDS:
            raise GateError("required_gate_set_mismatch")
        gate_decisions = dict(validated)
        all_passed = all(value == "PASSED" for value in gate_decisions.values())

        required_docs = (
            "release/back01-minimal/PUBLIC_RELEASE_DECISION.md",
            "release/back01-minimal/RELEASE_SCOPE_AND_OS_MATRIX.md",
            "release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md",
            "release/back01-minimal/RELEASE_SIGNATURE_RUNBOOK.md",
        )
        for relative in required_docs:
            if not (root / relative).is_file():
                raise GateError(f"required_gate_document_missing:{relative}")

        if decision == "PUBLIC_RELEASE_BLOCKED":
            if index.get("policy", {}).get("public_release_decision") != "PUBLIC_RELEASE_BLOCKED":
                raise GateError("traceability_policy_must_be_blocked")
            if candidate.get("public_release_eligible") is not False or "PUBLIC_RELEASE_BLOCKED" not in str(candidate.get("status", "")):
                raise GateError("candidate_status_must_remain_publicly_blocked")
            if all_passed:
                raise GateError("all_gates_passed_requires_explicit_promotion_review")
        else:
            if not all_passed:
                raise GateError("public_approval_requires_all_gates_passed")
            if index.get("policy", {}).get("public_release_decision") != "PUBLIC_RELEASE_APPROVED":
                raise GateError("traceability_policy_not_promoted")
            if candidate.get("public_release_eligible") is not True:
                raise GateError("candidate_not_marked_publicly_eligible")
    except GateError as error:
        print(json.dumps({"valid": False, "error": str(error)}, sort_keys=True))
        return 1

    pending = [gate_id for gate_id, gate_decision in gate_decisions.items() if gate_decision != "PASSED"]
    print(json.dumps({"decision": decision, "gate_decisions": gate_decisions, "pending_gates": pending, "valid": True}, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
