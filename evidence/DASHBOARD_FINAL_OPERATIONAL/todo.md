# Todo — Dashboard final

| Sujet | Statut | Condition de clôture |
|---|---|---|
| Signature cryptographique et attestation de provenance | `NOT_IMPLEMENTED_PLACEHOLDER` | Contrat Core explicite, vérification cryptographique réelle, provenance attestée et tests négatifs. |
| Gitleaks | `FAIL` | Examiner les 10 findings de l’arbre complet ; aucune allowlist globale ajoutée dans ce lot. |
| Gosec | `FAIL` | Traiter les 128 findings `internal/...` ou produire des exceptions individuelles justifiées. |
| OSV | `FAIL` | Résoudre les 46 références advisory ou rejouer avec une toolchain/scanner compatible et documenter chaque exception. |
| Camoufox | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Archive Camoufox vérifiée fournie par l’utilisateur ; aucune substitution Chromium. |
| SystemVault natif | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Secret Service/keyring réel disponible en environnement de test dédié. |
| Docker/Buildx | `NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE` | Daemon Docker/Buildx réel disponible ; build, inspection et Trivy image. |
| Persistance des espaces/réglages/notifications | `LOCAL_SESSION_SCOPE` | Décision produit et contrat Core si une persistance inter-session est requise. |
| T29 | `NOT_STARTED` | Hors périmètre de cette mission. |
| T31–T38 | `UNCHANGED` | Ne pas modifier dans cette mission. |

Les gates `PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false` et `release_authorized=false` restent actifs.
