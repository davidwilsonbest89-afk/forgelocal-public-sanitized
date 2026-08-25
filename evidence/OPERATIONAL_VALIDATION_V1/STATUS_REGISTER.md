# FORGELOCAL_OPERATIONAL_VALIDATION_V1 — registre

**Branche :** `validation/operational-v1`  
**HEAD courant documenté :** `b1559ca53852c493ba15e4a06ad89b0c171c7938`  
**T28 :** `T28_APPROVED_VERIFIABLE_LOCAL`, non rouvert.

| Étape | Statut | Preuve principale | Commentaire |
|---|---|---|---|
| V-CORE | FAIL | `V-CORE_RAW.log` | 23 PASS ; expiration/révocation du token admin absente, FAIL non critique. |
| V-SQLITE-CRASH-RECOVERY | PASS | `V-SQLITE-CRASH-RECOVERY_RAW.log` | WAL, concurrence, crash transactionnel, corruption/permissions et redaction vérifiés sur temporaires. |
| V-T28-RUNTIME-CONTRACT | PASS | `V-T28-RUNTIME-CONTRACT_RAW.log` | Tests T28 et checks Go exit 0 sur le nouveau HEAD. |
| V-RUNTIME-SYNTHETIQUE | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE pour Camoufox/extension complète | `V-RUNTIME-SYNTHETIQUE_RAW.log` | Chromium système PASS pour page/profils/crash ; Camoufox absent. |
| V-PROXY-COOKIES-SYNTHETIQUES | PASS | `V-PROXY-COOKIES-SYNTHETIQUES_RAW.log` | Proxy et fixtures locaux redacted uniquement. |
| V-SYSTEMVAULT | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE pour natif | `V-SYSTEMVAULT_RAW.log` | MemoryVault PASS ; Secret Service natif absent. |
| V-DOCKER | NOT_EXECUTED_ENVIRONMENT_UNAVAILABLE | `V-DOCKER_RAW.log` | Docker/Buildx absents. |
| V-DASHBOARD-API | FAIL | `V-DASHBOARD-API_RAW.log`, `V-DASHBOARD-API_A11Y_CONSOLE_RAW.log` | 5 specs PASS ; Axe contraste sérieux et erreurs assets/analytics locales conservées. |
| V-INSTALLATION-PROPRE | PASS avec réserve | `V-INSTALL-FRESH_RAW.log`, `V-INSTALL-FRESH_PRE_FIX_RAW.log` | Correctif 0700 publié ; `init --force` requis pour réécriture explicite. |
| V-SECURITY | FAIL / réserves | `V-SECURITY_RAW.log` | Gitleaks 0 ; Gosec 128 historiques ; OSV 24 packages exit 1 ; pnpm production 0 vulnérabilité ; autres outils absents. |
| V-EVIDENCE | EN COURS | répertoire courant | ZIP/bundle/clone frais à produire et vérifier. |

## Décision courante

Aucun `FAIL_CRITICAL` n’a été constaté. Le verdict de campagne reste **`FORGELOCAL_OPERATIONAL_VALIDATION_PARTIAL_ENVIRONMENT_UNAVAILABLE`**, non production-ready. Il ne pourra être finalisé qu’après publication et vérification des artefacts d’évidence.

## Gates conservées

`PUBLIC_RELEASE_BLOCKED`, `SCAN_BLOCKED_UNKNOWN`, `NATIVE_SYSTEMVAULT_NOT_TESTED`, `camoflox_execution_authorized=false`, `t08_authorized=false`, `release_authorized=false`.

T29 et T39–T42 ne sont pas démarrés.
