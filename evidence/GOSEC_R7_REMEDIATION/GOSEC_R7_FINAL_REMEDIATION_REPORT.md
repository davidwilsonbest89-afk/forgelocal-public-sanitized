# GOSEC-R7 — rapport final de remédiation locale ciblée

## Résumé exécutif

La remédiation locale a été exécutée sur la branche dédiée `validation/gosec-r7-final-remediation`, créée depuis le HEAD public R7 découvert dynamiquement `b907dfcd68c290144e2b922e352d5a937e9b3259`. Aucun code produit n’a été modifié. Les contrôles locaux réexécutés n’ont démontré ni fuite de secret réel, ni sortie de confinement, ni exécution arbitraire, ni bypass d’authentification, ni dial externe non contrôlé, ni mélange de profils, ni mutation partielle exploitable.

La campagne confirme que les 46 alertes restent ouvertes dans Gosec mais ne constituent pas, sur la base des preuves disponibles, 46 failles exploitables. Aucun défaut P0, P1 ou défaut P2 concret supplémentaire n’a été démontré. La conclusion est donc une décision de risque résiduel, pas une fermeture de sécurité.

## Baseline et lignée

Le HEAD R6 découvert est `3656dbad4bfef0381e1f9d837271d293ecffe292`. Il est ancêtre du HEAD R7 public et de la branche de remédiation. Le diff R6→R7 hors `evidence/GOSEC_R7/` est vide. Le package R7 v3 de départ a été vérifié depuis un clone neuf avec sidecars, extraction ZIP, Gitleaks `--no-git`, bundle et `git fsck --full` exit code 0. Les packages R6 ne sont ni reconstruits ni réécrits.

Le JSON Gosec autoritatif a le SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a` et contient exactement 46 findings : G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7 et G704=7.

## Contrôles exécutés sur la branche de remédiation

| Contrôle | Résultat |
|---|---|
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS, exit code 0 |
| `go vet ./cmd/... ./internal/...` | PASS, exit code 0 |
| `go build ./cmd/... ./internal/...` | PASS, exit code 0 |
| Gosec source-only | exit code 1, 46 findings attendus |
| Gitleaks source-only | PASS, exit code 0 |
| Gitleaks extraction R7 v3 avec `--no-git` | PASS, exit code 0 |
| Govulncheck | PASS, aucune vulnérabilité trouvée |
| OSV Go et pnpm | PASS, aucun problème trouvé |
| Trivy filesystem, vuln/secret/misconfig | PASS dans le périmètre excluant `artifacts`, `evidence`, `continuation` |
| Syft | PASS, inventaire produit |
| Dashboard TypeScript `pnpm run check` | PASS, exit code 0 |
| Chromium headless synthétique, profil jetable | PASS, exit code 0 |
| Processus Chromium résiduel | Aucun détecté |
| Firefox standard | Non exécuté, binaire absent |
| `git diff --check` | PASS, exit code 0 |

Les tests ciblés d’archives, runtime, permissions, token, workflow et réseau sont réutilisés depuis `R7_CONSOLIDATED_TESTS_RAW.log`, où ils ont tous retourné exit code 0. Cette distinction est conservée : ils sont une preuve R7 antérieure, tandis que race/vet/build/Gosec/scanners/Chromium ont été réexécutés sur la branche de remédiation.

## Findings et décision de remédiation

| Classification | Nombre | Décision |
|---|---:|---|
| `CORRECTION_REQUISE_P0` | 0 | Aucun défaut critique démontré |
| `CORRECTION_REQUISE_P1` | 0 | Aucun défaut urgent démontré |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 23 | Contrôles démontrés, findings conservés ouverts |
| `SCANNER_OPEN_MANUAL_REVIEW` | 18 | Revue manuelle conservée |
| `BLOCKED_ENVIRONMENT_REQUIRED` | 5 | Validation native nécessaire |
| `HISTORICAL_NOT_REACHABLE` | 0 | Aucun classement historique artificiel |
| **Total** | **46** | Baseline inchangée |

La matrice `R7_REMEDIATION_FINDINGS_MATRIX.tsv` reprend chaque finding avec fichier, ligne, entrée, chemin exécuté, préconditions, impacts, garde, test, classification et décision de remédiation. Les G115, G302, G305, G703 et G704 conservent leurs contrôles et leurs tests sans correction supplémentaire. Les G404 restent limités à la simulation/fingerprint non cryptographique. Le G101 reste une revue manuelle du nom `.api-token.meta`, sans credential brut. Les G204 restent bloqués par l’environnement natif et ne sont pas simulés.

## Outils et environnements indisponibles

Semgrep, Grype, Shellcheck et Yamllint n’étaient pas présents. L’espace disponible était tendu, avec un raw de campagne signalant 404 MiB disponibles et 100 % d’utilisation; leur installation n’a donc pas été tentée afin de ne pas mettre la sandbox en danger. Ils restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE`, jamais PASS.

Docker/Buildx, Camoufox, Firefox standard, SystemVault natif, Windows, macOS et les chemins natifs `open`, `rundll32` et `xattr` restent non exécutés. `xdg-open` est présent et `DISPLAY` est défini, mais la session GUI, la cible loopback, le cleanup et l’absence de processus résiduel n’étaient pas établis par un protocole indépendant; aucun lancement n’a été effectué.

## Absence de correction de code

La comparaison manuelle et Git montre qu’aucun fichier `cmd/` ou `internal/` n’a été modifié. Aucun `nosec`, `nolint`, skip global, allowlist globale, exclusion de fichier ou réduction artificielle de périmètre n’a été ajouté. La campagne ne produit donc pas de nouveau finding corrigé; elle confirme la décision R7 précédente sans masquer les 46 alertes.

## Conservation et verdict

Les raw logs, la matrice, le JSON Gosec, le rapport, les résultats Chromium et les artefacts de package seront publiés uniquement sur `validation/gosec-r7-final-remediation`. `evidence/SMOKE_INTEGRATED_PROXY/` reste local, non suivi et absent des packages. T28 n’est pas rouvert, T29 n’est pas démarré, T31–T38 restent intacts et aucune release n’est préparée.

```text
GOSEC_R7_FINAL_REMEDIATION_NO_CONCRETE_CRITICAL_DEFECT
GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R7_PARTIAL_ENVIRONMENT_UNAVAILABLE
FORGELOCAL_PRODUCTION_READY=false
```
