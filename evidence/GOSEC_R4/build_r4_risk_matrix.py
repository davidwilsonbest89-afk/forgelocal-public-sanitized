#!/usr/bin/env python3
import json
import re
from pathlib import Path

root = Path(__file__).parent
repo = root.parent.parent
issues = json.loads((root / 'gosec_source_only_baseline.json').read_text()).get('Issues', [])

def rel(path):
    marker = '/operational_validation_v1/repo/'
    return path.split(marker, 1)[1] if marker in path else path

def function_at(path, line):
    p = repo / path
    try:
        lines = p.read_text(errors='replace').splitlines()
    except OSError:
        return 'unavailable'
    try:
        n = int(str(line).split('-',1)[0])
    except ValueError:
        return 'unavailable'
    current = 'package scope'
    for text in lines[:n]:
        m = re.match(r'\s*func\s+(?:\([^)]*\)\s*)?([A-Za-z0-9_]+)', text)
        if m:
            current = m.group(1)
    return current

def lot(rule):
    if rule == 'G104': return 'R4-A'
    if rule in {'G703','G304','G305'}: return 'R4-B'
    if rule in {'G115','G404','G302','G101','G107'}: return 'R4-C'
    if rule in {'G204','G704'}: return 'R4-D'
    return 'R4-REVIEW'

def context(rule, path):
    if rule == 'G104': return ('local persistence, archive, transport or audit I/O', 'faulty/partial I/O or closed resource', 'integrity/availability; confidentiality if cleanup leaks data', 'bounded error propagation, rollback or cleanup exists on some paths', 'interrupted write/close/rename and rollback tests')
    if rule in {'G703','G304','G305'}: return ('user-selected local path, archive, backup or artifact', 'tainted path, link, malformed archive or race', 'confidentiality/integrity/availability', 'lexical confinement, os.Root, archive limits and staged activation on selected paths', 'traversal, symlink, hardlink, special-type and malformed archive tests')
    if rule in {'G115'}: return ('size, port, identifier or buffer value', 'out-of-range external or archive value', 'integrity/availability', 'positive/range checks exist on selected paths', 'boundary and overflow tests where available')
    if rule == 'G404': return ('humanization/fingerprint behavior', 'weak PRNG influences non-cryptographic variation', 'low integrity/privacy impact; no auth impact established', 'not used for token/key/auth decisions by current review', 'stability tests and manual behavior review')
    if rule == 'G302': return ('runtime executable or sensitive file mode', 'permission mode broader than intended', 'confidentiality/integrity', 'owner-only directories and explicit file modes on R3 paths', 'mode assertions on sensitive runtime paths')
    if rule == 'G101': return ('admin-token identifier or token metadata context', 'pattern is mistaken for embedded credential, or real secret is committed', 'confidentiality/integrity', 'redaction and synthetic-token-only policy', 'gitleaks and source inspection')
    if rule == 'G107': return ('URL/network request', 'uncontrolled external or redirectable target', 'confidentiality/integrity/availability', 'loopback validation and local HTTP policy on reviewed paths', 'loopback/external/redirect/timeout tests')
    if rule in {'G204','G704'}: return ('subprocess or WebSocket/HTTP bridge', 'tainted command/endpoint or missing timeout', 'integrity/availability/confidentiality', 'allowlisted local targets, deadlines and platform-specific guards on reviewed paths', 'synthetic local refusal/timeout tests')
    return ('source finding', 'requires review', 'undetermined', 'not yet established', 'not yet established')

rows=[]
for idx, i in enumerate(issues,1):
    rule=i.get('rule_id','')
    path=rel(i.get('file',''))
    asset, pre, impact, control, test = context(rule,path)
    rows.append({
      'id':f'R4-{idx:03d}','rule':rule,'severity':i.get('severity',''),'confidence':i.get('confidence',''),
      'file':path,'function':function_at(path,i.get('line','')),'line':i.get('line',''),
      'source':'GOSEC_R3_FINAL_SOURCE_ONLY_SCAN','executed_path':asset,'asset':asset,
      'precondition':pre,'impact_CIA':impact,'application_control':control,'existing_test':test,
      'proposed_fix':f'Inspect and harden the {lot(rule)} path; preserve scanner visibility and add negative/positive regression tests.',
      'lot':lot(rule),'status':'NEEDS_MANUAL_REVIEW','details':' '.join(str(i.get('details','')).split())
    })
headers=['id','rule','severity','confidence','file','function','line','source','executed_path','asset','precondition','impact_CIA','application_control','existing_test','proposed_fix','lot','status','details']
with (root/'GOSEC_R4_BASELINE_MATRIX.tsv').open('w',encoding='utf-8') as f:
    f.write('\t'.join(headers)+'\n')
    for r in rows: f.write('\t'.join(str(r[h]).replace('\t',' ').replace('\n',' ') for h in headers)+'\n')
with (root/'GOSEC_R4_BASELINE_MATRIX.md').open('w',encoding='utf-8') as f:
    f.write('# GOSEC-R4 baseline risk matrix\n\n')
    f.write(f'Source-only baseline from the published R3 final scan. All {len(rows)} rows begin as `NEEDS_MANUAL_REVIEW`; later lots may change status only with code, tests and evidence.\n\n')
    f.write('| Lot | Rule set | Count |\n|---|---|---:|\n')
    for l in ['R4-A','R4-B','R4-C','R4-D','R4-REVIEW']:
        n=sum(r['lot']==l for r in rows)
        rules=', '.join(sorted({r['rule'] for r in rows if r['lot']==l}))
        f.write(f'| `{l}` | `{rules}` | {n} |\n')
    f.write('\nThe complete per-finding matrix is in `GOSEC_R4_BASELINE_MATRIX.tsv`.\n')
print(f'rows={len(rows)}')
for l in ['R4-A','R4-B','R4-C','R4-D','R4-REVIEW']:
    print(l, sum(r['lot']==l for r in rows))
