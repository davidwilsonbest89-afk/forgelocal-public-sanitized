# GOSEC-R4 — Lot C

## Périmètre

Le lot R4-C couvre les conversions entières, l’aléatoire non cryptographique, les permissions signalées et le motif G101. Il ne couvre ni T28 ni T29, et n’exécute aucun navigateur natif, Camoufox, SystemVault ou release.

## Source et preuves

Le code du lot est publié sur la branche dédiée `validation/gosec-r4` dans le commit `3cd3ac0fb9eb684e61eafb8746c69dda9781df64`, après le commit R4-A `dac05e6d597e98116e90176a69203223354f7aca` et le commit R4-B `53c5e7a9652b4a01e87a6aac729218547fe158a1`.

La campagne finale a été exécutée avec Go 1.25.13 et `GOTOOLCHAIN=local` sur le périmètre source-only `./cmd/... ./internal/...`.

| Contrôle | Résultat |
|---|---:|
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS |
| `go vet ./cmd/... ./internal/...` | PASS |
| `go build ./cmd/... ./internal/...` | PASS |
| Test local téléchargement runtime avec fixture HTTP | PASS |
| Test bornes archive et seed persona | PASS |
| Gosec source-only | exit 1, findings ouverts |

## Évolution Gosec

Le scan final R4-C compte **63 findings**. Les compteurs pertinents sont `G101=1`, `G104=0`, `G107=0`, `G115=3`, `G204=5`, `G302=5`, `G304=15`, `G305=1`, `G404=17`, `G703=9` et `G704=7`.

Les corrections principales sont l’ajout d’un client HTTP explicite avec timeout de 60 secondes et refus des redirections, ainsi que le bornage des tailles d’archive et le refus des conversions négatives ou au-delà de la limite `int64`. Les tests de régression utilisent exclusivement des fixtures synthétiques et un serveur HTTP local.

Les trois G115 restent visibles car le scanner ne prouve pas toutes les bornes par analyse statique, bien que les contrôles applicatifs soient démontrables. Les G302 concernent des répertoires owner-only et des exécutables runtime volontairement en 0755 sous des parents privés. Les G404 concernent uniquement la randomisation d’humanisation et de sélection de fingerprints, sans usage cryptographique. Le G101 porte sur le nom d’un fichier de métadonnées d’admin-token, et non sur une credential embarquée; les tests vérifient la redaction du token brut.

## Verdict du lot

```text
GOSEC_R4_C_MITIGATED_WITH_OPEN_FINDINGS
GOSEC_R4_CLASSIFIED_WITH_OPEN_FINDINGS
```

Aucun finding n’a été masqué par `nosec`, `nolint`, skip global ou allowlist. La matrice complète figure dans `R4_C_CLASSIFICATION.tsv`; la sortie brute est dans `R4_C_FINAL_RAW.log` et le JSON dans `gosec_after_r4c_final.json`.
