#!/usr/bin/env python3
import json
import pathlib
import sys

root = pathlib.Path(sys.argv[1])

def normalize(path):
    data = json.loads(path.read_text())
    rows = []
    for item in data.get("Issues", []):
        filename = str(item.get("file", ""))
        for marker in ("/internal/", "/cmd/", "/pkg/"):
            if marker in filename:
                filename = filename[filename.index(marker) + 1:]
                break
        rows.append((item.get("rule_id"), filename, item.get("details")))
    return sorted(rows, key=lambda row: tuple("" if value is None else str(value) for value in row))

baseline = normalize(root / "POSTFIX_GOSEC_BASELINE.json")
head = normalize(root / "POSTFIX_GOSEC_HEAD.json")
new = sorted(set(head) - set(baseline))
resolved = sorted(set(baseline) - set(head))
print(f"baseline_count={len(baseline)}")
print(f"head_count={len(head)}")
print(f"new_count={len(new)}")
print(f"resolved_count={len(resolved)}")
print("new_findings=" + json.dumps(new, ensure_ascii=False))
print("resolved_findings=" + json.dumps(resolved, ensure_ascii=False))
