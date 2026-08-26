# GOSEC-R5 Lot R5-A — chemins et confinement

## Périmètre

Le lot R5-A couvre uniquement le confinement de fichiers de configuration, de journalisation, du marqueur `.version` runtime et du token bootstrap `.api-token`. Il ne modifie ni T28, ni T29, ni T31–T38, et ne démarre aucun runtime navigateur ou environnement de production.

Le commit source publié est `54ed3a4964806eeb4880c9ebb3949d410c335174` sur `validation/gosec-r5`. Les preuves de ce lot sont conservées séparément dans le commit d’évidence qui suit.

## Résultats exécutés

| Contrôle | Résultat | Preuve |
|---|---:|---|
| Tests ciblés config/browser/API/server | PASS | `R5_A_FINAL_RAW.log` |
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS | `R5_A_FINAL_RAW.log` |
| `go vet ./cmd/... ./internal/...` | PASS | `R5_A_FINAL_RAW.log` |
| `go build ./cmd/... ./internal/...` | PASS | `R5_A_FINAL_RAW.log` |
| Gosec source-only | exit 1 attendu avec findings ouverts | `gosec_after_r5a.json`, `R5_A_FINAL_RAW.log` |

Le scan Gosec source-only est exécuté sur `./cmd/... ./internal/...` uniquement. Il passe de **63 à 60 findings**, sans finding nouveau. Les trois suppressions sont trois occurrences G304 : deux dans `internal/config/config.go` et une dans `internal/browser/download.go`.

| Règle | Baseline R5 | Après R5-A | Variation |
|---|---:|---:|---:|
| G101 | 1 | 1 | +0 |
| G104 | 0 | 0 | 0 |
| G107 | 0 | 0 | 0 |
| G115 | 3 | 3 | +0 |
| G204 | 5 | 5 | +0 |
| G302 | 5 | 5 | +0 |
| G304 | 15 | 12 | -3 |
| G305 | 1 | 1 | +0 |
| G404 | 17 | 17 | +0 |
| G703 | 9 | 9 | +0 |
| G704 | 7 | 7 | +0 |
| **Total** | **63** | **60** | **-3** |

## Contrôles et limites

Les nouveaux tests vérifient qu’un marqueur `.version` régulier est lu, qu’un marqueur symlinké est refusé, qu’un fichier de configuration symlinké est refusé et que l’absence de configuration conserve les valeurs par défaut. Le token bootstrap est également relu lorsqu’il est régulier et refusé lorsqu’il est symlinké. Les chemins sont ouverts relativement à une racine contrôlée avec `os.OpenRoot`; aucune suppression Gosec artificielle n’a été ajoutée.

Les 60 findings restants sont individualisés dans `R5_A_CLASSIFICATION.tsv` et restent `OPEN_REVIEW`; ce lot ne les clôt pas par convention. Les outils et gates environnementaux non exécutés restent non exécutés. Le statut demeure `GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
