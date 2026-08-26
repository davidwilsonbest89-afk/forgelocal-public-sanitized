# GOSEC-R5 Lot R5-B — subprocess et réseau

## Périmètre

R5-B examine les findings G204 et G704 restants ainsi que le contrôle token bootstrap apparu dans le commit source R5-A. Aucun refactor global n’est appliqué : les appels `open`, `xdg-open`, `rundll32` et `xattr` restent des cas dépendants de la plateforme, tandis que les appels HTTP CLI passent déjà par une validation loopback, un timeout et un refus de redirection externe.

## Résultats exécutés

| Contrôle | Résultat | Preuve |
|---|---:|---|
| Tests réseau CLI positifs et négatifs synthétiques | PASS | `R5_B_FINAL_RAW.log` |
| Suite race initiale | FAIL transitoire sur `TestRequest_GlobalLimit`; échec conservé | `R5_B_FINAL_RAW.log` |
| Suite race isolée puis rerun complet | PASS | `R5_B_RERUN_RAW.log` |
| `go vet ./cmd/... ./internal/...` | PASS | `R5_B_FINAL_RAW.log` |
| `go build ./cmd/... ./internal/...` | PASS | `R5_B_FINAL_RAW.log` |
| Gosec source-only | exit 1 avec findings ouverts | `gosec_after_r5b.json`, `R5_B_FINAL_RAW.log` |

Le scan post-R5-B observe **59 findings**, contre 60 dans `gosec_after_r5a.json`, sans finding nouveau. Une occurrence G304 disparaît dans `internal/api/router.go:833` après le confinement du fichier `.api-token`; ce contrôle est couvert par `TestLoadOrCreateTokenRejectsSymlinkedTokenFile`. Les sept G704 et cinq G204 restent visibles et ouverts, sans suppression ni nosec/nolint.

| Règle | Après R5-A | Après R5-B | Variation |
|---|---:|---:|---:|
| G101 | 1 | 1 | +0 |
| G104 | 0 | 0 | +0 |
| G107 | 0 | 0 | +0 |
| G115 | 3 | 3 | +0 |
| G204 | 5 | 5 | +0 |
| G302 | 5 | 5 | +0 |
| G304 | 12 | 11 | -1 |
| G305 | 1 | 1 | +0 |
| G404 | 17 | 17 | +0 |
| G703 | 9 | 9 | +0 |
| G704 | 7 | 7 | +0 |
| **Total** | **60** | **59** | **-1** |

La matrice individualisée est `R5_B_CLASSIFICATION.tsv`. Le statut global reste `GOSEC_R5_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R5_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
