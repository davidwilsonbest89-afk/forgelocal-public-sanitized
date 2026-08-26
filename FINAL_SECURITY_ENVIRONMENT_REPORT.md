# Rapport final de sécurité et de qualification d’environnement ForgeLocal

## Conclusion exécutive

La qualification locale séparée a été menée depuis un clone neuf du dépôt public, sans reprendre le développement produit et sans modifier les findings Gosec. Le code source de la branche de départ est resté inchangé. Les contrôles Go, Gitleaks, Govulncheck, OSV, Trivy, Syft, Grype et Chromium synthétique ont été rejoués avec des sorties brutes conservées. Les findings Gosec restent ouverts et la qualification d’environnement est partielle, car Docker/Buildx, Firefox standard, Camoufox, SystemVault natif et les plateformes natives nécessaires ne sont pas disponibles dans cet environnement.

> `GOSEC_R7_FINAL_ENVIRONMENT_QUALIFICATION_PARTIAL`
> `GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS`
> `FORGELOCAL_PRODUCTION_READY=false`

## Identité Git et provenance R7

| Référence | Valeur | Qualification |
| --- | --- | --- |
| Dépôt | `https://github.com/davidwilsonbest89-afk/forgelocal-public-sanitized.git` | dépôt public utilisé pour le clone neuf |
| Branche de départ | `validation/gosec-r7` | découverte par `git ls-remote` |
| HEAD public vérifié de départ | `b907dfcd68c290144e2b922e352d5a937e9b3259` | vérifié localement et contre le remote |
| Branche dédiée | `validation/final-environment-qualification` | seule branche de qualification créée et publiée |
| HEAD final de la branche dédiée | à lire avec `git rev-parse validation/final-environment-qualification` après le commit final | commit de conservation de cette mission, non auto-référencé dans son propre manifeste |
| Commit d’évidence R7 | `1dca108197f90e379184474ad37bbb9f386fe309` | `INHERITED_FROM_R7` |
| Commit du package R7 | `5cd16b11f62b8c16582973001ce81bff9ef03dcf` | `INHERITED_FROM_R7` |
| Commit source R7 | `3656dbad4bfef0381e1f9d837271d293ecffe292` | `INHERITED_FROM_R7` |

La chaîne historique a été relue depuis `evidence/GOSEC_R7/R7_PUBLIC_VERIFICATION_V4_RAW.log` et `evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt`. Ces fichiers ne sont pas des preuves d’une nouvelle exécution : ils sont explicitement classés `INHERITED_FROM_R7`.

## Artefacts R7 v4 vérifiés

| Artefact | Nom exact | SHA-256 | Statut |
| --- | --- | --- | --- |
| ZIP | `evidence/GOSEC_R7/forgelocal-gosec-r7-final-v4.zip` | `80b0546eca714826023bfc4cc3e33381b1b9d3dfe52900771dd00a2cd2ba5ed8` | `PASS` et `INHERITED_FROM_R7` |
| TAR | `evidence/GOSEC_R7/forgelocal-gosec-r7-final-v4.tar.gz` | `faf43da4e69e20aa4dc59863173b0f64b8a411f102d7dcd4c46b22fc8089fa7e` | `PASS` et `INHERITED_FROM_R7` |
| Bundle | `evidence/GOSEC_R7/forgelocal-gosec-r7-delta-3656dba-1dca108.bundle` | `6f732c627ee58529898753c09f292c5c0e79b9ee4a32cf6bd58b755f7ae4edb0` | `PASS` et `INHERITED_FROM_R7` |
| Manifeste v4 | `evidence/GOSEC_R7/R7_FINAL_MANIFEST_V4.txt` | conservé tel quel, manifest non auto-référentiel | `INHERITED_FROM_R7` |
| Log public v4 | `evidence/GOSEC_R7/R7_PUBLIC_VERIFICATION_V4_RAW.log` | conservé tel quel | `INHERITED_FROM_R7` |
| Baseline Gosec | `evidence/GOSEC_R7/R7_GOSEC_BASELINE.json` | `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a` | `PASS` et `INHERITED_FROM_R7` |

La vérification nouvelle a confirmé les sommes SHA-256 attendues, `unzip -t`, le test de lecture TAR, `git bundle verify`, les merge-bases et `git fsck --full`. Le clone neuf documenté par R7 v4 est une preuve historique relue ; le clone neuf de la présente mission est consigné séparément dans le raw log.

## Findings et contrôles

| Contrôle | Statut | Observation principale |
| --- | --- | --- |
| `go test -count=1 -race ./cmd/... ./internal/...` | `PASS` | replay final réussi après mise à disposition d’une toolchain C Ubuntu sûre |
| `go vet ./cmd/... ./internal/...` | `PASS` | exit code 0 |
| `go build ./cmd/... ./internal/...` | `PASS` | exit code 0 |
| `git diff --check` | `PASS` | aucune erreur de format |
| Gosec source-only `./cmd/... ./internal/...` | `FAIL` | 46 findings conservés : G101=1, G115=3, G204=5, G302=5, G304=0, G305=1, G404=17, G703=7, G704=7 |
| Gitleaks source-only avec `--no-git` | `PASS` | aucun leak trouvé dans l’extraction source-only |
| Semgrep source-only | `FAIL` | 18 résultats et 4 erreurs de règles ; aucune allowlist artificielle |
| ShellCheck source-only | `FAIL` | exit code 123, sorties et stderr conservés |
| Yamllint source-only | `FAIL` | exit code 123, notamment lignes longues dans le lockfile |
| Govulncheck Go | `PASS` | exit code 0 dans le périmètre Go |
| OSV source récursif | `PASS` | exit code 0 |
| Trivy filesystem | `PASS` | exit code 0, JSON conservé |
| Syft JSON/CycloneDX | `PASS` | SBOM conservés ; 745 artefacts dans la sortie JSON |
| Grype sur SBOM Syft | `PASS` | exit code 0 |
| Chromium loopback avec deux profils jetables | `PASS` | page synthétique loopback rendue et profils distincts utilisés |
| Chromium proxy contrôlé indisponible | `PASS` | page d’erreur observée avec proxy loopback invalide ; portée limitée au fixture |
| Docker et Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | client et daemon absents |
| Firefox standard | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | binaire absent |
| Camoufox | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | non téléchargé faute de provenance/checksum traçable |
| SystemVault natif Linux | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Secret Service/keyring non disponible ; `NATIVE_SYSTEMVAULT_NOT_TESTED` |
| Windows/macOS natifs | `BLOCKED_ENVIRONMENT_REQUIRED` | aucun OS et service natif correspondant ; aucune simulation |
| Preuves et résultats R7 existants | `INHERITED_FROM_R7` | relus pour provenance uniquement, non revendiqués comme nouveaux |

## Conservation nouvelle

Les résultats nouveaux sont conservés sous `evidence/FINAL_ENVIRONMENT_QUALIFICATION/` avec les raw logs, manifestes, résultats JSON des scanners et scripts de reproduction. Le dossier historique `evidence/SMOKE_INTEGRATED_PROXY/` n’est pas touché et n’est inclus dans aucun package. Le package de cette mission porte un manifeste non auto-référentiel ; son propre commit final est obtenu par `git rev-parse` après publication de la branche dédiée.

Aucune gate n’a été levée. Aucune release n’a été préparée. T29 n’a pas été démarré. Le statut `PASS` d’un scanner ou d’un test local ne ferme pas les findings Gosec et ne qualifie pas les environnements natifs absents.
