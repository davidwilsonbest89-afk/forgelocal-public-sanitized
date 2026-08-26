# ForgeLocal — matrice finale de qualification d’environnement

## Périmètre et règle de lecture

Cette qualification est exécutée depuis un clone neuf du dépôt public, sur une branche dédiée `validation/final-environment-qualification` créée à partir du HEAD public vérifié de `validation/gosec-r7`. Elle ne modifie pas le code produit et ne réinterprète pas les findings Gosec. Le dossier historique `evidence/SMOKE_INTEGRATED_PROXY/` est exclu de toute nouvelle exécution et de tout package.

Les statuts utilisés sont **PASS**, **FAIL**, **BLOCKED_ENVIRONMENT_REQUIRED**, **NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE** et **INHERITED_FROM_R7**. Un statut `INHERITED_FROM_R7` désigne une preuve existante relue et vérifiée, non une exécution nouvelle de cette qualification.

## Références de base vérifiées

| Élément | Valeur | Statut | Preuve |
| --- | --- | --- | --- |
| HEAD public de départ `validation/gosec-r7` | `b907dfcd68c290144e2b922e352d5a937e9b3259` | PASS | `FINAL_ENVIRONMENT_BASELINE_DISCOVERY_RAW.log` |
| `source_head` R7 | `3656dbad4bfef0381e1f9d837271d293ecffe292` | INHERITED_FROM_R7 | `evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt` |
| `evidence_head` R7 | `1dca108197f90e379184474ad37bbb9f386fe309` | INHERITED_FROM_R7 | `evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt` |
| `package_commit` R7 | `5cd16b11f62b8c16582973001ce81bff9ef03dcf` | INHERITED_FROM_R7 | `evidence/GOSEC_R7/R7_PUBLIC_VERIFICATION_V4_RAW.log` |
| ZIP R7 v4 | `evidence/GOSEC_R7/forgelocal-gosec-r7-final-v4.zip` | PASS / INHERITED_FROM_R7 | hash et `unzip -t` vérifiés dans le raw log |
| TAR R7 v4 | `evidence/GOSEC_R7/forgelocal-gosec-r7-final-v4.tar.gz` | PASS / INHERITED_FROM_R7 | hash et test TAR vérifiés dans le raw log |
| Bundle R7 v4 | `evidence/GOSEC_R7/forgelocal-gosec-r7-delta-3656dba-1dca108.bundle` | PASS / INHERITED_FROM_R7 | hash, `git bundle verify` et log public R7 v4 |
| Baseline Gosec | SHA-256 `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a` | PASS / INHERITED_FROM_R7 | `evidence/GOSEC_R7/R7_GOSEC_BASELINE.json` |

## Contrôles nouvellement exécutés

| Contrôle | Commande ou méthode | Résultat | Preuve et limites |
| --- | --- | --- | --- |
| Clone public neuf | `gh repo clone ... repo -- --no-tags` | PASS | URL, branche et HEAD consignés dans le raw log |
| Checkout explicite | `git checkout -B validation/final-environment-qualification refs/remotes/origin/validation/gosec-r7` | PASS | branche dédiée créée depuis le HEAD vérifié |
| Intégrité Git du clone | `git fsck --full` | PASS | `reference-integrity-raw.log` et vérification finale |
| Go tests avec race | `go test -count=1 -race ./cmd/... ./internal/...` | PASS | replay final réussi ; un premier essai sans C compiler était non exécutable puis rejoué après installation sûre de `build-essential` |
| Go vet | `go vet ./cmd/... ./internal/...` | PASS | replay final réussi |
| Go build | `go build ./cmd/... ./internal/...` | PASS | replay final réussi |
| Format Git | `git diff --check` | PASS | aucune erreur |
| Gosec source-only | `gosec -fmt json ... ./cmd/... ./internal/...` | FAIL | 46 findings conservés ; distribution autoritative `G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7, G704=7, TOTAL=46` |
| Gitleaks source-only | `gitleaks detect --source <source-only> --no-git --redact` | PASS | aucun leak trouvé, périmètre source-only |
| Semgrep source-only | `semgrep scan --config auto --json ...` | FAIL | 18 résultats et 4 erreurs de règles ; aucune exclusion artificielle ajoutée |
| ShellCheck | scripts shell source-only | FAIL | code 123 ; findings consignés sans modification du code |
| Yamllint | YAML source-only | FAIL | code 123 ; principalement lignes trop longues dans le lockfile, sans correction artificielle |
| Govulncheck Go | `govulncheck -json ./cmd/... ./internal/...` | PASS | exit code 0 ; sortie brute conservée |
| OSV récursif | `osv-scanner scan source -r <source-only> --format json` | PASS | exit code 0 ; absence de vulnérabilité signalée par l’outil dans ce périmètre |
| Trivy filesystem | `trivy fs --scanners vuln,secret,misconfig ... <source-only>` | PASS | exit code 0 ; sortie JSON conservée |
| Syft SBOM | `syft scan dir:<source-only>` JSON et CycloneDX | PASS | SBOM conservés ; 745 artefacts détectés dans la sortie JSON |
| Grype sur SBOM | `grype sbom:<syft-json> -o json` | PASS | exit code 0 ; sortie brute conservée |
| Chromium loopback | deux profils jetables, page synthétique `127.0.0.1` | PASS | démarrage, page loopback et cleanup vérifiés ; ce n’est pas Camoufox |
| Chromium proxy indisponible | proxy contrôlé `127.0.0.1:1`, cible synthétique | PASS | page d’erreur Chromium observée ; contrôle limité à ce fixture, sans proxy réel |
| Firefox standard | binaire, headless, profils, proxy | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | binaire Firefox absent ; Camoufox non téléchargé |
| Docker/Buildx | `docker version`, `docker info`, `docker buildx version`, contextes | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | client et daemon Docker absents |
| SystemVault natif Linux | Secret Service/D-Bus/keyring réel | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | outillage Secret Service/keyring indisponible ; verdict `NATIVE_SYSTEMVAULT_NOT_TESTED` |
| Camoufox | installation et exécution | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | aucune provenance/checksum traçable fournie ; téléchargement volontairement non effectué |
| Windows/macOS natifs | `rundll32`, Keychain, `open`, `xattr` | BLOCKED_ENVIRONMENT_REQUIRED | aucun OS natif correspondant ; aucune simulation autorisée |

## Contrôles hérités, non revendiqués comme nouveaux

Les résultats historiques R7, y compris le package Gosec v4, la baseline à 46 findings, les logs publics, les bundles et les tests déjà documentés dans `evidence/GOSEC_R7/`, restent classés `INHERITED_FROM_R7`. Ils servent à établir la chaîne de provenance et la comparaison, mais ne sont pas présentés comme des exécutions de la branche finale.

## Verdict obligatoire

```text
GOSEC_R7_FINAL_ENVIRONMENT_QUALIFICATION_PARTIAL
GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS
FORGELOCAL_PRODUCTION_READY=false
```

Aucune gate n’est levée, aucune release n’est préparée, T29 n’est pas démarré et aucune alerte ouverte n’est transformée en absence de finding.
