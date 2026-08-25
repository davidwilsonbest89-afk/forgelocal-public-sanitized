# T28 — Décision finale locale

## Verdict

> `T28_IMPLEMENTED_VERIFIABLE_LOCAL_PENDING_INDEPENDENT_REVIEW`

Le lot T28 est implémenté et vérifiable localement, mais **non approuvé** : une revue indépendante du lot reste requise. Aucun résultat de ce lot ne lève une gate ou n’autorise le runtime d’extension.

## Références publiées

| Élément | Référence |
|---|---|
| Branche | `feature/t28-local-extensions-controlled` |
| Baseline V6 | tag `t00-t42-v6-local-qualified-2026-08-25`, commit `999374d99b7996504ba91e421850a2fe84afb78d` |
| Commit de contenu du package | `a9c47fd422d472c4a44d2703d844c25c27e766d9` |
| ZIP | `evidence/T28/t28-local-extensions-controlled-v1.zip` |
| SHA-256 ZIP | `fa657b372d516738fc8617ffb634be82871a17c6afeafb9e556ac9cb020128ee` |
| Bundle delta | `evidence/T28/t28-local-extensions-controlled.delta.bundle` |
| SHA-256 bundle | `845d20de9b8607f7d146175ba18b7ba4804e0428277cbaf134703e54461d73a2` |
| Sidecars | `*.zip.portable.sha256` et `*.bundle.portable.sha256` |

Le bundle est un bundle v2 manuel qui requiert explicitement la baseline ; `git bundle verify` reconnaît la ref T28 et le prerequisite. Le ZIP a passé `unzip -t`, extraction fraîche et 43 contrôles checksum. Le manifeste ne s’auto-référence pas.

## Implémentation

T28 fournit une surface locale-only pour importer un package ZIP, parser un manifest borné, conserver toutes les permissions déclarées, signaler les permissions/host patterns high-risk, demander un acknowledgement complet, approuver, affecter à un profil existant, versionner immuablement, rollbacker, révoquer/quarantiner et purger sous conditions. Les métadonnées API et l’audit sont redacted ; les blobs restent sous le répertoire géré, avec permissions owner-only. Les handlers utilisent les protections bearer, loopback et origin existantes.

Aucun navigateur, Camoufox, Chromium, runtime d’extension, chargement/exécution, téléchargement de package externe, proxy/cookie réel, SystemVault natif, migration, production runtime ou release n’a été démarré.

## Validation

| Contrôle | Résultat |
|---|---|
| `go test -count=1 -race ./internal/extensions ./internal/api -run '^TestT28'` | PASS |
| `go vet ./internal/extensions ./internal/api` | PASS |
| `go build ./...` | PASS |
| `go vet ./...` | PASS |
| `govulncheck ./...` | PASS — aucune vulnérabilité |
| Gitleaks extraction et diff | PASS — rapports vides ; limitation `0 commits scanned` du mode `--log-opts` conservée séparément |
| Gosec `internal/extensions` | PASS — `found=0` |
| Syft SPDX | Généré |
| OSV Scanner | NON EXÉCUTÉ — binaire absent, code 127 |
| Suite globale `go test -count=1 -race ./...` | Échec sur finding runtime V6 préexistant `TestNewRegistryLoadsBrowseForgeChromiumFromDefaultConfig`; aucun échec T28 |

Les journaux bruts, rapports, baseline et scripts de conservation sont sous `evidence/T28/`. Les erreurs procédurales initiales de CWD/option sont conservées mais ne sont pas utilisées comme preuves de succès ; les contrôles corrigés sont les références autoritatives.

## Statuts et arrêt

T29, T39, T40, T41 et T42 ne sont pas démarrés. T30 conserve son statut antérieur. Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoulox_execution_authorized=false` / `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent inchangées. L’exécution s’arrête ici en attente de la revue indépendante T28 et des autorisations séparées prévues pour les lots ultérieurs.
