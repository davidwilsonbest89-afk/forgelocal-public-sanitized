# GOSEC-R6 Lot A — paths et confinement

## Périmètre et méthode

Le Lot A a été exécuté sur `validation/gosec-r6` depuis le HEAD baseline `41057639b04d74e9476058c99d9ff7c2d2d2fe36`. Le périmètre autoritatif est le JSON Gosec source-only `./cmd/... ./internal/...`. La baseline contient 21 findings Lot A : G703=9, G304=11 et G305=1. La classification individuelle complète est dans `R6_A_FINDING_CLASSIFICATION.tsv`.

Les contrôles ont été effectués avec fixtures synthétiques locales uniquement. Les premiers résultats post-correctif ont été mesurés avant le commit source et sont conservés; la preuve finale `R6_A_FINAL_POSTCOMMIT_RAW.log` est rattachée au commit source publié `142477ae0d576eae937b16660899fd973d6f2464`. Les tests ajoutés couvrent la lecture normale et le refus de symlink pour workflow, fingerprints et fichiers browser. Les tests d’archives existants couvrent traversal, chemins absolus, profondeur excessive, symlinks, types spéciaux et conservation de la destination.

## Résultats mesurés

| Règle | Baseline Lot A | Après correctif | Évolution | Classification |
|---|---:|---:|---:|---|
| G304 | 11 | 0 | -11 | CORRECTED_ROOT_SCOPED |
| G703 | 9 | 7 | -2 | 7 MITIGATED_CONTROL_SCANNER_OPEN |
| G305 | 1 | 1 | 0 | MITIGATED_CONTROL_SCANNER_OPEN |
| **Total Lot A** | **21** | **8** | **-13** |  |

Le scan global passe de 59 à 46 findings dans le JSON après retry, avec un exit code Gosec égal à 1 parce que des findings restent ouverts. La réduction nette de 13 correspond aux 11 G304 corrigés et aux deux sinks G703 remplacés par des helpers root-scoped. Elle ne constitue pas une allowlist ni un masquage.

## Corrections appliquées

Les lectures de workflow, fingerprints, token bootstrap, profils migrés et préférences Chromium utilisent désormais des racines `os.Root` et refusent les symlinks ou types non réguliers. Les destinations de téléchargement, les entrées TAR et les écritures d’extraction/restore utilisent également des helpers root-scoped. Les bornes d’archives, les refus de symlink/type spécial, le staging et le rollback ont été conservés.

Les sept G703 restants sont classés `MITIGATED_CONTROL_SCANNER_OPEN`: les chemins sont filtrés ou confinés par les contrôles applicatifs existants, mais Gosec conserve une alerte de taint sur les opérations de chemin. Le G305 du lien de backup reste aussi ouvert avec preuve du contrôle `safeBackupPath`; il n’est pas déclaré clos automatiquement.

## Validation

Les tests ciblés browser/workflow/fingerprint/server, la suite `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et `git diff --check` sont PASS dans `R6_A_FINAL_POSTCOMMIT_RAW.log`. Le scan post-commit est `gosec_r6a_after_postcommit.json` et son SHA-256 est `e48747b6428418f32ab7552d886231da154e9fdcd4b1ca2062900786ef84691a`. Les raw précédents, y compris le scan avant commit et le retry, sont conservés séparément.

Le Lot A reste un lot de hardening partiel : G703/G305 ouverts au scanner, aucun runtime de production, aucun chemin Windows/macOS simulé, aucun secret ou compte réel. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.
