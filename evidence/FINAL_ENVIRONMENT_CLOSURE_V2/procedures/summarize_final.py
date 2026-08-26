#!/usr/bin/env python3
import json
from collections import Counter
from pathlib import Path

root = Path('/home/ubuntu/forgelocal-final-environment-closure')
out = root / 'independent-results'
summary = []
for name in ['gosec.json', 'semgrep.json', 'shellcheck.json', 'trivy.json', 'syft.json']:
    p = out / name
    if not p.exists():
        continue
    try:
        d = json.loads(p.read_text())
    except Exception as exc:
        summary.append(f'{name}\tparse_error={exc}')
        continue
    if name == 'gosec.json':
        summary.append(f'gosec_findings={len(d.get("Issues", []))}\tstats={d.get("Stats")}')
    elif name == 'semgrep.json':
        summary.append(f'semgrep_findings={len(d.get("results", []))}\tsemgrep_errors={len(d.get("errors", []))}')
        for i, r in enumerate(d.get('results', []), 1):
            e = r.get('extra', {})
            summary.append(f'SEMGREP_{i}\t{r.get("path")}\tline={r.get("start",{}).get("line")}\trule={r.get("check_id")}\tseverity={e.get("severity")}\tmessage={e.get("message")}')
        for i, e in enumerate(d.get('errors', []), 1):
            summary.append(f'SEMGREP_ERROR_{i}\tpath={e.get("path")}\tmessage={e.get("message")}')
    elif name == 'shellcheck.json':
        summary.append(f'shellcheck_findings={len(d)}')
        counts = Counter((x.get('code'), x.get('level')) for x in d)
        summary.append('shellcheck_codes=' + ';'.join(f'SC{k[0]}:{k[1]}={v}' for k, v in sorted(counts.items())))
        for i, x in enumerate(d, 1):
            summary.append(f'SHELLCHECK_{i}\t{x.get("file")}\tline={x.get("line")}\tcolumn={x.get("column")}\tcode=SC{x.get("code")}\tlevel={x.get("level")}\tmessage={x.get("message")}')
    elif name == 'trivy.json':
        results = d.get('Results') or []
        summary.append(f'trivy_result_sections={len(results)}')
        summary.append(f'trivy_findings={sum(len(x.get("Vulnerabilities") or []) + len(x.get("Secrets") or []) + len(x.get("Misconfigurations") or []) for x in results)}')
    elif name == 'syft.json':
        summary.append(f'syft_artifacts={len(d.get("artifacts", []))}')

lines = (out / 'yamllint.txt').read_text(errors='replace').splitlines() if (out / 'yamllint.txt').exists() else []
summary.append(f'yamllint_diagnostics={len(lines)}')
by_file = Counter()
for line in lines:
    if ':' in line:
        by_file[line.split(':', 1)[0]] += 1
summary.append('yamllint_by_file=' + ';'.join(f'{k}={v}' for k, v in sorted(by_file.items())))
(root / 'FINAL_STATIC_SUMMARY.tsv').write_text('\n'.join(summary) + '\n')
print('\n'.join(summary))
