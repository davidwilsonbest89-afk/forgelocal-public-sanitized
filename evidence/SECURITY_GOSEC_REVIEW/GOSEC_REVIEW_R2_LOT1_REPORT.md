# GOSEC-REVIEW-R2 — Lot 1 : filesystem, archives et provenance

## Verdict du lot

```text
GOSEC_REVIEW_R2_LOT1_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```

Ce lot est une revue automatisée/agent avec tests locaux synthétiques. Il ne constitue pas une revue humaine indépendante et ne valide ni Camoufox, ni SystemVault natif, ni Docker/Buildx, ni une release.

## Périmètre

Le périmètre fermé comprend les 43 findings baseline des règles `G703`, `G304`, `G305`, `G122` et `G110` liés aux chemins filesystem, téléchargements, archives et restauration. La matrice complète, une ligne par finding, se trouve dans `GOSEC_REVIEW_R2_LOT1_MATRIX.tsv` et `GOSEC_REVIEW_R2_LOT1_MATRIX.md`.

| Règle | Baseline | Post-correctif | Évolution observée |
|---|---:|---:|---:|
| G703 | 14 | 12 | 2 occurrences statiques de moins; plusieurs sinks restent signalés. |
| G304 | 22 | 21 | 1 occurrence statique de moins. |
| G305 | 3 | 1 | 2 occurrences statiques de moins. |
| G122 | 1 | 1 | WalkDir de création de backup reste à auditer séparément. |
| G110 | 3 | 0 | Les trois occurrences de décompression non bornée ont disparu du scan post-correctif. |
| **Total source-only** | **43** | **35** | **8 occurrences de ces règles en moins**. |

Le scan source-only complet passe de 176 à **155 findings**, mais cette baisse globale n’est pas attribuée exclusivement au lot 1. Gosec reste en échec attendu (`exit_code=1`) et aucun finding n’est masqué.

## Risques confirmés et corrections

Le code antérieur extrayait ZIP/TAR directement dans la destination, ignorait silencieusement les chemins traversal, autorisait des symlinks TAR et ne bornait pas le nombre d’entrées ni la taille développée. La restauration CLI écrivait également directement dans la base cible et pouvait produire un état partiel en cas d’erreur.

Le lot ajoute un staging temporaire transactionnel pour les archives navigateur, une validation de racine et de profondeur, le refus des chemins absolus et traversal, le refus des symlinks/types spéciaux dans les téléchargements, des limites de 4096 entrées, 1 GiB par fichier et 2 GiB par archive, ainsi qu’une activation après extraction complète. La restauration CLI utilise désormais un staging, les limites d’entrées/volume, refuse les hardlinks et types spéciaux, accepte uniquement les symlinks dont la cible résolue reste sous la racine du backup, puis active les racines avec rollback en cas d’échec.

Ces contrôles sont des mitigations applicatives. Les findings statiques G703/G304/G305 restent visibles lorsque le scanner conserve le sink tainté; ils ne sont pas déclarés faux positifs ni clos.

## Tests exécutés

| Test | Résultat |
|---|---:|
| ZIP régulier et remplacement atomique | PASS |
| ZIP traversal `..` avec destination existante préservée | PASS |
| ZIP chemin absolu | PASS |
| ZIP profondeur excessive | PASS |
| TAR régulier | PASS |
| TAR traversal | PASS |
| TAR symlink interdit pour téléchargement | PASS |
| CLI backup symlink interne autorisé | PASS |
| CLI backup symlink externe refusé | PASS |
| CLI backup hardlink refusé | PASS |
| CLI archive partiellement valide sans activation partielle | PASS |
| CLI limite de 4096 entrées | PASS |
| `go test -count=1 -race ./cmd/... ./internal/...` | PASS |
| `go vet ./cmd/... ./internal/...` | PASS |
| `go build ./cmd/... ./internal/...` | PASS |

Les sorties brutes sont conservées dans `GOSEC_REVIEW_R2_LOT1_TARGETED_TESTS_RAW.log`, `GOSEC_REVIEW_R2_LOT1_POSTFIX_SCANS_RAW.log` et `GOSEC_REVIEW_R2_LOT1_ALL_SCANS_RAW.log`. Le JSON Gosec autoritatif est `gosec_source_only_r2_lot1_published.json`; les JSON OSV actuels sont `osv_go_mod_r2_lot1.json` et `osv_pnpm_lock_r2_lot1.json`; le résultat Trivy est `trivy_r2_lot1.json`.

## Findings restant à revoir

La matrice classe 20 des 43 findings dans `MITIGATED_CONTROL_SCANNER_OPEN`, car les chemins concernés disposent maintenant de contrôles et de tests mais restent parfois signalés par le scanner. Les 23 autres restent `NEEDS_MANUAL_REVIEW`, notamment le `G122` sur `WalkDir`, les chemins de backup output/metadata non couverts par ce lot et les occurrences filesystem sans preuve de confinement dédiée.

Le prochain lot doit traiter subprocess/réseau (`G204` et G704 encore ouverts), sans réutiliser ce lot comme allowlist et sans déclarer les findings statiques fermés.

## Contraintes conservées

T28 n’a pas été redémarré, T29 n’a pas commencé et T31–T38 ne sont pas touchés. Les parcours Dashboard/Core/proxy déjà validés ne sont pas rejoués comme campagne complète. Les environnements Camoufox, SystemVault natif, Docker/Buildx, proxy commercial, cookies réels et release restent hors périmètre autorisé.

Le correctif source est publié dans le commit `cd0d2e61990ceb421765f75a26cfd986ad9dc558` sur `validation/operational-v1`. Le statut reste :

```text
GOSEC_REVIEW_R2_LOT1_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_REVIEW_R1_CLASSIFIED_WITH_OPEN_FINDINGS
OPERATIONAL_VALIDATION_PARTIAL_SECURITY_AND_ENVIRONMENT_GATES_OPEN
FORGELOCAL_PRODUCTION_READY=false
```
