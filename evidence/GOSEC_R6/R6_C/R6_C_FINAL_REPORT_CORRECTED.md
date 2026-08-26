# GOSEC-R6 Lot C — rapport final post-package

Ce document compagnon complète `R6_C_REPORT.md` avec les références du package et de la vérification publique. Le package R6-C v1 et son rapport inclus sont conservés sans reconstruction.

## Résultats

La baseline et le scan post-commit contiennent 26 findings Lot C sur 46 globaux : G302=5, G115=3, G404=17 et G101=1. Aucun finding n’a été masqué. Les G302/G115 restent `MITIGATED_CONTROL_SCANNER_OPEN`; les G404 de simulation/fingerprint et le G101 de nom de fichier restent `SCANNER_OPEN_MANUAL_REVIEW`.

Les tests ciblés permissions/bornes/token, `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et `git diff --check` sont PASS. Gosec reste à exit code 1 avec 46 findings globaux. Gitleaks, Govulncheck, OSV Go/pnpm et Trivy sont PASS dans les périmètres documentés.

## Conservation et vérification publique

Le commit source est `6bdda53c917b9617f530729528a0fa6bf80b94f2`. Le commit d’évidence est `b7a86b8c6fe9daa60745fd8fc3ae93257aa69cdf`. Le commit package est `d85ed7ed9ce3ac9485dc5d61afcde6d943c53dbb`. Le package est `forgelocal-gosec-r6c-final-v1.zip` et `.tar.gz`, avec sidecars SHA-256. Le bundle `forgelocal-gosec-r6c-delta-6bdda53-b7a86b8.bundle` retourne PASS avec `git bundle verify`.

La vérification publique `R6C_PUBLIC_VERIFICATION_RAW.log` retourne PASS pour les hashes, extractions fraîches, manifestes, absence réelle de membre `SMOKE_INTEGRATED_PROXY`, bundle, clone GitHub neuf, checkout explicite du commit package et `git fsck --full` avec exit code 0.

## Limites et verdict

Aucun secret réel, compte réel, cookie, proxy commercial, site externe, runtime de production, SystemVault natif, Camoufox, Docker/Buildx ou plateforme Windows/macOS n’a été utilisé. Les usages d’aléatoire non cryptographique sont limités à la simulation et au choix de fingerprints; la génération du token d’administration reste en `crypto/rand`. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.

Le verdict est `GOSEC_R6_LOT_C_CLASSIFIED_WITH_OPEN_FINDINGS`, `GOSEC_R6_PARTIAL_ENVIRONMENT_UNAVAILABLE` et `FORGELOCAL_PRODUCTION_READY=false`.
