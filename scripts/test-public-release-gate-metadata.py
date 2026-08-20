#!/usr/bin/env python3
"""Exercise public-release gate metadata validation without mutating the repository."""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


EXPECTED_ERROR = "passed_gate_owner_must_be_assigned:SYSTEMVAULT_NATIVE_PER_TARGET"


def run_validator(python: str, validator: Path, root: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [python, str(validator), "--root", str(root)],
        check=False,
        capture_output=True,
        text=True,
    )


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=Path(__file__).resolve().parents[1])
    args = parser.parse_args()
    source_root = args.root.resolve()
    validator = source_root / "scripts/check-public-release-gate.py"

    with tempfile.TemporaryDirectory(prefix="forgelocal-gate-v2-") as temp_dir:
        temp_root = Path(temp_dir) / "repo"
        (temp_root / "scripts").mkdir(parents=True)
        shutil.copytree(source_root / "release", temp_root / "release")
        shutil.copy2(validator, temp_root / "scripts/check-public-release-gate.py")

        baseline = run_validator(sys.executable, temp_root / "scripts/check-public-release-gate.py", temp_root)
        if baseline.returncode != 0:
            print(json.dumps({"valid": False, "stage": "baseline", "output": baseline.stdout.strip()}, sort_keys=True))
            return 1

        state_path = temp_root / "release/back01-minimal/PUBLIC_RELEASE_GATE_STATE.json"
        state = json.loads(state_path.read_text(encoding="utf-8"))
        state["gates"][0]["decision"] = "PASSED"
        state_path.write_text(json.dumps(state, indent=2) + "\n", encoding="utf-8")

        negative = run_validator(sys.executable, temp_root / "scripts/check-public-release-gate.py", temp_root)
        if negative.returncode == 0 or EXPECTED_ERROR not in negative.stdout:
            print(json.dumps({"valid": False, "stage": "negative", "output": negative.stdout.strip()}, sort_keys=True))
            return 1

    print(json.dumps({"baseline": "accepted", "incomplete_passed_gate": "rejected", "valid": True}, sort_keys=True))
    return 0


if __name__ == "__main__":
    sys.exit(main())
