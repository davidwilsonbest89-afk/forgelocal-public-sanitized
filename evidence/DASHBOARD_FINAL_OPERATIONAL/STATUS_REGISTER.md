# Status Register — Dashboard final

## Statut courant

> **`DASHBOARD_FINAL_FUNCTIONAL_LOCAL_SYNTHETIC_NOT_PRODUCTION_READY`**

Le Dashboard final est fonctionnel sur le périmètre local synthétique couvert. Ce statut ne lève ni les gates environnementales ni les réserves de sécurité.

| Catégorie | Statut | Détail |
|---|---|---|
| PASS | `PASS` | 5/5 scénarios Dashboard final ; 6/6 avec R2 auth. |
| PASS | `PASS` | Core build/init/health, typecheck, Go race tests ciblés, vet et diff-check. |
| PASS | `PASS` | Axe desktop/mobile 0 ; console/warnings/page errors 0 ; réseau 0 ; clavier et responsive pass. |
| PASS | `PASS` | Espaces, audit session, réglages, aide, notifications et filtres avancés cliqués réellement. |
| PASS | `PASS` | Actions profil et workflows T28 disponibles reliés au Core synthétique. |
| PASS | `PASS` | Cas négatifs 403/404/409/500 ; 401 expired/revoked couvert par la suite R2. |
| FAIL | `FAIL` | Gitleaks arbre complet : 10 findings ; aucune allowlist globale ni suppression introduite. |
| FAIL | `FAIL` | Gosec : 128 findings dans `internal/...`, non masqués. |
| FAIL | `FAIL` | OSV : exit 1 et 46 références advisory affichées. |
| NOT_IMPLEMENTED_PLACEHOLDER | `NOT_IMPLEMENTED_PLACEHOLDER` | Signature cryptographique et attestation de provenance non exposées par le contrat Core ; l’UI l’indique explicitement. |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Camoufox absent, archive utilisateur non fournie. |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Secret Service/SystemVault natif absent. |
| NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Docker/Buildx absent ; aucun daemon privilégié lancé. |
| FAIL_CRITICAL | `NONE_OBSERVED` | Aucun fail critique observé dans les tests exécutés ; cela ne transforme pas les réserves en PASS. |

## Discipline de conservation

T28 historique n’est pas rouvert et son code métier n’est pas modifié. T29 n’est pas démarré. T31–T38 ne sont pas modifiés. Les fixtures, profils, packages et réponses HTTP sont synthétiques et locales.

Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent actifs.
