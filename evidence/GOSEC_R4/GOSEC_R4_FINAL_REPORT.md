# GOSEC-R4 — Rapport final

## 1. Objet et périmètre

GOSEC-R4 a été exécuté sur la branche dédiée `validation/gosec-r4`, créée depuis le HEAD R3 distant découvert dynamiquement `80048180d4ec5241c08146ade698d53a3f29454d`. La branche historique `validation/operational-v1` n’a pas été modifiée. Le périmètre source-only autoritatif reste `./cmd/... ./internal/...`; les snippets partiels sous `artifacts/` ne sont pas utilisés pour déclarer un PASS.

La mission couvre quatre lots fermés : erreurs I/O et cleanup, chemins backup/download et confinement, conversions/permissions/aléatoire/motif G101, puis subprocess GUI et réseau/WebSocket. T28 n’a pas été rouvert, T29 n’a pas commencé et T31–T38 n’ont pas été modifiés.

## 2. Lignée et commits

| Référence | Rôle |
|---|---|
| `80048180d4ec5241c08146ade698d53a3f29454d` | HEAD R3 de départ vérifié |
| `78cf74e...` | Baseline et threat model R4 publiés |
| `dac05e6d597e98116e90176a69203223354f7aca` | R4-A : propagation I/O, MCP, cleanup et erreurs de transport |
| `53c5e7a9652b4a01e87a6aac729218547fe158a1` | R4-B : accès root-scoped des backups/downloads et tests symlink |
| `3cd3ac0fb9eb684e61eafb8746c69dda9781df64` | R4-C : tailles d’archives bornées, redirections download refusées, contrôles de conversions |

Le HEAD R4 source publié avant l’empaquetage est `3cd3ac0fb9eb684e61eafb8746c69dda9781df64`. Les preuves finales et la vérification publique ultérieure seront publiées dans des commits de conservation distincts; le package sera explicitement rattaché à ce `source_head`.

## 3. Résultats des lots

| Lot | Résultat | Évolution ou contrôle principal |
|---|---|---|
| R4-A | `GOSEC_R4_A_MITIGATED_WITH_OPEN_FINDINGS` | G104 réduit à 0; erreurs MCP, cleanup relay/browser, keyboard et réponses JSON rendues observables. |
| R4-B | `GOSEC_R4_B_MITIGATED_WITH_OPEN_FINDINGS` | Backup et download root-scoped; symlink externe, sortie symlinkée, hardlink, traversal et types spéciaux couverts. |
| R4-C | `GOSEC_R4_C_MITIGATED_WITH_OPEN_FINDINGS` | G107 réduit à 0; bornes archives et seed testées; G101/G115/G302/G404 restent scanner-visibles et classés. |
| R4-D | `GOSEC_R4_D_MITIGATED_WITH_OPEN_FINDINGS` | CLI/WebSocket loopback, ports, chemins, userinfo, query, fragment, timeout et redirections contrôlés; G204/G704 restent scanner-visibles. |

## 4. Scans et non-régressions

La campagne finale du 26 août 2026 a utilisé Go 1.25.13 avec `GOTOOLCHAIN=local`.

| Contrôle | Résultat |
|---|---:|
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS, exit 0 |
| `go vet ./cmd/... ./internal/...` | PASS, exit 0 |
| `go build ./cmd/... ./internal/...` | PASS, exit 0 |
| Govulncheck | PASS, aucune vulnérabilité trouvée |
| Gitleaks | PASS, aucun leak trouvé |
| OSV Go corrigé avec `--lockfile go.mod` | PASS, `results=[]` |
| OSV pnpm corrigé avec `--lockfile forge-dashboard/pnpm-lock.yaml` | PASS, `results=[]` |
| Trivy vuln/misconfig/secret | PASS, 0/0/0 |
| Syft | PASS, SBOM produite; les composants UNKNOWN ne sont pas assimilés à une revue de licence terminée |
| `forge-dashboard/pnpm run check` | PASS, exit 0 |
| Gosec source-only | exit 1, 63 findings ouverts |

La première invocation OSV du log global employait une syntaxe invalide et a retourné `127`; elle est conservée comme défaut procédural. La commande officielle `osv-scanner scan --lockfile ...` a ensuite été exécutée séparément et a retourné 0 pour Go et pnpm.

## 5. Gosec final

| Règle | Findings |
|---|---:|
| G101 | 1 |
| G104 | 0 |
| G107 | 0 |
| G115 | 3 |
| G204 | 5 |
| G302 | 5 |
| G304 | 15 |
| G305 | 1 |
| G404 | 17 |
| G703 | 9 |
| G704 | 7 |
| **Total** | **63** |

Les classifications détaillées sont dans `R4_C_CLASSIFICATION.tsv` et `R4_D_CLASSIFICATION.tsv`. Elles distinguent `MITIGATED_CONTROL_SCANNER_OPEN`, `SCANNER_OPEN_MANUAL_REVIEW` et les véritables suppressions confirmées par le scan suivant. Aucun `nosec`, `nolint`, skip global ou allowlist n’a été ajouté.

## 6. Environnements et gates

Le préflight confirme un hôte Linux x86_64. Docker, Buildx, xattr Darwin et Camoufox ne sont pas disponibles. SystemVault natif n’a pas été testé. Les ports 19281, 3001 et 19282–19287 sont libres. Aucun proxy réel, cookie réel, compte externe, secret ou donnée utilisateur n’a été utilisé.

Les gates suivantes restent inchangées :

```text
PUBLIC_RELEASE_BLOCKED=true
SCAN_BLOCKED_UNKNOWN=true
NATIVE_SYSTEMVAULT_NOT_TESTED=true
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```

## 7. Verdict

```text
GOSEC_R4_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R4_PARTIAL_ENVIRONMENT_UNAVAILABLE
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
PUBLIC_RELEASE_BLOCKED
FORGELOCAL_PRODUCTION_READY=false
```

GOSEC-R4 ne constitue donc ni une clôture complète de sécurité ni une validation de production. Les findings restants sont visibles, classés et rattachés à des contrôles, mais leur fermeture nécessite une décision et une revue supplémentaires. La prochaine étape n’est pas démarrée automatiquement.
