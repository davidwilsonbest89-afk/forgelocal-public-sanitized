# GOSEC-R7 — décision finale de risque et conservation candidate

## Décision

R7 est une **décision de risque résiduel**, et non une clôture complète de sécurité. La revue a confirmé la baseline Gosec source-only autoritative de 46 findings et n’a démontré aucun P0 ni P1. Aucun correctif de code supplémentaire n’est justifié uniquement pour réduire le compteur Gosec.

La branche dédiée est `validation/gosec-r7`. Le HEAD R6 de départ découvert et vérifié est `3656dbad4bfef0381e1f9d837271d293ecffe292`. Les commits R7 ajoutés ensuite sont des preuves uniquement; le diff entre le HEAD R6 et le HEAD R7 final est limité à `evidence/GOSEC_R7/`. Le dossier local `evidence/SMOKE_INTEGRATED_PROXY/` reste non suivi et exclu.

## Baseline et classification

Le JSON R7 contient exactement 46 findings et son SHA-256 est `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`, identique au JSON R6-C post-commit.

| Règle | Nombre | Classification R7 |
|---|---:|---|
| G101 | 1 | `SCANNER_OPEN_MANUAL_REVIEW` |
| G115 | 3 | `MITIGATED_CONTROL_SCANNER_OPEN` |
| G204 | 5 | `BLOCKED_ENVIRONMENT_REQUIRED` |
| G302 | 5 | `MITIGATED_CONTROL_SCANNER_OPEN` |
| G304 | 0 | Aucun finding courant |
| G305 | 1 | `MITIGATED_CONTROL_SCANNER_OPEN` |
| G404 | 17 | `SCANNER_OPEN_MANUAL_REVIEW` |
| G703 | 7 | `MITIGATED_CONTROL_SCANNER_OPEN` |
| G704 | 7 | `MITIGATED_CONTROL_SCANNER_OPEN` |
| **Total** | **46** | **findings ouverts** |

| Classification | Nombre |
|---|---:|
| `CORRECTION_REQUISE_P0` | 0 |
| `CORRECTION_REQUISE_P1` | 0 |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 23 |
| `SCANNER_OPEN_MANUAL_REVIEW` | 18 |
| `HISTORICAL_NOT_REACHABLE` | 0 |
| `BLOCKED_ENVIRONMENT_REQUIRED` | 5 |

La matrice contient une ligne par finding avec fichier, ligne, règle, entrée, atteignabilité, contrôle, test, décision et preuve. Aucun finding encore signalé par Gosec n’est déclaré clos; aucun finding de fixture n’est déclaré historique sans preuve d’inatteignabilité depuis le code courant.

## Contrôles exécutés depuis R7

Depuis R7, les tests ciblés d’archives, runtime, permissions de groupes, token administrateur, workflow réseau et CLI réseau ont retourné exit code 0. La suite `go test -count=1 -race ./cmd/... ./internal/...`, `go vet ./cmd/... ./internal/...`, `go build ./cmd/... ./internal/...` et le contrôle diff ont également retourné exit code 0.

Gosec source-only a retourné exit code 1 avec exactement les 46 findings attendus. Gitleaks source-only et Gitleaks sur extraction avec `--no-git`, Govulncheck, OSV Go/pnpm, Trivy et Syft ont retourné exit code 0 dans leurs périmètres documentés. Le contrôle TypeScript Dashboard a retourné exit code 0.

## Contrôles hérités de R6-C

La vérification physique du package R6-C, ses hashes, ses extractions, son bundle, son clone public neuf, son checkout, son `git fsck --full` et son Gitleaks d’extraction sont des preuves héritées et explicitement conservées; elles ne sont pas présentées comme une nouvelle exécution R7. Les tests ciblés R7 ont été rejoués séparément et sont consignés dans `R7_CONSOLIDATED_TESTS_RAW.log`.

## Analyse des findings

Les G703, G305, G302, G115 et G704 restent scanner-visibles malgré les contrôles applicatifs démontrés et les tests négatifs; ils restent `MITIGATED_CONTROL_SCANNER_OPEN`. Les G404 concernent la simulation humanisée et la sélection non cryptographique de fingerprints; aucun usage ne protège un token, une clé, une session ou une autorisation. Le G101 vise le nom `.api-token.meta`; les métadonnées contiennent un digest et un état, pas le credential brut.

Les cinq G204 concernent les appels natifs `open`, `xdg-open`, `rundll32` et `xattr`. La simple présence d’une GUI Linux ne constitue pas une validation native. Aucune exécution dangereuse n’a été lancée; macOS et Windows ne sont pas disponibles. Ces findings restent `BLOCKED_ENVIRONMENT_REQUIRED`.

## Références exactes du package R7 v2

| Référence | Valeur |
|---|---|
| `source_head` | `3656dbad4bfef0381e1f9d837271d293ecffe292` |
| `evidence_head` | `dda55254f658f12f07e000212db983dd5aef665d` |
| `package_commit` | `ddd2c68e1fc108acf89fcd745d9ddd6a139a9d58` |
| `public_verification_head` | `f3ddc4d1da8f525103efee4ddc2a9554fbd1f50b` |
| ZIP SHA-256 | `7eb7ad17e2ddcfee4f9f99070020a553ce10996d075b45469083ac469dfc7caa` |
| TAR SHA-256 | `bbe4ba8e4cb4b4024b2745833c35dbd897552b8d6aff49dc4253f19439d0dcd8` |
| Bundle | `forgelocal-gosec-r7-delta-3656dba-dda5525.bundle` |
| Vérification publique | clone neuf, checkout explicite, `fsck=0`, hashes PASS, extraction ZIP/TAR PASS, Gitleaks extraction PASS, bundle verify PASS, SMOKE exclu |

Le package R7 v1 et ses artefacts historiques sont conservés. La correction de son libellé de scope est documentée par `R7_MANIFEST_RECONCILIATION.md` et le manifest v2, sans réécriture des archives v1.

## Environnements et outils indisponibles

Semgrep, Grype, Shellcheck, Yamllint, Docker/Buildx, Camoufox, SystemVault natif, Windows, macOS et une validation GUI native complète restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` ou bloqués par environnement selon la ligne. Aucun outil absent n’est traité comme PASS. Aucun `xdg-open`, `open`, `rundll32` ou `xattr` n’a été simulé.

## Invariants et verdict

T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts. Aucun compte réel, cookie, secret réel, proxy commercial, site externe ou donnée utilisateur n’a été utilisé. Aucun produit n’a été publié, déployé ou relâché.

```text
GOSEC_R7_CLASSIFIED_WITH_OPEN_FINDINGS
GOSEC_R7_PARTIAL_ENVIRONMENT_UNAVAILABLE
PUBLIC_RELEASE_BLOCKED
SCAN_BLOCKED_UNKNOWN
NATIVE_SYSTEMVAULT_NOT_TESTED
camoflox_execution_authorized=false
t08_authorized=false
release_authorized=false
FORGELOCAL_PRODUCTION_READY=false
```
