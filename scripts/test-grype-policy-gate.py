#!/usr/bin/env python3
from __future__ import annotations

import json
import subprocess
import sys
import tempfile
from pathlib import Path


def run_gate(root: Path, report: Path, exceptions: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, "scripts/grype-policy-gate.py", str(report), str(exceptions)],
        cwd=root,
        text=True,
        capture_output=True,
        check=False,
    )


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    with tempfile.TemporaryDirectory(prefix="grype-policy-test-") as temp:
        work = Path(temp)
        report = work / "report.json"
        exceptions = work / "exceptions.json"
        report.write_text(
            json.dumps(
                {
                    "matches": [
                        {
                            "vulnerability": {"id": "SYNTHETIC-CVE-0001", "severity": "Critical"},
                            "artifact": {"name": "synthetic-component", "version": "1.0.0"},
                        }
                    ]
                }
            ),
            encoding="utf-8",
        )
        exceptions.write_text('{"schema_version": 1, "exceptions": []}\n', encoding="utf-8")
        blocked = run_gate(root, report, exceptions)
        if blocked.returncode != 1 or "GRYPE_POLICY_GATE=FAIL_OPEN_CRITICAL_HIGH=1" not in blocked.stdout:
            print("GRYPE_POLICY_BLOCKING_TEST=FAIL")
            return 1
        exceptions.write_text(
            json.dumps(
                {
                    "schema_version": 1,
                    "exceptions": [
                        {
                            "vulnerability_id": "SYNTHETIC-CVE-0001",
                            "artifact_name": "synthetic-component",
                            "installed_version": "1.0.0",
                            "owner": "security@example.invalid",
                            "justification": "Synthetic policy gate test only",
                            "expires_at": "2099-01-01",
                        }
                    ],
                }
            ),
            encoding="utf-8",
        )
        approved = run_gate(root, report, exceptions)
        if approved.returncode != 0 or "GRYPE_POLICY_GATE=PASS" not in approved.stdout:
            print("GRYPE_POLICY_EXCEPTION_TEST=FAIL")
            return 1
    print("GRYPE_POLICY_BLOCKING_TEST=PASS")
    print("GRYPE_POLICY_EXCEPTION_TEST=PASS")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
