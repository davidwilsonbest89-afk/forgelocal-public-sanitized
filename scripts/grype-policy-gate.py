#!/usr/bin/env python3
"""Fail closed on unapproved Critical/High Grype matches.

The gate deliberately reports only counts and never prints match contents.
Exceptions are exact, reviewed records keyed by vulnerability ID, artifact name,
and installed version.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path


def load(path: Path) -> object:
    with path.open(encoding="utf-8") as handle:
        return json.load(handle)


def exception_key(vulnerability_id: str, artifact_name: str, artifact_version: str) -> str:
    return "|".join((vulnerability_id, artifact_name, artifact_version))


def main(argv: list[str]) -> int:
    if len(argv) != 3:
        print("usage: grype-policy-gate.py REPORT EXCEPTIONS", file=sys.stderr)
        return 2
    report_path = Path(argv[1])
    exceptions_path = Path(argv[2])
    report = load(report_path)
    exceptions_doc = load(exceptions_path)
    if not isinstance(report, dict) or not isinstance(exceptions_doc, dict):
        print("GRYPE_POLICY_GATE=FAIL_INVALID_JSON_SHAPE")
        return 2
    raw_exceptions = exceptions_doc.get("exceptions", [])
    if not isinstance(raw_exceptions, list):
        print("GRYPE_POLICY_GATE=FAIL_INVALID_EXCEPTION_SCHEMA")
        return 2
    approved = set()
    for item in raw_exceptions:
        if not isinstance(item, dict):
            print("GRYPE_POLICY_GATE=FAIL_INVALID_EXCEPTION_ENTRY")
            return 2
        required = (item.get("vulnerability_id"), item.get("artifact_name"), item.get("installed_version"), item.get("owner"), item.get("justification"), item.get("expires_at"))
        if not all(isinstance(value, str) and value.strip() for value in required):
            print("GRYPE_POLICY_GATE=FAIL_EXCEPTION_REQUIRES_OWNER_JUSTIFICATION_EXPIRY")
            return 2
        approved.add(exception_key(item["vulnerability_id"], item["artifact_name"], item["installed_version"]))

    matches = report.get("matches", [])
    if not isinstance(matches, list):
        print("GRYPE_POLICY_GATE=FAIL_INVALID_MATCHES")
        return 2
    blocking = 0
    approved_count = 0
    for match in matches:
        if not isinstance(match, dict):
            continue
        vulnerability = match.get("vulnerability") or {}
        artifact = match.get("artifact") or {}
        severity = str(vulnerability.get("severity", "")).strip().upper()
        if severity not in {"CRITICAL", "HIGH"}:
            continue
        key = exception_key(str(vulnerability.get("id", "")), str(artifact.get("name", "")), str(artifact.get("version", "")))
        if key in approved:
            approved_count += 1
        else:
            blocking += 1

    if blocking:
        print(f"GRYPE_POLICY_GATE=FAIL_OPEN_CRITICAL_HIGH={blocking}")
        print(f"GRYPE_POLICY_APPROVED_EXCEPTIONS={approved_count}")
        return 1
    print("GRYPE_POLICY_GATE=PASS")
    print(f"GRYPE_POLICY_APPROVED_EXCEPTIONS={approved_count}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
