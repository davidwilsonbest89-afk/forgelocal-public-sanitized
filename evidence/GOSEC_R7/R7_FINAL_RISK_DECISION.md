# GOSEC-R7 — décision finale de risque ciblé

## Décision

La baseline R7 a été créée depuis le HEAD R6 distant découvert dynamiquement `3656dbad4bfef0381e1f9d837271d293ecffe292`. Le JSON Gosec R7 source-only contient exactement 46 findings, avec le même SHA-256 que le JSON R6-C autoritatif : `62206b6c4e9375f112f5bc3dcfceba6700fb9147a521ff075eebd78b9c090e3a`.

Les tests d’atteignabilité autorisés n’ont démontré aucun P0 ni P1. Les chemins testés utilisent uniquement des fixtures locales, des profils jetables, des tokens synthétiques et des services loopback. Aucun secret réel, compte réel, cookie, proxy commercial, site externe ou donnée utilisateur n’a été utilisé.

| Classification | Nombre | Portée |
|---|---:|---|
| `CORRECTION_REQUISE_P0` | 0 | Aucun risque critique démontré |
| `CORRECTION_REQUISE_P1` | 0 | Aucun risque urgent démontré |
| `MITIGATED_CONTROL_SCANNER_OPEN` | 23 | G115=3, G302=5, G305=1, G703=7, G704=7 |
| `SCANNER_OPEN_MANUAL_REVIEW` | 18 | G101=1, G404=17 |
| `BLOCKED_ENVIRONMENT_REQUIRED` | 5 | G204=5, validation native/GUI requise |
| `HISTORICAL_NOT_REACHABLE` | 0 | Aucun finding courant classé historique |
| **Total** | **46** | Distribution exacte du JSON |

## Décision par risque

Les G703, G305, G302, G115 et G704 restent scanner-visibles avec des contrôles applicatifs démontrés et des tests négatifs; ils ne sont pas déclarés clos. Les G404 sont limités à la simulation humanisée et à la sélection non cryptographique de fingerprints; ils ne protègent ni token, ni clé, ni autorisation. Le G101 vise le nom de fichier `.api-token.meta`; les métadonnées stockent un digest et l’état du token, pas le credential brut.

Les cinq G204 concernent les appels natifs `open`, `xdg-open`, `rundll32` et `xattr`. La branche Linux possède une présence GUI signalée, mais aucune exécution native n’a été autorisée ou nécessaire pour cette revue; macOS et Windows ne sont pas disponibles. Ils restent donc `BLOCKED_ENVIRONMENT_REQUIRED`, sans PASS simulé.

## Décision avant code

Aucun dernier lot de correction ne démarre automatiquement. Une éventuelle correction future doit être décidée par le propriétaire avec un périmètre fermé, un risque P0/P1 démontré ou un défaut P2 concret, puis un commit source et des preuves séparés. La revue R7 n’a modifié aucun code de production.

## Gates

Les tests d’archives, runtime, permissions, token, workflow et réseau, la suite race, vet et build ont retourné exit code 0. Gosec retourne exit code 1 avec les 46 findings attendus. Gitleaks source-only et extraction `--no-git`, Govulncheck, OSV Go/pnpm, Trivy et Syft sont PASS dans leurs périmètres documentés. Semgrep, Grype, Shellcheck, Yamllint, Docker/Buildx, Camoufox, SystemVault natif, Windows, macOS et validation GUI native restent `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` ou bloqués par environnement selon la ligne concernée.

## Conservation et invariants

Le package R6-C a été vérifié depuis un clone neuf : ZIP/TAR, sidecars, SHA-256, manifestes, extraction, Gitleaks d’extraction, bundle, checkout et `git fsck --full` sont PASS. Les packages R6 ne sont pas recréés. R5 reste à `28f66a1`, `validation/operational-v1` reste à `8004818`, et `evidence/SMOKE_INTEGRATED_PROXY/` reste local et non suivi. T28 n’est pas rouvert, T29 n’est pas démarré et T31–T38 restent intacts.

## Verdict

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
