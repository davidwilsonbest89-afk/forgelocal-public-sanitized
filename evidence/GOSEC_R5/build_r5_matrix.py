import csv, json
from pathlib import Path
repo=Path('/home/ubuntu/forgelocal_self_validation_v6/operational_validation_v1/repo')
src=repo/'evidence/GOSEC_R5/gosec_source_only_baseline.json'
out_tsv=repo/'evidence/GOSEC_R5/GOSEC_R5_BASELINE_MATRIX.tsv'
out_md=repo/'evidence/GOSEC_R5/GOSEC_R5_BASELINE_MATRIX.md'
data=json.loads(src.read_text())
issues=data.get('Issues', [])
rows=[]
for n, issue in enumerate(issues, 1):
    rule=issue.get('rule_id','')
    file=issue.get('file','')
    rel=file
    marker='/repo/'
    if marker in file:
        rel=file.split(marker,1)[1]
    rows.append({
      'id': f'R5-{n:03d}',
      'rule': rule,
      'file': rel,
      'function': issue.get('function') or '-',
      'line': issue.get('line') or '-',
      'source_entry': 'Gosec source-only baseline on 079c452',
      'executed_path': 'To be confirmed per lot; source path is statically reachable',
      'protected_asset': 'application I/O, runtime, profile, token, network or integrity depending on path',
      'preconditions': 'Requires the relevant local caller/path and input condition; lot-specific proof required',
      'impact_cia': 'Triage required: confidentiality/integrity/availability classification',
      'existing_controls': 'R4 controls retained; see R4 evidence and R5 threat model',
      'existing_test': 'Existing package tests plus lot-specific regression to be selected',
      'proposed_fix': 'Narrow fix or documented manual review; no global suppression',
      'lot': {'G703':'R5-A','G304':'R5-A','G305':'R5-A','G204':'R5-B','G704':'R5-B','G302':'R5-C','G115':'R5-C','G404':'R5-C','G101':'R5-C'}.get(rule,'R5-C'),
      'decision': 'OPEN_REVIEW',
      'evidence': f'evidence/GOSEC_R5/gosec_source_only_baseline.json#{n}'
    })
fields=list(rows[0]) if rows else []
out_tsv.parent.mkdir(parents=True, exist_ok=True)
with out_tsv.open('w', newline='') as f:
    w=csv.DictWriter(f, fieldnames=fields, delimiter='\t', lineterminator='\n'); w.writeheader(); w.writerows(rows)
counts={}
for r in rows: counts[r['rule']]=counts.get(r['rule'],0)+1
with out_md.open('w') as f:
    f.write('# GOSEC-R5 baseline matrix\n\n')
    f.write('Cette matrice est générée depuis le JSON Gosec source-only du commit `079c452`. Elle contient une ligne par finding; aucune ligne n’est supprimée ou clôturée automatiquement.\n\n')
    f.write('| Rule | Count | Lot |\n|---|---:|---|\n')
    for rule in sorted(counts):
        f.write(f'| {rule} | {counts[rule]} | {rows[next(i for i,r in enumerate(rows) if r["rule"]==rule)]["lot"]} |\n')
    f.write(f'| **Total** | **{len(rows)}** | **R5-A/R5-B/R5-C** |\n\n')
    f.write('Les décisions autorisées après triage sont `CORRECTED_AND_VERIFIED`, `MITIGATED_CONTROL_SCANNER_OPEN`, `SCANNER_OPEN_MANUAL_REVIEW`, `HISTORICAL_NOT_REACHABLE` et `BLOCKED_ENVIRONMENT_REQUIRED`. La valeur initiale `OPEN_REVIEW` est provisoire et doit être remplacée avant le rapport final.\n')
print('rows', len(rows)); print('counts', counts)
