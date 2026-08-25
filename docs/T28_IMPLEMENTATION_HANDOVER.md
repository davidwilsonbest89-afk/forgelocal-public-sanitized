# T28 — Handover d’implémentation contrôlée

## Décision

> `T28_IMPLEMENTED_VERIFIABLE_LOCAL_PENDING_INDEPENDENT_REVIEW`

Cette décision signifie que l’implémentation locale et ses contrôles reproductibles sont publiés, mais qu’aucune revue indépendante physique de ce lot n’a encore été effectuée. Elle ne signifie ni approbation runtime, ni compatibilité navigateur, ni autorisation de release.

## Baseline et publication

| Élément | Valeur |
|---|---|
| Baseline V6 | tag `t00-t42-v6-local-qualified-2026-08-25` |
| Commit baseline | `999374d99b7996504ba91e421850a2fe84afb78d` |
| Branche publiée | `feature/t28-local-extensions-controlled` |
| HEAD publié d’implémentation | `4f0f6201e1d8f8da44d82c4245bd9b7dfee44578` |
| Baseline brute | `evidence/T28/BASELINE_DISCOVERY_RAW.log` |
| Rapport contrat | `docs/T28_EXTENSIONS_CONTRACT_AND_TEST_MATRIX.md` |
| Décisions produit | `docs/T28_EXTENSIONS_PRODUCT_DECISIONS.md` |
| Modèle de menace | `docs/T28_EXTENSIONS_THREAT_MODEL.md` |

## Fonctionnalités livrées

Le lot fournit un repository SQLite local-first sous le Profile Store root, une extraction ZIP bornée et refusant traversal/symlink, un stockage d’objet immuable et owner-only, l’import/update de versions, le signalement de permissions et host patterns à risque, l’acknowledgement explicite, l’approbation high-risk, l’affectation à un profil existant, le rollback, la révocation/quarantaine, la purge contrôlée et un audit redacted. Les handlers Core sont intégrés au middleware bearer, loopback et origin guard existant. Aucune fonction ne charge ou n’exécute une extension.

## Validation réelle

| Contrôle | Résultat |
|---|---|
| `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'` | PASS |
| `go vet ./internal/extensions ./internal/api` | PASS |
| `go build ./...` | PASS |
| `govulncheck ./...` | PASS — aucune vulnérabilité trouvée |
| Gitleaks extraction | PASS — rapport vide |
| Gitleaks plage diff | PASS — rapport vide ; plage Git matérialisée séparément avec 3 commits |
| Gosec `internal/extensions` | PASS — `found=0`, une annotation locale justifiée pour permissions 0700 |
| SBOM Syft SPDX | Généré |
| `osv-scanner` | NON EXÉCUTÉ — binaire absent, code 127 |
| Suite globale `go test -count=1 -race ./...` | BLOQUÉE par finding runtime V6 préexistant dans `internal/runtime/runtime_test.go`; les packages T28 passent |

Les preuves brutes et rapports sont sous `evidence/T28/`. Les premières commandes mal cadrées (CWD/option Gitleaks) sont conservées dans le journal, et les contrôles corrigés sont explicitement identifiés ; aucun résultat vide n’est utilisé comme preuve de succès.

## Limites et gates

La revue indépendante T28 reste requise. Aucun Camoufox, Chromium, proxy/cookie réel, SystemVault natif, migration, téléchargement de package externe, store d’extensions, runtime de production ou release n’a été lancé. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoulox_execution_authorized=false` / orthographe historique `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangées.

T29, T39, T40, T41 et T42 ne sont pas démarrés par ce lot. T30 conserve son statut antérieur et n’est pas requalifié par T28.
