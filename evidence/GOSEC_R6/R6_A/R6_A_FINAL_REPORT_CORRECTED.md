# GOSEC-R6 Lot A — rapport final corrigé après package

Le rapport `R6_A_REPORT.md` et le package R6-A v1 sont conservés. Ce document compagnon ajoute les références post-package et la vérification publique corrigée, sans modifier le contenu utilisé pour construire le package.

## Résultats

La baseline Lot A était G703=9, G304=11 et G305=1. Le scan post-commit retourne G703=7, G304=0 et G305=1. Le scan global passe de 59 à 46 findings. Les 11 G304 et deux sinks G703 supprimés sont rattachés aux helpers `os.Root`; les sept G703 et le G305 restants restent `MITIGATED_CONTROL_SCANNER_OPEN`.

Les tests ciblés browser/workflow/fingerprint/server, `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et `git diff --check` sont PASS dans `R6_A_FINAL_POSTCOMMIT_RAW.log`. Le JSON post-commit est `gosec_r6a_after_postcommit.json`, SHA-256 `e48747b6428418f32ab7552d886231da154e9fdcd4b1ca2062900786ef84691a`.

## Package et vérification publique

Le commit source est `142477ae0d576eae937b16660899fd973d6f2464`. Le commit d’évidence utilisé pour le package est `77ed72b31695d4beaa49e6697f4e0c5419849364`. Le commit package est `22737a4c03855bf1963278da6b1f3ab83dd23d88`. Le bundle delta `forgelocal-gosec-r6a-delta-142477a-77ed72b.bundle` est vérifié par `git bundle verify`.

Le premier raw `R6A_PUBLIC_VERIFICATION_RAW.log` est conservé : son contrôle textuel de SMOKE avait signalé à tort les mentions légitimes du chemin exclu dans les manifestes/raw. Le raw corrigé `R6A_PUBLIC_VERIFICATION_CORRECTED_RAW.log` vérifie les noms d’entrées d’archive et retourne PASS pour hashes, extractions fraîches, manifestes, absence de membre `SMOKE_INTEGRATED_PROXY`, bundle, clone public neuf, checkout et `git fsck --full` avec exit code 0.

## Limites et invariants

Le Lot A reste partiel : G703/G305 restent ouverts au scanner; aucun runtime de production, chemin Windows/macOS simulé, secret ou compte réel n’a été utilisé. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts. Les verdicts sont `GOSEC_R6_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
