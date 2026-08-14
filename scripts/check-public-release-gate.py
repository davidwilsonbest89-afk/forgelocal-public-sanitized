#!/usr/bin/env python3
"""Confirm that ForgeLocal public-release gates remain blocked until complete evidence exists."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path


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
        if state.get("format") != "forgelocal-public-release-gate-state/v1":
            raise GateError("unsupported_gate_state_format")
        if state.get("decision") != "PUBLIC_RELEASE_BLOCKED":
            raise GateError("public_release_decision_must_be_blocked")
        if index.get("policy", {}).get("public_release_decision") != "PUBLIC_RELEASE_BLOCKED":
            raise GateError("traceability_policy_must_be_blocked")
        candidate_id = state.get("artifact_chain_id")
        candidate = next((item for item in index.get("chains", []) if item.get("chain_id") == candidate_id), None)
        if not isinstance(candidate, dict):
            raise GateError("candidate_chain_missing")
        if candidate.get("public_release_eligible") is not False:
            raise GateError("candidate_must_not_be_public_release_eligible")
        if "PUBLIC_RELEASE_BLOCKED" not in str(candidate.get("status", "")):
            raise GateError("candidate_status_must_remain_publicly_blocked")

        gates = state.get("gates")
        if not isinstance(gates, list) or len(gates) != 5:
            raise GateError("exactly_five_public_gates_required")
        pending = []
        for gate in gates:
            if not isinstance(gate, dict) or not gate.get("id") or not gate.get("required_evidence"):
                raise GateError("invalid_gate_entry")
            if gate.get("status") == "PASSED":
                raise GateError(f"gate_cannot_be_marked_passed_without_promotion_review:{gate.get('id')}")
            pending.append(gate["id"])

        required_docs = (
            "release/back01-minimal/PUBLIC_RELEASE_DECISION.md",
            "release/back01-minimal/RELEASE_SCOPE_AND_OS_MATRIX.md",
            "release/back01-minimal/SYSTEMVAULT_NATIVE_HOST_RUNBOOK.md",
            "release/back01-minimal/RELEASE_SIGNATURE_RUNBOOK.md",
        )
        for relative in required_docs:
            if not (root / relative).is_file():
                raise GateError(f"required_gate_document_missing:{relative}")
    except GateError as error:
        print(json.dumps({"valid": False, "error": str(error)}, sort_keys=True))
        return 1

    print(json.dumps({"decision": "PUBLIC_RELEASE_BLOCKED", "pending_gates": pending, "valid": True}, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
