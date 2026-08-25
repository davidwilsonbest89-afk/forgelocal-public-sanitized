# Todo — OPV1 R2

## Restant avant tout verdict complet ou release

| Action | Statut R2 | Condition de clôture |
|---|---|---|
| Qualification Camoufox ciblée | `BLOCKED_ENVIRONMENT` | Archive Camoufox vérifiée et fournie par l’utilisateur ; aucune substitution Chromium. |
| Qualification SystemVault natif | `BLOCKED_ENVIRONMENT` | Secret Service réellement disponible avec session/keyring de test ; aucune écriture réelle non autorisée. |
| Qualification Docker/Buildx | `BLOCKED_ENVIRONMENT` | Environnement Docker/Buildx dédié ; build inspecté, non-root, permissions, HEALTHCHECK et Trivy image. |
| Revue Gosec | `OPEN_RESERVE` | Examiner les 128 findings historiques et documenter les corrections ou exceptions individuelles. |
| Revue OSV | `OPEN_RESERVE` | Rejouer avec un scanner compatible Go 1.25 et traiter les advisories ; le résultat actuel reste exit 1. |
| Outils non disponibles | `OPEN_RESERVE` | Semgrep, Grype, Shellcheck et Yamllint dans un environnement qui les fournit. |
| Rotation opérateur du token | `NOT_IN_R2` | Définir et implémenter ultérieurement une procédure d’émission/rotation ; R2 ne revendique pas une rotation automatique. |
| Dashboard final | `NOT_IMPLEMENTED_PLACEHOLDER` | Décisions produit et implémentation des espaces, audit, réglages, filtres avancés et parcours T28 UI. |
| T29 | `NOT_STARTED` | Ne pas commencer sans décision produit séparée. |
| T31–T38 | `UNCHANGED` | Ne pas modifier dans R2. |

## Garde-fous

Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent actifs. Aucun todo bloqué par l’environnement ne doit être converti en PASS par un skip ou une substitution silencieuse.
