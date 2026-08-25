import json
from collections import Counter
from pathlib import Path
base = Path('/home/ubuntu/forgelocal_self_validation_v5')
e = base / 'evidence'

def read_json(name, default):
    try:
        return json.loads((e / name).read_text())
    except Exception:
        return default

def read_json_path(path, default):
    try:
        return json.loads(Path(path).read_text())
    except Exception:
        return default

sem = read_json('SEMGREP_FINAL.json', {})
sem_results = sem.get('results', []) if isinstance(sem, dict) else []
sem_rules = Counter((x.get('check_id') or 'unknown') for x in sem_results)

def grype_stats(name):
    d = read_json(name, {})
    matches = d.get('matches', []) if isinstance(d, dict) else []
    sev = Counter(((x.get('vulnerability') or {}).get('severity') or 'Unknown') for x in matches)
    pkgs = Counter(((x.get('artifact') or {}).get('name') or 'unknown') for x in matches)
    return len(matches), sev, pkgs

cdx_count, cdx_sev, cdx_pkgs = grype_stats('GRYPE_CYCLONEDX.json')
spdx_count, spdx_sev, spdx_pkgs = grype_stats('GRYPE_SPDX.json')
axe = read_json_path(base / 'SELF_VALIDATION_AXE_RESULTS.json', {})
axe_violations = axe.get('violations', []) if isinstance(axe, dict) else []
axe_blocking = axe.get('blocking', []) if isinstance(axe, dict) else []
axe_impact = Counter((x.get('impact') or 'unknown') for x in axe_violations)

out = []
out.append('# Analyse des contrôles complémentaires — v5')
out.append('')
out.append('Cette analyse conserve les résultats non nuls et les distingue des erreurs d’outillage. Aucun finding n’est masqué ni transformé en PASS.')
out.append('')
out.append('| Contrôle | Résultat | Détail |')
out.append('|---|---|---|')
out.append(f'| Semgrep 1.174.0 | FINDINGS | {len(sem_results)} résultat(s), règles : ' + ', '.join(f'`{k}`={v}' for k,v in sem_rules.items()) + ' |')
out.append(f'| Grype 0.117.0 CycloneDX | PASS technique / findings à classer | {cdx_count} correspondance(s), sévérités : ' + ', '.join(f'{k}={v}' for k,v in cdx_sev.items()) + ' |')
out.append(f'| Grype 0.117.0 SPDX | PASS technique / findings à classer | {spdx_count} correspondance(s), sévérités : ' + ', '.join(f'{k}={v}' for k,v in spdx_sev.items()) + ' |')
out.append(f'| Axe Playwright | BLOCKED_BY_FINDINGS | {len(axe_violations)} violation(s), dont {len(axe_blocking)} sérieuse(s)/critique(s) : ' + ', '.join(f'{k}={v}' for k,v in axe_impact.items()) + ' |')
out.append('')
out.append('## Semgrep')
out.append('')
out.append('| # | Règle | Fichier | Ligne | Message | Risque | Propriétaire | Condition de levée |')
out.append('|---:|---|---|---:|---|---|---|---|')
for i, x in enumerate(sem_results, 1):
    start = x.get('start') or {}
    msg = (x.get('extra') or {}).get('message') or ''
    out.append(f"| {i} | `{x.get('check_id','unknown')}` | `{x.get('path','unknown')}` | {start.get('line','')} | {msg.replace('|','\\|')} | Revue SAST requise ; risque dépendant du contexte d’usage | Mainteneurs ForgeLocal | Revue humaine, correction ou justification ciblée versionnée, puis rerun |")
out.append('')
out.append('## Grype')
out.append('')
out.append('| SBOM | # correspondances | Sévérités | Packages principaux | Décision |')
out.append('|---|---:|---|---|---|')
out.append(f"| CycloneDX | {cdx_count} | {dict(cdx_sev)} | {', '.join(k for k,_ in cdx_pkgs.most_common(12))} | À trier par CVE/package/version ; aucun upgrade automatique |")
out.append(f"| SPDX | {spdx_count} | {dict(spdx_sev)} | {', '.join(k for k,_ in spdx_pkgs.most_common(12))} | À trier par CVE/package/version ; aucun upgrade automatique |")
out.append('')
out.append('## Axe')
out.append('')
for i, x in enumerate(axe_violations, 1):
    nodes = x.get('nodes') or []
    selectors = ', '.join(str(n.get('target')) for n in nodes[:4])
    out.append(f"{i}. `{x.get('id','unknown')}` — impact `{x.get('impact','unknown')}`, {x.get('description','')}. Cibles : `{selectors}`. Condition de levée : correction UI ciblée ou justification accessibilité revue humainement, puis rerun Axe.")
out.append('')
out.append('## Décision')
out.append('')
out.append('Les contrôles complémentaires ont accru la couverture mais ne lèvent pas les gates. Semgrep et Axe produisent des findings ; Grype termine techniquement mais ses correspondances doivent être triées ; la campagne reste en attente de revue indépendante.')
(base / 'SELF_VALIDATION_V5_COMPLEMENTS_ANALYSIS.md').write_text('\n'.join(out) + '\n')
(base / 'SELF_VALIDATION_V5_COMPLEMENTS_COUNTS.json').write_text(json.dumps({'semgrep_results': len(sem_results), 'semgrep_rules': dict(sem_rules), 'grype_cyclonedx_matches': cdx_count, 'grype_cyclonedx_severity': dict(cdx_sev), 'grype_spdx_matches': spdx_count, 'grype_spdx_severity': dict(spdx_sev), 'axe_violations': len(axe_violations), 'axe_blocking': len(axe_blocking), 'axe_impact': dict(axe_impact)}, indent=2) + '\n')
