import json
from collections import Counter, defaultdict
from pathlib import Path

base = Path('/home/ubuntu/forgelocal_self_validation_v5')
e = base / 'evidence'

def load(path, default):
    try:
        return json.loads(path.read_text())
    except Exception:
        return default

def issues(path):
    d = load(path, {})
    if isinstance(d, list):
        return d
    return d.get('Issues') or d.get('issues') or []

def key(x):
    return (x.get('FromLinter') or x.get('linter') or x.get('Linter') or '', x.get('Text') or x.get('message') or x.get('Description') or '', x.get('Pos', {}).get('Filename') if isinstance(x.get('Pos'), dict) else x.get('file') or x.get('File') or '', str((x.get('Pos') or {}).get('Line', '') if isinstance(x.get('Pos'), dict) else x.get('line') or x.get('Line') or ''))

baseline = issues(e / 'GOLANGCI_BASELINE.json')
head = issues(e / 'GOLANGCI_FINAL.json')
bset = {key(x) for x in baseline}
hset = {key(x) for x in head}
new = [x for x in head if key(x) not in bset]
resolved = [x for x in baseline if key(x) not in hset]

tree_reports = sorted(e.glob('GITLEAKS_TREE_*.json'))
gitleaks_counts = Counter()
gitleaks_files = Counter()
gitleaks_total = 0
for p in tree_reports:
    for x in issues(p):
        gitleaks_total += 1
        gitleaks_counts[x.get('RuleID') or x.get('rule_id') or 'UNKNOWN'] += 1
        f = x.get('File') or x.get('file') or ''
        gitleaks_files[f.split('/repo/')[-1].split('/tmp/forgelocal-v5-gitleaks-trees.')[-1]] += 1

out = []
out.append('# Analyse comparative obligatoire — v5')
out.append('')
out.append('Les résultats sont produits avec GolangCI-Lint 2.13.1 (binaire publié, compilé avec Go 1.27.0) contre la baseline et le HEAD, et avec un scan Gitleaks par arbre de commit pour compenser le mode `--log-opts` qui annonçait zéro commit. Les valeurs d’alerte Gitleaks restent redacted.')
out.append('')
out.append('| Contrôle | Baseline | HEAD | Nouveau | Résolu | Décision |')
out.append('|---|---:|---:|---:|---:|---|')
out.append(f'| GolangCI-Lint 2.13.1 | {len(baseline)} | {len(head)} | {len(new)} | {len(resolved)} | Findings conservés et classés ci-dessous |')
out.append(f'| Gitleaks arbres de commits | — | {gitleaks_total} détections cumulées | — | — | Findings historiques redacted, non-PASS |')
out.append('')
out.append('## Findings GolangCI-Lint nouveaux')
out.append('')
out.append('| # | Linter | Règle | Fichier | Ligne | Message | Risque | Propriétaire | Condition de levée |')
out.append('|---:|---|---|---|---:|---|---|---|---|')
for i, x in enumerate(new, 1):
    pos = x.get('Pos') or {}
    out.append(f"| {i} | {x.get('FromLinter') or x.get('linter') or 'unknown'} | `{x.get('FromLinter') or 'unknown'}` | `{pos.get('Filename') or x.get('file') or 'unknown'}` | {pos.get('Line') or x.get('line') or ''} | {(x.get('Text') or x.get('message') or '').replace('|','\\|')} | Revue sécurité/qualité requise avant correction ; aucune exploitation établie par le scan seul | Mainteneurs ForgeLocal | Correction/test ou justification versionnée, puis rerun avec code attendu |")
out.append('')
out.append('## Findings GolangCI-Lint résolus entre baseline et HEAD')
out.append('')
out.append(f'{len(resolved)} diagnostic(s) présents à la baseline et absents au HEAD selon la clé linter/message/position. Cette variation est conservée pour revue et ne constitue pas une approbation globale.')
out.append('')
out.append('## Gitleaks explicite par arbres')
out.append('')
out.append(f'La plage contient {len(tree_reports)} commits. Le scan explicite a produit {gitleaks_total} détections cumulées, principalement `generic-api-key` dans des fixtures et preuves historiques. Les JSON individuels sont livrés sans secret en clair ; les alertes ne sont pas reclassées automatiquement comme faux positifs.')
out.append('')
out.append('| Règle | Détections cumulées |')
out.append('|---|---:|')
for rule, count in gitleaks_counts.most_common():
    out.append(f'| `{rule}` | {count} |')
out.append('')
out.append('## Compléments Go')
out.append('')
out.append('`go test -shuffle=on -count=3 ./...` et `go test -shuffle=on -count=3 -race ./...` ont terminé avec code 0. Aucun code produit n’a été modifié pendant cette campagne.')
(base / 'SELF_VALIDATION_V5_MANDATORY_ANALYSIS.md').write_text('\n'.join(out) + '\n')
(base / 'SELF_VALIDATION_V5_MANDATORY_COUNTS.json').write_text(json.dumps({'golangci_baseline': len(baseline), 'golangci_head': len(head), 'golangci_new': len(new), 'golangci_resolved': len(resolved), 'gitleaks_tree_commits': len(tree_reports), 'gitleaks_tree_findings': gitleaks_total, 'gitleaks_rules': dict(gitleaks_counts)}, indent=2) + '\n')
