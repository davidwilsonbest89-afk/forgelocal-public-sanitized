#!/usr/bin/env python3
import json
from pathlib import Path

root = Path(__file__).parent
source = root / "gosec_source_only_baseline.json"
out_tsv = root / "GOSEC_R3_BASELINE_MATRIX.tsv"
out_md = root / "GOSEC_R3_BASELINE_MATRIX.md"

data = json.loads(source.read_text())
issues = data.get("Issues", [])
selected = {"G703", "G304", "G305", "G122", "G110", "G301", "G302", "G306", "G104", "G404", "G115", "G118", "G101"}
rows = []
for i, issue in enumerate(issues, 1):
    rule = issue.get("rule_id", "")
    if rule not in selected:
        continue
    file_name = issue.get("file", "")
    if file_name.startswith("/tmp/gosec_r3_phase0_clone/"):
        file_name = file_name.replace("/tmp/gosec_r3_phase0_clone/", "")
    detail = " ".join(str(issue.get("details", "")).split())
    rows.append({
        "id": f"R3-BL-{i:03d}",
        "rule": rule,
        "severity": issue.get("severity", ""),
        "confidence": issue.get("confidence", ""),
        "file": file_name,
        "line": issue.get("line", ""),
        "column": issue.get("column", ""),
        "details": detail,
        "initial_status": "OPEN_BASELINE_REQUIRES_TRIAGE",
    })

headers = ["id", "rule", "severity", "confidence", "file", "line", "column", "details", "initial_status"]
with out_tsv.open("w", encoding="utf-8") as f:
    f.write("\t".join(headers) + "\n")
    for row in rows:
        f.write("\t".join(str(row[h]).replace("\t", " ").replace("\n", " ") for h in headers) + "\n")

counts = {}
for row in rows:
    counts[row["rule"]] = counts.get(row["rule"], 0) + 1
with out_md.open("w", encoding="utf-8") as f:
    f.write("# GOSEC-R3 baseline matrix\n\n")
    f.write("This matrix is generated from the source-only scan at the final overnight HEAD. No finding is suppressed or treated as closed.\n\n")
    f.write("| Rule | Count | Initial status |\n|---|---:|---|\n")
    for rule in sorted(counts):
        f.write(f"| `{rule}` | {counts[rule]} | `OPEN_BASELINE_REQUIRES_TRIAGE` |\n")
    f.write(f"\n**Selected findings:** {len(rows)}. **All baseline findings:** {len(issues)}.\n\n")
    f.write("| ID | Rule | Severity | Confidence | File | Line | Details | Status |\n|---|---|---|---|---|---:|---|---|\n")
    for row in rows:
        details = row["details"].replace("|", "\\|")
        f.write(f"| {row['id']} | `{row['rule']}` | {row['severity']} | {row['confidence']} | `{row['file']}` | {row['line']} | {details} | `{row['initial_status']}` |\n")

print(f"all_findings={len(issues)}")
print(f"selected_findings={len(rows)}")
for rule in sorted(counts):
    print(f"{rule}={counts[rule]}")
print(f"tsv={out_tsv}")
print(f"markdown={out_md}")
